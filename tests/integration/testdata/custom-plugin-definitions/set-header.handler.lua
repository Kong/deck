return {
  VERSION = "1.0,0",
  PRIORITY = 500,
  access = function(self, config)
    kong.service.request.set_header(config.name, config.value)
  end
}
