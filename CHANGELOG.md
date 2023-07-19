# Table of Contents

- [v1.24.0](#v1240)
- [v1.23.0](#v1230)
- [v1.22.1](#v1221)
- [v1.22.0](#v1220)
- [v1.21.0](#v1210)
- [v1.20.0](#v1200)
- [v1.19.1](#v1191)
- [v1.19.0](#v1190)
- [v1.18.1](#v1181)
- [v1.18.0](#v1180)
- [v1.17.3](#v1173)
- [v1.17.2](#v1172)
- [v1.17.1](#v1171)
- [v1.17.0](#v1170)
- [v1.16.1](#v1161)
- [v1.16.0](#v1160)
- [v1.15.1](#v1151)
- [v1.15.0](#v1150)
- [v1.14.0](#v1140)
- [v1.13.0](#v1130)
- [v1.12.4](#v1124)
- [v1.12.3](#v1123)
- [v1.12.2](#v1122)
- [v1.12.1](#v1121)
- [v1.12.0](#v1120)
- [v1.11.0](#v1110)
- [v1.10.0](#v1100)
- [v1.9.0](#v190)
- [v1.8.2](#v182)
- [v1.8.1](#v181)
- [v1.8.0](#v180)
- [v1.7.0](#v170)
- [v1.6.0](#v160)
- [v1.5.1](#v151)
- [v1.5.0](#v150)
- [v1.4.0](#v140)
- [v1.3.0](#v130)
- [v1.2.4](#v124)
- [v1.2.3](#v123)
- [v1.2.2](#v122)
- [v1.2.1](#v121)
- [v1.2.0](#v120)
- [v1.1.0](#v110)
- [v1.0.3](#v103)
- [v1.0.2](#v102)
- [v1.0.1](#v101)
- [v1.0.0](#v100)
- [v0.7.2](#v072)
- [v0.7.1](#v071)
- [v0.7.0](#v070)
- [v0.6.2](#v062)
- [v0.6.1](#v061)
- [v0.6.0](#v060)
- [v0.5.2](#v052)
- [v0.5.1](#v051)
- [v0.5.0](#v050)
- [v0.4.0](#v040)
- [v0.3.0](#v030)
- [v0.2.0](#v020)
- [v0.1.0](#v010)

## [v1.24.0]

> Release date: to-be-set

### Added

- Add a new flag (`--json-output`) to enable JSON output when using `sync` and `diff` commands
  [#798](https://github.com/Kong/deck/pull/798)
- Improved error logs coming from files validation against Kong's schemas.
  [#976](https://github.com/Kong/deck/pull/976)
- Added a new command `file openapi2kong` that will generate a deck file from an OpenAPI
  3.0 spec. This is the replacement for the similar `inso` functionality.
  The functionality is imported from the [go-apiops library](https://github.com/Kong/go-apiops).
  [#939](https://github.com/Kong/deck/pull/939)
- Added a new command `file merge` that will merge multiple deck files. The files will not be
  validated, which allows for working with incomplete or even invalid files in a pipeline.
  The functionality is imported from the [go-apiops library](https://github.com/Kong/go-apiops).
  [#939](https://github.com/Kong/deck/pull/939)
- Added a new command `file patch` for applying patches on top of a decK file. The patches can be
  provided on the commandline, or via patch files. The deck file will not be
  validated, which allows for working with incomplete or even invalid files in a pipeline.
  The functionality is imported from the [go-apiops library](https://github.com/Kong/go-apiops).
  [#939](https://github.com/Kong/deck/pull/939)
- Added a new commands `file add-tags/list-tags/remove-tags` to manage tags in a decK file. The deck file will not be
  validated, which allows for working with incomplete or even invalid files in a pipeline.
  The functionality is imported from the [go-apiops library](https://github.com/Kong/go-apiops).
  [#939](https://github.com/Kong/deck/pull/939)
- Added a new command `file add-plugins` for adding plugins to a decK file. The plugins can be
  provided on the commandline, or via config files. The deck file will not be
  validated, which allows for working with incomplete or even invalid files in a pipeline.
  The functionality is imported from the [go-apiops library](https://github.com/Kong/go-apiops).
  [#939](https://github.com/Kong/deck/pull/939)

### Fixes

- Fix Certificates & SNIs handling when running against Konnect.
  [#978](https://github.com/Kong/deck/pull/978)


## [v1.23.0]

> Release date: 2023/07/03

### Add

- Honor HTTPS_PROXY and HTTP_PROXY proxy environment variables
  [#952](https://github.com/Kong/deck/pull/952)

## [v1.22.1]

> Release date: 2023/06/22

### Fixes

- Handle missing service and route names detecting duplicates
  [#945](https://github.com/Kong/deck/pull/945)
- Update go-kong to fix a bug causing a panic when
  filling record defaults of an empty array.
  [#345](https://github.com/Kong/go-kong/pull/345)

## [v1.22.0]

> Release date: 2023/06/07

### Add

- Add indent function to support multi-line content
  [#929](https://github.com/Kong/deck/pull/929)

### Fixes

- Update go-kong to fix a bug causing wrong injection of
  defaults for non-required fields and set of record.
  [go-kong #333](https://github.com/Kong/go-kong/pull/333)
  [go-kong #336](https://github.com/Kong/go-kong/pull/336)

## [v1.21.0]

> Release date: 2023/05/31

### Add

- Add support for updating Services, Routes, and Consumers by changing their IDs,
  but retaining their names.
  [#918](https://github.com/Kong/deck/pull/918)

### Fix

- Return proper error when HTTP calls fail on validate.
  [#869](https://github.com/Kong/deck/pull/869)
- Replace old docs link in `convert` and fix its docstring.
  [#905](https://github.com/Kong/deck/pull/905)

### Misc

- Bump Go toolchain to 1.20.
  [#898](https://github.com/Kong/deck/pull/898)

## [v1.20.0]

> Release date: 2023/04/24

### Add

- Add the license type to the file package.

## [v1.19.1]

> Release date: 2023/03/21

### Add

- Add support to numeric environment variables injection
  via the `toInt` and `toFloat` functions.
  [#868](https://github.com/Kong/deck/pull/868)
- Add support to bolean environment variables injection
  via the `toBool` function.
  [#867](https://github.com/Kong/deck/pull/867)

### Fix

- Skip Consumer Groups and the related plugins when `--skip-consumers`
  [#863](https://github.com/Kong/deck/pull/863)

## [v1.19.0]

> Release date: 2023/02/21

### Add

- Add `instance_name` field to plugin schema
  [#844](https://github.com/Kong/deck/pull/844)

## [v1.18.1]

> Release date: 2023/02/10

### Fix

- Use global endpoint to retrieve Konnect Organization info.
  [#845](https://github.com/Kong/deck/pull/845)

## [v1.18.0]

> Release date: 2023/02/09

### Add

- Remove deprecated endpoint for pinging Konnect so
  to add Konnect System Accounts access token support.
  [#843](https://github.com/Kong/deck/pull/843)

## [v1.17.3]

> Release date: 2023/02/08

### Fixes

- Handle konnect runtime groups pagination properly.
  [#841](https://github.com/Kong/deck/pull/841)
- Fix workspaces validation with multiple files
  [#839](https://github.com/Kong/deck/pull/839)

## [v1.17.2]

> Release date: 2023/01/24

### Fixes

- Allow writing execution output to stdout in Konnect mode.
  [#829](https://github.com/Kong/deck/pull/829)
- Add `tags` support to Consumer Groups
  [#823](https://github.com/Kong/deck/pull/823)
- Add "update" functionality to Consumer Groups
  [#823](https://github.com/Kong/deck/pull/823)
- Do not error out when EE list endpoints are hit but no
  license is present in Kong Gateway.
  [#821](https://github.com/Kong/deck/pull/821)

## [v1.17.1]

> Release date: 2022/12/22

### Fixes

- Update go-kong to fix a bug causing wrong injection of
  defaults for arbitrary map fields.
  [go-kong #258](https://github.com/Kong/go-kong/pull/258)

## [v1.17.0]

> Release date: 2022/12/21

### Fixes

- Do not print full diffs with Create and Delete actions.
  [#806](https://github.com/Kong/deck/pull/806)
- Update go-kong to fix a bug causing misleading diffs with
  nested plugins configuration fields.
  [go-kong #244](https://github.com/Kong/go-kong/pull/244)

### Added

- Add support to Consumer Groups for both Kong Gateway and Konnect.
  [#760](https://github.com/Kong/deck/pull/760)
- Enhance error messages during dump for plugins.
  [#791](https://github.com/Kong/deck/pull/791)
- Add support to defaults injection when running
  against Konnect.
  [#788](https://github.com/Kong/deck/pull/788)

### Misc

- Remove `github.com/imdario/mergo` replace from `go.mod` and bump `go-kong` to
  `v0.34.0`
  [#809](https://github.com/Kong/deck/pull/809)

## [v1.16.1]

> Release date: 2022/11/09

### Fixes

- Fix issue with `ping` when running against Konnect
  using a PAT.
  [#790](https://github.com/Kong/deck/pull/790)

## [v1.16.0]

> Release date: 2022/11/09

### Added

- Mask DECK_ environment variable values at diff outputs.
  [#463](https://github.com/Kong/deck/pull/463)
- Add `--yes` flag to `convert` subcommand to bypass
  user confirmation and run non-interactively.
  [#775](https://github.com/Kong/deck/pull/775)
- Add support to Kong Vaults.
  [#761](https://github.com/Kong/deck/pull/761)

### Fixes

- Remove selected tags information from entities level when
  using `dump`.
  [#766](https://github.com/Kong/deck/pull/766)
- Use `kong.yaml` as default value with `convert` subcommand
  when no `--output-file` is provided.
  [#775](https://github.com/Kong/deck/pull/775)
- Add `-w` shorthand flag support to `sync`.
  [#765](https://github.com/Kong/deck/pull/765)
- Handle correctly encoded whitespaces into services' `url`
  [#755](https://github.com/Kong/deck/pull/755)
- Make sure decK can update in place consumers' username when
  `custom_id` is also set.
  [#707](https://github.com/Kong/deck/pull/707)

## [v1.15.1]

> Release date: 2022/09/27

### Fixes

- Fix decK dump when running against Konnect
  [#758](https://github.com/Kong/deck/pull/758)

## [v1.15.0]

> Release date: 2022/09/26

### Added

- Add support for Kong Gateway 3.0.
- Add support to 2.x to 3.x format migration.
  [#753](https://github.com/Kong/deck/pull/753)
- Add support to the new Kong `expressions` flavor fields.
  [#752](https://github.com/Kong/deck/pull/752)

### Fixes

- Make sure decK runs against Kong Gateway if a non-default
  `--kong-addr` is provided, even if some lingering Konnect
  environment variables are present.
  [#738](https://github.com/Kong/deck/pull/738)
- `deck validate` with `--online` flag is not currently supported
  with Konnect cloud. A clear error message is provided when this
  command is invoked.
  [#718](https://github.com/Kong/deck/pull/#718)
- Improve error message when workspace is used in Konnect mode.
  [#696](https://github.com/Kong/deck/pull/696)

## [v1.14.0]

> Release date: 2022/08/19

### Added

- Add support to multi-geo for Konnect.
  [#732](https://github.com/Kong/deck/pull/732)
- Support PAT (Personal Access Tokens) for Konnect authentication.
  [#710](https://github.com/Kong/deck/pull/710)

## [v1.13.0]

> Release date: 2022/07/14

### Fixes

- Fixed a failure when performing a diff against non-existent workspaces.
  [#702](https://github.com/Kong/deck/pull/702)

### Added

- Added support for plugin `ordering` field.
  [#710](https://github.com/Kong/deck/pull/710)

## [v1.12.4]

> Release date: 2022/07/07

### Fixes

- Make sure decK correctly includes select_tags when dumping
  from Konnect.
  [#711](https://github.com/Kong/deck/pull/711)

## [v1.12.3]

> Release date: 2022/07/05

### Added

- Add rate-limiting capabilities to Konnect client.
  [#705](https://github.com/Kong/deck/pull/705)

## [v1.12.2]

> Release date: 2022/06/06

### Fixes

- Only include referenced objects IDs in API requests to fix issue with Konnect.
  [#693](https://github.com/Kong/deck/pull/693)
- Make Oauth2Credential's `redirect_uris` field not required as it's been so
  since Kong `1.4.0`.
  [#688](https://github.com/Kong/deck/pull/688)

## [v1.12.1]

> Release date: 2022/05/13

### Added

- Log descriptive names when configuring plugins on other entities.
  [#662](https://github.com/Kong/deck/pull/662)
- Docker images now include `jq` to assist with pre-processing environment
  substitutions.
  [#660](https://github.com/Kong/deck/pull/660)

### Fixes

- Service's `enabled` field is now correctly parsed when present in config files.
  [#677](https://github.com/Kong/deck/pull/677)
- Route references to services by name are now properly handled when printing
  diffs.
  [#657](https://github.com/Kong/deck/pull/657)
- decK  uses its own user-agent header value for Konnect commands also.
  [#654](https://github.com/Kong/deck/pull/654)


## [v1.12.0]

> Release date: 2022/04/22

### Added

- Inject decK version into `User-Agent` header for requests originating from decK
  [#652](https://github.com/Kong/deck/pull/652)
- Konnect can now be configured via the main `deck` command, while
  `deck konnect` is now considered deprecated.
  [#645](https://github.com/Kong/deck/pull/645)
- Added `--skip-ca-certificates` flag. When present, decK will not attempt to
  sync CA certificates. This assists with using decK to manage multiple
  workspaces. CA certificates do not belong to a specific workspace, but can be
  seen and managed through workspaced endpoints. decK will attempt to
  remove them if they are not present in a state file when syncing a workspace,
  even if they were created for use with configuration in another workspace.
  [#617](https://github.com/Kong/deck/pull/617)
- Users can no longer set default values for fields where a default value
  does not make sense (such as fields with unique constraints, like `id`), and
  will print an error indicating the restricted field.
  [#613](https://github.com/Kong/deck/pull/613)
- Universal binaries are now available for OS X.
  [#585](https://github.com/Kong/deck/pull/585)
- Validation failures now log the name or ID of the invalid entity.
  [#588](https://github.com/Kong/deck/pull/588)

### Fixes

- Fixed unreliable diff outputs due to external dependency (gojsondiff).
  [#538](https://github.com/Kong/deck/issues/538)
- De-duplicated `select_tags` in file metadata to avoid erroneous mismatch
  reports when using `--select-tags` with multiple state files.
  [#623](https://github.com/Kong/deck/pull/623)
- Fixed several marshalling and URL construction issues for RBAC endpoint
  permissions.
  [#619](https://github.com/Kong/deck/pull/619)
  [go-kong #148](https://github.com/Kong/go-kong/pull/148)
- Fixed a regression that broke workspace creation in v1.11.0.
  [#608](https://github.com/Kong/deck/pull/608)
- Fixed a regression that broke plugins dedup feature in v1.11.0.
  [#594](https://github.com/Kong/deck/pull/594)
- Invalid YAML in state files no longer parses as an empty target state.
  [#590](https://github.com/Kong/deck/pull/590)

### Under the hood

- decK now uses Go 1.18.
  [#626](https://github.com/Kong/deck/pull/626)


## [v1.11.0]

> Release date: 2022/02/17

### Added

- decK now populates core entities defaults from their schema leveraging the Admin API.
  [#573](https://github.com/Kong/deck/pull/573)
- decK now populates plugins defaults from their schema leveraging the Admin API.
  [#562](https://github.com/Kong/deck/pull/562)
- decK prevents user interaction with any internal-only Konnect plugins.
  [#564](https://github.com/Kong/deck/pull/564)
- decK now supports mTLS Kong Admin API authentication via
  `--tls-client-cert`/`--tls-client-cert-file` and
  `--tls-client-key`/`--tls-client-key-file` flags.
  [#509](https://github.com/Kong/deck/pull/509)
- decK `validate` now has an `--online` flag enabling entity validation against
  the Admin API. This lets users validate their configuration before deploying it.
  [#502](https://github.com/Kong/deck/pull/502)

### Fixes

- decK now can set zero values (`false`, `""`, `0`) in entity fields.
  [#580](https://github.com/Kong/deck/pull/580)
- Attempting to run `sync` on multiple workspaces now returns an error
  rather than applying a single workspace to all entities.
  [#576](https://github.com/Kong/deck/pull/576)
- Skip `consumers` when using `--skip-consumers` with the `sync` command.
  [#559](https://github.com/Kong/deck/pull/559)
- The `--analytics` flag now works with the `konnect ping` command.
  [#569](https://github.com/Kong/deck/pull/569)
- Duplicate `select_tags` in separate files and `--select-tags` no longer cause errors.
  [#571](https://github.com/Kong/deck/pull/571)
- The `--header` flag is now enabled for Konnect commands.
  [#557](https://github.com/Kong/deck/pull/557)


## [v1.10.0]

> Release date: 2021/12/14
### Added

- decK can now inject cookies in its request to Kong Gateway.
  These cookies can be session cookies set by the Admin server for auth.
  `--kong-cookie-jar-path` is the cli flag that indicates path to cookie-jar file
  [#545](https://github.com/Kong/deck/pull/545)

## [v1.9.0]

> Release date: 2021/12/09

### Breaking changes

- The https://hub.docker.com/r/hbagdi/deck image is deprecated. 1.8.2 is the
  last release uploaded to it. You must switch to https://hub.docker.com/r/kong/deck
  to use 1.9.0 and future releases.

### Added

- decK now handles `_transform` state file metadata.
  [#520](https://github.com/Kong/deck/pull/520)

### Fixed

- `--select-tag` applies its tags to newly-created entities whether or not the
  tag is also present in the state file metadata.
  [#517](https://github.com/Kong/deck/pull/517)
- Timeouts in `Syncer.Run()` now return an error instead of syncing only a
  subset of the requested changes and reporting success. Downstream clients
  using decK as a library can determine when their sync attempt failed due to a
  timeout. [#529](https://github.com/Kong/deck/pull/529)

## [v1.8.2]

> Release date: 2021/09/27

### Added

- ARM64 Linux and Darwin binaries are now available.

### Fixed

- Workspace existence checks now work with workspace admins.

## [v1.8.1]

> Release date: 2021/09/22

### Fixed

- Update go-kong to v0.22.0 to fix a bug with detecting non-existent
  workspaces.

## [v1.8.0]

> Release date: 2021/09/13

### Added

- Flag `--silence-events` has been added to `deck diff` and `deck sync` commands.
  The flag disables output of events to stdout.
  [#393](https://github.com/Kong/deck/pull/393)
- decK now supports shell completions. To set up completion, please read the output
  of `deck completion --help` command. Support shells are Bash, Zsh, Fish and Powershell.
  [#416](https://github.com/Kong/deck/pull/416)
- decK now support defaults. This feature helps with avoiding repetition of common
  fields in Kong's configuration and instead specifying them in a single place.
  [#419](https://github.com/Kong/deck/pull/419)
- A new `--timeout` flag has been added to the root command to specify timeouts
  in seconds for various requests to Kong.
  [#450](https://github.com/Kong/deck/pull/450)

### Fixed

- Fix a data race with operations counters
  [#381](https://github.com/Kong/deck/pull/381)
- Correct the formats for conversion
  [#460](https://github.com/Kong/deck/pull/460)
- Updates to target entity do not result in an error anymore
  [#480](https://github.com/Kong/deck/pull/480)

### Misc

- Variety of linting updates
- Variety of refactors to improve the code health of the project
- A security policy has been added to the repository

## [v1.7.0]

> Release date: 2021/05/20

### Added

- State files now support environment variable-based templating. decK can
  substitute the value of an environment variable into an object in the state
  file. This is useful for avoiding persistent cleartext storage of sensitive
  values and populating values that vary between similar configurations.
  [#286](https://github.com/Kong/deck/pull/286)
- Sort state file objects by name, to ease comparing state files from
  similarly-configured instances that do not share object IDs.
  [#327](https://github.com/Kong/deck/pull/327)
- Added a default timeout to HTTP requests.
  [37eeec8](https://github.com/Kong/deck/commit/37eeec8606583d2ecfacb3265f7ff15921f0ab8d)
- Implemented `convert` command for converting state files between Kong Gateway
  and Konnect configuration formats. This is aimed to solving migration problem between
  on-premise Kong clusters and Konnect SaaS.
  [#330](https://github.com/Kong/deck/pull/330)
- Add `--konnect-addr` flag to set Konnect address. This can be used to target Konnect
  data-centers in geographical regions other than the US.
  [#374](https://github.com/Kong/deck/pull/374)
- Added support for document objects for Service Packages and Versions in Konnect.
  [#388](https://github.com/Kong/deck/pull/388)

### Fixed

- Fixed duplicate error message prints.
  [#317](https://github.com/Kong/deck/pull/317)
- Handle mtls-auth credential API behavior when Kong Enterprise is running in
  free mode. decK no longer treats the free mode mtls-auth behavior as a fatal
  error.
  [#321](https://github.com/Kong/deck/pull/321)
- `--select-tag` tags are now applied to credentials.
  [#282](https://github.com/Kong/deck/pull/282)
- Fix empty Service Package descriptions not syncing correctly.
  [#347](https://github.com/Kong/deck/pull/347)
- Updating certificate fields no longer deletes SNI associations.
  [#386](https://github.com/Kong/deck/pull/386)

### Misc

- Refactored utility functionality to take advantage of new features in
  go-kong.
- Added reworked usage analytics.
  [#379](https://github.com/Kong/deck/pull/379)

## [v1.6.0]

> Release date: 2021/04/08

### Added

- decK now prompts by default before overwriting existing state files when
  dumping config. Including `--yes` in args assumes yes and overwrites state
  files without prompting.
  [#285](https://github.com/Kong/deck/pull/285)

### Misc

- Removed analytics.
  [#301](https://github.com/Kong/deck/pull/301)

### Breaking changes

- Changed `github.com/blang/semver` module to `github.com/blang/semver/v4`. If
  you use decK's `file` package in other applications, you will also need to
  update the semver module used in your application.
  [#303](https://github.com/Kong/deck/pull/303)

## [v1.5.1]

> Release date: 2021/03/23

### Fixed

- Targets with identical IP and port values no longer conflict when created for
  different upstreams.
  [#280](https://github.com/Kong/deck/pull/280)
- Fixed issue where Konnect flag defaults overwrote non-Konnect flag defaults,
  which broke the `--all-workspaces` flag.
  [#290](https://github.com/Kong/deck/pull/290)
- Diff output no longer prints resource timestamp information.
  [#283](https://github.com/Kong/deck/pull/283)
- Tracebacks no longer include unwanted information specific to the build
  environment.
  [#284](https://github.com/Kong/deck/pull/284)

## [v1.5.0]

> Release date: 2021/03/06

### Added

- decK now supports Kong Konnect. Configuration for Kong Konnect can be exported,
  diffed and synced using decK. A new command `konnect` has been introduced for
  this purpose, which has 4 sub-commands: `ping`, `dump`, `diff`, and  `sync`.
  This feature in decK is currently in `alpha` state, which means there can be
  breaking changes to these commands in future releases.
- decK now supports two new Kong Enterprise resources: RBAC role and RBAC
  endpoint-permission. Special thanks to [@tjrivera](https://github.com/tjrivera)
  for this contribution. A new flag `--rbac-resources-only` has been introduced
  to manage RBAC-only configuration via decK.
  [#276](https://github.com/Kong/deck/pull/276)
- Certificates and Kong Services can now be managed separately. A check for
  existence of Certificate has been relaxed to make this possible.
  [#269](https://github.com/Kong/deck/pull/269)

## [v1.4.0]

> Release date: 2021/02/01

### Added

- deck now handles the `request_buffering` and `response_buffering` options for `Route`
  [#261](https://github.com/Kong/deck/pull/261)

### Fixes

- Updated brew syntax
  [#252](https://github.com/Kong/deck/pull/252)
- Fixed YAML/JSON file detection logic
  [#255](https://github.com/Kong/deck/pull/255)

## [v1.3.0]

> Release date: 2021/01/15

### Added

- decK will now retry sync operations that encounter a 500 error several times
  before failing completely.
  [#226](https://github.com/Kong/deck/pull/226)

### Fixed

- Fixed regression that broke workspace creation.
  [#252](https://github.com/Kong/deck/pull/252)
- Analytics failures no longer delay execution.
  [#254](https://github.com/Kong/deck/pull/254)

## [v1.2.4]

> Release date: 2021/01/06

### Fixed

- Fixed a bug that disabled verbose output.
  [#243](https://github.com/Kong/deck/pull/243)
- decK no longer considers tag order significant. This avoids unnecessary
  resource updates for Cassandra-backed clusters.
  [#240](https://github.com/Kong/deck/pull/240)

## [v1.2.3]

> Release date: 2020/11/18

### Fixed

- Sync operations now handle plugins with array configuration correctly.
  [#229](https://github.com/Kong/deck/pull/229)
- Removed unecessary permissions requirement for checking workspace existence.
  [#225](https://github.com/Kong/deck/pull/225)

## [v1.2.2]

> Release date: 2020/10/19

### Added

- decK now prints a change summary even if it encountered an error.
  [#197](https://github.com/hbagdi/deck/pull/197)
- decK now prints the ID of entities that it could not successfully sync.
  [#199](https://github.com/hbagdi/deck/pull/199)
- Issues sending analytics will now emit a panic.
  [#200](https://github.com/hbagdi/deck/pull/200)
- decK now creates the workspace specified with `--workspace` if it is not
  already present.
  [#201](https://github.com/hbagdi/deck/pull/201)
- decK prints descriptive information about duplicated entities.
  [#204](https://github.com/hbagdi/deck/pull/204)

### Fixed

- Resolved a concurrency bug during syncs.
  [#202](https://github.com/hbagdi/deck/pull/202)

## [v1.2.1]

> Release date: 2020/08/04

### Summary

decK has move under Kong's umbrella.
Due to this change, the package path has changed from `github.com/hbagdi/deck`
to `github.com/kong/deck`.
This release contains the updated `go.mod` over v1.2.0. There are no
other changes introduced in this release.

## [v1.2.0]

> Release date: 2020/08/04

### Added

- decK is now compatible with Kong 2.1:
  - New Admin API properties for entities are added.
  - Ordering of operations has changed to incorporate for new foreign-relations
    [#192](https://github.com/hbagdi/deck/pull/192)
- New flag `--db-update-propagation-delay` to add an artifical delay
  between Admin API calls. This is introduced for better compatibility with
  Cassandra backed installations of Kong.
  [#160](https://github.com/hbagdi/deck/pull/160)
  [#154](https://github.com/hbagdi/deck/pull/154)
- decK now errors out if there are invalid positional arguments supplied to
  any command.
- Stricter validation of state files.
  [#162](https://github.com/hbagdi/deck/pull/162)
- ID property of CACertificate is always exported.
  [#193](https://github.com/hbagdi/deck/pull/193)

### Fixed

- Ignore error for missing `.deck` config file
  [#168](https://github.com/hbagdi/deck/pull/168)
- Correctly populate port in Service's URL (a sugar attribute)
  [#166](https://github.com/hbagdi/deck/pull/166)
- Correct the help text for `--tls-server-name` flag
  [#170](https://github.com/hbagdi/deck/pull/170)
- Better sanitization of `--kong-addr` input
  [#171](https://github.com/hbagdi/deck/pull/171)
- Fix typos in the output of `--help`
  [#174](https://github.com/hbagdi/deck/pull/174)
- Improve language of warning message for basic-auth credentials
  [#145](https://github.com/hbagdi/deck/pull/145)
- Deduplicate `select_tags` input
  [#183](https://github.com/hbagdi/deck/pull/183)


### Enterprise-only

- Added support for managing `mtls-auth` credentials.
  [#175](https://github.com/hbagdi/deck/pull/175)
- decK now automatically creates a workspace if one does not already exist
  during a `sync` operation.
  [#187](https://github.com/hbagdi/deck/pull/187)
- Added `--workspace` flag to `ping` command. This can be used to verify
  connectivity with Kong Enterprise when running as an RBAC role with lower
  priviliges.
- New `--workspace` flag for `diff` and `sync` command to provide workspace
  via the CLI instead of state file. Workspace defined in state file will be
  overriden if this flag is provided.
- New `--skip-workspace-crud` flag to skip any workspace related operations.
  This flag can be used when running as as an RBAC role with lower priviliges.
  The content can be synced to specific workspaces but decK will not attempt
  to create or verify existence of a workspace.
  [#157](https://github.com/hbagdi/deck/pull/157)
- Additional checks for existence of workspace before performing dump or reset
  [#167](https://github.com/hbagdi/deck/pull/167)
- Improve end-user error message when workspace doesn't exist

#### Misc

- CI changed from Travis to Github Actions
- Improved code quality with addition of golangci-lint
- Default branch for the project has been changed from `master` to `main`

## [v1.1.0]

> Release date: 2020/04/05

### Added

- Added support for multiple files or directories to `-s/--state`
  flag. Use `-s` multiple times or specify multiple files/directories using
  a comma separated list.
  [#137](https://github.com/hbagdi/deck/pull/137)
- **Performance**
  decK should be much faster than before. Requests to Kong are
  now concurrent. `dump`, `sync`, `diff` and `reset`
  commands will be faster than before, by at least 2x.
- SNI entity in Kong is not supported natively supported
  [#139](https://github.com/hbagdi/deck/pull/139). Most users will not observe
  any changes. `id` and `tags` are now supported for the SNI entity in Kong.

### Under the hood

- Go has been upgraded to 1.14.1
- Alpine base image for Docker has been upgraded to 3.11
- Multiple other dependencies have also been upgraded, but these have no
  user-visible changes.

### Fixed

- Default values for `retries` in Service entity and
  `HTTPSVerifyCertificate` in Upstream entity have been removed.
  These values can be set to `0` and `false` respectively now.
  [#134](https://github.com/hbagdi/deck/issues/134)

## [v1.0.3]

> Release date: 2020/03/14

### Fixed

- Fix certificate diff for certificates with no associated snis
  [#131](https://github.com/hbagdi/deck/issues/131)

## [v1.0.2]

> Release date: 2020/02/21

### Fixed

- Fix broken `ca_certificate` entity support
  [#127](https://github.com/hbagdi/deck/pull/127)

## [v1.0.1]

> Release date: 2020/02/14

### Added

- decK now supports the `url` sugar property on Service entity.
  [#123](https://github.com/hbagdi/deck/issues/123)

## [v1.0.0]

> Release date: 2020/01/18

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

## [v0.7.2]

> Release date: 2019/12/29

### Fixed

- Kong's version is correctly parsed; v0.7.1 is unusable because
  of this bug.
  [#117](https://github.com/hbagdi/deck/issues/117)

## [v0.7.1]

> Release date: 2019/12/24

### Fixed

- Backward compatibility for credentials; tags are no longer injected into
  credentials for Kong versions below 1.4
  [#114](https://github.com/hbagdi/deck/issues/114)

## [v0.7.0]

> Release date: 2019/12/07

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

## [v0.6.2]

> Release date: 2019/11/16

### Fixed

- Service-less routes are correctly processed
  [#103](https://github.com/hbagdi/deck/issues/103)
- Plugins for routes are correctly processed
  [#104](https://github.com/hbagdi/deck/issues/104)

## [v0.6.1]

> Release date: 2019/11/08

### Fixed

- Check for workspace makes call the right endpoint
  [#94](https://github.com/hbagdi/deck/issues/94)
- Error checking is performed correctly when ensuring existence of a workspace
  [#95](https://github.com/hbagdi/deck/issues/95)
- Multiple upstream definitions are read correctly and synced up
  [#96](https://github.com/hbagdi/deck/issues/96)

## [v0.6.0]

> Release date: 2019/11/03

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

## [v0.5.2]

> Release date: 2019/09/15

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

## [v0.5.1]

> Release date: 2019/08/24

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

## [v0.5.0]

> Release date: 2019/08/18

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

## [v0.4.0]

> Release date: 2019/06/10

### Summary

This release introduces support for Kong 1.2.x.

### Breaking changes

- `strip_path` attribute of Route can now be set to false. The default value
  is now false, which was true previously.
  [#18](https://github.com/hbagdi/deck/issues/18)

### Added

- `https_redirect_status_code` attribute of Route in Kong can be set,
  and defaults to `426`.

## [v0.3.0]

> Release date: 2019/05/14

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

## [v0.2.0]

> Release date: 2019/04/01

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

## [v0.1.0]

> Release date: 2019/01/12

### Summary

Debut release of decK

[v1.24.0]: https://github.com/kong/deck/compare/v1.23.0...v1.24.0
[v1.23.0]: https://github.com/kong/deck/compare/v1.22.1...v1.23.0
[v1.22.1]: https://github.com/kong/deck/compare/v1.22.0...v1.22.1
[v1.22.0]: https://github.com/kong/deck/compare/v1.21.0...v1.22.0
[v1.21.0]: https://github.com/kong/deck/compare/v1.20.0...v1.21.0
[v1.20.0]: https://github.com/kong/deck/compare/v1.19.1...v1.20.0
[v1.19.1]: https://github.com/kong/deck/compare/v1.19.0...v1.19.1
[v1.19.0]: https://github.com/kong/deck/compare/v1.18.1...v1.19.0
[v1.18.1]: https://github.com/kong/deck/compare/v1.18.0...v1.18.1
[v1.18.0]: https://github.com/kong/deck/compare/v1.17.3...v1.18.0
[v1.17.3]: https://github.com/kong/deck/compare/v1.17.2...v1.17.3
[v1.17.2]: https://github.com/kong/deck/compare/v1.17.1...v1.17.2
[v1.17.1]: https://github.com/kong/deck/compare/v1.17.0...v1.17.1
[v1.17.0]: https://github.com/kong/deck/compare/v1.16.1...v1.17.0
[v1.16.1]: https://github.com/kong/deck/compare/v1.16.0...v1.16.1
[v1.16.0]: https://github.com/kong/deck/compare/v1.15.1...v1.16.0
[v1.15.1]: https://github.com/kong/deck/compare/v1.15.0...v1.15.1
[v1.15.0]: https://github.com/kong/deck/compare/v1.14.0...v1.15.0
[v1.14.0]: https://github.com/kong/deck/compare/v1.13.0...v1.14.0
[v1.13.0]: https://github.com/kong/deck/compare/v1.12.4...v1.13.0
[v1.12.4]: https://github.com/kong/deck/compare/v1.12.3...v1.12.4
[v1.12.3]: https://github.com/kong/deck/compare/v1.12.2...v1.12.3
[v1.12.2]: https://github.com/kong/deck/compare/v1.12.1...v1.12.2
[v1.12.1]: https://github.com/kong/deck/compare/v1.12.0...v1.12.1
[v1.12.0]: https://github.com/kong/deck/compare/v1.11.0...v1.12.0
[v1.11.0]: https://github.com/kong/deck/compare/v1.10.0...v1.11.0
[v1.10.0]: https://github.com/kong/deck/compare/v1.9.0...v1.10.0
[v1.9.0]: https://github.com/kong/deck/compare/v1.8.2...v1.9.0
[v1.8.2]: https://github.com/kong/deck/compare/v1.8.1...v1.8.2
[v1.8.1]: https://github.com/kong/deck/compare/v1.8.0...v1.8.1
[v1.8.0]: https://github.com/kong/deck/compare/v1.7.0...v1.8.0
[v1.7.0]: https://github.com/kong/deck/compare/v1.6.0...v1.7.0
[v1.6.0]: https://github.com/kong/deck/compare/v1.5.1...v1.6.0
[v1.5.1]: https://github.com/kong/deck/compare/v1.5.0...v1.5.1
[v1.5.0]: https://github.com/kong/deck/compare/v1.4.0...v1.5.0
[v1.4.0]: https://github.com/kong/deck/compare/v1.3.0...v1.4.0
[v1.3.0]: https://github.com/kong/deck/compare/v1.2.4...v1.3.0
[v1.2.4]: https://github.com/kong/deck/compare/v1.2.3...v1.2.4
[v1.2.3]: https://github.com/kong/deck/compare/v1.2.2...v1.2.3
[v1.2.2]: https://github.com/kong/deck/compare/v1.2.1...v1.2.2
[v1.2.1]: https://github.com/hbagdi/deck/compare/v1.2.0...v1.2.1
[v1.2.0]: https://github.com/hbagdi/deck/compare/v1.1.0...v1.2.0
[v1.1.0]: https://github.com/hbagdi/deck/compare/v1.0.3...v1.1.0
[v1.0.3]: https://github.com/hbagdi/deck/compare/v1.0.2...v1.0.3
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
