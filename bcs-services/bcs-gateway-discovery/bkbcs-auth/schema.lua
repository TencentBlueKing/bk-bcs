local typedefs = require "kong.db.schema.typedefs"

return {
  name = "bkbcs-auth",
  fields = {
    { protocols = typedefs.protocols_http },
    { config = {
        type = "record",
        fields = {
          -- NOTE: any field added here must be also included in the handler's get_queue_id method
          { bkbcs_auth_endpoints = typedefs.url({ required = true }) },
          { timeout = { type = "number", default = 3 }, }, 
          { keepalive = { type = "number", default = 60000 }, },
          { retry_count = { type = "integer", default = 1 }, },
    }, }, },
  },
}