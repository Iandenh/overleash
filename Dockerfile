# Stage 1: Rust Build for a static library using MUSL on Alpine
FROM --platform=$BUILDPLATFORM rust:1.90-alpine AS rust-build-stage
WORKDIR /yggdrasil-bindings
ARG TARGETPLATFORM

# Install the C build toolchain for Alpine
RUN apk add --no-cache build-base

RUN case "$TARGETPLATFORM" in \
  "linux/arm64") echo aarch64-unknown-linux-musl > /rust_target.txt ;; \
  "linux/amd64") echo x86_64-unknown-linux-musl > /rust_target.txt ;; \
  *) exit 1 ;; \
esac

# Add the MUSL target to the Rust toolchain
RUN rustup target add $(cat /rust_target.txt)
COPY ./yggdrasil-bindings /yggdrasil-bindings

# Build the static library against the MUSL target
RUN cargo build --release --target $(cat /rust_target.txt)
RUN cp target/$(cat /rust_target.txt)/release/libyggdrasilffi.a libyggdrasilffi.a

# ----------------------------------------------------------------

# Stage 2: Go Build for a fully static binary on Alpine
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build-stage
ARG VERSION
ARG TARGETOS
ARG TARGETARCH

# Install the C build toolchain for Alpine
RUN apk add --no-cache build-base

WORKDIR /app

# Create a non-root user. We will copy the user definition to the final stage.
RUN adduser -D -u 1001 -s /bin/sh noroot && mkdir /data && chown -R noroot:noroot /data

# Copy the static library from the Rust stage
COPY --from=rust-build-stage /yggdrasil-bindings/libyggdrasilffi.a /app/unleashengine/libyggdrasilffi.a

# Standard Go build steps
RUN go install github.com/a-h/templ/cmd/templ@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . /app
RUN templ generate

# Build the final static binary. The Alpine toolchain inherently supports this.
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -tags yggdrasil_static \
    -ldflags="-linkmode external -extldflags \"-static\" -s -w -X github.com/Iandenh/overleash/internal/version.Version=${VERSION}" \
    -o /entrypoint main.go

# ----------------------------------------------------------------

# Stage 3: Final, minimal image using distroless
FROM gcr.io/distroless/static-debian12 AS release-stage

ENV OVERLEASH_LISTEN_ADDRESS=":8080"
ENV DATA_DIR="/data"

LABEL org.opencontainers.image.title="Overleash"
LABEL org.opencontainers.image.description="Override your Unleash feature flags blazing fast"
LABEL org.opencontainers.image.source="https://github.com/Iandenh/overleash"
LABEL org.opencontainers.image.licenses="MIT"

WORKDIR /

# Copy the static binary, user definitions, and data directory from the build stage
COPY --from=build-stage --chown=noroot:noroot /entrypoint /entrypoint
COPY --from=build-stage /etc/passwd /etc/group /etc/
COPY --from=build-stage --chown=noroot:noroot /data /data

USER noroot
EXPOSE 8080
ENTRYPOINT ["/entrypoint"]