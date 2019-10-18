# Plex Backup

[![Build Status](https://travis-ci.org/gebn/plexbackup.svg?branch=master)](https://travis-ci.org/gebn/plexbackup)
[![GoDoc](https://godoc.org/github.com/gebn/plexbackup?status.svg)](https://godoc.org/github.com/gebn/plexbackup)
[![Go Report Card](https://goreportcard.com/badge/github.com/gebn/plexbackup)](https://goreportcard.com/report/github.com/gebn/plexbackup)

Backs up the [`Plex Media Server`](https://www.plex.tv) directory to S3.
Intended to run as a cron job, ideally soon after the configured maintenance period.
The directory (excluding `Cache`) is passed through `tar` and `gzip`, then uploaded, without writing to disk.

## Setup

### IAM

Regardless of how the job runs, it requires list, put and delete permissions on the destination bucket. This can be achieved with the following IAM policy:

    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": "s3:ListBucket",
                "Resource": "arn:aws:s3:::<bucket>"
            },
            {
                "Effect": "Allow",
                "Action": [
                    "s3:PutObject",
                    "s3:DeleteObject"
                ],
                "Resource": "arn:aws:s3:::<bucket>/<prefix>*"
            }
        ]
    }

*N.B. if using EC2, an [instance profile](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2.html) can make management much easier.*

### Sudoers

In order to ensure consistency of the backup, Plex is stopped then started once the process is complete.
This tool effectively runs `sudo systemctl stop|start <unit>` to do this (Polkit was also considered but [ultimately rejected](https://github.com/gebn/plexbackup/issues/6#issuecomment-452899467)).
Assuming a vanilla installation, this can be made to work by allowing the `plex` user to execute the two required commands without having to re-authenticate:

    # cat <<EOF > /etc/sudoers.d/10-plex-backup
    plex ALL=NOPASSWD: /bin/systemctl stop plexmediaserver.service
    plex ALL=NOPASSWD: /bin/systemctl start plexmediaserver.service
    EOF

### Cron

Download the [latest release](https://github.com/gebn/plexbackup/releases/latest) to `/opt/plexbackup/` on the Plex server.
Add a line similar to the following to the `plex` user's crontab:

    22 6 * * * /opt/plexbackup/plexbackup --bucket backup.eu-west-2.thebrightons.co.uk --region eu-west-2 --prefix plex/newton- 2> /your/log/file

Choose a time that doesn't overlap with the server's background task hours. The best time to run the backup is soon after these tasks have finished.
Note logs are written to `stderr`, rather than `stdout`.

If you have a relatively underpowered machine but a fast network, consider installing [`pigz`](https://zlib.net/pigz/).
This will be used in place of `gz` if available on the `$PATH`, and can give a good speed boost if bottlenecking on the compression.

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
      --directory="/var/lib/plexmediaserver/Library/Application Support/Plex Media Server"
                        Location of the 'Plex Media Server' directory to back up.
      --version         Show application version.
