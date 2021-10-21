-- configuration url view

local Class = require("pl.class")
local View = require("bcs-statistic.http.view")
local Response = require("bcs-statistic.http.response")
local ConfigView = Class(View)

function ConfigView:_init(collector)
    self:super("ConfigView")
    self.collector = collecotr
end

-- register all method callback with router
function ConfigView:Register(router)
    -- body
end

-- interface for handling get
-- @return response, see @http.response
function ConfigView:Get(request)
    --get proxy configuration from collector
    return nil
end

-- interface for handling POST
-- @return response, see @http.response
function ConfigView:Post(request)
    -- body
    return Response(501, "Not Implemented")
end

-- interface for handling Delete
-- @return response, see @http.response
function ConfigView:Delete(request)
    -- body
    return Response(501, "Not Implemented")
end

-- interface for handling put
-- @return response, see @http.response
function ConfigView:Put(request)
    -- body
    return Response(501, "Not Implemented")
end

return ConfigView