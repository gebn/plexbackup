package main

import (
	"log"

	"github.com/gebn/plexbackup/backup"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	region = "eu-west-2"
	bucket = "backup.eu-west-2.thebrightons.co.uk"
	prefix = "plex/newton/"

	plexService = "plexmediaserver.service"
	plexDir     = "/var/lib/plexmediaserver/Library/Application Support"
)

func main() {
	log.Printf("%v %v", aws.SDKName, aws.SDKVersion)
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := s3.New(sess)

	err := backup.Run(svc, &backup.Opts{
		Service:       plexService,
		AppSupportDir: plexDir,
		Region:        region,
		Bucket:        bucket,
		Prefix:        prefix,
	})
	if err != nil {
		log.Fatal(err)
	}
}
