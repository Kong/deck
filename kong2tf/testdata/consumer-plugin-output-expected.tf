resource "konnect_gateway_consumer" "example-user" {
  username         = "example-user"
  control_plane_id = var.control_plane_id
}
resource "konnect_gateway_plugin_rate_limiting_advanced" "example-user_rate-limiting-advanced" {
  config = {"hide_client_headers":false,"identifier":"consumer","limit":[5],"namespace":"example_namespace","strategy":"local","sync_rate":-1,"window_size":[30]}
  consumer = {
    id = konnect_gateway_consumer.example-user.id
  }
  control_plane_id = var.control_plane_id
}
