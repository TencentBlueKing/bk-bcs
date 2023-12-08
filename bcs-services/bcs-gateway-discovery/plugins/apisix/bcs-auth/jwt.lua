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
local error = error

local pcall = pcall
local ngx_decode_base64 = ngx.decode_base64
local ngx_time = ngx.time

local plugin_error_msg = "BCS Auth Plugin Error"


-- 生成 redis client
local function get_redis_client(conf)
    local red = redis_new()

    red:set_timeout(1000)

    local ok, err = red:connect(conf.redis_host, conf.redis_port)
    if not ok then
        core.log.error("failed to connect redis, err:", err)
        error("failed to connect redis")
    end

    local res, err = red:auth(conf.redis_password)
    if not res then
        core.log.error("failed to authenticate redis, err:", err)
        error("failed to authenticate redis")
    end
    red:select(conf.redis_database)

    return red
end

local function retry(max_attempts, func, ...)
    local args = { ... }
    local attempts = 0
    repeat
        attempts = attempts + 1
        local success, result = pcall(func, unpack(args))
        if success then
            return result
        end
        if attempts >= max_attempts then
            error("Retry failed after " .. attempts .. " attempts")
        end
    until false
end

local function conn_and_get(key, conf)
    local ok, red = pcall(get_redis_client, conf)
    if not ok then
        core.log.error("failed to connect redis:", red)
        error("failed to connect redis")
    end

    local jwt_token, err = red:get(key)
    if not jwt_token then
        core.log.error("failed to get jwt_token, err: ", err)
        error("failed to get jwt_token")
    end

    local ok, err = red:set_keepalive(10000, 100) -- tcp status : TIME_WAIT
    if not ok then
        core.log.error("failed to set keepalive:", err)
    end
    return jwt_token
end


local function conn_and_set(key, jwt_token, conf)
    local ok, red = pcall(get_redis_client, conf)
    if not ok then
        core.log.error("failed to connect redis:", red)
        error("failed to connect redis")
    end

    -- redis key 过期时间设置为 conf.exp-10 秒，防止 jwt exp 过期问题
    local exp = conf.exp - 10
    local setok, err = red:set(key, jwt_token, "EX", exp)
    if not setok then
        core.log.error("failed to set jwt_token, err: ", err)
        error("failed to set jwt_token")
    end

    local ok, err = red:set_keepalive(10000, 100) -- tcp status : TIME_WAIT
    if not ok then
        core.log.error("failed to set keepalive:", err)
    end
    return setok
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


local function get_real_payload(userinfo, auth_conf)
    local retTable = {
        sub_type = "user",
        exp = ngx_time() + auth_conf.exp,
        iss = "bcs-auth-plugin",
    }
    if type(userinfo) == "string" then
        retTable["username"] = userinfo
    elseif type(userinfo) == "number" then
        retTable["username"] = tostring(userinfo)
    elseif type(userinfo) == "table" then
        for key, value in pairs(userinfo) do
            retTable[key] = value
        end
    end

    return retTable
end


local function sign_jwt_with_RS256(userinfo, auth_conf)
    local auth_secret = get_secret(auth_conf)
    local ok, jwt_token = pcall(jwt.sign, _M,
            auth_secret,
            {
                header = {
                    typ = "JWT",
                    alg = "RS256"
                },
                payload = get_real_payload(userinfo, auth_conf)
            }
    )
    if not ok then
        core.log.error("failed to sign jwt, err: ", jwt_token.reason)
        core.response.exit(500, plugin_error_msg)
    end
    return jwt_token
end


local _M = {}


function _M:get_jwt_from_redis(credential, conf, ctx, key_prefix, create_if_null, get_userinfo_handler)
    local key = key_prefix
    if not credential.redis_key then
        key = key .. credential.user_token
    else
        key = key .. credential.redis_key
    end

    local jwt_token = retry(3, conn_and_get, key, conf)
    if not jwt_token then
        core.log.error("failed to get jwt_token, err: ", err)
        core.response.exit(500, plugin_error_msg)
    end
    -- redis 的 key 过期或者并未创建

    if (jwt_token == ngx.null and create_if_null) then
        local userinfo
        if conf.bk_login_host_esb then
            local data = get_userinfo_handler(credential, conf)
            if data["result"] then
                userinfo = data["data"]["bk_username"]
            end
            if ctx ~= nil then
                ctx.var["bk_login_code"] = data["code"]
                ctx.var["bk_login_message"] = data["message"]
            end
        else
            userinfo = get_userinfo_handler(credential, conf.bk_login_host)
        end

        if userinfo then
            jwt_token = sign_jwt_with_RS256(userinfo, conf)

            local ok = retry(3, conn_and_set, key, jwt_token, conf)
            if not ok then
                core.log.error("failed to set jwt_token, err: ", err)
                core.response.exit(500, plugin_error_msg)
            end
        end
    end

    if jwt_token == ngx.null then
        return nil
    end

    return jwt_token
end


return _M
