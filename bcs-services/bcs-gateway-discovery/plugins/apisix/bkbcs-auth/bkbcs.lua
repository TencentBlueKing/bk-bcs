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

local apiBalancer = require("apisix.balancer")
local core = require("apisix.core")
local http = require("resty.http")
local url = require("net.url")
local ngx_re = require("ngx.re")
local ngx = ngx

-- more module definition check https://github.com/Tencent/bk-bcs/blob/master/bcs-common/common/types/serverInfo.go
local KUBEAGENT = "kubeagent"
local MESOSDRIVER = "mesosdriver"
--[[
local KUBEDRIVER = "kubernetedriver"
local MESHMANAGER = "meshmanager"
local LOGMANAGER = "logmanager"
local STORAGE = "storage"
local NETWORKDETECTION = "networkdetection"
]]--
local CLUSTER_HEADER = "BCS-ClusterID"

-- bcs-user-manager permission verify info
local USERMGR_METHOD = "GET"
local USERMGR_URL = "/usermanager/v1/permissions/verify"
local USERMGR_HOST = "usermanager"

local userTarget = {
  type = "name",
  host = USERMGR_HOST,
  try_count = 0,
}

local function is_cluster_resource(mod)
  return mod == KUBEAGENT or mod == MESOSDRIVER
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
    parsed_url.path = USERMGR_URL
  end
  return parsed_url
end

-- get ring balancer instance from local global cache
--@return: ip, port, error if happened
local function instance_balancer(key)
  if not key then
    core.log.error("no host information to get bcs-user-manager balance instance")
    return nil, nil, "lost user-manager host"
  end
  -- get upstream from local cache data
  local upstreams = core.config.fetch_created_obj("/upstreams")
  if not upstreams then
    return nil, nil, "no user-manager registe"
  end
  local userkey = ngx_re.split(key, "\\.", nil, nil, 3)
  local userUpstream = upstreams:get(userkey[1])
  if not userUpstream then
    return nil, nil, "search usermanager " .. userkey[1] .. " failed"
  end
  for i, _ in ipairs(userUpstream.value.nodes) do
    if not userUpstream.value.nodes[i].priority then
      userUpstream.value.nodes[i].priority = 0
    end
  end
  local user_conf = {
    type = userUpstream.value.type,
    nodes = userUpstream.value.nodes,
  }
  local cxt = {
    upstream_conf = user_conf,
    upstream_version = userUpstream.modifiedIndex,
    upstream_key = "bkbcs-" .. userkey[1],
  }
  local server, err = apiBalancer.pick_server({}, cxt)
  if err then
    core.log.error("user_cli get upstream ", key, " balance target failed, ", err)
    return nil, nil, "user balance error: " .. err
  end
  return server.host, server.port, nil
end

-- bk auth cli
local BKUserCli = {}
local BKUserCli_mt = {
    __index = BKUserCli
}

function BKUserCli:new(conf)
  -- parse endpoint for request details
  local parsed_url = parse_url(conf.bkbcs_auth_endpoints)
  local ip, port, err = instance_balancer(parsed_url.host)
  if err then
    return nil, err
  end
  local cli = {
    url = parsed_url,
    ip = ip,
    port = port,
    timeout = conf.timeout,
  }
  return setmetatable(cli, BKUserCli_mt), nil
end

-- init http client according kong balancer setting
-- @param endpoints: comes from plugin configuration
function BKUserCli:init()
  -- init http for authentication later
  self.httpc = http.new()
  self.httpc:set_timeout(self.timeout * 1000)
  local ok, err = self.httpc:connect(self.ip, self.port)
  if not ok then
    core.log.error("user_cli connect to ", self.ip, ":", self.port, " failed, ", err)
    return err
  end
  if self.url.scheme == "https" then
    local _, err = self.httpc:ssl_handshake(true, self.url.host, false)
    if err then
      core.log.error("user_cli ssl handshake with ", self.ip, ":", self.port, " failed, ", err)
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
  -- kubeagent & mesosdriver construct ClusterId
  if conf.module == KUBEAGENT then
    -- retrieve bcs-cluster-id from url as resource
    local id_iterator, id_err = ngx.re.gmatch(ngx.var.uri, "BCS-K8S-([0-9]+[^/])")
    if not id_iterator then
      core.log.error(" user_cli get no BCS-ClusterID in kubernetes request ", ngx.var.uri, " error: ", id_err)
      return nil, "lost BCS-ClusterID in url"
    end
    local id, err = id_iterator()
    if not id or #id < 1 then
      core.log.error(" user_cli parse kubernetes BCS-ClusterID in request ", ngx.var.uri, " failed, ", err)
      return nil, "kuberentes BCS-ClusterID parse failed"
    end
    auth.resource = id[0]
  elseif conf.module == MESOSDRIVER then
    local headers = request.get_headers()
    if not headers[CLUSTER_HEADER] then
      core.log.error(" user_cli get no BCS-ClusterID from request ", ngx.var.uri)
      return nil, "lost BCS-ClusterID in header"
    end
    -- retrieve bcs-cluster-id from header
    auth.resource = headers[CLUSTER_HEADER]
  end

  -- get token for http header
  local auth_header = request.get_headers()
  if not auth_header["Authorization"] then
    core.log.error(" user_cli get no Authorization from http header, request path:", ngx.var.uri, " header details: ", core.json.encode(auth_header))
    return nil, "lost Authorization"
  end
  local iterator, iter_err = ngx.re.gmatch(auth_header["Authorization"], "\\s*[Bb]earer\\s+(.+)")
  if not iterator then
    core.log.error(" user_cli search token for request ", ngx.var.uri, " failed, ", iter_err, " Authorization: ", auth_header["Authorization"])
    return nil, "Authorization format error"
  end
  local m, err = iterator()
  if not m or #m < 1 then
    core.log.error(" user_cli get token information from request ",  ngx.var.uri, " failed: ", err)
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
  if not conf or not conf.token then
    return false, "lost plugin configuration"
  end
  -- serialization from table to string
  local payload = core.json.encode(info)
  core.log.debug("authentication payload: ", payload)
  local httpc = http.new()
  httpc:set_timeout(conf.timeout * 1000)
  local params = {
    method = USERMGR_METHOD,
    body =  payload,
    ssl_verify = false,
    keepalive = conf.keepalive * 1000,
    headers = {
      ["Content-Type"] = "application/json",
      ["Accept"] = "application/json",
      ["Content-Length"] = #payload,
      ["Authorization"] = "Bearer " .. conf.token
    }
  }
  local req_endpoint = self.url.scheme .. "://" .. self.ip .. ":" .. tostring(self.port) .. USERMGR_URL
  local httpc_res, httpc_err = httpc:request_uri(req_endpoint, params)
  if not httpc_res then
    core.log.error(" user_cli send request to user-manager failed, ", httpc_err, ". info: ", req_endpoint)
    return false, "authentication connection failed" .. ": " .. httpc_err
  end
  if httpc_res.status >= 400 then
    core.log.error(" user_cli send request to user-manager ", req_endpoint, " failed, http code: ", httpc_res.status)
    return false, "authentication internal err: " .. tostring(httpc_res.status)
  end

  local auth_response = core.json.decode(httpc_res.body)
  if not auth_response or not auth_response.data then
    core.log.error(" user_cli get no correct response from user-manager, body: ", httpc_res.body, " endpoint: ", req_endpoint)
    return false, "invalid verify response"
  end
  core.log.debug("auth result: ", auth_response.data.allowed, " details: ", payload)
  return auth_response.data.allowed, nil
end

return BKUserCli
