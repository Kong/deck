variable "control_plane_id" {
  type = string
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_service" "example_service" {
  enabled = true
  name = "example-service"
  client_certificate = "4e3ad2e4-0bc4-4638-8e34-c84a417ba39b"
  connect_timeout = 5000
  host = "example-api.com"
  path = "/v1"
  port = 80
  protocol = "http"
  read_timeout = 60000
  retries = 5
  tags = ["example", "api"]
  tls_verify = true
  tls_verify_depth = 1
  write_timeout = 60000

  control_plane_id = var.control_plane_id
}

