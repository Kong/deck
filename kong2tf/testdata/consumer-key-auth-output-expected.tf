resource "konnect_gateway_consumer" "example-user" {
  username         = "example-user"
  control_plane_id = var.control_plane_id
}
resource "konnect_gateway_key_auth" "example-user_my_api_key" {
  key              = "my_api_key"
  tags             = ["internal"]
  consumer_id      = konnect_gateway_consumer.example-user.id
  control_plane_id = var.control_plane_id
}
