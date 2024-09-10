variable "control_plane_id" {
  type = "string"
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_consumer" "example_user" {
  username = "example-user"

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_acl" "example_user_acl_demo_group" {
  group = "demo_group"
  tags = ["internal"]

  consumer_id = konnect_gateway_consumer.example_user.id

  control_plane_id = var.control_plane_id
}

