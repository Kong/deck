variable "control_plane_id" {
  type = string
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_service" "example_service" {
  name = "example-service"
  host = "example-api.com"
  port = 80
  protocol = "http"

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_route" "example_route" {
  name = "example-route"
  destinations = [
    {
      ip = "10.10.10.10"
      port = 8080
    },
  ]
  headers = {
    x-another-header = ["first-header-value", "second-header-value"]
    x-my-header = ["~*foos?bar$"]
  }
  hosts = ["example.com", "another-example.com", "yet-another-example.com"]
  https_redirect_status_code = 302
  methods = ["GET", "POST"]
  paths = ["~/v1/example/?$", "/v1/another-example", "/v1/yet-another-example"]
  preserve_host = true
  protocols = ["http", "https"]
  regex_priority = 1
  snis = ["example.com"]
  sources = [
    {
      ip = "192.168.0.1"
    },
  ]
  strip_path = false
  tags = ["version:v1"]

  service = {
    id = konnect_gateway_service.example_service.id
  }

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_route" "top_level_route" {
  name = "top-level-route"
  hosts = ["top-level.example.com"]

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_route" "top_level_with_service_route" {
  name = "top-level-with-service-route"
  hosts = ["top-level-with-service.example.com"]

  service = {
    id = konnect_gateway_service.example_service.id
  }

  control_plane_id = var.control_plane_id
}

