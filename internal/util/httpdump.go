package util

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

type DumpTransport struct {
	R http.RoundTripper
}

func (d *DumpTransport) RoundTrip(h *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequestOut(h, true)
	fmt.Printf("****REQUEST****\n%q\n", dump)
	resp, err := d.R.RoundTrip(h)
	dump, _ = httputil.DumpResponse(resp, true)
	fmt.Printf("****RESPONSE****\n%q\n****************\n\n", dump)
	return resp, err
}
