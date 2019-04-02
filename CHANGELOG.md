# Table of Contents

- [v0.2.0](#v020---20190401)
- [v0.1.0](#v010---20190112)

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
  Thanks to @davidcv5 for the contribution.

### Fixed

- Infinite loop in pagination for exporting entities in Kong
  [#2](https://github.com/hbagdi/deck/pull/2)  
  Thanks to @lmika for the contribution.
- Plugins are updated using PUT requests instead of PATCH to
  avoid any schema violations.

## [v0.1.0] - 2019/01/12

### Summary

Debut release of decK

[v0.2.0]: https://github.com/hbagdi/deck/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/hbagdi/deck/compare/0c7e839...v0.1.0
