#!/usr/bin/env bash
# BCS OpenAPI 统一生成脚本
#
# 功能:
#   - 从 service_config.yaml 读取各服务配置
#   - 支持 proto 类服务一键生成 (protoc-gen-swagger -> swagger2openapi 转换)
#   - 支持 swag 类服务一键生成 (swag init -> swagger2openapi 转换)
#   - 支持 manual 类服务同步 (直接复制 openapi.yaml)
#   - 生成后自动验证输出文件
#
# 用法:
#   ./openapi/generate.sh                          # 生成所有服务
#   ./openapi/generate.sh --service bcs-cluster-manager   # 生成指定服务
#   ./openapi/generate.sh --proto-only             # 仅生成 proto 类服务
#   ./openapi/generate.sh --swag-only              # 仅生成 swag 类服务
#   ./openapi/generate.sh --manual-only            # 仅同步 manual 类服务
#   ./openapi/generate.sh --list                   # 列出所有已配置的服务
#
# 依赖:
#   - protoc + protoc-gen-swagger (proto 类):
#       go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@latest
#   - swag (swag 类):
#       go install github.com/swaggo/swag/cmd/swag@latest
#   - swagger2openapi (proto/swag 类转换，需要 Node.js):
#       npm install -g swagger2openapi
#   - python3 + pyyaml
#   - 可选: python3 openapi/scripts/validate_openapi.py (自动验证)

set -euo pipefail

# ============================================================
# 路径初始化
# ============================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BCS_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
OPENAPI_DIR="${SCRIPT_DIR}"
CONFIG_FILE="${OPENAPI_DIR}/service_config.yaml"
SCRIPTS_DIR="${OPENAPI_DIR}/scripts"

# ============================================================
# 颜色输出
# ============================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC}   $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
error()   { echo -e "${RED}[ERR]${NC}  $*" >&2; }
step()    { echo -e "\n${CYAN}==>${NC} $*"; }

# ============================================================
# YAML 配置读取（使用 python3）
# ============================================================

# 读取 service_config.yaml 中某服务的指定字段值
get_config() {
    local service="$1"
    local field="$2"
    python3 - <<EOF
import yaml, sys
with open("${CONFIG_FILE}") as f:
    cfg = yaml.safe_load(f)
svc = cfg.get("services", {}).get("${service}", {})
val = svc.get("${field}", "")
if isinstance(val, list):
    print("\n".join(str(v) for v in val))
else:
    print(val)
EOF
}

# 列出所有服务名
list_services() {
    python3 - <<EOF
import yaml
with open("${CONFIG_FILE}") as f:
    cfg = yaml.safe_load(f)
for name in cfg.get("services", {}).keys():
    print(name)
EOF
}

# 按类型列出服务名
list_services_by_type() {
    local svc_type="$1"
    python3 - <<EOF
import yaml
with open("${CONFIG_FILE}") as f:
    cfg = yaml.safe_load(f)
for name, svc in cfg.get("services", {}).items():
    if svc.get("type") == "${svc_type}":
        print(name)
EOF
}

# ============================================================
# Proto 类服务生成
# 流程: proto -> protoc-gen-swagger -> output_dir/tmp/swagger.json -> swagger2openapi_convert.py -> output_dir/openapi.yaml
# ============================================================

generate_proto() {
    local service="$1"
    step "生成 proto 服务: ${service}"

    local source_dir module proto_file output_dir
    source_dir=$(get_config "${service}" "source_dir")
    module=$(get_config "${service}" "module")
    proto_file=$(get_config "${service}" "proto_file")
    output_dir=$(get_config "${service}" "output_dir")

    if [[ -z "${source_dir}" || -z "${proto_file}" || -z "${output_dir}" ]]; then
        error "${service}: 缺少 source_dir、proto_file 或 output_dir 配置"
        return 1
    fi

    local abs_source="${BCS_ROOT}/${source_dir}"
    if [[ ! -d "${abs_source}" ]]; then
        error "${service}: source_dir 不存在: ${abs_source}"
        return 1
    fi

    # 检查 protoc 和 protoc-gen-swagger 是否安装
    if ! command -v protoc &>/dev/null; then
        error "protoc 未安装，请参考 https://grpc.io/docs/protoc-installation/"
        return 1
    fi
    if ! command -v protoc-gen-swagger &>/dev/null; then
        error "protoc-gen-swagger 未安装，请运行: go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@latest"
        return 1
    fi

    # 构造 protoc -I 参数
    local proto_includes_raw
    proto_includes_raw=$(get_config "${service}" "proto_includes")
    local proto_i_args=""
    while IFS= read -r inc; do
        [[ -z "${inc}" ]] && continue
        if [[ "${inc}" == /* ]]; then
            proto_i_args="${proto_i_args} -I${inc}"
        else
            proto_i_args="${proto_i_args} -I${abs_source}/${inc}"
        fi
    done <<< "${proto_includes_raw}"

    # 第一步: protoc-gen-swagger 生成 swagger.json 到 output_dir/tmp/（临时，完成后清理）
    local abs_out_dir="${BCS_ROOT}/${output_dir}"
    local tmp_dir="${abs_out_dir}/tmp"
    mkdir -p "${tmp_dir}"

    info "运行 protoc-gen-swagger 生成 swagger.json..."
    if ! (cd "${abs_source}" && protoc \
        ${proto_i_args} \
        --proto_path=. \
        --swagger_out=logtostderr=true,allow_delete_body=true:"${tmp_dir}" \
        "${proto_file}" 2>&1); then
        rm -rf "${tmp_dir}"
        error "${service}: protoc-gen-swagger 执行失败"
        return 1
    fi

    # 找到生成的 swagger.json（文件名由 proto 文件名决定，可能含子目录）
    local proto_basename
    proto_basename=$(basename "${proto_file}" .proto)
    local swagger_json
    swagger_json=$(find "${tmp_dir}" -name "${proto_basename}.swagger.json" 2>/dev/null | head -1)
    if [[ -z "${swagger_json}" || ! -f "${swagger_json}" ]]; then
        rm -rf "${tmp_dir}"
        error "${service}: protoc-gen-swagger 未生成 ${proto_basename}.swagger.json (在 ${tmp_dir} 中未找到)"
        return 1
    fi

    # 第二步: swagger.json -> output_dir/openapi.yaml
    local dest="${abs_out_dir}/openapi.yaml"
    mkdir -p "${abs_out_dir}"

    info "转换 swagger.json -> openapi.yaml..."
    if ! swagger2openapi --yaml --patch --outfile "${dest}" "${swagger_json}" 2>&1; then
        rm -rf "${tmp_dir}"
        error "${service}: swagger 转换失败"
        return 1
    fi

    # 清理临时目录
    rm -rf "${tmp_dir}"

    success "${service}: 已生成 -> ${dest}"
    _validate "${dest}" "${service}"
}

# ============================================================
# Gorestful 类服务生成
# ============================================================

generate_gorestful() {
    local service="$1"
    step "生成 gorestful 服务: ${service}"

    local source_dir module base_path title version
    source_dir=$(get_config "${service}" "source_dir")
    module=$(get_config "${service}" "module")
    base_path=$(get_config "${service}" "base_path")
    title=$(get_config "${service}" "title")
    version=$(get_config "${service}" "version")

    if [[ -z "${source_dir}" ]]; then
        error "${service}: 缺少 source_dir 配置"
        return 1
    fi

    local abs_source="${BCS_ROOT}/${source_dir}"
    if [[ ! -d "${abs_source}" ]]; then
        error "${service}: source_dir 不存在: ${abs_source}"
        return 1
    fi

    # 获取 router_files 列表
    local router_files_raw
    router_files_raw=$(get_config "${service}" "router_files")
    local router_args=""
    while IFS= read -r rf; do
        [[ -z "${rf}" ]] && continue
        router_args="${router_args} ${abs_source}/${rf}"
    done <<< "${router_files_raw}"

    if [[ -z "${router_args}" ]]; then
        error "${service}: 缺少 router_files 配置"
        return 1
    fi

    local dest="${OPENAPI_DIR}/${module}/${service}/openapi.yaml"
    mkdir -p "$(dirname "${dest}")"

    info "解析 go-restful 路由并生成 OpenAPI..."
    if ! python3 "${SCRIPTS_DIR}/gorestful2openapi.py" \
        --service-dir "${abs_source}" \
        --routers ${router_args} \
        --base-path "${base_path}" \
        --title "${title}" \
        --version "${version}" \
        --output "${dest}" 2>&1; then
        error "${service}: gorestful2openapi 执行失败"
        return 1
    fi

    success "${service}: 已生成 -> ${dest}"

    # 验证
    _validate "${dest}" "${service}"
}

# ============================================================
# Swag 类服务生成
# 流程: swag init -> output_dir/tmp/swagger.json + docs.go -> swagger2openapi_convert.py -> output_dir/openapi.yaml
# ============================================================

generate_swag() {
    local service="$1"
    step "生成 swag 服务: ${service}"

    local source_dir module entry output_dir
    source_dir=$(get_config "${service}" "source_dir")
    module=$(get_config "${service}" "module")
    entry=$(get_config "${service}" "entry")
    output_dir=$(get_config "${service}" "output_dir")

    if [[ -z "${source_dir}" || -z "${entry}" || -z "${output_dir}" ]]; then
        error "${service}: 缺少 source_dir、entry 或 output_dir 配置"
        return 1
    fi

    local abs_source="${BCS_ROOT}/${source_dir}"
    if [[ ! -d "${abs_source}" ]]; then
        error "${service}: source_dir 不存在: ${abs_source}"
        return 1
    fi

    # 检查 swag 是否安装
    if ! command -v swag &>/dev/null; then
        error "swag 未安装，请运行: go install github.com/swaggo/swag/cmd/swag@latest"
        return 1
    fi

    # swag init 输出到 output_dir/tmp/（临时，完成后清理）
    local abs_out_dir="${BCS_ROOT}/${output_dir}"
    local tmp_dir="${abs_out_dir}/tmp"
    mkdir -p "${tmp_dir}"

    info "运行 swag init..."
    if ! (cd "${abs_source}" && swag init \
        --outputTypes go,json \
        --parseDependency \
        -g "${entry}" \
        --output "${tmp_dir}" \
        --exclude ./ 2>&1); then
        rm -rf "${tmp_dir}"
        error "${service}: swag init 执行失败"
        return 1
    fi

    local swagger_json="${tmp_dir}/swagger.json"
    if [[ ! -f "${swagger_json}" ]]; then
        rm -rf "${tmp_dir}"
        error "${service}: swag 未生成 swagger.json"
        return 1
    fi

    # 转换为 output_dir/openapi.yaml
    local dest="${abs_out_dir}/openapi.yaml"
    mkdir -p "${abs_out_dir}"
    info "转换 Swagger 2.0 -> OpenAPI 3.0.1..."
    if ! swagger2openapi --yaml --patch --outfile "${dest}" "${swagger_json}" 2>&1; then
        rm -rf "${tmp_dir}"
        error "${service}: swagger 转换失败"
        return 1
    fi

    # 清理临时目录
    rm -rf "${tmp_dir}"

    success "${service}: 已生成 -> ${dest}"

    # 验证
    _validate "${dest}" "${service}"
}

# ============================================================
# Manual 类服务同步
# ============================================================

sync_manual() {
    local service="$1"
    step "同步手写服务: ${service}"

    local source module
    source=$(get_config "${service}" "source")
    module=$(get_config "${service}" "module")

    if [[ -z "${source}" ]]; then
        error "${service}: 缺少 source 配置"
        return 1
    fi

    local abs_source="${BCS_ROOT}/${source}"
    if [[ ! -f "${abs_source}" ]]; then
        error "${service}: 源文件不存在: ${abs_source}"
        return 1
    fi

    local dest="${OPENAPI_DIR}/${module}/${service}/openapi.yaml"
    mkdir -p "$(dirname "${dest}")"

    # 如果源和目标是同一文件（原地维护类服务），跳过 cp
    if [[ "$(realpath "${abs_source}")" == "$(realpath "${dest}" 2>/dev/null || echo "")" ]]; then
        info "${service}: 原地维护，无需同步"
        success "${service}: 已就绪 -> ${dest}"
    else
        cp "${abs_source}" "${dest}"
        success "${service}: 已同步 -> ${dest}"
    fi

    # 验证
    _validate "${dest}" "${service}"
}

# ============================================================
# 验证
# ============================================================

_validate() {
    local yaml_file="$1"
    local service="$2"
    local validator="${SCRIPTS_DIR}/validate_openapi.py"
    if [[ -f "${validator}" ]]; then
        if python3 "${validator}" "${yaml_file}" &>/dev/null; then
            info "${service}: 验证通过"
        else
            warn "${service}: 验证存在问题，运行以查看详情: python3 ${validator} ${yaml_file}"
        fi
    fi
}

# ============================================================
# 主逻辑
# ============================================================

usage() {
    cat <<EOF
用法: $(basename "$0") [选项]

选项:
  --service <name>    仅生成指定服务（可多次指定）
  --proto-only        仅处理 proto 类服务
  --swag-only         仅处理 swag 类服务
  --manual-only       仅同步 manual 类服务
  --list              列出所有已配置的服务及类型
  --no-validate       跳过验证步骤
  -h, --help          显示帮助

示例:
  $(basename "$0")                                  # 生成所有服务
  $(basename "$0") --service bcs-cluster-manager    # 生成单个服务
  $(basename "$0") --proto-only                     # 仅生成 proto 类
  $(basename "$0") --swag-only                      # 仅生成 swag 类
EOF
}

list_all_services() {
    echo "已配置的服务列表:"
    python3 - <<EOF
import yaml
with open("${CONFIG_FILE}") as f:
    cfg = yaml.safe_load(f)
for name, svc in cfg.get("services", {}).items():
    t = svc.get("type", "?")
    m = svc.get("module", "?")
    print(f"  {name:<40} [{t}]  ({m})")
EOF
}

process_service() {
    local service="$1"
    local svc_type
    svc_type=$(get_config "${service}" "type")

    case "${svc_type}" in
        proto)      generate_proto "${service}" ;;
        gorestful)  generate_gorestful "${service}" ;;
        swag)       generate_swag "${service}" ;;
        manual)     sync_manual "${service}" ;;
        "")
            error "服务 '${service}' 未在 service_config.yaml 中配置"
            return 1
            ;;
        *)
            error "未知服务类型: ${svc_type} (服务: ${service})"
            return 1
            ;;
    esac
}

main() {
    local specific_services=()
    local filter_type=""
    local do_list=false

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --service)
                specific_services+=("$2")
                shift 2
                ;;
            --proto-only)
                filter_type="proto"
                shift
                ;;
            --swag-only)
                filter_type="swag"
                shift
                ;;
            --manual-only)
                filter_type="manual"
                shift
                ;;
            --list)
                do_list=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                # 兼容直接传服务名
                if [[ "$1" != -* ]]; then
                    specific_services+=("$1")
                    shift
                else
                    error "未知选项: $1"
                    usage
                    exit 1
                fi
                ;;
        esac
    done

    if ${do_list}; then
        list_all_services
        exit 0
    fi

    # 检查依赖
    if ! command -v python3 &>/dev/null; then
        error "python3 未安装"
        exit 1
    fi
    if ! python3 -c "import yaml" &>/dev/null; then
        error "pyyaml 未安装，请运行: pip install pyyaml"
        exit 1
    fi
    if ! command -v swagger2openapi &>/dev/null; then
        error "swagger2openapi 未安装，请运行: npm install -g swagger2openapi"
        exit 1
    fi
    if [[ ! -f "${CONFIG_FILE}" ]]; then
        error "配置文件不存在: ${CONFIG_FILE}"
        exit 1
    fi

    # 确定要处理的服务列表
    local services_to_process=()
    if [[ ${#specific_services[@]} -gt 0 ]]; then
        services_to_process=("${specific_services[@]}")
    elif [[ -n "${filter_type}" ]]; then
        while IFS= read -r svc; do
            [[ -n "${svc}" ]] && services_to_process+=("${svc}")
        done < <(list_services_by_type "${filter_type}")
    else
        while IFS= read -r svc; do
            [[ -n "${svc}" ]] && services_to_process+=("${svc}")
        done < <(list_services)
    fi

    if [[ ${#services_to_process[@]} -eq 0 ]]; then
        warn "没有找到需要处理的服务"
        exit 0
    fi

    info "共 ${#services_to_process[@]} 个服务需要处理"

    local success_count=0
    local fail_count=0
    local failed_services=()

    for svc in "${services_to_process[@]}"; do
        if process_service "${svc}"; then
            ((success_count++)) || true
        else
            ((fail_count++)) || true
            failed_services+=("${svc}")
        fi
    done

    echo ""
    echo "============================================"
    echo " 生成完成: 成功 ${success_count} / 失败 ${fail_count}"
    echo "============================================"
    if [[ ${#failed_services[@]} -gt 0 ]]; then
        error "以下服务生成失败:"
        for svc in "${failed_services[@]}"; do
            echo "  - ${svc}"
        done
        exit 1
    fi
}

main "$@"
