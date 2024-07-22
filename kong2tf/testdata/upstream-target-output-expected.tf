resource "konnect_gateway_upstream" "example-apicom" {
  name             = "example-api.com"
  slots            = 10000
  host_header      = "example.com"
  algorithm        = "round-robin"
  hash_on          = "none"
  hash_fallback    = "none"
  hash_on_cookie_path = "/"
  use_srv_name     = false
  tags             = ["user-level", "low-priority"]
  healthchecks = {
    active = {
        concurrency = 10
        http_path = "/"
        https_sni = "example.com"
        https_verify_certificate = true
        timeout = 1
        type = "http"
        headers = {
          "x-another-header" = jsonencode(["bla"])
          "x-my-header" = jsonencode(["foo","bar"])
        }
      healthy = {
        http_statuses = [200, 302]
        interval = 0
        successes = 0
      }
      unhealthy = {
        http_failures = 0
        http_statuses = [429, 404, 500, 501, 502, 503, 504, 505]
        tcp_failures = 0
        timeouts = 0
        interval = 0
      }
    }
    passive = {
        healthy = {
            http_statuses = [200, 201, 202, 203, 204, 205, 206, 207, 208, 226, 300, 301, 302, 303, 304, 305, 306, 307, 308]
            successes = 0
        }
        unhealthy = {
            http_failures = 0
            http_statuses = [429, 500, 503]
            tcp_failures = 0
            timeouts = 0
        }
        type = "http"
    }
  }
  control_plane_id = var.control_plane_id
}
resource "konnect_gateway_target" "example-apicom_101010108000" {
  target           = "10.10.10.10:8000"
  weight           = 100
  upstream_id      = konnect_gateway_upstream.example-apicom.id
  control_plane_id = var.control_plane_id
}
resource "konnect_gateway_target" "example-apicom_101010118000" {
  target           = "10.10.10.11:8000"
  weight           = 100
  upstream_id      = konnect_gateway_upstream.example-apicom.id
  control_plane_id = var.control_plane_id
}
