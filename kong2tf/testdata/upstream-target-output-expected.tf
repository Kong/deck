variable "control_plane_id" {
  type = "string"
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_upstream" "upstream_example_api_com" {
  name = "example-api.com"
  algorithm = "round-robin"
  hash_fallback = "none"
  hash_on = "none"
  hash_on_cookie_path = "/"
  healthchecks = {
    active = {
      concurrency = 10
      headers = {
        x-another-header = ["bla"]
        x-my-header = ["foo", "bar"]
      }
      healthy = {
        http_statuses = [200, 302]
        interval = 0
        successes = 0
      }
      http_path = "/"
      https_sni = "example.com"
      https_verify_certificate = true
      timeout = 1
      type = "http"
      unhealthy = {
        http_failures = 0
        http_statuses = [429, 404, 500, 501, 502, 503, 504, 505]
        interval = 0
        tcp_failures = 0
        timeouts = 0
      }
    }
    passive = {
      healthy = {
        http_statuses = [200, 201, 202, 203, 204, 205, 206, 207, 208, 226, 300, 301, 302, 303, 304, 305, 306, 307, 308]
        successes = 0
      }
      type = "http"
      unhealthy = {
        http_failures = 0
        http_statuses = [429, 500, 503]
        tcp_failures = 0
        timeouts = 0
      }
    }
    threshold = 0
  }
  host_header = "example.com"
  slots = 10000
  tags = ["user-level", "low-priority"]
  use_srv_name = false

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_target" "upstream_example_api_com_target_10_10_10_10_8000" {
  target = "10.10.10.10:8000"
  weight = 100

  upstream_id = konnect_gateway_upstream.upstream_example_api_com.id

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_target" "upstream_example_api_com_target_10_10_10_11_8000" {
  target = "10.10.10.11:8000"
  weight = 200

  upstream_id = konnect_gateway_upstream.upstream_example_api_com.id

  control_plane_id = var.control_plane_id
}

