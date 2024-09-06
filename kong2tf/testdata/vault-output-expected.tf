variable "control_plane_id" {
  type = "string"
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_vault" "env" {
  name = "env"
  config = jsonencode({
    prefix = "MY_SECRET_"
  })
  description = "ENV vault for secrets"
  prefix = "my-env-vault"
  tags = ["env-vault"]

  control_plane_id = var.control_plane_id
}

