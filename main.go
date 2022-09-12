// Package main implements a CLI frontend for the backup package.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gebn/plexbackup/backup"

	"github.com/alecthomas/kingpin"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gebn/go-stamp/v2"
)

var (
	bucket = kingpin.Flag("bucket", "Name of the S3 bucket to upload the backup to.").
		Required().
		String()
	region = kingpin.Flag("region", "Region of the --bucket.").
		Required().
		String()
	prefix = kingpin.Flag("prefix", `Location within the bucket to upload to. This will be suffixed with <RFC3339 date>.tar.gz, e.g. "2019-01-06T22:38:21Z.tar.gz".`).
		Default("plex/").
		String()

	noPause = kingpin.Flag("no-pause", "Do not stop Plex while the backup is performed. This is not recommended, as it risks an inconsistent backup.").
		Bool()
	service = kingpin.Flag("service", "Name of the Plex systemd unit to stop while the backup is performed.").
		Default("plexmediaserver.service").
		String()
	directory = kingpin.Flag("directory", "Location of the 'Plex Media Server' directory to back up.").
			Default("/var/lib/plexmediaserver/Library/Application Support/Plex Media Server").
			String() // ExistingDir() breaks --version if does not exist (#261)
)

func main() {
	if err := actualMain(context.Background()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func actualMain(ctx context.Context) error {
	kingpin.Version(stamp.Summary())
	kingpin.Parse()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(*region),
		config.WithUseDualStackEndpoint(aws.DualStackEndpointStateEnabled))
	if err != nil {
		return fmt.Errorf("failed to initialise AWS SDK: %w", err)
	}

	s3client := s3.NewFromConfig(cfg)
	return backup.Run(ctx, s3client, &backup.Opts{
		NoPause:   *noPause,
		Service:   *service,
		Directory: *directory,
		Bucket:    *bucket,
		Prefix:    *prefix,
	})
}
