# Changelog

## [1.0.0](https://github.com/Iandenh/overleash/compare/v0.11.0...v1.0.0) (2025-10-13)


### ⚠ BREAKING CHANGES

* move to listen port instead of only port

### Features

* add env_from_token option. That will use the env from the token in the header ([7b81cb1](https://github.com/Iandenh/overleash/commit/7b81cb1165af7ec9e4a54d0d247538c4948f7d2f))
* add first version of consuming delta api ([6328b7b](https://github.com/Iandenh/overleash/commit/6328b7bcc320c5a3312481d48ab999e19a4da63c))
* add first version of delta api ([4518a65](https://github.com/Iandenh/overleash/commit/4518a657bc6cbcd2c38694d197808075bba40678))
* add headless mode ([c331b26](https://github.com/Iandenh/overleash/commit/c331b267277dcb2c00d5eb89b0b1935cf4aa35a7))
* add prometheus metrics ([62c9d02](https://github.com/Iandenh/overleash/commit/62c9d02af8653ca490196cb8d6120989108e5446))
* add proxy edge validate ([fa33437](https://github.com/Iandenh/overleash/commit/fa334373d6cf5c10767966887674f0f9172122d5))
* add ReadTimeout, WriteTimeout and IdleTimeout back ([2b1049e](https://github.com/Iandenh/overleash/commit/2b1049eff895d525b1f4de2a6f2b846fe518acbf))
* add set frontend api behind setting ([4e7a715](https://github.com/Iandenh/overleash/commit/4e7a7159b87e52829ae1821efa11e84e2b164705))
* add support for bulk metrics ([e1717e2](https://github.com/Iandenh/overleash/commit/e1717e2933fdfbda14a0880d870f108472caf4aa))
* add support for redis as storage with pub sub ([fa61395](https://github.com/Iandenh/overleash/commit/fa61395438efff991f2a81207c5f4700c2c6b821))
* add support for syncing overrides over the sse ([a0f6d3b](https://github.com/Iandenh/overleash/commit/a0f6d3b5ec6cab758b623cbb4c995fb31b13ebee))
* add webhook endpoint that wil refresh the token ([e9c859b](https://github.com/Iandenh/overleash/commit/e9c859b4744f427b3fb4809e4c9f9a00103bb957))
* allow container to be reused as a sidecar ([6c2552e](https://github.com/Iandenh/overleash/commit/6c2552e643a9ede7e9bceb0e2aced7fa4df73e9d))
* broadcast webhook received if possible ([be03ba5](https://github.com/Iandenh/overleash/commit/be03ba5d921290a76fb72d5bcc7e9306f442f665))
* don't process if no subscribers ([6117352](https://github.com/Iandenh/overleash/commit/61173527876bd4cf2fd96df76fae192fd4baa4f6))
* move to listen port instead of only port ([de0135c](https://github.com/Iandenh/overleash/commit/de0135ccabe92e22a9d08fae212f394d68a354b3))
* only expose static files in non-headless mode ([d1948d2](https://github.com/Iandenh/overleash/commit/d1948d2d1ca343f894db687423e2463dd6b4f81c))
* send not send metrics before shutting down ([a9c3e62](https://github.com/Iandenh/overleash/commit/a9c3e62d9ac551ac165e0e03611c4549c5533711))
* statically compile for docker ([8f24f3c](https://github.com/Iandenh/overleash/commit/8f24f3c972e055d3bbcc44ec1c403a55a1d46096))
* support time unit in reload ([4420b36](https://github.com/Iandenh/overleash/commit/4420b3688b330cc98da710b527b50cd47ae3775e))
* update README.md ([7deab78](https://github.com/Iandenh/overleash/commit/7deab78c72b67e4a95feb3ecdd76389c3ac60ccd))
* update the .env.example ([cb30514](https://github.com/Iandenh/overleash/commit/cb30514f540e3efd1eb71cb0e043d5f2816b07db))
* update unleash_engine.h ([e760782](https://github.com/Iandenh/overleash/commit/e760782b04f30fe5bf2a59b968cc4827f6500610))
* use static distroless image ([b4d00ad](https://github.com/Iandenh/overleash/commit/b4d00ad7949a25ad29dc307d5cd8636fa8e6bec1))
* use staticly build in docker file ([5a78262](https://github.com/Iandenh/overleash/commit/5a7826209a060fc2ac6249b56252682c868293a5))


### Bug Fixes

* add locks on updates from redis ([2f97e19](https://github.com/Iandenh/overleash/commit/2f97e19d9fbc700b961e454992a8531f8b2ec1a1))
* add missing method of fakeClient ([9b3a85c](https://github.com/Iandenh/overleash/commit/9b3a85c1f8671eddaf8c85cc0e86d781566a1a6c))
* better text rendering ([c17e94d](https://github.com/Iandenh/overleash/commit/c17e94d617ec6fc8b7e13385ea98ace79b37e94c))
* client tests ([43734e4](https://github.com/Iandenh/overleash/commit/43734e48334fc9ff8bafcddadfdeba41c73bb1a5))
* ConnectVia is array for sending to edge ([67cd366](https://github.com/Iandenh/overleash/commit/67cd366862f904e56c56411cb647eb1e533e64a8))
* constrain json contained both value and values ([e3bd9d6](https://github.com/Iandenh/overleash/commit/e3bd9d6eec335ec952c13a6079b4685d92ff75bf))
* correct return after returning error 500 ([02452eb](https://github.com/Iandenh/overleash/commit/02452eb91ecb2cd3b260c023f1153d4988215d9f))
* example name of OVERLEASH_REDIS_ADDRESS ([c85f691](https://github.com/Iandenh/overleash/commit/c85f6912db3c9d1e4370b4d1da83414570e7824a))
* frontend api metrics ([a7ba31a](https://github.com/Iandenh/overleash/commit/a7ba31a53ac916a9c61c44f04cfe204c1b751322))
* improve some nil pointers ([4934412](https://github.com/Iandenh/overleash/commit/49344125dc0f9e0da45f9bc09aeaeaa386c1233a))
* resolve broken test after adding extra option for streamer ([eab2d2a](https://github.com/Iandenh/overleash/commit/eab2d2a6e7b90c50b8710574cf96c23219c98121))
* send delete override sse event ([640b78c](https://github.com/Iandenh/overleash/commit/640b78cc2425c02e5637508318b43dd0d5ea9625))
* set correct default env_from_token ([16021d1](https://github.com/Iandenh/overleash/commit/16021d11bc1fc2bb52869ddfa596699240851c65))
* tests ([66f765d](https://github.com/Iandenh/overleash/commit/66f765d091b5a15bb1ad81b7686c62bb05680d80))
* tests ([5f220bc](https://github.com/Iandenh/overleash/commit/5f220bc32f9fe4f4e4da54771b177ec308775023))
* tests after moving to yggdrasil-bindings ([483040c](https://github.com/Iandenh/overleash/commit/483040ce6817371e24a07326569c8467076b70cc))

## [0.11.0](https://github.com/Iandenh/overleash/compare/v0.10.0...v0.11.0) (2025-08-18)


### Features

* add register to unleash option ([cf4de71](https://github.com/Iandenh/overleash/commit/cf4de7130c5064bb00c3bcbacea31f83aa27340c))


### Bug Fixes

* don't shrink constrain name ([517aad6](https://github.com/Iandenh/overleash/commit/517aad66e650d27f515e3c891c20bca5a413238c))
* interval in register call ([3078ad4](https://github.com/Iandenh/overleash/commit/3078ad4b1fa06ea31730e811117aa69c396aa0b1))

## [0.10.0](https://github.com/Iandenh/overleash/compare/v0.9.4...v0.10.0) (2025-07-23)


### Features

* improve constrain detail view ([7f74683](https://github.com/Iandenh/overleash/commit/7f74683d17afc81fa27c2e8610b18b415c63c3d3))
* refactor structure of multiple environments ([1bad421](https://github.com/Iandenh/overleash/commit/1bad421df94707e5bc9c17d8fb4c2671874a7f46))
* show all environment on constraint detail view ([ca4f7e5](https://github.com/Iandenh/overleash/commit/ca4f7e526b748d55f9dbe23baffb25138fb3c1e6))


### Bug Fixes

* overleash tests ([30c2c7c](https://github.com/Iandenh/overleash/commit/30c2c7cf6c4682321a0f236e7a162a8f9b91c3a9))

## [0.9.4](https://github.com/Iandenh/overleash/compare/v0.9.3...v0.9.4) (2025-07-11)


### Bug Fixes

* installing ca-certificate ([52d06ef](https://github.com/Iandenh/overleash/commit/52d06efe745272f463021123f52ff690bba00154))

## [0.9.3](https://github.com/Iandenh/overleash/compare/v0.9.2...v0.9.3) (2025-07-11)


### Bug Fixes

* make sure ca-certificates are installed ([8dfd3bf](https://github.com/Iandenh/overleash/commit/8dfd3bfcebd0caf8cf697be5537fcbec002c7423))

## [0.9.2](https://github.com/Iandenh/overleash/compare/v0.9.1...v0.9.2) (2025-07-10)


### Bug Fixes

* for code scanning alert no. 1: Workflow does not contain permissions ([#38](https://github.com/Iandenh/overleash/issues/38)) ([c9ab6d9](https://github.com/Iandenh/overleash/commit/c9ab6d9e3a36dfcc2ab068451657b61560ac6ecf))
* for code scanning alert no. 3: Disabled TLS certificate check ([#36](https://github.com/Iandenh/overleash/issues/36)) ([a3f7da5](https://github.com/Iandenh/overleash/commit/a3f7da5418ff72132346ee7a8bc49687ff2691ea))

## [0.9.1](https://github.com/Iandenh/overleash/compare/v0.9.0...v0.9.1) (2025-05-30)


### Bug Fixes

* add autocomplete="off" to remote select ([511a035](https://github.com/Iandenh/overleash/commit/511a035297fd5e9b65629d7f02dfb124f63eea4f))

## [0.9.0](https://github.com/Iandenh/overleash/compare/v0.8.0...v0.9.0) (2025-04-28)


### Features

* add automaxprocs ([bc1c109](https://github.com/Iandenh/overleash/commit/bc1c109714b488846baa5a1531000f1a4b3cf5a8))
* add change remote keyboard shortcuts ([d942476](https://github.com/Iandenh/overleash/commit/d9424766ccef81bcd3ad46caf343d4d0b1216838))
* add extra deploy to help chart ([f6231af](https://github.com/Iandenh/overleash/commit/f6231afd3c800e0e77164938b912736288cf6ca2))
* improve empty state ([6a28bb9](https://github.com/Iandenh/overleash/commit/6a28bb9c2ca0b0184d0bbed60b57bba232c3af85))
* rename url env to upstream ([25e151c](https://github.com/Iandenh/overleash/commit/25e151c11928ecf331a1e19b5a3e1bcf2de1cb6a))
* small help modal improvements ([5bff270](https://github.com/Iandenh/overleash/commit/5bff270fd589310e1c7ec4e1d2a1355851e3717f))


### Bug Fixes

* errors when switching remote when there are no multiple remotes ([1e039fa](https://github.com/Iandenh/overleash/commit/1e039fa142951e44d20bb585dd18b4239284e1e4))

## [0.8.0](https://github.com/Iandenh/overleash/compare/v0.7.0...v0.8.0) (2025-04-08)


### Features

* add description to flags ([69dd35d](https://github.com/Iandenh/overleash/commit/69dd35d6457b2dd875494e7775ed914c47c7912f))
* add stale chip when flag is stale ([c4e7464](https://github.com/Iandenh/overleash/commit/c4e7464ac1d586eb034d036c80e4a14da2e97c63))


### Bug Fixes

* git modules branch ([7a392d0](https://github.com/Iandenh/overleash/commit/7a392d0d5bc8237b335de04bb9eba91486eea46b))
* remove toolchain from go.mod ([6888b02](https://github.com/Iandenh/overleash/commit/6888b0211efd4c681b56a45a783378a14f93117e))

## [0.7.0](https://github.com/Iandenh/overleash/compare/v0.6.4...v0.7.0) (2025-04-05)


### Features

* also push docker images to github packages ([75d6b87](https://github.com/Iandenh/overleash/commit/75d6b87a95d0cf67816c6436930852b69eafd45c))


### Bug Fixes

* lint css ([dffc2f0](https://github.com/Iandenh/overleash/commit/dffc2f0916c374a49fc31cbb231117010d49ee5b))
* rename css variable ([0359778](https://github.com/Iandenh/overleash/commit/0359778f453c41787202c6ef5fd67903d55d2d77))

## [0.6.4](https://github.com/Iandenh/overleash/compare/v0.6.3...v0.6.4) (2025-03-24)


### Bug Fixes

* set docker tags ([7f4f58e](https://github.com/Iandenh/overleash/commit/7f4f58eb74b1e24b4461d1a1fa107bccb2f20c81))

## [0.6.3](https://github.com/Iandenh/overleash/compare/v0.6.2...v0.6.3) (2025-03-24)


### Bug Fixes

* set correct version when building ([b1b1bbc](https://github.com/Iandenh/overleash/commit/b1b1bbcb8b015274750d9ee660520de3299d4f41))

## [0.6.2](https://github.com/Iandenh/overleash/compare/0.6.1...v0.6.2) (2025-03-24)


### Bug Fixes

* set release type for release-please ([3c9b363](https://github.com/Iandenh/overleash/commit/3c9b363095d2076f310bb14be23e4b6bbcd52c96))
