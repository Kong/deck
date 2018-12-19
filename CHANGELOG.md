# Table of Contents

- [0.3.0](#030---20181219)
- [0.2.0](#020---20181219)
- [0.1.0](#010---20181201)

## [0.3.0] - 2018/12/19

### Summary

- This release adds support for Kong 1.0.
  It is not compatible with 0.x.y  versions of Kong due to breaking
  Admin API changes as the deprecated API entity is dropped.
- The code and API for the library is same as 0.2.0, with the exception
  that struct defs and services related to `API` is dropped.

### Breaking changes

- `API` struct definition is no longer available.
- `APIService` is no longer available. Please ensure your code doesn't rely
  on these before upgrading.
- `Plugin` struct has dropped the `API` field.

## [0.2.0] - 2018/12/19

### Summary

- This release adds support for Kong 0.15.x.
  It is not compatible with any other versions of Kong due to breaking
  Admin API changes in Kong for Plugins, Upstreams and Targets entities.

### Breaking changes

- `Target` struct now has an `Upstream` member in place of `UpstreamID`.
- `Plugin` struct now has `Consumer`, `API`, `Route`, `Service` members
  instead of `ConsumerID`, `APIID`, `RouteID` and `ServiceID`.

### Added

- `RunOn` property has been added to `Plugin`.
- New properties are added to `Route` for L4 proxy support.

## [0.1.0] - 2018/12/01

### Summary

- Debut release of this library
- This release comes with support for Kong 0.14.x
- The library is not expected to work with previous or later
  releases of Kong since every release of Kong is introducing breaking changes
  to the Admin API.

[0.3.0]: https://github.com/hbagdi/go-kong/compare/0.2.0...0.3.0
[0.2.0]: https://github.com/hbagdi/go-kong/compare/0.1.0...0.2.0
[0.1.0]: https://github.com/hbagdi/go-kong/compare/87666c7fe73477d1874d35d690301241cd23059f...0.1.0
