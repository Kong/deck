variable "control_plane_id" {
  type    = string
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_consumer" "my_consumer_prod_example" {
  username  = "my-consumer.prod.example"
  custom_id = "1234567890"
  tags      = ["internal"]

  control_plane_id = var.control_plane_id
}

