// Package main implements a CLI frontend for the backup package.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gebn/plexbackup/backup"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gebn/go-stamp/v2"
)

var (
	ErrNoBucket = errors.New("bucket name must be specified with -bucket")

	version = flag.Bool("version", false, "display software version and exit")

	bucket = flag.String("bucket", "", "name of the S3 bucket to upload the backup to")
	region = flag.String("region", "us-east-1", "region of the -bucket")
	prefix = flag.String("prefix", "plex/", `suffixed with "<RFC3339 date>.tar.zst" to form the upload key`)

	noPause   = flag.Bool("no-pause", false, "suppresses stopping Plex while the backup is performed, risks an inconsistent backup")
	service   = flag.String("service", "plexmediaserver.service", "name of the Plex systemd unit to stop, redundant if -no-pause used")
	directory = flag.String("directory", "/var/lib/plexmediaserver/Library/Application Support/Plex Media Server", "path of the 'Plex Media Server' directory to back up")
)

func main() {
	if err := actualMain(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func actualMain(ctx context.Context) error {
	flag.Parse()

	if *version {
		fmt.Println(stamp.Summary())
		return nil
	}

	if *bucket == "" {
		return ErrNoBucket
	}

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
