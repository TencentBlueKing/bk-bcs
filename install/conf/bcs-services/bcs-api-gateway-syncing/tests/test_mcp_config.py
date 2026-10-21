"""Offline validation of the gateway configuration; never calls a live gateway."""

import re
import unittest
from pathlib import Path

import yaml
from django.template import Context, Engine
from jsonschema import Draft4Validator, ValidationError
from openapi_spec_validator import validate


ROOT = Path(__file__).resolve().parents[1]
GATEWAY = ROOT / "bcs-api-gateway"
HTTP_METHODS = {"get", "post", "put", "patch", "delete", "head", "options", "trace"}
# POST resources that only query, validate or preview and are safe for the read-only MCP.
READ_ONLY_POSTS = {
    "check_drain_node", "check_node_in_cluster", "get_chart_releases", "get_template_version",
    "labels_by_post", "list_template_vaiables", "preview_addons", "preview_deploy_template",
    "query_by_post", "query_range_by_post", "release_preview", "series_by_post",
}


class UniqueKeyLoader(yaml.SafeLoader):
    """Do not silently discard duplicate YAML mapping keys."""

    def construct_mapping(self, node, deep=False):
        keys = [self.construct_object(key, deep=deep) for key, _ in node.value]
        if len(keys) != len(set(keys)):
            raise ValueError("duplicate YAML mapping key")
        return super().construct_mapping(node, deep=deep)


def load_yaml(text):
    return yaml.load(text, Loader=UniqueKeyLoader)


def render_definition(apps=(), managers=(), gateways=()):
    template = Engine().from_string((GATEWAY / "definition.yaml").read_text())
    return load_yaml(template.render(Context({
        "environ": {
            "BKAPI_RELEASE_VERSION": "test-mcp-1",
            "BKAPI_RELEASE_COMMENT": "offline validation",
            "BKAPI_DESCRIPTION": "BCS",
            "BKAPI_STAGE_HOST": "http://bcs-app.svc",
        },
        "settings": {
            "BK_APP_CODE": "bcs",
            "BK_APIGW_MANAGERS": list(managers),
            "BK_API_GRANT_PERMISSIONS_APPS": list(apps),
            "BK_API_APPLY_PERMISSIONS_GATEWAYS": list(gateways),
            "BK_APIGW_RESOURCE_DOCS_BASE_DIR": "/data/docs",
        },
    }, autoescape=False, use_l10n=False)))


class MCPConfigTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.resources = load_yaml((GATEWAY / "resources.yaml").read_text())
        cls.operations = [
            (path, method, operation)
            for path, methods in cls.resources["paths"].items()
            for method, operation in methods.items()
            if method in HTTP_METHODS
        ]
        cls.by_name = {operation["operationId"]: operation for _, _, operation in cls.operations}

    def resolve(self, value):
        if isinstance(value, dict):
            if "$ref" in value:
                self.assertTrue(value["$ref"].startswith("#/components/"))
                target = self.resources
                for key in value["$ref"][2:].split("/"):
                    target = target[key]
                return self.resolve(target)
            return {key: self.resolve(item) for key, item in value.items()}
        if isinstance(value, list):
            return [self.resolve(item) for item in value]
        return value

    def parameters(self, name):
        return self.resolve(self.by_name[name].get("parameters", []))

    def body(self, name):
        return self.resolve(self.by_name[name]["requestBody"]["content"]["application/json"]["schema"])

    def test_valid_openapi_and_local_references(self):
        validate(self.resources)
        self.resolve(self.resources)

    def test_all_130_resources_and_docs_are_in_mcp(self):
        self.assertEqual(len(self.operations), 130)
        self.assertEqual(len(self.by_name), 130)
        definition = render_definition()
        self.assertEqual(definition["spec_version"], 2)
        self.assertNotIn("stage", definition)
        self.assertEqual(len(definition["stages"]), 1)
        servers = {server["name"]: server for server in definition["stages"][0]["mcp_servers"]}
        self.assertEqual(set(servers), {"mcp", "mcp-readonly"})
        for server in servers.values():
            names = server["resource_names"]
            self.assertEqual(len(names), len(set(names)))
            self.assertNotIn("tool_names", server)
        names = servers["mcp"]["resource_names"]
        self.assertEqual(set(names), set(self.by_name))
        self.assertEqual(set(names), {p.stem for p in (GATEWAY / "docs/zh").glob("*.md")})
        self.assertTrue(READ_ONLY_POSTS <= self.by_name.keys())
        read_only = {operation["operationId"] for _, method, operation in self.operations
                     if method == "get" or operation["operationId"] in READ_ONLY_POSTS}
        self.assertEqual(set(servers["mcp-readonly"]["resource_names"]), read_only)

    def test_stage_and_authorization(self):
        for apps in ((), ("bk_apigateway", "bcs-agent")):
            with self.subTest(apps=apps):
                definition = render_definition(apps, ("maintainer",), ("bk-test",))
                stage = definition["stages"][0]
                self.assertEqual(stage["name"], "prod")
                self.assertEqual(stage["vars"], {})
                self.assertEqual(stage["backends"], [{
                    "name": "default",
                    "config": {
                        "timeout": 60,
                        "loadbalance": "roundrobin",
                        "hosts": [{"host": "http://bcs-app.svc", "weight": 100}],
                    },
                }])
                self.assertFalse(stage["rate_limit"]["enabled"])
                self.assertEqual([server["name"] for server in stage["mcp_servers"]], ["mcp", "mcp-readonly"])
                for server in stage["mcp_servers"]:
                    self.assertEqual(server["status"], 1)
                    self.assertEqual(server["protocol_type"], "streamable_http")
                    self.assertIs(server["is_public"], True)
                    self.assertIs(server["oauth2_public_client_enabled"], True)
                    self.assertIs(server["oauth2_personal_client_enabled"], True)
                    self.assertEqual(server["category_names"], ["Official"])
                    self.assertEqual(server["target_app_codes"], ["bcs", *apps])
                    self.assertEqual(server["target_app_codes"], [
                        grant["bk_app_code"] for grant in definition["grant_permissions"]
                    ])
                self.assertEqual(definition["apigateway"]["maintainers"], ["admin", "maintainer"])
                self.assertEqual(definition["apply_permissions"], [{"api_name": "bk-test"}])

    def test_every_resource_has_confirmed_request_parameters(self):
        no_schema = []
        for path, _, operation in self.operations:
            with self.subTest(operation=operation["operationId"]):
                parameters = self.resolve(operation.get("parameters", []))
                path_parameters = [param for param in parameters if param["in"] == "path"]
                self.assertEqual({p["name"] for p in path_parameters}, set(re.findall(r"\{([^}]+)\}", path)))
                self.assertTrue(all(p["required"] for p in path_parameters))
                keys = [(p["in"], p["name"]) for p in parameters]
                self.assertEqual(len(keys), len(set(keys)))
                config = operation["x-bk-apigateway-resource"]
                has_schema = bool(parameters or operation.get("requestBody"))
                self.assertNotEqual(has_schema, config.get("noneSchema", False))
                if config.get("noneSchema"):
                    no_schema.append(operation["operationId"])
                # Do not weaken gateway app verification when adding MCP schemas.
                self.assertTrue(config["authConfig"]["appVerifiedRequired"])
        self.assertEqual(no_schema, ["projects_for_iam"])

    def test_monitor_post_uses_supported_query_parameters(self):
        for name in ("query", "query_range", "labels", "series"):
            with self.subTest(operation=name):
                self.assertEqual(self.parameters(name), self.parameters(name + "_by_post"))
                self.assertNotIn("requestBody", self.by_name[name + "_by_post"])
                headers = {p["name"]: p for p in self.parameters(name) if p["in"] == "header"}
                self.assertTrue(headers["X-Tenant-Project-Code"]["required"])
        for name in ("query_range", "query_range_by_post"):
            required = {p["name"] for p in self.parameters(name) if p.get("required")}
            self.assertTrue({"query", "start", "end", "step"} <= required)
        for name in ("series", "series_by_post", "labels", "labels_by_post", "label_values"):
            match = next(p for p in self.parameters(name) if p["name"] == "match[]")
            self.assertEqual(match["schema"]["type"], "array")
            self.assertEqual(match["style"], "form")
            self.assertTrue(match["explode"])

    def test_resource_tags(self):
        allowed = {"Project", "Namespace", "Cluster", "Helm", "Addons", "TemplateFile", "Monitor", "Webconsole"}
        for _, _, operation in self.operations:
            with self.subTest(operation=operation["operationId"]):
                self.assertEqual(len(operation["tags"]), 1)
                self.assertIn(operation["tags"][0], allowed)

    def test_activity_filters_match_handler(self):
        query_names = {p["name"] for p in self.parameters("get_activity_logs") if p["in"] == "query"}
        self.assertEqual(query_names, {
            "resource_type", "activity_type", "status", "start_time", "end_time", "limit", "offset",
        })

    def test_json_bodies_and_webconsole_constraints(self):
        for path, _, operation in self.operations:
            name = operation["operationId"]
            if "requestBody" in operation:
                with self.subTest(operation=name):
                    body = self.body(name)
                    Draft4Validator.check_schema(body)
                    self.assertEqual(body["type"], "object")
                    self.assertTrue(body["properties"])
                    self.assertFalse(set(re.findall(r"\{([^}]+)\}", path)) & body["properties"].keys())
                    self.assertTrue(set(body.get("required", [])) <= body["properties"].keys())
                    self.assertFalse({"app_secret", "bk_app_secret", "access_token"} & body["properties"].keys())
        console = Draft4Validator(self.body("create_web_console_sessions"))
        console.validate({"operator": "test-user", "container_id": "container-id"})
        console.validate({"operator": "test-user", "namespace": "default", "pod_name": "pod", "container_name": "app"})
        for invalid in (
            {"operator": "test-user"},
            {"container_id": "container-id"},
            {"operator": "test-user", "container_id": ""},
            {"operator": "test-user", "container_id": "container-id", "session_timeout": 1440},
        ):
            with self.subTest(invalid=invalid), self.assertRaises(ValidationError):
                console.validate(invalid)

    def test_image_enables_mcp_sync(self):
        dockerfile = (ROOT / "Dockerfile").read_text()
        self.assertIn("FROM hub.bktencent.com/blueking/apigw-manager:4.2.4", dockerfile)
        self.assertIn("ENV ENABLE_SYNC_MCP_SERVERS=true", dockerfile)
        self.assertIn("COPY bcs-api-gateway /data", dockerfile)


if __name__ == "__main__":
    unittest.main()
