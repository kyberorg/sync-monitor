# Sync Monitor
Small tool to check delta from last repo sync. 
It reads `lastsync` file, which should contain last repository sync timestamp in unix format.

## Usage
### Minimal working example
```shell
sync-monitor --file /path/to/repo/lastsync
```
### Full example
```shell
sync-monitor --file /path/to/repo/lastsync --port 1221 --interval 2m
```

## Params
### File
Lastsync file. Should exist.

### Port
Port to expose metrics at. Leave empty or 0 for disable metrics.

### Interval
Interval between checks.

## SystemD daemon example
```unit file (systemd)
[Unit]
Description=Sync Monitor and Metrics
Wants=network-online.target
After=network-online.target nss-lookup.target

[Service]
Type=simple
Restart=always
RestartSec=5s
ExecStart=/usr/local/bin/sync-monitor --file /path/to/repo/lastsync --port 1221 --interval 2m

[Install]
WantedBy=multi-user.target
```

## Build
### With make
```shell
make binary
```

### Without make
```shell
 CGO_ENABLED=0 go build github.com/kyberorg/sync-monitor/cmd/sync-monitor
```

## About: State monitoring
Application can monitor Manjaro's `state` files. 
It reads timestamp from repo's `state` file and calculates delta (seconds between now and timestamp).

### Usage
```shell
/usr/local/bin/sync-monitor
--manjaro.repo.path=/path/to/repo
--manjaro.repo.list=testing,stable,unstable
(other params are omiited)
```



