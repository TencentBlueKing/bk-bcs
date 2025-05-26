-- * Tencent is pleased to support the open source community by making Blueking Container Service available.
-- Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
-- Licensed under the MIT License (the "License"); you may not use this file except
-- in compliance with the License. You may obtain a copy of the License at
-- http://opensource.org/licenses/MIT
-- Unless required by applicable law or agreed to in writing, software distributed under
-- the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
-- either express or implied. See the License for the specific language governing permissions and
-- limitations under the License.
--
local BKUserCli = require("apisix.plugins.bkbcs-auth.bkbcs")
local bcs_jwt = require("apisix.plugins.bcs-auth.jwt")
local resty_jwt = require("resty.jwt")
local core = require("apisix.core")
local ngx = ngx
local attr = {}
local token_user_map_cache = core.lrucache.new(
    {
        ttl = 300,
        count = 1000,
    }
)

local schema = {
    type = "object",
    properties = {
        bkbcs_auth_endpoints = {
            description = "bkbcs auth endpoint",
            type = "string",
            minLength = 1,
            maxLength = 125,
        },
        module = {
            description = "bk bcs module name",
            type = "string",
            minLength = 2,
            maxLength = 64,
        },
        token = {
            description = "token for auth module api request",
            type = "string",
            minLength = 16,
            maxLength = 64,
            pattern = "^[0-9a-zA-Z-.]+$",
        },
        retry_count = {
            description = "auth retry time when http request failed",
            type = "number",
            minimum = 1,
            maxnum = 3,
        },
        timeout = {
            description = "timeout seconds for request to auth module",
            type = "number",
            minimum = 1,
            maxnum = 10,
            default = 10,
        },
        keepalive = {
            description = "keepalive seconds for request to auth module",
            type = "number",
            minimum = 1,
            maxnum = 60,
            default = 60,
        },
        redis_host = {
            type = "string",
            description = "redis for bcs-auth plugin: host",
        },
        redis_port = {
            type = "integer",
            default = 6379,
            description = "redis for bcs-auth plugin: port",
        },
        redis_password = {
            type = "string",
            description = "redis for bcs-auth plugin: password",
        },
        redis_database = {
            type = "integer",
            default = 0,
            description = "redis for bcs-auth plugin: database num",
        },
    },
    required = {
        "bkbcs_auth_endpoints",
        "module",
    },
    additionalProperties = false,
}

local attr_schema = {
    type = "object",
    properties = {
        gateway_token = {
            type = "string",
            description = "token for auth module api request",
        },
        usermanager_upstream_name = {
            type = "string",
            default = "usermanager",
            description = "upstream name for usermanager",
        },
    },
}

local plugin_name = "bkbcs-auth"

local _M = {
    version = 0.1,
    priority = 2788,
    name = plugin_name,
    schema = schema,
}

function _M.check_schema(conf)
    local ok, err = core.schema.check(schema, conf)
    if not ok then
        return false, err
    end
    return true, nil
end

local function get_username_from_redis(token, conf, ctx)
    local jwt_str = bcs_jwt:get_jwt_from_redis(
        {
            user_token = token,
        }, conf, ctx, "bcs_auth:token:", false
    )
    if not jwt_str then
        return nil, ""
    end
    local jwt_obj = resty_jwt:load_jwt(jwt_str)
    if not jwt_obj then
        core.log.error("load jwt from jwt token failed, jwt token:" .. jwt_str)
        return nil, ""
    end
    return jwt_obj.payload
end

-- * rewrite stage: authentication before request to Service
function _M.rewrite(conf, apictx)
    if conf == nil then
        return 503, {
            code = 503,
            result = false,
            message = "no plugin conf",
        }
    end
    if conf.token == nil then
        conf.token = attr.token
    end
    conf["usermanager_upstream_name"] = attr.usermanager_upstream_name
    if conf.bkbcs_auth_endpoints == nil or conf.token == nil then
        core.log.error("bkbcs auth rewrite configuration: ", core.json.encode(conf))
        return 503, {
            code = 503,
            result = false,
            message = "plugin configuration fatal",
        }
    end
    -- ready to creat bkbcs userclient for authentication
    local userc, err = BKUserCli:new(apictx, conf)
    if err then
        core.log.error("create user-manager client with endpoint failed: ", err)
        -- response internal error
        return 503, {
            code = 503,
            result = false,
            message = "An unexpected error occurred when authentication",
        }
    end

    -- construct request according configuration
    local req, err = userc:construct_identity(conf, ngx.req)
    if err then
        core.log.error("construct auth request for [", ngx.req.get_method(), "]", ngx.var.uri, ", err:", err)
        return 400, {
            code = 400,
            result = false,
            message = "Bad Request: " .. err,
        }
    end
    local payload, err = token_user_map_cache(req.user_token, nil, get_username_from_redis, req.user_token, conf, ctx)
    if not payload or not payload.sub_type then
        core.log.warn(
            "no username can be found for token " .. req.user_token .. ",payload: " .. core.json.encode(payload, true)
        )
    else
        apictx.var["bcs_usertype"] = payload.sub_type
        if payload.username and #payload.username > 0 then
            apictx.var["bcs_username"] = payload.username
        else
            apictx.var["bcs_username"] = payload.client_id
        end
    end
    -- init success, try to anthentication
    local ok, err = userc:authentication(conf, req)
    if err then
        return 503, {
            code = 503,
            result = false,
            message = "An unexpected error occurred in verify: " .. err,
        }
    end
    if not ok then
        core.log.warn("token is not allow to access to [", ngx.req.get_method(), "]", ngx.var.uri)
        return 401, {
            code = 401,
            result = false,
            message = "Resource is Unauthorized",
        }
    end
end

function _M.init()
    local local_conf = core.config.local_conf()
    attr = core.table.try_read_attr(local_conf, "plugin_attr", plugin_name)
    local ok, err = core.schema.check(attr_schema, attr)
    if not ok then
        core.log.error("failed to check the plugin_attr[", plugin_name, "]", ": ", err)
        return
    end
end

return _M
