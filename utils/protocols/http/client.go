package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

var client http.Client

type config struct {
	TotalCount   int
	CountPerHost int
	Timeout      time.Duration
}

func Init(config config) {
	client = http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        config.TotalCount,
			MaxConnsPerHost:     config.CountPerHost,
			MaxIdleConnsPerHost: config.CountPerHost,
			IdleConnTimeout:     config.Timeout,
			DialContext: (&net.Dialer{
				Timeout: config.Timeout,
			}).DialContext,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: config.Timeout,
	}
}
