--* Tencent is pleased to support the open source community by making Blueking Container Service available.
-- Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
-- Licensed under the MIT License (the "License"); you may not use this file except
-- in compliance with the License. You may obtain a copy of the License at
-- http://opensource.org/licenses/MIT
-- Unless required by applicable law or agreed to in writing, software distributed under
-- the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
-- either express or implied. See the License for the specific language governing permissions and
-- limitations under the License.
--

local typedefs = require "kong.db.schema.typedefs"

return {
  name = "bkbcs-auth",
  fields = {
    { protocols = typedefs.protocols_http },
    { config = {
        type = "record",
        fields = {
          -- NOTE: any field added here must be also included in the handler's get_queue_id method
          { bkbcs_auth_endpoints = typedefs.url({ required = true }) },
          -- module for bkbcs module
          { module = {type = "string", required = true}, },
          -- token for post request to bkbcs-user-mananger
          { token = {type = "string", required = true}, },
          -- network timeout
          { timeout = { type = "number", default = 3000 }, },
          -- keepalive for connection reuse
          { keepalive = { type = "number", default = 60000 }, },
          -- retry count when failing
          { retry_count = { type = "integer", default = 1 }, },
        },
      },
    },
  },
}