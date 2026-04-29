package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	socks5proxy "golang.org/x/net/proxy"
)

const (
	SCRAPE_WORKERS = 50
	CHECK_WORKERS  = 350
	CHECK_TIMEOUT  = 3 * time.Second
	RETRY_COUNT    = 1
)

type Proxy struct {
	IP   string
	Port string
}

var proxySources = []string{
	"https://raw.githubusercontent.com/Thordata/awesome-free-proxy-list/main/proxies/socks5.txt",
	"https://raw.githubusercontent.com/fyvri/fresh-proxy-list/archive/storage/classic/socks5.txt",
	"https://raw.githubusercontent.com/Ian-Lusule/Proxies/main/proxies/socks5.txt",
	"https://raw.githubusercontent.com/joy-deploy/free-proxy-list/main/data/latest/types/socks5/proxies.txt",
	"https://raw.githubusercontent.com/gfpcom/free-proxy-list/main/list/socks5.txt",
	"https://raw.githubusercontent.com/proxifly/free-proxy-list/refs/heads/main/proxies/protocols/socks5/data.txt",
	"https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/socks5.txt",
	"https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-socks5.txt",
	"https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS5_RAW.txt",
	"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-socks5.txt",
	"https://raw.githubusercontent.com/vakhov/fresh-proxy-list/refs/heads/master/socks5.txt",
	"https://raw.githubusercontent.com/Zaeem20/FREE_PROXIES_LIST/master/socks5.txt",
	"https://raw.githubusercontent.com/simatwa/free-proxies/main/socks5.txt",
	"https://raw.githubusercontent.com/zevtyardt/proxy-list/main/socks5.txt",
	"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/socks5.txt",
	"https://raw.githubusercontent.com/sunny9577/proxy-scraper/master/proxies/socks5.txt",
	"https://raw.githubusercontent.com/officialputuid/KangProxy/main/socks5.txt",
	"https://raw.githubusercontent.com/elli0t43/proxy-list/master/socks5.txt",
	"https://raw.githubusercontent.com/hookzof/socks5_list/master/txt/socks5.txt",
	"https://raw.githubusercontent.com/ALIILAPRO/Proxy/main/socks5.txt",
	"https://raw.githubusercontent.com/MuhamadZainal8/proxy-scraper/main/socks5.txt",
	"https://raw.githubusercontent.com/ZenulAbidin/proxy-list/main/socks5.txt",
	"https://raw.githubusercontent.com/B4RC0DE-TM/proxy-list/main/socks5.txt",
	"https://raw.githubusercontent.com/yemixzy/proxy-list/main/socks5.txt",
	"https://raw.githubusercontent.com/rxndydev/proxy-list/main/socks5.txt",
	"https://raw.githubusercontent.com/TheLonelyWolf/stuff/main/proxy/socks5.txt",
	"https://raw.githubusercontent.com/rdavydov/proxy-list/main/socks5.txt",
	"https://raw.githubusercontent.com/a2u/free-proxy-list/main/socks5.txt",
	"https://raw.githubusercontent.com/mmpx12/proxy-list/master/socks5.txt",
}

func main() {
	scraped := scrapeProxies()
	total := len(scraped)

	logo := fmt.Sprintf(`
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⡾⠃⠀⠀⠀⠀⠀⠀Runing : Termux
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⣀⣾⠋⠀⠀⠀⠀⠀⠀⠀⠀Server : None
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⠛⠻⢿⣷⣄⠀⠀⠀⠀⠀⠀⠀⠀Version : v1.3.0
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⡾⢛⣿⣿⣶⣄⠙⠿⠀⠀⠀⠀⠀⠀⠀⠀Connection : Wifi
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⡟⢀⣾⣿⣿⣿⣿⣷⡀⠀⠀⠀⠀⠀⠀⠀⠀Dns : dns.adguard.com
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡿⢀⣾⣿⣿⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀Requests By : Golang
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣸⡇⣼⣿⣿⣿⣿⣿⣿⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀Sumber : go.sum
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣿⣷⣿⣿⣿⣿⣿⡿⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Modifikasi : go.mod
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣼⣿⣿⣿⣿⣿⣿⠏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Country : Indonesia
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣾⣿⣿⣿⣿⡿⠟⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Device : Redmi-Xiaomi
⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣶⣿⣿⣿⣿⠿⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Cpu : 8 Core
⠀⠀⠀⠀⠀⠀⢀⣤⣾⣿⣿⣿⠿⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Ram : 8+8 gb
⠀⠀⠀⠀⣠⣶⡿⠿⠛⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀Rom : 256 gb

🚀 Faster Scrapers
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
+ Scrape  : %d
+ Type    : SOCKS5
+ Ulimit  : 32768
+ Layers  : 7 Support
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

`, total)
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
		saveToFile(activeList.proxies, "interrupted")
		activeList.mu.Unlock()
		os.Exit(0)
	}()

	checkProxies(activeList, scraped)

	if len(activeList.proxies) > 0 {
		saveToFile(activeList.proxies, "active")
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
		transport := &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
			DialContext: (&net.Dialer{
				Timeout:   CHECK_TIMEOUT,
				KeepAlive: 0,
			}).DialContext,
		}
		dialer, err := socks5proxy.SOCKS5("tcp", net.JoinHostPort(p.IP, p.Port), nil, socks5proxy.Direct)
		if err != nil {
			continue
		}
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
		transport.Proxy = nil
		client := &http.Client{Transport: transport, Timeout: CHECK_TIMEOUT}

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
			time.Sleep(100 * time.Millisecond)
		}
		if ok {
			fmt.Printf("%s:%s\n", p.IP, p.Port)
			active.mu.Lock()
			active.proxies = append(active.proxies, p)
			active.mu.Unlock()
		}
	}
}

func saveToFile(proxies []Proxy, typ string) {
	if len(proxies) == 0 {
		return
	}
	name := fmt.Sprintf("proxy.txt", typ, time.Now().Format("20060102_150405"))
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
	fmt.Printf("[+] Saved %d proxies to %s\n", len(proxies), name)
}
