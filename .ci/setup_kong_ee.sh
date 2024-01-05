#!/bin/bash

set -e

MY_SECRET_CERT='''-----BEGIN CERTIFICATE-----
MIIEczCCAlugAwIBAgIJAMw8/GAiHIFBMA0GCSqGSIb3DQEBCwUAMDYxCzAJBgNV
BAYTAlVTMRMwEQYDVQQIDApDYWxpZm9ybmlhMRIwEAYDVQQDDAlsb2NhbGhvc3Qw
HhcNMjIxMDA0MTg1MjI5WhcNMjcxMDAzMTg1MjI5WjA2MQswCQYDVQQGEwJVUzET
MBEGA1UECAwKQ2FsaWZvcm5pYTESMBAGA1UEAwwJbG9jYWxob3N0MIIBIjANBgkq
hkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3LUv6RauFfFn4a2BNSTE5oQNhASBh2Lk
0Gd0tfPcmTzJbohFwyAGskYj0NBRxnRVZdLPeoZIQyYSaiPWyeDITnXyKk3Nh9Zk
xQ03YsbZCk9jIsp78/ECdnYCCS4dpYGswu8b37dxUta+6AhEEte73ezrAhc+ZIy5
2yjcix4P5+vfhBf0EzBT8D7z+wZjji3F/A969EqphFwPz3KudkTOR6d0bQEVbN3x
cg4lcj49RzwS4sPbq6ub52QrKcx8s+d9bqC/nhHLn1HM/eef+cxROedcIWZs5RvG
mk/H+K2lcKL33gIcXgzSeunobV+8xwwoYk4GZroXjavUkgelKKjQBQIDAQABo4GD
MIGAMFAGA1UdIwRJMEehOqQ4MDYxCzAJBgNVBAYTAlVTMRMwEQYDVQQIDApDYWxp
Zm9ybmlhMRIwEAYDVQQDDAlsb2NhbGhvc3SCCQCjgi452nKnUDAJBgNVHRMEAjAA
MAsGA1UdDwQEAwIE8DAUBgNVHREEDTALgglsb2NhbGhvc3QwDQYJKoZIhvcNAQEL
BQADggIBAJiKfxuh2g0fYcR7s6iiiOMhT1ZZoXhyDhoHUUMlyN9Lm549PBQHOX6V
f/g+jqVawIrGLaQbTbPoq0XPTZohk4aYwSgU2Th3Ku8Q73FfO0MM1jjop3WCatNF
VZj/GBc9uVunOV5aMcMIs2dFU2fwH4ZjOwnv7rJRldoW1Qlk3uWeIktbmv4yZKvu
FWPuo3ks+7B+BniqKXuYkNcuhlE+iVr3kJ55qRgX1RxKo4CB3Tynkp7sikp4al7x
jlHSM9YAqvPFFMOhlU2U3SxE4CLasL99zP0ChINKp9XqzW/qo+F0/Jd4rZmddU2f
M9Cx62cc0L5IlsHLVJj3zwHZzc/ifpBUeebB98IjoQAfiRkbX0Oe/c2TxtR4o/RH
GWNeKCThdliZkXaLiOPswOV1BYfA00etorcY0CIy0aTaZgfvrYsJe4aT/hkF/JvF
tHJ/iD67m8RhLysRL/w50+quVMluUDqJps0HhKrB9wzNJWrddWhvplVeuOXEJfTM
i80W1JE4OApdISMEn56vRi+BMQMgIeYWznfyQnI4G3rUJQFMI5KzLxkvfYNTF3Ci
3Am0KaJ7X2oLOq4Qc6CM3qDkvvId61COlfJb2Fo3ETnoT66mxcb6bjtz1WTWOopm
UcmBKErRUKksINUxwuvP/VW007tXOjZH7wmiM2IW5LUZVkbhB1iE
-----END CERTIFICATE-----'''

MY_SECRET_KEY='''-----BEGIN RSA PRIVATE KEY-----
MIIEpgIBAAKCAQEA3LUv6RauFfFn4a2BNSTE5oQNhASBh2Lk0Gd0tfPcmTzJbohF
wyAGskYj0NBRxnRVZdLPeoZIQyYSaiPWyeDITnXyKk3Nh9ZkxQ03YsbZCk9jIsp7
8/ECdnYCCS4dpYGswu8b37dxUta+6AhEEte73ezrAhc+ZIy52yjcix4P5+vfhBf0
EzBT8D7z+wZjji3F/A969EqphFwPz3KudkTOR6d0bQEVbN3xcg4lcj49RzwS4sPb
q6ub52QrKcx8s+d9bqC/nhHLn1HM/eef+cxROedcIWZs5RvGmk/H+K2lcKL33gIc
XgzSeunobV+8xwwoYk4GZroXjavUkgelKKjQBQIDAQABAoIBAQDcd/nmAwvfS4iT
vTgGmDZAdsTxjXa+gSFEtTO21mUUhc5JpcLaSdGmn74DRzWI4oiz8EPlhuIEgbF/
aVGT1AEDr3o6nAGloZqD5NHgz/XbALZs+IudgLEPGI6sEO74d3LWPvg/IAYJ1A5b
xnYJxIscAyA2tHVVB+ZYcJbuORd2eSKeZSjfEfX8/DN8sKD+4DK9g/GiVlOJBG/d
bSoZmcRv3HXpnSKTvCydkxbBliD/8H7jRvkCOi2VcYT9+08rucwXc6v+q9wiQ/b7
hPdBn6KqDKRO9HPZYVkztlsdHXnthq16QyNPOk2umggfyXMIPhYBcW/dZ5xNqxBD
KiInqjbBAoGBAP3s/FS8GvFJ80pwtA0AUtB7Lo3ZASSs9dq+Dr8binPpp/1qz3hJ
Q/gRC9EP63MOWA2PK22D4qsjTfrBpqxLaalZCbCJGDGT+2knTN+qsOJ3//qI5zjj
cFEcnWcJ3bI5eLAU/2GKViyXzdGlZxBbc4zKBUSyxMAUewr3PsqEO0SJAoGBAN6C
vEYAbNuCZagtDrhhGYb+8rbSKZA7K4QjJcLTyZZ+z4ohWRW9tVyEXkszIwCRrs3y
rhHJU+z1bJOXxIin1i5tCzVlG6iNLct9dm2Z5v4hbDGNw4HhIV0l+tXrVGhkM/1v
vbRhldQA0H9iwWx+bKNS3lVLeUvYu4udmzrY74idAoGBAJ/8zQ9mZWNZwJxKXmdC
qOsKcc6Vx46gG1dzID9wzs8xjNKylX2oS9bkhpl2elbH1trUNfyOeCZz3BH+KVGt
QimdG+nKtx+lqWYbiOfz1/cYvIPR9j11r7KrYNEm+jPs2gm3cSC31IvMKbXJjSJV
PHycXK1oJWcQgGXsWfenUOBhAoGBAKezvRa9Z04h/2A7ZWbNuCGosWHdD/pmvit/
Ggy29q54sQ8Yhz39l109Xpwq1GyvYCJUj6FULe7gIo8yyat9Y83l3ZbGt4vXq/Y8
fy+n2RMcOaE3iWywMyczYtQr45gyPYT73OzAx93bJ0l7MvEEb/jAklWS5r6lgOR/
SumVayN5AoGBALLaG16NDrV2L3u/xzxw7uy5b3prpEi4wgZd4i72XaK+dMqtNVYy
KlBs7O9y+fc4AIIn6JD+9tymB1TWEn1B+3Vv6jmtzbztuCQTbJ6rTT3CFcE6TdyJ
8rYuG3/p2VkcG29TWbQARtj5ewv9p5QNfaecUzN+tps89YzawWQBanwI
-----END RSA PRIVATE KEY-----'''

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
    echo "waiting for $1"
    docker exec --user root $2 $3 && break;
    sleep 0.2
  done
}

# create docker network
docker network create $NETWORK_NAME

# Start a PostgreSQL container
docker run --rm -d --name $PG_CONTAINER_NAME \
  --network=$NETWORK_NAME \
  -p 5432:5432 \
  -e "POSTGRES_USER=$DATABASE_USER" \
  -e "POSTGRES_DB=$DATABASE_NAME" \
  -e "POSTGRES_PASSWORD=$KONG_DB_PASSWORD" \
  postgres:9.6

waitContainer "PostGres" $PG_CONTAINER_NAME pg_isready

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
  -e "MY_SECRET_CERT=$MY_SECRET_CERT" \
  -e "MY_SECRET_KEY=$MY_SECRET_KEY" \
  -p 8000:8000 \
  -p 8443:8443 \
  -p 8001:8001 \
  -p 8444:8444 \
  -p 8002:8002 \
  -p 8445:8445 \
  -p 8003:8003 \
  -p 8004:8004 \
  $KONG_IMAGE

waitContainer "Kong" $GATEWAY_CONTAINER_NAME "kong health"
