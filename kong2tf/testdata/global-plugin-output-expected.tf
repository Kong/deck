resource "konnect_gateway_plugin_openid_connect" "openid-connect" {
  config = {"auth_methods":["authorization_code","session"],"client_id":["\u003cclient-id\u003e"],"client_secret":["\u003cclient-secret\u003e"],"issuer":"http://example.org","response_mode":"form_post","session_secret":"\u003csession-secret\u003e"}
  control_plane_id = var.control_plane_id
}
