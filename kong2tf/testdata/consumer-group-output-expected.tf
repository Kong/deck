variable "control_plane_id" {
  type = string
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_consumer_group" "example_consumer_group" {
  name = "example-consumer-group"

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_consumer_group_member" "example_consumer_group_example_user" {
  consumer_id = konnect_gateway_consumer.example_user.id
  consumer_group_id = konnect_gateway_consumer_group.example_consumer_group.id
  control_plane_id = var.control_plane_id
}

