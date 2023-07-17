#!/usr/bin/env bash


function run_test {
  # the suite name below will only be used when running this file directly, when
  # running through "test.sh" it must be provided using the "--suite" option.
  tinitialize "decK-cli test suite" "${BASH_SOURCE[0]}"

  tchapter "deck file openapi2kong"

  ttest "accepts format 'json'" 
  ./deck file openapi2kong --format json > /dev/null < tests/cli/fixtures/mock-a-rena-oas.yml
  if [ $? -ne 0 ]; then
    tfailure "--format=json was not accepted"
  else
    tsuccess
  fi

  ttest "accepts format 'yaml'" 
  ./deck file openapi2kong --format yaml > /dev/null < tests/cli/fixtures/mock-a-rena-oas.yml
  if [ $? -ne 0 ]; then
    tfailure "--format=yaml was not accepted"
  else
    tsuccess
  fi

  ttest "--format has a default value" 
  ./deck file openapi2kong > /dev/null < tests/cli/fixtures/mock-a-rena-oas.yml
  if [ $? -ne 0 ]; then
    tfailure "unspecified --format was not accepted"
  else
    tsuccess
  fi

  tfinish
}

# No need to modify anything below this comment

# shellcheck disable=SC1090  # do not follow source
[[ "$T_PROJECT_NAME" == "" ]] && set -e && if [[ -f "${1:-$(dirname "$(realpath "$0")")/test.sh}" ]]; then source "${1:-$(dirname "$(realpath "$0")")/test.sh}"; else source "${1:-$(dirname "$(realpath "$0")")/run.sh}"; fi && set +e
run_test
