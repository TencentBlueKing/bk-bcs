
-- local proxyModule = os.getenv("BCS_PROXY_MODULE")

local Adaptor = require("bcs-statistic.adaptor")
local Server = require("bcs-statistic.server.server")

-- initial
local proxy = Adaptor.GetProxy()
local collector = Adaptor.GetCollector()
local srv = Server("bcs-statistic", collector)

proxy:SetServer(srv)
srv:SetupViews()
-- register proxy callback
proxy:Register()

