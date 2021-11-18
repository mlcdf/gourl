package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

var tmpl = `{{.Method}} {{.URL}}
{{- range $name, $values := .Header }}
{{ $name | bold }}: {{ range $index, $value := $values -}} {{- $value }}{{ if last $index $values | not }}, {{ end }}{{ end -}}
{{ end }}
`

var once sync.Once

type Reporter interface {
	request(*http.Request) error
	result(*http.Response, *Trace) error
	// export() error
}

type Result struct {
	res   *http.Response
	trace *Trace
}

type SimpleReport struct {
	stdout io.Writer
	stderr io.Writer
}

type Report struct {
	SimpleReport
	results []Result
}

func newReporter(interval float64) Reporter {
	if interval > 0 {
		return &Report{
			SimpleReport: SimpleReport{stdout: os.Stdout, stderr: os.Stderr},
			results:      make([]Result, 0),
		}
	}

	return &SimpleReport{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (r *Report) result(res *http.Response, trace *Trace) error {
	// https://stackoverflow.com/questions/17948827/reusing-http-connections-in-go
	defer res.Body.Close()
	_, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	//

	r.results = append(r.results, Result{res, trace})

	once.Do(func() {
		fmt.Fprintln(r.stdout, "Code    DNS lookup    TLS handshake        Connect      Start to first byte          Total")
	})

	fmt.Fprintf(r.stdout, "%d      %9v        %9v      %9v                %9s      %9s\n",
		res.StatusCode, trace.DNSLookUp.Round(time.Millisecond),
		trace.TLSHandshake.Round(time.Millisecond),
		trace.Connect.Round(time.Millisecond),
		trace.FromStartToFirstByte.Round(time.Millisecond),
		trace.Total.Round(time.Millisecond))
	return nil
}

func (r *SimpleReport) result(res *http.Response, trace *Trace) error {

	fmt.Fprintln(r.stdout, "--------------------------------------------------------------------------------")
	fmt.Fprintln(r.stdout, trace)

	fmt.Fprintln(r.stdout, "--------------------------------------------------------------------------------")
	fmt.Fprintf(r.stdout, "%s\n", res.Status)

	if isVerbose {
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		fmt.Fprintf(r.stdout, "%s\n", body)
	}
	return nil
}

func stdoutIsTerm() bool {
	_, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TCGETS)
	return err == nil
}

var fns = template.FuncMap{
	"last": func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()-1
	},

	"bold": func(x string) string {
		if !stdoutIsTerm() {
			return x
		}
		return "\033[1m" + x + "\033[0m"
	},
}

func (r *SimpleReport) request(req *http.Request) error {
	tmpl, err := template.New("request").Funcs(fns).Parse(tmpl)
	if err != nil {
		return err
	}

	return tmpl.Execute(r.stderr, req)
}

func (r *Report) request(req *http.Request) error {
	tmpl, err := template.New("request").Funcs(fns).Parse(tmpl)
	if err != nil {
		return err
	}

	return tmpl.Execute(r.stderr, req)
}
