<template>
  <div class="service-detail-header">
    <section class="summary-wrapper">
      <div :class="['status-tag', versionData.status.publish_status]">
        {{ VERSION_STATUS_MAP[versionData.status.publish_status as keyof typeof VERSION_STATUS_MAP] || '编辑中' }}
      </div>
      <div class="version-name" :title="versionData.spec.name">{{ versionData.spec.name }}</div>
      <InfoLine
        v-if="versionData.spec.memo"
        v-bk-tooltips="{
          content: versionData.spec.memo,
          placement: 'bottom-start',
          theme: 'light',
        }"
        class="version-desc"
      />
    </section>
    <template v-if="!props.versionDetailView">
      <div class="detail-header-tabs" v-if="isFileType">
        <BkTab type="unborder-card" v-model:active="activeTab" :label-height="41" @change="handleTabChange">
          <BkTabPanel v-for="tab in tabs" :key="tab.name" :name="tab.name" :label="tab.label"></BkTabPanel>
        </BkTab>
      </div>
      <section class="version-operations">
        <div
          v-if="isShowReleasedGroups"
          v-bk-tooltips="{
            disabled: versionData.status.publish_status !== 'partial_released',
            placement: 'bottom-end',
            theme: 'light',
            content: getReleasedGroupsPopoverContent(versionData.status.released_groups)
          }"
          class="released-groups">
          <i class="bk-bscp-icon icon-resources"></i>
          <div class="groups-tag">
            <div class="first-group-name">{{ firstReleasedGroupName }}</div>
            <div v-if="versionData.status.publish_status === 'partial_released'" class="remaining-count">
              ;
              <span class="count">+{{ versionData.status.released_groups.length - 1 }}</span>
            </div>
          </div>
        </div>
        <CreateVersion
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :perm-check-loading="permCheckLoading"
          :has-perm="perms.create"
          @confirm="handleVersionCreated"
        />
        <PublishVersion
          ref="publishVersionRef"
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :perm-check-loading="permCheckLoading"
          :has-perm="perms.publish"
          @confirm="refreshVesionList"
        />
        <ModifyGroupPublish
          :bk-biz-id="props.bkBizId"
          :app-id="props.appId"
          :perm-check-loading="permCheckLoading"
          :has-perm="perms.publish"
          @confirm="refreshVesionList"
        />
      </section>
    </template>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { InfoLine } from 'bkui-vue/lib/icon';
import { storeToRefs } from 'pinia';
import useConfigStore from '../../../../../store/config';
import useServiceStore from '../../../../../store/service';
import { VERSION_STATUS_MAP } from '../../../../../constants/config';
import { IConfigVersion } from '../../../../../../types/config';
import { permissionCheck } from '../../../../../api/index';
import getReleasedGroupsPopoverContent from '../config/components/get-released-groups-popover-content';
import PublishVersion from './publish-version/index.vue';
import CreateVersion from './create-version/index.vue';
import ModifyGroupPublish from './modify-group-publish.vue';

const route = useRoute();
const router = useRouter();

const configStore = useConfigStore();
const serviceStore = useServiceStore();
const { versionData } = storeToRefs(configStore);
const { isFileType } = storeToRefs(serviceStore);
const permCheckLoading = ref(false);
const perms = ref({
  create: false,
  publish: false,
});

const props = defineProps<{
  bkBizId: string;
  appId: number;
  versionDetailView: Boolean;
}>();

const tabs = ref([
  { name: 'config', label: '配置管理', routeName: 'service-config' },
  { name: 'script', label: '前/后置脚本', routeName: 'init-script' },
]);

const getDefaultTab = () => {
  const tab = tabs.value.find(item => item.routeName === route.name);
  return tab ? tab.name : 'config';
};
const activeTab = ref(getDefaultTab());
const publishVersionRef = ref();

const isShowReleasedGroups = computed(() => {
  return versionData.value.status?.released_groups.length > 0;
});

const firstReleasedGroupName = computed(() => {
  if (isShowReleasedGroups.value) {
    return versionData.value.status.publish_status === 'partial_released' ? versionData.value.status.released_groups[0].name : '全部实例';
  }
  return ''
})

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
  const tab = tabs.value.find(item => item.name === val);
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
  box-shadow: 0 3px 4px 0 #0000000a;
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
    top: 5px;
    right: 24px;
    display: flex;
    align-items: center;
    z-index: 10;
    .released-groups {
      display: flex;
      align-items: center;
      padding: 5px 8px;
      background: #F0F1F5;
      border-radius: 2px;
      cursor: pointer;
    }
    .icon-resources {
      margin-right: 4px;
      font-size: 14px;
      color: #979BA5;
    }
    .groups-tag {
      display: flex;
      align-items: center;
      line-height: 22px;
      font-size: 12px;
      color: #63656e;
      .count {
        padding: 4px;
        background: #FAFBFD;
        border-radius: 2px;
      }
    }
  }
}
</style>
