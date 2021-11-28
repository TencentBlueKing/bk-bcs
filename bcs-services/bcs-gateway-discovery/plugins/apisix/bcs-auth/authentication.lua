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
local http = require("resty.http")
local jwt = require("apisix.plugins.bcs-auth.jwt")

local RUN_ON_CE = "ce"


------------ LoginTicketAuthentication start ------------
local function get_username_for_ticket(credential, bk_login)
    local httpc = http.new()
    local res, err = httpc:request_uri(bk_login .. "/user/is_login/", {
        method = "GET",
        query = {bk_ticket = credential.user_token},
        headers = {
            ["Content-Type"] = "application/json",
        },
    })

    if not res then
        core.log.error("request login error: ", err)
        return nil
    end

    if not res.body or res.status ~= 200 then
        core.log.error("request login status: ", res.status)
        return nil
    end

    local data, err = core.json.decode(res.body)
    if not data then
        core.log.error("request login decode body error: ", err)
        return nil, err
    end

    if data["ret"] == 0 then
        return credential.username
    end

    return nil
end


local LoginTicketAuthentication = {}


-- 获取 cookie 中的用户凭证信息
function LoginTicketAuthentication:fetch_credential()
    local cookie, err = ck:new()
    if not cookie then
        return nil, err
    end

    local bk_ticket, err = cookie:get("bk_ticket")
    if not bk_ticket then
        if err and err:sub(1, #"no cookie") ~= "no cookie" then
            core.log.error("failed to fetch bk_ticket: ", err)
        end
    end

    local bk_uid, err = cookie:get("bk_uid")
    if not bk_uid then
        if err and err:sub(1, #"no cookie") ~= "no cookie" then
            core.log.error("failed to fetch bk_uid: ", err)
        end
    end

    return {user_token = bk_ticket, username = bk_uid}

end


function LoginTicketAuthentication:get_jwt(credential, conf)
    return jwt:get_jwt_from_redis(credential, conf, "session_id:", true, get_username_for_ticket)
end
------------ LoginTicketAuthentication end ------------


------------ LoginTokenAuthentication start ------------
local function get_username_for_token(credential, bk_login)
    -- TODO: 实现
    return nil
end


local LoginTokenAuthentication = {}


-- 获取 cookie 中的用户凭证信息
function LoginTokenAuthentication:fetch_credential()
    local cookie, err = ck:new()
    if not cookie then
        return nil, err
    end

    local bk_token, err = cookie:get("bk_token")
    if not bk_token then
        if err and err:sub(1, #"no cookie") ~= "no cookie" then
            core.log.error("failed to fetch bk_token: ", err)
        end
    end

    return {user_token = bk_token}
end


function LoginTokenAuthentication:get_jwt(credential, conf)
    return jwt:get_jwt_from_redis(credential, conf, "session_id:", true, get_username_for_token)
end
------------ LoginTokenAuthentication end ------------


------------ TokenAuthentication start ------------
local TokenAuthentication = {}


function TokenAuthentication:fetch_credential()
    local auth_header = core.request.header(ctx, "Authorization")
    if not auth_header then
        return {user_token = nil}
    end

    local m, err = ngx.re.match(auth_header, "Bearer\\s(.+)", "jo")
    if err then
        -- error authorization
        return {user_token = nil}
    end

    return {user_token = m[1]}
end


function TokenAuthentication:get_jwt(credential, conf)
    return jwt:get_jwt_from_redis(credential, conf, "token:", false)
end
------------ TokenAuthentication end ------------


local _M = {}


function _M:new(use_login, run_env)
    local o = {}
    setmetatable(o, self)

    self.__index = self

    if not use_login then
        self.backend = TokenAuthentication
        return o
    end

    if run_env == RUN_ON_CE then
        self.backend = LoginTokenAuthentication
    else
        self.backend = LoginTicketAuthentication
    end

    return o
end


function _M:authenticate(conf)
    local credential = self.backend:fetch_credential()

    if not credential.user_token then
        return nil
    end

    return self.backend:get_jwt(credential, conf)
end


return _M
