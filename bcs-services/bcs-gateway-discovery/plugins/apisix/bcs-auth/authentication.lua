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
local ck = require("resty.cookie")
local stringx = require('pl.stringx')
local jwt = require("apisix.plugins.bcs-auth.jwt")
local bklogin = require("apisix.plugins.bcs-auth.bklogin")
local resty_jwt = require("resty.jwt")
local ngx_decode_base64 = ngx.decode_base64

local RUN_ON_CE = "ce" -- 表示社区版
local TOKEN_TYPE_APIGW = "apigw"
local TOKEN_TYPE_BCS = "bcs"

local bcs_token_user_map_cache = core.lrucache.new(
    {
        ttl = 300,
        count = 1000,
    }
)

------------ LoginTicketAuthentication start ------------
local LoginTicketAuthentication = {}

-- 获取 cookie 中的用户凭证信息
function LoginTicketAuthentication:fetch_credential(conf, ctx)
    local cookie, err = ck:new()
    if not cookie then
        return nil, err
    end

    local bk_ticket, err = cookie:get("bk_ticket")
    if not bk_ticket then
        if err and not stringx.startswith(err, "no cookie") then
            core.log.error("failed to fetch bk_ticket: ", err)
        end
    end

    local bk_uid, err = cookie:get("bk_uid")
    if not bk_uid then
        if err and not stringx.startswith(err, "no cookie") then
            core.log.error("failed to fetch bk_uid: ", err)
        end
    end

    return {
        user_token = bk_ticket,
        username = bk_uid,
    }

end

function LoginTicketAuthentication:injected_user_info(credential, jwt_str, conf, ctx)
    return {
        username = credential.username,
        usertype = "user",
    }
end

function LoginTicketAuthentication:get_jwt(credential, conf)
    return jwt:get_jwt_from_redis(credential, conf, "bcs_auth:session_id:", true, bklogin.get_username_for_ticket)
end
------------ LoginTicketAuthentication end ------------

------------ LoginTokenAuthentication start ------------
local LoginTokenAuthentication = {}

-- 获取 cookie 中的用户凭证信息
function LoginTokenAuthentication:fetch_credential(conf, ctx)
    local cookie, err = ck:new()
    if not cookie then
        return nil, err
    end

    local bk_token, err = cookie:get("bk_token")
    if not bk_token then
        if err and not stringx.startswith(err, "no cookie") then
            core.log.error("failed to fetch bk_token: ", err)
        end
    end
    return {
        user_token = bk_token,
    }
end

function LoginTokenAuthentication:injected_user_info(credential, jwt_str, conf, ctx)
    return {
        username = bklogin.get_username_for_token(credential, conf.bk_login_host),
        usertype = "user",
    }
end

function LoginTokenAuthentication:get_jwt(credential, conf)
    return jwt:get_jwt_from_redis(credential, conf, "bcs_auth:session_id:", true, bklogin.get_username_for_token)
end
------------ LoginTokenAuthentication end ------------

------------ TokenAuthentication start ------------
local TokenAuthentication = {}

function TokenAuthentication:fetch_credential(conf, ctx)
    local auth_header = core.request.header(ctx, "Authorization")
    if not auth_header then
        return {
            user_token = nil,
        }
    end

    local m, err = ngx.re.match(auth_header, "Bearer\\s(.+)", "jo")
    if err then
        -- error authorization
        return {
            user_token = nil,
        }
    end

    return {
        user_token = m[1],
        token_type = TOKEN_TYPE_BCS,
    }
end

local function get_username_by_TokenAuthentication_jwt(jwt_str)
    local jwt_obj = resty_jwt:load_jwt(jwt_str)
    if not jwt_obj then
        core.log.error("load jwt from apigw jwt token failed, jwt token:" .. jwt_str)
        return nil, ""
    end
    return jwt_obj.payload, nil
end

function TokenAuthentication:injected_user_info(credential, jwt_str, conf, ctx)
    local payload = bcs_token_user_map_cache(
        credential.user_token, nil, get_username_by_TokenAuthentication_jwt, jwt_str
    )
    if not payload or not payload.sub_type then
        core.log.warn(
            "no username can be found for token " .. credential.user_token .. ",payload: " ..
                core.json.encode(payload, true)
        )
    end
    local retV = {}
    retV["usertype"] = payload.sub_type
    if payload.username and #payload.username > 0 then
        retV["username"] = payload.username
    else
        retV["username"] = payload.client_id
    end
    return retV
end

function TokenAuthentication:get_jwt(credential, conf)
    return jwt:get_jwt_from_redis(credential, conf, "bcs_auth:token:", false)
end
------------ TokenAuthentication end ------------

------------ APIGWAuthentication start ------------
local APIGWAuthentication = {}

-- 获取 cookie 中的用户凭证信息
function APIGWAuthentication:fetch_credential(conf, ctx)
    local jwt_str = core.request.header(ctx, "X-Bkapi-JWT")
    if not jwt_str then
        return TokenAuthentication:fetch_credential(conf, ctx)
    end

    local jwt_obj = resty_jwt:load_jwt(jwt_str)
    if not jwt_obj then
        core.log.error("load jwt from apigw jwt token failed, jwt token:" .. jwt_str)
        core.response.exit(401, "Bad Bkapi JWT token")
    end
    local key = nil
    if conf.bkapigw_jwt_verify_key_map and conf.bkapigw_jwt_verify_key_map[jwt_obj.header.kid] then
        key = conf.bkapigw_jwt_verify_key_map[jwt_obj.header.kid]
    elseif conf.bkapigw_jwt_verify_key then
        key = conf.bkapigw_jwt_verify_key
    end
    if not key then
        core.log.error("no verify key for apigw jwt token, jwt token:" .. jwt_str)
        core.response.exit(500, "BCS Auth Plugin Error")
    end
    local decode_res = ngx_decode_base64(key)
    if decode_res ~= nil then
        key = decode_res
    end
    local jwt_obj = resty_jwt:verify_jwt_obj(key, jwt_obj)
    if not jwt_obj or not jwt_obj.verified or not jwt_obj.valid then
        core.log.error("verify apigw jwt token failed, jwt token:" .. jwt_str)
        if jwt_obj then
            core.log.error("verify apigw jwt token failed, reason:" .. jwt_obj.reason)
        end
        core.response.exit(401, "Bad Bkapi JWT token")
    end
    -- situation that anonymous requests from bkapigw will carry a token with all fields unverified, fallback to bcs token auth
    if jwt_obj.payload.app and not jwt_obj.payload.app.verified and jwt_obj.payload.user and
        not jwt_obj.payload.user.verified then
        core.log.warn("Neither app nor user has been verified, jwt obj: " .. core.json.encode(jwt_obj))
        return TokenAuthentication:fetch_credential(conf, ctx)
    end
    local redis_key = jwt_obj.payload.app.app_code
    local credential = {
        token_type = TOKEN_TYPE_APIGW,
    }
    credential["user_token"] = {
        bk_app_code = jwt_obj.payload.app.app_code,
    }
    if jwt_obj.payload.user and jwt_obj.payload.user.verified then
        credential["user_token"]["username"] = jwt_obj.payload.user.username
        redis_key = redis_key .. "," .. jwt_obj.payload.user.username
    end
    credential["redis_key"] = redis_key
    core.request.set_header(ctx, "X-Bkapi-JWT", nil)
    return credential
end

function APIGWAuthentication.get_userinfo(credential, useless)
    return credential.user_token
end

function APIGWAuthentication:injected_user_info(credential, jwt_str, conf, ctx)
    if credential.token_type == TOKEN_TYPE_BCS then
        return TokenAuthentication:injected_user_info(credential, jwt_str, conf, ctx)
    end
    if credential.user_token.username then
        return {
            username = credential.user_token.username,
            usertype = "user",
        }
    end
    return {
        username = credential.user_token.bk_app_code,
        usertype = "bk_app",
    }
end

function APIGWAuthentication:get_jwt(credential, conf)
    if not credential or not credential.user_token then
        return nil
    end
    if credential.token_type == TOKEN_TYPE_BCS then
        return TokenAuthentication:get_jwt(credential, conf)
    end
    return jwt:get_jwt_from_redis(credential, conf, "bcs_auth:apigw:", true, APIGWAuthentication.get_userinfo)
end
------------ APIGWAuthentication end ------------

local _M = {}

function _M:new(use_login, run_env)
    local o = {}
    setmetatable(o, self)

    self.__index = self

    if not use_login then
        self.backend = APIGWAuthentication
        return o
    end

    if run_env == RUN_ON_CE then
        self.backend = LoginTokenAuthentication
    else
        self.backend = LoginTicketAuthentication
    end

    return o
end

function _M:authenticate(conf, ctx)
    local credential = self.backend:fetch_credential(conf, ctx)

    if not credential or not credential.user_token then
        return nil
    end
    local jwt_str = self.backend:get_jwt(credential, conf)
    -- 向上下文注入用户信息
    if jwt_str then
        local userinfo = self.backend:injected_user_info(credential, jwt_str, conf, ctx)
        if not userinfo then
            core.log.error(
                "generate userinfo failed with credential " .. core.json.encode(credential) .. ", jwt: " .. jwt_str
            )
            return jwt_str
        end
        ctx.var["bcs_usertype"] = userinfo.usertype
        ctx.var["bcs_username"] = userinfo.username
    end
    return jwt_str
end

return _M
