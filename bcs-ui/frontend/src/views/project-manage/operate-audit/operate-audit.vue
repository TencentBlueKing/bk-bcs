<!-- eslint-disable max-len -->
<template>
  <div class="biz-content">
    <Header hide-back :title="$t('projects.operateAudit.record')" />
    <div class="biz-content-wrapper" style="padding: 0;" v-bkloading="{ isLoading: isInitLoading, opacity: 0.1 }">
      <template v-if="!isInitLoading">
        <div class="biz-panel-header biz-operate-audit-query">
          <div class="left">
            <bk-selector
              :placeholder="$t('projects.operateAudit.opObjType')"
              :selected.sync="resourceTypeIndex"
              :list="resourceTypeList"
              :setting-key="'id'"
              :display-key="'name'"
              :allow-clear="true"
              @clear="resourceTypeClear">
            </bk-selector>
          </div>
          <div class="left">
            <bk-selector
              :placeholder="$t('projects.operateAudit.opType')"
              :selected.sync="activityTypeIndex"
              :list="activityTypeList"
              :setting-key="'id'"
              :display-key="'name'"
              :allow-clear="true"
              @clear="activityTypeClear">
            </bk-selector>
          </div>
          <div class="left">
            <bk-selector
              :placeholder="$t('generic.label.status')"
              :selected.sync="activityStatusIndex"
              :list="activityStatusList"
              :setting-key="'id'"
              :display-key="'name'"
              :allow-clear="true"
              @clear="activityStatusClear">
            </bk-selector>
          </div>
          <div class="left range-picker">
            <bk-date-picker
              :placeholder="$t('generic.placeholder.searchDate')"
              :shortcuts="shortcuts"
              :type="'datetimerange'"
              :placement="'bottom-end'"
              @change="change">
            </bk-date-picker>
          </div>
          <div class="left">
            <bk-button type="primary" :title="$t('generic.button.query')" icon="search" @click="handleClick">
              {{$t('generic.button.query')}}
            </bk-button>
          </div>
        </div>
        <bk-table
          v-bkloading="{ isLoading: isPageLoading && !isInitLoading }"
          ext-cls="biz-operate-audit-table"
          :data="dataList"
          :page-params="pageConf"
          size="medium"
          @page-change="pageChange"
          @page-limit-change="changePageSize">
          <bk-table-column :label="$t('generic.label.time')" prop="activityTime"></bk-table-column>
          <bk-table-column :label="$t('projects.operateAudit.opType')" prop="activityType"></bk-table-column>
          <bk-table-column :label="$t('projects.operateAudit.objType')">
            <template slot-scope="{ row }">
              <p class="extra-info lh20" :title="row.extra.resourceType || '--'">
                <span>{{$t('generic.label.type1')}}</span>{{row.extra.resourceType || '--'}}
              </p>
              <p class="extra-info lh20" :title="row.extra.resource || '--'">
                <span>{{$t('projects.operateAudit.obj')}}</span>{{row.extra.resource || '--'}}
              </p>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.label.status')">
            <template slot-scope="{ row }">
              <i class="bk-icon" :class="row.activity_status === 'completed' || row.activity_status === 'succeed' ? 'success icon-check-circle' : 'fail icon-close-circle'"></i>
              {{row.activityStatus}}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('projects.operateAudit.operator')" prop="user"></bk-table-column>
          <bk-table-column :label="$t('cluster.create.label.desc')" prop="description" :show-overflow-tooltip="true" min-width="160"></bk-table-column>
          <template #empty>
            <BcsEmptyTableStatus :type="searchEmpty ? 'search-empty' : 'empty'" @clear="handleClearSearchData" />
          </template>
        </bk-table>
      </template>
    </div>
  </div>
</template>

<script>
import Header from '@/components/layout/Header.vue';

export default {
  name: 'OperateAudit',
  components: { Header },
  data() {
    // 操作类型下拉框 list
    const activityTypeList = [
      { id: 'all', name: this.$t('generic.label.total') },
      { id: 'note', name: this.$t('projects.operateAudit.note') },
      { id: 'add', name: this.$t('generic.button.create') },
      { id: 'modify', name: this.$t('generic.button.update') },
      { id: 'delete', name: this.$t('generic.button.delete') },
      { id: 'begin', name: this.$t('projects.operateAudit.begin') },
      { id: 'end', name: this.$t('projects.operateAudit.end') },
      { id: 'start', name: this.$t('projects.operateAudit.start') },
      { id: 'pause', name: this.$t('projects.operateAudit.pause') },
      { id: 'carryon', name: this.$t('projects.operateAudit.continue') },
      { id: 'stop', name: this.$t('projects.operateAudit.stop') },
      { id: 'restart', name: this.$t('projects.operateAudit.restart') },
      { id: 'query', name: this.$t('generic.button.query') },
    ];
    // 操作类型 map
    const activityTypeMap = {};
    activityTypeList.forEach((item) => {
      activityTypeMap[item.id] = item.name;
    });

    // 状态下拉框 list
    const activityStatusList = [
      { id: 'all', name: this.$t('generic.label.total') },
      { id: 'unknown', name: this.$t('generic.status.unknown') },
      { id: 'completed', name: this.$t('generic.status.done') },
      { id: 'error', name: this.$t('projects.operateAudit.error') },
      { id: 'busy', name: this.$t('projects.operateAudit.doing') },
      { id: 'succeed', name: this.$t('generic.status.success') },
      { id: 'failed', name: this.$t('generic.status.failed') },
    ];
    // 操作状态 map
    const activityStatusMap = {};
    activityStatusList.forEach((item) => {
      activityStatusMap[item.id] = item.name;
    });
    // 服务类型
    let serviceType = 'container-service';
    if (this.$route.fullPath.indexOf('/bcs') === 0) {
      serviceType = 'container-service';
    } else if (this.$route.fullPath.indexOf('/monitor') === 0) {
      serviceType = 'monitor';
    }
    // 自定义页数
    const pageCountData = [
      { count: 10, id: '10' },
      { count: 20, id: '20' },
      { count: 50, id: '50' },
      { count: 100, id: '100' },
    ];
    return {
      // 操作类型 map
      activityTypeMap,
      // 操作类型下拉框 list
      activityTypeList,
      activityTypeIndex: -1,

      // 操作状态 map
      activityStatusMap,
      // 状态下拉框 list
      activityStatusList,
      activityStatusIndex: -1,

      // 操作对象类型下拉框 list
      resourceTypeList: [],
      // 操作对象类型下拉框 map
      resourceTypeMap: {},
      resourceTypeIndex: -1,

      // 查询时间范围
      dataRange: ['', ''],
      // 列表数据
      dataList: [],

      // 服务类型
      serviceType,

      isInitLoading: true,
      isPageLoading: false,

      pageConf: {
        // 总数
        total: 0,
        // 总页数
        totalPage: 1,
        // 每页多少条
        pageSize: 10,
        // 当前页
        curPage: 1,
        // 是否显示翻页条
        show: false,
      },
      // 自定义页数 对象
      pageCountList: pageCountData,
      pageCountListIndex: '10',
      bkMessageInstance: null,
      shortcuts: [
        {
          text: this.$t('units.time.today'),
          value() {
            const end = new Date();
            const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
            return [start, end];
          },
        },
        {
          text: this.$t('units.time.lastDays'),
          value() {
            const end = new Date();
            const start = new Date();
            start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
            return [start, end];
          },
        },
        {
          text: this.$t('units.time.last15Days'),
          value() {
            const end = new Date();
            const start = new Date();
            start.setTime(start.getTime() - 3600 * 1000 * 24 * 15);
            return [start, end];
          },
        },
        {
          text: this.$t('units.time.last30Days'),
          value() {
            const end = new Date();
            const start = new Date();
            start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
            return [start, end];
          },
        },
      ],
    };
  },
  computed: {
    isEn() {
      return this.$store.state.isEn;
    },
    searchEmpty() {
      return (this.resourceTypeIndex !== -1 && this.resourceTypeIndex)
      || (this.activityTypeIndex !== -1 && this.activityTypeIndex)
      || (this.activityStatusIndex !== -1 && this.activityStatusIndex)
      || this.dataRange.some(item => item);
    },
  },
  watch: {
    '$route.params.projectId': {
      handler: 'routerChangeHandler',
      immediate: true,
    },
  },
  mounted() {
  },
  destroyed() {
    this.bkMessageInstance?.close();
  },
  methods: {
    /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
    changePageSize(pageSize) {
      this.pageConf.pageSize = pageSize;
      this.pageConf.curPage = 1;
      this.pageChange();
    },
    /**
             * router change 回调(根据projectId变化更新数据)
             */
    async routerChangeHandler() {
      this.projId = this.$route.params.projectId || '000';
      this.fetchData({
        projId: this.projId,
        limit: this.pageConf.pageSize,
        offset: 0,
      });
      this.getResourceTypes();
    },

    /**
             * 获取所有的操作对象类型
             */
    async getResourceTypes() {
      try {
        const res = await this.$store.dispatch('mc/getResourceTypes', {
          serviceType: this.serviceType,
        });
        this.resourceTypeMap = res.data;
        this.resourceTypeList.splice(0, this.resourceTypeList.length);
        Object.keys(this.resourceTypeMap).forEach((key) => {
          this.resourceTypeList.push({
            id: key,
            name: this.resourceTypeMap[key],
          });
        });
        this.resourceTypeList.unshift({ id: 'all', name: this.$t('generic.label.total') });
      } catch (e) {
      }
    },

    /**
             * 获取表格数据
             *
             * @param {Object} params ajax 查询参数
             */
    async fetchData(params = {}) {
      // 操作类型
      // const activityType = this.activityTypeIndex !== -1
      //     ? this.activityTypeList[this.activityTypeIndex].id
      //     : null
      const activityType = this.activityTypeIndex === -1 ? null : this.activityTypeIndex;

      // 状态
      // const activityStatus = this.activityStatusIndex !== -1
      //     ? this.activityStatusList[this.activityStatusIndex].id
      //     : null
      const activityStatus = this.activityStatusIndex === -1 ? null : this.activityStatusIndex;

      // 操作对象类型
      // const resourceType = this.resourceTypeIndex !== -1
      //     ? this.resourceTypeList[this.resourceTypeIndex].id
      //     : null
      const resourceType = this.resourceTypeIndex === -1 ? null : this.resourceTypeIndex;

      // 开始结束时间
      const [beginTime, endTime] = this.dataRange;

      this.isPageLoading = true;
      try {
        const res = await this.$store.dispatch('mc/getActivityLogs', Object.assign({}, params, {
          activityType,
          activityStatus,
          resourceType,
          beginTime,
          endTime,
          serviceType: this.serviceType,
        }));

        this.dataList = [];

        const { count } = res;
        if (count <= 0) {
          this.pageConf.totalPage = 0;
          this.total = 0;
          this.pageConf.show = false;
          return;
        }

        this.pageConf.total = count;
        this.pageConf.totalPage = Math.ceil(count / this.pageConf.pageSize);
        if (this.pageConf.totalPage < this.pageConf.curPage) {
          this.pageConf.curPage = 1;
        }
        this.pageConf.show = true;

        const list = res.results || [];
        list.forEach((item) => {
          this.dataList.push({
            activity_status: item.activity_status,
            // 操作时间
            activityTime: item.activity_time,
            // 操作类型
            activityType: this.activityTypeMap[item.activity_type],
            extra: {
              // 操作对象类型
              resourceType: this.resourceTypeMap[item.resource_type] || item.resource_type,
              // 操作对象
              resource: item.resource,
            },
            // 状态
            activityStatus: this.activityStatusMap[item.activity_status],
            user: item.user,
            description: item.description,
          });
        });
      } catch (e) {
      } finally {
        this.isPageLoading = false;
        // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
        setTimeout(() => {
          this.isInitLoading = false;
        }, 200);
      }
    },

    /**
             * 翻页
             *
             * @param {number} page 页码
             */
    pageChange(page = 1) {
      this.pageConf.curPage = page;
      this.fetchData({
        projId: this.projId,
        limit: this.pageConf.pageSize,
        offset: this.pageConf.pageSize * (page - 1),
      });
    },

    /**
             * 清除操作对象类型
             */
    resourceTypeClear() {
      this.resourceTypeIndex = -1;
    },

    /**
             * 清除操作类型
             */
    activityTypeClear() {
      this.activityTypeIndex = -1;
    },

    /**
             * 清除状态
             */
    activityStatusClear() {
      this.activityStatusIndex = -1;
    },

    /**
             * 日期范围搜索条件
             *
             * @param {string} newValue 变化前的值
             */
    change(newValue) {
      this.dataRange = newValue;
    },

    /**
             * 搜索按钮点击
             *
             * @param {Object} e 事件对象
             */
    handleClick() {
      this.pageConf.curPage = 1;
      this.fetchData({
        projId: this.projId,
        limit: this.pageConf.pageSize,
        offset: 0,
      });
    },

    /**
             * 分页大小更改
             *
             * @param {Object} e 事件对象
             */
    handlePageSizeChange() {
      this.fetchData({
        projId: this.projId,
        limit: this.pageConf.pageSize,
        offset: 0,
      });
    },

    handleClearSearchData() {
      this.resourceTypeIndex = -1;
      this.activityTypeIndex = -1;
      this.activityStatusIndex = -1;
      this.dataRange = [];
      this.handleClick();
    },
  },
};
</script>

<style scoped>
    @import './operate-audit.css';
</style>
