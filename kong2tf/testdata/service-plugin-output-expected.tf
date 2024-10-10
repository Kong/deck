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

resource "konnect_gateway_plugin_rate_limiting_advanced" "example_service_rate_limiting_advanced" {
  config = {
    hide_client_headers = false
    identifier = "consumer"
    limit = [5]
    namespace = "example_namespace"
    strategy = "local"
    sync_rate = -1
    window_size = [30]
  }
  ordering = {
    after = {
      access = ["yet-another-plugin"]
    }
    before = {
      access = ["another-plugin"]
    }
  }

  service = {
    id = konnect_gateway_service.example_service.id
  }

  control_plane_id = var.control_plane_id
}

