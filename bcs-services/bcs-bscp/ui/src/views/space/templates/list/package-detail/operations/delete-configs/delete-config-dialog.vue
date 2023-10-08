<template>
  <bk-dialog
    ext-cls="delete-configs-dialog"
    confirm-text="确认删除"
    footer-align="center"
    :width="400"
    :show-header="false"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @confirm="handleConfirm"
    @closed="close"
  >
    <template #header>
      <div class="header-icon"><Warn /></div>
    </template>
    <div class="title-area">
      <template v-if="props.configs.length > 1">
        确认删除<span>{{ props.configs.length }}</span
        >条配置项?
      </template>
      <template v-else> 确认删除配置项【{{ props.configs[0] ? props.configs[0].spec.name : '' }}】？ </template>
    </div>
    <div class="tips">删除后不可找回，请谨慎操作。</div>
    <template #footer>
      <div class="actions-wrapper">
        <bk-button theme="primary" :loading="pending" @click="handleConfirm">确认删除</bk-button>
        <bk-button @click="close">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { storeToRefs } from 'pinia';
import { Warn } from 'bkui-vue/lib/icon';
import { Message } from 'bkui-vue';
import useGlobalStore from '../../../../../../../store/global';
import useTemplateStore from '../../../../../../../store/template';
import { ITemplateConfigItem } from '../../../../../../../../types/template';
import { deleteTemplate } from '../../../../../../../api/template';

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
      message: '删除配置项成功',
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
<style lang="scss" scoped>
.header-icon {
  line-height: 1;
  font-size: 48px;
  color: #ff9c01;
  text-align: center;
}
.title-area {
  margin: 8px 0;
  font-size: 20px;
  line-height: 32px;
  color: #313238;
  text-align: center;
  .num {
    color: #313238;
  }
}
.tips {
  color: #63656e;
  font-size: 14px;
  text-align: center;
}
.actions-wrapper {
  padding-bottom: 20px;
  .bk-button:not(:last-of-type) {
    margin-right: 8px;
  }
}
</style>
<style lang="scss">
.delete-configs-dialog.bk-modal-wrapper.bk-dialog-wrapper {
  .bk-dialog-header {
    padding-bottom: 0;
  }
  .bk-modal-footer {
    // padding: 32px 0 48px;
    background: #ffffff;
    border-top: none;
    .bk-button {
      min-width: 88px;
    }
  }
}
</style>
