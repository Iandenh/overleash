# Overleash

## Override your Unleash feature flags blazing fast 🚀

Overleash is a powerful developer tool that allows you to easily **override feature flags** on and off in your environment. It simplifies the process of making and testing changes that rely on feature flags, enabling faster iterations and smoother development without interrupting your teammates or the need for complex configurations.

---

## Key Concepts & Features

### Dynamic Overrides with Multiple Storage Backends
Override feature flags dynamically without modifying upstream configs. Overleash supports two storage backends for these overrides:

* **File (Default):** Simple and requires no external dependencies. Perfect for local development.
* **Redis:** Enables a high-availability (HA) setup. When using Redis, Overleash utilizes a Pub/Sub connection to **instantly synchronize overrides** across multiple Overleash instances, ensuring consistent behavior in a distributed environment.

### Near-Instant Updates with Delta Streaming
Instead of periodically polling for changes, Overleash can connect directly to an upstream Unleash instance's Server-Sent Events (SSE) stream. This provides **near-instant updates** for any feature flag changes, ensuring your local environment is always up-to-date with minimal latency and reduced network traffic.

### Flexible Environment Handling
Overleash can operate in two modes for handling environments:
1.  **Dashboard-Driven (Default):** The active environment is managed and can be changed directly within the Overleash dashboard. This is the recommended setup for local development.
2.  **Token-Driven:** For more advanced use cases, you can enable a mode where the environment is dynamically resolved from the client token sent in the `Authorization` header. This allows a single Overleash instance to serve flags for multiple environments (e.g., dev, staging) based on the token provided by the client SDK.

### Additional Features
* **Dashboard:** View and manage feature flags with ease.
* **Multi-Token Support:** Seamlessly test with multiple Unleash tokens.
* **Proxy Metrics:** Enable metrics forwarding for your Unleash setup.
* **Prometheus Metrics:** Expose metrics for monitoring with Prometheus.
* **Webhook Refresh:** Trigger a refresh of feature flags from an external source, like an Unleash webhook.
## 🚀 Quick Start

The simplest way to get started is by using the provided `docker-compose` configuration:

1. Copy the `.env.example` to `.env` and fill in your configuration options.
2. Run:
   ```bash
   docker compose up
   ```

## Docker image
The Overleash Docker image is published to **Docker Hub**:
```bash
docker pull iandenh/overleash
```

### Traefik

The docker-compose file provide as labels to register as a service Traefik 2+. For this to work correctly make sure to
set the correct `NETWORK_NAME`.
The default host name for optimizely is `overleash.test`.

### Using in application
To use Overleash in your application, simply update your Unleash host:
```plaintext
From: https://unleash.mysite.com
To:   http://overleash.test
```

## Yggdrasil - Frontend API

Overleash supports the **Frontend API**, powered by [yggdrasil](https://github.com/Unleash/yggdrasil), the reusable Unleash SDK logic.

> Note: This project uses a custom fork of Yggdrasil-bindings with additional FFI functions: [resolve_all branch](https://github.com/Iandenh/yggdrasil-bindings/tree/resolve_all).

## Local build

This project is written in golang. For templating it
uses [templ](https://github.com/a-h/templ/tree/main), so make sure that is installed.
Generating templ templates can be done with `templ generate`

Because we are using yggdrasil you also need to make sure the `libyggdrasilffi.so` is correctly linked. This can be done
with:
`export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$(pwd)/unleashengine`

## Config

Configuration can be set via command-line flags or environment variables. Environment variables are prefixed with `OVERLEASH_` and are derived from the flag names (e.g., `--listen-address` becomes `OVERLEASH_LISTEN_ADDRESS`).

### **Core**
| Flag         | Environment Variable | Description                                                                                                                   | Default |
|:-------------|:---------------------|:------------------------------------------------------------------------------------------------------------------------------|:--------|
| `--upstream` | `OVERLEASH_UPSTREAM` | Unleash upstream URL to load feature flags (e.g., `https://unleash.my-site.com`), can be an Unleash instance or Unleash Edge. | `""`    |
| `--token`    | `OVERLEASH_TOKEN`    | Comma-separated Unleash client token(s) to fetch feature flag configurations.                                                 | `""`    |
| `--url`      | `OVERLEASH_URL`      | **DEPRECATED**. Use `--upstream` instead.                                                                                     | `""`    |

---
### **Server & Logging**
| Flag                        | Environment Variable                | Description                                                                                         | Default |
|:----------------------------|:------------------------------------|:----------------------------------------------------------------------------------------------------|:--------|
| `--listen-address`          | `OVERLEASH_LISTEN_ADDRESS`          | Address to listen on (e.g., `:5433`, `127.0.0.1:5433`).                                             | `:5433` |
| `--reload`                  | `OVERLEASH_RELOAD`                  | Reload frequency for refreshing feature flags (e.g., `5m`, `1h`). `0` disables automatic reloading. | `0`     |
| `--verbose`                 | `OVERLEASH_VERBOSE`                 | Enable verbose logging to troubleshoot and diagnose issues.                                         | `false` |
| `--prometheus-metrics`      | `OVERLEASH_PROMETHEUS_METRICS`      | Whether to collect and expose Prometheus metrics.                                                   | `false` |
| `--prometheus-metrics-port` | `OVERLEASH_PROMETHEUS_METRICS_PORT` | Port to expose Prometheus metrics on.                                                               | `9100`  |

---
### **Unleash-specific**
| Flag                    | Environment Variable            | Description                                                                                                                     | Default |
|:------------------------|:--------------------------------|:--------------------------------------------------------------------------------------------------------------------------------|:--------|
| `--register-metrics`    | `OVERLEASH_REGISTER_METRICS`    | Register metrics with the upstream Unleash server.                                                                              | `false` |
| `--register`            | `OVERLEASH_REGISTER`            | Register this Overleash instance with the upstream Unleash server.                                                              | `false` |
| `--headless`            | `OVERLEASH_HEADLESS`            | Disable the dashboard API.                                                                                                      | `false` |
| `--streamer`            | `OVERLEASH_STREAMER`            | Enable streaming of delta events from this instance.                                                                            | `false` |
| `--enable-frontend-api` | `OVERLEASH_ENABLE_FRONTEND_API` | Enable the Frontend API.                                                                                                        | `true`  |
| `--delta`               | `OVERLEASH_DELTA`               | Use the upstream delta streaming API for near-instant updates.                                                                  | `false` |
| `--env-from-token`      | `OVERLEASH_ENV_FROM_TOKEN`      | Resolve the environment from the client token in the `Authorization` header. Recommended to keep `false` for local development. | `false` |
| `--webhook`             | `OVERLEASH_WEBHOOK`             | Expose a webhook that will force a refresh of the flags when triggered.                                                         | `false` |

---
### **Storage**
| Flag         | Environment Variable | Description                                                                                                | Default |
|:-------------|:---------------------|:-----------------------------------------------------------------------------------------------------------|:--------|
| `--storage`  | `OVERLEASH_STORAGE`  | Storage backend for overrides. Options: `file` or `redis`. Use `redis` for multi-instance synchronization. | `file`  |

---
### **Redis**
(These settings are only used if `--storage` is set to `redis`)*

| Flag                | Environment Variable        | Description                                             | Default             |
|:--------------------|:----------------------------|:--------------------------------------------------------|:--------------------|
| `--redis-address`   | `OVERLEASH_REDIS_ADDRESS`   | Redis address (host:port).                              | `localhost:6379`    |
| `--redis-password`  | `OVERLEASH_REDIS_PASSWORD`  | Redis password.                                         | `""`                |
| `--redis-db`        | `OVERLEASH_REDIS_DB`        | Redis database number.                                  | `0`                 |
| `--redis-channel`   | `OVERLEASH_REDIS_CHANNEL`   | Redis Pub/Sub channel for override updates.             | `overrides-updates` |
| `--redis-sentinel`  | `OVERLEASH_REDIS_SENTINEL`  | Use Redis Sentinel for high availability.               | `false`             |
| `--redis-master`    | `OVERLEASH_REDIS_MASTER`    | Redis master name (for Sentinel).                       | `mymaster`          |
| `--redis-sentinels` | `OVERLEASH_REDIS_SENTINELS` | Comma-separated list of Sentinel addresses (host:port). | `""`                |

## API Endpoints

### **Client API**
| Method  | Endpoint                     | Description                                                                                  |
|---------|------------------------------|----------------------------------------------------------------------------------------------|
| `GET`   | `/api/client/features`       | Fetch all feature flags.                                                                     |
| `GET`   | `/api/client/features/{key}` | Fetch a specific feature flag.                                                               |
| `POST`  | `/api/client/metrics`        | Proxy metrics to Unleash server when proxy metrics is enabled; otherwise, returns 200 OK.    |
| `POST`  | `/api/client/register`       | Register client. Always returns 200 OK.                                                      |

---
### **Frontend API**
| Method   | Endpoint                               | Description                                                                                 |
|----------|----------------------------------------|---------------------------------------------------------------------------------------------|
| `GET`    | `/api/frontend`                        | Fetch evaluated toggles.                                                                    |
| `POST`   | `/api/frontend`                        | Fetch toggles with custom context.                                                          |
| `GET`    | `/api/frontend/features/{featureName}` | Fetch a specific feature evaluation.                                                        |
| `POST`   | `/api/frontend/client/metrics`         | Proxy metrics to Unleash server when proxy metrics is enabled; otherwise, returns 200 OK.   |
| `POST`   | `/api/frontend/client/register`        | Register frontend client. Always returns 200 OK.                                            |

---
### **Dashboard & Control API**
These endpoints are primarily for interacting with the Overleash dashboard or for automation.

| Method   | Endpoint                              | Description                                                                                                                                                                                        |
|----------|---------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `POST`   | `/override/{key}/{enabled}`           | Override a feature flag. Set `{enabled}` to `true` or `false`.                                                                                                                                     |
| `POST`   | `/override/constrain/{key}/{enabled}` | Add a constraint override.                                                                                                                                                                         |
| `DELETE` | `/override/{key}`                     | Remove an override.                                                                                                                                                                                |
| `POST`   | `/dashboard/refresh`                  | Manually refresh feature flag data from the upstream.                                                                                                                                              |
| `POST`   | `/dashboard/pause`                    | Pause Overleash updates.                                                                                                                                                                           |
| `POST`   | `/dashboard/unpause`                  | Resume Overleash updates.                                                                                                                                                                          |
| `POST`   | `/webhook/refresh`                    | **Webhook Endpoint**. Triggers a forced refresh of feature flags. Can be configured in the Unleash UI to notify Overleash of changes instantly. No authentication or specific payload is required. |