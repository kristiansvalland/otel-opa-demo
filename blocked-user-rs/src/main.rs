use actix_web::middleware::Logger;
use actix_web::{web, App, HttpServer, Result};
use actix_web_opentelemetry::RequestTracing;
use log::info;
use opentelemetry::sdk::propagation::TraceContextPropagator;
use opentelemetry::trace::TraceContextExt;
use opentelemetry::trace::Tracer;
use opentelemetry::Context;
use opentelemetry::{
    global,
    sdk::trace::{self, IdGenerator, Sampler},
    sdk::Resource,
    trace::TraceError,
    Key,
};
use opentelemetry_otlp::WithExportConfig;
use std::time::Duration;

use serde::{Deserialize, Serialize};

fn setup_tracer() -> Result<trace::Tracer, TraceError> {
    info!("Setting up tracer");
    // Create a new trace pipeline that prints to stdout
    global::set_text_map_propagator(TraceContextPropagator::new());
    let exporter = opentelemetry_otlp::new_exporter()
        .tonic()
        .with_endpoint("http://otel-collector:4317")
        .with_timeout(Duration::from_secs(5));
    let trace_config = trace::config()
        .with_sampler(Sampler::AlwaysOn)
        .with_id_generator(IdGenerator::default())
        .with_max_events_per_span(16)
        .with_resource(Resource::default());
    opentelemetry_otlp::new_pipeline()
        .tracing()
        .with_exporter(exporter)
        .with_trace_config(trace_config)
        .install_batch(opentelemetry::runtime::Tokio)
}

#[derive(Deserialize)]
struct BlockedRequest {
    username: String,
}

#[derive(Serialize)]
struct BlockedResponse {
    result: bool,
}

async fn blocked_user(user_info: web::Json<BlockedRequest>) -> Result<web::Json<BlockedResponse>> {
    let tracer = global::tracer("blocked-user-rs");
    tracer.in_span("Request blocked user", |ctx: Context| {
        // TODO
        let span = ctx.span();
        span.set_attribute(Key::new("username").string(user_info.username.clone()));
        let is_blocked = match user_info.username.as_str() {
            "mallory" => {
                span.add_event(
                    "User is blocked",
                    vec![Key::new("username").string(user_info.username.clone())],
                );
                true
            }
            _ => {
                span.add_event(
                    "User is allowed",
                    vec![Key::new("username").string(user_info.username.clone())],
                );
                false
            }
        };
        Ok(web::Json(BlockedResponse { result: is_blocked }))
    })
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    std::env::set_var("RUST_LOG", "debug");
    info!("Starting app.");
    env_logger::init();
    let _ = setup_tracer().expect("Failed to initialise tracer.");
    info!("Tracer set up.");

    info!("Starting server");
    HttpServer::new(move || {
        App::new()
            .wrap(Logger::default())
            .wrap(RequestTracing::new())
            .route("/blocked", web::post().to(blocked_user))
    })
    .bind("0.0.0.0:8088")
    .unwrap()
    .run()
    .await?;

    info!("Shutting down ...");
    Ok(())
}
