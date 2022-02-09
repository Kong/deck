#!/bin/bash

set -e

KONG_IMAGE=${KONG_IMAGE}
NETWORK_NAME=deck-test

PG_CONTAINER_NAME=pg
DATABASE_USER=kong
DATABASE_NAME=kong
KONG_DB_PASSWORD=kong
KONG_PG_HOST=pg

GATEWAY_CONTAINER_NAME=kong

# create docker network
docker network create $NETWORK_NAME

waitContainer() {
  for try in {1..100}; do
    echo "waiting for $1.."
    nc localhost $2 && break;
    sleep $3
  done
}

# Start a PostgreSQL container
docker run --rm -d --name $PG_CONTAINER_NAME \
  --network=$NETWORK_NAME \
  -p 5432:5432 \
  -e "POSTGRES_USER=$DATABASE_USER" \
  -e "POSTGRES_DB=$DATABASE_NAME" \
  -e "POSTGRES_PASSWORD=$KONG_DB_PASSWORD" \
  postgres:9.6

waitContainer "PostgreSQL" 5432 0.2

# Prepare the Kong database
docker run --rm --network=$NETWORK_NAME \
  -e "KONG_DATABASE=postgres" \
  -e "KONG_PG_HOST=$KONG_PG_HOST" \
  -e "KONG_PG_PASSWORD=$KONG_DB_PASSWORD" \
  -e "KONG_PASSWORD=$KONG_DB_PASSWORD" \
  -e "KONG_LICENSE_DATA=$KONG_LICENSE_DATA" \
  $KONG_IMAGE kong migrations bootstrap

# Start Kong Gateway EE
docker run -d --name $GATEWAY_CONTAINER_NAME \
  --network=$NETWORK_NAME \
  -e "KONG_DATABASE=postgres" \
  -e "KONG_PG_HOST=$KONG_PG_HOST" \
  -e "KONG_PG_USER=$DATABASE_USER" \
  -e "KONG_PG_PASSWORD=$KONG_DB_PASSWORD" \
  -e "KONG_PROXY_ACCESS_LOG=/dev/stdout" \
  -e "KONG_ADMIN_ACCESS_LOG=/dev/stdout" \
  -e "KONG_PROXY_ERROR_LOG=/dev/stderr" \
  -e "KONG_ADMIN_ERROR_LOG=/dev/stderr" \
  -e "KONG_ADMIN_LISTEN=0.0.0.0:8001" \
  -e "KONG_PORTAL_GUI_URI=127.0.0.1:8003" \
  -e "KONG_ADMIN_GUI_URL=http://127.0.0.1:8002" \
  -e "KONG_LICENSE_DATA=$KONG_LICENSE_DATA" \
  -p 8000:8000 \
  -p 8443:8443 \
  -p 8001:8001 \
  -p 8444:8444 \
  -p 8002:8002 \
  -p 8445:8445 \
  -p 8003:8003 \
  -p 8004:8004 \
  $KONG_IMAGE

waitContainer "Kong" 8001 0.2