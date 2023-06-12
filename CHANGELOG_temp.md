Temporary changelog for the long-lived branch for decK-apiops. To minimize conflicts
when rebaseing on the `main` branch. To be integrated into `CHANGELOG.md` when the
branch is ready for merging.

Note: the V2 below, will NOT be the final version it lands in!

# Table of Contents

- [v2.0.0](#v200)

## [v2.0.0]

> Release date: to-be-set

### Added

- Added a new command `openapi2kong` that will generate a deck file from an OpenAPI
  3.0 spec. This is the replacement for the similar `inso` functionality.
  The functionality is imported from the [go-apiops library](https://github.com/Kong/go-apiops).
  [#939](https://github.com/Kong/deck/pull/939)
- Added a new command `merge` that will merge multiple deck files. The files will not be
  validated, which allows for working with incomplete or even invalid files in a pipeline.
  The functionality is imported from the [go-apiops library](https://github.com/Kong/go-apiops).
  [#939](https://github.com/Kong/deck/pull/939)
- Added a new command `patch` for applying patches on top of a decK file. The patches can be
  provided on the commandline, or via patch files. The deck file will not be
  validated, which allows for working with incomplete or even invalid files in a pipeline.
  The functionality is imported from the [go-apiops library](https://github.com/Kong/go-apiops).
  [#939](https://github.com/Kong/deck/pull/939)

### Fixes


### Misc


[v2.0.0]: https://github.com/kong/deck/compare/main...v2
