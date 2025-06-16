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
local OPERATOR_HEADER = "X-BCS-Operator" -- 运维人员头部 header

local bcs_token_user_map_cache = core.lrucache.new(
    {
        ttl = 300,
        count = 1000,
        serial_creating = true,
        invalid_stale = true,
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
    return jwt:get_jwt_from_redis(credential, conf, nil,  "bcs_auth:session_id:", true, bklogin.get_username_for_ticket)
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
    if conf.bk_login_host_esb then
        local data = bklogin.get_username_for_token_esb(credential, conf)
        local username
        if data["result"] then
            username = data["data"]["bk_username"]
        end
        return {
            username = username,
            usertype = "user",
            bk_login_code = data["code"],
            bk_login_message= data["message"],
        }
    end

    if conf.bk_login_host_tenant then
        --core.log.warn("conf.bk_login_host_tenant: ", conf.bk_login_host_tenant)
        local data = bklogin.get_username_and_tenant_for_token(credential, conf)
        local username, tenant_id
        if data ~= nil then
            if data["data"] ~= nil then
                username, tenant_id = data["data"]["bk_username"], data["data"]["tenant_id"]
            end
        end
        return {
            username = username,
            usertype = "user",
            tenant_id = tenant_id,
        }
    end

    return {
        username = bklogin.get_username_for_token(credential, conf.bk_login_host),
        usertype = "user",
    }
end

function LoginTokenAuthentication:get_jwt(credential, conf, ctx)
    if conf.bk_login_host_esb then
        return jwt:get_jwt_from_redis(credential, conf, ctx,  "bcs_auth:session_id:", true, bklogin.get_username_for_token_esb)
    end
    if conf.bk_login_host_tenant then
        --core.log.warn("conf.bk_login_host_tenant: ", conf.bk_login_host_tenant)
        return jwt:get_jwt_from_redis(credential, conf, ctx,  "bcs_auth:session_id:", true, bklogin.get_username_and_tenant_for_token)
    end

    return jwt:get_jwt_from_redis(credential, conf, ctx, "bcs_auth:session_id:", true, bklogin.get_username_for_token)
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
    if err or not m then
        -- error authorization
        return {
            user_token = nil,
        }
    end

    local user_token = m[1]
    -- 如果有操作人头部, 使用操作人+token查询
    local operator = core.request.header(ctx, OPERATOR_HEADER)
    if operator then
        user_token = "op-" .. operator .. ":" .. user_token
        core.log.warn("req use operator instead: ", operator)
    end

    return {
        user_token = user_token,
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

function TokenAuthentication:get_jwt(credential, conf, ctx)
    return jwt:get_jwt_from_redis(credential, conf, ctx, "bcs_auth:token:", false)
end
------------ TokenAuthentication end ------------

------------ APIGWAuthentication start ------------
local APIGWAuthentication = {}

-- 获取 cookie 中的用户凭证信息
function APIGWAuthentication:fetch_credential(conf, ctx)
    local credential = {
        token_type = TOKEN_TYPE_APIGW,
        user_token = {},
    }
    local bcs_credential = TokenAuthentication:fetch_credential(conf, ctx)
    credential.bcs_token = bcs_credential.user_token
    local jwt_str = core.request.header(ctx, "X-Bkapi-JWT")
    if not jwt_str then
        return bcs_credential
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
        return bcs_credential
    end
    local redis_key = jwt_obj.payload.app.app_code
    credential["user_token"]["bk_app_code"] = jwt_obj.payload.app.app_code
    if jwt_obj.payload.user and jwt_obj.payload.user.verified then
        credential["user_token"]["username"] = jwt_obj.payload.user.username
        redis_key = redis_key .. "," .. jwt_obj.payload.user.username
    end
    if credential.bcs_token then
        local bcs_payload = bcs_token_user_map_cache(credential.bcs_token, nil, function ()
            return nil, nil
        end)
        if not bcs_payload then
            local bcs_jwt_token = TokenAuthentication:get_jwt({user_token=credential.bcs_token}, conf, ctx)
            bcs_payload = get_username_by_TokenAuthentication_jwt(bcs_jwt_token)
        end

        if bcs_payload then
            credential.user_token.sub_type = bcs_payload.sub_type
            credential.user_token.client_id = bcs_payload.client_id
            credential.user_token.client_secret = bcs_payload.client_secret

            redis_key = redis_key .. "," .. credential.bcs_token
        end
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
    local retV = credential.user_token
    if retV.sub_type then
        retV.usertype = retV.sub_type
    elseif retV.username then
        retV.usertype = "user"
    else
        retV.usertype = "bk_app"
    end
    return retV
end

function APIGWAuthentication:get_jwt(credential, conf)
    if not credential or not credential.user_token then
        return nil
    end
    if credential.token_type == TOKEN_TYPE_BCS then
        return TokenAuthentication:get_jwt(credential, conf)
    end
    return jwt:get_jwt_from_redis(credential, conf, nil, "bcs_auth:apigw:", true, APIGWAuthentication.get_userinfo)
end
------------ APIGWAuthentication end ------------

local _M = {}

function _M:new(use_login, run_env)
    local o = {}
    setmetatable(o, self)

    self.__index = self

    if not use_login then
        o.backend = APIGWAuthentication
        return o
    end

    if run_env == RUN_ON_CE then
        o.backend = LoginTokenAuthentication
    else
        o.backend = LoginTicketAuthentication
    end

    return o
end

function _M:authenticate(conf, ctx)
    local credential = self.backend:fetch_credential(conf, ctx)

    if not credential or not credential.user_token then
        return nil
    end
    local jwt_str = self.backend:get_jwt(credential, conf, ctx)

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
        ctx.var["bk_app_code"] = userinfo.bk_app_code
        ctx.var["bcs_client_id"] = userinfo.client_id
        ctx.var["bk_login_code"] = userinfo.bk_login_code
        ctx.var["bk_login_message"] = userinfo.bk_login_message
    end
    return jwt_str
end

return _M
