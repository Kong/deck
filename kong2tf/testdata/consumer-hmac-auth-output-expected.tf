variable "control_plane_id" {
  type = "string"
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_consumer" "example_user" {
  username = "example-user"

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_hmac_auth" "example_user_hmac_auth_example_user" {
  username = "example-user"
  secret = "just-a-secret"

  consumer_id = konnect_gateway_consumer.example_user.id

  control_plane_id = var.control_plane_id
}

