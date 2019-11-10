# decK Documentation

decK provides declarative configuration and drift detection for Kong.

## Summary

Here is an introductory screen-cast explaining decK:

[![asciicast](https://asciinema.org/a/238318.svg)](https://asciinema.org/a/238318)

## Table of content

- [Design](#design)
- [Guides](#guides)
- [References](#references)
- [Security](#security)
- [FAQS](#frequently-asked-questions-faqs)
- [Roadmap](#roadmap)
- [Getting help](#getting-help)
- [Reporting a bug](#reporting-a-bug)

## Design

- [Terminology](terminology.md)
- [Architecture](design-architecture.md)

## Guides

- [Installation](guides/installation.md)
- [Getting started with decK](guides/getting-started.md)
- [Backup and restore of Kong's configuration](guides/backup-restore.md)
- [Configuration as code and Git-ops using decK](guides/ci-driven-configuration.md)
- [Distributed configuration with decK](guides/distributed-configuration.md)
- [Best practices for using decK](guides/best-practices.md)
- [Using decK with Kong Enterprise](guides/kong-enterprise.md)
- [Using multiple files to store configuration](guides/multi-file-state.md)

## References

The command-line `--help` flag on the main command or a sub-command (like diff,
sync, reset, etc.) shows the help text along with supported flags for those
commands.

A gist of all commands that are available in decK can be found
[here](commands.md).

## Frequently Asked Questions (FAQs)

You can find answers to FAQs [here](faqs.md).

## Roadmap

decK's roadmap is public and can be found under the open
[Github issues](https://github.com/hbagdi/deck/issues) and
[milestones](https://github.com/hbagdi/deck/milestones).

If you would like a feature to be added to decK, please open a Github issue,
or add a `+1` reaction to an existing open issues, if you feel that's
an addition you would like to see in decK.
Features with more reactions take a higher precedence usually.

## Security

decK does not offer to secure your Kong deployment but only configures it.
It encourages you to protect your Kong's Admin API with authentication but
doesn't offer such a service itself.

decK's state file can contain sensitive data such as private keys of
certificates, credentials, etc. It is left up to the user to manage
and store the state file in a secure fashion.

If you believe that you have found a security vulnerability in decK, please
submit a detailed report, along-with reproducible steps
to Harry Bagdi (email address is first name last name At gmail Dot com).
I will try to respond in a timely manner and will really appreciate it you
report the issue privately first.

## Getting help

One of the design goals of decK is deliver a good developer experience to you.
And part of it is getting the required help when you need it.
To seek help, use the following resources:
- `--help` flag gives you the necessary help in the terminal itself and should
  solve most of your problems.
- Please read through the pages under the `docs` directory of this repository.
- If you still need help, please open a
  [Github issue](https://github.com/hbagdi/deck/issues/new) to ask your
  question.
- decK has a very wide adoption by Kong's community and you can seek help
  from the larger community at [Kong Nation](https://discuss.konghq.com).

One thing I humbly ask for when you need help or run into a bug is patience.
I'll do my best to respond you at the earliest possible.

## Reporting a bug

If you believe you have run into a bug with decK, please open
a [Github issue](https://github.com/hbagdi/deck/issues/new).

If you think you've found a security issue with decK, please read the
[Security](#security) section.
