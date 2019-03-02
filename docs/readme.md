# Docs

## Deployment

### MacOS

Copy the binary to `/usr/local/bin`

Make the file `/Library/LaunchDaemons/com.gavinelder.puppet-nanny.plist` with contents

```shell
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Disabled</key>
    <false/>
    <key>Label</key>
    <string>com.gavinelder.puppet-nanny</string>
    <key>Program</key>
    <string>/usr/local/bin/puppet-nanny</string>
    <key>RunAtLoad</key>
    <true/>
    <key>StartCalendarInterval</key>
    <array>
        <dict>
            <key>Minute</key>
            <integer>0</integer>
        </dict>
        <dict>
            <key>Minute</key>
            <integer>30</integer>
        </dict>
    </array>
</dict>
</plist>
```

Load the service

```shell
launchctl load -w /Library/LaunchDaemons/com.gavinelder.puppet-nanny.plist
```

### Debian

Copy the binary to `/usr/local/bin`

Make the unit file `/lib/systemd/system/puppet-nanny.service`

```shell
[Unit]
Description=Puppet wrapper to handle issues such as incomplete runs
After=network.target

 [Service]
ExecStart=/usr/local/bin/puppet-nanny.py

 [Install]
WantedBy=default.target
```

Reload the service daemon

```shell
systemctl daemon-reload
```

Start the service

```shell
systemctl service puppet-nanny start
```
