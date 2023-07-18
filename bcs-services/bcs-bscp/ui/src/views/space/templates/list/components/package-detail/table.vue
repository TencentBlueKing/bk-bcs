<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { storeToRefs } from 'pinia'
  import { Ellipsis } from 'bkui-vue/lib/icon'
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
    <bk-table :border="['outer']" :data="list">
      <bk-table-column label="配置项名称">
        <template #default="{ row }">
          <div v-if="row.spec">{{ row.spec.name }}</div>
        </template>
      </bk-table-column>
      <bk-table-column label="配置项路径" prop="spec.path"></bk-table-column>
      <bk-table-column label="配置项描述" prop="spec.memo"></bk-table-column>
      <bk-table-column label="所在套餐"></bk-table-column>
      <bk-table-column label="被引用"></bk-table-column>
      <bk-table-column label="更新人" prop="revision.reviser"></bk-table-column>
      <bk-table-column label="更新时间" prop="revision.update_at"></bk-table-column>
      <bk-table-column label="操作">
        <template #default="{ row }">
          <div class="actions-wrapper">
            <bk-button theme='primary' text>版本管理</bk-button>
            <div class="more-actions">
              <Ellipsis class="ellipsis-icon" />
            </div>
          </div>
        </template>
      </bk-table-column>
    </bk-table>
  </div>
</template>
<style lang="scss" scoped>
  .actions-wrapper {
    display: flex;
    align-items: center;
    height: 100%;
    .more-actions {
      display: flex;
      align-items: center;
      justify-content: center;
      margin-left: 16px;
      width: 16px;
      height: 16px;
      border-radius: 50%;
      cursor: pointer;
      &:hover {
        background: #dcdee5;
        color: #3a84ff;
      }
    }
    .ellipsis-icon {
      transform: rotate(90deg);
    }
  }
</style>
