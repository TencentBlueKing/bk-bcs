<template>
  <bk-dialog
    title="新增分组"
    ext-cls="create-group-dialog"
    confirm-text="提交"
    :width="640"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @closed="handleClose"
    @confirm="handleConfirm"
  >
    <group-edit-form ref="groupFormRef" :group="groupData" @change="updateData"></group-edit-form>
  </bk-dialog>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { IGroupEditing, ECategoryType } from '../../../../types/group';
import { createGroup } from '../../../api/group';
import groupEditForm from './components/group-edit-form.vue';

const route = useRoute();

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
        rules: [{ key: '', op: '', value: '' }],
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
    const { name, public: isPublic, bind_apps, rule_logic, rules } = groupData.value;
    const params = {
      biz_id: route.params.spaceId,
      name,
      public: isPublic,
      bind_apps: isPublic ? [] : bind_apps,
      mode: ECategoryType.Custom,
      selector: rule_logic === 'AND' ? { labels_and: rules } : { labels_or: rules },
    };
    await createGroup(route.params.spaceId as string, params);
    handleClose();
    emits('reload');
  } catch (e) {
    console.error(e);
  } finally {
    pending.value = false;
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
