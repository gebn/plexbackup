# Plex Backup

[![Build Status](https://travis-ci.org/gebn/plexbackup.svg?branch=master)](https://travis-ci.org/gebn/plexbackup)
[![GoDoc](https://godoc.org/github.com/gebn/plexbackup?status.svg)](https://godoc.org/github.com/gebn/plexbackup)

This tool backs up the [`Plex Media Server`](https://www.plex.tv) directory (sans `Cache`) to S3.
It is `tar`red, `gz`ipped and uploaded without writing to disk.
The tool is envisaged to be run as a cron job, preferably soon after the configured maintenance period.
The process is usually CPU-bound on the compression, so [`pigz`](https://zlib.net/pigz/) will be used in place of `gz` if available on the `$PATH`.

## Setup

In order to ensure consistency of the backup, Plex is stopped then started once the process is complete.
This tool effectively runs `sudo systemctl stop|start <unit>` to do this (Polkit was also considered but [ultimately rejected](https://github.com/gebn/plexbackup/issues/6#issuecomment-452899467)).
Assuming a vanilla installation, this can be made to work by allowing the `plex` user to execute the two required commands without having to re-authenticate:

    # cat <<EOF > /etc/sudoers.d/10-plex-backup
    plex ALL=NOPASSWD: /bin/systemctl stop plexmediaserver.service
    plex ALL=NOPASSWD: /bin/systemctl start plexmediaserver.service
    EOF

## Usage

    $ plexbackup --help
    usage: plexbackup --bucket=BUCKET --region=REGION [<flags>]

    Flags:
      --help            Show context-sensitive help (also try --help-long and --help-man).
      --bucket=BUCKET   Name of the S3 bucket to upload the backup to.
      --region=REGION   Region of the --bucket.
      --prefix="plex/"  Location within the bucket to upload to. This will be suffixed with <RFC3339
                        date>.tar.gz, e.g. "2019-01-06T22:38:21Z.tar.gz".
      --no-pause        Do not stop Plex while the backup is performed. This is not recommended, as it
                        risks an inconsistent backup.
      --service="plexmediaserver.service"  
                        Name of the Plex systemd unit to stop while the backup is performed.
      --directory=/var/lib/plexmediaserver/Library/Application Support/Plex Media Server
                        Location of the 'Plex Media Server' directory to back up.
