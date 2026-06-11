use hyper::client::HttpConnector;
use hyper::Client;
use hyper_rustls::{HttpsConnector, HttpsConnectorBuilder};
use std::sync::Arc;
use std::time::Duration;
use tokio::net::TcpStream;
use tokio::signal;
use tokio::sync::watch;
use tokio::time::timeout;
use std::env;
use std::process;

#[tokio::main]
async fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() < 2 {
        eprintln!("Usage: {} <url>", args[0]);
        process::exit(1);
    }
    let target_url = args[1].clone();

    let mut http = HttpConnector::new();
    http.set_connect_timeout(Some(Duration::from_secs(2)));
    http.set_nodelay(true);
    http.set_keepalive(Some(Duration::from_secs(30)));

    let tls = HttpsConnectorBuilder::new()
        .with_native_roots()
        .unwrap_or_else(|_| HttpsConnectorBuilder::new().with_webpki_roots())
        .https_only()
        .enable_http2()
        .with_connector(http);

    let client = Client::builder()
        .pool_idle_timeout(Duration::from_secs(60))
        .pool_max_idle_per_host(5000)
        .http2_only(true)
        .build::<_, hyper::Body>(tls);

    let client = Arc::new(client);

    let workers = 550;
    let (tx, mut rx) = watch::channel(false);
    let mut handles = vec![];

    for _ in 0..workers {
        let client_clone = Arc::clone(&client);
        let url_clone = target_url.clone();
        let mut rx_clone = rx.clone();

        let handle = tokio::spawn(async move {
            let req = hyper::Request::head(&url_clone)
                .header("user-agent", "curl/8.4.0")
                .body(hyper::Body::empty())
                .unwrap();

            loop {
                if *rx_clone.borrow() {
                    break;
                }
                let _ = timeout(Duration::from_secs(3), client_clone.request(req.clone())).await;
            }
        });
        handles.push(handle);
    }

    tokio::select! {
        _ = signal::ctrl_c() => {
            println!("\n[*] Stopping workers...");
        }
    }

    let _ = tx.send(true);
    for handle in handles {
        let _ = handle.await;
    }
}
