# decK: Declarative configuration for Kong

decK provides declarative configuration and drift detection for Kong.

[![Build Status](https://github.com/kong/deck/workflows/CI%20Test/badge.svg)](https://github.com/kong/deck/actions?query=branch%3Amain+event%3Apush)
[![codecov](https://codecov.io/gh/Kong/deck/branch/main/graph/badge.svg?token=m9WNK9rFEG)](https://codecov.io/gh/Kong/deck)
[![Go Report Card](https://goreportcard.com/badge/github.com/kong/deck)](https://goreportcard.com/report/github.com/kong/deck)

[![asciicast](https://asciinema.org/a/238318.svg)](https://asciinema.org/a/238318)

## Table of Content

- [**Features**](#features)
- [**Compatibility**](#compatibility)
- [**Installation**](#installation)
- [**Documentation**](#documentation)
- [**Stale issue and pull request policy**](#stale-issue-and-pull-request-policy)
- [**License**](#license)

## Features

- **Export**  
  Existing Kong configuration to a YAML configuration file
  This can be used to backup Kong's configuration.
- **Import**  
  Kong's database can be populated using the exported or a hand written config
  file.
- **Diff and sync capabilities**  
  decK can diff the configuration in the config file and
  the configuration in Kong's DB and then sync it as well.
  This can be used to detect config drifts or manual interventions.
- **Reverse sync**  
  decK supports a sync the other way as well, meaning if an
  entity is created in Kong and doesn't add it to the config file,
  decK will detect the change.
- **Validation**  
  decK can validate a YAML file that you backup or modify to catch errors
  early on.
- **Reset**  
  This can be used to drops all entities in Kong's DB.
- **Parallel operations**  
  All Admin API calls to Kong are executed in parallel using multiple
  threads to speed up the sync process.
- **Authentication with Kong**
  Custom HTTP headers can be injected in requests to Kong's Admin API
  for authentication/authorization purposes.
- **Manage Kong's config with multiple config file**  
  Split your Kong's configuration into multiple logical files based on a shared
  set of tags amongst entities.
- **Designed to automate configuration management**  
  decK is designed to be part of your CI pipeline and can be used to not only
  push configuration to Kong but also detect drifts in configuration.

## Compatibility

decK is compatible with Kong Gateway >= 1.x and Kong Enterprise >= 0.35.

## Installation

### macOS

If you are on macOS, install decK using brew:

```shell
$ brew tap kong/deck
$ brew install deck
```

### Linux

If you are Linux, you can either use the Debian or RPM archive from
the GitHub [release page](https://github.com/kong/deck/releases)
or install by downloading the binary:

```shell
$ curl -sL https://github.com/kong/deck/releases/download/v1.41.2/deck_1.41.2_linux_amd64.tar.gz -o deck.tar.gz
$ tar -xf deck.tar.gz -C /tmp
$ sudo cp /tmp/deck /usr/local/bin/
```

### Windows

If you are on Windows, you can download the binary from the GitHub
[release page](https://github.com/kong/deck/releases) or via PowerShell:

```shell
$ curl -sL https://github.com/kong/deck/releases/download/v1.41.2/deck_1.41.2_windows_amd64.tar.gz -o deck.tar.gz
$ tar -xzvf deck.tar.gz
```

### Docker image

Docker image is hosted on [Docker Hub](https://hub.docker.com/r/kong/deck).

You can get the image with the command:

```
docker pull kong/deck
```

## Documentation

You can use `--help` flag once you've decK installed on your system
to get help in the terminal itself.

The project's documentation is hosted at
[https://docs.konghq.com/deck/overview](https://docs.konghq.com/deck/overview).

## Changelog

Changelog can be found in the [CHANGELOG.md](CHANGELOG.md) file.

## Stale issue and pull request policy

To ensure our backlog is organized and up to date, we will close issues and
pull requests that have been inactive awaiting a community response for over 2
weeks. If you wish to reopen a closed issue or PR to continue work, please
leave a comment asking a team member to do so.

## License

decK is licensed with Apache License Version 2.0.
Please read the [LICENSE](LICENSE) file for more details.
