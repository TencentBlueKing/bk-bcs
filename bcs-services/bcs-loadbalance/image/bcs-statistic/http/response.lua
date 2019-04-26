-- Simple HTTP Response for BCS

local Class = require("pl.class")
local Tablex = require("pl.tablex")

local Response = Class()

function Response:_init(status, body, header)
    self.StatusCode = status or 500
    self.Body = body or ""
    -- setting data length by default
    local defualtHeader = {["Content-Length"] = self.Body:len()}
    self.Header = Tablex.update(defualtHeader, header or {})
end

return Response