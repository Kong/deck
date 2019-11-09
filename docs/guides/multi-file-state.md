# Using multiple files to store configuration

decK can construct a state by combining multiple JSON or YAML files inside a
directory instead of a single file.

In most use case, a single-file will suffice but you might want to use
multiple files if:
- You want to organizes the files for each service, in such a case, you
  can have a file per service, and keep a service, it's associate routes,
  plugins, and other entities in that file.
- You have a large configuration file and want to break down the large files
  into smaller digestible chunks.

You can specify an entire directory for decK to consumer using the `--state`
flag.

Under the hood, decK combines the YAML/JSON files in a very dumb fashion,
meaning it just concatenates the various arrays in the file together, before
starting to process the state.

There is no automated way of generating multiple files using decK. You will
have to export the entire configuration using the `deck dump` command and then
split the configuration into different files as you see fit for your use-case.


Please note that having the state split across different files is not same
as [distributed configuration](distributed-configuration.md).

