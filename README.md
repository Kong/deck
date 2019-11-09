# decK: Declarative configuration for Kong

decK provides declarative configuration and drift detection for Kong.

[![Build Status](https://travis-ci.com/hbagdi/deck.svg?branch=master)](https://travis-ci.com/hbagdi/deck)

[![asciicast](https://asciinema.org/a/238318.svg)](https://asciinema.org/a/238318)

## Table of Content

- [**Features**](#features)
- [**Compatibility**](#compatibility)
- [**Installation**](#installation)
- [**Documentation**](#documentation)
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

decK is compatible with Kong 1.x. 

## Installation

### macOS

If you are on macOS, install decK using brew:

```shell
$ brew tap hbagdi/deck
$ brew install deck
```

### Linux

If you are Linux, you can either use the Debian or RPM archive from
the Github [release page](https://github.com/hbagdi/deck/releases)
or install by downloading the binary:

```shel
$ curl -sL https://github.com/hbagdi/deck/releases/download/v0.6.0/deck_0.6.0_linux_amd64.tar.gz -o deck.tar.gz
$ tar -xf deck.tar.gz -C /tmp
$ sudo cp /tmp/deck /usr/local/bin/
```

### Docker image

Docker image is hosted on [Docker Hub](https://hub.docker.com/r/hbagdi/deck).

You can get the image with the command:

```
docker pull hbagdi/deck
```

## Documentation

Documentation can be found in the [docs](docs/README.md) directory.
You can use `--help` flag to get help in the terminal itself.

## License

decK is licensed with Apache License Version 2.0. Please read the LICENSE
file for more details.
