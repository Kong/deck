return {
  name = "set-header",
  fields = {
    { protocols = require("kong.db.schema.typedefs").protocols_http },
    {
      config = {
        type = "record",
        fields = {
          { name = { description = "The name of the header to set.", type = "string", required = true, }, },
          { value = { description = "The value for the header.", type = "string", required = true, }, },
        },
      },
    },
  },
}
