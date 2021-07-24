// Package backup creates and uploads Plex Media Server backups to S3.
//
// Example
//
// The following snippet shows how to perform a vanilla backup. Plex will be
// stopped before the backup begins, and started again after it finishes.
//
//   sess := session.Must(session.NewSession(&aws.Config{
//   	Region: "eu-west-2",
//   }))
//   svc := s3.New(sess)
//
//   opts := &backup.Opts{
//   	Service:   "plexmediaserver.service",
//   	Directory: "/var/lib/plexmediaserver/Library/Application Support/Plex Media Server",
//   	Bucket:    "thebrightons-backup-euw2",
//   	Prefix:    "plex/newton/",
//   }
//   if err := backup.Run(context.Background(), svc, opts); err != nil {
//   	log.Fatal(err)
//   }
package backup

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gebn/plexbackup/internal/pkg/countingreader"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	// gibibyteBytes is the number of bytes in a GiB.
	gibibyteBytes = 1024 * 1024 * 1024
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

	// Prefix is prepended to "<RFC3339 date>.tar.gz" to form the path of the
	// backup object, e.g. "2019-01-06T22:38:21Z.tar.gz". N.B. no slash is
	// automatically added to the end of the prefix. This is also the prefix
	// under which we query for old backups - if it changes, unless the new
	// value is a prefix of the old one, the previous backup will not be
	// discovered and deleted by this tool.
	Prefix string
}

// oldestObject returns the object with the oldest LastModified attribute within
// a given bucket under a given prefix, or nil if no objects exist there. It
// assumes the prefix contains <=1000 objects (no pagination is attempted).
func oldestObject(ctx context.Context, svc *s3.S3, bucket, prefix string) (*s3.Object, error) {
	result, err := svc.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket: &bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return nil, err
	}

	var oldest *s3.Object
	for _, object := range result.Contents {
		if oldest == nil || object.LastModified.Before(*oldest.LastModified) {
			oldest = object
		}
	}
	return oldest, nil
}

// gzCommand determines the correct implementation of gzip to use: if pigz is
// available, it is preferred, otherwise we fall back on gz, assuming it exists.
func findGzCommand() string {
	if _, err := exec.LookPath("pigz"); err == nil {
		return "pigz"
	}
	return "gz"
}

// backup performs the actual archive, compression and upload of the backup. It
// blocks until the operation is complete.
func (o *Opts) backup(ctx context.Context, svc *s3.S3) error {
	tar := exec.CommandContext(ctx,
		"tar", "-cf", "-",
		"-C", filepath.Dir(o.Directory),
		"--exclude", "Cache",
		"--exclude", "Crash Reports",
		"--exclude", "Diagnostics",
		"--exclude", "plexmediaserver.pid",
		filepath.Base(o.Directory))
	tar.Stderr = os.Stderr

	gz := exec.CommandContext(ctx, findGzCommand(), "-c")
	gz.Stderr = os.Stderr
	var err error
	gz.Stdin, err = tar.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout from tar: %v", err)
	}

	gzStdout, err := gz.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout from gz: %v", err)
	}

	start := time.Now()

	if err = tar.Start(); err != nil {
		return fmt.Errorf("failed to start tar: %v", err)
	}
	if err = gz.Start(); err != nil {
		return fmt.Errorf("failed to start gz: %v", err)
	}

	// N.B. tar interprets names containing colons as network locations, so it
	// must be piped in, e.g. tar -xzf - < name:with:colons.tar.xz.
	key := o.Prefix + time.Now().UTC().Format(time.RFC3339) + ".tar.gz"
	uploader := s3manager.NewUploaderWithClient(svc)
	reader := countingreader.New(gzStdout)
	_, uploadErr := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: &o.Bucket,
		Key:    &key,
		Body:   reader,
	})

	if err = gz.Wait(); err != nil {
		return fmt.Errorf("gz completed improperly: %v", err)
	}
	if err = tar.Wait(); err != nil {
		return fmt.Errorf("tar completed improperly: %v", err)
	}

	if uploadErr != nil {
		return fmt.Errorf("failed to upload new backup: %v", uploadErr)
	}

	elapsed := time.Since(start)
	gib := float64(reader.ReadBytes) / float64(gibibyteBytes)
	log.Printf("backed up %.3f GiB to %v in %v",
		gib, key, elapsed.Round(time.Millisecond))

	return nil
}

// Run stops Plex, performs the backup, then starts Plex again. It should
// ideally be run soon after the server maintenance period.
func Run(ctx context.Context, svc *s3.S3, o *Opts) error {
	oldest, err := oldestObject(ctx, svc, o.Bucket, o.Prefix)
	if err != nil {
		return fmt.Errorf("failed to retrieve oldest backup: %v", err)
	}

	if !o.NoPause {
		if err = exec.CommandContext(ctx, "sudo", "systemctl", "stop", o.Service).Run(); err != nil {
			return fmt.Errorf("failed to stop plex: %v", err)
		}
	}

	if err = o.backup(ctx, svc); err != nil {
		return err
	}

	// we could have deferred this after stopping plex, however this would not
	// allow us to report an error - this way the caller can be confident Plex
	// is running if they get back a nil error
	if !o.NoPause {
		if err = exec.CommandContext(ctx, "sudo", "systemctl", "start", o.Service).Run(); err != nil {
			return fmt.Errorf("failed to start plex: %v", err)
		}
	}

	if oldest != nil {
		_, err := svc.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: &o.Bucket,
			Key:    oldest.Key,
		})
		if err != nil {
			// not regarded as significant enough to report
			log.Printf("failed to delete %v: %v\n", oldest, err)
		}
	}

	return nil
}
