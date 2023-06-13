<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../../../../../../store/global'
  import { useServiceStore } from '../../../../../../../../../store/service'
  import { IConfigVersion } from '../../../../../../../../../../types/config'
  import { getServiceGroupList } from '../../../../../../../../../api/group'
  import { getConfigVersionList } from '../../../../../../../../../api/config'
  import { IGroupToPublish, IGroupItemInService } from '../../../../../../../../../../types/group'
  import Group from './group.vue'
  import Preview from './preview.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const { appData } = storeToRefs(useServiceStore())


  const props = defineProps<{
    groups: IGroupToPublish[];
    disabled?: number[];
  }>()
  const emits = defineEmits(['openPreviewVersionDiff', 'change'])

  const groupListLoading = ref(true)
  const groupList = ref<IGroupToPublish[]>([])
  const versionListLoading = ref(true)
  const versionList = ref<IConfigVersion[]>([])
  const allowPreviewDelete = ref(true)

  onMounted(() => {
    getAllGroupData()
    getAllVersionData()
  })

  // 获取所有分组，并转化为tree组件需要的结构
  const getAllGroupData = async() => {
    groupListLoading.value = true
    const res = await getServiceGroupList(spaceId.value, <number>appData.value.id)
    groupList.value = res.details.map((group: IGroupItemInService) => {
      const { group_id, group_name, release_id, release_name } = group
      const selector = group.new_selector
      const rules = selector.labels_and || selector.labels_or || []
      return { id: group_id, name: group_name, release_id, release_name, rules: rules }
    })
    groupListLoading.value = false
  }

    // 加载全量版本列表
    const getAllVersionData = async() => {
    versionListLoading.value = true
    const params = {
      start: 0,
      limit: 1000
    }
    const res = await getConfigVersionList(spaceId.value, Number(appData.value.id), params)
    // 只需要已上线的分组
    versionList.value = res.data.details.filter((item: IConfigVersion) => item.status.publish_status !== 'not_released')
    versionListLoading.value = false
  }

</script>
<template>
  <div class="select-group-wrapper">
    <div class="group-tree-area">
      <bk-loading style="height: 100%;" :loading="groupListLoading">
        <Group
          v-if="!groupListLoading"
          :group-list="groupList"
          :group-list-loading="groupListLoading"
          :version-list="versionList"
          :version-list-loading="versionListLoading"
          :disabled="props.disabled"
          :value="props.groups"
          @togglePreviewDelete="allowPreviewDelete = $event"
          @change="emits('change', $event)" />
      </bk-loading>
    </div>
    <div class="preview-area">
      <Preview
        :group-list="groupList"
        :group-list-loading="groupListLoading"
        :version-list="versionList"
        :version-list-loading="versionListLoading"
        :allow-preview-delete="allowPreviewDelete"
        :disabled="props.disabled"
        :value="props.groups"
        @diff="emits('openPreviewVersionDiff', $event)"
        @change="emits('change', $event)"  />
    </div>
  </div>
</template>
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