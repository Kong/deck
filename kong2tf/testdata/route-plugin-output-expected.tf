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
  paths = ["~/v1/example/?$"]

  service = {
    id = konnect_gateway_service.example_service.id
  }

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_plugin_cors" "example_route_cors" {
  config = {
    credentials = true
    exposed_headers = ["X-My-Header"]
    headers = ["Authorization"]
    max_age = 3600
    methods = ["GET", "POST"]
    origins = ["example.com"]
  }

  route = {
    id = konnect_gateway_route.example_route.id
  }

  control_plane_id = var.control_plane_id
}

