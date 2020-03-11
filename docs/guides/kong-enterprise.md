# decK and Kong Enterprise

All features of decK work with open-source and enterprise versions of Kong.

For Kong Enterprise, decK provides a few additional features leveraging the
power of enterprise features.

## Compatibility

decK is compatible with Kong Enterprise 0.35 and above.

## Entities managed by decK

decK manages only the core proxy entities in Kong Enterprise. It doesn't
manage enterprise only entities such as admins, RBAC permissions, RBAC roles
or any entities related to Developer Portal.

## RBAC

You should have authentication and RBAC configured for Kong's Admin API.
You can supply the RBAC token to decK so that decK can authenticate itself
against the Admin API:
- use `--headers` flag (example: `--headers "kong-admin-token:<your-token>"`).
  Please note that this is not a secure method. The entire command along-with
  it's flags will be logged to your shell's history file, potentially leaking
  the token. You can store the token in a file and load it as your execute the
  command ,example: `--headers "kong-admin-token:$(cat token.txt)"`
- use `DECK_HEADERS` environment variable to supply the same token, but via
  an environment variable.

It is advised that you do not use an RBAC token with super-admin privileges
with decK, and always scope down the exact permissions you need to give
decK.

## Workspaces

decK is workspace aware, meaning it can interact with multiple workspaces.

### Dump

To export configuration of a specific workspace, use the `--workspace` flag:

```
deck dump --workspace my-workspace
```

If you do not specify a flag, the configuration of `default` workspace will
be managed.


You can export configuration of all workspaces in Kong Enterprise with
the `--all-workspaces` flag:

```
deck dump --all-workspaces
```

This creates one configuration file per workspace.

### Sync

If a workspace is not present, decK will error out.
You should ensure that a workspace already exists before using decK.

`diff` and `sync` command work with workspaces and the workspace to sync
to is determined via the `_workspace` property inside the state file.

It is recommended to manage of one workspace at a time and not club
configuration of all the workspaces at the same time.

### Reset

Same as `dump` command, you can use `-workspace` to reset configuration of a
specific workspace, or use `--all-workspaces` to reset configuration of all
workspaces in Kong.
Please note that decK doesn't delete the workspace itself but deletes the
entire configuration inside the workspace.
