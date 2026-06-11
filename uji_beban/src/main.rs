use reqwest::{Client, Method};
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
    let target_url = &args[1];

    let client = Client::builder()
        .danger_accept_invalid_certs(true)
        .danger_accept_invalid_hostnames(true)
        .http2_prior_knowledge()
        .pool_max_idle_per_host(5000)
        .pool_idle_timeout(Duration::from_secs(60))
        .timeout(Duration::from_secs(3))
        .build()
        .unwrap();

    let client = Arc::new(client);
    let url = Arc::new(target_url.clone());

    const WORKERS: usize = 550;
    let (tx, mut rx) = watch::channel(false);
    let mut handles = vec![];

    println!("[*] Memulai pengujian dengan {} worker berbasis Rust...", WORKERS);

    for _ in 0..WORKERS {
        let client_clone = Arc::clone(&client);
        let url_clone = Arc::clone(&url);
        let mut rx_clone = rx.clone();

        let handle = tokio::spawn(async move {
            let user_agent = "curl/8.4.0";

            loop {
                if *rx_clone.borrow() {
                    break;
                }

                let req = client_clone
                    .request(Method::HEAD, url_clone.as_str())
                    .header("User-Agent", user_agent)
                    .build();

                if let Ok(request) = req {
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
