resource "konnect_gateway_service" "example-service" {
  name             = "example-service"
  protocol         = "http"
  host             = "example-api.com"
  port             = 80
  control_plane_id = var.control_plane_id
}

  
resource "konnect_gateway_route" "example-route" {
  name = "example-route"
  paths = ["~/v1/example/?$"]
  service = {
    id = konnect_gateway_service.example-service.id
  }  
  control_plane_id = var.control_plane_id
}
resource "konnect_gateway_plugin_cors" "example-service_example-route_cors" {
  config = {"credentials":true,"exposed_headers":["X-My-Header"],"headers":["Authorization"],"max_age":3600,"methods":["GET","POST"],"origins":["example.com"]}
  route = {
    id = konnect_gateway_route.example-route.id
  }
  control_plane_id = var.control_plane_id
}
