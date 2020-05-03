--* Tencent is pleased to support the open source community by making Blueking Container Service available.
-- Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
-- Licensed under the MIT License (the "License"); you may not use this file except
-- in compliance with the License. You may obtain a copy of the License at
-- http://opensource.org/licenses/MIT
-- Unless required by applicable law or agreed to in writing, software distributed under
-- the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
-- either express or implied. See the License for the specific language governing permissions and
-- limitations under the License.
--

local basic = require "kong.plugins.base_plugin"
local BKUserCli = require "kong.plugins.bkbcs-auth.bkbcs"
local kong = kong

local BKBCSAuthHandler = basic:extend()
BKBCSAuthHandler.PRIORITY = 1003
BKBCSAuthHandler.VERSION = "0.1.0"

function BKBCSAuthHandler:new(name)
  BKBCSAuthHandler.super.new(self, "bkbcs-auth")
end

-- * access authentication before request to Service
function BKBCSAuthHandler:access(conf)
  BKBCSAuthHandler.super.access(self, conf)
  -- ready to creat bkbcs userclient for authentication
  local userc = BKUserCli(conf.bkbcs_auth_endpoints)
  local err = userc:init(conf)
  if err then
    kong.log.err("init user-manager client with endpoint [", conf.bkbcs_auth_endpoints, "] failed: ", err)
    -- response internal error
    return kong.response.exit(500, {code = 400, result = false, message = "An unexpected error occurred when authentication"})
  end
  -- construct request according configuration
  local req, err = userc:construct_identity(conf, kong.request)
  if err then
    kong.log.err("construct auth request for [", kong.request.get_method(), "]", kong.request.get_path(), ", err:", err)
    return kong.response.exit(400, {code = 400, result = false, message = "Bad Request: " .. err})
  end
  -- init success, try to anthentication
  local ok, err = userc:authentication(conf, req)
  if err then
    return kong.response.exit(500, {code = 500, result = false, message = "An unexpected error occurred in verify: " .. err})
  end
  if not ok then
    kong.log.warn("token is not allow to access to [", kong.request.get_method(), "]", kong.request.get_path())
    return kong.response.exit(401, {code = 401, result = false, message = "Resource is Unauthorized"})
  end
end

return BKBCSAuthHandler
