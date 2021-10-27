-- inner statistic data structure

local Class = require("pl.class")

local Info = Class()

function Info:_init()
    self.Name = "Unknown"
    self.Node = "Unknown"
    self.Version = ""
    self.ReleaseData = ""
    self.NbProc = 0
    self.ProccessNum = 0
    self.Pid = 1
    self.UptimeSec = 0
    -- some other attributes
    self.IdlePct = 100
end

-- frontend, listen ports, algorithm
local Frontend  = Class()

function Frontend:_init()
    self.Name = ""
    self.CurSessRate  = ""  -- current sessions
    self.MaxSessRate  = ""  -- max sessions
    self.SessionLimit = ""  -- configured session limit
    self.CumSession   = ""  -- cumulative session
    self.BytesIn      = ""  -- bytes in
    self.BytesOut     = ""  -- bytes out   
    self.RequestDeny  = ""  -- requests denied because of security concerns.
    self.ResponseDeny = ""  -- responses denied because of security concerns.
    self.RequestError = ""  -- request errors
    self.Status       = ""  -- status (OPEN/UP/DOWN/NOLB/MAINT/MAINT)
    self.Rate         = ""  -- number of sessions per second over last elapsed second
    self.RateMax      = ""  -- max number of new sessions per second
    self.ConnRate     = ""  -- number of connections over the last elapsed second
    self.ConnRateMax  = ""  -- highest known conn_rate
    self.ConnTot      = ""  -- cumulative number of connections
end

-- backend statistic, container several Server instances
local Backend = Class()

function Backend:_init()
    self.Name = ""
    self.CurQueue      = "" -- current queued requests. For the backend this reports the number queued without a server assigned.
    self.CurQMax       = "" -- max value of qcur
    self.CurSessRate   = "" -- current sessions
    self.MaxSessRate   = "" -- max sessions
    self.SessionLimit  = "" -- configured session limit
    self.CumSession    = "" -- cumulative session
    self.BytesIn       = "" -- bytes in
    self.BytesOut      = "" -- bytes out
    self.RequestDeny   = "" -- requests denied because of security concerns.
    self.ResponseDeny  = "" -- responses denied because of security concerns.
    self.RequestError  = "" -- request errors
    self.Status        = "" -- status
    self.Active        = "" -- [..BS]: number of active servers (backend), server is active (server)
    self.ErrorCnt      = "" -- [..BS]: number of requests that encountered an error trying to connect to a backend server
    self.ErrResp       = "" -- [..BS]: response errors. srv_abrt will be counted here also.
    self.WRetry        = "" -- [..BS]: number of times a connection to a server was retried.
    self.WRedispath    = "" -- [..BS]: number of times a request was redispatched to another
    self.ChkDown       = "" -- [..BS]: number of UP->DOWN transitions. The backend counter counts transitions to the whole backend being down, rather than the sum of the counters for each server.
    self.DownTime      = "" -- total downtime (in seconds). The value for the backend is the downtime for the whole backend, not the sum of the server downtime.
    self.Rate          = "" -- number of sessions per second over last elapsed second
    self.RateMax       = "" -- max number of new sessions per second
    self.Hrsp_1xx      = "" -- http responses with 1xx code
    self.Hrsp_2xx      = "" -- http responses with 2xx code
    self.Hrsp_3xx      = "" -- http responses with 3xx code
    self.Hrsp_4xx      = "" -- http responses with 4xx code
    self.Hrsp_5xx      = "" -- http responses with 5xx code
    self.Mode          = "" -- proxy mode (tcp, http, health, unknown)
    self.Algo          = "" -- load balancing algorithm
    -- server list
    self.Servers = {}
end

-- detail server info under backend
local Server = Class()

function Server:_init(  )
    self.Name           = ""
    self.Pxname         = ""
    self.CurQueue       = ""    -- current queued requests. For the backend this reports the number queued without a server assigned.
    self.CurQMax        = ""    -- max value of qcur
    self.CurSessRate    = ""    -- current sessions
    self.MaxSessRate    = ""    -- max sessions
    self.SessionLimit   = ""    -- configured session limit
    self.CumSession     = ""    -- cumulative session
    self.BytesIn        = ""    -- bytes in 
    self.BytesOut       = ""    -- bytes out
    self.RequestDeny    = ""    -- requests denied because of security concerns.
    self.ResponseDeny   = ""    -- responses denied because of security concerns.
    self.RequestError   = ""    -- request errors 
    self.Status         = "UP" -- status
    self.Weight         = ""    -- total weight (backend), server weight (server)
    self.Active         = "Y"  -- [..BS]: number of active servers (backend), server is active (server)
    self.ErrorCnt       = ""    -- [..BS]: number of requests that encountered an error trying to connect to a backend server
    self.ErrResp        = ""    -- [..BS]: response errors. srv_abrt will be counted here also.
    self.WRetry         = ""    -- [..BS]: number of times a connection to a server was retried.
    self.WRedispath     = ""    -- [..BS]: number of times a request was redispatched to another
    self.ChkDown        = ""    -- [..BS]: number of UP->DOWN transitions. The backend counter counts transitions to the whole backend being down, rather than the sum of the counters for each server.
    self.DownTime       = ""    -- total downtime (in seconds). The value for the backend is the downtime for the whole backend, not the sum of the server downtime.
    self.LBTot          = ""    -- [..BS]: total number of times a server was selected, either for new sessions, or when re-dispatching. The server counter is the number of times that server was selected.
    self.Rate           = ""    -- number of sessions per second over last elapsed second
    self.RateMax        = ""    -- max number of new sessions per second
    self.Hrsp_1xx       = "" -- http responses with 1xx code
    self.Hrsp_2xx       = "" -- http responses with 2xx code
    self.Hrsp_3xx       = "" -- http responses with 3xx code
    self.Hrsp_4xx       = "" -- http responses with 4xx code
    self.Hrsp_5xx       = "" -- http responses with 5xx code
    self.QTime          = ""    -- [..BS]: the average queue time in ms over the 1024 last requests
    self.CTime          = ""    -- [..BS]: the average connect time in ms over the 1024 last requests
    self.RTime          = ""    -- [..BS]: the average response time in ms over the 1024 last requests (0 for TCP)
    self.TTime          = ""    -- [..BS]: the average total session time in ms over the 1024 last requests
    self.AgentStatus    = ""   -- status of last agent check
    self.Addr           = ""   -- ip address & port 
    self.AgentDur       = ""    -- [...S]: time in ms taken to finish last check
    self.LastSess       = ""    -- [..BS]: number of seconds since last session assigned to server/backend
end

-- haproxy
local Stats = Class()

-- stats init, create default value for attributes
function Stats:_init()
    -- all frontends that can not connect to backend
    -- key/value
    self.Frontends = {}
    -- all backends that can not connect to frontend
    -- key/value
    self.Backends = {}
end

function Stats:AddFrontend(frontend)
    self.Frontends[frontend.Name] = frontend
end

function Stats:AddBackend(backend)
    if self.Backends[backend.Name] then
        local oldBackend = self.Backends[backend.Name]
        backend.Servers = oldBackend.Servers
        self.Backends[backend.Name] = backend
        return
    end
    self.Backends[backend.Name] = backend
end

function Stats:AddServer(server)
    if self.Backends[server.Pxname] then
        table.insert(self.Backends[server.Pxname].Servers, server)
        -- self.Backends[server.Pxname].Servers[server.Name] = server
        return
    end
    local back = Backend()
    -- back.Servers[server.Name] = server
    table.insert(back.Servers, server)
    self.Backends[server.Pxname] = back
end

return {
    Stats = Stats,
    Info = Info,
    Frontend = Frontend,
    Backend = Backend,
    Server = Server,
}
