variable "control_plane_id" {
  type = "string"
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_consumer" "example_user" {
  username = "example-user"

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_jwt" "example_user_jwt_my_jwt_secret" {
  algorithm = "HS256"
  key = "my_jwt_secret"
  secret = "my_secret_key"
  tags = ["internal"]

  consumer_id = konnect_gateway_consumer.example_user.id

  control_plane_id = var.control_plane_id
}

