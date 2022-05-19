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
local upstream   = require("apisix.upstream")
local bcs_upstreams_util = require("apisix.plugins.bcs-common.upstreams")
local stringx = require('pl.stringx')
local timers = require("apisix.timers")
local http = require("resty.http")
local ipmatcher  = require("resty.ipmatcher")
local sub_str     = string.sub
local str_find    = core.string.find

local ngx = ngx
local ngx_shared = ngx.shared
local ngx_time = ngx.time
local ngx_update_time = ngx.update_time
local ngx_re = ngx.re
local table_insert = table.insert

local plugin_name = "bcs-dynamic-route"
local bcsapi_prefix = "/bcsapi/v4"
local clustermanager_credential_path = "/clustermanager/v1/clustercredential"
local clustermanager_tunnel_path = "/clustermanager/clusters/"
local last_sync_time
local last_sync_status
local credential_global_cache = ngx_shared[plugin_name]
local credential_worker_cache = core.lrucache.new({
    ttl = 30,
    count = 5000,
    serial_creating = true,
    invalid_stale = true,
})
local attr = {}

local schema = {
    type = "object",
    properties = {
        reg_extract_pattern = {type = "string", default = "/clusters/(BCS-K8S-[0-9]+)/(.*)", description = "regex pattern which will be used to extract clusterID ($1) and request uri ($2) from url"},
        clustermanager_upstream_name = {type = "string", default = "clustermanager-http", description = "clustermanager upstream name"},
        grayscale_clusterid_pattern = {type = "string", default = "BCS-K8S-2[0-9]+", description = "regex pattern for grayscale environment"},
        grayscale_clustermanager_address = {type = "string", description = "grayscale clustermanager url (For example, your.url:port)"},
        grayscale_gateway_token = {type = "string", description = "gateway token for access clustermanager via grayscale_clustermanager_address. If not specified, plugin attr config's gateway token will be used"},
        grayscale_clustermanager_upstream_name = {type = "string", description = "grayscale clustermanager upstream name (will use upstream when upstream is specified)"},
        timeout = {
            type = "object",
            properties = {
                send = {type = "integer", default = 60, description = "send timeout for proxying to cluster apiserver"},
                read = {type = "integer", default = 3600, description = "read timeout for proxying to cluster apiserver"},
                connect = {type = "integer", default = 60, description = "connect timeout for proxying to cluster apiserver"},
            },
            default = {send = 60, read = 3600, connect = 60}
        },
    }
}

local attr_schema = {
    type = "object",
    properties = {
        gateway_token = {type = "string", description = "gateway token for access clustermanager via apisix"},
        sync_cluster_credential_interval = {type = "integer", default = 10, description = "time interval for syncing cluster credential (s)"},
        gateway_insecure_port = {type = "integer", default = 8000, description = "apisix gateway insecure port"},
        cm_timeout = {description = "timeout seconds for request to clustermanager module", type = "number", minimum = 1, maxnum = 10, default = 10},
        cm_keepalive = {description = "keepalive seconds for request to clustermanager module", type = "number", minimum = 1, maxnum = 60, default = 60},
    }
}

local _M = {
    version = 0.1,
    priority = 0,
    name = plugin_name,
    schema = schema,
}

local function parse_domain_for_node(node)
    local host = node.host
    if not ipmatcher.parse_ipv4(host)
       and not ipmatcher.parse_ipv6(host)
    then
        node.domain = host

        local ip, err = core.resolver.parse_domain(host)
        if ip then
            node.host = ip
        end

        if err then
            core.log.error("dns resolver domain: ", host, " error: ", err)
        end
    end
end

local function set_upstream(upstream_info, ctx)
    local nodes = upstream_info.nodes
    local new_nodes = {}
    if core.table.isarray(nodes) then
        for _, node in ipairs(nodes) do
            parse_domain_for_node(node)
            table_insert(new_nodes, node)
        end
    else
        for addr, weight in pairs(nodes) do
            local node = {}
            local port, host
            host, port = core.utils.parse_addr(addr)
            node.host = host
            parse_domain_for_node(node)
            node.port = port
            node.weight = weight
            table_insert(new_nodes, node)
        end
    end
    upstream_info["nodes"] = new_nodes
    upstream_info["timeout"] = {
        send = upstream_info.timeout and upstream_info.timeout.send or 15,
        read = upstream_info.timeout and upstream_info.timeout.read or 15,
        connect = upstream_info.timeout and upstream_info.timeout.connect or 15
    }

    core.log.info("upstream_info: ", core.json.delay_encode(upstream_info, true))

    local ok, err = upstream.check_schema(upstream_info)
    if not ok then
        core.log.error("failed to validate generated upstream: ", err)
        return 500, err
    end

    local matched_route = ctx.matched_route
    ctx.matched_upstream = upstream_info
    upstream.set_by_route(matched_route, ctx)
end

-- time check
local function periodly_sync_cluster_credentials_in_master()
    if last_sync_status then
        ngx_update_time()
        local now_time = ngx_time()
        if not last_sync_time then
            last_sync_time = 0
        end
        if now_time - last_sync_time < attr.sync_cluster_credential_interval then
            return
        end
        last_sync_time = now_time
    elseif last_sync_status == false then
        core.log.warn("Last syncing cluster credential failed, retry")
    end
    
    -- start sync
    local httpCli = http.new()
    httpCli:set_timeout(attr.cm_timeout * 1000)
    local params = {
        method = "GET",
        query = "connectMode=direct",
        ssl_verify = false,
        keepalive = attr.cm_timeout * 1000,
        headers = {
          ["Content-Type"] = "application/json",
          ["Accept"] = "application/json",
          ["Authorization"] = "Bearer " .. attr.gateway_token
        }
    }
    local res, err = httpCli:request_uri("http://127.0.0.1:" .. attr.gateway_insecure_port .. bcsapi_prefix .. clustermanager_credential_path, params)
    if not res then
        core.log.error("request clustermanager error: ", err)
        last_sync_status = false
        return nil
    end
    if not res.body or res.status ~= 200 then
        core.log.error("request clustermanager status: ", res.status)
        last_sync_status = false
        return nil
    end

    local data, err = core.json.decode(res.body)
    if not data then
        core.log.error("request clustermanager decode body error: ", err)
        last_sync_status = false
        return nil
    end

    if data["code"] ~= 0 then
        core.log.error("request clustermanager return failed: ", data["message"])
        last_sync_status = false
        return nil
    end

    for _, cluster_credential in ipairs(data["data"]) do
        local cluster_info_cache = credential_global_cache:get(cluster_credential["clusterID"])
        local cluster_info = {}
        cluster_info["user_token"] = cluster_credential["userToken"]
        local upstream = {
            type = "roundrobin",
            scheme = "https",
        }
        if cluster_credential["clientCert"] and cluster_credential["clientKey"] then
            upstream["tls"] = {
                client_cert = cluster_credential["clientCert"], 
                client_key = cluster_credential["clientKey"],
            }
        end
        local upstream_nodes = {}
        local addresses = stringx.split(cluster_credential["serverAddress"], ",")
        for i, address in ipairs(addresses) do
            local splited = stringx.split(address, "://")
            local scheme = "https"
            if #splited ~= 2 then
                upstream_nodes[i] = {
                    host = address,
                    weight = 100,
                }
            else
                scheme = splited[1]
                upstream_nodes[i] = {
                    host = splited[2],
                    weight = 100,
                }
            end
            -- port is required, 443 as default port
            local host, port = core.utils.parse_addr(upstream_nodes[i].host)
            upstream_nodes[i].host = host
            if not port then
                if scheme == "http" then
                    core.log.warn("apiserver port auto-derived as 80 with scheme http")
                    upstream_nodes[i].port = 80
                else
                    core.log.warn("apiserver port auto-derived as 443 with scheme "..scheme)
                    upstream_nodes[i].port = 443
                end
            else
                upstream_nodes[i].port = port
            end
        end
        upstream["nodes"] = upstream_nodes
        cluster_info["upstream"] = upstream
        if cluster_info_cache then
            core.log.debug("cached credential: ", core.json.delay_encode(cluster_info_cache["upstream"]))
        end
        core.log.debug("new credential: ", core.json.delay_encode(cluster_info["upstream"]))
        local cluster_info_str = core.json.encode(cluster_info)
        if not cluster_info_cache or cluster_info_cache ~= cluster_info_str then
            local succ, err = credential_global_cache:set(cluster_credential["clusterID"], cluster_info_str)
            if not succ then
                core.log.error("insert cluster info into shared dict failed: ", err, "ClusterID: ", cluster_credential["clusterID"])
                goto continue
            end
            core.log.info("Sync cluster: ", cluster_credential["clusterID"])
        else
            core.log.info("Cluster (" .. cluster_credential["clusterID"] .. ") credential does not change")
        end
        ::continue::
    end
    last_sync_status = true
end

-- local cluster info from shared memory
local function load_cluster_info(clusterID)
    local cluster_info_str, err = credential_global_cache:get(clusterID)
    if not cluster_info_str then
        return nil, err
    end
    local cluster_info, err = core.json.decode(cluster_info_str)
    if not cluster_info then
        core.log.error("decode cluster info of " .. clusterID .. " failed, error: ", err)
        core.log.error("raw message: ", cluster_info_str)
        return nil, err
    end
    return cluster_info
end

-- proxy to clustermanager cluster websocket tunnel
local function traffic_to_clustermanager(conf, ctx, clusterID, upstream_uri)
    if conf.grayscale_clusterid_pattern and conf.grayscale_clusterid_pattern ~= "" then
        local captures = ngx_re.match(clusterID, conf.grayscale_clusterid_pattern)
        if captures then
            -- clustermanager upstream
            if conf.grayscale_clustermanager_upstream_name and conf.grayscale_clustermanager_upstream_name ~= "" then
                ctx.var.upstream_uri = clustermanager_tunnel_path .. clusterID .. "/" .. upstream_uri
                local upstream = bcs_upstreams_util.get_upstream_by_name(conf.grayscale_clustermanager_upstream_name)
                return set_upstream(upstream, ctx)
                -- ctx.upstream_id = conf.grayscale_clustermanager_upstream_name
                -- return
            end
            
            -- clustermanager url
            local host, port = core.utils.parse_addr(conf.grayscale_clustermanager_address)
            if not port then
                port = 443
            end
            local upstream = {
                type = "roundrobin",
                scheme = "https",
                pass_host = "rewrite",
                upstream_host = host,
                nodes = {
                    {
                        host = host,
                        port = port,
                        weight = 100,
                    },
                },
                timeout = conf.timeout
            }
            local token = ""
            if conf.grayscale_gateway_token and conf.grayscale_gateway_token ~= "" then
                token = conf.grayscale_gateway_token
            else
                token = attr.gateway_token
            end
            core.request.set_header(ctx, "Authorization", "Bearer " .. token)
            ctx.var.upstream_uri = bcsapi_prefix .. clustermanager_tunnel_path .. clusterID .. "/" .. upstream_uri
            return set_upstream(upstream, ctx)
        end
    end
    ctx.var.upstream_uri = clustermanager_tunnel_path .. clusterID .. "/" .. upstream_uri
    local upstream = bcs_upstreams_util.get_upstream_by_name(conf.clustermanager_upstream_name)
    return set_upstream(upstream, ctx)
    -- ctx.upstream_id = conf.clustermanager_upstream_name
end

-- proxy to apiserver directly
local function traffic_to_cluster_apiserver(conf, ctx, cluster_credential, upstream_uri)
    ctx.var.upstream_uri = "/" .. upstream_uri
    if cluster_credential["user_token"] then
        core.request.set_header(ctx, "Authorization", "Bearer " .. cluster_credential["user_token"])
    end
    cluster_credential["upstream"]["timeout"] = conf.timeout
    return set_upstream(cluster_credential["upstream"], ctx)
end

function _M.check_schema(conf)
    return core.schema.check(schema, conf)
end

function _M.access(conf, ctx)
    local captures, err = ngx_re.match(ngx.var.uri, conf.reg_extract_pattern, "jo")
    if not captures then
        core.log.error("extract clusterid and request path from url failed: ", err)
        return 404, {message = "Cluster not found"}
    end
    local clusterID = ""
    if #captures < 2 then
        core.log.error("regex captures does not contain clusterID or request path, captures:  ", core.json.encode(captures, true))
        return 404, {message = "Resource not found"}
    end
    clusterID = captures[1]
    local upstream_uri = captures[2]

    -- append url query parameter to upsrteam_uri
    local index = str_find(upstream_uri, "?")
    if index then
        upstream_uri = core.utils.uri_safe_encode(sub_str(upstream_uri, 1, index-1)) ..
                       sub_str(upstream_uri, index)
    else
        upstream_uri = core.utils.uri_safe_encode(upstream_uri)
    end

    if ctx.var.is_args == "?" then
        if index then
            upstream_uri = upstream_uri .. "&" .. (ctx.var.args or "")
        else
            upstream_uri = upstream_uri .. "?" .. (ctx.var.args or "")
        end
    end

    ctx.upstream_scheme = "https"
    ctx.var.upstream_scheme = "https"

    local cluster_credential = credential_worker_cache(clusterID, nil, load_cluster_info, clusterID)
    if cluster_credential then
        core.log.debug("ClusterID: ", clusterID, " matches cluster upstream: ", core.json.delay_encode(cluster_credential["upstream"], true))
    else
        core.log.debug("ClusterID: ", clusterID, " does not match any cluster upstream")
    end
    if not cluster_credential then
        traffic_to_clustermanager(conf, ctx, clusterID, upstream_uri)
        return
    end
    traffic_to_cluster_apiserver(conf, ctx, cluster_credential, upstream_uri)
end

function _M.init()
    local local_conf = core.config.local_conf()
    attr = core.table.try_read_attr(local_conf, "plugin_attr", plugin_name)
    local ok, err = core.schema.check(attr_schema, attr)
    timers.register_timer("plugin#"..plugin_name, periodly_sync_cluster_credentials_in_master, true)
    if not ok then
        core.log.error("failed to check the plugin_attr[", plugin_name, "]", ": ", err)
        return
    end
end

function _M.destroy()
    timers.unregister_timer("plugin#"..plugin_name, true)
end

return _M
