# Plex Backup

[![Build Status](https://travis-ci.org/gebn/plexbackup.svg?branch=master)](https://travis-ci.org/gebn/plexbackup)
[![GoDoc](https://godoc.org/github.com/gebn/plexbackup?status.svg)](https://godoc.org/github.com/gebn/plexbackup)

This tool backs up [Plex](https://www.plex.tv)'s server data and uploads it to S3.
The entire `Plex Media Server` directory (sans `Cache`) is snapshotted, stopping the server beforehand and restarting it after to ensure consistency.
The directory is `tar`ed, `gz`ipped and uploaded without writing any data to disk.
The tool is envisaged to be run as a cron job, preferably soon after the configured maintenance period.
As the latest instance types have 2 or more cores, [`pigz`](https://zlib.net/pigz/) will be used in place of `gz` if available on the `$PATH`.

## Usage

    $ plexbackup --help
    usage: plexbackup [<flags>]

    Flags:
      --help                Show context-sensitive help (also try --help-long and --help-man).
      --bucket=BUCKET       Bucket to upload to
      --region="eu-west-2"  Region of the S3 bucket
      --prefix="plex"       Prefix to prepend to the backup object key
      --service="plexmediaserver.service"  
                            Name of the Plex systemd unit
      --directory=/var/lib/plexmediaserver/Library/Application Support  
                            Location of the 'Application Support' directory
