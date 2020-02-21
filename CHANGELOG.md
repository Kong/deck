# Table of Contents

- [v1.0.2](#v102---20200221)
- [v1.0.1](#v101---20200214)
- [v1.0.0](#v100---20200118)
- [v0.7.2](#v072---20191229)
- [v0.7.1](#v071---20191224)
- [v0.7.0](#v070---20191207)
- [v0.6.2](#v062---20191116)
- [v0.6.1](#v061---20191108)
- [v0.6.0](#v060---20191103)
- [v0.5.2](#v052---20190915)
- [v0.5.1](#v051---20190824)
- [v0.5.0](#v050---20190818)
- [v0.4.0](#v040---20190610)
- [v0.3.0](#v030---20190514)
- [v0.2.0](#v020---20190401)
- [v0.1.0](#v010---20190112)

## [v1.0.2] - 2020/02/21

### Fixed

- Fix broken `ca_certificate` entity support
  [#127](https://github.com/hbagdi/deck/pull/127)

## [v1.0.1] - 2020/02/14

### Added

- decK now supports the `url` sugar property on Service entity.
  [#123](https://github.com/hbagdi/deck/issues/123)

## [v1.0.0] - 2020/01/18

### Fixed

- decK doesn't error out if bundled plugins in Kong are disabled
  [#121](https://github.com/hbagdi/deck/pull/121)
- Consumer-specific plugins are excluded when `--skip-consumers` is used
  [#119](https://github.com/hbagdi/deck/issues/119)

### Internal

- `go-kong` has been upgraded to v0.11.0, which brings in support for
  Kong 2.0.
- All other dependencies have also been upgraded, but these have no
  user-visible changes.
  [b603f9](https://github.com/hbagdi/deck/commit/b603f9)

## [v0.7.2] - 2019/12/29

### Fixed

- Kong's version is correctly parsed; v0.7.1 is unusable because
  of this bug.
  [#117](https://github.com/hbagdi/deck/issues/117)

## [v0.7.1] - 2019/12/24

### Fixed

- Backward compatibility for credentials; tags are no longer injected into
  credentials for Kong versions below 1.4
  [#114](https://github.com/hbagdi/deck/issues/114)

## [v0.7.0] - 2019/12/07

### Breaking changes

- `sync` command now shows the progress of the sync. Previously, the command
  did not output anything but errors.

### Added

- Configuration of multiple plugin instances can now be de-duplicated using
  `_plugin_configs` field in the state file.
  [#93](https://github.com/hbagdi/deck/issues/93)
- A summary is now presented at the end of a `diff` or `sync`
  operation showing the count of resources created/updated/deleted.
  [#101](https://github.com/hbagdi/deck/issues/101)
- `sync` command now shows the progress of the sync as the sync takes place,
  making it easier to track progress in large environments.
  [#100](https://github.com/hbagdi/deck/issues/100)
- `--non-zero-exit-code` flag hsa been added to `diff` command. Using
  this flag causes decK to exit with a non-zero exit code if a diff is
  detected, making it easier to script decK in CI pipelines.
  [#98](https://github.com/hbagdi/deck/issues/98)
- A new docs website has been setup for the project:
  [https://deck.yolo42.com](https://deck.yolo42.com)

## [v0.6.2] - 2019/11/16

### Fixed

- Service-less routes are correctly processed
  [#103](https://github.com/hbagdi/deck/issues/103)
- Plugins for routes are correctly processed
  [#104](https://github.com/hbagdi/deck/issues/104)

## [v0.6.1] - 2019/11/08

### Fixed

- Check for workspace makes call the right endpoint
  [#94](https://github.com/hbagdi/deck/issues/94)
- Error checking is performed correctly when ensuring existence of a workspace
  [#95](https://github.com/hbagdi/deck/issues/95)
- Multiple upstream definitions are read correctly and synced up
  [#96](https://github.com/hbagdi/deck/issues/96)

## [v0.6.0] - 2019/11/03

### Breaking changes

- `ID` field is required for `Certificate` entity. Previous state files will
  break if `ID` is not present on this entity. You can use `dump` command
  to generate new state files which includes the `ID` field.
- SNIs are exported under the `name` key under Certificate entity to match
  Kong's declarative configuration format.

### Added

- Kong's configuration can now be synced/diffed/dumped using JSON format,
  in addition to the existing YAML format. Use the `--format` flag to specify
  the format.
  [#35](https://github.com/hbagdi/deck/issues/35)
- Plugins associated with multiple entities e.g. a plugin for a combination of
  route and a consumer in Kong are now supported.
  [#13](https://github.com/hbagdi/deck/issues/13)
- JSON-schema based validation is now performed on the input file(s) for every
  command.
- New `validate` command has been added to validate an existing state file.
  This performs a JSON-schema based sanity check on the file along-with foreign
  reference checks to check for dangling pointers.
- Service-less routes are now supported by decK.
- `name` is no longer a required field  for routes and services entities
  in Kong. If a `name` is not present, decK exports the entity with it's `ID`.
- Client-certificates on Service entity are now a supported.
- Credential entities like key-auth, basic-auth now support tagging.
- `--parallelism` flag has been added to `sync` and `diff` commands to control
  the number of concurrenty request to Kong's Admin API.
  [#85](https://github.com/hbagdi/deck/issues/85)
- `diff` and `sync` show a descriptive error when a workspace doesn't exist
  for Kong Enterprise.
  [102ed5dd](https://github.com/hbagdi/deck/commit/102ed5dd6f8ef)
- `--select-tag` flag has been added to `diff` and `sync` command for use-cases
  where the tags are not part of the state file. It is not recommended to
  use these flags unless you know what you are doing.
  [#81](https://github.com/hbagdi/deck/issues/81)
- ID for any entity can now be specified. decK previously ignored the ID for
  any entity if one was specified. Entities can also be exported with the `ID`
  field set using `--with-id` flag on the `dump` command.
  [#29](https://github.com/hbagdi/deck/issues/29)

### Fixed

- decK runs as non-root user in the Docker image.
  [#82](https://github.com/hbagdi/deck/issues/82)
- SNIs are now exported same as Kong's format i.e. they are exported under a
  `name` key under the certificates entity.
  [#76](https://github.com/hbagdi/deck/issues/76)
- Errors are made more descriptive in few commands.
- decK's binary inside the Docker image now contains versioning information.
  [#38](https://github.com/hbagdi/deck/issues/38)


### Internal

- Go has been bumped up to `1.13.4`.
- `go-kong` has been bumped up to `v0.10.0`.
- Reduced memory allocation, which should result in less GC pressure.

## [v0.5.2] - 2019/09/15

### Added

- `-w/--workspace` flag has been added to the `reset` command to reset a
  specific workspace in Kong Enterprise.
  [#74](https://github.com/hbagdi/deck/issues/74)
- `--all-workspaces` flag has been added to the `reset` command to reset
  all workspaces in Kong Enterprise.
  [#74](https://github.com/hbagdi/deck/issues/74)
- A warning is logged when basic-auth credentials are being synced.
  [#49](https://github.com/hbagdi/deck/issues/49)

### Fixed

- Kong Enterprise Developer Portal exposes the credentials (basic/key) of
  Developers on the Admin API, but doesn't expose the consumers causing
  issues during export. decK now ignores these credentials in Kong Enterprise.
  [#75](https://github.com/hbagdi/deck/issues/75)

### Internal

- Go version has been bumped to 1.13.

## [v0.5.1] - 2019/08/24

### Added

- `oauth2` credentials associated with consumers are now supported.
  [#67](https://github.com/hbagdi/deck/pull/67)

### Fixed

- The same target can be associated with multiple upstreams.
  [#57](https://github.com/hbagdi/deck/issues/57)
- Fix compatibility with Kong < 1.3.
  [#59](https://github.com/hbagdi/deck/issues/59)
- Ignore credentials for consumers which are not in the sub-set of
  the configuration being synced.
  [#65](https://github.com/hbagdi/deck/issues/65)

## [v0.5.0] - 2019/08/18

### Summary

This release brings the following features:
- Consumer credentials are now supported
- Support for Kong 1.3
- Kong Enterprise workspace support
- Reading configuration from multiple files in a directories

### Breaking changes

No breaking changes have been introduced in this release.

### Added

- **Consumer credentials**  
  The following entities associate with a consumer in Kong are now supported [#12](https://github.com/hbagdi/deck/issues/12):
  - `key-auth`
  - `basic-auth`
  - `hmac-auth`
  - `jwt`
  - `acl`
  
- decK's exported YAML is now compatible with Kong's declarative config
  file.
- **Homebrew support**  
  decK can now be installed using Homebrew on macOS:
  ```
  brew tap hbagdi/deck
  brew install deck
  ```
- **Multiple state files**  
  decK can now read the configuration of Kong from multiple YAML files in a directory. You can split your configuration
  into files in any way you would like.
  [#22](https://github.com/hbagdi/deck/issues/22)
- Upcoming Kong 1.3 is now supported.
  [#36](https://github.com/hbagdi/deck/issues/36)
- **Kong Enterprise only features:**  
  Workspaces are now natively supported in decK
  - `-w/--workspace` flag can be specified in the `dump` command to
    export configuration of a single workspace.
  - `--all-workspaces` flag in `dump` command will export all workspaces
    in Kong Enteprise. Each workspace lives in a separate state file.
  - `diff` and `sync` command now support workspaces via the `_workspace`
    attribute in the state file.

### Fixed

- decK now supports TCP services in Kong.
  [#44](https://github.com/hbagdi/deck/issues/44)
- Add missing `interval` field in Upstream entity's
  unhealthy active healthchecks
  [#45](https://github.com/hbagdi/deck/pull/45)
- Docker image now contains only the binary and not the entire source code.
  [#34](https://github.com/hbagdi/deck/pull/34)  
  Thanks to [David Cruz](https://github.com/davidcv5) for the contribution.

## [v0.4.0] - 2019/06/10

### Summary

This release introduces support for Kong 1.2.x.

### Breaking changes

- `strip_path` attribute of Route can now be set to false. The default value
  is now false, which was true previously.
  [#18](https://github.com/hbagdi/deck/issues/18)

### Added

- `https_redirect_status_code` attribute of Route in Kong can be set,
  and defaults to `426`.

## [v0.3.0] - 2019/05/14

### Breaking changes

No breaking changes have been introduced in this release.

### Added

- **Tag-based distributed configuration management**  
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
- Docker image is now available via Docker Hub.
  You can use `docker pull hbagdi/deck` to pull down decK in a Docker image.

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

[v1.0.2]: https://github.com/hbagdi/deck/compare/v1.0.1...v1.0.2
[v1.0.1]: https://github.com/hbagdi/deck/compare/v1.0.0...v1.0.1
[v1.0.0]: https://github.com/hbagdi/deck/compare/v0.7.2...v1.0.0
[v0.7.2]: https://github.com/hbagdi/deck/compare/v0.7.1...v0.7.2
[v0.7.1]: https://github.com/hbagdi/deck/compare/v0.7.0...v0.7.1
[v0.7.0]: https://github.com/hbagdi/deck/compare/v0.6.2...v0.7.0
[v0.6.2]: https://github.com/hbagdi/deck/compare/v0.6.1...v0.6.2
[v0.6.1]: https://github.com/hbagdi/deck/compare/v0.6.0...v0.6.1
[v0.6.0]: https://github.com/hbagdi/deck/compare/v0.5.2...v0.6.0
[v0.5.2]: https://github.com/hbagdi/deck/compare/v0.5.1...v0.5.2
[v0.5.1]: https://github.com/hbagdi/deck/compare/v0.5.0...v0.5.1
[v0.5.0]: https://github.com/hbagdi/deck/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/hbagdi/deck/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/hbagdi/deck/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/hbagdi/deck/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/hbagdi/deck/compare/0c7e839...v0.1.0
