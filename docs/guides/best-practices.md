# Best practices when using decK

- Always ensure that you've one decK process running at any time. Multiple
  process will step on each other and can corrupt Kong's configuration.
- Do not mix up decK's declarative configuration with `cURL` or any other
  script. Either manage the configuration with decK or manage it with your
  home-grown script. Mixing the two on the same data-set will get cumbersome
  and will be error-prone.
- If you've a very large installation, then it is recommended to split out
  your configuration into smaller sub-set. You can find more info for it
  in the guide to practising
  [distributed configuration](distributed-configuration.md).
- Always use a pinned version of decK and Kong.
  Achieving declarative configuration is
  not easy because details matter a lot. Use a specific version of decK in
  production. If you're going to start using a new version of decK or Kong,
  please safely test the changes in a staging environment first.
- decK does not manage encryption of sensitive information, meaning private
  keys of your certificates, credentials of consumers will be stored in
  plain-text in the state file. Please be careful in how and where you store
  this file as it can have a huge impact on your security.
  You should store these in an encrypted form and provide a plain-text version
  of this on a need-by-need basis.
- If you've a very large number of consumers in your database, do not export
  or manage them using decK. Declarative configuration is for ... configuration,
  it is not meant for end-user data, which can easily grow into hundreds of
  thousands or millions.
- Always run a `deck diff` command before running a `deck sync`
  to ensure the change that is taking place.
- Adopt a [CI-driven configuration](ci-driven-configuration.md) practice.
- Always secure your Kong's Admin API with some kind of authentication method.
- Do not write the state file by hand, it will be very error-prone.
  Instead using Kong's Admin API to
  configure Kong for the first time and then export the configuration. Any
  subsequent changes should be made by manually editing the file and pushing
  the change via CI. If you're making a large change, make the change in Kong,
  export the new file, and then diff the two state files to review the changes
  being made.
- Configure a `cronjob` to run `deck diff` periodically to ensure that Kong's
  database is same as the state file checking into your Git repositories.
  Trigger an alert if decK detects a drift in the configuration.
