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

_M.TENANT_MODE_SINGLE = "single"
_M.TENANT_MODE_MULTI = "multi"

-- get_tenant_mode 返回租户模式 single / multi，tenant 为 off 或空字符串时返回 nil 表示关闭租户
function _M.get_tenant_mode(conf)
    local mode = conf.tenant
    if mode == _M.TENANT_MODE_SINGLE or mode == _M.TENANT_MODE_MULTI then
        return mode
    end
    return nil
end

-- is_multi_tenant_enabled 判断是否开启多租户模式
function _M.is_multi_tenant_enabled(conf)
    return _M.get_tenant_mode(conf) == _M.TENANT_MODE_MULTI
end

-- is_single_tenant_enabled 判断是否开启单租户模式
function _M.is_single_tenant_enabled(conf)
    return _M.get_tenant_mode(conf) == _M.TENANT_MODE_SINGLE
end

-- is_tenant_enabled 判断是否启用租户版登录校验（单租户或多租户）
function _M.is_tenant_enabled(conf)
    return _M.get_tenant_mode(conf) ~= nil
end

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
    local res, err = httpc:request_uri(bk_login_host .. "/accounts/is_login/", {
        method = "GET",
        query = {bk_token = credential.user_token},
        headers = {
            ["Content-Type"] = "application/json",
            ["X-Bkapi-Authorization"] = authorization,
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

-- get_username_and_tenant_for_token used for LoginTokenAuthentication
function _M.get_username_and_tenant_for_token(credential, conf)
    if not conf.bk_login_host_tenant or conf.bk_login_host_tenant == "" then
        core.log.error("bk_login_host_tenant is empty when tenant mode is enabled")
        return nil
    end
    --core.log.warn("conf.bk_login_host_tenant: ", conf.bk_login_host_tenant)
    local httpc = http.new()
    local authorization = string.format('{"bk_app_code": "%s", "bk_app_secret": "%s"}',
        conf.bk_app_code or "", conf.bk_app_secret or "")
    --core.log.warn("authorization: ", authorization)
    local res, err = httpc:request_uri(conf.bk_login_host_tenant .. "/bk-tokens/verify/", {
        method = "GET",
        query = {bk_token = credential.user_token},
        headers = {
            ["Content-Type"] = "application/json",
            ["X-Bkapi-Authorization"] = authorization,
            ["X-Bk-Tenant-Id"] = "default",
        },
    })

    if not res then
        core.log.error("request login error: ", err)
        return nil
    end

    if not res.body or res.status ~= 200 then
        core.log.error("request login status: ", res.status)
        core.log.error("request login body: ", res.body)
        return nil
    end

    local data, err = core.json.decode(res.body)
    if not data then
        core.log.error("request login decode body error: ", err)
        return nil
    end
    --core.log.warn("data: ", core.json.encode(data))

    if data['error'] ~= nil then
        core.log.error("request login error: ", data["error"])
        return nil
    end

    return data
end

-- get_username_for_token_esb used for LoginTokenAuthentication
function _M.get_username_for_token_esb(credential, conf)
    local httpc = http.new()
    local authorization = string.format('{"bk_app_code": "%s", "bk_app_secret": "%s"}', conf.bk_app_code, conf.bk_app_secret)
    local res, err = httpc:request_uri(conf.bk_login_host_esb .. "/api/c/compapi/v2/bk_login/is_login/", {
        method = "GET",
        query = {bk_token = credential.user_token},
        headers = {
            ["Content-Type"] = "application/json",
            ["X-Bkapi-Authorization"] = authorization,
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

    return data
end


return _M