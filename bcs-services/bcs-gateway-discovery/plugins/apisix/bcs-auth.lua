--
-- Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
-- Edition) available.
-- Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
-- Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://opensource.org/licenses/MIT
--
-- Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
-- an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
-- specific language governing permissions and limitations under the License.
--
local core = require("apisix.core")
local stringx = require('pl.stringx')
local authentication = require("apisix.plugins.bcs-auth.authentication")

local ngx = ngx
local ngx_escape_uri = ngx.escape_uri
local tab_concat = table.concat
local plugin_name = "bcs-auth"

local schema = {
    type = "object",
    properties = {
        bk_login_host = {type = "string", description = "bk login host (with scheme prefix)"},
        private_key = {type = "string", description = "jwt private_key"},
        bkapigw_jwt_verify_key = {type = "string", description = "jwt verify key for bkapigw"},
        exp = {type = "integer", default = 300, description = "jwt exp time in seconds"},
        -- redis backend config
        redis_host = {type = "string", description = "redis for bcs-auth plugin: host"},
        redis_port = {type = "integer", default = 6379, description = "redis for bcs-auth plugin: port"},
        redis_password = {type = "string", description = "redis for bcs-auth plugin: password"},
        redis_database = {type = "integer", default = 0, description = "redis for bcs-auth plugin: database num"},
        run_env = {type = "string", default = "ce", description = "apisix on ce or cloud env"}
    },
    required = {"bk_login_host", "private_key", "redis_host", "redis_password"}
}


local _M = {
    version = 0.2,
    priority = 12,
    name = plugin_name,
    schema = schema,
}


function _M.check_schema(conf)
    return core.schema.check(schema, conf)
end


-- 判断 client 的类型
local function is_from_browser(user_agent)
    if not user_agent then
        return false
    end
    return stringx.startswith(user_agent, "Mozilla")
end


local function concat_login_uri(conf, ctx)
    local c_url = tab_concat({ctx.var.scheme, "://", ctx.var.host, ngx_escape_uri(ctx.var.request_uri)})
    return tab_concat({conf.bk_login_host, "/plain/?size=big&c_url=", c_url})
end


local function redirect_login(conf, ctx)
    core.response.set_header("Location", concat_login_uri(conf, ctx))
    return 302
end


function _M.rewrite(conf, ctx)
    local user_agent = core.request.header(ctx, "User-Agent")
    local use_login = is_from_browser(user_agent) -- 是否对接蓝鲸统一登录

    local auth = authentication:new(use_login, conf.run_env)
    local jwt_token = auth:authenticate(conf, ctx)

    if not jwt_token then
        if use_login then
            -- TODO 处理成无权限的数据返回 return 401, core.json.encode({code=40101, data={login_url={full='', simple=''}}})
            return redirect_login(conf, ctx)
        end
        return 401, {message = "bcs-auth plugin error: token is expired or is invalid"}
    else
        core.request.set_header(ctx, "Authorization", "Bearer " .. jwt_token)
    end

end


return _M
