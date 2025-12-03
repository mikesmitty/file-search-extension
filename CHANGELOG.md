# Changelog

## [0.6.2](https://github.com/mikesmitty/file-search/compare/v0.6.1...v0.6.2) (2025-11-26)


### Bug Fixes

* Update module gopkg.in/dnaeon/go-vcr.v4 to v4.0.6 ([#68](https://github.com/mikesmitty/file-search/issues/68)) ([2d2639f](https://github.com/mikesmitty/file-search/commit/2d2639f1b5480b78208f36f86985c28abd08d47c))

## [0.6.1](https://github.com/mikesmitty/file-search/compare/v0.6.0...v0.6.1) (2025-11-24)


### Bug Fixes

* Update module github.com/mark3labs/mcp-go to v0.43.1 ([59beac2](https://github.com/mikesmitty/file-search/commit/59beac2d0636fc2217ca2a18251dc4efcc301caf))

## [0.6.0](https://github.com/mikesmitty/file-search/compare/v0.5.1...v0.6.0) (2025-11-23)


### Features

* add some quality of life tweaks to query ([#55](https://github.com/mikesmitty/file-search/issues/55)) ([7f4f7e4](https://github.com/mikesmitty/file-search/commit/7f4f7e401b8dd0323e759568cfd0437d03eba4a6))
* allow batch uploads ([#59](https://github.com/mikesmitty/file-search/issues/59)) ([adc2c6e](https://github.com/mikesmitty/file-search/commit/adc2c6ee1ff185e91ffced1702bc6d7a95ca551a))
* improve plaintext output with grounding results ([#57](https://github.com/mikesmitty/file-search/issues/57)) ([8487601](https://github.com/mikesmitty/file-search/commit/848760186f63ff388bb3264edf1090d06712cbb8))

## [0.5.1](https://github.com/mikesmitty/file-search/compare/v0.5.0...v0.5.1) (2025-11-23)


### Bug Fixes

* trigger goreleaser workflow via tag ([#51](https://github.com/mikesmitty/file-search/issues/51)) ([fcdb621](https://github.com/mikesmitty/file-search/commit/fcdb62102ab5c062ca4b381436510db09008ee20))

## [0.5.0](https://github.com/mikesmitty/file-search/compare/v0.4.0...v0.5.0) (2025-11-23)


### Features

* add force flag to document deletion ([#45](https://github.com/mikesmitty/file-search/issues/45)) ([a7e5c5f](https://github.com/mikesmitty/file-search/commit/a7e5c5f7916488d561c54342f5a8db2b17b1bec5))
* add newly added models to completion list ([#47](https://github.com/mikesmitty/file-search/issues/47)) ([32580b4](https://github.com/mikesmitty/file-search/commit/32580b4d1fbb0b471e666643b995016025dc2bd0))
* print operation ID in import-file command ([#46](https://github.com/mikesmitty/file-search/issues/46)) ([c1649fc](https://github.com/mikesmitty/file-search/commit/c1649fc3d5f1f40ff4b5cf49dfea526eb1bdff9b))


### Bug Fixes

* add more subcommand aliases ([#43](https://github.com/mikesmitty/file-search/issues/43)) ([ee1c9d7](https://github.com/mikesmitty/file-search/commit/ee1c9d7d77971a72069b9021c6136e2dae920698))

## [0.4.0](https://github.com/mikesmitty/file-search/compare/v0.3.2...v0.4.0) (2025-11-23)


### Features

* add name parameter to upload_file tool ([#41](https://github.com/mikesmitty/file-search/issues/41)) ([3da4815](https://github.com/mikesmitty/file-search/commit/3da48151e14cda1af4dda14d4b599d3c7dfa6fef))


### Bug Fixes

* update MCP list tools to return objects instead of arrays ([#37](https://github.com/mikesmitty/file-search/issues/37)) ([9402f74](https://github.com/mikesmitty/file-search/commit/9402f7432fdecdf0653d616c7e14c813177f68b5))

## [0.3.2](https://github.com/mikesmitty/file-search/compare/v0.3.1...v0.3.2) (2025-11-23)


### Bug Fixes

* handle crash due to typed nil passing ([#35](https://github.com/mikesmitty/file-search/issues/35)) ([fbbf497](https://github.com/mikesmitty/file-search/commit/fbbf497938e6d9ae11696142cf00cfeb76c38168))

## [0.3.1](https://github.com/mikesmitty/file-search/compare/v0.3.0...v0.3.1) (2025-11-22)


### Bug Fixes

* remove experimental settings storage for now ([#33](https://github.com/mikesmitty/file-search/issues/33)) ([15314c7](https://github.com/mikesmitty/file-search/commit/15314c74392e1ff0fb117470b91a640c4c77a1e8))

## [0.3.0](https://github.com/mikesmitty/file-search/compare/v0.2.0...v0.3.0) (2025-11-22)


### Features

* enable all mcp tools by default ([#25](https://github.com/mikesmitty/file-search/issues/25)) ([d4f5492](https://github.com/mikesmitty/file-search/commit/d4f5492fa856ce9e6664134452d1c927cc3c460c))


### Bug Fixes

* try storing extension api key as sensitive setting ([#31](https://github.com/mikesmitty/file-search/issues/31)) ([4b80324](https://github.com/mikesmitty/file-search/commit/4b803242806e2f2bfbf6ed3da810327d3084d903))
* update go module name ([#27](https://github.com/mikesmitty/file-search/issues/27)) ([1d04a59](https://github.com/mikesmitty/file-search/commit/1d04a5952a1938aad3d36c6c79169d195378ee6d))

## [0.2.0](https://github.com/mikesmitty/file-search/compare/v0.1.4...v0.2.0) (2025-11-22)


### Features

* rename to File Search Query ([#21](https://github.com/mikesmitty/file-search/issues/21)) ([1ae3191](https://github.com/mikesmitty/file-search/commit/1ae31919a44b6a35a41feb3063cb8dc382076a36))


### Bug Fixes

* revert to static model list ([#19](https://github.com/mikesmitty/file-search/issues/19)) ([0b0341b](https://github.com/mikesmitty/file-search/commit/0b0341b4e72d18a0a85d564cc1a3763c4ba2a6a5))

## [0.1.4](https://github.com/mikesmitty/file-search/compare/v0.1.3...v0.1.4) (2025-11-22)


### Bug Fixes

* add missing friendly name resolution in store command ([#13](https://github.com/mikesmitty/file-search/issues/13)) ([3531604](https://github.com/mikesmitty/file-search-extension/commit/35316048ed1c3b86f9631d3cbafbaf0bade3d14a))

## [0.1.3](https://github.com/mikesmitty/file-search/compare/v0.1.2...v0.1.3) (2025-11-22)


### Bug Fixes

* update extension binary path ([#8](https://github.com/mikesmitty/file-search/issues/8)) ([fd72adc](https://github.com/mikesmitty/file-search-extension/commit/fd72adcb2c6b6d834a78de26a86d59898d4aa0bb))

## [0.1.2](https://github.com/mikesmitty/file-search/compare/v0.1.1...v0.1.2) (2025-11-22)


### Bug Fixes

* modernize goreleaser config ([#5](https://github.com/mikesmitty/file-search/issues/5)) ([40cb081](https://github.com/mikesmitty/file-search-extension/commit/40cb081f7b803e5b3cf8843a439a7676aeae7ee9))

## [0.1.1](https://github.com/mikesmitty/file-search/compare/v0.1.0...v0.1.1) (2025-11-22)


### Bug Fixes

* update release workflow ([8dc17f1](https://github.com/mikesmitty/file-search/commit/8dc17f19520c0fbf41a65d978abf5b4fc9bc880f))

## 0.1.0 (2025-11-22)


### Features

* initial version ([8994ba6](https://github.com/mikesmitty/file-search/commit/8994ba6362f4ab87b9613e2665856da4a9777e22))
