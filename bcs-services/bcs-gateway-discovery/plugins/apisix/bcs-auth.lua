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
        bk_login_host_tenant = {type = "string", description = "bk login host tenant"},
        bk_login_host_esb = {type = "string", description = "bk login host esb"},
        bk_login_host = {type = "string", description = "bk login host (with scheme prefix)"},
        private_key = {type = "string", description = "jwt private_key"},
        bkapigw_jwt_verify_key = {type = "string", description = "jwt verify key for bkapigw"},
        bkapigw_jwt_verify_key_map = {type = "object", description = "jwt verify keys for bkapigw"},
        exp = {type = "integer", default = 300, description = "jwt exp time in seconds"},
        -- redis backend config
        redis_host = {type = "string", description = "redis for bcs-auth plugin: host"},
        redis_port = {type = "integer", default = 6379, description = "redis for bcs-auth plugin: port"},
        redis_password = {type = "string", description = "redis for bcs-auth plugin: password"},
        redis_database = {type = "integer", default = 0, description = "redis for bcs-auth plugin: database num"},
        run_env = {type = "string", default = "ce", description = "apisix on ce or cloud env"},
        bk_app_code = {type = "string", description = "bk_app_code"},
        bk_app_secret = {type = "string", description = "bk_app_secret"}
    },
    required = {"bk_login_host", "private_key", "redis_host", "redis_password"},
}


local _M = {
    version = 0.2,
    priority = 2788,
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

-- 获取正确的 scheme，优先使用 X-Forwarded-Proto header（当 TLS 在 CLB 上终止时）
local function get_scheme(ctx)
    -- 优先检查 X-Forwarded-Proto header（CLB 通常会设置此 header）
    local forwarded_proto = core.request.header(ctx, "X-Forwarded-Proto")
    -- core.log.warn("X-Forwarded-Proto: ", tostring(forwarded_proto), ", ctx.var.scheme: ", tostring(ctx.var.scheme))
    
    if forwarded_proto then
        forwarded_proto = string.lower(forwarded_proto)
        if forwarded_proto == "https" or forwarded_proto == "http" then
            core.log.warn("Using X-Forwarded-Proto scheme: ", forwarded_proto)
            return forwarded_proto
        end
    end
    
    -- 回退到 ctx.var.scheme
    local fallback_scheme = ctx.var.scheme or "http"
    core.log.warn("Using fallback scheme: ", fallback_scheme)
    return fallback_scheme
end


local function concat_login_uri(conf, ctx)
    local scheme = get_scheme(ctx)
    local c_url = tab_concat({scheme, "://", ctx.var.host, ngx_escape_uri(ctx.var.request_uri)})
    return tab_concat({conf.bk_login_host, "/plain/?size=big&c_url=", c_url})
end

local function is_ajax_request(ctx)
    local x_requested_with = core.request.header(ctx, "X-Requested-With") or ""
    return x_requested_with:lower() == "xmlhttprequest"
end

local function redirect_login(conf, ctx)
    if is_ajax_request(ctx) then
        return 401, {message = "bcs-auth plugin error: token is expired or is invalid"}
    end
    core.response.set_header("Location", concat_login_uri(conf, ctx))
    return 302
end

local function redirect_403(conf, ctx)
    local msg_base64 = ngx.encode_base64(tostring(ctx.var.bk_login_message))
    local msg_base64_uri = ngx_escape_uri(msg_base64)
    local scheme = get_scheme(ctx)
    core.response.set_header("Location", tab_concat({scheme, "://", ctx.var.host, "/403.html?msg=", msg_base64_uri}))
    return 302
end

function _M.rewrite(conf, ctx)
    local user_agent = core.request.header(ctx, "User-Agent")
    local use_login = is_from_browser(user_agent) -- 是否对接蓝鲸统一登录

    local auth = authentication:new(use_login, conf.run_env)
    local jwt_token = auth:authenticate(conf, ctx)

    --core.log.warn("rewrite, code: " .. tostring(ctx.var.bk_login_code) .. " msg: "
    --        .. tostring(ctx.var.bk_login_message) .. " esb: " .. tostring(conf.bk_login_host_esb))

    -- 1302403 用户认证成功，但用户无应用访问权限，跳转到403页面
    if ctx.var.bk_login_code == 1302403 then
        return redirect_403(conf, ctx)
    end

    -- 1302100 用户认证失败，即用户登录态无效，跳转到login页面
    if ctx.var.bk_login_code == 1302100 then
        return redirect_login(conf, ctx)
    end

    --core.log.warn("jwt_token: ", tostring(jwt_token))

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
