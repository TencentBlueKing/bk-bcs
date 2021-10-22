-- init for proxy module

local Haproxy = require("bcs-statistic.adaptor.proxy.haproxy")
local HaColloctor = require("bcs-statistic.adaptor.collector.haproxy")
-- local nginx = require("bcs-plugins.adaptor.nginx")

-- get proxy implementation according system env
-- @return detailed proxy implementation
function GetProxy()
    local module = os.getenv("BCS_PROXY_MODULE")
    if module == "haproxy" then
        local ha = Haproxy()
        return ha
    end
end

-- get collector implementation according system env
-- @return detailed collector implementation
function GetCollector()
    local module = os.getenv("BCS_PROXY_MODULE")
    if module == "haproxy" then
        core.Info("Get haproxy collector here")
        local ha = HaColloctor()
        return ha
    end
end

return {
    GetProxy = GetProxy,
    GetCollector = GetCollector,
}