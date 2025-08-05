variable "control_plane_id" {
  type = string
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_partial" "Demo" {
  redis_ee = {
    name = "Demo"
    config = {
      username = "username_here"
      connect_timeout = 2000
      connection_is_proxied = false
      database = 0
      host = "127.0.0.1"
      keepalive_backlog = 0
      keepalive_pool_size = 256
      password = "password_here"
      port = 6379
      read_timeout = 2000
      send_timeout = 2000
      server_name = "redis.example.com"
      ssl = true
      ssl_verify = false
    }
  }

  control_plane_id = var.control_plane_id
}

