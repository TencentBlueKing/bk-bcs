-- proxy interface definition

local Class = require("pl.class")

local Proxy = Class()

function Proxy:_init( name )
    self.name = name or "unkown proxy"
end

-- inject server implementation
-- @param srv server implementation @bcs-statistic.server
function Proxy:SetServer(srv)
    self.server = srv
end

-- register call back proxy depandency
function Proxy:Register()
    -- body, virtual
end

return Proxy