resource "konnect_gateway_service" "example-service" {
  name             = "example-service"
  protocol         = "http"
  host             = "example-api.com"
  port             = 80
  control_plane_id = var.control_plane_id
}

  
resource "konnect_gateway_route" "example-route" {
  name = "example-route"
  hosts = ["example.com", "another-example.com", "yet-another-example.com"]
  headers = {
    "x-another-header" = jsonencode(["first-header-value","second-header-value"])
    "x-my-header" = jsonencode(["~*foos?bar$"])
  }
  methods = ["GET", "POST"]
  paths = ["~/v1/example/?$", "/v1/another-example", "/v1/yet-another-example"]
  preserve_host = true
  protocols = ["http", "https"]
  regex_priority = 1
  strip_path = false
  snis = ["example.com"]
  sources = [
    {
      ip   = "192.168.0.1"
    }
  ]
  destinations = [
    {
      ip   = "10.10.10.10"
      port = 8080
    }
  ]
  tags = ["version:v1"]
  https_redirect_status_code = 302
  service = {
    id = konnect_gateway_service.example-service.id
  }  
  control_plane_id = var.control_plane_id
}
