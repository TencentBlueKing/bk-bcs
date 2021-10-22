-- configuration url view

local Class = require("pl.class")
local Json = require("cjson")
local View = require("bcs-statistic.http.view")
local Response = require("bcs-statistic.http.response")
local StatsView = Class(View)

function StatsView:_init(collector)
    self:super("ConfigView")
    self.collector = collector
end

-- register all method callback with router
function StatsView:Register(router)
    router:get("/stats", function(req) return self:Get(req) end)
end

-- interface for handling get
-- @return response, see @http.response
function StatsView:Get(request)
    --get proxy configuration from collector
    local stats = self.collector:GetStats()
    local res
    if stats == nil then
        res = Response(204, "No Content")
    else
        local jsonStr = Json.encode(stats)
        res = Response(200, jsonStr, {["Content-Type"] = 'application/json'})
    end
    return res
end

-- interface for handling POST
-- @return response, see @http.response
function StatsView:Post(request)
    -- body
    return Response(501, "Not Implemented")
end

-- interface for handling Delete
-- @return response, see @http.response
function StatsView:Delete(request)
    -- body
    return Response(501, "Not Implemented")
end

-- interface for handling put
-- @return response, see @http.response
function StatsView:Put(request)
    -- body
    return Response(501, "Not Implemented")
end

return StatsView
