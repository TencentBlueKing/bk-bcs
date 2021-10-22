-- collector implementation for haproxy
-- collecting haproxy status with admin stats sock


local Class = require("pl.class")
local Collector = require("bcs-statistic.adaptor.collector.collector")
local Stringx = require("pl.stringx")
local Types = require("bcs-statistic.http.types")

local HaCollector = Class(Collector)

local sockpath = "/var/run/haproxy.sock"

function HaCollector:_init()
    self:super("haproxy")
end

-- get proxy information
-- @return Info, proxy information, see @http.Info
function HaCollector:GetInfo()
    -- None implementation
    return nil
end

-- get all statistic info
-- @return data, status data, see bcs-statistic.http.Stats
function HaCollector:GetStats()
    local rawData = self:connection("show stat")
    core.Info("HaCollector show stat got " .. #rawData .. " data")
    return statsAdaptor(rawData)
end

-- get service by index, like name, id
-- @return service, service info, see bcs-statistic.data.Service
function HaCollector:GetService( index )
    -- body
    return nil
end

-- create connection to haproxy, execute administer command,
-- receive all response, response is string, detail data structure 
-- in https://cbonte.github.io/haproxy-dconv/1.8/management.html#9.3
-- and use command socat for a trial, for example:
-- echo show stats | socat /var/run/haproxy.sock stdio
function HaCollector:connection(cmd)
    local socket = core.tcp()
    -- todo(developer): handle panic when connect refuse
    socket:connect(sockpath)
    socket:settimeout(3)
    local re = socket:send(cmd .. "\n")
    local stringdata = socket:receive("*a")
    socket:close()
    return stringdata
end

-- stats info adaptor
-- @param data, string, stat response from haproxy
-- @return Stats, see @bcs-statistic.http.types
-- all status structure please see @https://cbonte.github.io/haproxy-dconv/1.8/management.html#9
function statsAdaptor(data)
    if not data then
        core.Info("Get empty data from HaCollector")
        return nil
    end
    if #data < 666 then
        -- title length is 666
        core.Info("Get less data than expected")
        return nil
    end
    local lines = Stringx.splitlines(data)
    if #lines < 2 then
        -- no data
        core.Info("data line is less than expetced")
        return nil
    end
    local stats = Types.Stats()
    for index, line in pairs(lines) do
        -- first line is title, skip
        core.Info("index: " .. index .. " data: " .. line)
        if index > 1 and #line > 50 then
            lineData = Stringx.split(line, ",")
            if lineData[2] == "FRONTEND" then
                front = frontendAdaptor(lineData)
                stats:AddFrontend(front)
            elseif lineData[2] == "BACKEND" then
                back = backendAdaptor(lineData)
                stats:AddBackend(back)
            else
                -- server data
                serv = serverAdaptor(lineData)
                stats:AddServer(serv)
            end
        end
    end
    return stats
end

-- create frontend data
function frontendAdaptor(data)
    local frontend = Types.Frontend()
    frontend.Name = data[1]
    frontend.CurSessRate  = data[5]   -- current sessions
    frontend.MaxSessRate  = data[6]   -- max sessions
    frontend.SessionLimit = data[7]   -- configured session limit
    frontend.CumSession   = data[8]   -- cumulative session
    frontend.BytesIn      = data[9]      
    frontend.BytesOut     = data[10]     
    frontend.RequestDeny  = data[11]  
    frontend.ResponseDeny = data[12] 
    frontend.RequestError = data[13] 
    frontend.Status       = data[18]  -- status (OPEN/UP/DOWN/NOLB/MAINT/MAINT)
    frontend.Rate         = data[34]  -- number of sessions per second over last elapsed second
    frontend.RateMax      = data[36]  -- max number of new sessions per second
    frontend.ConnRate     = data[78]  -- number of connections over the last elapsed second
    frontend.ConnRateMax  = data[79]  -- highest known conn_rate
    frontend.ConnTot      = data[80]  -- cumulative number of connections
    return frontend
end

-- create backend data
function backendAdaptor(data)
    local backend = Types.Backend()
    backend.Name = data[1]
    backend.CurQueue      = data[3] -- current queued requests. For the backend this reports the number queued without a server assigned.
    backend.CurQMax       = data[4] -- max value of qcur
    backend.CurSessRate   = data[5] -- current sessions
    backend.MaxSessRate   = data[6] -- max sessions
    backend.SessionLimit  = data[7] -- configured session limit
    backend.CumSession    = data[8] -- cumulative session
    backend.BytesIn       = data[9] -- bytes in
    backend.BytesOut      = data[10] -- bytes out
    backend.RequestDeny   = data[11] -- requests denied because of security concerns.
    backend.ResponseDeny  = data[12] -- responses denied because of security concerns.
    backend.RequestError  = data[13] -- request errors
    backend.ErrorCnt      = data[14] -- [..BS]: number of requests that encountered an error trying to connect to a backend server
    backend.ErrResp       = data[15] -- [..BS]: response errors. srv_abrt will be counted here also.
    backend.WRetry        = data[16] -- [..BS]: number of times a connection to a server was retried.
    backend.WRedispath    = data[17] -- [..BS]: number of times a request was redispatched to another
    backend.Status        = data[18] -- status
    backend.Active        = data[20] -- [..BS]: number of active servers (backend), server is active (server)
    backend.ChkDown       = data[23] -- [..BS]: number of UP->DOWN transitions. The backend counter counts transitions to the whole backend being down, rather than the sum of the counters for each server.
    backend.DownTime      = data[25] -- total downtime (in seconds). The value for the backend is the downtime for the whole backend, not the sum of the server downtime.
    backend.Rate          = data[34] -- number of sessions per second over last elapsed second
    backend.RateMax       = data[36] -- max number of new sessions per second
    backend.Hrsp_1xx      = data[40] -- http responses with 1xx code
    backend.Hrsp_2xx      = data[41] -- http responses with 2xx code
    backend.Hrsp_3xx      = data[42] -- http responses with 3xx code
    backend.Hrsp_4xx      = data[43] -- http responses with 4xx code
    backend.Hrsp_5xx      = data[44] -- http responses with 5xx code
    backend.Mode          = data[76] -- proxy mode (tcp, http, health, unknown)
    backend.Algo          = data[77] -- load balancing algorithm
    return backend
end


-- create detail server
function serverAdaptor( data )
    local serv = Types.Server()
    serv.Name           = data[2]
    serv.Pxname         = data[1]
    serv.CurQueue       = data[3] -- current queued requests. For the backend this reports the number queued without a server assigned.
    serv.CurQMax        = data[4] -- max value of qcur
    serv.CurSessRate    = data[5] -- current sessions
    serv.MaxSessRate    = data[6] -- max sessions
    serv.SessionLimit   = data[7] -- configured session limit
    serv.CumSession     = data[8] -- cumulative session
    serv.BytesIn        = data[9] -- bytes in 
    serv.BytesOut       = data[10] -- bytes out
    serv.RequestDeny    = data[11] -- requests denied because of security concerns.
    serv.ResponseDeny   = data[12] -- responses denied because of security concerns.
    serv.RequestError   = data[13] -- request errors 
    serv.ErrorCnt       = data[14] -- [..BS]: number of requests that encountered an error trying to connect to a backend server
    serv.ErrResp        = data[15] -- [..BS]: response errors. srv_abrt will be counted here also.
    serv.WRetry         = data[16] -- [..BS]: number of times a connection to a server was retried.
    serv.WRedispath     = data[17] -- [..BS]: number of times a request was redispatched to another
    serv.Status         = data[18] -- status
    serv.Weight         = data[19] -- total weight (backend), server weight (server)
    serv.Active         = data[20] -- [..BS]: number of active servers (backend), server is active (server)
    serv.ChkDown        = data[23] -- [..BS]: number of UP->DOWN transitions. The backend counter counts transitions to the whole backend being down, rather than the sum of the counters for each server.
    serv.DownTime       = data[25] -- total downtime (in seconds). The value for the backend is the downtime for the whole backend, not the sum of the server downtime.
    serv.LBTot          = data[31] -- [..BS]: total number of times a server was selected, either for new sessions, or when re-dispatching. The server counter is the number of times that server was selected.
    serv.Rate           = data[34] -- number of sessions per second over last elapsed second
    serv.RateMax        = data[36] -- max number of new sessions per second
    serv.Hrsp_1xx       = data[40] -- http responses with 1xx code
    serv.Hrsp_2xx       = data[41] -- http responses with 2xx code
    serv.Hrsp_3xx       = data[42] -- http responses with 3xx code
    serv.Hrsp_4xx       = data[43] -- http responses with 4xx code
    serv.Hrsp_5xx       = data[44] -- http responses with 5xx code
    serv.LastSess       = data[56] -- [..BS]: number of seconds since last session assigned to server/backend
    serv.QTime          = data[59] -- [..BS]: the average queue time in ms over the 1024 last requests
    serv.CTime          = data[60] -- [..BS]: the average connect time in ms over the 1024 last requests
    serv.RTime          = data[61] -- [..BS]: the average response time in ms over the 1024 last requests (0 for TCP)
    serv.TTime          = data[62] -- [..BS]: the average total session time in ms over the 1024 last requests
    serv.AgentStatus    = data[63] -- status of last agent check
    serv.AgentDur       = data[65] -- [...S]: time in ms taken to finish last check
    serv.Addr           = data[74] -- ip address & port
    return serv
end

return HaCollector
