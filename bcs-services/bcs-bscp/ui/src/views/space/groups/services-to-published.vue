<template>
  <bk-sideslider :title="`${name}-上线服务`" :width="640" :is-show="props.show" @closed="handleClose">
    <div class="services-content">
      <div class="search-area">
        <SearchInput v-model="searchStr" placeholder="服务名称/服务版本" :width="320" @search="handleSearch" />
      </div>
      <bk-loading class="loading-wrapper" :loading="loading">
        <bk-table
          :data="list"
          :border="['outer']"
          :remote-pagination="true"
          :pagination="pagination"
          @page-limit-change="handlePageLimitChange"
          @page-value-change="loadServicesList"
        >
          <bk-table-column label="服务名称" prop="app_name"></bk-table-column>
          <bk-table-column label="服务版本">
            <template #default="{ row }">
              <bk-link v-if="row.app_id" class="link-btn" theme="primary" target="_blank" :href="getHref(row)">{{
                row.release_name
              }}</bk-link>
            </template>
          </bk-table-column>
        </bk-table>
      </bk-loading>
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
import useGlobalStore from '../../../store/global';
import { IGroupBindService } from '../../../../types/group';
import { getGroupReleasedApps } from '../../../api/group';
import { ICommonQuery } from '../../../../types/index';
import SearchInput from '../../../components/search-input.vue';

const router = useRouter();

const { spaceId } = storeToRefs(useGlobalStore());

const props = defineProps<{
  show: boolean;
  id: number;
  name: string;
}>();

const emits = defineEmits(['update:show']);

const loading = ref(true);
const list = ref<IGroupBindService[]>([]);
const searchStr = ref('')
const pagination = ref({
  count: 0,
  limit: 10,
  current: 1,
});

watch(
  () => props.show,
  (val) => {
    if (val) {
      loadServicesList();
    }
  },
);

const loadServicesList = async () => {
  loading.value = true;
  const params: ICommonQuery = {
    start: pagination.value.limit * (pagination.value.current - 1),
    limit: pagination.value.limit,
  };
  if (searchStr.value) {
    params.search_key = searchStr.value
  }
  const res = await getGroupReleasedApps(spaceId.value, props.id, params);
  list.value = res.details;
  pagination.value.count = res.count;
  loading.value = false;
};

const handleSearch = () => {
  pagination.value.current = 1
  pagination.value.count = 0
  loadServicesList()
}

const getHref = (service: IGroupBindService) => {
  const { href } = router.resolve({
    name: 'service-config',
    params: { spaceId: spaceId.value, appId: service.app_id, versionId: service.release_id },
  });
  return href;
};

const handlePageLimitChange = (val: number) => {
  pagination.value.limit = val;
  loadServicesList();
};

const handleClose = () => {
  emits('update:show', false);
};
</script>
<style lang="scss" scoped>
.services-content {
  padding: 24px;
  height: calc(100vh - 101px);
  overflow: auto;
}
.loading-wrapper {
  height: calc(100% - 48px);
}
.search-area {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}
.link-btn {
  font-size: 12px;
}
.action-btn {
  padding: 8px 24px;
  border-top: 1px solid #dcdee5;
  background: #fafbfd;
  .bk-button {
    min-width: 88px;
  }
}
</style>
