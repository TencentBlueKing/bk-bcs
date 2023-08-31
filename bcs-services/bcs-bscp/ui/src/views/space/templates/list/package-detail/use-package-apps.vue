<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { storeToRefs } from 'pinia'
  import { Plus } from 'bkui-vue/lib/icon';
  import { useGlobalStore } from '../../../../../store/global'
  import { useUserStore } from '../../../../../store/user'
  import { useTemplateStore } from '../../../../../store/template'
  import { getAppList } from '../../../../../api/index'
  import { getUnNamedVersionAppsBoundByPackage } from '../../../../../api/template'
  import { IAppItem } from '../../../../../../types/app'
  import { IPackageCitedByApps } from '../../../../../../types/template'
  import LinkToApp from '../components/link-to-app.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const { userInfo } = storeToRefs(useUserStore())
  const templateStore = useTemplateStore()
  const { currentTemplateSpace, currentPkg } = storeToRefs(templateStore)

  const userAppList = ref<IAppItem[]>([])
  const userAppListLoading = ref(false)
  const boundApps = ref<IPackageCitedByApps[]>([])
  const boundAppsLoading = ref(false)

  watch(() => currentPkg.value, () => {
    boundApps.value = []
    getBoundApps()
  })

  onMounted(() => {
    getBoundApps()
    getUserApps()
  })

  const getUserApps = async () => {
    userAppListLoading.value = true
    const params = {
      operator: userInfo.value.username,
      start: 0,
      all: true
    }
    const res = await getAppList(spaceId.value, params)
    userAppList.value = res.details
    userAppListLoading.value = false
  }

  const getBoundApps = async() => {
    if (typeof currentPkg.value !== 'number') return
    boundAppsLoading.value = true
    const params = {
      start: 0,
      limit: 1000
      // all: true
    }
    const res = await getUnNamedVersionAppsBoundByPackage(spaceId.value, currentTemplateSpace.value, <number>currentPkg.value, params)
    boundApps.value = res.details
    boundAppsLoading.value = false
  }

</script>
<template>
  <div class="use-package-apps">
    <bk-select
      :filterable="true"
      :inputSearch="false">
      <template #trigger>
        <div class="select-app-trigger">
          <Plus class="plus-icon" />
          新服务中使用
        </div>
      </template>
      <bk-option v-for="app in userAppList" :key="app.id" :id="app.id">{{ app.spec.name }}</bk-option>
    </bk-select>
    <div class="table-wrapper">
      <bk-table :border="['outer']" :data="boundApps">
        <bk-table-column label="当前使用此套餐的服务">
          <template #default="{ row }">
            <div class="app-info">
              <div class="name">{{ row.app_name }}</div>
              <LinkToApp :id="row.app_id" />
            </div>
          </template>
        </bk-table-column>
      </bk-table>
      <bk-pagination
        class="table-pagination"
        small
        align="center"
        :show-limit="false"
        :show-total-count="false">
      </bk-pagination>
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .use-package-apps {
    padding: 16px 24px;
    width: 240px;
    height: 100%;
    background: #ffffff;
  }
  .select-app-trigger {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 5px;
    width: 192px;
    height: 32px;
    line-height: 22px;
    border: 1px solid #c4c6cc;
    border-radius: 2px;
    color: #63656e;
    font-size: 14px;
    overflow: hidden;
    cursor: pointer;
    .plus-icon {
      font-size: 20px;
    }
  }
  .table-wrapper {
    margin-top: 16px;
    .table-pagination {
      margin-top: 16px;
    }
  }
</style>
