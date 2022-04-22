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
local redis_new = require("resty.redis").new
local jwt = require("resty.jwt")

local pcall = pcall
local ngx_decode_base64 = ngx.decode_base64
local ngx_time = ngx.time

local plugin_error_msg = "BCS Auth Plugin Error"


-- 生成 redis client
local function get_redis_client(conf)
    local red = redis_new()

    red:set_timeout(1000)

    red:connect(conf.redis_host, conf.redis_port)
    red:auth(conf.redis_password)
    red:select(conf.redis_database)

    return red
end


local function get_secret(conf)
    if not conf.private_key then
        core.log.error("no jwt private key provided, conf:"..core.json.encode(conf, true))
        core.response.exit(500, plugin_error_msg)
    end
    local auth_secret = ngx_decode_base64(conf.private_key)
    if not auth_secret then
        return conf.private_key
    end
    return auth_secret
end


local function get_real_payload(username, auth_conf)
    return {
        sub_type = "user",
        username = username,
        exp = ngx_time() + auth_conf.exp,
        iss = "bcs-auth-plugin",
    }
end


local function sign_jwt_with_RS256(username, auth_conf)
    local auth_secret = get_secret(auth_conf)
    local ok, jwt_token = pcall(jwt.sign, _M,
        auth_secret,
        {
            header = {
                typ = "JWT",
                alg = "RS256"
            },
            payload = get_real_payload(username, auth_conf)
        }
    )
    if not ok then
        core.log.error("failed to sign jwt, err: ", jwt_token.reason)
        core.response.exit(500, plugin_error_msg)
    end
    return jwt_token
end


local _M = {}


function _M:get_jwt_from_redis(credential, conf, key_prefix, create_if_null, get_username_handler)
    local ok, red = pcall(get_redis_client, conf)
    if not ok then
        core.log.error("failed to connect redis:", red)
        core.response.exit(500, plugin_error_msg)
    end

    local key = key_prefix .. credential.user_token
    local jwt_token, err = red:get(key)
    if not jwt_token then
        core.log.error("failed to get jwt_token, err: ", err)
        core.response.exit(500, plugin_error_msg)
    end
    -- redis 的 key 过期或者并未创建

    if (jwt_token == ngx.null and create_if_null) then
        local username = get_username_handler(credential, conf.bk_login_host)
        if username then
            jwt_token = sign_jwt_with_RS256(username, conf)

            local ok, err = red:set(key, jwt_token, "EX", conf.exp)
            if not ok then
                core.log.error("failed to set jwt_token, err: ", err)
                core.response.exit(500, plugin_error_msg)
            end
        end
    end

    local ok, err = red:set_keepalive(10000, 100) -- tcp status : TIME_WAIT
    if not ok then
        core.log.error("failed to set keepalive:", err)
        core.response.exit(500, plugin_error_msg)
    end

    if jwt_token == ngx.null then
        return nil
    end

    return jwt_token
end


return _M
