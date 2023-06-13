<script setup lang="ts">
  import { ref, watch, onMounted } from 'vue'
  import { useRoute } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { cloneDeep } from 'lodash'
  import { useUserStore } from '../../../../store/user'
  import { IGroupEditing, EGroupRuleType, IGroupRuleItem, IGroupBindService } from '../../../../../types/group'
  import { GROUP_RULE_OPS } from '../../../../constants/group'
  import { getAppList } from '../../../../api/index'
  import { IAppItem } from '../../../../../types/app'

  const getDefaultRuleConfig = (): IGroupRuleItem => {
    return { key: '', op: '', value: '' }
  }

  const route = useRoute()
  const { userInfo } = storeToRefs(useUserStore())

  const props = defineProps<{
    group: IGroupEditing
  }>()

  const emits = defineEmits(['change'])

  const serviceLoading = ref(false)
  const serviceList = ref<IAppItem[]>([])
  const formData = ref(cloneDeep(props.group))
  const formRef = ref()
  const rulesValid = ref(true)

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
    ],
    public: [
      {
        validator: (val: boolean) => {
          if (!val && formData.value.bind_apps.length === 0) {
            return false
          }
          return true
        },
        message: '指定服务不能为空'
      }
    ]
  }

  watch(() => props.group, (val) => {
    formData.value = cloneDeep(val)
    getServiceList()
  })

  onMounted(() => {
    getServiceList()
  })

  const getServiceList = async () => {
    serviceLoading.value = true;
    try {
      const bizId = <string>route.params.spaceId
      const query = {
        start: 0,
        limit: 1000, // @todo 确认拉全量列表参数
        operator: userInfo.value.username
      }
      const resp = await getAppList(bizId, query)
      serviceList.value = resp.details
    } catch (e) {
      console.error(e)
    } finally {
        serviceLoading.value = false;
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
    change()
  }

  const handleLogicChange = (index: number, val: EGroupRuleType) => {
    validateRules()
    const rule = formData.value.rules[index]
    if (['in', 'nin'].includes(val) && !['in', 'nin'].includes(rule.op)) {
      rule.value = []
    } else if (!['in', 'nin'].includes(val) && ['in', 'nin'].includes(rule.op)) {
      rule.value = ''
    }
    rule.op = val
  }

  const change = () => {
    validateRules()
    emits('change', formData.value)
  }

  const validate = () => {
    const isRulesValid = validateRules()
    
    return formRef.value.validate().then(() => {
      return isRulesValid
    })
  }

  // 校验分组规则是否有表单项为空
  const validateRules = () => {
    const inValid = formData.value.rules.some(item => {
      const { key, op, value } = item
      return key === '' || op === '' || (Array.isArray(value) ? (<string[]>value).length === 0 : value === '')
    })

    rulesValid.value = !inValid

    return !inValid
  }

  defineExpose({
    validate
  })

</script>
<template>
  <bk-form form-type="vertical" ref="formRef" :model="formData" :rules="rules">
    <bk-form-item label="分组名称" required property="name">
      <bk-input v-model="formData.name" placeholder="请输入分组名称" @blur="change"></bk-input>
    </bk-form-item>
    <bk-form-item class="radio-group-form" label="服务可见范围" required property="public">
      <bk-radio-group v-model="formData.public" @change="change">
        <bk-radio :label="true">公开</bk-radio>
        <bk-radio :label="false">指定服务</bk-radio>
      </bk-radio-group>
      <bk-select
        v-if="!formData.public"
        v-model="formData.bind_apps"
        class="service-selector"
        multiple
        filterable
        placeholder="请选择服务"
        @change="change">
        <bk-option v-for="service in serviceList" :key="service.id" :label="service.spec.name" :value="service.id"></bk-option>
      </bk-select>
    </bk-form-item>
    <bk-form-item class="radio-group-form" label="分组规则" required property="rules"> 
      <div v-for="(rule, index) in formData.rules" class="rule-config" :key="index">
        <bk-input v-model="rule.key" style="width: 176px;" placeholder="" @change="change"></bk-input>
        <bk-select :model-value="rule.op" style="width: 82px;" :clearable="false" @change="handleLogicChange(index, $event)">
          <bk-option v-for="op in GROUP_RULE_OPS" :key="op.id" :value="op.id" :label="op.name"></bk-option>
        </bk-select>
        <div class="value-input">
          <bk-tag-input
            v-if="['in', 'nin'].includes(rule.op)"
            v-model="rule.value"
            :allow-create="true"
            :collapse-tags="true"
            :has-delete-icon="true"
            :show-clear-only-hover="true"
            :allow-auto-match="true"
            :list="[]"
            @change="change">
          </bk-tag-input>
          <bk-input
            v-else
            v-model="rule.value"
            placeholder=""
            :type="['gt', 'ge', 'lt', 'le'].includes(rule.op) ? 'number' : 'text'"
            @change="change">
          </bk-input>
        </div>
        <div class="action-btns">
          <i v-if="index > 0 || formData.rules.length > 1" class="bk-bscp-icon icon-reduce" @click="handleDeleteRule(index)"></i>
          <i v-if="index === formData.rules.length - 1" style="margin-left: 10px;" class="bk-bscp-icon icon-add" @click="handleAddRule(index)"></i>
        </div>
      </div>
      <div v-if="!rulesValid" class="bk-form-error">分组规则表单不能为空</div>
    </bk-form-item>
  </bk-form>
</template>
<style lang="scss" scoped>
  .bk-form {
    :deep(.bk-form-label) {
      font-size: 12px;
    }
    :deep(.radio-group-form .bk-form-content) {
      line-height: 1;
    }
  }
  .service-selector {
    margin-top: 10px;
  }
  .published-version {
    line-height: 16px;
    font-size: 12px;
    color: #313238;
  }
  .rule-config {
    display: flex;
    align-items: center;
    justify-content: space-between;
    position: relative;
    margin-top: 15px;
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
    .value-input {
      width: 270px;
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
</style>