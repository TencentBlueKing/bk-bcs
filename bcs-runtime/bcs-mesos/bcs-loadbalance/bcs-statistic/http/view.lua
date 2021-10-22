-- View is implementation for GET/POST/DELETE method

local Class = require("pl.class")

local View  = Class()

function View:_init(name)
    self.name = name or "basicView"
end

-- register all method callback with router
function View:Register(router)
    -- body
end

-- interface for handling get
-- @return response, see @http.response
function View:Get(request)
    -- body
    return nil
end

-- interface for handling POST
-- @return response, see @http.response
function View:Post(request)
    -- body
    return nil
end

-- interface for handling Delete
-- @return response, see @http.response
function View:Delete(request)
    -- body
    return nil
end

-- interface for handling put
-- @return response, see @http.response
function View:Put(request)
    -- body
    return nil
end

return View