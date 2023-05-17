<template>
  <LayoutContent :title="$t('日志采集规则')" :desc="cluster.clusterName">
    <div class="content-header mb20">
      <bcs-button icon="plus" theme="primary" @click="handleCreateLog">{{$t('新建规则')}}</bcs-button>
      <div class="right">
        <NamespaceSelect :cluster-id="clusterId" v-model="searchData.namespace" class="mw248"></NamespaceSelect>
        <bcs-select
          class="ml10 mw248"
          :placeholder="$t('应用类型')"
          v-model="searchData.kind">
          <bcs-option
            v-for="name in kinds"
            :key="name"
            :id="name"
            :name="name">
          </bcs-option>
        </bcs-select>
        <bcs-input
          class="ml10 mw300"
          :placeholder="$t('应用名')"
          right-icon="bk-icon icon-search"
          v-model="searchData.name"></bcs-input>
      </div>
    </div>
    <bcs-table
      :data="curPageData"
      :pagination="pagination"
      v-bkloading="{ isLoading: loading }"
      size="medium"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <bcs-table-column :label="$t('名称')">
        <template #default="{ row }">
          <span v-bk-tooltips="{ disabled: !row.deleted, content: $t('采集规则不存在') }">
            <bcs-button
              text
              :disabled="row.deleted"
              @click="handleShowDetail(row)"
            >
              <span class="bcs-ellipsis">{{row.name}}</span>
            </bcs-button>
          </span>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('集群')">
        <template #default>
          {{cluster.clusterName}}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('命名空间')" prop="namespace"></bcs-table-column>
      <bcs-table-column :label="$t('日志源')" width="100">
        <template #default="{ row }">
          {{logSourceTypeMap[row.config_selected]}}
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('选择器')" show-overflow-tooltip min-width="300">
        <template #default="{ row }">
          <template v-if="!row.deleted">
            <!-- 选择容器 -->
            <template v-if="row.config_selected === 'SelectedContainers'">
              <p>{{$t('类型')}}：{{row.config.workload.kind || '--'}}</p>
              <p>{{$t('名称')}}：{{row.config.workload.name || '--'}}</p>
            </template>
            <!-- 选择标签 -->
            <template v-else-if="row.config_selected === 'SelectedLabels'">
              <div class="row-label" v-if="Object.keys(row.config.label_selector.match_labels || {}).length">
                <div class="mb5">{{$t('匹配标签')}}：</div>
                <div class="row-label-content">
                  <span
                    v-for="key in Object.keys(row.config.label_selector.match_labels || {})"
                    :key="key"
                    class="tag mr5 mb5">
                    {{`${key} : ${row.config.label_selector.match_labels[key]}`}}
                  </span>
                </div>
              </div>
              <div class="row-label" v-if="(row.config.label_selector.match_expressions || []).length">
                <div class="mb5">{{$t('匹配表达式')}}：</div>
                <div class="row-label-content">
                  <span
                    v-for="item in row.config.label_selector.match_expressions"
                    :key="item.key"
                    class="tag mr5 mb5">
                    {{item.key || '--'}}
                    <span>{{item.operator || '--'}}</span>
                    {{item.values || '--'}}
                  </span>
                </div>
              </div>
            </template>
            <!-- 所有容器 -->
            <template v-else-if="row.config_selected === 'AllContainers'">
              --
            </template>
          </template>
          <template v-else>--</template>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('日志信息')" show-overflow-tooltip min-width="300">
        <template #default="{ row }">
          <template v-if="!row.deleted">
            <template v-if="row.config_selected === 'SelectedContainers'">
              <div
                v-for="(item, index) in row.config.workload.containers"
                :key="index"
                class="bcs-ellipsis">
                {{`${item.container_name}: ${(item.paths || []).join(';') || '--'}`}}
              </div>
            </template>
            <template v-else-if="row.config_selected === 'SelectedLabels'">
              {{row.config.label_selector.paths.join(';') || '--'}}
            </template>
            <template v-else-if="row.config_selected === 'AllContainers'">
              --
            </template>
          </template>
          <template v-else>--</template>
        </template>
      </bcs-table-column>
      <bcs-table-column :label="$t('更新人')" prop="updator" width="120"></bcs-table-column>
      <bcs-table-column :label="$t('更新时间')" prop="updated_at"></bcs-table-column>
      <bcs-table-column :label="$t('操作')" width="120">
        <template #default="{ row }">
          <span v-bk-tooltips="{ disabled: !row.deleted, content: $t('采集规则不存在') }">
            <bcs-button
              text
              :disabled="row.deleted"
              @click="handleUpdateLog(row)"
            >
              {{$t('更新')}}
            </bcs-button>
          </span>
          <bcs-button text class="ml10" @click="handleDeleteLog(row)">{{$t('删除')}}</bcs-button>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="searchEmpty ? 'search-empty' : 'empty'" @clear="handleClearSearchData" />
      </template>
    </bcs-table>
    <!-- 详情 -->
    <bcs-sideslider
      :is-show.sync="showDetail"
      :quick-close="true"
      :width="800"
      :title="currentRow ? currentRow.name : ''">
      <template #content>
        <LogDetail
          :log-source-type-map="logSourceTypeMap"
          :data="currentRow"
          :cluster="cluster">
        </LogDetail>
      </template>
    </bcs-sideslider>
    <!-- 编辑 -->
    <bcs-sideslider
      :is-show.sync="showEdit"
      :width="800"
      :title="currentRow ? $t('编辑规则') : $t('新建规则')">
      <template #content>
        <LogListEdit
          class="pd30"
          :kinds="kinds"
          :name="currentRow ? currentRow.name : null"
          :cluster-id="clusterId"
          @cancel="showEdit = false"
          @confirm="handleEditConfirm">
        </LogListEdit>
      </template>
    </bcs-sideslider>
  </LayoutContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue';
import LayoutContent from '@/components/layout/Content.vue';
import $store from '@/store';
import usePage from '@/composables/use-page';
import $i18n from '@/i18n/i18n-setup';
import LogListEdit from './log-list-edit.vue';
import LogDetail from './log-detail.vue';
import useLog from './use-log';
import NamespaceSelect from '@/components/namespace-selector/namespace-select.vue';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';

export default defineComponent({
  components: { LayoutContent, LogListEdit, LogDetail, NamespaceSelect },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const cluster = computed(() => ($store.state as any).cluster.clusterList
      .find(item => item.clusterID === props.clusterId) || {});
    const showDetail = ref(false);
    const showEdit = ref(false);
    const logSourceTypeMap = {
      SelectedContainers: $i18n.t('指定容器'),
      SelectedLabels: $i18n.t('指定标签'),
      AllContainers: $i18n.t('所有容器'),
    };
    const kinds = ref([
      'Deployment',
      'DaemonSet',
      'StatefulSet',
      'CronJob',
      'Job',
      'Pod',
      'GameStatefulSet',
      'GameDeployment',
    ]);

    const {
      handleGetLogRules,
      handleDeleteLogRule,
    } = useLog();
    // 日志规则列表
    const logList = ref<any[]>([]);
    const loading = ref(false);
    const handleGetLogList = async () => {
      loading.value = true;
      logList.value = await handleGetLogRules(props.clusterId);
      loading.value = false;
    };
    const searchData = ref({
      namespace: '',
      kind: '',
      name: '',
    });
    const searchEmpty = computed(() => Object.keys(searchData.value).some(key => !!searchData.value[key]));

    const filterLogList = computed(() =>
    // 搜索
      logList.value.filter(item => (!searchData.value.namespace || item.namespace === searchData.value.namespace)
                      && (!searchData.value.name || item.name.includes(searchData.value.name))
                      && (!searchData.value.kind || item.config.workload?.kind === searchData.value.kind)));
    const {
      pageChange,
      pageSizeChange,
      curPageData,
      pagination,
    } = usePage(filterLogList);

    // 操作
    const currentRow = ref<any>(null);
    const handleCreateLog = () => {
      currentRow.value = null;
      showEdit.value = true;
    };
    const handleShowDetail = (row) => {
      currentRow.value = row;
      showDetail.value = true;
    };
    const handleUpdateLog = (row) => {
      currentRow.value = row;
      showEdit.value = true;
    };
    const handleDeleteLog = (row) => {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('确认删除规则 {name}', { name: row.name }),
        subTitle: row.name,
        defaultInfo: true,
        confirmFn: async () => {
          const result = await handleDeleteLogRule({
            $name: row.name,
            $clusterId: props.clusterId,
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('删除成功'),
            });
            handleGetLogList();
          }
        },
      });
    };
    const handleEditConfirm = () => {
      showEdit.value = false;
      handleGetLogList();
    };

    const handleClearSearchData = () => {
      Object.keys(searchData.value).forEach(key => searchData.value[key] = '');
    };

    onMounted(() => {
      handleGetLogList();
    });
    return {
      searchEmpty,
      handleClearSearchData,
      currentRow,
      searchData,
      cluster,
      showEdit,
      showDetail,
      kinds,
      loading,
      pageChange,
      pageSizeChange,
      curPageData,
      pagination,
      handleShowDetail,
      logSourceTypeMap,
      handleUpdateLog,
      handleDeleteLog,
      handleCreateLog,
      handleEditConfirm,
    };
  },
});
</script>
<style lang="postcss" scoped>
.content-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  .right {
      display: flex;
      background: #fff;
  }
  .mw248 {
      min-width: 248px;
  }
  .mw300 {
      min-width: 300px;
  }
}
.pd30 {
  padding: 30px;
}
.row-label {
  overflow: hidden;
  margin: 5px 0;
  &-content {
      display: flex;
      flex-wrap: wrap;
      .tag {
          background-color: rgba(151,155,165,.1);
          border-color: rgba(220,222,229,.6);
          color: #63656e;
          display: inline-block;
          font-size: 12px;
          padding: 0 10px;
          min-height: 22px;
          cursor: default;
          box-sizing: border-box;
          line-height: 22px;
      }
  }
}
</style>
