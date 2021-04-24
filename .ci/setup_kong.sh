#!/bin/bash

set -e
# download Kong deb

sudo apt-get update
sudo apt-get install openssl libpcre3 procps perl wget zlibc

function setup_kong(){
  SWITCH="1.3.100"

  URL="https://kong.bintray.com/kong-deb/kong-${KONG_VERSION}.xenial.all.deb"

  if [[ "$KONG_VERSION" > "$SWITCH" ]];
  then
  URL="https://kong.bintray.com/kong-deb/kong-${KONG_VERSION}.xenial.amd64.deb"
  fi

  /usr/bin/curl -sL $URL -o kong.deb
}

function setup_kong_enterprise(){
  KONG_VERSION="${KONG_VERSION#enterprise-}"
  URL="https://kong.bintray.com/kong-enterprise-edition-deb/dists/kong-enterprise-edition-${KONG_VERSION}.xenial.all.deb"
  RESPONSE_CODE=$(/usr/bin/curl -sL \
    -w "%{http_code}" \
    -u $KONG_ENTERPRISE_REPO_USERNAME:$KONG_ENTERPRISE_REPO_PASSSWORD \
    $URL -o kong.deb)
  if [[ $RESPONSE_CODE != "200" ]]; then
    echo "error retrieving kong enterprise package from ${URL}. response code ${RESPONSE_CODE}"
    exit 1 
  fi
}

if [[ $KONG_VERSION == *"enterprise"* ]]; then
  setup_kong_enterprise
else
  setup_kong
fi

sudo dpkg -i kong.deb
echo $KONG_LICENSE_DATA | sudo tee /etc/kong/license.json
export KONG_LICENSE_PATH=/tmp/license.json
export KONG_PASSWORD=kong
export KONG_ENFORCE_RBAC=on
export KONG_PORTAL=on

sudo kong migrations bootstrap
sudo kong version
sudo kong start
