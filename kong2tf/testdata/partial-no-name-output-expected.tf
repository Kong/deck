variable "control_plane_id" {
  type = string
  default = "YOUR_CONTROL_PLANE_ID"
}

resource "konnect_gateway_partial" "partial_1f8f3b1e_7a8b_4c2d_9e6f_0a1b2c3d4e5f" {
  redis_ee = {
    config = {
      host = "127.0.0.1"
      port = 6379
    }
  }

  control_plane_id = var.control_plane_id
}

resource "konnect_gateway_partial" "partial_redis_ce_1" {
  redis_ce = {
    config = {
      host = "redis.example.com"
      port = 6380
    }
  }

  control_plane_id = var.control_plane_id
}

