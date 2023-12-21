<template>
  <DeleteConfirmDialog title="确认删除该配置文件？" :is-show="props.show" @confirm="handleConfirm" @close="close">
    <div style="margin-bottom: 8px">
      配置文件:
      <span style="color: #313238; font-weight: 600">{{ props.configs[0] ? props.configs[0].spec.name : '' }}</span>
    </div>
    <div>一旦删除，该操作将无法撤销，请谨慎操作</div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { storeToRefs } from 'pinia';
import Message from 'bkui-vue/lib/message';
import useGlobalStore from '../../../../../../../store/global';
import useTemplateStore from '../../../../../../../store/template';
import { ITemplateConfigItem } from '../../../../../../../../types/template';
import { deleteTemplate } from '../../../../../../../api/template';
import DeleteConfirmDialog from '../../../../../../../components/delete-confirm-dialog.vue';

const { spaceId } = storeToRefs(useGlobalStore());
const { currentTemplateSpace } = storeToRefs(useTemplateStore());

const props = defineProps<{
  show: boolean;
  configs: ITemplateConfigItem[];
}>();

const emits = defineEmits(['update:show', 'deleted']);

const pending = ref(false);

const handleConfirm = async () => {
  try {
    pending.value = true;
    const ids = props.configs.map(config => config.id);
    await deleteTemplate(spaceId.value, currentTemplateSpace.value, ids);
    close();
    emits('deleted');
    Message({
      theme: 'success',
      message: '删除配置文件成功',
    });
  } catch (e) {
    console.log(e);
  } finally {
    pending.value = false;
  }
};

const close = () => {
  emits('update:show', false);
};
</script>
