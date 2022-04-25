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

    return {user_token = bk_ticket, username = bk_uid}

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
    return {user_token = bk_token}
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
        return {user_token = nil}
    end

    local m, err = ngx.re.match(auth_header, "Bearer\\s(.+)", "jo")
    if err then
        -- error authorization
        return {user_token = nil}
    end

    return {user_token = m[1], token_type = TOKEN_TYPE_BCS}
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

    local key = conf.bkapigw_jwt_verify_key
    local jwt_obj = resty_jwt:load_jwt(jwt_str)
    if not jwt_obj then
        core.log.error("load jwt from apigw jwt token failed, jwt token:"..jwt_str)
        core.response.exit(401, "Bad Bkapi JWT token")
    end
    if not key then
        core.log.error("no verify key for apigw jwt token, jwt token:"..jwt_str)
        core.response.exit(500, "BCS Auth Plugin Error")
    end
    local decode_res = ngx_decode_base64(key)
    if decode_res ~= nil then
        key = decode_res
    end
    local verified = resty_jwt:verify_jwt_obj(key, jwt_obj)
    if not verified then
        core.log.error("verify apigw jwt token failed, jwt token:"..jwt_str)
        core.response.exit(401, "Bad Bkapi JWT token")
    end
    if not jwt_obj.payload.user then
        core.log.error("app is not yet allowed")
        core.log.error("apigw jwt:"..jwt_str)
        core.log.error("app_code:"..jwt_obj.payload.app.app_code)
        core.response.exit(401, "User identify required")
    end

    return {user_token = jwt_obj.payload.app.app_code..","..jwt_obj.payload.user.username, token_type = TOKEN_TYPE_APIGW}
end

function APIGWAuthentication.get_username(credential, useless)
    local appcode_and_username = credential.user_token
    if not appcode_and_username then
        core.log.error("No token for APIGWAuthentication.get_username func, credential is: "..core.json.encode(credential, true))
        core.response(500, "BCS Auth Plugin Error")
    end
    local splited = stringx.split(appcode_and_username, ",")
    if #splited ~= 2 then
        core.log.error("appcode username token length is incorrect, raw string: "..appcode_and_username)
        core.response(500, "BCS Auth Plugin Error")
    end
    return splited[2]
end


function APIGWAuthentication:get_jwt(credential, conf)
    if not credential or not credential.user_token then
        return nil
    end
    if credential.token_type == TOKEN_TYPE_BCS then
        return TokenAuthentication:get_jwt(credential, conf)
    end
    return jwt:get_jwt_from_redis(credential, conf, "bcs_auth:apigw:", true, APIGWAuthentication.get_username)
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

    return self.backend:get_jwt(credential, conf)
end


return _M
