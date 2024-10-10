variable "control_plane_id" {
  type = string
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_plugin_rate_limiting" "rate_limiting" {
  enabled = false
  config = {
    hour = 10000
    policy = "local"
    second = 5
  }

  control_plane_id = var.control_plane_id
}

