# gourl

## Features
- Performs simple HTTP requests and gathers stats
- Provides an API as close to cURL as possible.
- Distributed as a static binary. Works great in containers.

## Usage

```
Performs simple HTTP requests and gathers stats while providing an API as close
to cURL as possible.

Usage:
    gourl [options...] <url>

Options:
    -X, --request          HTTP method
    -H, --header           Pass custom header(s)
    -d, --data             Request data/payload
    -V, --version          Print version
    -v, --verbose          Verbose terminal output
    -h, --help             Show info usage

    --interval             Interval between requests
    --no-connection-reuse  Turn off HTTP connection reuse

Examples:
    gourl https://httpbin.org -d "yolo"
    gourl https://httpbin.org -H "Authorization: ${token}"
```

## To do
- better error handling
- tests
- export data as csv
- Fancy colored terminal output