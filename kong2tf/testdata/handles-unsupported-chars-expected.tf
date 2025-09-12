variable "control_plane_id" {
  type    = string
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_consumer" "my_consumer123_my_org_prod_2_example" {
  username = "my-consumer123@my-org.prod_2.example"
  custom_id = "1234567890"
  tags = ["internal"]

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_consumer" "宮本武蔵_my_org_prod_example" {
  username = "宮本武蔵@my-org.prod.example"
  custom_id = "1234567891"
  tags = ["internal"]

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_consumer" "my_consumer_124_my_org_prod_example" {
  username = "my-consumer©124@my-org.prod.example"
  custom_id = "1234567892"
  tags = ["internal"]

  control_plane_id = var.control_plane_id
}