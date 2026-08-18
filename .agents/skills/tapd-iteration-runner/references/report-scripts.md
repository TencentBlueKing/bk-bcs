# 迭代报告脚本参考

> 本文档包含 `tapd-iteration-report` 生成 changelog 和 summary 的详细脚本实现。
> report SKILL.md 中保留精简逻辑，需要实现细节时读取本文档。

## 1. Changelog 生成脚本

### Linux / macOS (Bash)

```bash
CHANGELOG_FILE="specs/${VERSION}/changelog.md"

# 如果 changelog.md 已存在，追加新版本章节；否则创建新文件
if [ ! -f "$CHANGELOG_FILE" ]; then
  echo "# Changelog" > "$CHANGELOG_FILE"
  echo "" >> "$CHANGELOG_FILE"
fi

# 构建本次版本章节
echo "## ${VERSION_TAG}" >> "$CHANGELOG_FILE"
echo "" >> "$CHANGELOG_FILE"
echo "_发布时间：$(date '+%Y-%m-%d %H:%M:%S')_" >> "$CHANGELOG_FILE"
echo "" >> "$CHANGELOG_FILE"

# 遍历 sequence 中的需求，按顺序提取 commit 信息
for ID in $(jq -r '.sequence[]' "$STATE_FILE"); do
  COMMIT_FILE="specs/${VERSION}/${ID}/commit.md"
  if [ -f "$COMMIT_FILE" ]; then
    # 提取需求名称
    STORY_NAME=$(jq -r --arg id "$ID" '
      .stories[$id].name // (.stories[] | .children[$id].name // empty)
    ' "$STATE_FILE")
    
    # 提取 Commit Message 章节内容
    COMMIT_MSG=$(sed -n '/^## Commit Message/,/^## /{/^## Commit Message/d;/^## /d;p}' "$COMMIT_FILE" | sed '/^$/d')
    
    echo "### ${STORY_NAME} (#${ID})" >> "$CHANGELOG_FILE"
    echo "" >> "$CHANGELOG_FILE"
    echo "${COMMIT_MSG}" >> "$CHANGELOG_FILE"
    echo "" >> "$CHANGELOG_FILE"
  fi
done
```

### Windows (PowerShell)

```powershell
$changelogFile = "specs\$VERSION\changelog.md"

if (-not (Test-Path $changelogFile)) {
    "# Changelog`n" | Set-Content $changelogFile -Encoding UTF8
}

$content = @()
$content += "## $VERSION_TAG"
$content += ""
$content += "_发布时间：$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')_"
$content += ""

$sequence = (Get-Content $STATE_FILE | ConvertFrom-Json).sequence
foreach ($id in $sequence) {
    $commitFile = "specs\$VERSION\$id\commit.md"
    if (Test-Path $commitFile) {
        $storyName = # 从 iteration-state.json 提取需求名称
        $commitContent = Get-Content $commitFile -Raw
        # 提取 Commit Message 章节
        $content += "### $storyName (#$id)"
        $content += ""
        $content += $commitMsg
        $content += ""
    }
}

$content -join "`n" | Add-Content $changelogFile -Encoding UTF8
```

## 2. 效能数据汇总脚本

### Linux / macOS (Bash)

```bash
SUMMARY_FILE="specs/${VERSION}/summary.md"

# 如果 summary.md 已存在，追加新版本章节；否则创建新文件
if [ ! -f "$SUMMARY_FILE" ]; then
  echo "# 迭代研发效能报告" > "$SUMMARY_FILE"
  echo "" >> "$SUMMARY_FILE"
fi

# 使用 jq 从 iteration-state.json 提取所有叶子需求
# 注意：children 字段需兼容 null / {} / 不存在三种空值
# 注意：stats.cost / per_stage / duration_sec 均可能不存在，缺失时视为 0

# 提取代码统计
CODE_STATS=$(jq '
  [.stories | to_entries[] | 
    if .value.children and (.value.children | type == "object") and (.value.children | length > 0) then
      .value.children | to_entries[] | .value
    else
      .value
    end
  ] | {
    total_lines: (map(.stats.total // 0) | add),
    add_code: (map(.stats.add_code // 0) | add),
    delete_code: (map(.stats.delete_code // 0) | add),
    logic_code: (map(.stats.logic_code // 0) | add),
    test_code: (map(.stats.test_code // 0) | add),
    docs: (map(.stats.docs // 0) | add),
    files: (map(.stats.files // 0) | add),
    story_count: length
  }
' "$STATE_FILE")

# 提取成本统计——第一层：迭代总体
# credit 和 tokens 仅为 per_stage 各阶段求和；耗时来自 stats.duration_sec
COST_SUMMARY=$(jq '
  [.stories | to_entries[] | 
    if .value.children and (.value.children | type == "object") and (.value.children | length > 0) then
      .value.children | to_entries[] | {id: .key, data: .value}
    else
      {id: .key, data: .value}
    end
  ] | {
    sum_duration: (map(.data.stats.duration_sec // 0) | add),
    sum_credit: (map([.data.stats.cost.per_stage // {} | to_entries[] | .value.credit // 0] | add) | add),
    sum_input_tokens: (map([.data.stats.cost.per_stage // {} | to_entries[] | .value.total_input_tokens // 0] | add) | add),
    sum_output_tokens: (map([.data.stats.cost.per_stage // {} | to_entries[] | .value.total_output_tokens // 0] | add) | add),
    sum_cache_tokens: (map([.data.stats.cost.per_stage // {} | to_entries[] | .value.total_cache_tokens // 0] | add) | add),
    sum_speckit_calls: (map([.data.stats.cost.per_stage // {} | to_entries[] | .value.calls // 0] | add) | add),
    story_count: length
  }
' "$STATE_FILE")

# 提取成本统计——第二层：各需求独立统计
STORY_COSTS=$(jq '
  [.stories | to_entries[] | 
    if .value.children and (.value.children | type == "object") and (.value.children | length > 0) then
      .value.children | to_entries[] | {id: .key, name: .value.name, data: .value}
    else
      {id: .key, name: .value.name, data: .value}
    end
  ] | map({
    id: .id,
    name: .name,
    duration: (.data.stats.duration_sec // 0),
    credit: ([.data.stats.cost.per_stage // {} | to_entries[] | .value.credit // 0] | add),
    input_tokens: ([.data.stats.cost.per_stage // {} | to_entries[] | .value.total_input_tokens // 0] | add),
    output_tokens: ([.data.stats.cost.per_stage // {} | to_entries[] | .value.total_output_tokens // 0] | add),
    cache_tokens: ([.data.stats.cost.per_stage // {} | to_entries[] | .value.total_cache_tokens // 0] | add),
    calls: ([.data.stats.cost.per_stage // {} | to_entries[] | .value.calls // 0] | add)
  }) | map(. + {total_tokens: (.input_tokens + .output_tokens + .cache_tokens)})
' "$STATE_FILE")

# 提取成本统计——第三层：各阶段成本统计（执行层 per_stage 聚合）
STAGE_COSTS=$(jq '
  [.stories | to_entries[] |
    if .value.children and (.value.children | type == "object") and (.value.children | length > 0) then
      .value.children | to_entries[] | .value
    else
      .value
    end
  ] | map(select(.stats.cost.per_stage != null) | .stats.cost.per_stage) |
  reduce .[] as $ps ({};
    reduce ($ps | to_entries[]) as $e (.;
      .[$e.key].credit += ($e.value.credit // 0) |
      .[$e.key].input_tokens += ($e.value.total_input_tokens // 0) |
      .[$e.key].output_tokens += ($e.value.total_output_tokens // 0) |
      .[$e.key].cache_tokens += ($e.value.total_cache_tokens // 0) |
      .[$e.key].calls += ($e.value.calls // 0)
    )
  )
' "$STATE_FILE")

# 将汇总写入 summary.md
cat >> "$SUMMARY_FILE" << EOF

## ${VERSION_TAG}

_迭代周期：${STARTED_AT} → ${END_AT}_

### 代码统计

| 指标 | 数值 |
|------|------|
| 总变更行数 | $(echo "$CODE_STATS" | jq -r '.total_lines') |
| 新增代码 | $(echo "$CODE_STATS" | jq -r '.add_code') |
| 删除代码 | $(echo "$CODE_STATS" | jq -r '.delete_code') |
| 逻辑代码 | $(echo "$CODE_STATS" | jq -r '.logic_code') |
| 测试代码 | $(echo "$CODE_STATS" | jq -r '.test_code') |
| 文档变更 | $(echo "$CODE_STATS" | jq -r '.docs') |
| 变更文件数 | $(echo "$CODE_STATS" | jq -r '.files') |

### 成本汇总

#### 第一层：迭代总体

| 指标 | 值 |
|------|-----|
| 迭代总耗时 | $(echo "$COST_SUMMARY" | jq -r '.sum_duration')s |
| 迭代总 credit | $(echo "$COST_SUMMARY" | jq -r '.sum_credit') |
| 迭代总 tokens（输入） | $(echo "$COST_SUMMARY" | jq -r '.sum_input_tokens') |
| 迭代总 tokens（输出） | $(echo "$COST_SUMMARY" | jq -r '.sum_output_tokens') |
| 迭代总 tokens（缓存） | $(echo "$COST_SUMMARY" | jq -r '.sum_cache_tokens') |
| 迭代 speckit 调用次数 | $(echo "$COST_SUMMARY" | jq -r '.sum_speckit_calls') |
| 需求数量 | $(echo "$COST_SUMMARY" | jq -r '.story_count') |
| 单需求平均耗时 | <sum_duration / story_count> |
| 单需求平均 credit | <sum_credit / story_count> |
| 单需求最高 credit | <max_credit>（需求 #<ID>）|

#### 第二层：各需求独立统计

| 需求 ID | 需求名称 | 耗时 | Credit | 输入 tokens | 输出 tokens | 缓存 tokens | 总 tokens | 调用次数 |
|---------|---------|------|--------|------------|------------|------------|----------|---------|
$(echo "$STORY_COSTS" | jq -r '.[] | "| #\(.id) | \(.name) | \(.duration)s | \(.credit) | \(.input_tokens) | \(.output_tokens) | \(.cache_tokens) | \(.total_tokens) | \(.calls) |"')
| **合计** | — | $(echo "$COST_SUMMARY" | jq -r '.sum_duration')s | $(echo "$COST_SUMMARY" | jq -r '.sum_credit') | $(echo "$COST_SUMMARY" | jq -r '.sum_input_tokens') | $(echo "$COST_SUMMARY" | jq -r '.sum_output_tokens') | $(echo "$COST_SUMMARY" | jq -r '.sum_cache_tokens') | <sum_total> | $(echo "$COST_SUMMARY" | jq -r '.sum_speckit_calls') |

#### 第三层：各阶段成本统计（执行层）

| 阶段 | 总 credit | 占比 | 输入 tokens | 输出 tokens | 缓存 tokens | 调用次数 |
|------|-----------|------|------------|------------|------------|---------|
$(echo "$STAGE_COSTS" | jq -r 'to_entries[] | "| \(.key) | \(.value.credit) | <占比> | \(.value.input_tokens) | \(.value.output_tokens) | \(.value.cache_tokens) | \(.value.calls) |"')
| **执行层合计** | <sum> | 100% | <sum> | <sum> | <sum> | <sum> |
EOF
```

> **实现说明**：以上 bash 脚本为逻辑示意，实际执行时建议通过编排层直接读取
> `iteration-state.json`，用编程方式完成数据汇总和时间计算，确保精确性。
> 时间计算使用 ISO 8601 格式的 `started_at` 和 `end_at` 字段求差值，
> 输出格式为 "Xh Ym"（如 "2h 35m"）。
> 
> **credit 计算规则**：每个需求的 credit = Σ(`per_stage.*.credit`)，tokens 同理（仅执行层）。
> 需求耗时直接取 runner 记录的 `stats.duration_sec`。

### Windows (PowerShell)

```powershell
$summaryFile = "specs\$VERSION\summary.md"

if (-not (Test-Path $summaryFile)) {
    "# 迭代研发效能报告`n" | Set-Content $summaryFile -Encoding UTF8
}

$state = Get-Content $STATE_FILE | ConvertFrom-Json

# 提取所有叶子需求
$leaves = @()
foreach ($key in $state.stories.PSObject.Properties.Name) {
    $story = $state.stories.$key
    if ($story.children -and $story.children.PSObject.Properties.Count -gt 0) {
        foreach ($childKey in $story.children.PSObject.Properties.Name) {
            $leaves += @{ id = $childKey; data = $story.children.$childKey }
        }
    } else {
        $leaves += @{ id = $key; data = $story }
    }
}

# 代码统计
$codeStats = @{ total = 0; add_code = 0; delete_code = 0; logic_code = 0; test_code = 0; docs = 0; files = 0 }
foreach ($leaf in $leaves) {
    $s = $leaf.data.stats
    if ($s) {
        $codeStats.total += ($s.total ?? 0)
        $codeStats.add_code += ($s.add_code ?? 0)
        $codeStats.delete_code += ($s.delete_code ?? 0)
        $codeStats.logic_code += ($s.logic_code ?? 0)
        $codeStats.test_code += ($s.test_code ?? 0)
        $codeStats.docs += ($s.docs ?? 0)
        $codeStats.files += ($s.files ?? 0)
    }
}

# 成本统计——各需求独立计算（第二层）
$storyCosts = @()
$sumDuration = 0; $sumCredit = 0; $sumInput = 0; $sumOutput = 0; $sumCache = 0; $sumCalls = 0
$stageCosts = @{}

foreach ($leaf in $leaves) {
    $cost = $leaf.data.stats.cost
    $perStage = if ($cost -and $cost.per_stage) { $cost.per_stage } else { $null }

    # 耗时来自 runner 记录的 duration_sec
    $dur = $leaf.data.stats.duration_sec ?? 0

    # per_stage 数据求和（执行层即全部成本）
    $stageCredit = 0; $stageInput = 0; $stageOutput = 0; $stageCache = 0; $stageCalls = 0
    if ($perStage) {
        foreach ($stage in $perStage.PSObject.Properties) {
            $stageCredit += ($stage.Value.credit ?? 0)
            $stageInput += ($stage.Value.total_input_tokens ?? 0)
            $stageOutput += ($stage.Value.total_output_tokens ?? 0)
            $stageCache += ($stage.Value.total_cache_tokens ?? 0)
            $stageCalls += ($stage.Value.calls ?? 0)
            # 第三层聚合
            if (-not $stageCosts[$stage.Name]) {
                $stageCosts[$stage.Name] = @{ credit = 0; input_tokens = 0; output_tokens = 0; cache_tokens = 0; calls = 0 }
            }
            $stageCosts[$stage.Name].credit += ($stage.Value.credit ?? 0)
            $stageCosts[$stage.Name].input_tokens += ($stage.Value.total_input_tokens ?? 0)
            $stageCosts[$stage.Name].output_tokens += ($stage.Value.total_output_tokens ?? 0)
            $stageCosts[$stage.Name].cache_tokens += ($stage.Value.total_cache_tokens ?? 0)
            $stageCosts[$stage.Name].calls += ($stage.Value.calls ?? 0)
        }
    }

    $storyCosts += @{
        id = $leaf.id; name = $leaf.data.name
        duration = $dur; credit = $stageCredit
        input_tokens = $stageInput; output_tokens = $stageOutput
        cache_tokens = $stageCache; calls = $stageCalls
    }

    $sumDuration += $dur; $sumCredit += $stageCredit
    $sumInput += $stageInput; $sumOutput += $stageOutput
    $sumCache += $stageCache; $sumCalls += $stageCalls
}

# 写入 summary.md（格式同 Bash 版本，包含代码统计 + 三层成本汇总）
# ...（按第一层/第二层/第三层格式输出表格）
```
