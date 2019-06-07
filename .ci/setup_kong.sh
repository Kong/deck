#!/bin/bash

set e
# download Kong deb

sudo apt-get update
sudo apt-get install openssl libpcre3 procps perl wget

/usr/bin/curl -sL \
  "https://kong.bintray.com/kong-deb/kong-${KONG_VERSION}.trusty.all.deb" \
  -o kong.deb

sudo dpkg -i kong.deb

sudo kong migrations bootstrap
sudo kong start
