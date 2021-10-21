-- Simple HTTP Request for BCS

local Class = require("pl.class")
local Stringx = require("pl.stringx")
local Url = require("pl.url")

local Request = Class()

function Request:_init(method, path, header, args)
    self.Method = method or ""
    self.Path = path or ""
    self.PathParameter = {}
    self.Header = header or {}
    self.Args = parseArgs(args)
    self.Data = ""
end

function Request:UpdateArgs()
    -- body
end

-- setting request data when in POST/PUT/PATCH
function Request:SetData(data)
    if not data then
        return
    end
    if self.Method == "POST" or self.Method == "PUT" or self.Method == "PATCH" then
        self.Data = data
    end
end

function parseArgs(args)
    local t = {}
    if not args then
        return t
    end
    local data = Stringx.split(Url.unquote(args), '&')
    for _, pair in pairs(data) do
        local k, _, v = Stringx.partition(pair, '=')
        t[k] = v
    end
    return t
end

return Request