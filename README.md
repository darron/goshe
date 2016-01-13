goshe
===========

Replacement for some old Ruby scripts that look for Apache processes and send their memory usage to Datadog.

## apache

```
darron@: bin/goshe apache -h
Grab stats from Apache2 processes - and mod_status - and sends to Datadog.

Usage:
  goshe apache [flags]

Flags:
  -m, --memory uint   Smallest Apache memory size to log. (default 10485760)

Global Flags:
  -i, --interval int     Interval when running in a loop. (default 5)
      --prefix string    Metric prefix. (default "goshe")
  -p, --process string   Process name to match.
      --verbose          log output to stdout
```

## ping

```
darron@: bin/goshe ping -h
Ping an address and send stats to Datadog. Need to be root to use.

Usage:
  goshe ping [flags]

Flags:
  -e, --endpoint string   Endpoint to ping. (default "www.google.com")

Global Flags:
  -i, --interval int     Interval when running in a loop. (default 5)
      --prefix string    Metric prefix. (default "goshe")
  -p, --process string   Process name to match.
      --verbose          log output to stdout
```

## match

```
darron@: bin/goshe match -h
Grab memory stats from matching processes and sends to Datadog.

Usage:
  goshe match [flags]

Global Flags:
  -i, --interval int     Interval when running in a loop. (default 5)
      --prefix string    Metric prefix. (default "goshe")
  -p, --process string   Process name to match.
      --verbose          log output to stdout
```

These are much faster and use significantly less memory.

[![wercker status](https://app.wercker.com/status/f25e70250066e5f1e03744ef4d5be79e/m "wercker status")](https://app.wercker.com/project/bykey/f25e70250066e5f1e03744ef4d5be79e)
