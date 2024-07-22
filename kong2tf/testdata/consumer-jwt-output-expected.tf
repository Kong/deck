resource "konnect_gateway_consumer" "example-user" {
  username         = "example-user"
  control_plane_id = var.control_plane_id
}
resource "konnect_gateway_jwt" "example-user_my_jwt_secret" {
  key              = "my_jwt_secret"
  algorithm        = "HS256"
  secret           = "my_secret_key"
  tags             = ["internal"]
  consumer_id      = konnect_gateway_consumer.example-user.id
  control_plane_id = var.control_plane_id
}
