#!/usr/bin/env bash
# detect-standards.sh — Level-1 standards preset detection (supports monorepo roots).
# Usage: detect-standards.sh [--json] [workspace_root]
# Exit 0 always when detection completes; prints human table or JSON.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="."
JSON=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --json) JSON=1; shift ;;
    -h|--help)
      echo "Usage: detect-standards.sh [--json] [workspace_root]"
      exit 0
      ;;
    *) ROOT="$1"; shift ;;
  esac
done

ROOT="$(cd "$ROOT" && pwd)"
INDEX="$SCRIPT_DIR/../assets/standards/index.yaml"
export DETECT_ROOT="$ROOT"
export DETECT_INDEX="$INDEX"
export DETECT_JSON="$JSON"

python3 - <<'PY'
import json, os, re, sys
from pathlib import Path

try:
    import yaml
except ImportError:
    print("ERROR: PyYAML required (python3 -c 'import yaml')", file=sys.stderr)
    sys.exit(2)

ROOT = Path(os.environ["DETECT_ROOT"])
INDEX = Path(os.environ["DETECT_INDEX"])
AS_JSON = os.environ.get("DETECT_JSON") == "1"

SKIP_DIR_NAMES = {
    "node_modules", "vendor", ".git", "dist", "build", "coverage",
    ".venv", "venv", "__pycache__", ".tox", "target", "bin",
}

def find_package_roots(max_depth=4):
    roots = []
    if (ROOT / "package.json").is_file():
        roots.append(ROOT)
    for dirpath, dirnames, filenames in os.walk(ROOT):
        rel = Path(dirpath).relative_to(ROOT)
        depth = 0 if str(rel) == "." else len(rel.parts)
        dirnames[:] = [d for d in dirnames if d not in SKIP_DIR_NAMES and not d.startswith(".")]
        if depth > max_depth:
            dirnames.clear()
            continue
        if "package.json" in filenames and Path(dirpath) != ROOT:
            # skip nested test/e2e package roots as primary candidates
            if any(p in ("e2e", "tests", "__tests__", "fixtures") for p in rel.parts):
                continue
            roots.append(Path(dirpath))
    return roots

def find_gomod_roots(max_depth=4):
    roots = []
    if (ROOT / "go.mod").is_file():
        roots.append(ROOT)
    for dirpath, dirnames, filenames in os.walk(ROOT):
        rel = Path(dirpath).relative_to(ROOT)
        depth = 0 if str(rel) == "." else len(rel.parts)
        dirnames[:] = [d for d in dirnames if d not in SKIP_DIR_NAMES and not d.startswith(".")]
        if depth > max_depth:
            dirnames.clear()
            continue
        if "go.mod" in filenames and Path(dirpath) != ROOT:
            roots.append(Path(dirpath))
    return roots

def parse_pkg_deps(pkg_json: Path):
    try:
        data = json.loads(pkg_json.read_text(encoding="utf-8"))
    except Exception:
        return {}
    deps = {}
    deps.update(data.get("dependencies") or {})
    deps.update(data.get("devDependencies") or {})
    return deps

def parse_version(v: str):
    if not v:
        return None
    m = re.search(r"(\d+)\.(\d+)\.(\d+)", v.lstrip("^~>=< "))
    if not m:
        m = re.search(r"(\d+)\.(\d+)", v.lstrip("^~>=< "))
        if not m:
            return None
        return (int(m.group(1)), int(m.group(2)), 0)
    return tuple(int(x) for x in m.groups())

def version_cmp(a, b):
    return (a > b) - (a < b)

def require_block_modules(go_mod: Path):
    """Direct require only — lines ending with '// indirect' are ignored (F2a)."""
    text = go_mod.read_text(encoding="utf-8", errors="replace")
    mods = set()
    for m in re.finditer(r"(?m)^require\s+(\S+)\s+\S+", text):
        # single-line require rarely has indirect; still check
        line = m.group(0)
        if "// indirect" in line:
            continue
        mods.add(m.group(1))
    in_block = False
    for line in text.splitlines():
        if re.match(r"^\s*require\s*\(\s*$", line):
            in_block = True
            continue
        if in_block:
            if ")" in line:
                in_block = False
                continue
            if line.strip().startswith("//") or "// indirect" in line:
                continue
            m = re.match(r"^\s*(\S+)\s+\S+", line)
            if m:
                mods.add(m.group(1))
    return mods

def _path_skipped(path: Path) -> bool:
    parts = set(path.parts)
    if parts & SKIP_DIR_NAMES:
        return True
    # IDE / install layout / hidden segments
    for part in path.parts:
        if part.startswith(".") and part not in (".", ".."):
            return True
    return False

def _rel_to_repo(path: Path):
    try:
        return path.resolve().relative_to(ROOT.resolve())
    except Exception:
        return path

def any_of_exists(root: Path, patterns):
    for p in patterns:
        if "**" in p or "*" in p:
            it = root.rglob(p[3:]) if p.startswith("**/") else root.glob(p)
            for hit in it:
                if hit.is_file() and not _path_skipped(_rel_to_repo(hit)):
                    return True
        else:
            for cand in (root / p, ROOT / p):
                if cand.is_file() and not _path_skipped(_rel_to_repo(cand)):
                    return True
    return False

def require_dirs_ok(root: Path, dirs):
    for d in dirs:
        if not (root / d).is_dir() and not (ROOT / d).is_dir():
            return False
    return True

def rule_ok(rule, pkg_roots, go_roots):
    # Evaluate rule against candidate roots; return (ok, matched_root_rel or None)
    if "any_of_files" in rule and "file" not in rule and "contains_dep" not in rule and "contains_require" not in rule:
        for r in ([ROOT] + pkg_roots + go_roots):
            if any_of_exists(r, rule["any_of_files"]):
                rel = "." if r == ROOT else str(r.relative_to(ROOT))
                return True, rel
        return False, None

    if rule.get("contains_dep") is not None or (
        rule.get("file") == "package.json" and ("contains_dep" in rule or "version_gte" in rule or "version_lt" in rule)
    ):
        deps_keys = rule.get("contains_dep") or []
        for r in pkg_roots:
            pkg = r / "package.json"
            if not pkg.is_file():
                continue
            deps = parse_pkg_deps(pkg)
            if deps_keys and not all(k in deps for k in deps_keys):
                continue
            # version constraints against first contains_dep key or "vue"
            key = deps_keys[0] if deps_keys else None
            if key and ("version_gte" in rule or "version_lt" in rule):
                ver = parse_version(str(deps.get(key, "")))
                if ver is None:
                    continue
                if "version_gte" in rule:
                    want = parse_version(str(rule["version_gte"]))
                    if want is None or version_cmp(ver, want) < 0:
                        continue
                if "version_lt" in rule:
                    want = parse_version(str(rule["version_lt"]))
                    if want is None or version_cmp(ver, want) >= 0:
                        continue
            # also need any_of_files if present in same rule? usually sibling rules
            rel = "." if r == ROOT else str(r.relative_to(ROOT))
            return True, rel
        return False, None

    if rule.get("contains_require") is not None or (
        rule.get("file") == "go.mod" and "contains_require" in rule
    ):
        reqs = rule.get("contains_require") or []
        for r in go_roots:
            gm = r / "go.mod"
            if not gm.is_file():
                continue
            mods = require_block_modules(gm)
            # OR within list
            if reqs and not any(x in mods for x in reqs):
                continue
            rel = "." if r == ROOT else str(r.relative_to(ROOT))
            return True, rel
        return False, None

    if "require_dirs" in rule:
        dirs = rule["require_dirs"]
        for r in go_roots + [ROOT]:
            if require_dirs_ok(r, dirs):
                rel = "." if r == ROOT else str(r.relative_to(ROOT))
                return True, rel
        return False, None

    if "any_of_files" in rule:
        for r in ([ROOT] + pkg_roots + go_roots):
            if any_of_exists(r, rule["any_of_files"]):
                rel = "." if r == ROOT else str(r.relative_to(ROOT))
                return True, rel
        return False, None

    if "file" in rule and "contains_dep" not in rule and "contains_require" not in rule:
        f = rule["file"]
        for r in ([ROOT] + pkg_roots + go_roots):
            if (r / f).is_file():
                rel = "." if r == ROOT else str(r.relative_to(ROOT))
                return True, rel
        return False, None

    return False, None

def preset_match(preset, pkg_roots, go_roots):
    detect = preset.get("detect")
    if detect in ("code-project", "skill-tooling"):
        return None  # cross-cutting handled elsewhere
    if not isinstance(detect, dict):
        return None
    if preset.get("status") == "planned":
        return {"id": preset["id"], "file": preset["file"], "level": "planned", "root": None}
    rules = detect.get("rules") or []
    match_mode = detect.get("match", "all")
    roots_hit = []
    oks = []
    for rule in rules:
        ok, root = rule_ok(rule, pkg_roots, go_roots)
        oks.append(ok)
        if ok and root is not None:
            roots_hit.append(root)
    if match_mode == "all":
        passed = all(oks) if oks else False
    else:
        passed = any(oks)
    if not passed:
        return None
    # prefer deepest / non-dot root
    root = None
    for r in roots_hit:
        if r != ".":
            root = r
            break
    if root is None and roots_hit:
        root = roots_hit[0]
    return {
        "id": preset["id"],
        "name": preset.get("name", preset["id"]),
        "file": preset["file"],
        "level": "1",
        "root": root or ".",
        "tags": preset.get("tags") or [],
    }

data = yaml.safe_load(INDEX.read_text(encoding="utf-8"))
pkg_roots = find_package_roots()
go_roots = find_gomod_roots()

matches = {"frontend": [], "api": [], "backend": []}
unmatched = []
for cat in ("frontend", "api", "backend"):
    cat_data = (data.get("categories") or {}).get(cat) or {}
    presets = cat_data.get("presets") or []
    found = []
    for p in presets:
        m = preset_match(p, pkg_roots, go_roots)
        if m and m.get("level") == "1":
            found.append(m)
    # first Level-1 per category as primary; extras listed
    matches[cat] = found
    if not found:
        unmatched.append(cat)

result = {
    "workspace": str(ROOT),
    "package_roots": ["." if r == ROOT else str(r.relative_to(ROOT)) for r in pkg_roots],
    "gomod_roots": ["." if r == ROOT else str(r.relative_to(ROOT)) for r in go_roots],
    "matches": matches,
    "primary": {c: (matches[c][0] if matches[c] else None) for c in matches},
    "unmatched": unmatched,
}

if AS_JSON:
    print(json.dumps(result, ensure_ascii=False, indent=2))
else:
    print(f"detect-standards workspace={ROOT}")
    print(f"package_roots={result['package_roots']}")
    print(f"gomod_roots={result['gomod_roots']}")
    print(f"{'CAT':<10} {'ID':<16} {'ROOT':<24} FILE")
    print(f"{'---':<10} {'--':<16} {'----':<24} ----")
    for cat, items in matches.items():
        if not items:
            print(f"{cat:<10} {'(none)':<16} {'-':<24} -")
            continue
        for i, m in enumerate(items):
            mark = "*" if i == 0 else " "
            print(f"{cat:<10} {mark}{m['id']:<15} {m['root']:<24} {m['file']}")
    if unmatched:
        print("unmatched:", ", ".join(unmatched))
sys.exit(0)
PY
