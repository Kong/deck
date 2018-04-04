#!/bin/bash

set e
# download Kong deb

sudo apt-get update
sudo apt-get install openssl libpcre3 procps perl wget

/usr/bin/curl -sL 'https://kong.bintray.com/kong-community-edition-deb/dists/kong-community-edition-0.13.0.trusty.all.deb' -o kong.deb -o kong.deb

sudo dpkg -i kong.deb

sudo kong migrations up
sudo kong start
