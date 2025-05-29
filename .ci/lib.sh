set -euo pipefail

readonly NETWORK_NAME=deck-test
readonly DOCKER_LABEL=com.konghq.test.deck=1
readonly KONG_PG_HOST=kong-postgres
readonly KONG_PG_USER=kong
readonly KONG_PG_DATABASE=kong
readonly KONG_PG_PASSWORD=kong
readonly KONG_PG_PORT=5432

readonly DOCKER_ARGS=(
    --label "$DOCKER_LABEL"
    --network "$NETWORK_NAME"
    --volume "$PWD/tests/integration/testdata/filters:/filters:ro"
    -e "KONG_DATABASE=postgres"
    -e "KONG_PG_HOST=$KONG_PG_HOST"
    -e "KONG_PG_PORT=$KONG_PG_PORT"
    -e "KONG_PG_USER=$KONG_PG_USER"
    -e "KONG_PG_DATABASE=$KONG_PG_DATABASE"
    -e "KONG_PG_PASSWORD=$KONG_PG_PASSWORD"
    -e "KONG_PROXY_ACCESS_LOG=/dev/stdout"
    -e "KONG_ADMIN_ACCESS_LOG=/dev/stdout"
    -e "KONG_PROXY_ERROR_LOG=/dev/stderr"
    -e "KONG_ADMIN_ERROR_LOG=/dev/stderr"
    -e "KONG_LOG_LEVEL=${KONG_LOG_LEVEL:-notice}"
    -e "KONG_CASSANDRA_CONTACT_POINTS=kong-database"
)

waitContainer() {
    local -r container=$1
    shift

    for _ in {1..100}; do
        echo "waiting for $container"
        if docker exec \
            --user root \
            "$container" \
            "$@"
        then
            return
        fi
        sleep 0.2
    done

    echo "FATAL: failed waiting for $container"
    exit 1
}

initNetwork() {
    docker network create \
        --label "$DOCKER_LABEL" \
        "$NETWORK_NAME"
}

initDb() {
    docker run \
        --rm \
        --detach \
        --name "$KONG_PG_HOST" \
        --label "$DOCKER_LABEL" \
        --network $NETWORK_NAME \
        -p "${KONG_PG_PORT}:${KONG_PG_PORT}" \
        -e "POSTGRES_USER=$KONG_PG_USER" \
        -e "POSTGRES_DB=$KONG_PG_DATABASE" \
        -e "POSTGRES_PASSWORD=$KONG_PG_PASSWORD" \
        postgres:9.6

    waitContainer "$KONG_PG_HOST" pg_isready
}

initMigrations() {
    local -r image=$1
    shift

    docker run \
        --rm \
        "${DOCKER_ARGS[@]}" \
        "$@" \
        "$image" \
            kong migrations bootstrap \
                --yes \
                --force \
                --db-timeout 30
}
