# Table of Contents

- [v0.5.1](#v051---20190805)
- [v0.5.0](#v050---20190607)
- [v0.4.1](#v041---20190411)
- [v0.4.0](#v040---20190406)
- [0.3.0](#030---20181219)
- [0.2.0](#020---20181219)
- [0.1.0](#010---20181201)

## [v0.5.1] - 2019/08/05

### Fix

- Add missing healthchecks.active.unhealthy.interval field to Upstream
  [#6](https://github.com/hbagdi/go-kong/issues/6)

## [v0.5.0] - 2019/06/07

### Summary

- This release adds support for Kong 1.2.

### Added

- Added HTTPSRedirectStatusCode property to Route struct.
  [#3](https://github.com/hbagdi/go-kong/pull/3)

### Breaking change

- `Create()` for Custom Entities now supports HTTP PUT method.
  If `id` is specified in the object, it will be used to PUT the entity.
  This was always POST previously.
  [#3](https://github.com/hbagdi/go-kong/pull/3)

## [v0.4.1] - 2019/04/11

### Fix

- Add `omitempty` property to Upstream fields for Kong 1.0 compatibility

## [v0.4.0] - 2019/04/06

### Summary

- This release adds support for features released in Kong 1.1.
  This version is compatible with Kong 1.0 and Kong 1.1.

### Breaking Change

- Please note that the version naming scheme for this library has changed from
  `x.y.z` to `vX.Y.Z`. This is to ensure compatibility with Go modules.

### Added

- `Tags` field has been added to all Kong Core entity structs.
- List methods now support tag based filtering introduced in Kong 1.1.
  Tags can be ANDed or ORed together. `ListOpt` struct can be used to
  specify the tags for filtering.
- `Protocols` field has been added to Plugin struct.
- New fields `Type`, `HTTPSSni` and `HTTPSVerifyCertificate` have been
  introduced for Active HTTPS healthchecks.
- `TargetService` has two new methods `MarkHealthy()` and `MarkUnhealthy()`
  to change the health of a target.

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

[v0.5.1]: https://github.com/hbagdi/go-kong/compare/v0.5.0...v0.5.1
[v0.5.0]: https://github.com/hbagdi/go-kong/compare/v0.4.1...v0.5.0
[v0.4.1]: https://github.com/hbagdi/go-kong/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/hbagdi/go-kong/compare/0.3.0...v0.4.0
[0.3.0]: https://github.com/hbagdi/go-kong/compare/0.2.0...0.3.0
[0.2.0]: https://github.com/hbagdi/go-kong/compare/0.1.0...0.2.0
[0.1.0]: https://github.com/hbagdi/go-kong/compare/87666c7fe73477d1874d35d690301241cd23059f...0.1.0
