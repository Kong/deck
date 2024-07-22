resource "konnect_gateway_consumer_group" "example-consumer-group" {
  name             = "example-consumer-group"
  control_plane_id = var.control_plane_id
}
resource "konnect_gateway_consumer_group_member" "example-user" {
  consumer_id       = konnect_gateway_consumer.example-user.id
  consumer_group_id = konnect_gateway_consumer_group.example-consumer-group.id
  control_plane_id  = var.control_plane_id
}
