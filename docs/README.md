# decK Documentation

Following are the main commands in decK:

## dump

This command can be used to export all of Kong's configuration into a single
YAML file. All entitites are exported by default.

`--select-tag` can be used to export entities with the specific tag only.
This flag can be used in the following cases:

- If you would like to manage only a subset of entities in Kong.
- If multiple teams would like to configure Kong, one team can export
  and sync it's configuration without being aware of any other teams'
  configuration.

If you are a Kong Enterprise user, you can specify a specific workspace that
you want to export using `--workspace` flag or use `--all-workspaces` flag
to export routing configuration of all workspaces.

## diff

This command compares the content of the input file against the current
configuration fo Kong.
You can use this command for drift detection i.e. if the configuration
of Kong is out of sync with configuration of the input file.

## sync

This command will create, update or delete entities in Kong to exactly match
as described via the input file. You can use `diff` command to display
the actions that decK will take and then use `sync` commmand to actually
perform these actions.

## reset

This command will delete all the entities in Kong. Please use this
command with extreme caution as the actions are irreversible.
