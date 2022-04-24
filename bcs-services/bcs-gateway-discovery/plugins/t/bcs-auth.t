#
# Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
# Edition) available.
# Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://opensource.org/licenses/MIT
#
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.
#
use t::APISIX 'no_plan';

repeat_each(1);
no_long_string();
no_root_location();
no_shuffle();
run_tests;

__DATA__

=== TEST 1: sanity
--- config
    location /t {
        content_by_lua_block {
            local plugin = require("apisix.plugins.bcs-auth")
            local ok, err = plugin.check_schema({bk_login_host = "http://login.bk.com", private_key = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLR...", 
            redis_host = "127.0.0.1", redis_password = "123@lua"})
            if not ok then
                ngx.say(err)
            end

            ngx.say("done")
        }
    }
--- request
GET /t
--- response_body
done
--- no_error_log
[error]

=== TEST 2: wrong type of string
--- config
    location /t {
        content_by_lua_block {
            local plugin = require("apisix.plugins.bcs-auth")
            local ok, err = plugin.check_schema({bk_login_host = "http://login.bk.com", private_key = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLR...", 
            redis_host = "127.0.0.1", redis_password = 123})
            if not ok then
                ngx.say(err)
            end

            ngx.say("done")
        }
    }
--- request
GET /t
--- response_body
property "redis_password" validation failed: wrong type: expected string, got number
done
--- no_error_log
[error]

=== TEST 3: enable bcs auth plugin using admin api
--- config
    location /t {
        content_by_lua_block {
            local t = require("lib.test_admin").test
            local code, body = t('/apisix/admin/routes/1',
                ngx.HTTP_PUT,
                [[{
                    "plugins": {
                        "bcs-auth": {
                            "bk_login_host": "http://login.bk.com",
                            "private_key": "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCT3dJQkFBSkJBTVg0UW9idVQvSzhWUzIyN3NYZEVKTThCSXJjNnVRU21HOHA0a2Z1S1lLKzJUNy9PaWRUCjRWQ1Y5TUpJdjdidG1BcmI2UFJsYURqdmZiVi9jV3FRQVRrQ0F3RUFBUUpCQUxLaVVXVnZwTFJqUEhrRG1IRHkKQ1FMU0pVY29FTXU3KzlCUyt0dnRDNGZ0RjhRSlMrejcrTXZiOGtyS0Mxb3dzTlFRR2hVR0ovanBoMG5JTXA4UgpBQUVDSVFEbEd5YTBPaWszbTAzcXhDNy96V0p3a0dmZ1pwNit0TStpQm93MTRYSnUrUUloQU4wMWJ5d3p5Q1QyCjJwd2NVTW02dlBvRlFTSGgzSkFIN1c3YTcxNk5RWFJCQWlCeUhpZ1ZOYk02STMyWUpzaFNXbmRpSWt2YmxzSVQKcy9TSWZFSnl4QzAvNFFJZ0R1M1hSZTFzdVlucmNSTzhKQkUxUmM1cStlVnJaRkVVcGlHaWZBZ2VmY0VDSVFEUgpvaU1FS3E3d1dIa1Y2WjRWMUllYnkrSS9UUmVhQmxaR2Q4d2NVTkgrVEE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQ==",
                            "redis_host": "127.0.0.1",
                            "redis_password": "Ps2+5f@10"
                        }
                    },
                    "upstream": {
                        "nodes": {
                            "127.0.0.1:1980": 1
                        },
                        "type": "roundrobin"
                    },
                    "uri": "/hello"
                }]]
                )

            if code >= 300 then
                ngx.status = code
            end
            ngx.say(body)
        }
    }
--- request
GET /t
--- response_body
passed
--- no_error_log
[error]

=== TEST 4: verify, missing authorization
--- request
GET /hello
--- error_code: 401
--- response_body
{"message":"bcs-auth plugin error: token is expired or is invalid"}
--- no_error_log
[error]

=== TEST 5: setup valid token in redis
--- config
    location /t {
        content_by_lua_block {
            local red = require("resty.redis").new()
            red:connect("127.0.0.1", "6379")
            red:auth("Ps2+5f@10")
            red:select(0)
            red:set("bcs_auth:token:ce1d6ce7996a4ad8a9b2e3e36410acea", "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UifQ...")

            ngx.say("done")
        }
    }
--- request
GET /t
--- response_body
done
--- no_error_log
[error]

=== TEST 6: verify, invalid token
--- request
GET /hello
--- more_headers
Authorization: Bearer YmFyOmJhcgo=
--- error_code: 401
--- response_body
{"message":"bcs-auth plugin error: token is expired or is invalid"}
--- no_error_log
[error]

=== TEST 7: verify valid token
--- request
GET /hello
--- more_headers
Authorization: Bearer ce1d6ce7996a4ad8a9b2e3e36410acea
--- error_code: 200
--- response_body
hello world
--- no_error_log
[error]

=== TEST 8: verify no login token (in cookie)
--- request
GET /hello
--- more_headers
User-Agent: Mozilla
--- error_code: 302
--- no_error_log
[error]

=== TEST 9: verify login token (in cookie)
--- init_by_lua_block
    apisix = require("apisix")
    apisix.http_init()

    package.loaded["apisix.plugins.bcs-auth.bklogin"]= require("apisix.plugins.bcs-auth.mock-bklogin")
--- request
GET /hello
--- more_headers
User-Agent: Mozilla
Cookie: bk_token=2b0827e0-d932-4342-aa0c-cbaa409495a3
--- error_code: 200
--- response_body
hello world
--- no_error_log
[error]


=== TEST 10: JWT verify use RS256 algorithm(private_key numbits = 512)
--- config
    location /t {
        content_by_lua_block {
            local core = require("apisix.core")
            local jwt = require("resty.jwt")
            local red = require("resty.redis").new()
            red:connect("127.0.0.1", "6379")
            red:auth("Ps2+5f@10")
            red:select(0)

            jwt_token = red:get("bcs_auth:session_id:2b0827e0-d932-4342-aa0c-cbaa409495a3")
            local jwt_obj = jwt:load_jwt(jwt_token)
            local auth_secret = "-----BEGIN PUBLIC KEY-----\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAMX4QobuT/K8VS227sXdEJM8BIrc6uQS\nmG8p4kfuKYK+2T7/OidT4VCV9MJIv7btmArb6PRlaDjvfbV/cWqQATkCAwEAAQ==\n-----END PUBLIC KEY-----"
            jwt_obj = jwt:verify_jwt_obj(auth_secret, jwt_obj)
            ngx.say(jwt_obj.payload.username .. " from " ..jwt_obj.payload.iss)
        }
    }
--- request
GET /t
--- response_body
bcs_user from bcs-auth-plugin
--- no_error_log
[error]


=== TEST 11: verify invalid login token (in cookie)
--- init_by_lua_block
    apisix = require("apisix")
    apisix.http_init()

    package.loaded["apisix.plugins.bcs-auth.bklogin"]= require("apisix.plugins.bcs-auth.mock-bklogin")
--- request
GET /hello
--- more_headers
User-Agent: Mozilla
Cookie: bk_token=7365a41c-86db-4f6f-977a-af5906cd5bc1
--- error_code: 302
--- no_error_log
[error]


=== TEST 12: teardown token from redis
--- config
    location /t {
        content_by_lua_block {
            local red = require("resty.redis").new()
            red:connect("127.0.0.1", "6379")
            red:auth("Ps2+5f@10")
            red:select(0)
            red:del("bcs_auth:session_id:2b0827e0-d932-4342-aa0c-cbaa409495a3")
            red:del("bcs_auth:token:ce1d6ce7996a4ad8a9b2e3e36410acea")

            ngx.say("done")
        }
    }
--- request
GET /t
--- response_body
done
--- no_error_log
[error]
