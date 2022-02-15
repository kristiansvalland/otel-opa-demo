FROM rust:1.58.1 as builder

WORKDIR /usr/src/app

COPY Cargo.lock Cargo.toml .
RUN mkdir src && touch src/main.rs
RUN cargo fetch
COPY src src
RUN cargo install --path .

FROM debian:buster-slim
#RUN apt-get update && apt-g&& rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/local/cargo/bin/blocked-user-rs /usr/local/bin/blocked-user-rs
EXPOSE 8088
CMD ["blocked-user-rs"]