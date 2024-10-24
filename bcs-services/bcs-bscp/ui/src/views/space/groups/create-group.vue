<template>
  <bk-dialog
    :title="t('新增分组')"
    ext-cls="create-group-dialog"
    :confirm-text="t('提交')"
    :cancel-text="t('取消')"
    :width="640"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @closed="handleClose"
    @confirm="handleConfirm">
    <group-edit-form ref="groupFormRef" :group="groupData" @change="updateData"></group-edit-form>
    <template #footer>
      <bk-button
        theme="primary"
        @click="handleConfirm"
        :disabled="pending"
        :loading="pending"
        style="margin-right: 8px">
        {{ t('提交') }}
      </bk-button>
      <bk-button @click="handleClose">{{ t('取消') }}</bk-button>
    </template>
  </bk-dialog>
</template>
<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import { IGroupEditing, ECategoryType } from '../../../../types/group';
  import { createGroup } from '../../../api/group';
  import groupEditForm from './components/group-edit-form.vue';
  import Message from 'bkui-vue/lib/message';

  const route = useRoute();
  const { t } = useI18n();

  const props = defineProps<{
    show: boolean;
  }>();

  const emits = defineEmits(['update:show', 'reload']);

  const groupData = ref<IGroupEditing>({
    name: '',
    public: true,
    bind_apps: [],
    rule_logic: 'AND',
    rules: [{ key: '', op: 'eq', value: '' }],
  });
  const groupFormRef = ref();
  const pending = ref(false);

  watch(
    () => props.show,
    (val) => {
      if (val) {
        groupData.value = {
          name: '',
          public: true,
          bind_apps: [],
          rule_logic: 'AND',
          rules: [{ key: '', op: 'eq', value: '' }],
        };
      }
    },
  );

  const updateData = (data: IGroupEditing) => {
    groupData.value = data;
  };

  const handleConfirm = async () => {
    const result = await groupFormRef.value.validate();
    if (!result) {
      return;
    }
    try {
      pending.value = true;
      const { name, public: isPublic, bind_apps, rule_logic, rules } = groupData.value;
      const params = {
        biz_id: route.params.spaceId,
        name,
        public: isPublic,
        bind_apps: isPublic ? [] : bind_apps,
        mode: ECategoryType.Custom,
        selector: rule_logic === 'AND' ? { labels_and: rules } : { labels_or: rules },
      };
      const res = await createGroup(route.params.spaceId as string, params);
      Message({
        message: t('创建分组成功'),
        theme: 'success',
      });
      handleClose();
      emits('reload', res.id);
    } catch (e) {
      console.error(e);
    } finally {
      setTimeout(() => {
        pending.value = false;
      }, 300);
    }
  };

  const handleClose = () => {
    emits('update:show', false);
  };
</script>
<style lang="scss">
  .create-group-dialog.bk-modal-wrapper {
    .bk-dialog-header {
      padding-top: 16px;
    }
    .bk-modal-content {
      max-height: 386px;
      overflow: auto;
    }
  }
</style>
