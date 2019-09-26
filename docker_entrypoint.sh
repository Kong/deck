#!/bin/sh
set +e

kong version &> /dev/null
KONG_AVAILABLE=$?

if [ "$KONG_AVAILABLE" -eq "0" ]
then
  kong start
  if [ "$?" -ne "0" ]
  then
    return 1
  fi
fi


deck "$@"
STATUS=$?


if [ "$KONG_AVAILABLE" -eq "0" ]
then
  kong stop
fi

exit $STATUS
