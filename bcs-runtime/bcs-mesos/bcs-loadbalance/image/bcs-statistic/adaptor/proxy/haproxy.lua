-- proxy implementation for haproxy

local Class = require("pl.class")
local Set = require("pl.Set")
local Proxy = require("bcs-statistic.adaptor.proxy.proxy")
local Request = require("bcs-statistic.http.request")
local Response = require("bcs-statistic.http.response")

local Haproxy = Class(Proxy)

local dataMethod = Set{"POST", "PUT", "PATCH"}

function Haproxy:_init()
    self:super("haproxy")
end

-- inject server implementation
-- @param srv server implementation @bcs-statistic.server
function Haproxy:SetServer(srv)
    self.server = srv
end

-- register call back proxy depandency
function Haproxy:Register()
    -- register http request from haproxy context
    core.register_service("bcs-statistic", "http", function(app) self:Handle(app) end)
end

-- handle haproxy http request, response statistic datas
-- @return None, http response 
function Haproxy:Handle(applet)
    local request = RequestFormat(applet)
    local response
    if not self.server then
        -- response: 503, no available backend
        response = Response(503, "No available backend server")
    else
        response = self.server:Serve(request)
    end
    ResponseFormat(applet, response)
end

-- change haproxy http request to local formation
-- @return request, see @http.request
function RequestFormat(applet)
    -- body
    local req = Request(applet.method, applet.path, applet.header, applet.qs)
    if dataMethod[applet.method] then
        req:SetData(applet:receive())
    end
    return req
end

-- change local response to haproxy http response
function ResponseFormat(applet, res)
    -- body
    applet:set_status(res.StatusCode)
    for i,v in pairs(res.Header) do
        applet:add_header(i, v)
    end
    applet:start_response()
    applet:send(res.Body)
end

return Haproxy
