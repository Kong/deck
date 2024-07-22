resource "konnect_gateway_consumer" "example-user" {
  username         = "example-user"
  custom_id        = "1234567890"
  tags             = ["internal"]
  control_plane_id = var.control_plane_id
}
