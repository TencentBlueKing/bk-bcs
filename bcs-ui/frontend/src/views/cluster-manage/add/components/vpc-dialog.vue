<template>
  <bk-dialog
    render-directive="show"
    header-position="left"
    :value="value"
    :show-footer="false"
    :mask-close="false"
    :width="1000"
    :title="$t('tke.label.selectVPC')"
    :z-index="99"
    @value-change="handleValueBefore">
    <bcs-search-select
      clearable
      class="bg-[#fff]"
      :data="searchSelectData"
      :show-condition="false"
      :show-popover-tag-change="false"
      :placeholder="$t('generic.placeholder.searchWith', [`${$t('generic.label.name')} / ID`])"
      ref="searchSelect"
      v-model="searchSelectValue"
      @change="searchSelectChange"
      @clear="handleClear">
    </bcs-search-select>
    <bcs-table
      class="mt-[16px]"
      v-bkloading="{ isLoading: vpcLoading }"
      :ref="(v) => tableRef = v"
      :data="curPageData"
      :loading="vpcLoading"
      :pagination="pagination"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange"
      @row-click="handleClickCell"
      @sort-change="handleSortChange">
      <bcs-table-column width="50" :resizable="false">
        <template #default="{ row }">
          <span
            v-bk-tooltips="{
              disabled: !(disabledFn?.(row)),
              content: $t('cluster.create.label.vpc.deficiencyIpNumTips')
            }"
          >
            <bcs-radio
              :value="localValue === row.vpcID"
              :disabled="disabledFn?.(row)">
            </bcs-radio>
          </span>
        </template>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('tke.label.vpcName')"
        prop="vpcName"
        min-width="100"
        show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.vpcName || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('tke.label.vpcID')"
        prop="vpcID"
        min-width="100">
        <template #default="{ row }">
          {{ row.vpcID || '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column
        :label="`VPC ${$t('generic.label.kind')}`"
        min-width="100">
        <template #default="{ row }">
          {{ row.businessID ?
            $t('cluster.create.label.vpc.businessSpecificVPC')
            : $t('cluster.create.label.vpc.publicVPC')
          }}
        </template>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('tke.label.availableUnderlay')"
        :sort-method="(row1, row2) => row1.underlay?.availableIPNum - row2.underlay?.availableIPNum"
        prop="underlayAvaliableIPNum"
        sortable
        min-width="100">
        <template #default="{ row }">
          {{ row.underlay?.availableIPNum ?? '--' }}
        </template>
      </bcs-table-column>
      <bcs-table-column
        :label="$t('tke.label.availableOverlay')"
        :sort-method="(row1, row2) => row1.overlay?.availableIPNum - row2.overlay?.availableIPNum"
        :sort-orders="['descending', 'ascending', null]"
        prop="overlayAvaliableIPNum"
        sortable
        min-width="100">
        <template #default="{ row }">
          {{ row.overlay?.availableIPNum ?? '--' }}
        </template>
      </bcs-table-column>
    </bcs-table>
    <template #footer>
      <div class="h-[48px] bg-[#FAFBFD] flex items-center justify-end">
        <bcs-button
          theme="primary"
          :disabled="!localValue"
          @click="handleConfirm">
          {{ $t('generic.button.confirm') }}
        </bcs-button>
        <bcs-button
          @click="handleCancel">
          {{ $t('generic.button.cancel') }}
        </bcs-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script lang="ts">
import { Table } from 'bk-magic-vue';
import { computed, defineComponent, ref, watch } from 'vue';

import { cloudVpc } from '@/api/modules/cluster-manager';
import usePage from '@/composables/use-page';
import useTableSearchSelect, { ISearchSelectData }  from '@/composables/use-table-search-select';
import $i18n from '@/i18n/i18n-setup';

export interface IVPCItem {
  cloudID: CloudID,
  region: string,
  regionName: string,
  networkType: 'overlay' | 'underlay',
  vpcID: string,
  vpcName: string,
  available: string,
  extra: string,
  reservedIPNum: number,
  availableIPNum: number,
  overlay: {
    cidrs: string[],
    reservedIPNum: number,
    availableIPNum: number
  },
  underlay: {
    cidrs: string[],
    reservedIPNum: number,
    availableIPNum: number
  },
  businessID: string
};

export default defineComponent({
  name: 'VpcCniTable',
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: Boolean,
      default: false,
    },
    cloudId: {
      type: String,
      default: '',
    },
    region: {
      type: String,
      default: '',
    },
    networkType: {
      type: String,
      default: '',
    },
    businessId: {
      type: String,
      default: '',
    },
    isEdit: {
      type: Boolean,
      default: false,
    },
    disabledFn: {
      type: Function,
      default: () => false,
    },
  },
  setup(props, ctx) {
    const tableRef = ref<InstanceType<typeof Table> | null>(null);
    const localValue = ref('');
    const curRow = ref<IVPCItem | null>(null);

    // vpc列表

    const vpcList = ref<IVPCItem[]>([]);
    const vpcLoading = ref(false);
    const getVpcList = async () => {
      if (!props.cloudId || !props.region || !props.networkType || !props.businessId) {
        return;
      }
      handleResetPage();
      vpcLoading.value = true;
      const data = await cloudVpc({
        cloudID: props.cloudId,
        region: props.region,
        networkType: props.networkType,
        businessID: props.businessId,
        ...searchParams.value,
      }).catch(() => []);
      vpcList.value = data.filter(item => item.available === 'true');
      vpcLoading.value = false;
    };

    const {
      pagination,
      curPageData,
      pageChange,
      pageSizeChange,
      handleResetPage,
    } = usePage(vpcList);
    const searchSelectDataSource = computed<ISearchSelectData[]>(() => [
      {
        name: $i18n.t('generic.label.name'),
        id: 'vpcName',
      },
      {
        name: 'ID',
        id: 'vpcID',
      },
    ]);
    // searchSelect搜索框可选值
    const filteredValue = ref({});
    const {
      searchSelectData,
      searchSelectValue,
      handleClearSearchSelect,
    } = useTableSearchSelect({ searchSelectDataSource, filteredValue });
    const defaultParams = {
      sort: 'overlayAvaliableIPNum',
      order: 'desc',
    };
    const orderMap = {
      descending: 'desc',
      ascending: 'ase',
    };
    const searchParams = ref<Record<string, string>>({ ...defaultParams });
    const searchSelectChange = (list) => {
      searchParams.value = list.reduce((pre, item) => {
        pre[item.id] = item.values?.map?.(item => item.id)?.join?.(',');
        return pre;
      }, {});
      getVpcList();
    };
    function handleClear() {
      pagination.value.current = 1;
      handleClearSearchSelect();
      if (Object.keys(searchParams.value).length > 0) {
        searchParams.value = { ...defaultParams };
        getVpcList();
      }
    }

    function handleClickCell(row) {
      if (props.disabledFn?.(row)) return;
      curRow.value = row;
      localValue.value = row.vpcID;
    };

    /**
     * 确认选择
     */
    function handleConfirm() {
      ctx.emit('confirm', localValue.value, curRow.value);
      handleCancel();
    }

    /**
     * 取消选择
     */
    function handleCancel() {
      ctx.emit('change', false);
    }

    function handleValueBefore(val) {
      if (val) {
        return;
      }
      ctx.emit('change', false);
    }

    /**
     * 排序
     */
    function handleSortChange({ prop, order }) {
      searchParams.value = {
        sort: prop,
        order: orderMap[order],
      };
      getVpcList();
    }

    watch(() => [props.region, props.networkType], () => {
      getVpcList();
    });

    // tableRef 需动态获取
    watch(tableRef, () => {
      tableRef.value?.sort?.('overlayAvaliableIPNum', 'descending');
    });

    return {
      tableRef,
      localValue,
      vpcList,
      vpcLoading,
      pagination,
      curPageData,
      pageChange,
      pageSizeChange,
      handleResetPage,
      searchSelectData,
      searchSelectValue,
      handleClearSearchSelect,
      searchParams,
      searchSelectChange,
      handleClear,
      handleClickCell,
      handleConfirm,
      handleCancel,
      handleValueBefore,
      handleSortChange,
    };
  },
});
</script>
