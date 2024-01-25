<template>
  <bk-dialog
    ext-cls="create-version-confirm-dialog"
    :title="t('确认更新配置文件版本？')"
    header-align="center"
    footer-align="center"
    :width="400"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    @closed="close">
    <p class="tips">{{ t('以下套餐及服务未命名版本中引用的此配置文件也将更新') }}</p>
    <div class="service-table">
      <bk-loading style="min-height: 100px" :loading="loading">
      <bk-table :data="citedList" :max-height="maxTableHeight">
        <bk-table-column :label="t('所在套餐')" prop="template_set_name"></bk-table-column>
        <bk-table-column :label="t('引用此模板的服务')">
          <template #default="{ row }">
            <div v-if="row.app_id" class="app-info" @click="goToConfigPageImport(row.app_id)">
              <div v-overflow-title class="name-text">{{ row.app_name }}</div>
              <LinkToApp class="link-icon" :id="row.app_id" auto-jump />
            </div>
          </template>
        </bk-table-column>
      </bk-table>
    </bk-loading>
    </div>
    <template #footer>
      <div class="actions-wrapper">
        <bk-button theme="primary" :loading="pending" @click="emits('confirm')">{{ t('确定') }}</bk-button>
        <bk-button @click="close">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import { IPackagesCitedByApps } from '../../../../../../types/template';
import { getUnNamedVersionAppsBoundByLatestTemplateVersion } from '../../../../../api/template';
import LinkToApp from '../../list/components/link-to-app.vue';
import useTemplateStore from '../../../../../store/template';

const { currentTemplateSpace } = storeToRefs(useTemplateStore());
const { t } = useI18n();

const props = defineProps<{
  show: boolean;
  spaceId: string;
  templateSpaceId: number;
  templateId: number;
  versionId: number;
  pending: boolean;
}>();

const emits = defineEmits(['update:show', 'confirm']);

const router = useRouter();

const loading = ref(false);
const citedList = ref<IPackagesCitedByApps[]>([]);

const maxTableHeight = computed(() => {
  const windowHeight = window.innerHeight;
  return windowHeight * 0.6 - 200;
});

watch(
  () => props.show,
  (val) => {
    if (val) {
      getCitedData();
    }
  },
);

const goToConfigPageImport = (id: number) => {
  const { href } = router.resolve({
    name: 'service-config',
    params: { appId: id },
    query: { pkg_id: currentTemplateSpace.value },
  });
  window.open(href, '_blank');
};


const getCitedData = async () => {
  loading.value = true;
  const params = {
    start: 0,
    all: true,
  };
  const res = await getUnNamedVersionAppsBoundByLatestTemplateVersion(
    props.spaceId,
    props.templateSpaceId,
    props.templateId,
    params,
  );
  citedList.value = res.details;
  loading.value = false;
};

const close = () => {
  emits('update:show', false);
};

defineExpose({
  close,
});
</script>
<style lang="scss" scoped>
.app-info {
  display: flex;
  align-items: center;
  overflow: hidden;
  cursor: pointer;
  .name-text {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  .link-icon {
    flex-shrink: 0;
    margin-left: 10px;
  }
}
.actions-wrapper {
  padding-bottom: 20px;
  .bk-button:not(:last-of-type) {
    margin-right: 8px;
  }
}
</style>
<style lang="scss">
.create-version-confirm-dialog.bk-modal-wrapper.bk-dialog-wrapper {
  .bk-modal-footer {
    position: static;
    padding: 32px 0 48px;
    background: #ffffff;
    border-top: none;
    .bk-button {
      min-width: 88px;
    }
  }
  .bk-modal-body {
    padding: 0;
  }
}
</style>
