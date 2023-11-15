<template>
  <bk-sideslider title="关联配置文件" width="640" :is-show="props.show" @closed="handleClose">
    <div class="search-area">
      <SearchInput v-model="searchStr" placeholder="服务名称/版本名称/被引用的版本" @search="refreshList()" />
    </div>
    <div class="cited-data-table">
      <bk-table
        v-bkloading="{ loading }"
        :border="['outer']"
        :data="list"
        :remote-pagination="true"
        :pagination="pagination"
        @page-limit-change="handlePageLimitChange"
        @page-value-change="refreshList"
      >
        <bk-table-column label="脚本版本">
          <template #default="{ row }">
            <template v-if="row.hook_revision_name || row.revision_name">{{
              row.hook_revision_name || row.revision_name
            }}</template>
          </template>
        </bk-table-column>
        <bk-table-column label="脚本类型">
          <template #default="{ row }">{{ row.type === 'pre_hook' ? '前置脚本':'后置脚本' }}</template>
        </bk-table-column>
        <bk-table-column label="服务名称" prop="app_name"></bk-table-column>
        <bk-table-column label="服务版本">
          <template #default="{ row }">
            <bk-link
              v-if="row.release_name"
              class="link-btn"
              theme="primary"
              target="_blank"
              :href="getHref(row.app_id, row.release_id)"
              >{{ row.release_name }}</bk-link
            >
          </template>
        </bk-table-column>
      </bk-table>
    </div>
    <div class="action-btn">
      <bk-button @click="handleClose">关闭</bk-button>
    </div>
  </bk-sideslider>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import useGlobalStore from '../../../../store/global';
import { IScriptCiteQuery, IScriptCitedItem } from '../../../../../types/script';
import { getScriptCiteList, getScriptVersionCiteList } from '../../../../api/script';
import SearchInput from '../../../../components/search-input.vue';

const { spaceId } = storeToRefs(useGlobalStore());

const router = useRouter();

const props = defineProps<{
  show: boolean;
  id: number;
  versionId?: number;
}>();

const emits = defineEmits(['update:show']);

const loading = ref(false);
const list = ref<IScriptCitedItem[]>([]);
const searchStr = ref('');
const pagination = ref({
  current: 1,
  count: 0,
  limit: 10,
});

watch(
  () => props.show,
  (val) => {
    if (val) {
      getCitedData();
    }
  },
);

const getCitedData = async () => {
  loading.value = true;
  let res;
  const params: IScriptCiteQuery = {
    start: (pagination.value.current - 1) * pagination.value.limit,
    limit: pagination.value.limit,
  };
  if (searchStr.value) {
    params.searchKey = searchStr.value;
  }
  if (props.versionId) {
    res = await getScriptVersionCiteList(spaceId.value, props.id, props.versionId, params);
  } else {
    res = await getScriptCiteList(spaceId.value, props.id, params);
  }
  list.value = res.details;
  pagination.value.count = res.count;
  loading.value = false;
};

const getHref = (id: number, releaseId: number) => {
  const { href } = router.resolve({ name: 'service-config', params: { spaceId: spaceId.value, appId: id, versionId: releaseId } });
  return href;
};

const refreshList = (current = 1) => {
  pagination.value.current = current;
  getCitedData();
};

const handlePageLimitChange = (val: number) => {
  pagination.value.limit = val;
  getCitedData();
};

const handleClose = () => {
  emits('update:show', false);
  list.value = [];
  pagination.value.count = 0;
  pagination.value.current = 1;
};
</script>
<style lang="scss" scoped>
.search-area {
  padding: 24px 24px 16px;
  text-align: right;
  .search-input {
    width: 320px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
    background: #ffffff;
  }
}
.cited-data-table {
  padding: 0 24px;
  height: calc(100vh - 172px);
  overflow: auto;
}
.link-btn {
  font-size: 12px;
}
.action-btn {
  padding: 8px 24px;
  background: #fafbfd;
  box-shadow: 0 -1px 0 0 #dcdee5;
  .bk-button {
    min-width: 88px;
  }
}
</style>
