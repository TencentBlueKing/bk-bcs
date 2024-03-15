<template>
  <div v-bkloading="{ loading: groupListLoading || versionListLoading, opacity: 1 }" class="select-group-wrapper">
    <template v-if="!groupListLoading && !versionListLoading">
      <div class="group-tree-area">
        <Group
          :group-list="groupList"
          :version-list="versionList"
          :released-groups="props.releasedGroups"
          :release-type="releaseType"
          :disable-select="props.disableSelect"
          :value="props.groups"
          @release-type-change="emits('releaseTypeChange', $event)"
          @change="emits('change', $event)" />
      </div>
      <div class="preview-area">
        <Preview
          :group-list="groupList"
          :release-type="releaseType"
          :released-groups="props.releasedGroups"
          :value="props.groups"
          @diff="emits('openPreviewVersionDiff', $event)"
          @change="emits('change', $event)" />
      </div>
    </template>
  </div>
</template>
<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { storeToRefs } from 'pinia';
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

  const props = withDefaults(
    defineProps<{
      releaseType?: string;
      releasedGroups?: number[];
      groups: IGroupToPublish[];
      disableSelect?: boolean; // 是否隐藏【选择分组实例上线】方式
    }>(),
    {
      releaseType: 'select',
      disableSelect: false,
    },
  );
  const emits = defineEmits(['openPreviewVersionDiff', 'releaseTypeChange', 'change']);

  const groupListLoading = ref(true);
  const groupList = ref<IGroupToPublish[]>([]);
  const versionListLoading = ref(true);
  const versionList = ref<IConfigVersion[]>([]);

  onMounted(() => {
    getAllGroupData();
    getAllVersionData();
  });

  // 获取所有分组，并组装tree组件节点需要的数据
  const getAllGroupData = async () => {
    groupListLoading.value = true;
    const res = await getServiceGroupList(spaceId.value, appData.value.id as number);
    groupList.value = res.details.map((group: IGroupItemInService) => {
      const { group_id, group_name, release_id, release_name } = group;
      const selector = group.new_selector;
      const rules = selector.labels_and || selector.labels_or || [];
      return { id: group_id, name: group_name, release_id, release_name, rules };
    });

    groupListLoading.value = false;
  };

  // 加载全量版本列表
  const getAllVersionData = async () => {
    versionListLoading.value = true;
    const res = await getConfigVersionList(spaceId.value, Number(appData.value.id), { start: 0, all: true });
    // 只需要已上线版本
    versionList.value = res.data.details.filter((item: IConfigVersion) => {
      return item.status.publish_status !== 'not_released';
    });
    versionListLoading.value = false;
  };
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
