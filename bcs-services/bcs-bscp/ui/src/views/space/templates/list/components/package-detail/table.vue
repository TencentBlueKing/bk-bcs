<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../store/template'
  import { ITemplateConfigItem } from '../../../../../../../types/template';
  import { getPackageConfigList } from '../../../../../../api/template';

  const { spaceId } = storeToRefs(useGlobalStore())
  const templateStore = useTemplateStore()
  const { currentTemplateSpace } = storeToRefs(templateStore)

  const loading = ref(false)
  const list = ref<ITemplateConfigItem[]>([])
  const pagination = ref({
    current: 1,
    limit: 10,
    count: 0
  })

  watch(() => currentTemplateSpace.value, val => {
    if (val) {
      loadConfigList()
    }
  })

  // onMounted(() => {
  //   loadConfigList()
  // })

  const loadConfigList = async () => {
    loading.value = true
    const params = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    const res = await getPackageConfigList(spaceId.value, currentTemplateSpace.value, params)
    list.value = res.details
    loading.value = false
  }
</script>
<template>
  <div class="package-config-table">
    <bk-table :data="list">
      <bk-table-column label="配置项名称"></bk-table-column>
      <bk-table-column label="配置项路径"></bk-table-column>
      <bk-table-column label="配置项描述"></bk-table-column>
      <bk-table-column label="所在套餐"></bk-table-column>
      <bk-table-column label="被引用"></bk-table-column>
      <bk-table-column label="更新人"></bk-table-column>
      <bk-table-column label="更新时间"></bk-table-column>
      <bk-table-column label="操作"></bk-table-column>
    </bk-table>
  </div>
</template>
