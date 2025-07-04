# Table of Contents

- [v1.49.1](#v1491)
- [v1.49.0](#v1490)
- [v1.48.0](#v1480)
- [v1.47.1](#v1471)
- [v1.47.0](#v1470)
- [v1.46.3](#v1463)
- [v1.46.2](#v1462)
- [v1.46.1](#v1461)
- [v1.46.0](#v1460)
- [v1.45.0](#v1450)
- [v1.44.2](#v1442)
- [v1.44.1](#v1441)
- [v1.44.0](#v1440)
- [v1.43.1](#v1431)
- [v1.43.0](#v1430)
- [v1.42.1](#v1421)
- [v1.42.0](#v1420)
- [v1.41.4](#v1414)
- [v1.41.3](#v1413)
- [v1.41.2](#v1412)
- [v1.41.1](#v1411)
- [v1.41.0](#v1410)
- [v1.40.3](#v1403)
- [v1.40.2](#v1402)
- [v1.40.1](#v1401)
- [v1.40.0](#v1400)
- [v1.39.6](#v1396)
- [v1.39.5](#v1395)
- [v1.39.4](#v1394)
- [v1.39.3](#v1393)
- [v1.39.2](#v1392)
- [v1.39.1](#v1391)
- [v1.39.0](#v1390)
- [v1.38.1](#v1381)
- [v1.38.0](#v1380)
- [v1.37.0](#v1370)
- [v1.36.2](#v1362)
- [v1.36.1](#v1361)
- [v1.36.0](#v1360)
- [v1.35.0](#v1350)
- [v1.34.0](#v1340)
- [v1.33.0](#v1330)
- [v1.32.1](#v1321)
- [v1.32.0](#v1320)
- [v1.31.1](#v1311)
- [v1.31.0](#v1310)
- [v1.30.0](#v1300)
- [v1.29.2](#v1292)
- [v1.29.1](#v1291)
- [v1.29.0](#v1290)
- [v1.28.1](#v1281)
- [v1.28.0](#v1280)
- [v1.27.1](#v1271)
- [v1.27.0](#v1270)
- [v1.26.1](#v1261)
- [v1.26.0](#v1260)
- [v1.25.0](#v1250)
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

## [v1.49.1]
> Release date: 2025/06/27

### Fixed
- Sync failures due to missing names for keys, keysets.
[#1675](https://github.com/Kong/deck/pull/1675)
[go-database-reconciler #301](https://github.com/Kong/go-database-reconciler/pull/301)
- Fixed online validation for partials.
[#1666](https://github.com/Kong/deck/pull/1666)
- Fixed error message passed to end-users while using
offline (file) validation with default_lookup_tags.
[#1669](https://github.com/Kong/deck/pull/1669)
- Fixed service configuration generation when no servers
are present in an OAS document while using `deck file
openapi2kong` command.
[#1676](https://github.com/Kong/deck/pull/1676)
[go-apiops #273](https://github.com/Kong/go-apiops/pull/273)

## [v1.49.0]
> Release date: 2025/06/17

### Added
- Added default_lookup_tags for partials.
[#1653](https://github.com/Kong/deck/pull/1653)
[go-database-reconciler #291](https://github.com/Kong/go-database-reconciler/pull/291)

### Fixed
- Skipping custom-entities in dump if select tags are set, so
as to not cause sync failures or undesirable deletions.
[#1662](https://github.com/Kong/deck/pull/1662)
[go-database-reconciler #296](https://github.com/Kong/go-database-reconciler/pull/296)
- Fixed partial linking in case of nested plugins.
[go-database-reconciler #289](https://github.com/Kong/go-database-reconciler/pull/289)
- Fixed false diffs due to ca_cert order in service config.
[go-database-reconciler #288](https://github.com/Kong/go-database-reconciler/pull/288)
- Allowed empty string as path param if minLength is 0 in
`deck file openapi2kong`.
[#1660](https://github.com/Kong/deck/pull/1660)
[go-apiops #271](https://github.com/Kong/go-apiops/pull/271)
- Fixed extra service creation in absence of top-level servers block
in `deck file openapi2kong`.
[go-apiops #270](https://github.com/Kong/go-apiops/pull/270)

## [v1.48.0]
> Release date: 2025/05/30

### Added
- Added support for keys and key_sets in deck
[#1645](https://github.com/Kong/deck/pull/1645)
[go-database-reconciler #274](https://github.com/Kong/go-database-reconciler/pull/274)
[go-database-reconciler #279](https://github.com/Kong/go-database-reconciler/pull/279)
[go-database-reconciler #283](https://github.com/Kong/go-database-reconciler/pull/283)
[go-database-reconciler #286](https://github.com/Kong/go-database-reconciler/pull/286)
[go-database-reconciler #287](https://github.com/Kong/go-database-reconciler/pull/287)

### Fixed
- Gained performance boost during sync and diff operations
by caching schemas for plugins and partials.
[go-database-reconciler #285](https://github.com/Kong/go-database-reconciler/pull/285)
- Corrected request body for ConsumerGroupConsumer creation.
[go-kong #545](https://github.com/Kong/go-kong/pull/545)
- Fixed `deck file openapi2kong` command to fail fast if no 
paths are provided in OAS document, give a warning if 
explicitly set to empty.
[#1631](https://github.com/Kong/deck/pull/1631)
[go-apiops #263](https://github.com/Kong/go-apiops/pull/263)

### Chores
- Upgraded underlying alpine version for docker images to
v3.21.3
[1541](https://github.com/Kong/deck/pull/1541)

## [v1.47.1]
> Release date: 2025/05/12

### Fixed
- Fixed syncing errors faced during plugin creation
due to conflicts with global or other similarly-scoped
plugins.
[#1627](https://github.com/Kong/deck/pull/1627)
[go-database-reconciler #271](https://github.com/Kong/go-database-reconciler/pull/271)
- Improved error messaging for unsupported routes
by adding route IDs in the message.
[go-database-reconciler #257](https://github.com/Kong/go-database-reconciler/pull/257) 
- Fixed errors faced during partial apply for custom entities
[#1625](https://github.com/Kong/deck/pull/1625)
[go-database-reconciler #267](https://github.com/Kong/go-database-reconciler/pull/267)
- Bump Go version to 1.24.3
  [#1629](https://github.com/Kong/deck/pull/1629)

## [v1.47.0]
> Release date: 2025/04/29

### Added
- Extended `deck file convert` command to be used for configuration
migrations between LTS versions `2.8` and `3.4`. The command can
auto-fix the possible configurations and gives appropriate errors
or warnings for the others.
This is how it can be used: `deck file convert --from 2.8 --to 3.4
--input-file kong-28x.yaml -o kong-34x.yaml`
[#1610](https://github.com/Kong/deck/pull/1610)
- `_format_version` string can be parametrised now and works well with
`deck file merge` command as well as others.
[#1605](https://github.com/Kong/deck/pull/1605)
[go-apiops #259](https://github.com/Kong/go-apiops/pull/259)

### Fixed
- ID existence checks are limited to certificates now,
restoring sync performance.
[#1608](https://github.com/Kong/deck/pull/1608)
[go-database-reconciler #254](https://github.com/Kong/go-database-reconciler/pull/254)
- Bumped `golang.org/x/net` from 0.36.0 to 0.38.0 to account
for [CVE-2025-22872](https://github.com/advisories/GHSA-vvgc-356p-c3xw)
[#1601](https://github.com/Kong/deck/pull/1601)

## [v1.46.3]
> Release date: 2025/04/10

### Fixed
- Plugin configurations with foreign references
now work correctly, both via name and IDs.
[#1597](https://github.com/Kong/deck/pull/1597)
[go-database-reconciler #250](https://github.com/Kong/go-database-reconciler/pull/250)
- Scoped plugins with a nested entity of the same
type would return an error.
[go-database-reconciler #249](https://github.com/Kong/go-database-reconciler/pull/249)
- RBAC endpoint permissions are now paginated and not
limited to the first 100 endpoints.
[go-database-reconciler #248](https://github.com/Kong/go-database-reconciler/pull/248)
[go-kong #533](https://github.com/Kong/go-kong/pull/533)

## [v1.46.2]
> Release date: 2025/04/07

### Fixed
- Fixed false diff for ip-restriction plugin.
[#1585](https://github.com/Kong/deck/pull/1585)
[go-database-reconciler #240](https://github.com/Kong/go-database-reconciler/pull/240)
- Fixed an issue where invalid json was sent to stdout even if --json-output flag was used
[#1589](https://github.com/Kong/deck/pull/1589)
[go-database-reconciler #242](https://github.com/Kong/go-database-reconciler/pull/242)

## [v1.46.1]
> Release date: 2025/04/03

### Fixed
- Fixed an issue where entities were replaced silently when ID was present in input files and select_tags didn't match.
[#1572](https://github.com/Kong/deck/pull/1572)
[go-database-reconciler #214](https://github.com/Kong/go-database-reconciler/pull/214)
[go-kong #526](https://github.com/Kong/go-kong/pull/526)

## [v1.46.0]
> Release date: 2025/04/01

### Added
- Added support for `partials` entity in deck.
[#1570](https://github.com/Kong/deck/pull/1570)
[go-database-reconciler #215](https://github.com/Kong/go-database-reconciler/pull/215)
[go-kong #507](https://github.com/Kong/go-kong/pull/507)

### Fixed
- Updated `golang.org/x/net` from 0.34.0 to 0.36.0 to
account for vulnerability [CVE-2025-22870](https://github.com/advisories/GHSA-qxp5-gwg8-xv66)
- Fixed false errors coming for consumer creation while using
default_lookup_tags for consumer_groups, specifically when tags
on the consumer_group and consumer didn't match.
[#1576](https://github.com/Kong/deck/pull/1576)
[go-database-reconciler #228](https://github.com/Kong/go-database-reconciler/pull/228)
- Fixed `deck file kong2kic` command to correctly process
top-level plugin and route entities.
[#1555](https://github.com/Kong/deck/pull/1555)

## [v1.45.0]
> Release date: 2025/02/26

### Added
- Added `--skip-consumers-with-consumer-groups` flag to sync and 
diff commands. It ensures that list_consumers=false while querying 
a consumer-group. This is beneficial for enterprise customers who 
sync consumers separately on a different cadence but require 
frequent changes in other entity configurations. Not listing 
consumers leads to a boost in performance in such cases.
[#1545](https://github.com/Kong/deck/pull/1545)

## [v1.44.2]
> Release date: 2025/02/13

### Fixed
- Updated `golang` to version `v1.23.5` to account for
vulnerability [CVE-2022-28948](https://github.com/advisories/GHSA-hp87-p4gw-j4gq)
[#1497](https://github.com/Kong/deck/pull/1497)
[#1533](https://github.com/Kong/deck/pull/1533)

## [v1.44.1]
> Release date: 2025/02/11

### Fixed
- Fixed issue coming with using deck against open-source Kong 
gateways where operations were getting stuck due to 
custom-entities support. Custom Entities are now gated to
Enterprise gateways only.
[go-database-reconciler #202](https://github.com/Kong/go-database-reconciler/pull/202)
[#1525](https://github.com/Kong/deck/pull/1525)

## [v1.44.0]
> Release date: 2025/02/10

### Added
- Added support for consumer-group policy overrides in Kong Gateway
version 3.4+ (until next major version is released). This is enabled
via flag `--consumer-group-policy-overrides` in sync, diff and dump
commands. Consumer-group policy overrides, though supported, are a
deprecated feature in the Kong Gateway and users should consider
moving to Consumer-group scoped plugins instead. Mixing of the two
approaches should be avoided.
[#1518](https://github.com/Kong/deck/pull/1518)
[go-database-reconciler #191](https://github.com/Kong/go-database-reconciler/pull/191)
- Added support for managing `degraphql_routes` via deck for both
Kong Gateway and Konnect.
[#1505](https://github.com/Kong/deck/pull/1505)
[go-database-reconciler #154](https://github.com/Kong/go-database-reconciler/pull/154)

## [v1.43.1]
> Release date: 2025/01/29

### Fixed
- The `deck gateway apply` command added in v1.43.0 added additional
HTTP calls to discover which functions are enabled. This does not work 
well when using an RBAC user with restricted permissions. This change 
removes those additional checks and delegates the lookup of foreign 
keys for partial applications to `go-database-reconciler`.
[#1508](https://github.com/Kong/deck/pull/1508)
[go-database-reconciler #182](https://github.com/Kong/go-database-reconciler/pull/182)

## [v1.43.0]
> Release date: 2025/01/23

### Added
- Added `deck gateway apply` command that allows users 
to apply partial configuration to a running Gateway instance.
[#1459](https://github.com/Kong/deck/pull/1459)
[go-database-reconciler #143](https://github.com/Kong/go-database-reconciler/pull/143)
- Added support for private link global api endpoint for Konnect.
[#1500](https://github.com/Kong/deck/pull/1500)
[go-database-reconciler #165](https://github.com/Kong/go-database-reconciler/pull/165)
- Added flag `--skip-consumers-with-consumer-groups` for
`deck gateway dump` command. If set to true, deck skips listing 
consumers with consumer-groups, thus gaining some performance 
with large configs. It is not valid for Konnect.
[#1486](https://github.com/Kong/deck/pull/1486)

### Fixed
- Adjusted multiline string formatting in terraform resource generation.
[#1482](https://github.com/Kong/deck/pull/1482)
- Improved error messaging when mandatory flag is missing in 
`deck file convert`. [#1487](https://github.com/Kong/deck/pull/1487)
- Fixed `deck gateway dump` command that was missing associations
between consumer-groups and consumers.
[#1486](https://github.com/Kong/deck/pull/1486)
[go-database-reconciler #159](https://github.com/Kong/go-database-reconciler/pull/159)
[go-kong #494](https://github.com/Kong/go-kong/pull/494)
- Added checks for all conflicting nested configs in plugins. 
A foreign key nested under a plugin of a different scope would error out. 
This would make sure that a sync does not go through 
when wrong configurations are passed via deck.
[go-database-reconciler #157](https://github.com/Kong/go-database-reconciler/pull/157)
- Fixed req-validator config generation while using 
`deck file openapi2kong` command when both body and param schema 
are empty. [#1501](https://github.com/Kong/deck/pull/1501)
[go-apiops #244](https://github.com/Kong/go-apiops/pull/244)
- Fixed tags retention on entities while using select-tags.
[#1500](https://github.com/Kong/deck/pull/1500)
[go-database-reconciler #156](https://github.com/Kong/go-database-reconciler/pull/156)


## [v1.42.1]
> Release date: 2024/12/24

### Fixed
- Updated `golang.org/x/net` to version `v0.33.0` to account for
vulnerability [CVE-2024-45338](https://avd.aquasec.com/nvd/2024/cve-2024-45338/)
[#1481](https://github.com/Kong/deck/pull/1481)

## [v1.42.0]
> Release date: 2024/12/13

### Added
- Added a new flag `--online-entities-list` to validate the specified entities
via `deck gateway validate` command.
[#1458](https://github.com/Kong/deck/pull/1458)
- Added feature to ignore entities tagged with `konnect-managed` during
deck dump, sync and diff. This is valid for Konnect entities only.
[#1478](https://github.com/Kong/deck/pull/1478)
[go-database-reconciler #153](https://github.com/Kong/go-database-reconciler/pull/153)
- Improved speed for deck sync/diff operations involving consumer-groups 
for gw 3.9+. The underlying API call to `GET /consumer_group` is called
with query parameter `list_consumers=false`, making it faster for deck
to deal with cases where a consumer-group holds many consumers.
(#1475)[https://github.com/Kong/deck/pull/1475]
(go-kong #487)[https://github.com/Kong/go-kong/pull/487]


### Fixes
- Fixed issue where tags were not getting propagated to consumer-group plugins.
[#1478](https://github.com/Kong/deck/pull/1458)
[go-database-reconciler #151](https://github.com/Kong/go-database-reconciler/pull/151)
[go-kong #485](https://github.com/Kong/go-kong/pull/485)
- Enhanced help message for generate-imports-for-control-plane-id flag
[#1448](https://github.com/Kong/deck/pull/1448)
- Restored to using Gateway API generation in `deck file kong2kic`, rather than
Ingress API
[#1431](https://github.com/Kong/deck/pull/1431)

## [v1.41.4]
> Release date: 2024/11/26

### Fixes
- Added validation for ensuring that cookie parameters in parameter schemas are skipped 
and a warning is logged for the user while using `deck file openapi2kong` command.
[#1452](https://github.com/Kong/deck/pull/1452)
[go-apiops #255](https://github.com/Kong/go-apiops/pull/225)
- Fixed issue where creating arrays with mixed types using oneOf in OAS specifications were
failing while using `deck file openapi2kong` command. 
[#1452](https://github.com/Kong/deck/pull/1452)
[go-apiops #231](https://github.com/Kong/go-apiops/pull/231)



## [v1.41.3]
> Release date: 2024/11/25

### Fixes
- Updated Konnect authentication logic to properly handle geo rewrites. 
[#1451](https://github.com/Kong/deck/pull/1451)
[go-database-reconciler #146](https://github.com/Kong/go-database-reconciler/pull/146)
- Fixed false diffs for gateway by clearing unmatching deprecated fields from
plugin schemas. [#1451](https://github.com/Kong/deck/pull/1451)
[go-database-reconciler #145](https://github.com/Kong/go-database-reconciler/pull/145)
[go-kong #473](https://github.com/Kong/go-kong/pull/473)

## [v1.41.2]
> Release date: 2024/11/06

### Fixes
- Added fix to validate for top-level type in parameter schemas in request-validator plugin while
using `deck file openapi2kong`. [go-apiops #215](https://github.com/Kong/go-apiops/pull/215)
- Added support for defining path parameters outside REST methods for request-validation while
using `deck file openapi2kong`. [go-apiops #216](https://github.com/Kong/go-apiops/pull/216)
(#1429)[https://github.com/Kong/deck/pull/1429]

## [v1.41.1]
> Release date: 2024/10/22

### Fixes
- `deck gateway validate` for Konnect supports Konnect configs passed by CLI flags now.
Earlier, the validation was failing if control plane information was passed via CLI flags.

## [v1.41.0]
> Release date: 2024/10/21

### Added
- `deck gateway validate` command now supports Konnect. Konnect entities can be validated online with this change. 
[#1335](https://github.com/Kong/deck/pull/1335)

### Fixes
- Quoted type constraints are removed for Terraform. Type constraints in quotes were required in Terraform <= 0.11, 
It is now deprecated and will be removed in a future Terraform versions. Thus, removed them from kong2tf generation, so as
to avoid potential errors in `terraform apply`. [#1412](https://github.com/Kong/deck/pull/1412)

## [v1.40.3]
> Release date: 2024/09/26

### Fixes
- Fixed the behaviour of --konnect-addr flag in case default Konnect URL is used with it.
Earlier, using the default URL with the said flag ran the command against the gateway.
[#1398](https://github.com/Kong/deck/pull/1398)
- Bumped up go-apiops to `v0.1.38` and replaced yaml/v3 package with [Kong's own fork](https://github.com/Kong/yaml). This change allows deck commands to process OAS files with path lengths > 128 characters which was a limitation from the original yaml library.[#1405](https://github.com/Kong/deck/pull/1405) [go-apiops #208](https://github.com/Kong/go-apiops/pull/208) [Kong/yaml #1](https://github.com/Kong/yaml/pull/1)

## [v1.40.2]
> Release date: 2024/09/19

### Added
- Add support for default lookup services. [#1367](https://github.com/Kong/deck/pull/1367)
[go-database-reconciler #130](https://github.com/Kong/go-database-reconciler/pull/130)

## [v1.40.1]
> Release date: 2024/09/12

### Fixes
- Fixed the issue in `deck file kong2tf` command where users were facing a panic error with using jwt plugins when passing an empty list to cookie_names field. [#1399](https://github.com/Kong/deck/pull/1399)
- Bumped up go-apiops library. The updated lib has a fix for `deck file openapi2kong` command where parameters.required field was coming as null, if not passed by user. [#1400](https://github.com/Kong/deck/pull/1400) [go-apiops #205](https://github.com/Kong/go-apiops/pull/205)
- Bumped up go-kong library. The updated lib prevents unset plugin's configuration "record" fields to be filled with empty tables: {}
for deck files. Since, deck doesn't fill defaults anymore, this fix ensures that deck doesn't pass empty record fields while syncing plugin configurations.
[#1401](https://github.com/Kong/deck/pull/1401) [go-kong #467](https://github.com/Kong/go-kong/pull/467)

## [v1.40.0]
> Release date: 2024/09/10

### Added
- Added a new `file kong2tf` command to convert a deck file to Terraform configuration [#1391](https://github.com/Kong/deck/pull/1391), along with two command line flags:
  - `--generate-imports-for-control-plane-id`: If this is provided, import blocks will be added to Terraform to adopt existing resources.
  - `--ignore-credential-changes`: If this is provided, any credentials will be ignored until they are destroyed and recreated.

### Fixes

- Fixed the issue that was preventing a consumer to be in more than one consumer-groups [#1394](https://github.com/Kong/deck/pull/1394)
[go-database-reconciler #140](https://github.com/Kong/go-database-reconciler/pull/140)
- Fields marked as auto in schema are filled with nil in the config sent to the Control Plane. In case a field is marked as auto and is a required field, deck would throw an error if the user doesn't fill it in the declarative configuration file.
[#1394](https://github.com/Kong/deck/pull/1394) [go-database-reconciler #139](https://github.com/Kong/go-database-reconciler/pull/139)
- Defaults are no longer filled by deck. They will only be used for computing a diff, but not sent to the Control Plane.
[#1394](https://github.com/Kong/deck/pull/1394) [go-database-reconciler #133](https://github.com/Kong/go-database-reconciler/pull/133)


## [v1.39.6]
> Release date: 2024/08/22

### Fixes

- Fixed the issue where plugins scoped to consumer-groups were shown as global by deck. [#1380](https://github.com/Kong/deck/pull/1380)
[go-database-reconciler #134](https://github.com/Kong/go-database-reconciler/pull/134)

## [v1.39.5]
> Release date: 2024/08/22

### Fixes

- Fixed `deck file openapi2kong` command where parameter schema wasn't getting generated properly. [#1355](https://github.com/Kong/deck/pull/1355) [go-apiops #186](https://github.com/Kong/go-apiops/pull/186)

## [v1.39.4]
> Release date: 2024/08/01

### Fixes

- Correct --no-color flag behaviour in non-tty environments
The changes retain the default behaviour of showing colors in tty and no colors in non-tty if no flag is passed. However, on passing the --no-color=false, non-tty environments can also get colored output.[#1339](https://github.com/Kong/deck/pull/1339)
- Add validation on `deck file patch` to avoid confusing behaviour. The command intends to patch input files either via selector-value flags or command arguments. The change ensures that at least one of these is present, but not both at the same time.[#1342](https://github.com/Kong/deck/pull/1342) 
- Fix rendering for expression routes, keeping kong gateway version in consideration. [go-database-reconciler #118](https://github.com/Kong/go-database-reconciler/pull/118) [#1351](https://github.com/Kong/deck/pull/1351)

## [v1.39.3]
> Release date: 2024/07/16

### Chores

- Fixes [#1228](https://github.com/Kong/deck/issues/1282) by updating the golang version from 1.21 to 1.22, thus removing the inconsistency between decK releases' version and the one used in the project. 
  [#1336](https://github.com/Kong/deck/pull/1336)

## [v1.39.2]

> Release date: 2024/07/04

### Fixes

- Correct IPv6 targets comparison to avoid misleading diffs and failing syncs.
  [#1333](https://github.com/Kong/deck/pull/1333)
  [go-database-reconciler #109](https://github.com/Kong/go-database-reconciler/pull/109)
- Make lookups for consumer-group's consumers more performant.
  [#1333](https://github.com/Kong/deck/pull/1333)
  [go-database-reconciler #102](https://github.com/Kong/go-database-reconciler/pull/102)

## [v1.39.1]

> Release date: 2024/06/28

### Chores

- Bumped CodeGen
  [#1319](https://github.com/Kong/deck/pull/1319)

## [v1.39.0]

> Release date: 2024/06/28

### Fixes

- Bump Go version to 1.22.4
  [#1321](https://github.com/Kong/deck/pull/1321)

## [v1.38.1]

> Release date: 2024/05/29

### Fixes

- Correct bug on plugins config comparison.
  [#1311](https://github.com/Kong/deck/pull/1311)
  [go-database-reconciler #95](https://github.com/Kong/go-database-reconciler/pull/95)

## [v1.38.0]

> Release date: 2024/05/27

### Fixes

- Correct plugins config comparison to avoid misleading diffs.
  [#1306](https://github.com/Kong/deck/pull/1306)
  [go-database-reconciler #93](https://github.com/Kong/go-database-reconciler/pull/93)
- Make KIC v2 Gateway API v2 config generation deterministic.
  [#1302](https://github.com/Kong/deck/pull/1302)
- Correct tags filtering with Consumers and Consumer Groups.
  [#1293](https://github.com/Kong/deck/pull/1293)
  [go-database-reconciler #88](https://github.com/Kong/go-database-reconciler/pull/88)
- Correct tags filtering with Consumers and Consumer Groups.
  [#1293](https://github.com/Kong/deck/pull/1293)
  [go-database-reconciler #88](https://github.com/Kong/go-database-reconciler/pull/88)
- Correct typo in inso-compatible flag of `openapi2kong` command.
  [#1295](https://github.com/Kong/deck/pull/1295)
- Correct bad example on the `add-plugins` command cli help.
  [#1294](https://github.com/Kong/deck/pull/1294)
- Removed the unsupported `json-output` flag from `validate`
  [#1278](https://github.com/Kong/deck/pull/1278)
- Fixed race condition in `lint` command (bump vacuum library)
  [#1281](https://github.com/Kong/deck/pull/1281)

### Added

- `openapi2kong` now generates request-validator schemas for content-types with `+json` suffix.
  [#1303](https://github.com/Kong/deck/pull/1303)
  [go-apiops #175](https://github.com/Kong/go-apiops/pull/175)


## [v1.37.0]

> Release date: 2024/04/10

### Added

- Adds a `--konnect-compatibility` flag to `deck gateway validate` that validates Konnect
  readiness from an existing decK state.
  [#1227](https://github.com/Kong/deck/pull/1227)

## [v1.36.2]

> Release date: 2024/04/10

### Fixes

- Auto-generate rla (rate-limiting-advanced) namespaces in the `convert` subcommand
  when using Consumer Groups too.
  [#1263](https://github.com/Kong/deck/pull/1263)
- OpenAPI 2 Kong: change regex priority field to int from uint, to allow for negative
  priorities.
  [go-apiops # 162](https://github.com/Kong/go-apiops/pull/162)

## [v1.36.1]

> Release date: 2024/03/21

### Fixes

- Avoid showing bogus diffs due to endpoint_permission roles array
  not being sorted.
  [#71 go-database-reconciler](https://github.com/Kong/go-database-reconciler/pull/71)
- Do not fetch Kong version when using `validate` command.
  [#1247](https://github.com/Kong/deck/pull/1247)

## [v1.36.0]

> Release date: 2024/03/13

### Added

- This completes the namespace feature, by adding the host-based
  namespacing to the existing path-based namespacing.
  [#1241](https://github.com/Kong/deck/pull/1241)

### Fixes

- Use correct workspace when running online validation.
  [#1243](https://github.com/Kong/deck/pull/1243)
- Limit path-param names to 32 chars (`go-apiops`)
  [#153 go-apiops](https://github.com/Kong/go-apiops/pull/153)
- Correct various issues with the `file kong2kic` command.
  [#1230](https://github.com/Kong/deck/pull/1230)

## [v1.35.0]

> Release date: 2024/02/29

### Added

- Added a new `file kong2kic` command to convert a Kong declarative file to k8s
  resources for the Kong Ingress Controller (supports Ingress and Gateway resources).
  [#1050](https://github.com/Kong/deck/pull/1050)

### Fixes

- auto-generate rla (rate-limiting-advanced) namespaces in the `convert` subcommand.
  [#1206](https://github.com/Kong/deck/pull/1206)


## [v1.34.0]

> Release date: 2024/02/08

### Fixes

- Correct consumer_groups -> consumers reference and allow importing their relationships
  from upstream using `default_lookup_tags`.
  [#1212](https://github.com/Kong/deck/pull/1212)
  [go-database-reconciler #57](https://github.com/Kong/go-database-reconciler/pull/57)
- CLI fix: error out if `deck file addplugins` gets a `--selector` but no `--config`.
  [#1211](https://github.com/Kong/deck/pull/1211)

## [v1.33.0]

> Release date: 2024/02/01

### Fixes

- Correct a defect preventing TLS configuration flags from being used with Konnect.
  [#1194](https://github.com/Kong/deck/pull/1194)
  [go-database-reconciler #52](https://github.com/Kong/go-database-reconciler/pull/52)

## [v1.32.1]

> Release date: 2024/01/29

### Fixes

- Correct a defect preventing the use of plugins config deduplication when
  consumer-group scoped plugins are used.
  [#1190](https://github.com/Kong/deck/pull/1190)
  [go-database-reconciler #45](https://github.com/Kong/go-database-reconciler/pull/45)

## [v1.32.0]

> Release date: 2024/01/25

### Added

- Added a new `file namespace` command to facilitate path-based namespacing.
  [#1179](https://github.com/Kong/deck/pull/1179)

## [v1.31.1]

> Release date: 2024/01/22

### Fixes

- Fix bug when using consumer-group scoped plugins with multiple nested entities.
  [#1177](https://github.com/Kong/deck/pull/1177)
  [go-database-reconciler #45](https://github.com/Kong/go-database-reconciler/pull/45)

## [v1.31.0]

> Release date: 2024/01/22

### Fixes

- Add missing analytics for `file` commands.
  [#1171](https://github.com/Kong/deck/pull/1171)

### Added

- Add support to `default_lookup_tags` to pull entities not part of the configuration file.
  [#1124](https://github.com/Kong/deck/pull/1124)
  [#1173](https://github.com/Kong/deck/pull/1173)

## [v1.30.0]

> Release date: 2024/01/11

### Fixes

- Correct bug when consumer-group-consumer doesn't have an username.
  [#1113](https://github.com/Kong/deck/pull/1113)
- Improve deprecation warnings to reduce upgrade friction and show warning when reading STDIN from terminal.
  [#1115](https://github.com/Kong/deck/pull/1115)
- 'file openapi2kong': Server ports will now be properly parsed, 32767 to 65535 are now accepted.
  [apiops #105](https://github.com/Kong/go-apiops/pull/105)

### Added

- 'file openapi2kong': will now generate OpenIDConnect plugins.
  [apiops #107](https://github.com/Kong/go-apiops/pull/107)

### Refactored

- Moved the database reconciler to its own project.
  [#1109](https://github.com/Kong/deck/pull/1109)

## [v1.29.2]

> Release date: 2023/11/08

### Fixes

- Avoid unnecessary Konnect API call to retrieve its version.
  [#1095](https://github.com/Kong/deck/pull/1095)
- Correct default values when using `gateway dump`.
  [#1094](https://github.com/Kong/deck/pull/1094)

## [v1.29.1]

> Release date: 2023/11/07

### Fixes

- Correct a bug preventing logins with Konnect in the EU region.
  [#1089](https://github.com/Kong/deck/pull/1089)

## [v1.29.0]

> Release date: 2023/11/03

### Added

- Add support for konnect AU region.
  [#1082](https://github.com/Kong/deck/pull/1082)

### Fixes

- Resolved an issue in the `deck file validate` and `deck gateway validate` commands that prevented them from properly processing the provided file arguments.
  [#1084](https://github.com/Kong/deck/pull/1084)


## [v1.28.1]

> Release date: 2023/11/02

### Fixes

- Old cli commands would also output to stdout by default. Now back to default "kong.yaml".
  [#1073](https://github.com/Kong/deck/pull/1073)
- Deprecation warnings were send to stdout, mixing warnings with intended output. Now going to stderr.
  [#1075](https://github.com/Kong/deck/pull/1075)


## [v1.28.0]

> Release date: 2023/10/31

> __IMPORTANT__: _The top-level CLI commands have been restructured. There is backward
compatibility, but in the future that will be removed. Please update to the new structure
(see 'changes')._

> __IMPORTANT__: _The recently added decK command `deck file openapi2kong` implemented different techniques for
generating decK configuration from OpenAPI spec files then the legacy `inso` tool. In particular, entity names and identifiers
were generated differently in the more recent implementation. For existing `inso` users, this may cause issues with migrating
to the new tool as names and IDs are used by Kong Gateway to identify entities. In response we have added a
`--inso-compatible` flag to the `deck file openapi2kong` command to support a smoother migration for these users. For more
information on this and other APIOps commands, see the
[go-apiops documentation page](https://github.com/Kong/go-apiops/tree/main/docs).

### Added

- Allow arrays to be specified on the `file patch` CLI command.
  [#1056](https://github.com/Kong/deck/pull/1056)

### Fixes

- Do not overwrite `created_at` for existing resources when running `sync` command.
  [#1061](https://github.com/Kong/deck/pull/1061)
- `deck file openapi2kong` creates names for entities that differ from the older `inso`
  tool. This has been fixed, but requires the new `--inso-compatible` flag to not be breaking.
  Adding that flag will also skip id generation.
  [#962](https://github.com/Kong/deck/pull/962)

### Changes

- Add analytics for local operations
  [#1051](https://github.com/Kong/deck/pull/1051)
- The top-level CLI commands have been restructured. All commands now live under 2
  subcommands (`gateway` and `file`) to clarify their use and (in the future) reduce the clutter of
  the many global flags only relevant to a few commands.
  Using the old commands will still work but
  will print a deprecation notice. Please update your usage to the new commands.
  The new commands are more unix-like;

  - default to `stdin`/`stdout` and no longer to "`kong.yaml`"
  - the `-s` / `--state` flag is gone, files can be listed without the flag
  - the `--online` flag for `validate` is gone; use `gateway validate` for online, `file validate` for local.

  PR [#962](https://github.com/Kong/deck/pull/962)

## [v1.27.1]

> Release date: 2023/09/27

### Fixes

- Fix inconsistency when managing multiple consumers having equal `username` and `custom_id` fields.
  [#1037](https://github.com/Kong/deck/pull/1037)
- Correct a bug preventing the deprecated `--konnect-runtime-group-name` flag to work properly.
  [#1036](https://github.com/Kong/deck/pull/1036)


## [v1.27.0]

> Release date: 2023/09/25

### Added

- Add `--konnect-control-plane-name` flag and deprecate `--konnect-runtime-group-name`
  [#1000](https://github.com/Kong/deck/pull/1000)

### Fixes

- Bumped `go-apiops` to `v0.1.21` to include various fixes on APIOps functionality
  [#1029](https://github.com/Kong/deck/pull/1029)

## [v1.26.1]

> Release date: 2023/09/07

### Fixes

- Raise an error if state files have different Runtime Groups
  [#1014](https://github.com/Kong/deck/pull/1014)
- Correct consumers validation when `custom_id` is used
  [#1012](https://github.com/Kong/deck/pull/1012)
- Remove hardcoded default value for Routes' `strip_path` field. Defaults are pulled via
  API anyway.
  [#999](https://github.com/Kong/deck/pull/999)

## [v1.26.0]

> Release date: 2023/08/09

### Added

- Added support for scoping plugins to Consumer Groups for both Kong Gateway and Konnect.
  [#963](https://github.com/Kong/deck/pull/963)
  [#959](https://github.com/Kong/deck/pull/959)

### Fixes

- Remove fallback mechanism formely used to authenticate with either "old" or "new" Konnect.
  [#995](https://github.com/Kong/deck/pull/995)

## [v1.25.0]

> Release date: 2023/07/28

### Added

- Added a new command `file render` to render a final decK file. This will result in a file representing
  the state as it would be synced online.
  [#963](https://github.com/Kong/deck/pull/963)
- Added a new flag `--format` to `file convert` to enable JSON output.
  [#963](https://github.com/Kong/deck/pull/963)

### Fixes

- Use same interface to pull Consumer Groups with Kong Gateway and Konnect.
  This will help solving the issue of using tags with Consumer Groups when running against Konnect.
  [#984](https://github.com/Kong/deck/pull/984)
- Fix Consumers handling when a consumer's `custom_id` is equal to the `username` of another consumer.
  [#986](https://github.com/Kong/deck/pull/986)
- Avoid misleading diffs when configuration file has empty tags.
  [#985](https://github.com/Kong/deck/pull/985)

## [v1.24.0]

> Release date: 2023/07/24

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

[v1.49.1]: https://github.com/Kong/deck/compare/v1.49.0...v1.49.1
[v1.49.0]: https://github.com/Kong/deck/compare/v1.48.0...v1.49.0
[v1.48.0]: https://github.com/Kong/deck/compare/v1.47.1...v1.48.0
[v1.47.1]: https://github.com/Kong/deck/compare/v1.47.0...v1.47.1
[v1.47.0]: https://github.com/Kong/deck/compare/v1.46.3...v1.47.0
[v1.46.3]: https://github.com/Kong/deck/compare/v1.46.2...v1.46.3
[v1.46.2]: https://github.com/Kong/deck/compare/v1.46.1...v1.46.2
[v1.46.1]: https://github.com/Kong/deck/compare/v1.46.0...v1.46.1
[v1.46.0]: https://github.com/Kong/deck/compare/v1.45.0...v1.46.0
[v1.45.0]: https://github.com/Kong/deck/compare/v1.44.2...v1.45.0
[v1.44.2]: https://github.com/Kong/deck/compare/v1.44.1...v1.44.2
[v1.44.1]: https://github.com/Kong/deck/compare/v1.44.0...v1.44.1
[v1.44.0]: https://github.com/Kong/deck/compare/v1.43.1...v1.44.0
[v1.43.1]: https://github.com/Kong/deck/compare/v1.43.0...v1.43.1
[v1.43.0]: https://github.com/Kong/deck/compare/v1.42.1...v1.43.0
[v1.42.1]: https://github.com/Kong/deck/compare/v1.42.0...v1.42.1
[v1.42.0]: https://github.com/Kong/deck/compare/v1.41.4...v1.42.0
[v1.41.4]: https://github.com/Kong/deck/compare/v1.41.3...v1.41.4
[v1.41.3]: https://github.com/Kong/deck/compare/v1.41.2...v1.41.3
[v1.41.2]: https://github.com/Kong/deck/compare/v1.41.1...v1.41.2
[v1.41.1]: https://github.com/Kong/deck/compare/v1.40.0...v1.41.1
[v1.41.0]: https://github.com/Kong/deck/compare/v1.40.3...v1.41.0
[v1.40.3]: https://github.com/Kong/deck/compare/v1.40.2...v1.40.3
[v1.40.2]: https://github.com/Kong/deck/compare/v1.40.1...v1.40.2
[v1.40.1]: https://github.com/Kong/deck/compare/v1.40.0...v1.40.1
[v1.40.0]: https://github.com/Kong/deck/compare/v1.39.6...v1.40.0
[v1.39.6]: https://github.com/Kong/deck/compare/v1.39.5...v1.39.6
[v1.39.5]: https://github.com/Kong/deck/compare/v1.39.4...v1.39.5
[v1.39.4]: https://github.com/Kong/deck/compare/v1.39.3...v1.39.4
[v1.39.3]: https://github.com/Kong/deck/compare/v1.39.2...v1.39.3
[v1.39.2]: https://github.com/kong/deck/compare/v1.39.1...v1.39.2
[v1.39.1]: https://github.com/kong/deck/compare/v1.39.0...v1.39.1
[v1.39.0]: https://github.com/kong/deck/compare/v1.38.1...v1.39.0
[v1.38.1]: https://github.com/kong/deck/compare/v1.38.0...v1.38.1
[v1.38.0]: https://github.com/kong/deck/compare/v1.37.0...v1.38.0
[v1.37.0]: https://github.com/kong/deck/compare/v1.36.2...v1.37.0
[v1.36.2]: https://github.com/kong/deck/compare/v1.36.1...v1.36.2
[v1.36.1]: https://github.com/kong/deck/compare/v1.36.0...v1.36.1
[v1.36.0]: https://github.com/kong/deck/compare/v1.35.0...v1.36.0
[v1.35.0]: https://github.com/kong/deck/compare/v1.34.0...v1.35.0
[v1.34.0]: https://github.com/kong/deck/compare/v1.33.0...v1.34.0
[v1.33.0]: https://github.com/kong/deck/compare/v1.32.1...v1.33.0
[v1.32.1]: https://github.com/kong/deck/compare/v1.32.0...v1.32.1
[v1.32.0]: https://github.com/kong/deck/compare/v1.31.1...v1.32.0
[v1.31.1]: https://github.com/kong/deck/compare/v1.31.0...v1.31.1
[v1.31.0]: https://github.com/kong/deck/compare/v1.30.0...v1.31.0
[v1.30.0]: https://github.com/kong/deck/compare/v1.29.2...v1.30.0
[v1.29.2]: https://github.com/kong/deck/compare/v1.29.1...v1.29.2
[v1.29.1]: https://github.com/kong/deck/compare/v1.29.0...v1.29.1
[v1.29.0]: https://github.com/kong/deck/compare/v1.28.1...v1.29.0
[v1.28.1]: https://github.com/kong/deck/compare/v1.28.0...v1.28.1
[v1.28.0]: https://github.com/kong/deck/compare/v1.27.1...v1.28.0
[v1.27.1]: https://github.com/kong/deck/compare/v1.27.0...v1.27.1
[v1.27.0]: https://github.com/kong/deck/compare/v1.26.1...v1.27.0
[v1.26.1]: https://github.com/kong/deck/compare/v1.26.0...v1.26.1
[v1.26.0]: https://github.com/kong/deck/compare/v1.25.0...v1.26.0
[v1.25.0]: https://github.com/kong/deck/compare/v1.24.0...v1.25.0
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
