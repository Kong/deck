variable "control_plane_id" {
  type = "string"
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_consumer" "example_user" {
  username = "example-user"

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_plugin_rate_limiting_advanced" "example_user_rate_limiting_advanced" {
  config = {
    hide_client_headers = false
    identifier = "consumer"
    limit = [5]
    namespace = "example_namespace"
    strategy = "local"
    sync_rate = -1
    window_size = [30]
  }

  consumer = {
    id = konnect_gateway_consumer.example_user.id
  }

  control_plane_id = var.control_plane_id
}

