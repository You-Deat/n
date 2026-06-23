package main

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
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	SCRAPE_WORKERS = 1500
	CHECK_WORKERS  = 1500
	CHECK_TIMEOUT  = 3 * time.Second
	RETRY_COUNT    = 1
)

type Proxy struct {
	IP   string
	Port string
}

var proxySources = []string{
	"https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt",
	"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/http.txt",
	"https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-http.txt",
	"https://raw.githubusercontent.com/roosterkid/openproxylist/main/HTTP_RAW.txt",
	"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/http.txt",
	"https://raw.githubusercontent.com/proxifly/free-proxy-list/main/proxies/protocols/http/data.txt",
	"https://raw.githubusercontent.com/joy-deploy/free-proxy-list/main/data/latest/types/http/proxies.txt",
	"https://raw.githubusercontent.com/fyvri/fresh-proxy-list/archive/storage/classic/http.txt",
	"https://raw.githubusercontent.com/sunny9577/proxy-scraper/master/proxies/http.txt",
	"https://raw.githubusercontent.com/zevtyardt/proxy-list/main/http.txt",
	"https://raw.githubusercontent.com/ALIILAPRO/Proxy/main/http.txt",
	"https://raw.githubusercontent.com/B4RC0DE-TM/proxy-list/main/http.txt",
	"https://raw.githubusercontent.com/elli0t43/proxy-list/master/http.txt",
	"https://raw.githubusercontent.com/hookzof/socks5_list/master/txt/http.txt",
	"https://raw.githubusercontent.com/a2u/free-proxy-list/main/http.txt",
	"https://raw.githubusercontent.com/mmpx12/proxy-list/master/http.txt",
}

func main() {
	// Siapkan tempat menampung semua proxy aktif
	var allActive []Proxy

	// Cek file proxy.txt
	if _, err := os.Stat("proxy.txt"); err == nil {
		loaded, err := loadProxiesFromFile("proxy.txt")
		if err == nil && len(loaded) > 0 {
			fmt.Printf("📂 Loaded %d proxies from proxy.txt, checking...\n", len(loaded))
			activeList := &safeList{proxies: []Proxy{}}
			checkProxies(activeList, loaded)
			allActive = append(allActive, activeList.proxies...)
			fmt.Printf("✅ Found %d active from file\n", len(activeList.proxies))
		} else {
			fmt.Println("proxy.txt is empty or invalid, will scrape from GitHub.")
		}
	} else {
		fmt.Println("proxy.txt not found, scraping from GitHub...")
	}

	// Scrape dari GitHub
	fmt.Println("🌐 Scraping proxies from GitHub...")
	scraped := scrapeProxies()
	fmt.Printf("📥 Scraped %d proxies\n", len(scraped))

	if len(scraped) > 0 {
		fmt.Println("🔍 Checking scraped proxies...")
		activeList2 := &safeList{proxies: []Proxy{}}
		checkProxies(activeList2, scraped)
		allActive = append(allActive, activeList2.proxies...)
		fmt.Printf("✅ Found %d active from scraped sources\n", len(activeList2.proxies))
	}

	// Gabungkan dan hilangkan duplikat
	unique := make(map[string]Proxy)
	for _, p := range allActive {
		key := p.IP + ":" + p.Port
		unique[key] = p
	}
	finalProxies := make([]Proxy, 0, len(unique))
	for _, p := range unique {
		finalProxies = append(finalProxies, p)
	}

	if len(finalProxies) > 0 {
		saveToFile(finalProxies)
		fmt.Printf("💾 Saved %d active proxies to proxy.txt\n", len(finalProxies))
	} else {
		fmt.Println("⚠️ No active proxies found.")
	}
}

// ========== Fungsi pendukung ==========

type safeList struct {
	mu      sync.Mutex
	proxies []Proxy
}

// loadProxiesFromFile membaca proxy dari file teks (satu baris ip:port)
func loadProxiesFromFile(filename string) ([]Proxy, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var proxies []Proxy
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		ip, port := parseLine(line)
		if ip != "" && port != "" {
			proxies = append(proxies, Proxy{IP: ip, Port: port})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return proxies, nil
}

// scrapeProxies mengambil proxy dari semua sumber GitHub
func scrapeProxies() []Proxy {
	sourceCh := make(chan string, len(proxySources))
	resultCh := make(chan Proxy, 10000)
	var wg sync.WaitGroup

	for i := 0; i < SCRAPE_WORKERS; i++ {
		wg.Add(1)
		go scrapeWorker(sourceCh, resultCh, &wg)
	}
	for _, src := range proxySources {
		sourceCh <- src
	}
	close(sourceCh)
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	m := make(map[string]Proxy)
	for r := range resultCh {
		key := r.IP + ":" + r.Port
		if _, ok := m[key]; !ok {
			m[key] = r
		}
	}
	out := make([]Proxy, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func scrapeWorker(srcCh <-chan string, out chan<- Proxy, wg *sync.WaitGroup) {
	defer wg.Done()
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxIdleConnsPerHost: 1000,
		},
	}
	for url := range srcCh {
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		scanner := bufio.NewScanner(strings.NewReader(string(body)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			ip, port := parseLine(line)
			if ip != "" && port != "" {
				out <- Proxy{IP: ip, Port: port}
			}
		}
	}
}

func parseLine(line string) (ip, port string) {
	if idx := strings.Index(line, "://"); idx != -1 {
		line = line[idx+3:]
	}
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}
	ip = parts[0]
	port = strings.SplitN(parts[1], " ", 2)[0]
	port = strings.SplitN(port, "/", 2)[0]
	if net.ParseIP(ip) == nil && !strings.Contains(ip, ".") {
		return "", ""
	}
	return ip, port
}

// checkProxies memeriksa daftar proxy dan menambahkan yang aktif ke activeList
func checkProxies(active *safeList, proxies []Proxy) {
	input := make(chan Proxy, 5000)
	var wg sync.WaitGroup

	for i := 0; i < CHECK_WORKERS; i++ {
		wg.Add(1)
		go worker(input, active, &wg)
	}
	for _, p := range proxies {
		input <- p
	}
	close(input)
	wg.Wait()
}

func worker(input <-chan Proxy, active *safeList, wg *sync.WaitGroup) {
	defer wg.Done()
	for p := range input {
		proxyURL, _ := url.Parse("http://" + p.IP + ":" + p.Port)
		transport := &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
			DialContext: (&net.Dialer{
				Timeout:   CHECK_TIMEOUT,
				KeepAlive: 0,
			}).DialContext,
			Proxy: http.ProxyURL(proxyURL),
		}
		client := &http.Client{
			Transport: transport,
			Timeout:   CHECK_TIMEOUT,
		}

		ok := false
		for retry := 0; retry <= RETRY_COUNT; retry++ {
			ctx, cancel := context.WithTimeout(context.Background(), CHECK_TIMEOUT)
			req, _ := http.NewRequestWithContext(ctx, "GET", "http://clients3.google.com/generate_204", nil)
			resp, err := client.Do(req)
			cancel()
			if err == nil && resp.StatusCode == 204 {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				ok = true
				break
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
		if ok {
			fmt.Printf("%s:%s\n", p.IP, p.Port)
			active.mu.Lock()
			active.proxies = append(active.proxies, p)
			active.mu.Unlock()
		}
	}
}

func saveToFile(proxies []Proxy) {
	if len(proxies) == 0 {
		return
	}
	name := "proxy.txt"
	f, err := os.Create(name)
	if err != nil {
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, p := range proxies {
		w.WriteString(p.IP + ":" + p.Port + "\n")
	}
	w.Flush()
}
