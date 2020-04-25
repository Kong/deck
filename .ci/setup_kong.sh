#!/bin/bash

set e
# download Kong deb

sudo apt-get update
sudo apt-get install openssl libpcre3 procps perl wget

# -_-
# Kong broke the URL because of addition of arm support.

SWITCH="1.3.100"

URL="https://kong.bintray.com/kong-deb/kong-${KONG_VERSION}.xenial.all.deb"

if [[ "$KONG_VERSION" > "$SWITCH" ]];
then
URL="https://kong.bintray.com/kong-deb/kong-${KONG_VERSION}.xenial.amd64.deb"
fi

/usr/bin/curl -sL $URL -o kong.deb

sudo dpkg -i kong.deb

sudo kong migrations bootstrap
sudo kong start
