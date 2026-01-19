# Overleash

## Override your Unleash feature flags

Overleash is a developer tool to quickly and safely **override feature flags** for your Unleash-powered apps. You can toggle flags in your environment instantly—no need for tricky upstream config changes, and no need to coordinate with the whole team. This helps you move faster, test new features, and keep your development smooth.

---

## Key Concepts & Features

### Dynamic Overrides with Multiple Storage Backends
Easily override flags without changing upstream configs. Overleash lets you store overrides with either:
- **File (default):** Simple, zero dependencies, great for local development.
- **Redis:** For distributed/high-availability setups. Uses Redis Pub/Sub to synchronize overrides instantly between Overleash instances.

### Reliable and Resilient Bootstrapping
Overleash always tries to fetch configs from your upstream Unleash first. If that's not available, it uses the most recent backup from your storage (file or Redis). This keeps your app working even if Unleash is down.

### Instant Updates with Delta Streaming
Instead of polling for changes, Overleash can connect to Unleash’s Server-Sent Events (SSE) to receive updates as soon as feature flags change, keeping things fast and fresh.

### Environment Handling Modes
- **Dashboard-driven (default):** Control which environment’s flags you’re using directly in the Overleash dashboard. Ideal for dev/local work.
- **Token-driven:** Automatically select the environment based on the client token in the Authorization header. Useful for serving flag data to multiple environments (dev, staging, etc) from a single instance.

### Other Highlights
- Web dashboard to view/manage flags
- Multi-token support for testing multiple Unleash setups
- Proxy and forward metrics as needed
- Expose Prometheus metrics for easy monitoring
- Webhook support for on-demand refreshes from CI or Unleash

---

## 🚀 Quick Start

1. Copy `.env.example` to `.env` and update your config values.
2. Start with Docker Compose:
   ```bash
   docker compose up
   ```
3. Change your Unleash SDK configuration:
   ```
   From: https://unleash.mysite.com
   To:   http://overleash.test
   ```

---

## Docker image
You’ll find Overleash on Docker Hub:
```bash
docker pull iandenh/overleash
```

### Traefik
If you're using Traefik, ensure the right `NETWORK_NAME` in your `compose.yaml`. By default, Overleash is served as `overleash.test`.

### Using in your application
Just update your Unleash SDK’s base URL to Overleash—no code changes required!

---

## Yggdrasil - Frontend API

Overleash supports the Unleash Frontend API, powered by [yggdrasil](https://github.com/Unleash/yggdrasil) SDK logic.

*Note: This project uses a custom fork of Yggdrasil-bindings with extra FFI functions—see the [resolve_all branch](https://github.com/Iandenh/yggdrasil-bindings/tree/resolve_all).* 

---

## Local build
- Written in Go; uses [templ](https://github.com/a-h/templ/tree/main) for templates
- After editing templates, re-run: `templ generate`
- Make sure the `libyggdrasilffi.so` library is available and linked:
   ```bash
   export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$(pwd)/unleashengine
   ```

---

## Config

Configuration can be set via command-line flags or environment variables. Environment variables are prefixed with `OVERLEASH_` and are derived from the flag names (e.g., `--listen-address` becomes `OVERLEASH_LISTEN_ADDRESS`).

### **Core**
| Flag         | Environment Variable | Description                                                                                                                   | Default |
|:-------------|:---------------------|:------------------------------------------------------------------------------------------------------------------------------|:--------|
| `--upstream` | `OVERLEASH_UPSTREAM` | Unleash upstream URL to load feature flags (e.g., `https://unleash.my-site.com`), can be an Unleash instance or Unleash Edge. | `""`    |
| `--token`    | `OVERLEASH_TOKEN`    | Comma-separated Unleash client token(s) to fetch feature flag configurations.                                                 | `""`    |
| `--url`      | `OVERLEASH_URL`      | **DEPRECATED**. Use `--upstream` instead.                                                                                     | `""`    |

---

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