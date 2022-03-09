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
local http = require("resty.http")


local _M = {}

-- get_username_for_ticket used for LoginTicketAuthentication
function _M.get_username_for_ticket(credential, bk_login_host)
    local httpc = http.new()
    local res, err = httpc:request_uri(bk_login_host .. "/user/is_login/", {
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
        return nil
    end

    if data["ret"] == 0 then
        return credential.username
    end

    return nil
end

-- get_username_for_token used for LoginTokenAuthentication
function _M.get_username_for_token(credential, bk_login_host)
    local httpc = http.new()
    local res, err = httpc:request_uri(bk_login_host .. "/login/accounts/is_login/", {
        method = "GET",
        query = {bk_token = credential.user_token},
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
        return nil
    end

    if not data["result"] then
        core.log.error("request login error: ", data["message"])
        return nil
    end

    return data["data"]["username"]
end


return _M