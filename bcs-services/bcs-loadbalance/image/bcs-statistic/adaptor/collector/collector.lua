-- Collector interface definition

local Class = require("pl.class")

local Collector = Class()

function Collector:_init(name)
    self.name = name
end

-- get proxy configuration
-- @return config, plan text of configuration
function Collector:GetInfo()
    -- None implementation
    return nil
end

-- get all statistic info
-- @return data, status data, see bcs-statistic.data.StatusData
function Collector:GetStats()
    -- body
    return nil
end

-- get service by index, like name, id
-- @return service, service info, see bcs-statistic.data.Service
function Collector:GetService( index )
    -- body
    return nil
end


return Collector
