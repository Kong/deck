# decK: Declarative configuration for Kong

deck is a CLI tool to configure Kong declaratively using a single config file.

[![Build Status](https://travis-ci.com/hbagdi/deck.svg?branch=master)](https://travis-ci.com/hbagdi/deck)

## Table of Content

- [**Features**](#features)
- [**Compatibility**](#compatibility)
- [**Roadmap**](#roadmap)
- [**License**](#license)

## Features

- **Export**  
  Exisitng Kong configuration to a YAML configuration file
  This can be used to backup Kong's configuration.
- **Import**  
  Kong's database can be populated using the exported or a hand written config
  file.
- **Diff and sync capabilities**  
  deck can diff the configuration in the config file and
  the configuration in Kong's DB and then sync it as well.
  This can be used to detect config drifts or manual interventions.
- **Reverse sync**:  
  deck supports a sync the other way as well, meaning if an
  entity is created in Kong and doesn't add it to the config file,
  deck will detect the change.
- **Reset**  
  This can be used to drops all entities in Kong's DB.
- **Parallel operations**  
  All Admin API calls to Kong are executed in parallel using threads to
  speed up the sync.
- **Supported entities**
  - Routes and services
  - Upstreams and targets
  - Certificates and SNIs
  - Plugins (Global, per route and per service)

## Compatibility

deck is compatible with Kong 1.x.

## Roadmap

- Admin API authentication  
  Support providing credentials to authenticate deck against
  Admin API.
- Default attributes
  Support filling in defaults for entities and configs of plugins
  for cases when the config file doesn't contain the attribute.
- Tag entities in Kong for distributed config management.
- Add support for Consumers and custom entities in Kong.
- Support to skip sync entities like consumers.
- Complete end to end integration tests with Kong.
- Certificate encryption  
  Support in deck to fetch certificate from Vault or a cloud
  secret storage and then sync it to Kong.

## License

deck is licensed with Apache License Version 2.0. Please read the LICENSE
file for more details.