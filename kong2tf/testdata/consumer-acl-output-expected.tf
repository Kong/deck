resource "konnect_gateway_consumer" "example-user" {
  username         = "example-user"
  control_plane_id = var.control_plane_id
}
resource "konnect_gateway_acl" "example-user_acl_group" {
  group            = "acl_group"
  tags             = ["internal"]
  consumer_id      = konnect_gateway_consumer.example-user.id
  control_plane_id = var.control_plane_id
}
