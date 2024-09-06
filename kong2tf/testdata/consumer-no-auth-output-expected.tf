variable "control_plane_id" {
  type = "string"
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_consumer" "example_user" {
  username = "example-user"
  custom_id = "1234567890"
  tags = ["internal"]

  control_plane_id = var.control_plane_id
}

