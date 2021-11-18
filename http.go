package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Header struct {
	Key   string
	Value string
}

type Trace struct {
	client               *httptrace.ClientTrace
	DNSLookUp            time.Duration
	TLSHandshake         time.Duration
	Connect              time.Duration
	FromStartToFirstByte time.Duration
	Total                time.Duration
}

func (t *Trace) String() string {
	buf := bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("DNS Done:            %9v\n", t.DNSLookUp.Round(time.Millisecond)))
	buf.WriteString(fmt.Sprintf("TLS Handshake:       %9v\n", t.TLSHandshake.Round(time.Millisecond)))
	buf.WriteString(fmt.Sprintf("Connect time:        %9v\n", t.Connect.Round(time.Millisecond)))
	buf.WriteString(fmt.Sprintf("Start to first byte: %9v\n", t.FromStartToFirstByte.Round(time.Millisecond)))
	buf.WriteString(fmt.Sprintf("Total time:          %9v", t.Total.Round(time.Millisecond)))

	return buf.String()
}

func newTrace() *Trace {
	var start, connect, dns, tlsHandshake time.Time
	trace := &Trace{}

	trace.client = &httptrace.ClientTrace{
		GetConn:  func(hostPort string) { start = time.Now() },
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			// fmt.Printf("DNS Done: %v\n", time.Since(dns))
			trace.DNSLookUp = time.Since(dns)
		},

		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			// fmt.Printf("TLS Handshake: %v\n", time.Since(tlsHandshake))
			trace.TLSHandshake = time.Since(tlsHandshake)
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			// fmt.Printf("Connect time: %v\n", time.Since(connect))
			trace.Connect = time.Since(connect)
		},

		GotFirstResponseByte: func() {
			// fmt.Printf("Time from start to first byte: %v\n", time.Since(start))
			trace.FromStartToFirstByte = time.Since(start)
		},
	}
	return trace
}

func prepareRequest(u string, method string, headers http.Header, data []string) (*http.Request, error) {
	if len(data) != 0 {
		method = http.MethodPost
	}

	if headers == nil {
		headers = make(http.Header)
	}

	var body io.Reader
	if len(data) > 0 {
		body = strings.NewReader(strings.Join(data, "&"))
	}

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}

	if headers.Get("Content-type") == "" && len(data) > 0 {
		headers.Add("Content-type", "application/x-www-form-urlencoded")
	}

	if headers.Get("User-Agent") == "" {
		headers.Add("User-Agent", "gourl (https://github.com/mlcdf/gourl)")
	}

	req.Header = headers

	return req, nil
}

func newRoundTripper(noConnectionReuse bool) http.RoundTripper {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DisableKeepAlives = noConnectionReuse
	return t
}

func request(req *http.Request, rt http.RoundTripper) (*http.Response, *Trace, error) {
	trace := newTrace()
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace.client))

	start := time.Now()

	res, err := rt.RoundTrip(req)
	if err != nil {
		return nil, trace, err
	}

	trace.Total = time.Since(start)
	return res, trace, nil
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	clone := req.Clone(req.Context())
	var err error
	clone.Body, err = req.GetBody()
	if err != nil {
		return nil, errors.Wrap(err, "failed to clone the request body")
	}
	return clone, nil
}
