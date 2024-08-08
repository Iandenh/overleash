FROM rust AS rust-build-stage
WORKDIR /yggdrasil
RUN git clone --depth 5 --branch resolve_all https://github.com/Iandenh/yggdrasil.git .
RUN cargo build --release
RUN ls target/release

# Go Build.
FROM golang:1.22 AS build-stage
WORKDIR /app
COPY --from=rust-build-stage /yggdrasil/target/release/libyggdrasilffi.so /app/unleashengine/libyggdrasilffi.so
RUN go install github.com/a-h/templ/cmd/templ@latest
COPY go.mod go.sum ./
RUN go mod download
RUN mkdir /data
COPY . /app
RUN templ generate
RUN GOOS=linux go build -o /entrypoint main.go

# Deploy.
FROM debian AS release-stage
ENV OVERLEASH_PORT=8080
WORKDIR /
RUN useradd -ms /bin/sh -u 1001 nonroot
USER nonroot
COPY --from=build-stage /entrypoint /entrypoint
COPY --from=rust-build-stage /yggdrasil/target/release/libyggdrasilffi.so /usr/lib/libyggdrasilffi.so
VOLUME ["/data"]
ENV DATA_DIR="/data"
EXPOSE 8080
ENTRYPOINT ["/entrypoint"]