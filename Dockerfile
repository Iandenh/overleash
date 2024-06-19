# Build.
FROM golang:1.22 AS build-stage
WORKDIR /app
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
COPY --from=build-stage --chown=nonroot:nonroot /data /data
COPY --from=build-stage /entrypoint /entrypoint
COPY --from=build-stage /app/unleashengine/libyggdrasilffi.so /usr/lib/libyggdrasilffi.so
VOLUME ["/data"]
ENV DATA_DIR="/data"
EXPOSE 8080
#USER nonroot:nonroot
ENTRYPOINT ["/entrypoint"]