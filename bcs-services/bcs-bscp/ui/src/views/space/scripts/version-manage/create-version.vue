<template>
  <bk-popover
    ext-cls="create-tips-popover"
    theme="light"
    trigger="click"
    placement="bottom-start"
    :disabled="props.creatable"
    :is-show="popoverShow"
    @after-hidden="closePopover">
    <bk-button theme="primary" :disabled="props.disabled" @click="handleCreateClick">
      <Plus class="button-icon" />
      {{ t('新建版本') }}
    </bk-button>
    <template #content>
      <h3 class="tips">{{ t('当前已有「未上线」版本') }}</h3>
      <div class="actions">
        <bk-button theme="primary" size="small" @click="handleEditClick">{{ t('前往编辑') }}</bk-button>
        <bk-button size="small" @click="closePopover">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-popover>
  <bk-dialog
    :title="t('创建版本')"
    :confirm-text="t('创建')"
    :cancel-text="t('取消')"
    head-align="left"
    footer-align="right"
    width="480"
    :is-show="dialogShow"
    :is-loading="listLoading"
    @value-change="afterDialogShow"
    @confirm="handleLoadScript"
    @closed="dialogShow = false">
    <bk-form ref="formRef" form-type="vertical" :model="{ selectedScript }">
      <bk-form-item :label="t('选择载入版本')" required property="selectedScript">
        <bk-select
          v-model="selectedScript"
          :loading="listLoading"
          :clearable="false"
          :filterable="true"
          :list="list"
          id-key="id"
          display-key="name"
          :input-search="false">
          <template #optionRender="{ item }">
            <span>{{ item.name }}</span>
            <bk-tag :theme="item.isOnline ? 'success' : ''" style="margin-left: 8px">
              {{ item.isOnline ? t('已上线') : t('已下线') }}
            </bk-tag>
          </template>
        </bk-select>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>
<script setup lang="ts">
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Plus } from 'bkui-vue/lib/icon';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../../store/global';
  import { IScriptVersionListItem, IScriptMapItem } from '../../../../../types/script';
  import { getScriptVersionList } from '../../../../api/script';

  const { spaceId } = storeToRefs(useGlobalStore());
  const { t } = useI18n();

  const props = withDefaults(
    defineProps<{
      disabled: boolean;
      creatable?: boolean; // 是否编辑当前未上线版本
      scriptId: number;
    }>(),
    {
      creatable: false,
    },
  );

  const emits = defineEmits(['create', 'edit']);

  const popoverShow = ref(false);
  const dialogShow = ref(false);
  const list = ref<IScriptMapItem[]>([]);
  const listLoading = ref(false);
  const selectedScript = ref<number | string>('');
  const formRef = ref();

  const afterDialogShow = async (val: boolean) => {
    if (val) {
      selectedScript.value = '';
      listLoading.value = true;
      const res = await getScriptVersionList(spaceId.value, props.scriptId, { start: 0, all: true });
      list.value = res.details.map((item: IScriptVersionListItem) => {
        const { id, spec } = item.hook_revision;
        const name = spec.memo ? `${spec.name}(${spec.memo})` : spec.name;
        return { id, name, content: spec.content, isOnline: spec.state === 'deployed' };
      });
      listLoading.value = false;
      if (list.value.length > 0) {
        selectedScript.value = list.value[0].id;
      }
    }
  };

  const handleCreateClick = () => {
    if (!props.creatable) {
      setTimeout(() => {
        popoverShow.value = true;
      }, 100);
      return;
    }
    dialogShow.value = true;
  };

  const handleEditClick = () => {
    emits('edit');
    closePopover();
  };

  const handleLoadScript = async () => {
    await formRef.value.validate();
    const script = list.value.find((item) => item.id === selectedScript.value);
    if (script) {
      dialogShow.value = false;
      emits('create', script.content);
    }
  };

  const closePopover = () => {
    popoverShow.value = false;
  };
</script>
<style lang="scss" scoped>
  .button-icon {
    font-size: 18px;
  }
  .tips {
    margin: 0 0 16px;
    line-height: 24px;
    font-size: 16px;
    font-weight: normal;
    color: #313238;
  }
  .actions {
    text-align: right;
    .bk-button {
      margin-left: 8px;
    }
  }
</style>
<style lang="scss">
  .create-tips-popover.bk-popover {
    padding: 16px;
  }
</style>
