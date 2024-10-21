<template>
  <div class="service-detail-header">
    <section class="summary-wrapper">
      <div :class="['status-tag', versionData.spec.deprecated ? 'deprecated' : publishStatus]">
        {{ statusName }}
      </div>
      <div class="version-name" :title="versionData.spec.name">{{ versionData.spec.name }}</div>
      <InfoLine
        v-if="versionData.spec.memo"
        v-bk-tooltips="{
          content: versionData.spec.memo,
          placement: 'bottom-start',
          theme: 'light',
        }"
        class="version-desc" />
    </section>
    <template v-if="!props.versionDetailView">
      <div class="detail-header-tabs" v-if="isFileType">
        <BkTab type="unborder-card" v-model:active="activeTab" :label-height="41" @change="handleTabChange">
          <BkTabPanel v-for="tab in tabs" :key="tab.name" :name="tab.name" :label="tab.label"></BkTabPanel>
        </BkTab>
      </div>
      <section class="version-operations">
        <ReleasedGroupViewer
          v-if="isShowReleasedGroups"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :groups="versionData.status.released_groups"
          :disabled="publishStatus === 'full_released'">
          <div class="released-groups">
            <i class="bk-bscp-icon icon-resources-fill"></i>
            <div class="groups-tag">
              <div class="first-group-name">{{ firstReleasedGroupName }}</div>
              <div
                v-if="publishStatus === 'partial_released' && versionData.status.released_groups.length > 1"
                class="remaining-count">
                ;
                <span class="count">+{{ versionData.status.released_groups.length - 1 }}</span>
              </div>
            </div>
          </div>
        </ReleasedGroupViewer>
        <VersionApproveStatus ref="verAppStatus" @send-data="getVerApproveStatus" />
        <CreateVersion
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :perm-check-loading="permCheckLoading"
          :has-perm="perms.create"
          @confirm="handleVersionCreated" />
        <PublishVersion
          ref="publishVersionRef"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :perm-check-loading="permCheckLoading"
          :has-perm="perms.publish"
          :approve-data="approveData"
          @confirm="refreshVesionList" />
        <ModifyGroupPublish
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :perm-check-loading="permCheckLoading"
          :has-perm="perms.publish"
          :approve-data="approveData"
          @confirm="refreshVesionList" />
        <!-- 更多选项 -->
        <!-- <HeaderMoreOptions v-show="['partial_released', 'not_released'].includes(publishStatus)" /> -->
        <HeaderMoreOptions
          :approve-status="approveData.status"
          :creator="creator"
          @handle-undo="verAppStatus.loadStatus()" />
      </section>
    </template>
  </div>
</template>
<script setup lang="ts">
  import { ref, computed, watch, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute, useRouter } from 'vue-router';
  import { InfoLine } from 'bkui-vue/lib/icon';
  import { storeToRefs } from 'pinia';
  import useConfigStore from '../../../../../store/config';
  import useServiceStore from '../../../../../store/service';
  import { IConfigVersion } from '../../../../../../types/config';
  import { permissionCheck } from '../../../../../api/index';
  import ReleasedGroupViewer from '../config/components/released-group-viewer.vue';
  import PublishVersion from './publish-version/index.vue';
  import CreateVersion from './create-version/index.vue';
  import ModifyGroupPublish from './modify-group-publish.vue';
  import HeaderMoreOptions from './header-more-options.vue';
  import VersionApproveStatus from './version-approve-status.vue';

  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();

  const configStore = useConfigStore();
  const serviceStore = useServiceStore();
  const { versionData } = storeToRefs(configStore);
  const { isFileType } = storeToRefs(serviceStore);
  const permCheckLoading = ref(false);
  const perms = ref({
    create: false,
    publish: false,
  });
  const verAppStatus = ref();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    versionDetailView: Boolean;
  }>();

  const tabs = ref([
    { name: 'config', label: t('配置管理'), routeName: 'service-config' },
    { name: 'script', label: t('前/后置脚本'), routeName: 'init-script' },
  ]);

  const approveData = ref<{
    status: string;
    time: string;
    type: string;
  }>({
    status: '',
    time: '',
    type: '',
  });

  const creator = ref('');

  const getDefaultTab = () => {
    const tab = tabs.value.find((item) => item.routeName === route.name);
    return tab ? tab.name : 'config';
  };
  const activeTab = ref(getDefaultTab());
  const publishVersionRef = ref();

  const publishStatus = computed(() => versionData.value.status.publish_status);

  const statusName = computed(() => {
    if (versionData.value.spec.deprecated) {
      return t('已废弃');
    }

    if (publishStatus.value === 'editing') {
      return t('编辑中');
    }
    if (publishStatus.value === 'not_released') {
      return t('未上线');
    }
    return t('已上线');
  });

  // 是否需要展示版本分组信息
  const isShowReleasedGroups = computed(() =>
    ['partial_released', 'full_released'].includes(versionData.value.status.publish_status),
  );

  // 当前版本是否上线到默认分组
  const hasDefaultGroup = computed(
    () => versionData.value.status.released_groups.findIndex((item) => item.id === 0) > -1,
  );

  // 第一个分组名称
  const firstReleasedGroupName = computed(() => {
    if (isShowReleasedGroups.value) {
      if (versionData.value.status.publish_status === 'full_released') {
        return t('全部实例');
      }
      return hasDefaultGroup.value ? t('全部实例') : versionData.value.status.released_groups[0].name;
    }
    return '';
  });

  watch(
    () => route.name,
    () => {
      activeTab.value = getDefaultTab();
    },
  );

  watch(
    () => props.bkBizId,
    () => {
      getVersionPerms();
    },
  );

  onMounted(() => {
    getVersionPerms();
  });

  const getVerApproveStatus = (approveStatusData: any, creatorData: string) => {
    approveData.value = approveStatusData;
    creator.value = creatorData;
  };

  const getVersionPerms = async () => {
    permCheckLoading.value = true;
    const [createRes, publishRes] = await Promise.all([
      permissionCheck({
        resources: [
          {
            biz_id: props.bkBizId,
            basic: {
              type: 'app',
              action: 'generate_release',
              resource_id: props.appId,
            },
          },
        ],
      }),
      permissionCheck({
        resources: [
          {
            biz_id: props.bkBizId,
            basic: {
              type: 'app',
              action: 'publish',
              resource_id: props.appId,
            },
          },
        ],
      }),
    ]);
    perms.value.create = createRes.is_allowed;
    perms.value.publish = publishRes.is_allowed;
    permCheckLoading.value = false;
  };

  // 创建版本成功后，刷新版本列表，若选择同时上线，则打开选择分组面板
  const handleVersionCreated = (version: IConfigVersion, isPublish: boolean) => {
    refreshVesionList();
    if (isPublish && publishVersionRef.value) {
      versionData.value = version;
      publishVersionRef.value.openSelectGroupPanel();
    }
  };

  const refreshVesionList = () => {
    configStore.$patch((state) => {
      state.refreshVersionListFlag = true;
    });
  };

  const handleTabChange = (val: string) => {
    const tab = tabs.value.find((item) => item.name === val);
    if (tab) {
      router.push({ name: tab.routeName });
    }
  };
</script>
<style lang="scss" scoped>
  .service-detail-header {
    position: relative;
    display: flex;
    align-items: center;
    padding: 0 24px;
    height: 41px;
    border-bottom: 1px solid #dcdee5;
    z-index: 1;
    .summary-wrapper {
      display: flex;
      align-items: center;
      justify-content: space-between;
      .status-tag {
        margin-right: 8px;
        padding: 0 10px;
        height: 22px;
        line-height: 20px;
        font-size: 12px;
        color: #63656e;
        border: 1px solid rgba(151, 155, 165, 0.3);
        border-radius: 11px;
        &.deprecated {
          color: #ea3536;
          background-color: #feebea;
          border-color: #ea35364d;
        }
        &.not_released {
          color: #fe9000;
          background: #ffe8c3;
          border-color: rgba(254, 156, 0, 0.3);
        }
        &.full_released,
        &.partial_released {
          color: #14a568;
          background: #e4faf0;
          border-color: rgba(20, 165, 104, 0.3);
        }
      }
      .version-name {
        max-width: 220px;
        color: #63656e;
        font-size: 14px;
        font-weight: bold;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .version-desc {
        margin-left: 8px;
        font-size: 15px;
        color: #979ba5;
        cursor: pointer;
      }
    }
    .detail-header-tabs {
      position: absolute;
      top: 0;
      left: 50%;
      transform: translateX(-50%);
      :deep(.bk-tab-header) {
        border-bottom: none;
      }
      :deep(.bk-tab-content) {
        display: none;
      }
    }
    .version-operations {
      position: absolute;
      top: 0;
      right: 24px;
      display: flex;
      align-items: center;
      height: 100%;
      z-index: 10;
      .released-groups {
        display: flex;
        align-items: center;
        padding: 2px 8px;
        background: #f0f1f5;
        border-radius: 2px;
        cursor: pointer;
      }
      .icon-resources-fill {
        margin-right: 4px;
        font-size: 14px;
        color: #979ba5;
      }
      .groups-tag {
        display: flex;
        align-items: center;
        line-height: 18px;
        font-size: 12px;
        color: #63656e;
        .count {
          padding: 2px 4px;
          line-height: 1;
          background: #fafbfd;
          border-radius: 2px;
        }
      }
    }
  }
</style>
