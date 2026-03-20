# 状态常量迁移指南

## 概述

将现有代码从字符串字面量迁移到常量定义，提升代码质量和可维护性。

**影响范围**: 112 处字符串字面量需要替换

---

## 现状分析

### 受影响的文件

```
internal/executor/subscription_executor.go    - 15 处
internal/executor/stage_executor.go           - 11 处
internal/executor/native_executor.go          - 43 处
internal/executor/localization_executor.go    - 15 处
internal/executor/k8s_resource_executor.go    - 12 处
internal/executor/http_executor.go            - 9 处
internal/executor/job_executor.go             - 7 处
---------------------------------------------------
总计                                          - 112 处
```

### 常见的字符串字面量

| 字符串 | 使用次数 | 常量名称 |
|--------|---------|---------|
| `"Succeeded"` | 45 | `drv1alpha1.PhaseSucceeded` |
| `"Failed"` | 52 | `drv1alpha1.PhaseFailed` |
| `"Running"` | 6 | `drv1alpha1.PhaseRunning` |
| `"Skipped"` | 9 | `drv1alpha1.PhaseSkipped` |

---

## 迁移步骤

### Step 1: 导入常量包

确保所有 executor 文件都导入了 `drv1alpha1` 包：

```go
import (
    drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)
```

✅ **已完成**: 所有 executor 文件已导入

---

### Step 2: 批量替换

使用编辑器的查找替换功能（支持正则表达式）：

#### 2.1 替换 Phase = "Succeeded"

**查找**:
```regex
\.Phase\s*=\s*"Succeeded"
```

**替换为**:
```go
.Phase = drv1alpha1.PhaseSucceeded
```

#### 2.2 替换 Phase = "Failed"

**查找**:
```regex
\.Phase\s*=\s*"Failed"
```

**替换为**:
```go
.Phase = drv1alpha1.PhaseFailed
```

#### 2.3 替换 Phase = "Running"

**查找**:
```regex
\.Phase\s*=\s*"Running"
```

**替换为**:
```go
.Phase = drv1alpha1.PhaseRunning
```

#### 2.4 替换 Phase = "Skipped"

**查找**:
```regex
\.Phase\s*=\s*"Skipped"
```

**替换为**:
```go
.Phase = drv1alpha1.PhaseSkipped
```

#### 2.5 替换条件判断

**查找**:
```regex
\.Phase\s*==\s*"Succeeded"
```

**替换为**:
```go
.Phase == drv1alpha1.PhaseSucceeded
```

**查找**:
```regex
\.Phase\s*!=\s*"Succeeded"
```

**替换为**:
```go
.Phase != drv1alpha1.PhaseSucceeded
```

---

### Step 3: 手动检查特殊情况

某些情况需要手动处理：

#### 3.1 结构体初始化

**Before**:
```go
status := &drv1alpha1.ActionStatus{
    Name:      actionStatus.Name,
    Phase:     "Running",
    StartTime: &metav1.Time{Time: time.Now()},
}
```

**After**:
```go
status := &drv1alpha1.ActionStatus{
    Name:      actionStatus.Name,
    Phase:     drv1alpha1.PhaseRunning,
    StartTime: &metav1.Time{Time: time.Now()},
}
```

#### 3.2 Switch 语句

**Before**:
```go
switch execution.Spec.OperationType {
case "Execute":
    return e.ExecutePlan(ctx, plan, execution)
case "Revert":
    return e.RevertPlan(ctx, plan, execution)
}
```

**After**:
```go
switch execution.Spec.OperationType {
case drv1alpha1.OperationTypeExecute:
    return e.ExecutePlan(ctx, plan, execution)
case drv1alpha1.OperationTypeRevert:
    return e.RevertPlan(ctx, plan, execution)
}
```

---

### Step 4: 验证编译

```bash
# 编译检查
go build ./internal/executor/...

# 运行测试
make test

# 代码格式化
go fmt ./...

# 静态检查
go vet ./...
```

---

### Step 5: 代码审查

使用 `grep` 查找是否有遗漏的字符串字面量：

```bash
# 检查是否还有 Phase 字符串字面量
grep -r 'Phase.*=.*"Succeeded"' internal/executor/
grep -r 'Phase.*=.*"Failed"' internal/executor/
grep -r 'Phase.*=.*"Running"' internal/executor/
grep -r 'Phase.*=.*"Skipped"' internal/executor/

# 检查条件判断
grep -r 'Phase.*==.*"Succeeded"' internal/executor/
grep -r 'Phase.*!=.*"Succeeded"' internal/executor/

# 如果没有输出，说明迁移完成！
```

---

## 快速迁移脚本（可选）

如果你想自动化这个过程，可以使用以下脚本：

```bash
#!/bin/bash

# migrate-status-constants.sh

FILES=$(find internal/executor -name "*.go" -type f)

for file in $FILES; do
    echo "Processing: $file"
    
    # 替换 Phase 赋值
    sed -i 's/\.Phase = "Succeeded"/\.Phase = drv1alpha1.PhaseSucceeded/g' "$file"
    sed -i 's/\.Phase = "Failed"/\.Phase = drv1alpha1.PhaseFailed/g' "$file"
    sed -i 's/\.Phase = "Running"/\.Phase = drv1alpha1.PhaseRunning/g' "$file"
    sed -i 's/\.Phase = "Skipped"/\.Phase = drv1alpha1.PhaseSkipped/g' "$file"
    sed -i 's/\.Phase = "Pending"/\.Phase = drv1alpha1.PhasePending/g' "$file"
    
    # 替换 Phase 比较
    sed -i 's/\.Phase == "Succeeded"/\.Phase == drv1alpha1.PhaseSucceeded/g' "$file"
    sed -i 's/\.Phase != "Succeeded"/\.Phase != drv1alpha1.PhaseSucceeded/g' "$file"
    sed -i 's/\.Phase == "Failed"/\.Phase == drv1alpha1.PhaseFailed/g' "$file"
    sed -i 's/\.Phase == "Running"/\.Phase == drv1alpha1.PhaseRunning/g' "$file"
    sed -i 's/\.Phase == "Pending"/\.Phase == drv1alpha1.PhasePending/g' "$file"
done

echo "Migration completed! Please run 'make test' to verify."
```

**使用方法**:
```bash
chmod +x migrate-status-constants.sh
./migrate-status-constants.sh
make test
```

---

## 迁移前后对比

### Before（字符串字面量）

```go
func (e *LocalizationActionExecutor) Rollback(...) (*drv1alpha1.ActionStatus, error) {
    rollbackStatus := &drv1alpha1.ActionStatus{
        Name:      actionStatus.Name,
        Phase:     "Running",
        StartTime: &metav1.Time{Time: time.Now()},
    }

    if err != nil {
        rollbackStatus.Phase = "Failed"
        rollbackStatus.Message = fmt.Sprintf("Failed: %v", err)
        return rollbackStatus, err
    }

    if actionStatus.Phase != "Succeeded" {
        rollbackStatus.Phase = "Skipped"
        return rollbackStatus, nil
    }

    rollbackStatus.Phase = "Succeeded"
    return rollbackStatus, nil
}
```

### After（使用常量）

```go
func (e *LocalizationActionExecutor) Rollback(...) (*drv1alpha1.ActionStatus, error) {
    rollbackStatus := &drv1alpha1.ActionStatus{
        Name:      actionStatus.Name,
        Phase:     drv1alpha1.PhaseRunning,
        StartTime: &metav1.Time{Time: time.Now()},
    }

    if err != nil {
        rollbackStatus.Phase = drv1alpha1.PhaseFailed
        rollbackStatus.Message = fmt.Sprintf("Failed: %v", err)
        return rollbackStatus, err
    }

    if actionStatus.Phase != drv1alpha1.PhaseSucceeded {
        rollbackStatus.Phase = drv1alpha1.PhaseSkipped
        return rollbackStatus, nil
    }

    rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
    return rollbackStatus, nil
}
```

---

## 完成清单

- [ ] 创建常量文件 `api/v1alpha1/constants.go`（✅ 已完成）
- [ ] 替换所有 Phase 赋值语句
- [ ] 替换所有 Phase 比较语句
- [ ] 替换结构体初始化中的字面量
- [ ] 替换 OperationType 相关字面量
- [ ] 替换 FailurePolicy 相关字面量
- [ ] 运行编译测试
- [ ] 运行单元测试
- [ ] 运行静态检查
- [ ] Code Review
- [ ] 提交代码

---

## 注意事项

1. **测试重要性**: 迁移后务必运行完整的测试套件
2. **一次迁移一个文件**: 降低出错风险
3. **提交频率**: 建议每迁移一个文件就提交一次
4. **向后兼容**: 常量值必须与原字符串完全一致
5. **Code Review**: 让团队成员审查迁移结果

---

## 获取帮助

- **Cursor Rule**: `.cursor/rules/status-constants.mdc` 提供了详细的规范和示例
- **常量定义**: `api/v1alpha1/constants.go` 包含所有可用常量
- **问题反馈**: 如遇到问题，请在团队中讨论

---

**预计迁移时间**: 1-2 小时（包括测试和验证）  
**优先级**: 高（提升代码质量，减少潜在错误）  
**风险等级**: 低（纯粹的重构，不改变功能）
