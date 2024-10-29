variable "control_plane_id" {
  type = string
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_plugin_openid_connect" "openid_connect" {
  enabled = true
  config = {
    auth_methods = ["authorization_code", "session"]
    client_id = ["<client-id>"]
    client_secret = ["<client-secret>"]
    issuer = "http://example.org"
    response_mode = "form_post"
    session_secret = "<session-secret>"
  }

  control_plane_id = var.control_plane_id
}

