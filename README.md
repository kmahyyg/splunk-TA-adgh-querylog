# ADGuard Home Query Log Preprocessor

## Background

- Avoid duplicated disk write
- ADGuard Home `querylog.json` does not involve any sender or syslog protocol
- `Answer` section in ADGuard Home `querylog.json` are `github.com/miekg/dns.Msg` struct marshalled into base64 text, which is difficult to search or extraction on Splunk.

## What will I do?

1. Read `querylog.json`, then record its last record timestamp before each time when program exit.
2. FULLY REWRITE - TODO

## Configuration

No need at all, since it's now a Splunk Addon.

## Installation

Download plugin from release. 

## License

GNU AGPL v3

## Assets

Lookup Tables:

https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml

- RCODE: https://www.iana.org/assignments/dns-parameters/dns-parameters-6.csv
- OpCode: https://www.iana.org/assignments/dns-parameters/dns-parameters-5.csv
