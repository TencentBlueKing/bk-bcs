<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue'
  import { storeToRefs } from 'pinia';
  import { useRoute, useRouter } from 'vue-router'
  import { ArrowsLeft, Plus, Search } from 'bkui-vue/lib/icon';
  import { useGlobalStore } from '../../../../store/global'
  import { ITemplateConfigItem, ITemplateVersionItem } from '../../../../../types/template'
  import { IPagination, ICommonQuery } from '../../../../../types/index';
  import { getTemplatesDetailByIds, getTemplateVersionList } from '../../../../api/template'
  import VersionFullTable from './version-full-table.vue';

  const getRouteId = (id: string) => {
    if (id && typeof Number(id) === 'number') {
      return Number(id)
    }
    return 0
  }

  const { spaceId } = storeToRefs(useGlobalStore())
  const route = useRoute()
  const router = useRouter()
  const templateDetailLoading = ref(false)
  const templateDetail = ref<ITemplateConfigItem>()
  const versionListLoading = ref(false)
  const versionList = ref<ITemplateVersionItem[]>([])
  const searchStr = ref('')
  const pagination = ref<IPagination>({
    count: 0,
    current: 1,
    limit: 10,
  })

  const templateSpaceId = computed(() => {
    return getRouteId(<string>route.params.templateSpaceId)
  })
  const packageId = computed(() => {
    return route.params.packageId
  })
  const templateId = computed(() => {
    return getRouteId(<string>route.params.templateId)
  })

  onMounted(() => {
    getTemplateDetail()
    getVersionList()
  })

  const getTemplateDetail = async() => {
    templateDetailLoading.value = true
    const res = await getTemplatesDetailByIds(spaceId.value, [templateId.value])
    if (res.details.length > 0) {
      templateDetail.value = res.details[0]
    }
    templateDetailLoading.value = false
  }

  const getVersionList = async() => {
    versionListLoading.value = true
    const params: ICommonQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit
    }
    const res = await getTemplateVersionList(spaceId.value, templateSpaceId.value, templateId.value, params)
    versionList.value = res.details
  }

  const refreshList = (current: number = 1) => {
    pagination.value.current = current
    getVersionList()
  }

  const handleSearchInputChange = (val: string) => {
    if (!val) {
      refreshList()
    }
  }

  const goToTemplateListPage = () => {
    router.push({ name: 'templates-list', params: {
      templateSpaceId: templateSpaceId.value,
      packageId: packageId.value
    }})
  }

</script>
<template>
  <div class="template-version-manage-page">
    <div class="page-header">
      <ArrowsLeft class="arrow-icon" @click="goToTemplateListPage" />
      <div v-if="templateDetail" class="title-name">
        版本管理 - {{ templateDetail.spec.name }}
        <span class="path">{{ templateDetail.spec.path }}</span>
      </div>
    </div>
    <div class="operation-area">
      <bk-button theme="primary">
        <Plus class="button-icon" />
        新建版本
      </bk-button>
      <bk-input
        v-model.trim="searchStr"
        class="search-input"
        placeholder="版本号/版本说明/更新人"
        :clearable="true"
        @enter="refreshList()"
        @clear="refreshList()"
        @change="handleSearchInputChange">
          <template #suffix>
            <Search class="search-input-icon" />
          </template>
      </bk-input>
    </div>
    <div class="version-content-area">
      <VersionFullTable
        :spaceId="spaceId"
        :currentTemplateSpace="templateSpaceId"
        :templateId="templateId"
        :list="versionList"
        :pagination="pagination" />
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .template-version-manage-page {
    height: 100%;
    background: #f5f7fa;
  }
  .page-header {
    display: flex;
    align-items: center;
    padding: 0 24px;
    height: 52px;
    background: #ffffff;
    box-shadow: 0 3px 4px 0 #0000000a;
    .arrow-icon {
      font-size: 24px;
      color: #3a84ff;
      cursor: pointer;
    }
    .title-name {
      padding: 14px 0;
      font-size: 16px;
      line-height: 24px;
      color: #313238;
      .path {
        margin-left: 12px;
        color: #979ba5;
        font-size: 12px;
      }
    }
  }
  .operation-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 24px;
    padding: 0 24px;
    .button-icon {
      font-size: 18px;
    }
    .search-input {
      width: 320px;
    }
    .search-input-icon {
      padding-right: 10px;
      color: #979ba5;
      background: #ffffff;
    }
  }
  .version-content-area {
    padding: 16px 24px 24px;
    height: calc(100% - 110px);
    overflow: hidden;
  }
</style>
