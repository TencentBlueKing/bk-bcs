<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import { cloneDeep } from 'lodash'
  import { IGroupEditing, ECategoryType, EGroupRuleType, ICategoryItem, IGroupEditArg, IGroupRuleItem, IAllCategoryGroupItem, IGroupItem } from '../../../types/group'
  import { createGroup, updateGroup } from '../../api/group'
  import groupEditForm from './components/group-edit-form.vue';

  const route = useRoute()

  const props = defineProps<{
    show: boolean,
    group: IGroupItem
  }>()

  const emits = defineEmits(['update:show', 'reload'])
  
  const groupData = ref<IGroupEditing>({
    id: 0,
    name: '',
    public: true,
    bind_apps: [],
    rule_logic: 'AND',
    rules: [{ key: '', op: <EGroupRuleType>'', value: '' }]
  })
  const groupFormRef = ref()
  const pending = ref(false)

  watch(() => props.show, (val) => {
    const { id, name, public: isPublic, bind_apps, selector } = props.group
    groupData.value = {
      id,
      name,
      bind_apps,
      public: isPublic,
      rule_logic: selector.labels_and ? 'AND' : 'OR',
      rules: (selector.labels_and || selector.labels_or) as IGroupRuleItem[]
    }
  })

  const updateData = (data: IGroupEditing) => {
    groupData.value = data
  }

  const handleConfirm = async() => {
    await groupFormRef.value.validate()
    pending.value = true
    try {
      const { id, name, public: isPublic, bind_apps, rule_logic, rules } = groupData.value
      const params = {
        biz_id: route.params.spaceId,
        name,
        public: isPublic,
        bind_apps: isPublic ? [] : bind_apps,
        mode: ECategoryType.Custom,
        selector: rule_logic === 'AND' ? { labels_and: rules } : { labels_or: rules }
      }
      await updateGroup(<string>route.params.spaceId, <number>id, params)
      handleClose()
      emits('reload')
    } catch (e) {
      console.error(e)
    } finally {
      pending.value = false
    }
  }

  const handleClose = () => {
    emits('update:show', false)
  }

</script>
<template>
  <bk-dialog
    ext-cls="edit-group-dialog"
    confirm-text="提交"
    :width="952"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :close-icon="false"
    :is-loading="pending"
    @closed="handleClose"
    @confirm="handleConfirm">
    <div class="group-edit-content">
      <section class="group-form-wrapper">
        <div class="dialog-title">编辑分组</div>
        <div class="group-edit-form">
          <group-edit-form v-if="props.show" ref="groupFormRef" :group="groupData" @change="updateData"></group-edit-form>
        </div>
      </section>
      <section class="rule-detail-wrapper">
        <div class="rule-preview-title">分组规则预览</div>
        <div class="rule-list">
          <div class="rule-item rule-logic">逻辑关系：{{ groupData.rule_logic }}</div>
        </div>
      </section>
    </div>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .group-edit-content {
    display: flex;
    align-items: stretch;
    justify-content: space-between;
    .group-form-wrapper {
      width: 632px;
      border-right: 1px solid #dcdee5;
    }
    .group-edit-form {
      padding: 0 16px 0 24px;
      min-height: 268px;
      max-height: 386px;
      overflow: auto;
    }
    .rule-detail-wrapper {
      width: 320px;
      background: #f5f7fa;
    }
    .dialog-title {
      margin: 16px 0 24px;
      padding: 0 24px;
      font-size: 20px;
      line-height: 28px;
      color: #313238;
    }
    .rule-preview-title {
      padding: 12px 24px 16px;
      font-size: 14px;
      line-height: 22px;
      color: #313238;
    }
    .rule-list {
      padding: 0 24px;
      .rule-item {
        padding: 6px 8px;
        line-height: 20px;
        font-size: 12px;
        color: #63656e;
        background: #ffffff;
        &.rule-logic {
          background: #eaebf0;
        }
      }
    }
  }
</style>
<style lang="scss">
  .edit-group-dialog.bk-modal-wrapper {
    .bk-dialog-header {
      display: none;
    }
    .bk-modal-content {
      padding: 0;
    }
  }
</style>
