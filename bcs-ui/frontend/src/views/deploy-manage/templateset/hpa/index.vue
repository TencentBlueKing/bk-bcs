<!-- eslint-disable vue/multi-word-component-names -->
<!-- eslint-disable max-len -->
<template>
  <BcsContent hide-back :title="$t('deploy.templateset.HPAManagement')">
    <div v-bkloading="{ isLoading: isInitLoading }">
      <Row class="mb-[16px]">
        <template #left>
          <div class="left">
            <bk-button @click.stop.prevent="removeHPAs">
              <span>{{$t('generic.button.batchDelete')}}</span>
            </bk-button>
          </div>
        </template>
        <template #right>
          <div class="right">
            <ClusterSelectComb
              :search.sync="searchKeyword"
              :cluster-id.sync="searchScope"
              cluster-type="all"
              @search-change="searchHPA"
              @refresh="refresh" />
          </div>
        </template>
      </Row>
      <div class="biz-hpa biz-table-wrapper">
        <bk-table
          v-bkloading="{ isLoading: isPageLoading && !isInitLoading }"
          :data="curPageData"
          :page-params="pageConf"
          @page-change="pageChangeHandler"
          @page-limit-change="changePageSize"
          @select="handlePageSelect"
          @select-all="handlePageSelectAll">
          <bk-table-column type="selection" width="60" :selectable="rowSelectable" />
          <bk-table-column :label="$t('generic.label.name')" prop="name" min-width="100">
            <template slot-scope="{ row }">
              {{row.name}}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.label.cluster')" prop="cluster_name" min-width="100">
            <template slot-scope="{ row }">
              {{row.cluster_name}}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('k8s.namespace')" prop="namespace" :show-overflow-tooltip="true" min-width="150" />
          <bk-table-column :label="$t('deploy.templateset.metricCurrentTarget')" prop="current_metrics_display" :show-overflow-tooltip="true" min-width="150">
          </bk-table-column>
          <bk-table-column width="30">
            <template slot-scope="{ row }">
              <i class="bcs-icon bcs-icon-info-circle" style="color: #ffb400;" v-if="row.needShowConditions" @click="showConditions(row, index)"></i>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('deploy.templateset.instanceNumCurrentRange')" prop="replicas" min-width="150">
            <template slot-scope="{ row }">
              {{ row.current_replicas }} / {{ row.min_replicas }}-{{ row.max_replicas }}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('deploy.templateset.associatedResources')" :show-overflow-tooltip="true" prop="deployment" min-width="150">
            <template slot-scope="{ row }">
              <bk-button
                :disabled="!['Deployment', 'StatefulSet'].includes(row.ref_kind)"
                text
                @click="handleGotoAppDetail(row)">
                <span class="bcs-ellipsis">{{row.ref_name}}</span>
              </bk-button>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('deploy.templateset.sources')" prop="source_type">
            <template slot-scope="{ row }">
              {{ row.source_type || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('cluster.labels.createdAt')" prop="create_time" min-width="100">
            <template slot-scope="{ row }">
              {{ row.create_time || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.label.createdBy')" prop="creator" min-width="100">
            <template slot-scope="{ row }">
              <bk-user-display-name :user-id="row.creator"></bk-user-display-name>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.label.action')" prop="permissions">
            <template slot-scope="{ row }">
              <div>
                <a href="javascript:void(0);" :class="['bk-text-button']" @click="removeHPA(row)">{{$t('generic.button.delete')}}</a>
              </div>
            </template>
          </bk-table-column>
          <template #empty>
            <BcsEmptyTableStatus :type="searchKeyword ? 'search-empty' : 'empty'" @clear="searchKeyword = ''" />
          </template>
        </bk-table>
      </div>

      <bk-dialog
        :is-show="batchDialogConfig.isShow"
        :width="430"
        :has-header="false"
        :quick-close="false"
        :title="$t('generic.title.confirmDelete')"
        @confirm="deleteHPAs(batchDialogConfig.data)"
        @cancel="batchDialogConfig.isShow = false">
        <template slot="content">
          <div class="biz-batch-wrapper">
            <p class="batch-title">{{$t('deploy.templateset.confirmDeleteFollowing')}} HPA？</p>
            <ul class="batch-list">
              <li v-for="(item, index) of batchDialogConfig.list" :key="index">{{item}}</li>
            </ul>
          </div>
        </template>
      </bk-dialog>

      <conditions-dialog
        :is-show="isShowConditions"
        :item="rowItem"
        @hide-update="hideConditionsDialog">
      </conditions-dialog>
    </div>
  </BcsContent>
</template>

<script>
import ConditionsDialog from './conditions-dialog';

import { catchErrorHandler } from '@/common/util';
import ClusterSelectComb from '@/components/cluster-selector/cluster-select-comb.vue';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';

export default {
  components: {
    ConditionsDialog,
    BcsContent,
    ClusterSelectComb,
    Row,
  },
  data() {
    return {
      isInitLoading: true,
      isPageLoading: false,
      curPageData: [],
      searchKeyword: '',
      searchScope: '',
      pageConf: {
        total: 1,
        totalPage: 1,
        pageSize: 10,
        curPage: 1,
        show: true,
      },
      alreadySelectedNums: 0,
      isBatchRemoving: false,
      curSelectedData: [],
      batchDialogConfig: {
        isShow: false,
        list: [],
        data: [],
      },
      isShowConditions: false,
      rowItem: null,
      hpaSelectedList: [],
    };
  },
  computed: {
    isEn() {
      return this.$store.state.isEn;
    },
    projectId() {
      return this.$route.params.projectId;
    },
    HPAList() {
      const list = this.$store.state.hpa.HPAList;
      return JSON.parse(JSON.stringify(list));
    },
    searchScopeList() {
      const { clusterList } = this.$store.state.cluster;
      const results = clusterList.map(item => ({
        id: item.cluster_id,
        name: item.name,
      }));

      return results;
    },
    curClusterId() {
      return this.$store.getters.curClusterId;
    },
  },
  watch: {
    searchScope() {
      this.init();
    },
    curClusterId() {
      this.searchScope = this.curClusterId;
      this.searchHPA();
    },
  },
  methods: {
    /**
             * 初始化入口
             */
    init() {
      this.initPageConf();
      this.getHPAList();
    },

    /**
             * 获取HPA 列表
             */
    async getHPAList() {
      try {
        this.isPageLoading = true;
        await this.$store.dispatch('hpa/getHPAList', {
          projectId: this.projectId,
          clusterId: this.searchScope,
        });

        this.initPageConf();
        this.curPageData = this.getDataByPage(this.pageConf.curPage);

        // 如果有搜索关键字，继续显示过滤后的结果
        if (this.searchScope || this.searchKeyword) {
          this.searchHPA();
        }
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
        this.isInitLoading = false;
        this.isPageLoading = false;
      }
    },

    /**
             * 每行的多选框点击事件
             */
    rowClick() {
      this.$nextTick(() => {
        this.alreadySelectedNums = this.HPAList.filter(item => item.isChecked).length;
      });
    },

    /**
             * 选择当前页数据
             */
    selectHPAs() {
      const list = this.curPageData;
      const selectList = list.filter(item => item.isChecked === true);
      this.curSelectedData.splice(0, this.curSelectedData.length, ...selectList);
    },

    /**
             * 刷新列表
             */
    refresh() {
      this.pageConf.curPage = 1;
      this.isPageLoading = true;
      this.getHPAList();
    },

    /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
    changePageSize(pageSize) {
      this.pageConf.pageSize = pageSize;
      this.pageConf.curPage = 1;
      this.initPageConf();
      this.pageChangeHandler();
    },

    /**
             * 单选
             * @param {array} selection 已经选中的行数
             * @param {object} row 当前选中的行
             */
    handlePageSelect(selection) {
      this.hpaSelectedList = selection;
    },

    /**
             * 全选
             */
    handlePageSelectAll(selection) {
      this.hpaSelectedList = selection;
    },

    /**
             * 搜索HPA
             */
    searchHPA() {
      const keyword = this.searchKeyword.trim();
      const keyList = ['name', 'cluster_name', 'creator'];
      let list = JSON.parse(JSON.stringify(this.$store.state.hpa.HPAList));
      const results = [];

      if (this.searchScope) {
        list = list.filter(item => item.cluster_id === this.searchScope);
      }

      list.forEach((item) => {
        item.isChecked = false;
        for (const key of keyList) {
          if (item[key]?.indexOf(keyword) > -1) {
            results.push(item);
            return true;
          }
        }
      });

      this.HPAList.splice(0, this.HPAList.length, ...results);
      this.pageConf.curPage = 1;
      this.initPageConf();
      this.curPageData = this.getDataByPage(this.pageConf.curPage);
    },

    /**
             * 初始化分页配置
             */
    initPageConf() {
      const total = this.HPAList.length;
      this.pageConf.total = total;
      this.pageConf.totalPage = Math.ceil(total / this.pageConf.pageSize);
      if (this.pageConf.curPage > this.pageConf.totalPage) {
        this.pageConf.curPage = this.pageConf.totalPage;
      }
    },

    /**
             * 获取分页数据
             * @param  {number} page 第几页
             * @return {object} data 数据
             */
    getDataByPage(page) {
      if (page < 1) {
        // eslint-disable-next-line no-multi-assign
        this.pageConf.curPage = page = 1;
      }
      let startIndex = (page - 1) * this.pageConf.pageSize;
      let endIndex = page * this.pageConf.pageSize;
      this.isPageLoading = true;
      if (startIndex < 0) {
        startIndex = 0;
      }
      if (endIndex > this.HPAList.length) {
        endIndex = this.HPAList.length;
      }
      setTimeout(() => {
        this.isPageLoading = false;
      }, 200);
      this.hpaSelectedList = [];
      return this.HPAList.slice(startIndex, endIndex);
    },

    /**
             * 页数改变回调
             * @param  {number} page 第几页
             */
    pageChangeHandler(page = 1) {
      this.pageConf.curPage = page;

      const data = this.getDataByPage(page);
      this.curPageData = data;
    },

    /**
             * 重新加载当面页数据
             */
    reloadCurPage() {
      this.initPageConf();
      if (this.pageConf.curPage > this.pageConf.totalPage) {
        this.pageConf.curPage = this.pageConf.totalPage;
      }
      this.curPageData = this.getDataByPage(this.pageConf.curPage);
    },

    /**
             * 清空当前页选择
             */
    clearSelectHPAs() {
      this.HPAList.forEach((item) => {
        item.isChecked = false;
      });
    },

    /**
             * 确认批量删除HPA
             */
    async removeHPAs() {
      const data = [];
      const names = [];

      if (this.hpaSelectedList.length) {
        this.hpaSelectedList.forEach((item) => {
          data.push({
            cluster_id: item.cluster_id,
            namespace: item.namespace,
            name: item.name,
          });
          names.push(item.name);
        });
      }
      if (!data.length) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.selectHPAToDelete'),
        });
        return false;
      }

      this.batchDialogConfig.list = names;
      this.batchDialogConfig.data = data;
      this.batchDialogConfig.isShow = true;
    },

    /**
             * 确认删除HPA
             * @param  {object} HPA HPA
             */
    async removeHPA(HPA) {
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const self = this;

      this.$bkInfo({
        title: this.$t('generic.title.confirmDelete'),
        clsName: 'biz-remove-dialog',
        content: this.$createElement('p', {
          class: 'biz-confirm-desc',
        }, `${this.$t('deploy.templateset.confirmDeleteHPA')}【${HPA.name}】？`),
        async confirmFn() {
          self.deleteHPA(HPA);
        },
      });
    },

    /**
             * 批量删除HPA
             * @param  {object} data HPAs
             */
    async deleteHPAs(data) {
      this.batchDialogConfig.isShow = false;
      this.isPageLoading = true;
      const { projectId } = this;

      try {
        await this.$store.dispatch('hpa/batchDeleteHPA', {
          projectId,
          params: { data },
        });

        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.delete'),
        });
        this.initPageConf();
        this.getHPAList();
      } catch (e) {
        // 4004，已经被删除过，但接口不能立即清除，防止重复删除
        if (e.code === 4004) {
          this.initPageConf();
          this.getHPAList();
        }
        this.$bkMessage({
          theme: 'error',
          delay: 8000,
          hasCloseIcon: true,
          message: e.message,
        });
        this.isPageLoading = false;
      }
    },

    /**
             * 删除HPA
             * @param {object} HPA HPA
             */
    async deleteHPA(HPA) {
      const { projectId } = this;
      const { namespace } = HPA;
      const clusterId = HPA.cluster_id;
      const { name } = HPA;
      this.isPageLoading = true;
      try {
        await this.$store.dispatch('hpa/deleteHPA', {
          projectId,
          clusterId,
          namespace,
          name,
        });

        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.delete'),
        });
        this.initPageConf();
        this.getHPAList();
      } catch (e) {
        catchErrorHandler(e, this);
        this.isPageLoading = false;
      }
    },

    /**
             * 显示 conditions 弹框
             *
             * @param {Object} item 当前行对象
             * @param {number} index 当前行对象索引
             *
             * @return {string} returnDesc
             */
    showConditions(item) {
      this.isShowConditions = true;
      this.rowItem = item;
    },

    /**
             * 关闭 conditions 弹框
             */
    hideConditionsDialog() {
      this.isShowConditions = false;
      setTimeout(() => {
        this.rowItem = null;
      }, 300);
    },

    rowSelectable(row) {
      return row.can_delete;
    },
    handleGotoAppDetail(row) {
      const kindMap = {
        Deployment: 'deploymentsInstanceDetail2',
        StatefulSet: 'statefulsetInstanceDetail2',
      };
      const location = this.$router.resolve({
        name: kindMap[row.ref_kind] || '404',
        params: {
          instanceName: row.ref_name,
          instanceNamespace: row.namespace,
          instanceCategory: row.ref_kind,
        },
        query: {
          cluster_id: row.cluster_id,
          name: row.ref_name,
          namespace: row.namespace,
        },
      });
      window.open(location.href);
    },
  },
};
</script>

<style scoped>
    @import './index.css';
</style>
