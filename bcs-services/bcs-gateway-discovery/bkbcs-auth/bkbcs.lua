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

local Object = require "kong.vendor.classic"
local kongBalancer = require "kong.runloop.balancer"
local kong = kong
local http = require "resty.http"
local url = require "socket.url"
local json = require "cjson"

local BKUserCli = Object:extend()

-- more module definition check https://github.com/Tencent/bk-bcs/blob/master/bcs-common/common/types/serverInfo.go
local KUBEAGENT = "kubeagent"
local MESOSDRIVER = "mesosdriver"
local KUBEDRIVER = "kubernetedriver"
local MESHMANAGER = "meshmanager"
local LOGMANAGER = "logmanager"
local STORAGE = "storage"
local NETWORKDETECTION = "networkdetection"
local CLUSTER_HEADER = "BCS-ClusterID"

-- bcs-user-manager permission verify info
local USERMGR_METHOD = "GET"
local USERMGR_URL = "/usermanager/v1/permissions/verify"
local USERMGR_HOST = "usermanager.bkbcs.tencent.com"

local userTarget = {
  type = "name",
  host = USERMGR_HOST,
  try_count = 0,
}

local function is_cluster_resource(mod)
  return mod == KUBEAGENT or mod == MESOSDRIVER or mod == KUBEDRIVER
end

local function is_no_cluster_header(mod)
  return mod == MESHMANAGER or 
  mod == LOGMANAGER or mod == KUBEAGENT or 
  mod == NETWORKDETECTION
end

-- Parse user-manager url
-- @param `endpoints` http(s)://host:port/url
-- @return `parsed_url` a table with host details: scheme, host, port, path, query, userinfo
local function parse_url(endpoints)
  local parsed_url = url.parse(endpoints)
  if not parsed_url.host then
    parsed_url.host = USERMGR_HOST
  end
  if not parsed_url.port then
    if parsed_url.scheme == "http" then
      parsed_url.port = 80
    elseif parsed_url.scheme == "https" then
      parsed_url.port = 443
    end
  end
  if not parsed_url.path then
    parsed_url.path = "/"
  end
  return parsed_url
end

function BKUserCli:new(name)
  self.name = name
end

-- get ring balancer instance from local global cache
--@return: ip, port, error if happened
function BKUserCli:instance_balancer(key)
  if key then
    userTarget.host = key
  end
  local ok, msg, code = kongBalancer.execute(userTarget)
  if not ok then
    kong.log.err("user_cli get upstream ", key, " balance target failed[", code, "] ", msg)
    return nil, nil, "balancer error"
  end
  return userTarget.ip, userTarget.port, nil
end

-- init http client according kong balancer setting
-- @param endpoints: comes from plugin configuration
function BKUserCli:init(config)
  -- parse endpoint for request details
  local parsed_url = parse_url(config.bkbcs_auth_endpoints)
  local ip, port, err = self:instance_balancer(parsed_url.host)
  if err then
    return err
  end
  -- init http for authentication later
  self.httpc = http.new()
  self.httpc:set_timeout(config.timeout)
  local ok, err = self.httpc:connect(ip, port)
  if not ok then
    kong.log.err(" user_cli connect to ", ip, ":", port, " failed, ", err)
    return err
  end
  if parsed_url.scheme == "https" then
    local _, err = self.httpc:ssl_handshake(true, parsed_url.host, false)
    if err then
      kong.log.err(" user_cli ssl handshake with ", ip, ":", port, " failed, ", err)
      return err
    end
  end
  return nil
end

-- contruct identity information for authentication
-- @param conf: bkbcs-auth plugin configuration
-- @param request: http request
-- @return: a table that use for `bkbcs_user_cli:authentication`
--        and error string if error happened
function BKUserCli:construct_identity(conf, request)
  if not conf or not request then
    return nil, "lost conf or request"
  end
  local auth = {
    user_token = "",
    resource_type = "",
    resource = "",
    action = "",
  }
  auth.action = request.get_method()
  if is_cluster_resource(conf.module) then
    auth.resource_type = "cluster"
  else
    auth.resource_type = conf.module
  end
  -- kubeagent & networkdetection has no ClusterId
  local cluster_id = request.get_header(CLUSTER_HEADER)
  if not cluster_id and 
  not is_no_cluster_header(conf.module) then
    kong.log.err(" user_cli get no BCS-ClusterID from request ", request.get_path())
    return nil, "lost BCS-ClusterID in header"
  end
  if conf.module == KUBEAGENT then
    -- retrieve bcs-cluster-id from url as resource
    local id_iterator, id_err = ngx.re.gmatch(request.get_path(), "BCS-K8S-([0-9]+[^/])")
    if not id_iterator then
      kong.log.err(" user_cli get no BCS-ClusterID in kubernetes request ", request.get_path(), " error: ", id_err)
      return nil, "lost BCS-ClusterID in url"
    end
    local id, err = id_iterator()
    if not id or #id < 1 then
      kong.log.err(" user_cli parse kubernetes BCS-ClusterID in request ", request.get_path(), " failed, ", err)
      return nil, "kuberentes BCS-ClusterID "
    end
    auth.resource = id[0]
  else
    -- retrieve bcs-cluster-id from header
    auth.resource = cluster_id
  end

  -- get token for http header
  local auth_header = request.get_header("Authorization")
  if not auth_header then
    kong.log.err(" user_cli get no Authorization from http header, request path:", request.get_path())
    return nil, "lost Authorization"
  end
  local iterator, iter_err = ngx.re.gmatch(auth_header, "\\s*[Bb]earer\\s+(.+)")
  if not iterator then
    kong.log.err(" user_cli search token for request ", request.get_path(), " failed, ", iter_err)
    return nil, "Authorization format error"
  end
  local m, err = iterator()
  if not m or #m < 1 then
    kong.log.err(" user_cli get token information from request ",  request.get_path(), " failed: ", err)
    return nil, "Authorization token lost"
  end
  auth.user_token = m[1]
  return auth, nil
end

-- try to use bcs-user-manager to authorize
-- bkbcs authentication link: http(s)://usermanager.bkbcs.tencent.com:8080/usermanager/v1/permissions/verify
--* @param  requset: '{"user_token":"", "resource_type":"cluster", "resource":"clsuterId", "action":"POST"}'
--* @return true/false: return true if authentication success, other false.
-- example: curl -H"Content-Type: application/json" http://127.0.0.L1:8080/usermanager/v1/permissions/verify -d \
--          -H "Authorization: Bearer ${token}" -d '{"user_token":"", "resource_type":"cluster", "resource":"clsuterId", "action":"POST"}'
-- response: {"result":true,"code":0,"message":"success","data":{"allowed":false,"message":"usertoken is invalid"}}
function BKUserCli:authentication(conf, info)
  -- ready to send http authentication information
  if self.httpc == nil then
    kong.log.err(" user_cli need to be initialized...")
    return false, "not initialization"
  end
  -- serialization from table to string
  local payload = json.encode(info)
  local res, err = self.httpc:request({
    method = USERMGR_METHOD,
    path = USERMGR_URL,
    headers = {
      ["Content-Type"] = "application/json",
      ["Accept"] = "application/json",
      ["Content-Length"] = #payload,
      ["Authorization"] = "Bearer " .. conf.token,
    },
    body = payload,
  })
  if not res then
    kong.log.err(" user_cli send request to user-manager failed, ", err, ". info: ", self.ip, ":", tostring(self.port))
    return false, "authentication connection failed" .. ": " .. err
  end
  if res.status >= 400 then
    kong.log.err(" user_cli send request to user-manager failed, http code: ", tostring(res.status))
    return false, "authentication internal err: " .. tostring(res.status)
  end
  local body = res:read_body()
  local auth_response = json.decode(body)
  if not auth_response or not auth_response.data then
    kong.log.err(" user_cli get no correct response from user-manager, body: ", body)
    return false, "invalid verify response"
  end
  -- connection setting keepalive
  local ok, err = self.httpc:set_keepalive(conf.keepalive)
  if not ok then
    -- the batch might already be processed at this point, so not being able to set the keepalive
    -- will not return false (the batch might not need to be reprocessed)
    kong.log.err(" user_cli failed keepalive for ", self.ip, ":", tostring(self.port), ": ", err)
  end
  return auth_response.data.allowed, nil
end

return BKUserCli