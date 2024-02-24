# ADGuard Home Query Log Preprocessor

## Background

- Avoid duplicated disk write
- ADGuard Home `querylog.json` does not involve any sender or syslog protocol
- `Answer` section in ADGuard Home `querylog.json` are `github.com/miekg/dns.Msg` struct marshalled into base64 text, which is DNS message on wire, and also is difficult to search or extraction on Splunk.

## What will this app do?

1. 

## Configuration



## Installation

Download plugin from release. 

## License

GNU AGPL v3

## Assets

Lookup Tables:

https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml

- RCODE: https://www.iana.org/assignments/dns-parameters/dns-parameters-6.csv
- OpCode: https://www.iana.org/assignments/dns-parameters/dns-parameters-5.csv
