// Package main implements a CLI frontend for the backup package.
package main

import (
	"log"

	"github.com/gebn/plexbackup/backup"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	bucket = kingpin.Flag("bucket", "Name of the S3 bucket to upload the backup to.").
		String()
	region = kingpin.Flag("region", "Region of the --bucket; defaults to eu-west-2, or AWS_REGION if set.").
		Default("eu-west-2").
		OverrideDefaultFromEnvar("AWS_REGION").
		String()
	prefix = kingpin.Flag("prefix", `Location within the bucket to upload to; a trailing slash is added if not present. The backup object is stored under this prefix as <RFC3339 date>.tar.xz, e.g. "2019-01-06T22:38:21Z.tar.xz".`).
		Default("plex").
		String()

	service = kingpin.Flag("service", "Name of the Plex systemd unit to stop while the backup is performed.").
		Default("plexmediaserver.service").
		String()
	directory = kingpin.Flag("directory", "Location of the 'Application Support' directory, whose contents will be backed up.").
			Default("/var/lib/plexmediaserver/Library/Application Support").
			ExistingDir()
)

func main() {
	kingpin.Parse()
	sess := session.Must(session.NewSession(&aws.Config{
		Region: region,
	}))
	svc := s3.New(sess)

	opts := &backup.Opts{
		Service:       *service,
		AppSupportDir: *directory,
		Bucket:        *bucket,
		Region:        *region,
		Prefix:        *prefix,
	}
	if err := opts.Run(svc); err != nil {
		log.Fatal(err)
	}
}
