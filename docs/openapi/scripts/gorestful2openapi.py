#!/usr/bin/env python3
"""
解析 go-restful 框架的路由注册文件（router.go）及 handler 源文件，
自动生成 OpenAPI 3.0.1 YAML。

工作原理:
1. 解析 router.go 中的 ws.Route(METHOD("/path").To(HandlerFunc)) 注册模式
2. 在 handler 源文件中查找对应函数的注释作为 summary/description
3. 从 handler 文件中提取 ReadEntity/request body struct 信息
4. 输出标准 OpenAPI 3.0.1 YAML

用法:
    python3 gorestful2openapi.py \\
        --service-dir bcs-services/bcs-user-manager \\
        --routers app/user-manager/v1http/router.go \\
                  app/user-manager/v3http/router.go \\
        --base-path /usermanager \\
        --title "BCS User Manager API" \\
        --version 0.0.1 \\
        --output openapi/bcs-services/bcs-user-manager/openapi.yaml

    # 通过 service_config.yaml 驱动（由 generate.sh 调用）
    python3 gorestful2openapi.py --config openapi/service_config.yaml --service bcs-user-manager
"""

import re
import os
import sys
import argparse
from typing import Optional

try:
    import yaml
except ImportError:
    print("ERROR: 需要安装 pyyaml: pip install pyyaml", file=sys.stderr)
    sys.exit(1)


# ============================================================
# 路由解析
# ============================================================

# go-restful 路由注册支持以下几种模式:
#   ws.Route(ws.GET("/path").To(Handler))
#   ws.Route(auth.Func(ws.POST("/path")).To(Handler))
#   ws.Route(auth.Func(auth.Func2(ws.PUT("/path"))).To(Handler))
#   ws.Route(auth.Func(auth.Func2(ws.GET("/path"))).To(Handler.Method))
# 统一策略: 在单行中提取 ws.METHOD("path") 和 .To(handler)

LINE_ROUTE_PATTERN = re.compile(
    r'ws\.(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\("([^"]+)"\)'
    r'.*?\.To\(([^)]+)\)'
    , re.DOTALL
)

# 多行路由（跨行，但实际上 bcs-user-manager 的路由都是单行）
# 用于处理 v3http 的跨行嵌套
MULTILINE_ROUTE_PATTERN = re.compile(
    r'ws\.Route\(([\s\S]+?)\)',
    re.DOTALL
)


def parse_routes_from_file(router_file: str, base_path: str = "") -> list:
    """
    解析 router.go 文件中的所有路由注册，返回路由列表。
    每个路由: { method, path, handler, tag }
    """
    with open(router_file, 'r', encoding='utf-8') as f:
        content = f.read()

    routes = []
    tag = _infer_tag_from_path(router_file)
    seen = set()

    # 策略：逐行扫描，每行只要包含 ws.METHOD("path") 和 .To(handler) 就提取
    lines = content.split('\n')
    for line in lines:
        m = LINE_ROUTE_PATTERN.search(line)
        if m:
            method = m.group(1).lower()
            path = m.group(2)
            handler = _clean_handler(m.group(3).strip())
            full_path = _normalize_path(base_path + path)
            key = (method, full_path)
            if key not in seen:
                seen.add(key)
                routes.append({
                    "method": method,
                    "path": full_path,
                    "handler": handler,
                    "tag": tag,
                })

    return routes


def _infer_tag_from_path(router_file: str) -> str:
    """从 router 文件路径推断 API tag"""
    # e.g. v1http/router.go -> V1
    # e.g. v3http/router.go -> V3
    parts = router_file.replace("\\", "/").split("/")
    for p in parts:
        if re.match(r'v\d+http', p):
            ver = re.search(r'v(\d+)', p)
            if ver:
                return f"V{ver.group(1)}"
    return "API"


def _normalize_path(path: str) -> str:
    """将 go-restful {param} 格式转为 OpenAPI {param} 格式（已兼容）"""
    # go-restful 使用 {param}，OpenAPI 也使用 {param}，无需转换
    # 但要处理 go-restful 特有的 {param:pattern} 格式
    path = re.sub(r'\{(\w+):[^}]+\}', r'{\1}', path)
    return path


def _clean_handler(handler: str) -> str:
    """清理 handler 名称，提取最后的函数名"""
    # tokenHandler.CreateToken -> CreateToken
    # user.CreateAdminUser -> CreateAdminUser
    # activity.SearchActivities -> SearchActivities
    if '.' in handler:
        return handler.split('.')[-1]
    return handler.strip()


# ============================================================
# Handler 注释和 struct 解析
# ============================================================

def parse_handler_comments(source_dir: str) -> dict:
    """
    扫描 source_dir 下所有 .go 文件，
    提取函数注释: { FuncName: { summary, description, request_body } }
    """
    result = {}
    go_files = []
    for root, _, files in os.walk(source_dir):
        if 'vendor' in root:
            continue
        for f in files:
            if f.endswith('.go') and not f.endswith('_test.go'):
                go_files.append(os.path.join(root, f))

    for filepath in go_files:
        _parse_single_file(filepath, result)

    return result


def _parse_single_file(filepath: str, result: dict):
    """解析单个 Go 文件中的函数注释和 ReadEntity 调用"""
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    i = 0
    while i < len(lines):
        line = lines[i].strip()

        # 匹配函数定义（handler 签名）
        func_match = re.match(
            r'^func\s+(?:\([^)]+\)\s+)?(\w+)\s*\('
            r'(?:[^)]*request\s+\*restful\.Request[^)]*|[^)]*\*restful\.Request[^)]*)',
            line
        )
        if func_match:
            func_name = func_match.group(1)
            # 向前查找注释
            comment_lines = []
            j = i - 1
            while j >= 0:
                comment_line = lines[j].strip()
                if comment_line.startswith('//'):
                    comment_lines.insert(0, comment_line[2:].strip())
                    j -= 1
                elif comment_line == '' and j > 0 and lines[j-1].strip().startswith('//'):
                    j -= 1
                else:
                    break

            # 过滤掉 nolint 等非描述注释
            desc_lines = [
                l for l in comment_lines
                if l and not l.startswith('nolint') and not l.startswith('NOCC:')
                and not l.startswith('Package ')
            ]

            summary = ""
            description = ""
            if desc_lines:
                # 第一行作为 summary
                first = desc_lines[0]
                # 去掉函数名前缀（如 "CreateAdminUser create a admin user" -> "create a admin user"）
                clean_first = re.sub(r'^' + re.escape(func_name) + r'\s+', '', first, flags=re.IGNORECASE)
                summary = clean_first.capitalize() if clean_first else first.capitalize()
                if len(desc_lines) > 1:
                    description = " ".join(desc_lines[1:])

            # 向后查找 ReadEntity 调用以确定 request body struct
            request_body_struct = None
            k = i + 1
            end = min(i + 30, len(lines))
            while k < end:
                rb_match = re.search(r'request\.ReadEntity\(&(\w+)\)', lines[k])
                if rb_match:
                    request_body_struct = rb_match.group(1)
                    break
                # 检查函数结束
                if re.match(r'^func\s+', lines[k].strip()):
                    break
                k += 1

            if func_name not in result:
                result[func_name] = {}
            if summary:
                result[func_name]['summary'] = summary
            if description:
                result[func_name]['description'] = description
            if request_body_struct:
                result[func_name]['request_body_var'] = request_body_struct

        i += 1


def parse_structs(source_dir: str) -> dict:
    """
    扫描 source_dir 下所有 .go 文件，提取 struct 定义。
    返回: { StructName: { description, fields: [{name, type, json_name, description}] } }
    """
    structs = {}
    go_files = []
    for root, _, files in os.walk(source_dir):
        if 'vendor' in root:
            continue
        for f in files:
            if f.endswith('.go') and not f.endswith('_test.go'):
                go_files.append(os.path.join(root, f))

    for filepath in go_files:
        _parse_structs_from_file(filepath, structs)

    return structs


def _parse_structs_from_file(filepath: str, structs: dict):
    """解析单个文件中的所有 struct 定义"""
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    i = 0
    while i < len(lines):
        line = lines[i].strip()
        struct_match = re.match(r'^type\s+(\w+)\s+struct\s*\{', line)
        if struct_match:
            struct_name = struct_match.group(1)

            # 向前查找注释
            comments = []
            j = i - 1
            while j >= 0 and lines[j].strip().startswith('//'):
                comments.insert(0, lines[j].strip()[2:].strip())
                j -= 1
            struct_desc = " ".join([
                c for c in comments
                if c and not c.startswith('nolint') and not c.startswith('NOCC:')
            ])
            # 去掉 "StructName xxx" 格式的 Go lint 注释前缀
            struct_desc = re.sub(r'^' + re.escape(struct_name) + r'\s+', '', struct_desc).strip()

            fields = []
            k = i + 1
            while k < len(lines):
                field_line = lines[k].strip()
                if field_line == '}':
                    break
                if not field_line or field_line.startswith('//'):
                    k += 1
                    continue
                field = _parse_struct_field(field_line)
                if field:
                    fields.append(field)
                k += 1

            if struct_name not in structs and fields:
                structs[struct_name] = {
                    'description': struct_desc,
                    'fields': fields,
                }
        i += 1


def _parse_struct_field(line: str) -> Optional[dict]:
    """解析 struct 字段，提取名称、类型、json tag 和注释"""
    # 字段格式: FieldName TypeName `json:"field_name" ...` // comment
    # 也处理: FieldName TypeName // comment (无 tag)
    comment = ""
    if '//' in line:
        parts = line.split('//', 1)
        line = parts[0].strip()
        comment = parts[1].strip()

    # 提取 json tag
    json_name = None
    tag_match = re.search(r'`[^`]*json:"([^",]+)', line)
    if tag_match:
        json_name = tag_match.group(1)
        if json_name == '-':
            return None

    # 移除 tag
    line = re.sub(r'`[^`]*`', '', line).strip()
    parts = line.split()
    if len(parts) < 2:
        return None

    field_name = parts[0]
    field_type = parts[1]

    if json_name is None:
        # 将驼峰转蛇形
        json_name = re.sub(r'(?<!^)(?=[A-Z])', '_', field_name).lower()

    return {
        'name': field_name,
        'json_name': json_name,
        'type': field_type,
        'description': comment,
    }


# ============================================================
# OpenAPI 生成
# ============================================================

GO_TYPE_MAP = {
    'string': {'type': 'string'},
    'int': {'type': 'integer'},
    'int32': {'type': 'integer', 'format': 'int32'},
    'int64': {'type': 'integer', 'format': 'int64'},
    'uint': {'type': 'integer'},
    'uint32': {'type': 'integer', 'format': 'int32'},
    'uint64': {'type': 'integer', 'format': 'int64'},
    'float32': {'type': 'number', 'format': 'float'},
    'float64': {'type': 'number', 'format': 'double'},
    'bool': {'type': 'boolean'},
    'time.Time': {'type': 'string', 'format': 'date-time'},
    '*time.Time': {'type': 'string', 'format': 'date-time'},
    'byte': {'type': 'string', 'format': 'byte'},
    '[]byte': {'type': 'string', 'format': 'byte'},
}


def go_type_to_schema(go_type: str, known_structs: dict) -> dict:
    """将 Go 类型转为 OpenAPI schema"""
    go_type = go_type.strip('*')
    if go_type in GO_TYPE_MAP:
        return dict(GO_TYPE_MAP[go_type])
    if go_type.startswith('[]'):
        inner = go_type[2:].strip('*')
        item_schema = go_type_to_schema(inner, known_structs)
        return {'type': 'array', 'items': item_schema}
    if go_type in known_structs:
        return {'$ref': f'#/components/schemas/{go_type}'}
    return {'type': 'object'}


def struct_to_schema(struct_info: dict, known_structs: dict) -> dict:
    """将 struct 定义转为 OpenAPI schema"""
    schema = {'type': 'object'}
    if struct_info.get('description'):
        schema['description'] = struct_info['description']
    properties = {}
    for field in struct_info.get('fields', []):
        prop = go_type_to_schema(field['type'], known_structs)
        if field.get('description'):
            prop['description'] = field['description']
        properties[field['json_name']] = prop
    if properties:
        schema['properties'] = properties
    return schema


def extract_path_params(path: str) -> list:
    """从路径中提取 path 参数"""
    params = []
    for m in re.finditer(r'\{(\w+)\}', path):
        params.append({
            'name': m.group(1),
            'in': 'path',
            'required': True,
            'description': f'[Path参数] {m.group(1)}',
            'schema': {'type': 'string'},
        })
    return params


def generate_operation_id(method: str, path: str, handler: str) -> str:
    """生成 operationId"""
    if handler:
        return handler
    # 从路径生成
    parts = [p for p in path.split('/') if p and not p.startswith('{')]
    return method.capitalize() + ''.join(p.capitalize() for p in parts[-2:])


def build_openapi(
    routes: list,
    handler_info: dict,
    structs: dict,
    title: str,
    version: str,
    base_path: str,
    servers_url: str = None,
) -> dict:
    """构建完整的 OpenAPI 3.0.1 文档"""
    doc = {
        'openapi': '3.0.1',
        'info': {
            'title': title,
            'description': f'{title}，提供用户、集群、Token、权限等管理能力。',
            'version': version,
        },
        'paths': {},
    }

    if servers_url:
        doc['servers'] = [{'url': servers_url}]
    elif base_path:
        doc['servers'] = [{'url': base_path}]

    # 收集使用到的 struct，用于 components/schemas
    used_structs = set()

    for route in routes:
        path = route['path']
        method = route['method']
        handler = route['handler']
        tag = route['tag']

        if path not in doc['paths']:
            doc['paths'][path] = {}

        operation = {}

        # tags
        operation['tags'] = [tag]

        # summary / description
        h_info = handler_info.get(handler, {})
        if h_info.get('summary'):
            operation['summary'] = h_info['summary']
        else:
            operation['summary'] = _handler_to_summary(handler)

        if h_info.get('description'):
            operation['description'] = h_info['description']

        # operationId
        operation['operationId'] = generate_operation_id(method, path, handler)

        # parameters (path params)
        params = extract_path_params(path)
        if params:
            operation['parameters'] = params

        # requestBody
        if method in ('post', 'put', 'patch') and h_info.get('request_body_var'):
            var_name = h_info['request_body_var']
            # 猜测 struct 名：通常 form 变量是 FormStruct 类型
            struct_name = _guess_struct_name(var_name, structs, handler)
            if struct_name:
                used_structs.add(struct_name)
                operation['requestBody'] = {
                    'required': True,
                    'content': {
                        'application/json': {
                            'schema': {'$ref': f'#/components/schemas/{struct_name}'}
                        }
                    }
                }

        # responses
        operation['responses'] = {
            '200': {
                'description': '操作成功',
                'content': {
                    'application/json': {
                        'schema': {'$ref': '#/components/schemas/CommonResponse'}
                    }
                }
            },
            '400': {
                'description': '请求参数错误',
                'content': {
                    'application/json': {
                        'schema': {'$ref': '#/components/schemas/ErrorResponse'}
                    }
                }
            },
        }

        doc['paths'][path][method] = operation

    # components/schemas
    components = {
        'schemas': {
            'CommonResponse': {
                'type': 'object',
                'properties': {
                    'result': {'type': 'boolean', 'description': '请求是否成功'},
                    'code': {'type': 'integer', 'description': '状态码，0 表示成功'},
                    'message': {'type': 'string', 'description': '返回信息'},
                    'data': {'type': 'object', 'description': '返回数据'},
                }
            },
            'ErrorResponse': {
                'type': 'object',
                'properties': {
                    'result': {'type': 'boolean'},
                    'code': {'type': 'integer', 'description': '错误码'},
                    'message': {'type': 'string', 'description': '错误信息'},
                    'data': {'type': 'object'},
                }
            }
        }
    }

    # 递归收集所有需要的 struct（包含嵌套引用）
    def collect_struct(name: str):
        if name in components['schemas'] or name not in structs:
            return
        components['schemas'][name] = struct_to_schema(structs[name], structs)
        # 递归处理嵌套引用
        for field in structs[name].get('fields', []):
            inner = field['type'].strip('*[]')
            if inner in structs:
                collect_struct(inner)

    for struct_name in used_structs:
        collect_struct(struct_name)

    doc['components'] = components
    return doc


def _handler_to_summary(handler: str) -> str:
    """将 handler 函数名转为人类可读的 summary"""
    # CreateAdminUser -> Create Admin User
    words = re.sub(r'(?<=[a-z])(?=[A-Z])|(?<=[A-Z])(?=[A-Z][a-z])', ' ', handler)
    return words


def _guess_struct_name(var_name: str, structs: dict, handler: str) -> Optional[str]:
    """
    根据变量名猜测对应的 struct 名。
    例如: var form CreateTokenForm -> struct CreateTokenForm
    """
    # 直接按变量类型查找（需要从源码读取）
    # 这里用启发式规则：查找包含 handler 名或 "Form"/"Request" 的 struct
    candidates = []
    for sname in structs:
        if 'Form' in sname or 'Request' in sname or 'Req' in sname:
            candidates.append(sname)

    # 找 handler 中包含的关键词
    handler_lower = handler.lower()
    for c in candidates:
        c_lower = c.lower().replace('form', '').replace('request', '').replace('req', '')
        if c_lower and c_lower in handler_lower:
            return c

    return None


# ============================================================
# 主逻辑
# ============================================================

def load_from_service_config(config_path: str, service_name: str, base_dir: str):
    """从 service_config.yaml 加载服务配置"""
    with open(config_path, 'r', encoding='utf-8') as f:
        cfg = yaml.safe_load(f)

    svc = cfg.get('services', {}).get(service_name)
    if not svc:
        print(f"ERROR: 服务 '{service_name}' 未在配置文件中找到", file=sys.stderr)
        sys.exit(1)

    source_dir = os.path.join(base_dir, svc.get('source_dir', ''))
    routers = [os.path.join(source_dir, r) for r in svc.get('router_files', [])]
    base_path = svc.get('base_path', '')
    title = svc.get('title', f'{service_name} API')
    version = svc.get('version', '0.0.1')
    module = svc.get('module', 'bcs-services')
    output = os.path.join(base_dir, 'openapi', module, service_name, 'openapi.yaml')

    return source_dir, routers, base_path, title, version, output


def main():
    parser = argparse.ArgumentParser(
        description='从 go-restful router.go 生成 OpenAPI 3.0.1 YAML'
    )
    parser.add_argument('--service-dir', help='服务根目录（相对于仓库根目录或绝对路径）')
    parser.add_argument('--routers', nargs='+', help='router.go 文件路径（可多个），相对于 service-dir')
    parser.add_argument('--base-path', default='', help='服务 base path (e.g. /usermanager)')
    parser.add_argument('--title', default='BCS API', help='API 文档标题')
    parser.add_argument('--version', default='0.0.1', help='API 版本')
    parser.add_argument('--output', '-o', help='输出 YAML 文件路径，省略则输出到 stdout')
    parser.add_argument('--config', help='service_config.yaml 路径（由 generate.sh 调用）')
    parser.add_argument('--service', help='服务名（配合 --config 使用）')
    parser.add_argument('--base-dir', default=None, help='仓库根目录')
    args = parser.parse_args()

    # 确定 base_dir
    base_dir = args.base_dir
    if not base_dir:
        script_dir = os.path.dirname(os.path.abspath(__file__))
        check = script_dir
        for _ in range(10):
            if os.path.isdir(os.path.join(check, 'bcs-services')):
                base_dir = check
                break
            check = os.path.dirname(check)
        if not base_dir:
            base_dir = os.getcwd()

    # 从 service_config.yaml 加载（generate.sh 调用路径）
    if args.config and args.service:
        service_dir, router_files, base_path, title, version, output_path = \
            load_from_service_config(args.config, args.service, base_dir)
    else:
        if not args.service_dir or not args.routers:
            parser.error("必须提供 --service-dir 和 --routers，或使用 --config + --service")

        service_dir = args.service_dir
        if not os.path.isabs(service_dir):
            service_dir = os.path.join(base_dir, service_dir)

        router_files = []
        for r in args.routers:
            if os.path.isabs(r):
                router_files.append(r)
            else:
                router_files.append(os.path.join(service_dir, r))

        base_path = args.base_path
        title = args.title
        version = args.version
        output_path = args.output

    # 解析所有 router 文件
    all_routes = []
    for rf in router_files:
        if not os.path.exists(rf):
            print(f"WARNING: router 文件不存在: {rf}", file=sys.stderr)
            continue
        print(f"  解析路由: {rf}")
        routes = parse_routes_from_file(rf, base_path)
        print(f"    发现 {len(routes)} 条路由")
        all_routes.extend(routes)

    if not all_routes:
        print("ERROR: 未解析到任何路由", file=sys.stderr)
        sys.exit(1)

    # 解析 handler 注释和 struct
    print(f"  扫描 handler 源码: {service_dir}")
    handler_info = parse_handler_comments(service_dir)
    print(f"    发现 {len(handler_info)} 个函数注释")
    structs = parse_structs(service_dir)
    print(f"    发现 {len(structs)} 个 struct 定义")

    # 构建 OpenAPI
    doc = build_openapi(
        routes=all_routes,
        handler_info=handler_info,
        structs=structs,
        title=title,
        version=version,
        base_path=base_path,
    )

    print(f"  共生成 {len(all_routes)} 个接口")

    # 输出
    yaml_str = yaml.dump(
        doc,
        allow_unicode=True,
        default_flow_style=False,
        sort_keys=False,
        width=120,
    )

    if output_path:
        os.makedirs(os.path.dirname(os.path.abspath(output_path)), exist_ok=True)
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(yaml_str)
        print(f"  已输出到: {output_path}")
    else:
        print(yaml_str)


if __name__ == '__main__':
    main()
