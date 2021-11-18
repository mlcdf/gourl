package main

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func Test(t *testing.T) {
	req, err := http.NewRequest("GET", "https://mlcdf.fr", nil)
	req.Header.Add("Access-Control-Allow-Headers", "content-type")
	req.Header.Add("Access-Control-Allow-Headers", "authorization")
	req.Header.Add("Authorization", "Bearer xxx")

	if err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	reporter := SimpleReport{
		stdout: io.Discard,
		stderr: buf,
	}
	err = reporter.request(req)
	if err != nil {
		t.Fatal(err)
	}

	txt := buf.String()
	expected := `GET https://mlcdf.fr
Access-Control-Allow-Headers: content-type, authorization
Authorization: Bearer xxx
`
	if txt != expected {
		t.Errorf("'%s' != '%s'", txt, expected)
	}
}
