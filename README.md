# ADGuard Home Query Log Preprocessor

## Background

- Avoid duplicated disk write
- ADGuard Home `querylog.json` does not involve any sender or syslog protocol
- `Answer` section in ADGuard Home `querylog.json` are `github.com/miekg/dns.Msg` struct marshalled into base64 text,
which is difficult to search or extraction on Splunk.

## What will I do?

1. Read `querylog.json`, then record its last record timestamp before each time when program exit.
2. Unmarshal `Answer` section, if `request type` is `TXT` or " `CNAME` but cannot find `A` ", will record CNAME and TXT data, else record IPs.
3. Final Parsed Answer will be in `"ParsedAnswer": ["IP", "TEXT"]`
4. Send to Splunk TCP input. (No TLS support due to internal network, it's for my home lab. I'm lazy.)

## Configuration

Configuration was set via environment variable and will be saved to `/etc/adgh-log-preproc/default`: 
```
DEST_TCP=10.77.1.233:12253
SRC_LOG=/usr/local/adgh/data/querylog.json
```

Progress recover file will be saved in `/etc/adgh-log-preproc/.recover.prog`.

## License

GNU AGPL v3
