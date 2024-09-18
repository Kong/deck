variable "control_plane_id" {
  type = "string"
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_consumer" "example_user" {
  username = "example-user"

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_key_auth" "example_user_key_auth_my_api_key" {
  key = "my_api_key"
  tags = ["internal"]

  consumer_id = konnect_gateway_consumer.example_user.id

  control_plane_id = var.control_plane_id
}

