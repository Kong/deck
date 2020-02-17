# Deduplicate plugin configuration

In some use cases, you might want to create a number of plugins associated with
different entities in Kong but with the same configuration. In such a case,
if you change anything in the configuration of the plugin, you will have to
repeat it for each instance of the plugin.

In other use cases, the plugin configuration could be decided by a different
team (Operations in some cases), and the configuration is directly used by
an API owner.

decK has support for both of these use-cases.

Let's take an example configuration file:

```yaml
consumers:
- username: foo
  tags:
  - silver-tier
  plugins:
  - name: rate-limiting
    config:
      day: null
      fault_tolerant: true
      hide_client_headers: false
      hour: null
      limit_by: consumer
      minute: 10
      month: null
      policy: redis
      redis_database: 0
      redis_host: redis.common.svc
      redis_password: null
      redis_port: 6379
      redis_timeout: 2000
      second: null
      year: null
    enabled: true
    run_on: first
    protocols:
    - http
    - https
- username: bar
  tags:
  - silver-tier
  plugins:
  - name: rate-limiting
    config:
      day: null
      fault_tolerant: true
      hide_client_headers: false
      hour: null
      limit_by: consumer
      minute: 10
      month: null
      policy: redis
      redis_database: 0
      redis_host: redis.common.svc
      redis_password: null
      redis_port: 6379
      redis_timeout: 2000
      second: null
      year: null
    enabled: true
    run_on: first
    protocols:
    - http
    - https
- username: baz
  tags:
  - gold-tier
  plugins:
  - name: rate-limiting
    config:
      day: null
      fault_tolerant: true
      hide_client_headers: false
      hour: null
      limit_by: consumer
      minute: 20
      month: null
      policy: redis
      redis_database: 0
      redis_host: redis.common.svc
      redis_password: null
      redis_port: 6379
      redis_timeout: 2000
      second: null
      year: null
    enabled: true
    run_on: first
    protocols:
    - http
    - https
- username: fub
  tags:
  - gold-tier
  plugins:
  - name: rate-limiting
    config:
      day: null
      fault_tolerant: true
      hide_client_headers: false
      hour: null
      limit_by: consumer
      minute: 20
      month: null
      policy: redis
      redis_database: 0
      redis_host: redis.common.svc
      redis_password: null
      redis_port: 6379
      redis_timeout: 2000
      second: null
      year: null
    enabled: true
    run_on: first
    protocols:
    - http
    - https
```

Here, we have two groups of consumers:
- `silver-tier` consumers who can access our APIs at 10 requests per minute
- `gold-tier` consumers who can access our APIs at 20 requests per minute

Now, if we want to increase the rate-limits or change the host of the Redis
server, then we have to edit the configuration of each and every instance of
the plugin.

To reduce this repetition, you can de-duplicate plugin configuration and
reference it where we you need to use it.
Please do note that this works across multiple files as well.

The above file now becomes:

```yaml
_plugin_configs:
  silver-tier-limit:
    day: null
    fault_tolerant: true
    hide_client_headers: false
    hour: null
    limit_by: consumer
    minute: 14
    month: null
    policy: redis
    redis_database: 0
    redis_host: redis.common.svc
    redis_password: null
    redis_port: 6379
    redis_timeout: 2000
    second: null
    year: null
  gold-tier-limit:
    day: null
    fault_tolerant: true
    hide_client_headers: false
    hour: null
    limit_by: consumer
    minute: 20
    month: null
    policy: redis
    redis_database: 0
    redis_host: redis.common.svc
    redis_password: null
    redis_port: 6379
    redis_timeout: 2000
    second: null
    year: null
consumers:
- username: foo
  tags:
  - silver-tier
  plugins:
  - name: rate-limiting
    _config: silver-tier-limit
    enabled: true
    protocols:
    - http
    - https
- username: bar
  tags:
  - silver-tier
  plugins:
  - name: rate-limiting
    _config: silver-tier-limit
    enabled: true
    protocols:
    - http
    - https
- username: baz
  tags:
  - gold-tier
  plugins:
  - name: rate-limiting
    _config: gold-tier-limit
    enabled: true
    protocols:
    - http
    - https
- username: fub
  tags:
  - gold-tier
  plugins:
  - name: rate-limiting
    _config: gold-tier-limit
    enabled: true
    protocols:
    - http
    - https
```

Now, you can edit plugin configuration in a single place and you can see it's
effect across multiple entities. Under the hood, decK takes the change and
applies it to each entity which references the plugin configuration that has
been changed. As always, use `deck diff` to inspect the changes before you
apply those to your Kong clusters.
