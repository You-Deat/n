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
	scraped := scrapeProxies()
	total := len(scraped)

	logo := fmt.Sprintf(`
🚀 Faster HTTP Scraper
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
+ Scrape  : %d
+ Type    : HTTP
+ Workers : %d
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`, total, SCRAPE_WORKERS)
	fmt.Print(logo)

	if total == 0 {
		return
	}

	activeList := &safeList{proxies: []Proxy{}}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTSTP)
	go func() {
		<-sigCh
		activeList.mu.Lock()
		saveToFile(activeList.proxies)
		activeList.mu.Unlock()
		os.Exit(0)
	}()

	checkProxies(activeList, scraped)

	if len(activeList.proxies) > 0 {
		saveToFile(activeList.proxies)
	}
}

type safeList struct {
	mu      sync.Mutex
	proxies []Proxy
}

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
	fmt.Println("[+] Saved proxy.txt")
}
