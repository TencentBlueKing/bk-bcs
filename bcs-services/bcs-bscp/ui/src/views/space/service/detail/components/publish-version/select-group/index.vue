<template>
  <div v-bkloading="{ loading: groupListLoading, opacity: 1 }" class="select-group-wrapper">
    <div class="group-tree-area">
      <Group
        v-if="!groupListLoading"
        ref="groupRef"
        :group-list="groupList"
        :group-list-loading="groupListLoading"
        :version-list="versionList"
        :version-list-loading="versionListLoading"
        :version-status="props.versionStatus"
        :released-groups="props.releasedGroups"
        :release-type="releaseType"
        :released-id="props.releaseId"
        :value="props.groups"
        @release-type-change="emits('releaseTypeChange', $event)"
        @change="emits('change', $event)" />
    </div>
    <div class="preview-area">
      <Preview
        :group-list="groupList"
        :group-list-loading="groupListLoading"
        :release-type="releaseType"
        :version-list="versionList"
        :version-list-loading="versionListLoading"
        :is-default-group-released="isDefaultGroupReleased"
        :released-groups="props.releasedGroups"
        :value="props.groups"
        @diff="emits('openPreviewVersionDiff', $event)"
        @change="emits('change', $event)" />
    </div>
  </div>
</template>
<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import Message from 'bkui-vue/lib/message';
  import useGlobalStore from '../../../../../../../store/global';
  import useServiceStore from '../../../../../../../store/service';
  import { IConfigVersion } from '../../../../../../../../types/config';
  import { getServiceGroupList } from '../../../../../../../api/group';
  import { getConfigVersionList } from '../../../../../../../api/config';
  import { IGroupToPublish, IGroupItemInService } from '../../../../../../../../types/group';
  import Group from './group.vue';
  import Preview from './preview.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const { appData } = storeToRefs(useServiceStore());
  const { t } = useI18n();

  const props = withDefaults(
    defineProps<{
      releaseType?: string;
      groups: IGroupToPublish[];
      versionStatus: string;
      releaseId: number;
      releasedGroups?: number[];
    }>(),
    {
      releaseType: 'select',
      releaseId: 0,
    },
  );
  const emits = defineEmits(['openPreviewVersionDiff', 'releaseTypeChange', 'change']);

  const groupListLoading = ref(true);
  const groupList = ref<IGroupToPublish[]>([]);
  const isDefaultGroupReleased = ref(false); // 默认分组是否已上线
  const versionListLoading = ref(true);
  const versionList = ref<IConfigVersion[]>([]);
  const groupRef = ref();

  onMounted(() => {
    getAllGroupData();
    getAllVersionData();
  });

  // 获取所有分组，并转化为tree组件需要的结构
  const getAllGroupData = async () => {
    groupListLoading.value = true;
    const res = await getServiceGroupList(spaceId.value, appData.value.id as number);
    groupList.value = res.details.map((group: IGroupItemInService) => {
      const { group_id, group_name, release_id, release_name } = group;
      const selector = group.new_selector;
      const rules = selector.labels_and || selector.labels_or || [];
      return { id: group_id, name: group_name, release_id, release_name, rules };
    });
    const defaultGroup = groupList.value.find((group) => group.id === 0);
    if (defaultGroup) {
      isDefaultGroupReleased.value = defaultGroup.release_id > 0;
    }
    groupListLoading.value = false;
  };

  // 加载全量版本列表
  const getAllVersionData = async () => {
    versionListLoading.value = true;
    const params = {
      start: 0,
      limit: 1000,
    };
    const res = await getConfigVersionList(spaceId.value, Number(appData.value.id), params);
    // 只需要已上线版本，且版本中不包含默认分组
    versionList.value = res.data.details.filter((item: IConfigVersion) => {
      const { publish_status, released_groups } = item.status;
      return publish_status !== 'not_released' && released_groups.findIndex((group) => group.id === 0) === -1;
    });
    versionListLoading.value = false;
  };

  const validate = () => {
    if (props.releaseType === 'exclude' && groupRef.value.selectedGroup.length === 0) {
      Message({
        theme: 'error',
        message: t('请至少选择一个排除分组实例'),
      });
      return false;
    }
    return true;
  };

  defineExpose({
    validate,
  });
</script>
<style lang="scss" scoped>
  .select-group-wrapper {
    display: flex;
    align-items: center;
    min-width: 1366px;
    height: 100%;
    background: #ffffff;
  }
  .group-tree-area {
    padding: 24px;
    width: 566px;
    height: 100%;
    border-right: 1px solid #dcdee5;
  }
  .preview-area {
    flex: 1;
    padding: 24px 0;
    height: 100%;
    background: #f5f7fa;
  }
</style>
