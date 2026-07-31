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
	SCRAPE_WORKERS = 7500
	CHECK_WORKERS  = 7500
	CHECK_TIMEOUT  = 4 * time.Second
	RETRY_COUNT    = 1
	TARGET         = 2000
)

type Proxy struct {
	IP   string
	Port string
}

var (
	activeList    = &safeList{proxies: []Proxy{}}
	seenProxies   = &safeSet{items: make(map[string]bool)} // track proxy yang sudah diperiksa
	proxySources  = []string{
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
)

type safeList struct {
	mu      sync.Mutex
	proxies []Proxy
}

type safeSet struct {
	mu    sync.Mutex
	items map[string]bool
}

func (s *safeSet) Add(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = true
}

func (s *safeSet) Has(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.items[key]
}

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTSTP)
	go func() {
		<-sigCh
		fmt.Println("\nSaving active proxies...")
		activeList.mu.Lock()
		if len(activeList.proxies) > 0 {
			saveToFile(activeList.proxies)
			fmt.Printf("Saved %d proxies.\n", len(activeList.proxies))
		}
		activeList.mu.Unlock()
		os.Exit(0)
	}()

	// Load existing proxy.txt, tandai sebagai sudah dilihat, dan periksa ulang
	if _, err := os.Stat("proxy.txt"); err == nil {
		if proxies, err := loadProxiesFromFile("proxy.txt"); err == nil && len(proxies) > 0 {
			for _, p := range proxies {
				seenProxies.Add(p.IP + ":" + p.Port)
			}
			checkProxies(proxies)
		}
	}

	// Loop utama: scrape dan check sampai target tercapai atau kehabisan proxy baru
	for len(activeList.proxies) < TARGET {
		fmt.Printf("Active: %d, scraping new proxies...\n", len(activeList.proxies))

		scraped := scrapeProxies()
		var newProxies []Proxy
		for _, p := range scraped {
			key := p.IP + ":" + p.Port
			if !seenProxies.Has(key) {
				newProxies = append(newProxies, p)
				seenProxies.Add(key)
			}
		}

		if len(newProxies) == 0 {
			fmt.Println("Tidak ada proxy baru dari sumber. Berhenti.")
			break
		}

		fmt.Printf("Memeriksa %d proxy baru...\n", len(newProxies))
		checkProxies(newProxies)
		// TIDAK ADA JEDA, LANGSUNG LOOP LAGI
	}

	// Simpan hasil akhir
	activeList.mu.Lock()
	if len(activeList.proxies) > 0 {
		unique := make(map[string]Proxy)
		for _, p := range activeList.proxies {
			unique[p.IP+":"+p.Port] = p
		}
		final := make([]Proxy, 0, len(unique))
		for _, p := range unique {
			final = append(final, p)
		}
		saveToFile(final)
		fmt.Printf("Saved %d active proxies.\n", len(final))
	} else {
		fmt.Println("Tidak ada proxy aktif ditemukan.")
	}
	activeList.mu.Unlock()
}

// ===== FUNGSI LAIN (TIDAK DIUBAH) =====

func loadProxiesFromFile(name string) ([]Proxy, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	var res []Proxy
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		ip, port := parseLine(line)
		if ip != "" && port != "" {
			res = append(res, Proxy{ip, port})
		}
	}
	return res, s.Err()
}

func scrapeProxies() []Proxy {
	chSrc := make(chan string, len(proxySources))
	chRes := make(chan Proxy, 10000)
	var wg sync.WaitGroup
	for i := 0; i < SCRAPE_WORKERS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{
				Timeout: 5 * time.Second,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
			for url := range chSrc {
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
						chRes <- Proxy{ip, port}
					}
				}
			}
		}()
	}
	for _, src := range proxySources {
		chSrc <- src
	}
	close(chSrc)
	go func() {
		wg.Wait()
		close(chRes)
	}()
	m := make(map[string]Proxy)
	for p := range chRes {
		m[p.IP+":"+p.Port] = p
	}
	out := make([]Proxy, 0, len(m))
	for _, p := range m {
		out = append(out, p)
	}
	return out
}

func parseLine(line string) (ip, port string) {
	if i := strings.Index(line, "://"); i != -1 {
		line = line[i+3:]
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

func checkProxies(proxies []Proxy) {
	input := make(chan Proxy, 5000)
	var wg sync.WaitGroup
	for i := 0; i < CHECK_WORKERS; i++ {
		wg.Add(1)
		go func() {
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
					activeList.mu.Lock()
					activeList.proxies = append(activeList.proxies, p)
					activeList.mu.Unlock()
				}
			}
		}()
	}
	for _, p := range proxies {
		input <- p
	}
	close(input)
	wg.Wait()
}

func saveToFile(proxies []Proxy) {
	if len(proxies) == 0 {
		return
	}
	f, err := os.Create("proxy.txt")
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
