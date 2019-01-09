// Package backup creates and uploads Plex backups to S3.
package backup

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Opts encapsulates parameters for backing up Plex's database.
type Opts struct {

	// NoPause takes the backup without stopping Plex. The server will remain
	// available throughout, but the backup may be unusable.
	// It is specified negatively in order to default to false, which is the
	// recommended setting.
	NoPause bool

	// Service is the name of Plex's systemd unit, e.g. plexmediaserver.service
	Service string

	// Directory is the path to the 'Plex Media Server' directory, which will
	// form the root directory of the produced backup.
	Directory string

	// Bucket is the name of the bucket to upload the backup to
	Bucket string

	// Region is the region of the bucket, used to connect to the
	// correct region's endpoint
	Region string

	// Prefix is the path within the bucket to look for the old backup,
	// and upload the new one to
	Prefix string
}

// oldestObject returns the object with the oldest LastModified attribute within
// a given bucket under a given prefix, or nil if no objects exist there. It
// assumes the prefix contains <=1000 objects (no pagination is attempted).
func oldestObject(svc *s3.S3, bucket, prefix string) (*s3.Object, error) {
	result, err := svc.ListObjects(&s3.ListObjectsInput{
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

// calculateKey returns the location to upload the new backup to.
// N.B. tar interprets names containing colons as network locations, so it must
// be piped in, e.g. tar -xzf - < name:with:colons.tar.xz.
func calculateKey(prefix string) string {
	now := time.Now().UTC()
	return path.Join(prefix, now.Format(time.RFC3339)+".tar.xz")
}

// gzCommand determines the correct implementation of gzip to use: if pigz
// is available, it is prefered, otherwise we fall back on gz, assuming it
// exists.
func findGzCommand() string {
	if _, err := exec.LookPath("pigz"); err == nil {
		return "pigz"
	}
	return "gz"
}

// Run stops Plex, performs the backup, then starts Plex again.
// It should ideally be run soon after the server maintenance period.
func (o *Opts) Run(svc *s3.S3) error {
	oldest, err := oldestObject(svc, o.Bucket, o.Prefix)
	if err != nil {
		return fmt.Errorf("failed to retrieve oldest backup: %v", err)
	}

	if !o.NoPause {
		if err = exec.Command("sudo", "systemctl", "stop", o.Service).Run(); err != nil {
			return fmt.Errorf("failed to stop plex: %v", err)
		}
	}

	tar := exec.Command(
		"tar", "-cf", "-",
		"-C", filepath.Dir(o.Directory),
		"--exclude", "Cache",
		"--exclude", "Crash Reports",
		"--exclude", "Diagnostics",
		filepath.Base(o.Directory))

	gz := exec.Command(findGzCommand(), "-c")
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

	key := calculateKey(o.Prefix)
	uploader := s3manager.NewUploaderWithClient(svc)
	_, uploadErr := uploader.Upload(&s3manager.UploadInput{
		Bucket: &o.Bucket,
		Key:    &key,
		Body: struct {
			io.Reader
		}{gzStdout},
	})

	if !o.NoPause {
		if err = exec.Command("sudo", "systemctl", "start", o.Service).Run(); err != nil {
			return fmt.Errorf("failed to start plex: %v", err)
		}
	}

	// TODO unclear whether this is necessary; the example uses it
	if err = tar.Wait(); err != nil {
		return fmt.Errorf("failed to wait for tar: %v", err)
	}

	elapsed := time.Since(start)

	if uploadErr != nil {
		return fmt.Errorf("failed to upload new backup: %v", uploadErr)
	}

	log.Printf("Completed backup to %v in %v", key, elapsed.Round(time.Millisecond))

	if oldest != nil {
		_, err := svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: &o.Bucket,
			Key:    oldest.Key,
		})
		if err != nil {
			log.Printf("Failed to delete %v: %v\n", oldest, err)
		}
	}
	return nil
}
