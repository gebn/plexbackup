# Plex Backup

[![CI](https://github.com/gebn/plexbackup/actions/workflows/ci.yaml/badge.svg)](https://github.com/gebn/plexbackup/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/gebn/plexbackup.svg)](https://pkg.go.dev/github.com/gebn/plexbackup)
[![Go Report Card](https://goreportcard.com/badge/github.com/gebn/plexbackup)](https://goreportcard.com/report/github.com/gebn/plexbackup)

Backs up the [`Plex Media Server`](https://www.plex.tv) directory to S3.
Intended to run as a cron job, ideally soon after the configured maintenance period.
The directory (excluding `Cache`) is passed through `tar`, compressed with Zstandard, then uploaded, all without writing to disk.

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
Open the `plex` user's crontab with `crontab -eu plex` as root. Add a line similar to the following:

    22 6 * * * /opt/plexbackup/plexbackup --bucket thebrightons-backup-euw2 --region eu-west-2 --prefix plex/newton- 2>> /your/log/file

Choose a time that doesn't overlap with the server's background task hours. The best time to run the backup is soon after these tasks have finished.
Note logs are written to `stderr`, rather than `stdout`.

## Usage

    $ plexbackup --help
    Usage of ./plexbackup:
      -bucket string
            name of the S3 bucket to upload the backup to
      -directory string
            path of the 'Plex Media Server' directory to back up (default "/var/lib/plexmediaserver/Library/Application Support/Plex Media Server")
      -no-pause
            suppresses stopping Plex while the backup is performed, risks an inconsistent backup
      -prefix string
            suffixed with "<RFC3339 date>.tar.zst" to form the upload key (default "plex/")
      -region string
            region of the -bucket (default "us-east-1")
      -service string
            name of the Plex systemd unit to stop, redundant if -no-pause used (default "plexmediaserver.service")
      -version
            display software version and exit
