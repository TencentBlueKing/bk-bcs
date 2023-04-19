<script setup lang="ts">
  import { ref, computed, watch, nextTick } from 'vue'
  import { cloneDeep } from 'lodash'
  import { IGroupEditing, ECategoryType, EGroupRuleType, ICategoryItem, IGroupEditArg, IGroupRuleItem, IAllCategoryGroupItem } from '../../../../../types/group'
  import { createCategory, createGroup, updateGroup } from '../../../../api/group'
  import { GROUP_RULE_OPS } from '../../../../constants'
  
  const props = defineProps<{
    show: boolean,
    appId: number,
    categoryList: IAllCategoryGroupItem[],
    group: IGroupEditing
  }>()

  const emits = defineEmits(['update:show', 'refreshCategoryList'])

  const getDefaultRuleConfig = (): IGroupRuleItem => {
    return { key: '', op: <EGroupRuleType>'', value: '' }
  }

  const formRef = ref()
  const pending = ref(false)
  const createCategoryPending = ref(false)
  const mode = {
    custom: ECategoryType.Custom,
    debug: ECategoryType.Debug
  }
  const formData = ref<IGroupEditing>(cloneDeep(props.group));
  const rules = {
    name: [
      {
        validator: (value: string) => value.length < 128,
        message: '最大长度128个字符'
      },
      {
        validator: (value: string) => {
          if (value.length > 0) {
            return /^[\u4e00-\u9fa5a-zA-Z0-9][\u4e00-\u9fa5a-zA-Z0-9_\-]*[\u4e00-\u9fa5a-zA-Z0-9]?$/.test(value)
          }
          return true
        },
        message: '仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾'
      }
    ]
  }

  // 是否为编辑态
  const isEditMode = computed(() => {
    return 'id' in props.group
  })

  const title = computed(() => {
    return isEditMode.value ? '编辑分组' : '创建分组'
  })

  watch(() => props.show, (val) => {
    if (val) {
      formData.value = cloneDeep(props.group)
    }
  })

  // 选择分类，所选值为字符串类型时创建新分类
  const handleCategoryChange = async (val: string) => {
    if (typeof val === 'string') {
      try {
        createCategoryPending.value = true
        const res = await createCategory(props.appId, val)
        formData.value.group_category_id = res.id
        emits('refreshCategoryList')
      } catch (e) {
        formData.value.group_category_id = ''
      } finally {
        createCategoryPending.value = false
      }
    } else {
      formData.value.group_category_id = val
    }
  }

  // 增加规则
  const handleAddRule = (index: number) => {
    const rule = getDefaultRuleConfig()
    formData.value.rules.splice(index + 1, 0, rule)
  }

  // 删除规则
  const handleDeleteRule = (index: number) => {
    formData.value.rules.splice(index, 1)
  }

  // 切换规则与/或逻辑
  const handleToggleRuleLogic = () => {
    formData.value.rule_logic = formData.value.rule_logic === 'AND' ? 'OR' : 'AND'
  }

  const handleConfirm = async () => {
    await formRef.value.validate()
    const { name, group_category_id, mode, rules, rule_logic, uid } = formData.value
    const params: IGroupEditArg = { name }
    if (mode === ECategoryType.Custom) {
      params['selector'] = rule_logic === 'AND' ? { labels_and: rules } : { labels_or: rules }
    } else {
      params.uid = uid
    }
    try {
      pending.value = true
      // 编辑分组
      if (isEditMode.value) {
        params.id = <number>formData.value.id
        await updateGroup(props.appId, params.id, params)
      } else { // 创建分组
        params.group_category_id = <number>group_category_id
        params.mode = <ECategoryType>mode
        await createGroup(props.appId, params)
      }
    } catch (e) {
      console.error(e)
    } finally {
      pending.value = false
    }

    handleClose()
  }

  const handleClose = () => {
    emits('update:show', false)
    nextTick(() => {
      formData.value = {
        name: '',
        group_category_id: '',
        mode: ECategoryType.Custom,
        rule_logic: 'AND',
        rules: [getDefaultRuleConfig()],
        uid: ''
      }
    })
  }

</script>
<template>
  <bk-dialog
    ext-cls="create-group-dialog"
    :title="title"
    :is-show="props.show"
    :width="960"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @closed="handleClose"
    @confirm="handleConfirm">
      <bk-alert theme="info" title="同分类下的不同分组之间需要保证没有交集，否则客户端将产生配置错误的风险！" />
      <div class="config-wrapper">
        <bk-form class="form-content-area" ref="formRef" :label-width="100" :model="formData" :rules="rules">
          <bk-form-item label="分组名称" property="name" required>
            <bk-input v-model="formData.name" size="small" placeholder="请输入"></bk-input>
          </bk-form-item>
          <bk-form-item label="分组标签" property="group_category_id" required>
            <bk-select
              v-model="formData.group_category_id"
              placeholder="请选择"
              size="small"
              :allow-create="true"
              :filterable="true"
              :disabled="createCategoryPending"
              @change="handleCategoryChange">
              <bk-option
                v-for="category in categoryList"
                :key="category.group_category_id"
                :value="category.group_category_id"
                :label="category.group_category_name">
              </bk-option>
            </bk-select>
          </bk-form-item>
          <!-- <bk-form-item label="调试用分组" required>
            <bk-switcher
              v-model="formData.mode"
              theme="primary"
              :disabled="isEditMode"
              :true-value="mode.debug"
              :false-value="mode.custom">
            </bk-switcher>
            <p class="debug-tips">启用调试用分组后，仅可使用 UID 作为分组规则，且配置版本将不跟随主线</p>
          </bk-form-item> -->
          <bk-form-item v-if="formData.mode === ECategoryType.Custom" label="分组规则" required>
            <div v-for="(rule, index) in formData.rules" class="rule-config" :key="index">
              <div
                v-if="index > 0"
                v-bk-tooltips="'可同时将所有条件间关系切换为 AND/OR'"
                class="rule-logic"
                @click="handleToggleRuleLogic">
                {{ formData.rule_logic }}
              </div>
              <bk-input v-model="rule.key" style="width: 80px;" size="small" placeholder=""></bk-input>
              <bk-select v-model="rule.op" style="width: 72px;" size="small" :clearable="false">
                <bk-option v-for="op in GROUP_RULE_OPS" :key="op.id" :value="op.id" :label="op.name"></bk-option>
              </bk-select>
              <bk-input
                v-model="rule.value"
                style="width: 120px;"
                size="small"
                placeholder=""
                :type="['gt', 'ge', 'lt', 'le'].includes(rule.op) ? 'number' : 'text'">
              </bk-input>
              <div class="action-btns">
                <i v-if="index > 0 || formData.rules.length > 1" class="bk-bscp-icon icon-reduce" @click="handleDeleteRule(index)"></i>
                <i v-if="index === formData.rules.length - 1" style="margin-left: 10px;" class="bk-bscp-icon icon-add" @click="handleAddRule(index)"></i>
              </div>
            </div>
          </bk-form-item>
          <bk-form-item v-else label="UID" required property="uid">
            <bk-input v-model="formData.uid" placeholder="请输入"></bk-input>
          </bk-form-item>
        </bk-form>
        <div class="group-intersection-detect">
          <h4 class="title">分组交集检测</h4>
          <bk-table class="rule-table" :border="['outer']">
            <bk-table-column label="分组名称"></bk-table-column>
            <bk-table-column label="分组规则"></bk-table-column>
          </bk-table>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <bk-button theme="primary" :loading="pending" @click="handleConfirm">提交</bk-button>
          <bk-button :disabled="pending" @click="handleClose">取消</bk-button>
        </div>
      </template>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .config-wrapper {
    display: flex;
    padding: 24px 0 14px;
    height: 530px;
  }
  .form-content-area {
    padding: 24px 24px 24px 0;
    width: 50%;
    height: 100%;
    overflow: auto;
  }
  :deep(.bk-form-label) {
    font-size: 12px;
  }
  .debug-tips {
    margin: 0;
    font-size: 12px;
    line-height: 18px;
    color: #979ba5;
  }
  .rule-config {
    display: flex;
    align-items: center;
    justify-content: space-between;
    position: relative;
    .rule-logic {
      position: absolute;
      top: 3px;
      left: -48px;
      height: 26px;
      line-height: 26px;
      width: 40px;
      background: #e1ecff;
      color: #3a84ff;
      font-size: 12px;
      text-align: center;
      cursor: pointer;
    }
  }
  .action-btns {
    display: flex;
    align-items: center;
    width: 38px;
    font-size: 14px;
    color: #979ba5;
    cursor: pointer;
    i:hover {
      color: #3a84ff;
    }
  }
  .group-intersection-detect {
    padding: 16px;
    width: 50%;
    background: #f5f7fa;
    .title {
      margin: 0;
    }
    .rule-table {
      margin-top: 16px;
    }
  }
  .dialog-footer {
    .bk-button {
      margin-left: 8px;
    }
  }
</style>
<style lang="scss">
  .create-group-dialog.bk-dialog-wrapper {
    .bk-dialog-header {
      padding-bottom: 20px;
    }
    .bk-modal-footer {
      height: auto;
      padding: 8px 24px;
    }
  }
    
</style>
