# Plex Backup

[![Build Status](https://travis-ci.org/gebn/plexbackup.svg?branch=master)](https://travis-ci.org/gebn/plexbackup)
[![GoDoc](https://godoc.org/github.com/gebn/plexbackup?status.svg)](https://godoc.org/github.com/gebn/plexbackup)

This tool backs up [Plex](https://www.plex.tv)'s server data to S3.
The entire `Plex Media Server` directory (sans `Cache`) is snapshotted, stopping the server beforehand and restarting it after to ensure consistency.
It is `tar`red, `gz`ipped then piped to S3 without writing to disk.
The tool is envisaged to be run as a cron job, preferably soon after the configured maintenance period.
The process is usually CPU-bound on the compression, so [`pigz`](https://zlib.net/pigz/) will be used in place of `gz` if available on the `$PATH`.

## Usage

    $ plexbackup --help
    usage: plexbackup [<flags>]

    Flags:
      --help                Show context-sensitive help (also try --help-long and --help-man).
      --bucket=BUCKET       Name of the S3 bucket to upload the backup to.
      --region="eu-west-2"  Region of the --bucket; defaults to eu-west-2, or AWS_REGION if set.
      --prefix="plex"       Location within the bucket to upload to; a trailing slash is added if not present.
                            The backup object is stored under this prefix as <RFC3339 date>.tar.xz, e.g.
                            "2019-01-06T22:38:21Z.tar.xz".
      --service="plexmediaserver.service"  
                            Name of the Plex systemd unit to stop while the backup is performed.
      --directory=/var/lib/plexmediaserver/Library/Application Support/Plex Media Server
                            Location of the 'Plex Media Server' directory to back up.
