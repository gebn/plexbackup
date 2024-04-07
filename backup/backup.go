// Package backup creates and uploads Plex Media Server backups to S3.
// Plex will be stopped before the backup begins, and started again after it
// finishes.
package backup

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gebn/plexbackup/internal/pkg/countingreader"

	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/klauspost/compress/zstd"
)

// Opts encapsulates parameters for backing up Plex's database.
type Opts struct {

	// NoPause performs the backup without stopping Plex. The server will remain
	// available throughout, but the backup may be unusable. It is specified
	// negatively in order to default to false, which is the recommended
	// setting.
	NoPause bool

	// Service is the name of Plex's systemd unit, e.g. plexmediaserver.service,
	// which will be stopped while the backup is performed, and started again
	// after it completes.
	Service string

	// Directory is the path to the 'Plex Media Server' directory, which will
	// form the root directory of the produced backup.
	Directory string

	// Bucket is the name of the S3 bucket to upload the backup to.
	Bucket string

	// Prefix is prepended to "<RFC3339 date>.tar.zst" to form the path of the
	// backup object, e.g. "2019-01-06T22:38:21Z.tar.zst". N.B. no slash is
	// automatically added to the end of the prefix. This is also the prefix
	// under which we query for old backups - if it changes, unless the new
	// value is a prefix of the old one, the previous backup will not be
	// discovered and deleted by this tool.
	Prefix string
}

// oldestObject returns the object with the oldest LastModified attribute within
// a given bucket under a given prefix, or nil if no objects exist there. It
// assumes the prefix contains <=1000 objects (no pagination is attempted).
func oldestObject(ctx context.Context, client *s3.Client, bucket, prefix string) (*s3types.Object, error) {
	result, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return nil, err
	}

	var oldest *s3types.Object
	for _, object := range result.Contents {
		if oldest == nil || object.LastModified.Before(*oldest.LastModified) {
			oldest = &object
		}
	}
	return oldest, nil
}

// backup performs the actual archive, compression and upload of the backup. It
// blocks until the operation is complete.
func (o *Opts) backup(ctx context.Context, logger *slog.Logger, client *s3.Client) error {
	tar := exec.CommandContext(ctx,
		"tar", "-cf", "-",
		"-C", filepath.Dir(o.Directory),
		"--exclude", "Cache",
		"--exclude", "Crash Reports",
		"--exclude", "Diagnostics",
		"--exclude", "plexmediaserver.pid",
		filepath.Base(o.Directory))
	tar.Stderr = os.Stderr
	tarStdoutReader, err := tar.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe from tar: %w", err)
	}

	// Turns the bytes written by zstd into something that can be read by the
	// AWS SDK.
	zstdReader, zstdWriter := io.Pipe()

	enc, err := zstd.NewWriter(zstdWriter)
	if err != nil {
		return err
	}

	type compressResult struct {
		UncompressedBytes uint64
		Error             error
	}

	compressResultChan := make(chan compressResult)
	go func() {
		uncompressedBytes, err := enc.ReadFrom(tarStdoutReader)
		compressResultChan <- compressResult{uint64(uncompressedBytes), err}
	}()

	uploader := s3manager.NewUploader(client)
	key := o.Prefix + time.Now().UTC().Format(time.RFC3339) + ".tar.zst"
	reader := countingreader.New(zstdReader)
	uploadErr := make(chan error)
	go func() {
		_, err := uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket: &o.Bucket,
			Key:    &key,
			Body:   reader,
		})
		uploadErr <- err
	}()

	start := time.Now()

	if err = tar.Start(); err != nil {
		return fmt.Errorf("failed to start tar: %w", err)
	}

	if err = tar.Wait(); err != nil {
		return fmt.Errorf("tar completed with error: %w", err)
	}

	zstdResult := <-compressResultChan
	if err := zstdResult.Error; err != nil {
		return fmt.Errorf("zstd completed with error: %w", err)
	}

	if err := enc.Close(); err != nil {
		return fmt.Errorf("failed to close zstd stream: %w", err)
	}

	// Should indicate to the S3 uploader that we are done, so it returns.
	zstdWriter.Close()

	if <-uploadErr != nil {
		return fmt.Errorf("failed to upload new backup: %w", uploadErr)
	}

	logger.InfoContext(ctx, "uploaded backup",
		slog.String("key", key),
		slog.Duration("elapsed", time.Since(start)),
		slog.Uint64("uncompressed_bytes", zstdResult.UncompressedBytes),
		slog.Uint64("compressed_bytes", reader.ReadBytes))

	return nil
}

// Run stops Plex, performs the backup, then starts Plex again. It should
// ideally be run soon after the server maintenance period.
func Run(ctx context.Context, logger *slog.Logger, client *s3.Client, o *Opts) error {
	oldest, err := oldestObject(ctx, client, o.Bucket, o.Prefix)
	if err != nil {
		return fmt.Errorf("failed to retrieve oldest backup: %w", err)
	}

	if !o.NoPause {
		logger.DebugContext(ctx, "stopping Plex")
		if err = exec.CommandContext(ctx, "sudo", "systemctl", "stop", o.Service).Run(); err != nil {
			return fmt.Errorf("failed to stop plex: %w", err)
		}
		logger.DebugContext(ctx, "stopped Plex")
	}

	if err = o.backup(ctx, logger, client); err != nil {
		return err
	}

	// We could have deferred this after stopping plex, however this would not
	// allow us to report an error - this way the caller can be confident Plex
	// is running if they get back a nil error.
	if !o.NoPause {
		logger.DebugContext(ctx, "starting Plex")
		if err = exec.CommandContext(ctx, "sudo", "systemctl", "start", o.Service).Run(); err != nil {
			return fmt.Errorf("failed to start plex: %w", err)
		}
		logger.DebugContext(ctx, "started Plex")
	}

	if oldest != nil {
		_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &o.Bucket,
			Key:    oldest.Key,
		})
		if err != nil {
			// Not regarded as significant enough to report.
			logger.WarnContext(ctx, "failed to delete old backup",
				slog.String("key", *oldest.Key),
				slog.String("error", err.Error()))
		} else {
			logger.DebugContext(ctx, "deleted oldest backup",
				slog.String("key", *oldest.Key))
		}
	}

	return nil
}
