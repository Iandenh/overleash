# Overleash

## Override your Unleash feature flags blazing fast 🚀

Overleash is a powerful developer tool that allows you to easily **override feature flags** on and off in your environment. It simplifies the process of making and testing changes that rely on feature flags, enabling faster iterations and smoother development without interrupting your teammates or the need for complex configurations.

---

# Features
- **Dynamic Overrides:** Override feature flags dynamically without modifying upstream configs.
- **Dashboard:** View and manage feature flags with ease.
- **Multi-Token Support:** Seamlessly test with multiple Unleash tokens.
- **Blazing Fast:** Fetch, cache, and reload feature flag configurations efficiently.
- **Proxy Metrics:** Enable metrics forwarding for your Unleash setup.

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

> Note: This project uses a custom fork of Yggdrasil with additional FFI functions: [resolve_all branch](https://github.com/Iandenh/yggdrasil/tree/resolve_all).

## Local build

This project is written in golang. For templating it
uses [templ](https://github.com/a-h/templ/tree/main), so make sure that is installed.
Generating templ templates can be done with `templ generate`

Because we are using yggdrasil you also need to make sure the `libyggdrasilffi.so` is correctly linked. This can be done
with:
`export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$(pwd)/unleashengine`

### Config

Setting the config for Overleash can be set in two ways, environment variable or cli flags.

Here are the config options:

| Name 	                        | Value	                                                                                                                                                                   | Env	                       | Flag	               |
|-------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------|---------------------|
| Unleash upstream url	         | 	 example: `https://unleash.my-site.com`, this url without `/api`, can be a Unleash instance or Unleash edge                                                             | `OVERLEASH_UPSTREAM`	      | `--upstream`	       |
| Reload frequency	             | 	  default: `0`, example: `1`, value is in minutes, `0` is no reload                                                                                                     | `OVERLEASH_RELOAD`	        | `--reload`	         |
| Listen address	               | default: `:5433`                                                                                                                                                         | `OVERLEASH_LISTEN_ADDRESS`	          | `--listen-address`	 |
| Unleash token (Client token)	 | example:  `*:development.a2a6261d38fe4f9c86aceddce09a00df6c348fd0feeab3c24a9547f2` Token or tokens that are used to fetch the feature flag config from upstream unleash. | `OVERLEASH_TOKEN`	         | `--token`	          |
| Verbose	                      | default:  `false` Logs a bit more information for diagnose issues                                                                                                        | `OVERLEASH_VERBOSE`	       | `--verbose`	        |
| Proxy metrics to upstream	    | default:  `false` Proxy the metric calls to the upstream. Make sure the correct token are in the authorization header.                                                   | `OVERLEASH_PROXY_METRICS`	 | `--proxy-metrics`	  |

## API Endpoints
### Client API
| Method | Endpoint                | Description                                                                                 |
|--------|-------------------------|---------------------------------------------------------------------------------------------|
| GET | /api/client/features    | Fetch all feature flags.                                                                    |
| GET | /api/client/features/{key} | Fetch a specific feature flag.                                                              |
| POST | /api/client/metrics     | Proxy metrics to Unleash server when proxy metrics is enabled, otherwise return always 200. |
| POST | /api/client/register    | Register client. Always returns 200.                                                        |


### Frontend API
| Method | Endpoint                         | Description                                                                                |
|---------|----------------------------------|--------------------------------------------------------------------------------------------|
| GET | /api/frontend                    | Fetch evaluated toggles.                                                                   |
| POST | /api/frontend                    | Fetch toggles with custom context.                                                         |
| GET | /api/frontend/features/{featureName} | Fetch a specific feature evaluation.                                                       |
| POST | /api/frontend/client/metrics     | Proxy metrics to Unleash server when proxy metrics is enabled, otherwise return always 200. |
| POST | /api/frontend/client/register    | Register frontend client. Always returns 200.                                              |


### Dashboard Overrides
These endpoints return html.

| Method | Endpoint                          | Description                                                                        |
|--------|-----------------------------------|------------------------------------------------------------------------------------|
| POST | /override/{key}/{enabled}         | Override a feature flag `true` for enabled and `false` for disabled. |
| POST | /override/constrain/{key}/{enabled} | Add a constraint override.                                                         |
| DELETE | /override/{key}                   | Remove an override.                                                                |
| POST | /dashboard/refresh                | Refresh feature flag data.                                                         |
| POST | /dashboard/pause                  | Pause Overleash updates.                                                           |
| POST | /dashboard/unpause                | Resume Overleash updates.                                                          |
