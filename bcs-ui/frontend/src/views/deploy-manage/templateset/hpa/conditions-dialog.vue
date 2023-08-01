<!-- eslint-disable max-len -->
<template>
  <bk-dialog
    :is-show.sync="isVisible"
    :width="1050"
    :title="title"
    :close-icon="true"
    :ext-cls="'biz-rolling-update-dialog'"
    :quick-close="false"
    @confirm="confirm"
    @cancel="hide">
    <template slot="content">
      <div class="gamestatefulset-update-wrapper" v-bkloading="{ isLoading: isLoading, opacity: 1 }">
        <div class="biz-namespace">
          <bk-table
            :size="'medium'"
            :data="curPageData"
            :pagination="pageConf"
            :key="isVisible"
            v-bkloading="{ isLoading: isPageLoading }"
            @page-limit-change="handlePageLimitChange"
            @page-change="handlePageChange">
            <bk-table-column label="Type" :show-overflow-tooltip="true" width="200">
              <template slot-scope="props">
                {{props.row.type}}
              </template>
            </bk-table-column>
            <bk-table-column label="Status" :show-overflow-tooltip="true" width="200">
              <template slot-scope="props">
                <i v-if="props.row.statusStr === 'false'" class="bcs-icon bcs-icon-close-circle" style="color: #ff5656; margin-left: 12px;"></i>
                <i v-else class="bcs-icon bcs-icon-check-circle" style="color: #30d878; margin-left: 12px;"></i>
              </template>
            </bk-table-column>
            <bk-table-column label="Reason" :show-overflow-tooltip="true" width="200">
              <template slot-scope="props">
                {{props.row.reason}}
              </template>
            </bk-table-column>
            <bk-table-column label="LastTransitionTime" :show-overflow-tooltip="true" width="200">
              <template slot-scope="props">
                {{props.row.lastTransitionTime}}
              </template>
            </bk-table-column>
            <bk-table-column label="Message" :show-overflow-tooltip="true" width="200">
              <template slot-scope="props">
                {{props.row.message}}
              </template>
            </bk-table-column>
          </bk-table>
        </div>
      </div>
    </template>
    <div slot="footer">
      <div class="bk-dialog-outer">
        <template>
          <bk-button type="button" @click="hide">
            {{$t('generic.button.close')}}
          </bk-button>
        </template>
      </div>
    </div>
  </bk-dialog>
</template>

<script>
export default {
  components: {
  },
  props: {
    isShow: {
      type: Boolean,
      default: false,
    },
    item: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      title: '',
      isVisible: false,
      isLoading: true,
      renderItem: {},
      isPageLoading: false,
      pageConf: {
        total: 1,
        totalPage: 1,
        limit: 5,
        current: 1,
        count: 1,
        show: true,
      },
      conditionsList: [],
      curPageData: [],
    };
  },
  computed: {
    projectId() {
      return this.$route.params.projectId;
    },
    isEn() {
      return this.$store.state.isEn;
    },
  },
  watch: {
    isShow: {
      async handler(newVal) {
        this.isVisible = newVal;
        if (!this.isVisible) {
          return;
        }
        this.renderItem = Object.assign({}, this.item || {});
        this.title = this.renderItem.name;
        const list = this.renderItem.conditions || [];
        list.forEach((item) => {
          item.statusStr = item.status.toLowerCase();
        });

        this.conditionsList.splice(0, this.conditionsList.length, ...list);
        this.initPageConf();
        this.curPageData = this.getDataByPage(this.pageConf.current);
        setTimeout(() => {
          this.isLoading = false;
        }, 300);
      },
      immediate: true,
    },
  },
  methods: {
    /**
             * 初始化弹层翻页条
             */
    initPageConf() {
      const total = this.conditionsList.length;
      this.pageConf.count = total;
      this.pageConf.current = 1;
      this.pageConf.totalPage = Math.ceil(total / this.pageConf.limit) || 1;
    },

    /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
    handlePageLimitChange(pageSize) {
      this.pageConf.limit = pageSize;
      this.pageConf.current = 1;
      this.initPageConf();
      this.handlePageChange(this.pageConf.current);
    },

    /**
             * 翻页回调
             *
             * @param {number} page 当前页
             */
    handlePageChange(page) {
      this.pageConf.current = page;
      const data = this.getDataByPage(page);
      this.curPageData.splice(0, this.curPageData.length, ...data);
      this.isPageLoading = true;
      setTimeout(() => {
        this.isPageLoading = false;
      }, 100);
    },

    /**
             * 获取当前这一页的数据
             *
             * @param {number} page 当前页
             *
             * @return {Array} 当前页数据
             */
    getDataByPage(page) {
      // 如果没有page，重置
      if (!page) {
        // eslint-disable-next-line no-multi-assign
        this.pageConf.current = page = 1;
      }
      let startIndex = (page - 1) * this.pageConf.limit;
      let endIndex = page * this.pageConf.limit;
      this.isPageLoading = true;
      if (startIndex < 0) {
        startIndex = 0;
      }
      if (endIndex > this.conditionsList.length) {
        endIndex = this.conditionsList.length;
      }
      setTimeout(() => {
        this.isPageLoading = false;
      }, 200);
      return this.conditionsList.slice(startIndex, endIndex);
    },

    async confirm() {
      console.error('confirm');
    },

    hide() {
      this.isLoading = true;
      this.$emit('hide-update');
    },
  },
};
</script>

<style scoped>
    @import './conditions-dialog.css';
</style>
