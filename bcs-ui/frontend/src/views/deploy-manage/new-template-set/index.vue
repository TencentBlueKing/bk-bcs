<template>
  <BcsContent :title="$t('templateSet.title.set')" hide-back>
    <Row class="mb-[16px]">
      <template #left>
        <bcs-button theme="primary" class="min-w-[88px]" @click="addTemplateSet">
          {{ $t('generic.button.create') }}
        </bcs-button>
        <bcs-button>{{ $t('generic.button.import') }}</bcs-button>
        <bcs-button>{{ $t('generic.button.export') }}</bcs-button>
      </template>
      <template #right>
        <bcs-input
          :placeholder="$t('templateSet.placeholder.searchSet')"
          class="w-[420px]"
          clearable
          right-icon="bk-icon icon-search"
          v-model.trim="searchValue">
        </bcs-input>
        <span class="flex items-center ml-[8px]">
          <bcs-icon
            :class="[
              'bcs-icon-btn bcs-icon bcs-icon-lie',
              'flex items-center justify-center w-[32px] h-[32px] border-[1px] border-[#C4C6CC]',
              activeType === 'list' ? 'active bg-[#E1ECFF] !border-[#3A84FF] !text-[#3A84FF] z-10' : ''
            ]"
            type=""
            v-bk-tooltips="$t('templateSet.tips.listView')"
            @click="handleChangeType('list')" />
          <bcs-icon
            :class="[
              'bcs-icon-btn bcs-icon bcs-icon-kuai ml-[-1px]',
              'flex items-center justify-center w-[32px] h-[32px] border-[1px] border-[#C4C6CC]',
              activeType === 'card' ? 'active bg-[#E1ECFF] !border-[#3A84FF] !text-[#3A84FF] z-10' : ''
            ]"
            type=""
            v-bk-tooltips="$t('templateSet.tips.cardView')"
            @click="handleChangeType('card')" />
        </span>
      </template>
    </Row>
    <template v-if="activeType === 'list'">
      <bcs-table
        v-bkloading="{ isLoading: loading }"
        :data="curPageData"
        :pagination="pagination"
        @page-change="pageChange"
        @page-limit-change="pageSizeChange">
        <bcs-table-column type="selection" width="60"></bcs-table-column>
        <bcs-table-column :label="$t('templateSet.label.name')" prop="name">
          <template #default="{ row }">
            <bcs-button text @click="goDetail(row)">{{ row.name }}</bcs-button>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('templateSet.label.setVersion')" prop="latestVersion"></bcs-table-column>
        <bcs-table-column :label="$t('templateSet.label.label')"></bcs-table-column>
        <bcs-table-column :label="$t('templateSet.label.key')" prop="key"></bcs-table-column>
        <bcs-table-column :label="$t('generic.label.updator')" prop="updateBy">
          <template #default="{ row }">
            <bk-user-display-name v-if="row.updateBy" :user-id="row.updateBy">
            </bk-user-display-name>
            <span v-else>--</span>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('generic.label.updatedAt')" prop="updateTime"></bcs-table-column>
        <bcs-table-column :label="$t('generic.label.action')" width="160">
          <template #default="{ row }">
            <bcs-button text class="mr-[5px]" @click="editTemplateSet(row)">{{ $t('generic.button.edit') }}</bcs-button>
            <bcs-button text @click="deployTemplateSet(row)">{{ $t('templateSet.button.deploy') }}</bcs-button>
            <PopoverSelector>
              <span class="bcs-icon-more-btn"><i class="bcs-icon bcs-icon-more"></i></span>
              <template #content>
                <ul>
                  <li
                    class="bcs-dropdown-item"
                    @click="handleShowVersionList(row)">
                    {{ $t('templateSet.button.versionManage') }}
                  </li>
                </ul>
              </template>
            </PopoverSelector>
          </template>
        </bcs-table-column>
      </bcs-table>
    </template>
    <template v-else-if="activeType === 'card'">
      <div class="grid grid-cols-3 gap-[16px]" v-bkloading="{ isLoading: loading }">
        <div
          v-for="item in chartList"
          :key="item.name"
          class="bcs-border group relative bg-[#fff] rounded-sm cursor-pointer"
          @click="goDetail(item)">
          <div
            :class="[
              'absolute left-0 top-0 opacity-0',
              'group-hover:opacity-100',
              'w-0 h-0 border-solid border-t-[#DCDEE5] border-t-[50px] border-r-[50px] border-r-[transparent]'
            ]" @click.stop></div>
          <i
            :class="[
              'absolute opacity-0 left-[8px] top-[8px] text-[14px] text-[#fff] z-10 font-bold',
              'group-hover:opacity-100',
              'bcs-icon bcs-icon-check-1'
            ]" @click.stop></i>
          <PopoverSelector class="absolute right-[12px] top-[12px]">
            <span class="bcs-icon-more-btn"><i class="bcs-icon bcs-icon-more"></i></span>
            <template #content>
              <ul>
                <li class="bcs-dropdown-item" @click="handleShowVersionList(item)">
                  {{ $t('templateSet.button.versionManage') }}
                </li>
              </ul>
            </template>
          </PopoverSelector>
          <div class="flex items-center h-[84px] px-[24px] py-[16px]">
            <i class="bcs-icon bcs-icon-templateset text-[#979BA5] text-[36px]"></i>
            <div class="ml-[16px]">
              <div class="leading-[22px] mb-[6px]">
                <span class="text-[#313238] text-[14px]">{{ item.name }}</span>
                <span class="ml-[16px] text-[12px] text-[#979BA5]">
                  {{ item.latestVersion }}
                </span>
              </div>
              <div>
                <bcs-tag class="!ml-[0px]">Game</bcs-tag>
                <bcs-tag>Test</bcs-tag>
              </div>
            </div>
          </div>
          <div class="bcs-border-top h-[72px] bg-[#FAFBFD] px-[24px] py-[16px] text-[12px] leading-[20px]">
            {{ item.latestDescription }}
          </div>
        </div>
      </div>
    </template>
    <!-- 版本管理 -->
    <bcs-sideslider
      :is-show.sync="showVersionList"
      quick-close
      :title="$t('templateSet.title.versionManage')"
      :width="960">
      <template #content>
        <VersionList :name="curRow?.name" :repo-name="projectCode" class="px-[24px] py-[20px]" />
      </template>
    </bcs-sideslider>
  </BcsContent>
</template>
<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue';

import VersionList from './version-list.vue';

import { HelmManagerService } from '@/api/modules/new-helm-manager';
// import $bkMessage from '@/common/bkmagic';
// import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import usePage from '@/composables/use-page';
import useSearch from '@/composables/use-search';
// import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

// 视图类型
const activeType = ref<'list'|'card'>('list');

const handleChangeType = (type: 'list'|'card') => {
  activeType.value = type;
};

// chart数据
const loading = ref(false);
const chartList = ref<Chart[]>([]);
const keys = ref(['name', 'latestVersion', 'key', 'updateBy']);
const { searchValue, tableDataMatchSearch } = useSearch(chartList, keys);
const {
  pagination,
  curPageData,
  pageChange,
  pageSizeChange,
} = usePage<Chart>(tableDataMatchSearch);

const projectCode = computed(() => $store.getters.curProjectCode);
const listChartV1 = async () => {
  loading.value = true;
  const { data = [] } = await HelmManagerService.ListChartV1({
    $repoName: projectCode.value,
    name: '',
    page: 1,
    size: 99999, // todo 支持获取全量数据
  });
  chartList.value = data;
  loading.value = false;
};

// 新建模板集
const addTemplateSet = () => {
  $router.push({
    name: 'createTemplatesetV2',
  });
};

// 当前操作行
const curRow = ref<Chart>();

// 详情
const goDetail = (row: Chart) => {
  console.log(row);
};

// 编辑模板集
const editTemplateSet = (row: Chart) => {
  console.log(row);
};

// 去部署
const deployTemplateSet = (row: Chart) => {
  console.log(row);
};

// 版本列表
const showVersionList  = ref(false);
const handleShowVersionList = (row) => {
  curRow.value = row;
  showVersionList.value = true;
};

onBeforeMount(() => {
  listChartV1();
});
</script>
