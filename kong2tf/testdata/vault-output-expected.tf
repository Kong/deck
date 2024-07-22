resource "konnect_gateway_vault" "env" {
  name   = "env"
  prefix = "my-env-vault"
  config = jsonencode({"prefix":"MY_SECRET_"})
  control_plane_id = var.control_plane_id
}