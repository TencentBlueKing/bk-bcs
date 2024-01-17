<template>
  <section class="create-version">
    <bk-button
      v-if="versionData.status.publish_status === 'not_released'"
      v-cursor="{ active: !props.hasPerm }"
      theme="primary"
      :class="['trigger-button', { 'bk-button-with-no-perm': !props.hasPerm }]"
      :disabled="props.permCheckLoading"
      @click="handleBtnClick"
    >
      {{ t('上线版本') }}
    </bk-button>
    <Teleport to="body">
      <VersionLayout v-if="isSelectGroupPanelOpen">
        <template #header>
          <section class="header-wrapper">
            <span class="header-name" @click="handlePanelClose">
              <ArrowsLeft class="arrow-left" />
              <span class="service-name">{{ appData.spec.name }}</span>
            </span>
            <AngleRight class="arrow-right" />
            {{ t('上线版本') }}：{{ versionData.spec.name }}
          </section>
        </template>
        <select-group
          :release-type="releaseType"
          :groups="groups"
          :version-status="versionData.status.publish_status"
          :release-id="versionData.id"
          @open-preview-version-diff="openPreviewVersionDiff"
          @release-type-change="releaseType = $event"
          @change="groups = $event"
        />
        <template #footer>
          <section class="actions-wrapper">
            <bk-button class="publish-btn" theme="primary" @click="handleDiffOrPublish">{{
              versionListByGroup.length ? t('对比并上线') : t('上线版本')
            }}</bk-button>
            <bk-button @click="handlePanelClose">{{ t('取消') }}</bk-button>
          </section>
        </template>
      </VersionLayout>
    </Teleport>
    <ConfirmDialog
      v-model:show="isConfirmDialogShow"
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      :release-id="versionData.id"
      :release-type="releaseType"
      :groups="groups"
      @confirm="handleConfirm"
    />
    <VersionDiff
      v-model:show="isDiffSliderShow"
      :current-version="versionData"
      :base-version-id="baseVersionId"
      :show-publish-btn="true"
      @publish="handleOpenPublishDialog"
      :version-diff-list="versionListByGroup"
    />
  </section>
</template>
<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { ArrowsLeft, AngleRight } from 'bkui-vue/lib/icon';
import { InfoBox } from 'bkui-vue';
import BkMessage from 'bkui-vue/lib/message';
import { storeToRefs } from 'pinia';
import useGlobalStore from '../../../../../../store/global';
import { IGroupToPublish } from '../../../../../../../types/group';
import useServiceStore from '../../../../../../store/service';
import useConfigStore from '../../../../../../store/config';
// import { permissionCheck } from '../../../../../../api/index';
import VersionLayout from '../../config/components/version-layout.vue';
import ConfirmDialog from './confirm-dialog.vue';
import SelectGroup from './select-group/index.vue';
import VersionDiff from '../../config/components/version-diff/index.vue';
import { useRoute,useRouter } from 'vue-router';
import { getConfigVersionList } from '../../../../../../api/config';
import { IConfigVersion } from '../../../../../../../types/config';

const { permissionQuery, showApplyPermDialog } = storeToRefs(useGlobalStore());
const serviceStore = useServiceStore();
const versionStore = useConfigStore();
const { appData } = storeToRefs(serviceStore);
const { versionData, publishedVersionId } = storeToRefs(versionStore);
const { t } = useI18n();

const props = defineProps<{
  bkBizId: string;
  appId: number;
  permCheckLoading: boolean;
  hasPerm: boolean;
}>();

const emit = defineEmits(['confirm']);

const route = useRoute();
const router = useRouter();
const bkBizId = String(route.params.spaceId);
const appId = Number(route.params.appId);
const versionList = ref<IConfigVersion[]>([]);
const isSelectGroupPanelOpen = ref(false);
const isDiffSliderShow = ref(false);
const isConfirmDialogShow = ref(false);
const releaseType = ref('select');
const groups = ref<IGroupToPublish[]>([]);
const baseVersionId = ref(0);

const permissionQueryResource = computed(() => [
  {
    biz_id: props.bkBizId,
    basic: {
      type: 'app',
      action: 'publish',
      resource_id: props.appId,
    },
  },
]);

// 获取所有选择上线范围的上线版本
const versionListByGroup = computed(() => {
  const list = [] as IConfigVersion[];
  versionList.value.forEach((version) => {
    groups.value.forEach((group) => {
      if (version.spec.name === group.release_name && !list.includes(version)) list.push(version);
    });
  });
  return list;
});

// 判断是否需要对比上线版本
const handleDiffOrPublish = () => {
  if (versionListByGroup.value.length) {
    baseVersionId.value = versionListByGroup.value[0].id;
    isDiffSliderShow.value = true;
    return;
  }
  handleOpenPublishDialog();
};

// 获取所有对比基准版本
const getVersionList = async () => {
  try {
    const res = await getConfigVersionList(bkBizId, appId, { start: 0, all: true });
    versionList.value = res.data.details.filter((item: IConfigVersion) => item.id !== versionData.value.id && item.status.publish_status === 'partial_released');
  } catch (e) {
    console.error(e);
  }
};

const handleBtnClick = () => {
  getVersionList();
  if (props.hasPerm) {
    isSelectGroupPanelOpen.value = true;
  } else {
    permissionQuery.value = { resources: permissionQueryResource.value };
    showApplyPermDialog.value = true;
  }
};

const handleOpenPublishDialog = () => {
  if (groups.value.length === 0) {
    BkMessage({ theme: 'error', message: t('请选择上线分组') });
    return;
  }
  isConfirmDialogShow.value = true;
};

// 选择分组面板上线预览版本对比
const openPreviewVersionDiff = (id: number) => {
  baseVersionId.value = id;
  isDiffSliderShow.value = true;
};

// 版本上线成功
const handleConfirm = (haveCredentials: boolean) => {
  isDiffSliderShow.value = false;
  publishedVersionId.value = versionData.value.id;
  handlePanelClose();
  emit('confirm');
  if (haveCredentials) {
    InfoBox({
      infoType: 'success',
      title: t('版本已上线'),
      dialogType: 'confirm',
    });
  } else {
    InfoBox({
      infoType: 'success',
      title: t('版本已上线'),
      confirmText: t('新增服务密钥'),
      cancelText: t('稍后再说'),
      onConfirm: () => {
        router.push({ name: 'credentials-management' });
      },
    });
  }
};

const handlePanelClose = () => {
  releaseType.value = 'select';
  isSelectGroupPanelOpen.value = false;
  groups.value = [];
};

defineExpose({
  openSelectGroupPanel: handleBtnClick,
});
</script>
<style lang="scss" scoped>
.trigger-button {
  margin-left: 8px;
}
.header-wrapper {
  display: flex;
  align-items: center;
  padding: 0 24px;
  height: 100%;
  font-size: 12px;
  line-height: 1;
}
.header-name {
  display: flex;
  align-items: center;
  font-size: 12px;
  color: #3a84ff;
  cursor: pointer;
}
.arrow-left {
  font-size: 26px;
  color: #3884ff;
}
.arrow-right {
  font-size: 24px;
  color: #c4c6cc;
}
.actions-wrapper {
  display: flex;
  align-items: center;
  padding: 0 24px;
  height: 100%;
  .publish-btn {
    margin-right: 8px;
  }
  .bk-button {
    min-width: 88px;
  }
}
.version-selector {
  display: flex;
  align-items: center;
  height: 100%;
  padding: 0 24px;
  font-size: 12px;
}
</style>
