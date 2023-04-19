<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { useRoute } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { cloneDeep } from 'lodash'
  import { useUserStore } from '../../../store/user'
  import { IGroupEditing, ECategoryType, EGroupRuleType, ICategoryItem, IGroupEditArg, IGroupRuleItem, IAllCategoryGroupItem } from '../../../../types/group'
  import { GROUP_RULE_OPS } from '../../../constants'
  import { getAppList } from '../../../api/index'
  import { IAppItem } from '../../../../types/app'

  const getDefaultRuleConfig = (): IGroupRuleItem => {
    return { key: '', op: <EGroupRuleType>'', value: '' }
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

  onMounted(() => {
    getServiceList()
  })

  const getServiceList = async () => {
    serviceLoading.value = true;
    try {
      const bizId = <string>route.params.spaceId
      const query = {
        start: 0,
        limit: 100,
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
  }

  const change = () => {
    emits('change', formData.value)
  }

  const validate = () => {
    return formRef.value.validate()
  }

  defineExpose({
    validate
  })

</script>
<template>
  <bk-form form-type="vertical" ref="formRef" :model="formData" :rules="rules">
    <bk-form-item label="分组名称" required property="name">
      <bk-input v-model="formData.name" size="small" placeholder="请输入分组名称" @blur="change"></bk-input>
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
        size="small"
        placeholder="请选择服务"
        @change="change">
        <bk-option v-for="service in serviceList" :key="service.id" :label="service.spec.name" :value="service.id"></bk-option>
      </bk-select>
    </bk-form-item>
    <bk-form-item class="radio-group-form" label="分组规则" required property="rule_logic"> 
      <bk-radio-group v-model="formData.rule_logic" @change="change">
        <bk-radio label="AND">AND</bk-radio>
        <bk-radio label="OR">OR</bk-radio>
      </bk-radio-group>
      <div v-for="(rule, index) in formData.rules" class="rule-config" :key="index">
        <bk-input v-model="rule.key" style="width: 176px;" size="small" placeholder="" @change="change"></bk-input>
        <bk-select v-model="rule.op" style="width: 72px;" size="small" :clearable="false" @change="change">
          <bk-option v-for="op in GROUP_RULE_OPS" :key="op.id" :value="op.id" :label="op.name"></bk-option>
        </bk-select>
        <bk-input
          v-model="rule.value"
          style="width: 280px;"
          size="small"
          placeholder=""
          :type="['gt', 'ge', 'lt', 'le'].includes(rule.op) ? 'number' : 'text'"
          @change="change">
        </bk-input>
        <div class="action-btns">
          <i v-if="index > 0 || formData.rules.length > 1" class="bk-bscp-icon icon-reduce" @click="handleDeleteRule(index)"></i>
          <i v-if="index === formData.rules.length - 1" style="margin-left: 10px;" class="bk-bscp-icon icon-add" @click="handleAddRule(index)"></i>
        </div>
      </div>
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