# CI-driven configuration

## or Configuration as code

decK can be, rather should be, used in a CI pipeline to push out configuration
Kong.

It is advisable to store configuration of Kong in a Git (or any other
Version Control System (VCS)) and then perform Git-ops on Kong's configuration:

- Any time a change needs to be performed, ask the developer to open a
  Pull or Merge Request, which can be reviewed by other humans.
  You should use `deck validate` and `deck diff` commands in the CI to validate
  and see if the target changes will be performed or not.
  Although unlikely, it is possible that a `deck sync` command might fail
  even if the above two pass. If this happens, a human has to intervene and
  resolve the conflict or error manually.
- Once the configuration change is merged in, the CI should execute `deck diff`
  again (to have a log of what is changing), followed by `deck sync`.

You should also have a `cronjob` in your CI or any other system, which verifies
if the source of truth, meaning Kong's database is in the exact same state as
you want it to be (the state file in VCS repository).
Unless you do this step, you do not have a truly declarative configuration
as your are configure Kong but are never verifying. The system could be
out of sync and can go undetected until another change is performed.

Anytime you use decK within such an automated environment, including a
`deck ping` command in the beginning of your script can ease debugging
in future as it usually rules out connectivity issues between decK and Kong.
