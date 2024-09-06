variable "control_plane_id" {
  type = "string"
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_consumer_group" "example_consumer_group" {
  name = "example-consumer-group"

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_plugin_rate_limiting_advanced" "example_consumer_group_rate_limiting_advanced" {
  config = {
    hide_client_headers = false
    identifier = "consumer"
    limit = [5]
    namespace = "example_namespace"
    retry_after_jitter_max = 0
    strategy = "local"
    sync_rate = -1
    window_size = [30]
    window_type = "sliding"
  }

  consumer_group = {
    id = konnect_gateway_consumer_group.example_consumer_group.id
  }

  control_plane_id = var.control_plane_id
}

