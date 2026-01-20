FROM --platform=$BUILDPLATFORM rust:1.92 AS rust-build-stage
WORKDIR /yggdrasil-bindings
ARG TARGETPLATFORM

RUN apt-get update && apt-get install -y musl-tools gcc-aarch64-linux-gnu

RUN case "$TARGETPLATFORM" in \
  "linux/arm64") echo aarch64-unknown-linux-musl > /rust_target.txt ;; \
  "linux/amd64") echo x86_64-unknown-linux-musl > /rust_target.txt ;; \
  *) exit 1 ;; \
esac

RUN rustup target add $(cat /rust_target.txt)
COPY ./yggdrasil-bindings /yggdrasil-bindings

RUN cargo build --release --target $(cat /rust_target.txt)
RUN cp target/$(cat /rust_target.txt)/release/libyggdrasilffi.a libyggdrasilffi.a

# ----------------------------------------------------------------

# Stage 2: Go Build for a fully static binary
FROM --platform=$BUILDPLATFORM golang:1.25 AS build-stage
ARG VERSION
ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

RUN apt-get update && apt-get install -y musl-tools gcc-aarch64-linux-gnu

WORKDIR /app

# Create a non-root user. We will copy the user definition to the final stage.
RUN addgroup --gid 1001 noroot && adduser --disabled-password --no-create-home --uid 1001 --gid 1001 --shell /bin/sh noroot && mkdir /data && chown -R noroot:noroot /data

# Copy the static library from the Rust stage
COPY --from=rust-build-stage /yggdrasil-bindings/libyggdrasilffi.a /app/unleashengine/libyggdrasilffi.a

# Standard Go build steps
RUN go install github.com/a-h/templ/cmd/templ@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . /app
RUN templ generate

RUN \
    # Ensure the script exits on any error
    set -e; \
    \
    # Print the target platform for logging purposes
    echo "Building for platform: $TARGETPLATFORM"; \
    \
    # Start the case statement on the TARGETPLATFORM variable
    case "$TARGETPLATFORM" in \
        # If building for arm64...
        "linux/arm64") \
            echo "-> Cross-compiling for arm64. Installing cross-compiler."; \
            # Install the aarch64 C cross-compiler
            apt-get update && apt-get install -y --no-install-recommends gcc-aarch64-linux-gnu; \
            # Export the CC variable so the 'go build' command uses the correct compiler
            export CC=aarch64-linux-gnu-gcc; \
            ;; \
        \
        # If building for amd64...
        "linux/amd64") \
            echo "-> Natively compiling for amd64."; \
            # No special compiler needed, but we set CC for consistency.
            # The 'build-essential' package or default gcc is usually present.
            export CC=gcc; \
            ;; \
        \
        # If the platform is not supported...
        *) \
            echo "Error: Unsupported TARGETPLATFORM: $TARGETPLATFORM" >&2; \
            exit 1; \
            ;; \
    esac; \
    \
    # === UNIVERSAL BUILD COMMAND ===
    # This command runs after the case statement and uses the exported CC variable.
    # GOARCH is set dynamically using the automatic $TARGETARCH variable.
    CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH go build \
    -tags yggdrasil_static \
    -ldflags="-linkmode external -extldflags "-static" -s -w -X github.com/Iandenh/overleash/internal/version.Version=${VERSION}" \
    -o /overleash main.go

# ----------------------------------------------------------------

# Stage 3: Final, minimal image using distroless
FROM gcr.io/distroless/static-debian13 AS release-stage

ENV OVERLEASH_LISTEN_ADDRESS=":8080"
ENV DATA_DIR="/data"

LABEL org.opencontainers.image.title="Overleash"
LABEL org.opencontainers.image.description="Override your Unleash feature flags blazing fast"
LABEL org.opencontainers.image.source="https://github.com/Iandenh/overleash"
LABEL org.opencontainers.image.licenses="MIT"

WORKDIR /

# Copy the static binary, user definitions, and data directory from the build stage
COPY --from=build-stage --chown=noroot:noroot /overleash /overleash
COPY --from=build-stage /etc/passwd /etc/group /etc/
COPY --from=build-stage --chown=noroot:noroot /data /data

USER noroot
EXPOSE 8080
ENTRYPOINT ["/overleash"]