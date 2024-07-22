resource "konnect_gateway_service" "example-service" {
  name             = "example-service"
  protocol         = "http"
  host             = "example-api.com"
  port             = 80
  control_plane_id = var.control_plane_id
}
resource "konnect_gateway_plugin_rate_limiting_advanced" "example-service_rate-limiting-advanced" {
  config = {"hide_client_headers":false,"identifier":"consumer","limit":[5],"namespace":"example_namespace","strategy":"local","sync_rate":-1,"window_size":[30]}
  service = {
    id = konnect_gateway_service.example-service.id
  }
  control_plane_id = var.control_plane_id
}

