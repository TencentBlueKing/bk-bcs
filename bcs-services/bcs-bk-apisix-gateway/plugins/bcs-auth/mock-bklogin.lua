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

local _M = {}

local valid_token = "2b0827e0-d932-4342-aa0c-cbaa409495a3"

-- get_username_for_ticket used for LoginTicketAuthentication
function _M.get_username_for_ticket(credential, bk_login_host)
    if credential.user_token == valid_token then
        return "bcs_user"
    end
    return nil
end

-- get_username_for_token used for LoginTokenAuthentication
function _M.get_username_for_token(credential, bk_login_host)
    if credential.user_token == valid_token then
        return "bcs_user"
    end
    return nil
end

return _M
