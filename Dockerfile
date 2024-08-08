FROM --platform=$BUILDPLATFORM rust AS rust-build-stage
WORKDIR /yggdrasil
ARG TARGETPLATFORM
RUN case "$TARGETPLATFORM" in \
  "linux/arm64") echo aarch64-unknown-linux-gnu > /rust_target.txt ;; \
  "linux/amd64") echo x86_64-unknown-linux-gnu > /rust_target.txt ;; \
  *) exit 1 ;; \
esac
RUN rustup target add $(cat /rust_target.txt)
RUN git clone --depth 5 --branch resolve_all https://github.com/Iandenh/yggdrasil.git .
RUN cargo build --release --target $(cat /rust_target.txt)
RUN cp target/$(cat /rust_target.txt)/release/libyggdrasilffi.so libyggdrasilffi.so

# Go Build.
FROM --platform=$BUILDPLATFORM golang:1.22 AS build-stage
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY --from=rust-build-stage /yggdrasil/libyggdrasilffi.so /app/unleashengine/libyggdrasilffi.so
RUN go install github.com/a-h/templ/cmd/templ@latest
COPY go.mod go.sum ./
RUN go mod download
RUN mkdir /data
COPY . /app
RUN templ generate
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /entrypoint main.go

# Deploy.
FROM --platform=$BUILDPLATFORM debian AS release-stage
ENV OVERLEASH_PORT=8080
WORKDIR /
RUN useradd -ms /bin/sh -u 1001 nonroot
USER nonroot
COPY --from=build-stage /entrypoint /entrypoint
COPY --from=rust-build-stage /yggdrasil/libyggdrasilffi.so /usr/lib/libyggdrasilffi.so
VOLUME ["/data"]
ENV DATA_DIR="/data"
EXPOSE 8080
ENTRYPOINT ["/entrypoint"]