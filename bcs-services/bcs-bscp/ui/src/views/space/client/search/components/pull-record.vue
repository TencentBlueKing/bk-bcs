<template>
  <bk-sideslider :is-show="show" width="1200" quick-close @closed="emits('close')">
    <template #header>
      <div class="header">
        <span class="title">配置拉取记录</span>
        <span class="uid">UID : {{ uid }}</span>
      </div>
    </template>
    <div class="content-wrap">
      <div class="operate-area">
        <bk-date-picker
          class="date-picker"
          v-model="initDateTime"
          type="datetimerange"
          format="yyyy-MM-dd"
          disable-date
          @change="handleYearChange" />
        <SearchInput v-model="searchStr" :width="600" :placeholder="'当前配置版本/目标配置版本/最近一次拉取配置状态'" />
      </div>
      <bk-table :data="tableData" :border="['outer', 'row']" :pagination="pagination">
        <bk-table-column label="开始时间" width="154">
          <template #default="{ row }">
            <span v-if="row.spec">
              {{ datetimeFormat(row.spec.start_time) }}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column label="结束时间" width="154">
          <template #default="{ row }">
            <span v-if="row.spec">
              {{ datetimeFormat(row.spec.end_time) }}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column label="当前配置版本" width="133">
          <template #default="{ row }">
            <div
              v-if="row.spec && row.spec.original_release_id"
              class="config-version"
              @click="linkToApp(row.spec.original_release_id)">
              <Share fill="#979BA5" />
              <span class="text">{{ row.original_release_name }}</span>
            </div>
          </template>
        </bk-table-column>
        <bk-table-column label="目标配置版本" width="134">
          <template #default="{ row }">
            <div
              v-if="row.spec && row.spec.target_release_id"
              class="config-version"
              @click="linkToApp(row.spec.target_release_id)">
              <Share fill="#979BA5" />
              <span class="text">{{ row.target_release_name }}</span>
            </div>
          </template>
        </bk-table-column>
        <bk-table-column label="配置拉取方式" prop="attachment.client_mode"></bk-table-column>
        <bk-table-column label="配置拉取耗时(秒)" width="128">
          <template #default="{ row }">
            <span v-if="row.spec">
              {{ parseInt(row.spec.total_seconds) }}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column label="配置拉取文件数" prop="spec.download_file_num"></bk-table-column>
        <bk-table-column label="配置拉取文件大小" prop="spec.total_file_size"></bk-table-column>
        <bk-table-column label="配置拉取状态" width="104">
          <template #default="{ row }">
            <div v-if="row.spec" class="release_change_status">
              <Spinner v-if="row.spec.release_change_status === 'Skip'" class="spinner-icon" fill="#3A84FF" />
              <div v-else :class="['dot', row.spec.release_change_status]"></div>
              <span>{{ CLIENT_STATUS_MAP[row.spec.release_change_status as keyof typeof CLIENT_STATUS_MAP] }}</span>
              <InfoLine
                v-if="row.spec.release_change_status === 'Failed'"
                class="info-icon"
                fill="#979BA5"
                v-bk-tooltips="{ content: row.spec.failed_detail_reason }" />
            </div>
          </template>
        </bk-table-column>
      </bk-table>
    </div>
  </bk-sideslider>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useRouter } from 'vue-router';
  import { Share, Spinner, InfoLine } from 'bkui-vue/lib/icon';
  import SearchInput from '../../../../../components/search-input.vue';
  import { getClientPullRecord } from '../../../../../api/client';
  import { datetimeFormat } from '../../../../../utils';
  import { CLIENT_STATUS_MAP } from '../../../../../constants/client';

  const router = useRouter();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    show: boolean;
    id: number;
    uid: string;
  }>();
  const emits = defineEmits(['close']);

  const isShowSlider = ref(false);
  const initDateTime = ref([new Date(), new Date()]);
  const searchStr = ref('');
  const tableData = ref();

  const pagination = ref({
    count: 0,
    current: 1,
    limit: 10,
  });
  watch(
    () => props.show,
    (val) => {
      if (val) {
        isShowSlider.value = true;
        loadTableData();
      }
    },
  );

  const loadTableData = async () => {
    const params = {
      start: 0,
      all: true,
      limit: 10,
    };
    const resp = await getClientPullRecord(props.bkBizId, props.appId, props.id, params);
    pagination.value.count = resp.data.count;
    tableData.value = resp.data.details;
  };
  const handleYearChange = (val: string[]) => {
    console.log(val);
  };
  const linkToApp = (versionId: number) => {
    router.push({ name: 'service-config', params: { spaceId: props.bkBizId, appId: props.appId, versionId } });
  };
</script>

<style scoped lang="scss">
  .header {
    .title {
      position: relative;
      &::after {
        position: absolute;
        right: -9px;
        content: '';
        width: 1px;
        height: 24px;
        background: #dcdee5;
      }
    }
    .uid {
      font-size: 14px;
      color: #63656e;
      margin-left: 17px;
    }
  }
  .content-wrap {
    padding: 20px 24px;
    .operate-area {
      display: flex;
      justify-content: flex-end;
      margin-bottom: 16px;
      .date-picker {
        width: 300px;
        margin-right: 8px;
      }
    }
  }
  .release_change_status {
    display: flex;
    align-items: center;
    .spinner-icon {
      font-size: 12px;
      margin-right: 10px;
    }
    .dot {
      margin-right: 10px;
      width: 8px;
      height: 8px;
      background: #f0f1f5;
      border: 1px solid #c4c6cc;
      border-radius: 50%;
      &.Success {
        background: #e5f6ea;
        border: 1px solid #3fc06d;
      }
      &.Failed {
        background: #ffe6e6;
        border: 1px solid #ea3636;
      }
    }
    .info-icon {
      font-size: 14px;
      margin-left: 9px;
    }
  }
  .config-version {
    display: flex;
    align-items: center;
    cursor: pointer;
    .text {
      margin-left: 4px;
    }
  }
</style>
