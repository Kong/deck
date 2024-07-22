resource "konnect_gateway_consumer_group" "example-consumer-group" {
  name             = "example-consumer-group"
  control_plane_id = var.control_plane_id
}
resource "konnect_gateway_plugin_rate_limiting_advanced" "example-consumer-group_rate-limiting-advanced" {
  config = {"hide_client_headers":false,"identifier":"consumer","limit":[5],"namespace":"example_namespace","retry_after_jitter_max":0,"strategy":"local","sync_rate":-1,"window_size":[30],"window_type":"sliding"}
  consumer_group = {
    id = konnect_gateway_consumer_group.example-consumer-group.id
  }
  control_plane_id = var.control_plane_id
}
