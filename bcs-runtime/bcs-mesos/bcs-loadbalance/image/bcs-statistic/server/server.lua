-- BCS loadbalance server layer for statistics

local Class = require("pl.class")
local Router = require("router")
-- local ConfigView = require("bcs-statistic.server.config")
local StatsView  = require("bcs-statistic.server.stats")

local Server = Class()

function Server:_init(name, co)
    self.name = name or "bcs-loadbalance"
    self.collector = co
    self.router = Router.new()
    -- init all views
    -- self.config = ConfigView(self.collector)
    self.stats = StatsView(self.collector)
end

-- Server handle statistic http request, dispatch to View by router
-- @param request simple http request from http.Request
-- @return response simple http response from http.Response
function Server:Serve(request)
    -- find function
    local fnx, args = self.router:resolve(request.Method, request.Path)
    request.PathArgs = args
    -- execute function
    return fnx(request)
end

-- setting all views for http request
function Server:SetupViews()
    -- setting config view
    -- self.config:Register(self.router)
    -- setting stats view 
    self.stats:Register(self.router)
end

return Server
