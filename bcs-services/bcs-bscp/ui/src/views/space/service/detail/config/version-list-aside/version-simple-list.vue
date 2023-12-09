<template>
  <section class="version-container">
    <bk-loading :loading="versionListLoading">
      <div class="version-search-wrapper">
        <SearchInput
          v-model="searchStr"
          class="config-search-input"
          placeholder="版本名称"/>
      </div>
      <section class="versions-wrapper">
        <section
          v-for="version in versionsInView"
          :key="version.id"
          :class="['version-item', { active: versionData.id === version.id }]"
          @click="handleSelectVersion(version)">
          <div :class="['dot', version.status.publish_status]"></div>
          <div class="version-name">{{ version.spec.name }}</div>
          <bk-popover
            v-if="version.status.publish_status !== 'editing'"
            theme="light config-version-actions-popover"
            placement="bottom-end"
            :popover-delay="[0, 100]"
            :arrow="false">
            <Ellipsis class="action-more-icon" />
            <template #content>
              <div class="action-list">
                <div class="action-item" @click="handleDiffDialogShow(version)">版本对比</div>
                <!-- <bk-dropdown-item @click="handleDeprecate(version.id)">废弃</bk-dropdown-item> -->
              </div>
            </template>
          </bk-popover>
        </section>
        <TableEmpty v-if="searchStr && versionsInView.length === 0" :is-search-empty="true" @clear="searchStr = ''" />
      </section>
    </bk-loading>
    <VersionDiff v-model:show="showDiffPanel" :current-version="diffVersion" />
  </section>
</template>
<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import useConfigStore from '../../../../../../store/config';
import { Ellipsis } from 'bkui-vue/lib/icon';
// import { InfoBox } from 'bkui-vue/lib';
import { getConfigVersionList } from '../../../../../../api/config';
import { GET_UNNAMED_VERSION_DATA } from '../../../../../../constants/config';
import { IConfigVersion } from '../../../../../../../types/config';
import SearchInput from '../../../../../../components/search-input.vue';
import TableEmpty from '../../../../../../components/table/table-empty.vue';
import VersionDiff from '../../config/components/version-diff/index.vue';

const configStore = useConfigStore();
const { versionData, refreshVersionListFlag } = storeToRefs(configStore);

const route = useRoute();
const router = useRouter();

const props = defineProps<{
  bkBizId: string;
  appId: number;
}>();

const unNamedVersion: IConfigVersion = GET_UNNAMED_VERSION_DATA();
const versionListLoading = ref(false);
const versionList = ref<IConfigVersion[]>([]);
const searchStr = ref('');
const showDiffPanel = ref(false);
const diffVersion = ref();

const versionsInView = computed(() => {
  if (searchStr.value === '') {
    return versionList.value;
  }
  return versionList.value.filter(item => item.spec.name.toLowerCase().includes(searchStr.value.toLocaleLowerCase()));
});

// 监听刷新版本列表标识，处理新增版本场景，默认选中新增的版本
watch(refreshVersionListFlag, async (val) => {
  if (val) {
    await getVersionList();
    const versionDetail = versionList.value[1];
    if (versionDetail) {
      versionData.value = versionDetail;
      refreshVersionListFlag.value = false;
    }
  }
});

watch(
  () => props.appId,
  () => {
    getVersionList();
  },
);

onMounted(async () => {
  init();
});

const init = async () => {
  await getVersionList();
  if (route.params.versionId) {
    const version = versionList.value.find(item => item.id === Number(route.params.versionId));
    if (version) {
      versionData.value = version;
    }
  }
};

const getVersionList = async () => {
  try {
    versionListLoading.value = true;
    const params = {
      // 未命名版本不在实际的版本列表里，需要特殊处理
      start: 0,
      all: true,
    };
    const res = await getConfigVersionList(props.bkBizId, props.appId, params);
    versionList.value = [unNamedVersion, ...res.data.details];
  } catch (e) {
    console.error(e);
  } finally {
    versionListLoading.value = false;
  }
};

const handleSelectVersion = (version: IConfigVersion) => {
  versionData.value = version;
  const params: { spaceId: string, appId: number, versionId?: number } = {
    spaceId: props.bkBizId,
    appId: props.appId,
  };
  if (version.id !== 0) {
    params.versionId = version.id;
  }
  router.push({ name: 'service-config', params });
};

const handleDiffDialogShow = (version: IConfigVersion) => {
  diffVersion.value = version;
  showDiffPanel.value = true;
};

// const handleDeprecate = (id: number) => {
//   InfoBox({
//     title: '确认废弃此版本？',
//     subTitle: '废弃操作无法撤回，请谨慎操作！',
//     headerAlign: 'center' as const,
//     footerAlign: 'center' as const,
//     onConfirm: () => {},
//   } as any);
// };
</script>

<style lang="scss" scoped>
.version-container {
  height: 100%;
}
.bk-nested-loading {
  height: 100%;
}
.version-search-wrapper {
  padding: 8px 16px;
}
.versions-wrapper {
  height: calc(100% - 48px);
  overflow: auto;
}
.version-steps {
  padding: 16px 0;
  overflow: auto;
}
.version-item {
  position: relative;
  padding: 0 40px 0 48px;
  cursor: pointer;
  &.active {
    background: #e1ecff;
  }
  &:hover {
    background: #e1ecff;
  }
  .dot {
    position: absolute;
    left: 28px;
    top: 16px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    border: 1px solid #c4c6cc;
    background: #f0f1f5;
    &.not_released {
      border: 1px solid #ff9c01;
      background: #ffe8c3;
    }
    &.full_released,
    &.partial_released {
      border: 1px solid #3fc06d;
      background: #e5f6ea;
    }
  }
}
.version-name {
  height: 40px;
  line-height: 40px;
  font-size: 12px;
  color: #313238;
  text-align: left;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.action-more-icon {
  position: absolute;
  top: 10px;
  right: 10px;
  transform: rotate(90deg);
  width: 22px;
  height: 22px;
  color: #979ba5;
  border-radius: 50%;
  cursor: pointer;
  &:hover {
    background: rgba(99, 101, 110, 0.1);
    color: #3a84ff;
  }
}
.list-pagination {
  margin-top: 16px;
}
</style>
<style lang="scss">
.config-version-actions-popover.bk-popover.bk-pop2-content {
  padding: 4px 0;
  border: 1px solid #dcdee5;
  box-shadow: 0 2px 6px 0 #0000001a;
  .action-list {
    .action-item {
      padding: 0 12px;
      min-width: 58px;
      height: 32px;
      line-height: 32px;
      color: #63656e;
      font-size: 12px;
      cursor: pointer;
      &:hover {
        background: #f5f7fa;
      }
    }
  }
}
</style>
