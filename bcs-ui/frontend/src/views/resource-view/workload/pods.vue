<template>
  <div class="h-full">
    <BaseLayout title="Pods" kind="Pod" category="pods" type="workloads" ref="baseLayoutRef">
      <template
        #search="{
          curNsList, isViewConfigShow, showFilter,
          clusterID, handleShowViewConfig, handleNsChange }">
        <template v-if="showFilter">
          <div class="flex items-start justify-end flex-1 pl-[24px] text-[12px] h-[32px] z-10">
            <span
              :class="[
                'inline-flex items-center justify-center bg-[#fff] w-[32px] h-[32px] mr-[8px]',
                'border border-solid border-[#C4C6CC] rounded-sm cursor-pointer',
                isViewConfigShow ? '!border-[#3a84ff] text-[#3a84ff]' : 'text-[#979BA5] hover:!border-[#979BA5]',
              ]"
              @click="handleShowViewConfig">
              <i class="bk-icon icon-funnel text-[14px]"></i>
            </span>
            <span
              :class="[
                'inline-flex items-center justify-center h-[32px] px-[8px] text-[12px]',
                'border border-solid border-[#C4C6CC] rounded-l-sm bg-[#FAFBFD] mr-[-1px]',
              ]">
              {{$t('k8s.namespace')}}
            </span>
            <NsSelect
              :value="curNsList"
              class="flex-1 bg-[#fff] max-w-[240px] mr-[8px]"
              :cluster-id="clusterID"
              display-tag
              @change="handleNsChange" />
            <bcs-search-select
              class="flex-1 bg-[#fff] max-w-[460px]"
              clearable
              :show-condition="false"
              :show-popover-tag-change="false"
              :data="searchSelectData"
              :values="searchSelectValue"
              :placeholder="$t('view.placeholder.searchNameOrCreatorOrIP')"
              :key="searchSelectKey"
              @change="searchSelectChange"
              @clear="searchSelectChange()" />
          </div>
        </template>
      </template>
      <template
        #default="{
          curPageData,
          pageConf,
          handlePageChange,
          handlePageSizeChange,
          handleGetExtData,
          gotoDetail,
          handleSortChange,
          handleUpdateResource,
          handleDeleteResource,
          handleShowViewConfig,
          clusterNameMap,
          goNamespace,
          handleFilterChange,
          isViewEditable,
          isClusterMode,
          sourceTypeMap
        }">
        <bk-table
          :data="curPageData"
          :pagination="pageConf"
          @page-change="handlePageChange"
          @page-limit-change="handlePageSizeChange"
          @sort-change="handleSortChange"
          @filter-change="handleFilterChange">
          <bk-table-column :label="$t('generic.label.name')" min-width="130" prop="metadata.name" sortable fixed="left">
            <template #default="{ row }">
              <bk-button
                class="bcs-button-ellipsis"
                text
                :disabled="isViewEditable"
                @click="gotoDetail(row)">{{ row.metadata.name }}</bk-button>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('cluster.labels.nameAndId')" min-width="160" v-if="!isClusterMode">
            <template #default="{ row }">
              <div class="flex flex-col py-[6px] h-[50px]">
                <span class="bcs-ellipsis">{{ clusterNameMap[handleGetExtData(row.metadata.uid, 'clusterID')] }}</span>
                <span class="bcs-ellipsis mt-[6px]">{{ handleGetExtData(row.metadata.uid, 'clusterID') }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column
            :label="$t('k8s.namespace')"
            width="150"
            prop="metadata.namespace"
            sortable>
            <template #default="{ row }">
              <bk-button
                class="bcs-button-ellipsis"
                text
                :disabled="isViewEditable"
                @click="goNamespace(row)">
                {{ row.metadata.namespace }}
              </bk-button>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('k8s.image')" min-width="200" :show-overflow-tooltip="false">
            <template #default="{ row }">
              <span v-bk-tooltips.top="(handleGetExtData(row.metadata.uid, 'images') || []).join('<br />')">
                {{ (handleGetExtData(row.metadata.uid, 'images') || []).join(', ') }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column
            label="Status"
            width="120"
            prop="status"
            :resizable="false"
            column-key="status"
            :filters="podStatusFilters"
            filter-multiple>
            <template #default="{ row }">
              <StatusIcon :status="handleGetExtData(row.metadata.uid, 'status')"></StatusIcon>
            </template>
          </bk-table-column>
          <bk-table-column label="Ready" width="100" :resizable="false">
            <template #default="{ row }">
              {{handleGetExtData(row.metadata.uid, 'readyCnt') || 0}}
              / {{handleGetExtData(row.metadata.uid, 'totalCnt') || 0}}
            </template>
          </bk-table-column>
          <bk-table-column label="Restarts" width="100" :resizable="false">
            <template #default="{ row }">{{handleGetExtData(row.metadata.uid, 'restartCnt') || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Host IP" width="140">
            <template #default="{ row }">{{row.status.hostIP || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Pod IPv4" width="140">
            <template #default="{ row }">{{handleGetExtData(row.metadata.uid, 'podIPv4') || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Pod IPv6" min-width="200">
            <template #default="{ row }">{{handleGetExtData(row.metadata.uid, 'podIPv6') || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Node">
            <template #default="{ row }">{{row.spec.nodeName || '--'}}</template>
          </bk-table-column>
          <bk-table-column label="Age" sortable="custom" prop="createTime">
            <template #default="{ row }">
              <span>{{handleGetExtData(row.metadata.uid, 'age')}}</span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.label.source')" :show-overflow-tooltip="false">
            <template #default="{ row }">
              <div class="flex items-center">
                <bk-popover
                  class="size-[16px] mr-[4px]"
                  :content="sourceTypeMap?.[handleGetExtData(row.metadata.uid, 'createSource')]?.iconText"
                  :tippy-options="{ interactive: false }">
                  <i
                    class="text-[14px] p-[1px]"
                    :class="sourceTypeMap?.[handleGetExtData(row.metadata.uid, 'createSource')]?.iconClass"></i>
                </bk-popover>
                <span
                  v-bk-overflow-tips="{ interactive: false }" class="bcs-ellipsis"
                  v-if="handleGetExtData(row.metadata.uid, 'createSource') === 'Template'">
                  {{ `${handleGetExtData(row.metadata.uid, 'templateName') || '--'}:${
                    handleGetExtData(row.metadata.uid, 'templateVersion') || '--'}` }}
                </span>
                <span
                  v-bk-overflow-tips="{ interactive: false }"
                  class="bcs-ellipsis" v-else>{{ handleGetExtData(row.metadata.uid, 'createSource') }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.label.editMode.text')" width="100">
            <template #default="{ row }">
              <span>
                {{handleGetExtData(row.metadata.uid, 'editMode') === 'form'
                  ? $t('generic.label.editMode.form') : 'YAML'}}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column
            :label="$t('generic.label.action')"
            :resizable="false"
            width="180"
            fixed="right"
            v-if="!isViewEditable">
            <template #default="{ row }">
              <bk-button text @click="handleShowLog(row, handleGetExtData(row.metadata.uid))">
                {{ $t('generic.button.log1') }}
              </bk-button>
              <bk-button
                text class="ml10"
                @click="handleUpdateResource(row)">{{ $t('generic.button.update') }}</bk-button>
              <bk-button
                class="ml10" text
                @click="handleDeleteResource(row)">{{ $t('generic.button.delete') }}</bk-button>
            </template>
          </bk-table-column>
          <template #empty>
            <BcsEmptyTableStatus
              :button-text="$t('generic.button.resetSearch')"
              type="search-empty"
              @clear="handleShowViewConfig" />
          </template>
        </bk-table>
      </template>
    </BaseLayout>
    <BcsLog
      v-model="showLog"
      :cluster-id="currentRow.clusterID"
      :namespace="currentRow.metadata.namespace"
      :name="currentRow.metadata.name">
    </BcsLog>
  </div>
</template>
<script lang="ts">
import { cloneDeep } from 'lodash';
import { computed, defineComponent, ref, watch } from 'vue';

import NsSelect from '../view-manage/ns-select.vue';
import useViewConfig from '../view-manage/use-view-config';

import { ISearchSelectValue } from '@/@types/bkui-vue';
// import CustomTableFilter from './custom-table-filter.vue';
import BcsLog from '@/components/bcs-log/log-dialog.vue';
import StatusIcon from '@/components/status-icon';
import useDebouncedRef from '@/composables/use-debounce';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store';
import BaseLayout from '@/views/resource-view/common/base-layout';

export default defineComponent({
  name: 'WorkloadPods',
  components: { BaseLayout, StatusIcon, BcsLog, NsSelect },
  setup() {
    // 显示日志
    const showLog = ref(false);
    const currentRow = ref<Record<string, any>>({ metadata: {} });
    const handleShowLog = (row, ext = {}) => {
      currentRow.value = {
        ...row,
        ...ext,
      };
      showLog.value = true;
    };

    // Pod状态
    const podStatusFilters = ref([
      'Pending',
      'Running',
      'Succeeded',
      'Failed',
      'Unknown',
      'CrashLoopBackOff',
      'ImagePullBackOff',
      'Terminating',
      'Evicted',
      'NotReady',
      'Completed',
    ].map(status => ({
      text: status,
      value: status,
    })));

    // 表头IP输入过滤
    // const baseLayoutRef = ref();
    // const renderFilterHeader = (_, { column }) => h(CustomTableFilter, {
    //   props: {
    //     label: column.label,
    //   },
    //   on: {
    //     confirm: (v: string) => {
    //       const ipList = v.split('\n');
    //       baseLayoutRef.value.customFilters.ip = ipList;
    //       baseLayoutRef.value.handleGetTableData();
    //     },
    //   },
    // });

    // 搜索
    const searchSelectKey = ref('');
    const baseLayoutRef = ref();
    const ipData = useDebouncedRef<string>('', 300);// IP过滤信息
    watch(ipData, () => {
      const ipList = ipData.value.split(' ');
      baseLayoutRef.value.customFilters.ip = ipList;
      baseLayoutRef.value.handleGetTableData();
    });
    const { curViewData } = useViewConfig();
    const searchSelectValue = computed<ISearchSelectValue[]>(() => {
      const data: ISearchSelectValue[] = [];
      if (curViewData.value?.name) {
        data.push({
          id: 'name',
          name: $i18n.t('view.labels.resourceName'),
          values: [{ name: curViewData.value?.name }],
        });
      }
      if (curViewData.value?.creator?.length) {
        data.push({
          id: 'creator',
          name: $i18n.t('view.labels.creator'),
          values: curViewData.value?.creator?.map(name => ({ name })),
        });
      }
      if (ipData.value) {
        data.push({
          id: 'ip',
          name: 'IP',
          values: [{ name: ipData.value }],
        });
      }
      return data;
    });
    const searchSelectDataSource = ref([
      {
        name: $i18n.t('view.labels.resourceName'),
        id: 'name',
      },
      {
        name: $i18n.t('view.labels.creator'),
        id: 'creator',
      },
      {
        name: 'IP',
        id: 'ip',
      },
    ]);
    const searchSelectData = computed(() => {
      const ids = searchSelectValue.value.map(item => item.id);
      return searchSelectDataSource.value.filter(item => !ids.includes(item.id));
    });
    const searchSelectChange = (v: any[] = []) => {
      if (!v.length) {
        ipData.value = '';
      }
      const data = v.reduce((pre, item) => {
        let newItem = cloneDeep(item);
        // 没有选择字段时默认搜索名称
        if (newItem.values === undefined) {
          newItem = {
            id: 'name',
            values: [item],
          };
        }
        if (newItem.id === 'name') { // 名称API只能给字符串类型
          pre[newItem.id] = newItem.values?.map(item => item.name)?.join('');
        } else if (newItem.id === 'ip') {
          // IP单独存储
          ipData.value = newItem.values?.map(item => item.name).join(',');
        } else {
          pre[newItem.id] = newItem.values?.map(item => item.name);
        }

        return pre;
      }, {});
      $store.commit('updateTmpViewData', {
        filter: data,
      });
      // hack 修复search select搜索完后还展示的BUG
      setTimeout(() => {
        searchSelectKey.value = new Date().getTime()
          .toString();
      });
    };

    return {
      baseLayoutRef,
      showLog,
      currentRow,
      podStatusFilters,
      searchSelectData,
      searchSelectKey,
      searchSelectValue,
      searchSelectChange,
      handleShowLog,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './detail/pod-log.css';
/deep/ .base-layout {
    width: 100%;
}
</style>
