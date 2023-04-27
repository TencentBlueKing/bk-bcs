<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../../../../../store/global'
  import { useServiceStore } from '../../../../../../../../store/service'
  import { IConfigVersion } from '../../../../../../../../../types/config'
  import { getServiceGroupList } from '../../../../../../../../api/group'
  import { getConfigVersionList } from '../../../../../../../../api/config'
  import { IGroupTreeItem, IGroupItemInService, IGroupItem } from '../../../../../../../../../types/group'

  const { spaceId } = storeToRefs(useGlobalStore())
  const { appData } = storeToRefs(useServiceStore())

  import Group from './group.vue'
  import Preview from './preview.vue'

  const props = defineProps<{
    groups: IGroupTreeItem[]
  }>()
  const emits = defineEmits(['change'])

  const groupListLoading = ref(true)
  const groupList = ref<IGroupItemInService[]>([])
  const versionListLoading = ref(true)
  const versionList = ref<IConfigVersion[]>([])

  onMounted(() => {
    getAllGroupData()
    getAllVersionData()
  })

  // 获取所有分组，并转化为tree组件需要的结构
  const getAllGroupData = async() => {
    groupListLoading.value = true
    const res = await getServiceGroupList(spaceId.value, <number>appData.value.id)
    groupList.value = res.details
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
      <Group
        :group-list="groupList"
        :group-list-loading="groupListLoading"
        :version-list="versionList"
        :version-list-loading="versionListLoading"
        :groups="props.groups"
        @change="emits('change', $event)" />
    </div>
    <div class="preview-area">
      <Preview
        :group-list="groupList"
        :group-list-loading="groupListLoading"
        :version-list="versionList"
        :version-list-loading="versionListLoading"
        :value="props.groups" />
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .select-group-wrapper {
    display: flex;
    align-items: center;
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
    padding: 24px;
    height: 100%;
    background: #f5f7fa;
  }
</style>