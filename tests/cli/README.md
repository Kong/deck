# CLI tests

This suite of tests is not intended to do complex integration tests, but simple
unit-test like tests of the CLI flags. Testing allowed inputs, flag-combinations, etc.

## Running all tests

Using the `Makefile`;

    make test-cli

or directly;

    tests/cli/test.sh --suite "decK-cli test suite"

## Running a single file

Execute the file:

    tests/cli/mytestfile.test.sh

## Create a new test-file

To create a new testfile run:

    tests/cli/test.sh --create tests/cli/mytestfile "decK-cli test suite"

instructions will be in the generated file.
