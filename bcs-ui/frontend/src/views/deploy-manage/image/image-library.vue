<template>
  <div class="biz-content">
    <div class="biz-content-wrapper" style="padding: 0;" v-bkloading="{ isLoading: isInitLoading, opacity: 0.1 }">
      <template v-if="!isInitLoading">
        <div class="biz-panel-header biz-image-library-query !p-[0]">
          <div class="right">
            <div class="biz-search-input" style="width: 300px;">
              <bk-input
                clearable
                :placeholder="$t('deploy.image.search')"
                v-model="searchKey"
                right-icon="bk-icon icon-search"
                @change="handleSearch">
              </bk-input>
            </div>
          </div>
        </div>

        <div class="biz-image-library-list" v-bkloading="{ isLoading: isPageLoading && !isInitLoading }">
          <template v-if="dataList.length">
            <div class="list-item" v-for="(item, index) in dataList" :key="index">
              <div class="left-wrapper">
                <img src="@/images/default_logo.jpg" class="logo" />
              </div>
              <div class="right-wrapper">
                <div class="content">
                  <div class="info">
                    <div class="title">
                      <span>{{item.name}}</span>
                    </div>
                    <div class="attr">
                      <span>{{$t('generic.label.type1')}}{{item.type || '--'}}</span>
                      <span>{{$t('deploy.image.sources')}}{{item.deployBy || '--'}}</span>
                    </div>
                    <div class="desc">
                      {{$t('plugin.tools._intro')}}{{item.desc || '--'}}
                    </div>
                  </div>
                  <div class="detail" @click="toImageDetail(item)">
                    {{$t('deploy.image.detail')}}<i class="bcs-icon bcs-icon-angle-right"></i>
                  </div>
                </div>
              </div>
            </div>
          </template>
          <template v-else-if="!isPageLoading">
            <bcs-exception type="empty" scene="part"></bcs-exception>
          </template>
          <template v-else>
            <div class="loading"></div>
          </template>
        </div>
        <template v-if="pageConf.total">
          <bk-pagination
            ext-cls="m20"
            align="right"
            :current.sync="pageConf.curPage"
            :count="pageConf.total"
            :show-total-count="true"
            @change="pageChange"
            @limit-change="changePageSize">
          </bk-pagination>
        </template>
      </template>
    </div>
  </div>
</template>

<script>
import { throttle } from '@/common/util';
export default {
  data() {
    return {
      winHeight: 0,
      isInitLoading: true,
      isPageLoading: false,
      // 查询条件
      searchKey: '',
      // 列表数据
      // dataList: [],
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
      bkMessageInstance: null,
      handleSearch: null,
    };
  },
  computed: {
    projectId() {
      return this.$route.params.projectId;
    },
    projectCode() {
      return this.$route.params.projectCode;
    },
    dataList() {
      return this.$store.state.depot.imageLibrary.dataList;
    },
    isEn() {
      return this.$store.state.isEn;
    },
  },
  mounted() {
    this.winHeight = window.innerHeight;
    this.projId = this.$route.params.projectId || '000';
    localStorage.removeItem('backRouterName');
    this.getFirstPage();
    this.handleSearch = throttle(this.getFirstPage, 400);
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
             * 搜索框清除事件
             */
    clearSearch() {
      this.searchKey = '';
      this.handleClick();
    },

    /**
             * 跳转到镜像详情
             *
             * @param {Object} item 当前镜像对象
             */
    toImageDetail(item) {
      localStorage.setItem('backRouterName', 'imageLibrary');
      this.$store.commit('depot/forceUpdateCurImage', item);
      this.$router.push({
        name: 'imageDetail',
        params: {
          imageRepo: item.repo,
        },
      });
    },

    /**
             * 获取数据
             *
             * @param {Object} params ajax 查询参数
             */
    async fetchData(params = {}) {
      this.isPageLoading = true;

      const search = this.searchKey;

      try {
        const res = await this.$store.dispatch('depot/getImageLibrary', Object.assign({}, params, { search }));

        const count = res.count || 0;
        this.pageConf.total = count;
        this.pageConf.totalPage = Math.ceil(count / this.pageConf.pageSize);
        if (this.pageConf.totalPage < this.pageConf.curPage) {
          this.pageConf.curPage = 1;
        }
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
             * 首页
             */
    getFirstPage() {
      this.fetchData({
        limit: this.pageConf.pageSize,
        projId: this.projId,
        offset: 0,
      });
    },

    /**
             * 翻页
             *
             * @param {number} page 页码
             */
    pageChange(page = 1) {
      this.fetchData({
        limit: this.pageConf.pageSize,
        offset: this.pageConf.pageSize * (page - 1),
      });
    },

    /**
             * 搜索按钮点击
             *
             * @param {Object} e 对象
             */
    handleClick() {
      this.getFirstPage();
    },
  },
};
</script>

<style scoped>
    @import './image-library.css';
</style>
