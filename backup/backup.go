package backup

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Opts struct {
	Service       string
	AppSupportDir string

	Region string
	Bucket string
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
	if path, err := exec.LookPath("pigz"); err == nil {
		log.Printf("found pigz at %v", path)
		return "pigz"
	}
	log.Println("pigz not found; falling back to gz")
	return "gz"
}

func Run(svc *s3.S3, opts *Opts) error {
	oldest, err := oldestObject(svc, opts.Bucket, opts.Prefix)
	if err != nil {
		return fmt.Errorf("failed to retrieve oldest backup: %v", err)
	}

	//if err = exec.Command("systemctl", "stop", plexService).Run(); err != nil {
	//	return fmt.Errorf("failed to stop plex: %v", err)
	//}

	//tar := exec.Command(
	//	"tar", "-cf", "-",
	//	"-C", plexDir,
	//	"--exclude", "Cache",
	//	"--exclude", "Crash Reports",
	//	"--exclude", "Diagnostics",
	//	"Plex Media Server")
	tar := exec.Command(
		"tar", "-cf", "-",
		"-C", "/home/george/Documents",
		"--exclude", "Cache",
		"--exclude", "Crash Reports",
		"--exclude", "Diagnostics",
		"debian")

	gz := exec.Command(findGzCommand(), "-c")
	gz.Stdin, err = tar.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout from tar: %v", err)
	}

	gzStdout, err := gz.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout from gz: %v", err)
	}

	if err = tar.Start(); err != nil {
		return fmt.Errorf("failed to start tar: %v", err)
	}
	if err = gz.Start(); err != nil {
		return fmt.Errorf("failed to start gz: %v", err)
	}

	key := calculateKey(opts.Prefix)
	uploader := s3manager.NewUploaderWithClient(svc)
	_, uploadErr := uploader.Upload(&s3manager.UploadInput{
		Bucket: &opts.Bucket,
		Key:    &key,
		Body:   gzStdout,
	})

	//if err = exec.Command("systemctl", "start", plexService).Run(); err != nil {
	//	return fmt.Errorf("failed to start plex: %v", err)
	//}

	// TODO unclear whether this is necessary; the example uses it
	if err = tar.Wait(); err != nil {
		return fmt.Errorf("failed to wait for tar: %v", err)
	}

	if uploadErr != nil {
		return fmt.Errorf("failed to upload new backup: %v", uploadErr)
	}

	if oldest != nil {
		_, err := svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: &opts.Bucket,
			Key:    oldest.Key,
		})
		if err != nil {
			log.Printf("Failed to delete %v: %v\n", oldest, err)
		}
	}
	return nil
}
