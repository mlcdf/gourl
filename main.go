package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const usage = `
Performs (multiple) HTTP requests and gathers stats while providing an API as close
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
`

// Version is set at build time using -ldflags="-X 'main.Version=v1.0.0'
var Version = "devel"
var isVerbose bool

type headerFlag struct {
	http.Header
}

func (f headerFlag) String() string {
	return ""
}

func (f headerFlag) Set(value string) error {
	splitted := strings.SplitN(value, ":", 2)
	f.Header[splitted[0]] = []string{strings.TrimSpace(splitted[1])}
	return nil
}

func newHeaderFlag() headerFlag {
	return headerFlag{
		Header: make(map[string][]string),
	}
}

type dataFlag []string

func (f *dataFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func (f *dataFlag) String() string {
	return "n"
}

func main() {
	log.SetFlags(0)
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	var method = http.MethodGet
	flag.StringVar(&method, "X", method, "HTTP Method")
	flag.StringVar(&method, "request", method, "HTTP Method")

	headerFlag := newHeaderFlag()
	flag.Var(&headerFlag, "H", "Pass custom header(s)")
	flag.Var(&headerFlag, "header", "Pass custom header(s)")

	var data dataFlag
	flag.Var(&data, "data", "Request data/payload")
	flag.Var(&data, "d", "Request data/payload")

	var interval float64
	flag.Float64Var(&interval, "interval", interval, "Interval between requests")

	var noConnectionReuse bool
	flag.BoolVar(&noConnectionReuse, "no-connection-reuse", noConnectionReuse, "Turn off HTTP connection reuse")

	flag.BoolVar(&isVerbose, "verbose", isVerbose, "Verbose terminal output")
	flag.BoolVar(&isVerbose, "v", isVerbose, "Verbose terminal output")

	var versionFlag bool
	flag.BoolVar(&versionFlag, "version", versionFlag, "print the version")
	flag.BoolVar(&versionFlag, "V", versionFlag, "print the version")

	flag.CommandLine.Parse(os.Args[firstArgWithDash(os.Args):])

	if versionFlag {
		fmt.Printf("gourl (%s)\n", Version)
		return
	}

	var url string
	if !strings.HasPrefix(os.Args[1], "-") {
		url = os.Args[1]
	} else {
		url = flag.Arg(0)
	}

	roundTripper := newRoundTripper(noConnectionReuse)
	reporter := newReporter(interval)

	req, err := prepareRequest(url, method, headerFlag.Header, data)
	if err != nil {
		log.Fatalln(err)
	}

	if isVerbose {
		reporter.request(req)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if interval == 0 {
		runSimple(req, roundTripper, reporter)
	} else {
		run(req, roundTripper, reporter, interval)
	}
}

func runSimple(req *http.Request, roundTripper http.RoundTripper, reporter Reporter) {
	res, trace, err := request(req, roundTripper)
	if err != nil {
		log.Fatalln(err)
	}

	err = reporter.result(res, trace)
	if err != nil {
		log.Fatalln(err)
	}
}

func run(req *http.Request, roundTripper http.RoundTripper, reporter Reporter, interval float64) {
	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for {
			req, err := cloneRequest(req)
			if err != nil {
				log.Fatalf("failed to clone request: %v", err)
			}

			res, trace, err := request(req, roundTripper)
			if err != nil {
				log.Fatalf("failed to perform request: %v", err)
			}

			err = reporter.result(res, trace)
			if err != nil {
				log.Fatalln(err)
			}

			time.Sleep(time.Second * time.Duration(interval))
		}
	}()
	<-cancelChan
}

// If a commandline app works like this: ./app subcommand -flag -flag2
// `flag.Parse` won't parse anything after `subcommand`.
// To still be able to use `flag.String/flag.Int64` etc without creating
// a new `flag.FlagSet`, we need this hack to find the first arg that has a dash
// so we know when to start parsing
func firstArgWithDash(args []string) int {
	index := 1
	for i := 1; i < len(args); i++ {
		index = i

		if len(args[i]) > 0 && args[i][0] == '-' {
			break
		}
	}
	return index
}
