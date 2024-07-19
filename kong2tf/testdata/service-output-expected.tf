resource "konnect_gateway_service" "example-service" {
  name             = "example-service"
  protocol         = "http"
  host             = "example-api.com"
  port             = 80
  path             = "/v1"
  connect_timeout  = 5000
  read_timeout     = 60000
  write_timeout    = 60000
  retries          = 5
  tls_verify       = true
  tls_verify_depth = 1
  tags             = ["example", "api"]
  control_plane_id = var.control_plane_id
}

