# Table of Contents

- [v0.3.0](#v030---20190514)
- [v0.2.0](#v020---20190401)
- [v0.1.0](#v010---20190112)

## [v0.3.0] - 2019/05/14

### Breaking changes

No breaking changes have been introduced in this release.

### Added

- **Tag-based distrubuted configuration management**  
  Only a subset of Kong entities sharing a (set of) tag can now be exported,
  deleted, diffed or synced.
  decK can now manage your Kong's configuration in a distributed manner,
  whereby you can split Kong's configuration by team and each team can manage
  it's own configuration. Use `select-tag` feature in all the commands and
  config file for this purpose.
  [#17](https://github.com/hbagdi/deck/pull/17)
- **Read/write state from stdout/stdin**
  Config file can now be read in from standard-input and written out to
  standard-output.
  [#10](https://github.com/hbagdi/deck/pull/10),
  [#11](https://github.com/hbagdi/deck/pull/11)  
  Thanks to [@matthewbednarski](https://github.com/matthewbednarski) for the contribution.
- **Automated defaults**  
  No need to specify default values for all core Kong entities,
  further simplifying your Kong's configuration.
  Default values for plugin configuration still need to be defined, this is on
  the roadmap.
  [b448d4f](https://github.com/hbagdi/deck/commit/b448d4f)
- Add support for new properties in Upstream entity in Kong.
  [080200d](https://github.com/hbagdi/deck/commit/080200d)
- Empty plugins and other Kong entities are not populated in the config file
  as empty arrays to keep the file concise and clean.
  [ae38f1b](https://github.com/hbagdi/deck/commit/ae38f1b)

### Fixed

- Empty arrays in plugin configs are not treated as nil anymore.
  [#9](https://github.com/hbagdi/deck/pull/9)
- Correctly sync plugins which are out of sync. Protocols field
  in plugins can be confused with protocols field in routes in Kong
  [#6](https://github.com/hbagdi/deck/pull/6)  
  Thanks to [@davidcv5](https://github.com/davidcv5) for the contribution.
- Throw an error if an object is not marshalled into YAML correctly.
- Correctly create service-level plugins for Kong >= 1.1
  [#16](https://github.com/hbagdi/deck/pull/16)  

### Misc

- `go-kong` has been bumped up to v0.4.1.

## [v0.2.0] - 2019/04/01

### Breaking changes

No breaking changes have been introduced in this release.

### Added

- **Consumers and consumer-level plugins** can now be exported from Kong and
  synced to Kong.
- `--skip-consumers` flag has been introduced to various sub-commands to skip
  management of consumers in environments where they are created dynamically.`
- **Authentication support**: custom HTTP Headers (key:value) can be injected
  into requests that decK makes to Kong's Admin API using the `--headers`
  CLI flag.
  [#1](https://github.com/hbagdi/deck/pull/1)  
  Thanks to [@davidcv5](https://github.com/davidcv5) for the contribution.

### Fixed

- Infinite loop in pagination for exporting entities in Kong
  [#2](https://github.com/hbagdi/deck/pull/2)  
  Thanks to [@lmika](https://github.com/lmika) for the contribution.
- Plugins are updated using PUT requests instead of PATCH to
  avoid any schema violations.

## [v0.1.0] - 2019/01/12

### Summary

Debut release of decK

[v0.3.0]: https://github.com/hbagdi/deck/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/hbagdi/deck/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/hbagdi/deck/compare/0c7e839...v0.1.0
