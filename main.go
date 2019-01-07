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
	bucket = kingpin.Flag("bucket", "Bucket to upload to").
		String()
	region = kingpin.Flag("region", "Region of the S3 bucket").
		Default("eu-west-2").
		OverrideDefaultFromEnvar("AWS_REGION").
		String()
	prefix = kingpin.Flag("prefix", "Prefix to prepend to the backup object key").
		Default("plex").
		String()

	service = kingpin.Flag("service", "Name of the Plex systemd unit").
		Default("plexmediaserver.service").
		String()
	directory = kingpin.Flag("directory", "Location of the 'Application Support' directory").
			Default("/var/lib/plexmediaserver/Library/Application Support").
			ExistingDir()
)

func main() {
	kingpin.Parse()
	sess := session.Must(session.NewSession(&aws.Config{
		Region: region,
	}))
	svc := s3.New(sess)

	err := backup.Run(svc, &backup.Opts{
		Service:       *service,
		AppSupportDir: *directory,
		Bucket:        *bucket,
		Region:        *region,
		Prefix:        *prefix,
	})
	if err != nil {
		log.Fatal(err)
	}
}
