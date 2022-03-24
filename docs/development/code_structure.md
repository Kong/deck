# deck codebase overview

This is an overview of the deck codebase based on version 1.10. It covers the subsystems used when processing a command in the order they run and provides an example of adding a new resource type.

## deck command execution path

To show how deck processes a command, we'll follow the life of a `deck sync` request.

### cmd

cmd holds the cobra/viper code for the various deck commands. Each subcommand (dump, sync, etc.) has a file named after it that contains the function run for that command.

root.go contains generic deck args, common.go contains utility functions and entrypoint functions, and there's a parallel set of konnect_<cmd>.go for the Konnect equivalents.

Not much of interest here, but it's where you go if you need to add a new argument or want to figure out the entrypoint into the rest of the code for a given command.

sync.go contains [the cobra function for sync](https://github.com/Kong/deck/blob/v1.10.0/cmd/sync.go#L16-L35), which calls [syncMain()](https://github.com/Kong/deck/blob/v1.10.0/cmd/common.go#L56-L164). `syncMain()` is responsible for invoking everything else we need to complete a sync.

### Building state

sync requires that deck read a state file (`kong.yaml` by default) and gather state from Kong's admin API.

#### Reading from a file

`syncMain()` first [reads the state file](https://github.com/Kong/deck/blob/v1.10.0/cmd/common.go#L60), returning a [`file.Content` struct](https://github.com/Kong/deck/blob/v1.10.0/file/types.go#L597-L618), a struct containing slices of other Go structures for every Kong entity deck supports. For example, [FService](https://github.com/Kong/deck/blob/v1.10.0/file/types.go#L33-L42) contains a go-kong Service and references to other file types that use this service: the FRoute slice are routes that use this service, and the FPlugins slice are plugins applied to it.

This nesting within the Content struct matches the structure of state files, as we're basically just [running a straight YAML unmarshal into it](https://github.com/Kong/deck/blob/v1.10.0/file/readfile.go#L83-L113), with some preprocessing to fill in template values.

#### Reading from the admin API

To read from the admin API, we first perform some prework to create a Kong client, checking to see if the requested workspace exists and creating it if not. We then [call `fetchCurrentState()`](https://github.com/Kong/deck/blob/v1.10.0/cmd/common.go#L119), which in turn [calls `dump.Get()`](https://github.com/Kong/deck/blob/v1.10.0/dump/dump.go#L244-L270).

`dump.Get()` calls several functions, which fetch various sections of the Kong configuration. Most configuration is [retrieved in `getProxyConfiguration()`](https://github.com/Kong/deck/blob/v1.10.0/dump/dump.go#L146-L219), which in turn calls [various wrapper functions around go-kong type `List()` functions](https://github.com/Kong/deck/blob/v1.10.0/dump/dump.go#L272-L293). It returns a [raw state struct](https://github.com/Kong/deck/blob/v1.10.0/utils/types.go#L26-L52), which is similar to a `file.Content`, but it only contains go-kong structs, without any nesting for relationships.

To find those relationships, we will build a [state.KongState](https://github.com/Kong/deck/blob/27094f2333fc3aa380f5c926e88600dafc0db019/state/state.go#L15-L41), which contains a memdb table for each type deck supports. [`state.Get()` loops over each entity in the raw state and calls its database add function](https://github.com/Kong/deck/blob/v1.10.0/state/builder.go#L9-L199). These each [perform some basic validation before inserting the entity](https://github.com/Kong/deck/blob/v1.10.0/state/service.go#L37-L63).

#### Massaging the file data

We don't yet have comparable formats. We need to transform our `file.Content` into the same state database we have for the admin API entities. [`file.Get()` transforms a `file.Content` into a `RawState` and then runs the state builder to create `KongState`](https://github.com/Kong/deck/blob/v1.10.0/cmd/common.go#L139-L153)

[`file.Get()` is similar to `state.Get()`](https://github.com/Kong/deck/blob/v1.10.0/file/builder.go#L36-L75), but with [additional logic to handle the entity relationships](https://github.com/Kong/deck/blob/v1.10.0/file/builder.go#L531-L585). We don't need this for dumping state from the admin API, because those resources come with their relationship IDs already set, but when building from a file, we're generating IDs on the fly and need to insert them as we go.

### Comparing state

Now that we both the target state (from the file) and current state (from the admin API) in the same format, we can compare the two to find the difference. This starts from `cmd.performDiff`, which [builds a new `diff.Syncer` from the states and runs its `Solve()` function](https://github.com/Kong/deck/blob/v1.10.0/cmd/common.go#L201-L221).

`Solve()` [sets up some stats counters and builds an anonymous function to log and perform individual CRUD events](https://github.com/Kong/deck/blob/v1.10.0/diff/diff.go#L323-L384), which it uses to invoke `syncer.Run()`.

`Run()` handles many CRUD events in parallel. It [runs `n` event consumers](https://github.com/Kong/deck/blob/v1.10.0/diff/diff.go#L212-L220) and [one event producer](https://github.com/Kong/deck/blob/v1.10.0/diff/diff.go#L222-L231).

The producer determines which entities need changes and emits CRUD events for each. It [first determines resources that need to be created or updated and then resources that need to be deleted](https://github.com/Kong/deck/blob/v1.10.0/diff/diff.go#L130-L141). These both loop over entities in one or the other state: for create/update we loop over entities in the target state and see if they're already in the current state, and emit an event if not, and for deletes we loop over entities in the current state to see if they're in the target state, e.g. [for services](https://github.com/Kong/deck/blob/v1.10.0/types/service.go#L112-L159).

### Updating configuration

Emitted events are [sent to the event channel](https://github.com/Kong/deck/blob/v1.10.0/diff/diff.go#L169-L178), where they are then [handled by one of the consumers](https://github.com/Kong/deck/blob/v1.10.0/diff/diff.go#L267-L283).

The consumers handle events by [wrapping `Do()` (the anonymous function from earlier, which processes the event and logs its change) and `postprocessor.Do()`](https://github.com/Kong/deck/blob/v1.10.0/crud/registry.go#L105-L128) in retries.

`Do()` will call the syncer's `processor.Do()`. `processor` is a `crud.Registry`, and will [call a CRUD function based on the operation resource type and action](https://github.com/Kong/deck/blob/v1.10.0/crud/registry.go#L105-L128). The individual type CRUD functions build an appropriate struct from the event object and use a go-kong client to issue the admin API HTTP call, e.g. [to create a service](https://github.com/Kong/deck/blob/v1.10.0/types/service.go#L29-L37).

Having processed all entities, the producer will close down the event channel and `syncer.Run()` will return any errors it encountered. After that, we're done!

## Other commands

Sync conveniently hits all deck's major code paths. The other major commands omit part of what we do in sync.

### diff

Same as sync, but with a syncer option that indicates it's a dry run. Dry run syncers don't actually call `processor.Do()`.

### reset

Sync, but with an empty target state.

### dump

Only collects the current state. Instead of comparing it and resolving a diff, [calls `KongStateToFile()` to get a `file.Content`, which it then marshals to YAML and writes to disk.

## Adding code for a new entity

Any new entity requires adding per-entity functions to each of the subsystems mentioned above, and then needs to call those entity functions from the functions that process all entities. For an example, I'll use the [PR that added mtls-auth support](https://github.com/Kong/deck/pull/175/files)

### diff

We [add mtls-auth calls](https://github.com/Kong/deck/commit/b1b6633f1296346168dcc2b8d003367f9fcff03c#diff-9b5d2fa5162a9c23e7b0c96446b28d54aa4c6676a697d2eb26073a5a76db602f) to `syncer.createUpdate()` and `syncer.delete()`.

These calls are implemented in a [diff package source file named after the entity](https://github.com/Kong/deck/commit/b1b6633f1296346168dcc2b8d003367f9fcff03c#diff-3bafe2ee4faaf2a190f3b0847648ea0fab2e0a3e7188934af5b61892c2d754f8).

A `deleteEntities()` function loops over all the entities of that type in the current state and calls `deleteEntity()` on each. The latter function is a bit of a misnomer, since it is not guaranteed to delete that entity: it will perform a lookup against the target state attached to the syncer and delete only if the entity is not found there.

`createUpdateEntities()` essentially does the reverse, with an additional call to `entity.EqualWithOpts()` to see if a found current entity matches the target entity. It emits an event if not.

Lastly, we add post-processing functions. These update the current state by adding the new entity to (or deleting a deleted entity from) the current state after its event succeeds.

### dump

Nothing too exciting here. The [added code](https://github.com/Kong/deck/commit/b1b6633f1296346168dcc2b8d003367f9fcff03c#diff-61c9587f2d7a5abbf52f69e0983b087ec5309c08721a20bc4155f57e1119fb6e) is basically just calling go-kong `List()` functions and adding the resulting structs to a slice.

### file

File statebuilder code [loops over entities in a file and adds them to a raw state](https://github.com/Kong/deck/commit/b1b6633f1296346168dcc2b8d003367f9fcff03c#diff-d0401391a7f45651a98e489a540d95eb1c3cda5d8e077414f234f15c06a8bc3e). The loop may be within the function for handling some other entity, e.g. here since credentials in the file will be under some consumer object, we add our mtls-auths from within the `consumers()` functions. Other entity types may need their own standalone function, which needs to be added to `build()`

[schema.go](https://github.com/Kong/deck/commit/b1b6633f1296346168dcc2b8d003367f9fcff03c#diff-8bb47c65c57c6acad7a9514929db77cc6e72064494364b76446434aa2586384b) gets new JSON schemas so we can (un)marshal to and from YAML.

[types.go](https://github.com/Kong/deck/commit/b1b6633f1296346168dcc2b8d003367f9fcff03c#diff-cfa628c722a6fefd09aeea50d3b473c589ca0a6c9263d4d0ac13c8eeb4dee809) either gets a new FEntity type (top-level entities) or updates an existing type (nested entities).

[writer.go](https://github.com/Kong/deck/commit/b1b6633f1296346168dcc2b8d003367f9fcff03c#diff-26fc8adff8d1657ea6fea89845726cec4c50879b6ea3f994b75f1d8b4d8199fe) adds entity handlers to `KongStateToFile()`, nested under its parent. This also strips out content we don't want, namely timestamps and IDs.

### solver

The solver package contains [functions that issue HTTP calls](https://github.com/Kong/deck/commit/b1b6633f1296346168dcc2b8d003367f9fcff03c#diff-0a5e7eafe1f13bf96aebd957f5c02e390e45408293d30c613cfed18a71e54c89) for an entity, with some validation. These are [registered](https://github.com/Kong/deck/commit/b1b6633f1296346168dcc2b8d003367f9fcff03c#diff-a62ce7bf064666bd923b3b06a05f40f5f63892cf0e70004096b1e61d6053c6c0) for their entity name, so that they're the functions called when the solver processes an event with Kind `entity-name`

Note that this code has since moved into the types package, but it's functionally the same.

### state

Most of what we add to state is a go-memdb [database schema and query functions](https://github.com/Kong/deck/commit/b1b6633f1296346168dcc2b8d003367f9fcff03c#diff-4e2f15e9ac7b96725f24802edd6043625922523d6a9efde81917625e9434882a).

Entities will have generic create/update/etc. functions that manipulate rows based on that specific entity (usually operating on an ID and/or name-like field) and possibly join queries. Credentials, for example, include a join query function to retrieve credentials for a given consumer.

Because most credentials are similar, we define [a generic schema for them](https://github.com/Kong/deck/blob/df3b6fae56f7cb6694e8c906a3474999f2af20b9/state/credentials.go#L24-L49). This showcases several types of [go-memdb indexers](https://pkg.go.dev/github.com/hashicorp/go-memdb#Indexer), e.g. the `StringFieldIndexer` (basic string equality for a given column in a row) and a `MethodIndexer`, which creates an index based on calling some receiver function on the entity. You can define your own Indexers, which is [what we've done for MethodIndexer](https://github.com/Kong/deck/blob/df3b6fae56f7cb6694e8c906a3474999f2af20b9/state/indexers/methodIndexer.go).

Note that because Go reasons, these methods [do need to be defined per entity](https://github.com/Kong/deck/blob/6ede1a8b4257a0bb8d8c8551687f61663ed37e20/state/types.go#L1148-L1153).

Those method definitions, the state type, and type equality functions are [added to state/types.go](https://github.com/Kong/deck/commit/b1b6633f1296346168dcc2b8d003367f9fcff03c#diff-d749ed180037f91f4f85e9d21a8a16ff32c26408a1025b16b26681cfb24c2b22).
