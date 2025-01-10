# Rust Build
FROM --platform=$BUILDPLATFORM rust:1.84 AS rust-build-stage
WORKDIR /yggdrasil
ARG TARGETPLATFORM
RUN git clone --depth 1 --branch resolve_all https://github.com/Iandenh/yggdrasil.git .
RUN case "$TARGETPLATFORM" in \
  "linux/arm64") echo aarch64-unknown-linux-gnu > /rust_target.txt ;; \
  "linux/amd64") echo x86_64-unknown-linux-gnu > /rust_target.txt ;; \
  *) exit 1 ;; \
esac
RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
        apt-get update && apt-get install -y gcc-aarch64-linux-gnu && mkdir -p .cargo && echo '[target.aarch64-unknown-linux-gnu]' >> .cargo/config && echo 'linker = "aarch64-linux-gnu-gcc"' >> .cargo/config ; \
fi
RUN rustup toolchain install stable --profile default
RUN rustup target add $(cat /rust_target.txt)
RUN cargo build --release --target $(cat /rust_target.txt)
RUN cp target/$(cat /rust_target.txt)/release/libyggdrasilffi.so libyggdrasilffi.so

# Go Build
FROM --platform=$BUILDPLATFORM golang:1.23 AS build-stage
ARG VERSION
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
        apt-get update && apt-get install -y gcc-aarch64-linux-gnu ; \
fi
WORKDIR /app
COPY --from=rust-build-stage /yggdrasil/libyggdrasilffi.so /app/unleashengine/libyggdrasilffi.so
RUN go install github.com/a-h/templ/cmd/templ@latest
COPY go.mod go.sum ./
RUN go mod download
RUN mkdir /data
COPY . /app
RUN templ generate
RUN case "$TARGETPLATFORM" in \
  "linux/arm64") CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc CC_FOR_TARGET=aarch64-linux-gnu-gcc GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-extld=aarch64-linux-gnu-gcc -s -w -X github.com/Iandenh/overleash/internal/version.Version=${VERSION}" -o /entrypoint main.go ;; \
  "linux/amd64") CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build  -ldflags="-s -w -X github.com/Iandenh/overleash/internal/version.Version=${VERSION}" -o /entrypoint main.go ;; \
  *) exit 1 ;; \
esac

# Deploy.
FROM --platform=$BUILDPLATFORM debian:bookworm-slim AS release-stage
ARG TARGETPLATFORM
ENV OVERLEASH_PORT=8080

LABEL org.opencontainers.image.title="Overleash"
LABEL org.opencontainers.image.description="Override your Unleash feature flags blazing fast"
LABEL org.opencontainers.image.source="https://github.com/Iandenh/overleash"
LABEL org.opencontainers.image.licenses="MIT"

RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
       dpkg --add-architecture arm64 && apt-get update && apt-get install -y gcc-aarch64-linux-gnu libc6:arm64 ; \
fi
WORKDIR /
RUN useradd -ms /bin/sh -u 1001 1001
USER 1001
COPY --from=build-stage /entrypoint /entrypoint
COPY --from=rust-build-stage /yggdrasil/libyggdrasilffi.so /usr/lib/libyggdrasilffi.so
VOLUME ["/data"]
ENV DATA_DIR="/data"
EXPOSE 8080
ENTRYPOINT ["/entrypoint"]