# Overleash

## Override your Unleash feature flags blazing fast

---

## Usage

The simplest way to get started is with the provided `docker compose` configuration. Copy the `.env.example` to `.env`
and fill in your options.

And then run:
`docker compose up`

## Docker image

The docker image of Overleash is also published to Docker Hub under `iandenh/overleash`.

### Traefik

The docker-compose file provide as labels to register as a service Traefik 2+. For this to work correctly make sure to
set the correct `NETWORK_NAME`.
The default host name for optimizely is `overleash.test`.

### Using in application

To make use of the overleash in your application you need to change your Unleash host. Instead
of https://unleash.mysite.com to for example `http://overleash.test`.

## Yggdrasil - Frontend API

This project supports the frontend API. This is build with [yggdrasil](https://github.com/Unleash/yggdrasil) the
reusable Unleash SDK domain logic.
Currently, an own fork is used with some extra FFI functions https://github.com/Iandenh/yggdrasil/tree/resolve_all

## Local build

This project is created with golang version 1.22. For templating it also
uses [templ](https://github.com/a-h/templ/tree/main), so make sure that is installed.
Generating templ templates can be done with `templ generate`

Because we are using yggdrasil you also need to make sure the `libyggdrasilffi.so` is correctly linked. This can be done
with:
`export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$(pwd)/unleashengine`

### Config

Setting the config for Overleash can be set in two ways, environment variable or cli flags.

Here are the config options:

| Name 	                        | Value	                                                                                                                                                                   | Env	                      | Flag	              |
|-------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------|--------------------|
| Unleash url	                  | 	 example: `https://unleash.my-site.com`, this url without `/api`, can be a Unleash instance or Unleash edge                                                             | `OVERLEASH_URL`	          | `--url`	           |
| Reload frequency	             | 	  default: `0`, example: `1`, value is in minutes, `0` is no reload                                                                                                     | `OVERLEASH_RELOAD`	       | `--reload`	        |
| Server port	                  | default: `5433`                                                                                                                                                          | `OVERLEASH_PORT`	         | `--port`	          |
| Unleash token (Client token)	 | example:  `*:development.a2a6261d38fe4f9c86aceddce09a00df6c348fd0feeab3c24a9547f2` Token or tokens that are used to fetch the feature flag config from upstream unleash. | `OVERLEASH_TOKEN`	        | `--token`	         |
| Dynamic mode	                 | default:  `false` If it's needs to start in Dynamic mode. We recommend providing the token as the config option above.                                                   | `OVERLEASH_DYNAMIC_MODE`	 | `--dynamic-mode`	  |
| Verbose	                      | default:  `false` Logs a bit more information for diagnose issues                                                                                                        | `OVERLEASH_VERBOSE`	      | `--verbose`	       |
| Proxy metrics to upstream	    | default:  `false` Proxy the metric calls to the upstream. Make sure the correct token are in the authorization header.                                                   | `OVERLEASH_PROXY_METRICS`	      | `--proxy-metrics`	 |

### Dynamic mode

Dynamic mode is an alternative of providing a token. With this mode we are extracting the token from the authorization
header. For this mode setting the overrides are available before the feature flag are loaded. Feature flag config will
be loaded after the first request that has a Client token included in the authorization header. Requesting this before
sending a Client token will result in an error 401. This includes request with frontend tokens. Only after the first
request with Client token all the api calls are unlocked.

This mode is discourage.