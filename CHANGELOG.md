# Changelog

## [0.12.0](https://github.com/Iandenh/overleash/compare/v0.11.0...v0.12.0) (2025-10-13)


### Features

* add env_from_token option. That will use the env from the token in the header ([a8f524f](https://github.com/Iandenh/overleash/commit/a8f524f6520c36a4ca7a187fe57e688e49f46999))
* add first version of consuming delta api ([aa737db](https://github.com/Iandenh/overleash/commit/aa737db651037ac1b57d27ea265318e1c1b7f794))
* add first version of delta api ([58f8e65](https://github.com/Iandenh/overleash/commit/58f8e651c9229b3ba4bd47a4e60d35aa12b232b7))
* add headless mode ([c331b26](https://github.com/Iandenh/overleash/commit/c331b267277dcb2c00d5eb89b0b1935cf4aa35a7))
* add prometheus metrics ([7d093c5](https://github.com/Iandenh/overleash/commit/7d093c5be2bd6dda5f9a69cdc83b5ced768c3622))
* add proxy edge validate ([dffe331](https://github.com/Iandenh/overleash/commit/dffe331cb386d64b862e8f0511ba5d293209beac))
* add ReadTimeout, WriteTimeout and IdleTimeout back ([0af8bc3](https://github.com/Iandenh/overleash/commit/0af8bc3b167a3425e3a09f52e36e2daeb1698d11))
* add set frontend api behind setting ([64455ce](https://github.com/Iandenh/overleash/commit/64455cedbe23f7bc5fa52229da72356b0bd8318d))
* add support for bulk metrics ([33bd446](https://github.com/Iandenh/overleash/commit/33bd44643c1cef8baa3289e30d8bb63821cfdfe9))
* add support for redis as storage with pub sub ([19a4cdd](https://github.com/Iandenh/overleash/commit/19a4cdd126b41b78fc02685008cdb35d7c4b3060))
* add support for syncing overrides over the sse ([a4fd7d9](https://github.com/Iandenh/overleash/commit/a4fd7d9232644d476b9497e7f62a9cf754a28d9b))
* add webhook endpoint that wil refresh the token ([15f2df5](https://github.com/Iandenh/overleash/commit/15f2df5cf29eaf7db84e56b3d2b8f9ba50b14d76))
* allow container to be reused as a sidecar ([702fbcb](https://github.com/Iandenh/overleash/commit/702fbcbe8f8dc9a5b76a9f38adf69e1a7de4d3c5))
* broadcast webhook received if possible ([4c07b80](https://github.com/Iandenh/overleash/commit/4c07b8046e5330a044adc960dbd3d7c3c5301450))
* don't process if no subscribers ([5460d48](https://github.com/Iandenh/overleash/commit/5460d48e7e368038a88e45cc5068d328f1cc3312))
* move to listen port instead of only port ([0c39936](https://github.com/Iandenh/overleash/commit/0c3993683e38491649baab26c6992a00fb28141d))
* only expose static files in non-headless mode ([6fa5029](https://github.com/Iandenh/overleash/commit/6fa50292618c7f42de71d73090fcacc0164df660))
* send not send metrics before shutting down ([4dccef3](https://github.com/Iandenh/overleash/commit/4dccef31c5b4a247a16836c47a979de500dd2c6a))
* statically compile for docker ([3bbf950](https://github.com/Iandenh/overleash/commit/3bbf95068175a65b3e869b7132355d5eb304110c))
* support time unit in reload ([7b47fde](https://github.com/Iandenh/overleash/commit/7b47fde539bc3cd6378327b46f5acee3b568f99a))
* update README.md ([f14f77d](https://github.com/Iandenh/overleash/commit/f14f77d124baa51c36496d844dbbe1c96456e060))
* update the .env.example ([5db1663](https://github.com/Iandenh/overleash/commit/5db166372bd6a607754e800d490e7057bf4b80a9))
* update unleash_engine.h ([106a52d](https://github.com/Iandenh/overleash/commit/106a52da7a70be977d5c12885f19264fd4d92de4))
* use static distroless image ([9250a93](https://github.com/Iandenh/overleash/commit/9250a93af80588349c1542ae57a2c602d0bd3340))
* use staticly build in docker file ([217273b](https://github.com/Iandenh/overleash/commit/217273bb9a8145ddd8d107e6e952e4fcb3782db3))


### Bug Fixes

* add locks on updates from redis ([3820b76](https://github.com/Iandenh/overleash/commit/3820b7667363873a2f0a2c67c7b47c4263fbf717))
* add missing method of fakeClient ([eb9350c](https://github.com/Iandenh/overleash/commit/eb9350cfec70a7c82bcd88507c453aa70c35a9fc))
* better text rendering ([70677ab](https://github.com/Iandenh/overleash/commit/70677ab907b2c22b574579f5187ea49b3f63be4b))
* client tests ([128c6dd](https://github.com/Iandenh/overleash/commit/128c6dd0afe9874adeab9c42e8689b91585d6357))
* ConnectVia is array for sending to edge ([354f06b](https://github.com/Iandenh/overleash/commit/354f06b08d52d61f291e35a8563fbf577d19724e))
* constrain json contained both value and values ([5a1e032](https://github.com/Iandenh/overleash/commit/5a1e032dcb44aea00142c7e80a3474765b15c339))
* correct return after returning error 500 ([d9d10aa](https://github.com/Iandenh/overleash/commit/d9d10aaded1d0109c972387de7af90cbe44b74e7))
* example name of OVERLEASH_REDIS_ADDRESS ([67bba27](https://github.com/Iandenh/overleash/commit/67bba2779bf32a38ae8e8d036c544ae308e09efd))
* frontend api metrics ([4d25921](https://github.com/Iandenh/overleash/commit/4d25921d0dcd182fe9884a6123366eb6c4042a08))
* improve some nil pointers ([c8c2e69](https://github.com/Iandenh/overleash/commit/c8c2e69ff17759c818080e5c502f1c74b5f9ffe7))
* resolve broken test after adding extra option for streamer ([25b5fcf](https://github.com/Iandenh/overleash/commit/25b5fcf1cabefdafc95b3f34f287c5cb010f92c9))
* send delete override sse event ([551adc1](https://github.com/Iandenh/overleash/commit/551adc17ae220bfe24b72c58aa04dc7e17497328))
* set correct default env_from_token ([a50cd3f](https://github.com/Iandenh/overleash/commit/a50cd3ff5b20589bc12be87c4cc782394500949b))
* tests ([83dd09a](https://github.com/Iandenh/overleash/commit/83dd09adfd61f22839f63c7f2a8ffa8b4bf46466))
* tests ([e5a19fa](https://github.com/Iandenh/overleash/commit/e5a19fa0dface13cb5858a3adb05da40b50614a6))
* tests after moving to yggdrasil-bindings ([e223cbf](https://github.com/Iandenh/overleash/commit/e223cbf4851a9d97af5aebd1e56b0e649be5a502))

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
