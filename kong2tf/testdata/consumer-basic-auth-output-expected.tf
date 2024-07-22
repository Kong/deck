resource "konnect_gateway_consumer" "example-user" {
  username         = "example-user"
  control_plane_id = var.control_plane_id
}
resource "konnect_gateway_basic_auth" "example-user_my_basic_user" {
  username         = "my_basic_user"
  password         = "my_basic_password"
  tags             = ["internal"]
  consumer_id      = konnect_gateway_consumer.example-user.id
  control_plane_id = var.control_plane_id
}
