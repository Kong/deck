variable "control_plane_id" {
  type = "string"
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_consumer" "example_user" {
  username = "example-user"

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_basic_auth" "example_user_basic_auth_my_basic_user" {
  username = "my_basic_user"
  password = "my_basic_password"
  tags = ["internal"]

  consumer_id = konnect_gateway_consumer.example_user.id

  control_plane_id = var.control_plane_id
}

