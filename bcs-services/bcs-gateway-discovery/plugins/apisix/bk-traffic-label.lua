--
-- TencentBlueKing is pleased to support the open source community by making
-- 蓝鲸智云 - API 网关(BlueKing - APIGateway) available.
-- Copyright (C) 2017 THL A29 Limited, a Tencent company. All rights reserved.
-- Licensed under the MIT License (the "License"); you may not use this file except
-- in compliance with the License. You may obtain a copy of the License at
--
--     http://opensource.org/licenses/MIT
--
-- Unless required by applicable law or agreed to in writing, software distributed under
-- the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
-- either express or implied. See the License for the specific language governing permissions and
-- limitations under the License.
--
-- We undertake not to change the open source license (MIT license) applicable
-- to the current version of the project delivered to anyone in the future.
--

-- this plugin is impls based on the doc of api7 traffic-label
-- link: https://docs.api7.ai/hub/traffic-label/

local core       = require("apisix.core")
local expr       = require("resty.expr.v1")
local pairs      = pairs
local ipairs     = ipairs

local match_schema = {
    type = "array",
}

local actions_schema = {
    type = "array",
    items = {
        type = "object",
        properties = {
            set_headers = {
                type = "object",
                additionalProperties = {
                    type = "string"
                }
            },
            weight = {
                description = "percentage of all matched which would do the actions",
                type = "integer",
                default = 1,
                minimum = 0
            }
        }
    },
    minItems = 1,
    maxItems = 20
}

local schema = {
    type = "object",
    properties = {
        rules = {
            type = "array",
            items = {
                type = "object",
                properties = {
                    match = match_schema,
                    actions = actions_schema
                },
            }
        }
    },
    required = {"rules"},
}

local plugin_name = "bk-traffic-label"

local _M = {
    version = 0.1,
    priority = 17460,
    name = plugin_name,
    schema = schema
}

function _M.check_schema(conf)
    -- Validate the configuration schema
    local ok, err = core.schema.check(schema, conf)
    if not ok then
        return false, err
    end

    if conf.rules then
        for _, rule in ipairs(conf.rules) do
            if rule.match then
                -- Validate the match expression
                local _, err2 = expr.new(rule.match)
                if err2 then
                    core.log.error("failed to validate the 'match' expression: ", err2)
                    return false, "failed to validate the 'match' expression: " .. err2
                end
            end

            -- Calculate total weight of all actions and preprocess actions to set default weight
            local total_weight = 0
            for _, action in ipairs(rule.actions) do
                if action.weight == nil or action.weight < 0 then
                    action.weight = 1
                end
                total_weight = total_weight + action.weight
            end
            rule.total_weight = total_weight
        end
    end

    return true
end

local function apply_actions(rule, ctx)
    -- Generate a random number between 1 and total_weight, [1, total_weight]
    local random_weight = math.random(1, rule.total_weight)
    local current_weight = 0

    for _, action in ipairs(rule.actions) do
        current_weight = current_weight + action.weight
        -- Apply the action if the random number falls within the current weight range
        if random_weight <= current_weight then
            if action.set_headers then
                -- Set the specified headers
                for k, v in pairs(action.set_headers) do
                    core.request.set_header(ctx, k, v)
                end
            end
            break
        end
    end
end

function _M.access(conf, ctx)
    if not conf or not conf.rules then
        return
    end

    for _, rule in ipairs(conf.rules) do
        if rule.match then
            -- Evaluate the match expression
            local ex, _ = expr.new(rule.match)
            local match_passed = ex:eval(ctx.var)
            if match_passed then
                -- Apply the actions if the match condition is met
                apply_actions(rule, ctx)
            end
        end
    end

    return
end

return _M