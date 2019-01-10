// Package backup creates and uploads Plex Media Server backups to S3.
//
// Example
//
// The following snippet shows how to perform a vanilla backup. Plex will be stopped before
// the backup begins, and started again after it finishes.
//
//   region := "eu-west-2"
//
//   sess := session.Must(session.NewSession(&aws.Config{
//   	Region: region,
//   }))
//   svc := s3.New(sess)
//
//   opts := &backup.Opts{
//   	Service:   "plexmediaserver.service",
//   	Directory: "/var/lib/plexmediaserver/Library/Application Support/Plex Media Server",
//   	Bucket:    "eu-west-2.backups.thebrightons.co.uk",
//   	Region:    region,
//   	Prefix:    "plex/newton/",
//   }
//   if err := opts.Run(svc); err != nil {
//   	log.Fatal(err)
//   }
package backup

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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

	// Service is the name of Plex's systemd unit, e.g. plexmediaserver.service,
	// which will be stopped while the backup is performed, and started again
	// after it completes.
	Service string

	// Directory is the path to the 'Plex Media Server' directory, which will
	// form the root directory of the produced backup.
	Directory string

	// Bucket is the name of the S3 bucket to upload the backup to.
	Bucket string

	// Region is the region of the Bucket, used to connect to the correct
	// endpoint.
	Region string

	// Prefix is prepended to "<RFC3339 date>.tar.gz" to form the path of the
	// backup object, e.g. "2019-01-06T22:38:21Z.tar.gz". N.B. no slash is
	// automatically added to the end of the prefix.
	// This is also the prefix under which we query for old backups - if it
	// changes, unless the new value is a prefix of the old one, the previous
	// backup will not be discovered and deleted by this tool.
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
	return prefix + now.Format(time.RFC3339) + ".tar.gz"
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
		"--exclude", "plexmediaserver.pid",
		filepath.Base(o.Directory))
	tar.Stderr = os.Stderr

	gz := exec.Command(findGzCommand(), "-c")
	gz.Stderr = os.Stderr
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

	if err = gz.Wait(); err != nil {
		return fmt.Errorf("gz completed improperly: %v", err)
	}
	if err = tar.Wait(); err != nil {
		return fmt.Errorf("tar completed improperly: %v", err)
	}

	// we could have deferred this after stopping plex, however this would not allow us
	// to report an error - this way the caller can be confident Plex is running if they
	// get back a nil error
	if !o.NoPause {
		if err = exec.Command("sudo", "systemctl", "start", o.Service).Run(); err != nil {
			return fmt.Errorf("failed to start plex: %v", err)
		}
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
			// not regarded as significant enough to report
			log.Printf("Failed to delete %v: %v\n", oldest, err)
		}
	}
	return nil
}
