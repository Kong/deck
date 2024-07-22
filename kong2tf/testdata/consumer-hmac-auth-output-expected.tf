resource "konnect_gateway_consumer" "example-user" {
  username         = "example-user"
  control_plane_id = var.control_plane_id
}
resource "konnect_gateway_hmac_auth" "example-user_example-user" {
  username         = "example-user"
  secret           = "just-a-secret"
  consumer_id      = konnect_gateway_consumer.example-user.id
  control_plane_id = var.control_plane_id
}
