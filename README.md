# ADGuard Home Query Log Preprocessor

## Background

- Avoid duplicated disk write
- ADGuard Home `querylog.json` does not involve any sender or syslog protocol
- `Answer` section in ADGuard Home `querylog.json` are `github.com/miekg/dns.Msg` struct marshalled into base64 text, which is difficult to search or extraction on Splunk.

## What will I do?

1. Read `querylog.json`, then record its last record timestamp before each time when program exit.
2. Unmarshal `Answer` section, if `request type` is selected, will record target domains/IPs/texts for specific types of requests, check `ext/parse_adgh_answer.go` for more details.
3. Final Parsed Answer will be in `"ParsedAnswer": ["IP", "TEXT"]`
4. Send to Splunk TCP input. (No TLS support due to internal network, it's for my home lab. I'm lazy.)
5. Support TCP auto-reconnect and progress save and recovery based on timestamp of raw log.

## Configuration

Configuration was set via environment variable and will be saved to `/etc/adgh-log-preproc/default`: 
```
DEST_TCP=10.77.1.233:12253
SRC_LOG=/usr/local/adgh/data/querylog.json
```

Progress recover file will be saved in `/etc/adgh-log-preproc/.recover.prog`.

## Installation

Download binary from release. Write environment variable according to above. Put `.service` file in the same folder, enable service and run.

## License

GNU AGPL v3

## Assets

Lookup Tables:

https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml

- RCODE: https://www.iana.org/assignments/dns-parameters/dns-parameters-6.csv
- OpCode: https://www.iana.org/assignments/dns-parameters/dns-parameters-5.csv
