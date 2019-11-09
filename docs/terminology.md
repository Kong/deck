# Terminology

## State

State is a set of Kong configuration which is the source of truth. decK
will take the state and make Admin API calls to Kong to match configuration
stored in Kong's database with the state.
This is also referred to as the target state or desired state.

## State file(s)

State is a single file or a set of files in JSON or YAML format, which hold
the entire or a sub-set of configuration for Kong.
The files respect Kong's native declarative configuration format.

## Sync

Sync is a process of taking current configuration of Kong and making it same
as the state.

## Diff

Diff is a process of doing a dry-run sync process. It doesn't perform any
changes to Kong's database but provides a plan of entities that will be
created, deleted or updated.

## select-tag

Select-tag is a common tag in Kong's entity which is used to filter and group
related configuration when performing configuration changes.
