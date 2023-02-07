<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <div
    ref="terminal"
    :class="['bk-dropdown-menu biz-terminal active', { 'active': isActive }]"
    v-if="curProject && clusterList.length && !isSharedCluster"
    @mousedown="handlerMousedown"
    @mouseover="handlerMouseover"
    @mouseout="handlerMouseout">
    <div class="bk-dropdown-trigger">
      <div class="flex items-center h-10 shadow bg-white cursor-move px-2.5">
        <img src="@/images/terminal.svg">
        <span class="text-sm ml-2">WebConsole</span>
        <a
          href="https://bk.tencent.com/docs/document/7.0/173/14130"
          target="_blank"
          class="text-sm ml-2">
          <i class="bcs-icon bcs-icon-helper"></i>
        </a>
      </div>
      <transition name="fade">
        <div
          :class="['bk-dropdown-content is-show']"
          ref="terminalContent" style="bottom: 35px; right: 0; position: absolute;" v-show="isShow">
          <div class="search-box">
            <bkbcs-input
              v-model="keyword"
              :placeholder="$t('请输入集群名或ID')"
              @focus="isFocus = true"
              @blur="isFocus = false">
            </bkbcs-input>
          </div>
          <ul class="bk-dropdown-list" v-if="searchList.length">
            <li>
              <a
                href="javascript:;"
                v-for="(cluster, index) in searchList" :key="index"
                @click="goWebConsole(cluster)" style="padding: 0 10px;" :title="cluster.name">{{cluster.name}}</a>
            </li>
          </ul>
          <div v-else class="no-result">
            {{$t('无数据')}}
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex';
export default {
  data() {
    return {
      PROJECT_K8S: window.PROJECT_K8S,
      PROJECT_TKE: window.PROJECT_TKE,
      terminalWins: null,
      isActive: false,
      isShow: false,
      showTimer: 0,
      activeTimer: 0,
      keyword: '',
      isFocus: false,
    };
  },
  computed: {
    searchList() {
      const clusterList = [...this.$store.state.cluster.clusterList];
      const list = clusterList.filter((item) => {
        const name = item.name.toLowerCase();
        const clusterId = item.cluster_id.toLowerCase();
        const keyword = this.keyword.toLowerCase();
        return name.indexOf(keyword) > -1 || clusterId.indexOf(keyword) > -1;
      });

      return list;
    },
    clusterList() {
      return [...this.$store.state.cluster.clusterList];
    },
    curProject() {
      return this.$store.state.curProject;
    },
    projectId() {
      return this.$route.params.projectId;
    },
    routeName() {
      return this.$route.name;
    },
    ...mapGetters('cluster', ['isSharedCluster']),
  },
  methods: {
    handlerMousedown(event) {
      const { terminal } = this.$refs;
      const { terminalContent } = this.$refs;
      const e = event || window.event;
      const cursorX = e.pageX - terminal.offsetLeft;
      const cursorY = e.pageY - terminal.offsetTop;

      document.onmousemove = function (event) {
        const e = event || window.event;
        terminal.style.top = `${e.pageY - cursorY}px`;
        terminal.style.left = `${e.pageX - cursorX}px`;
        terminalContent.style.bottom = '35px';
        terminalContent.style.left = '0';
      };

      document.onmouseup = function () {
        document.onmousemove = null;
        document.onmouseup = null;
      };
    },
    handlerMouseover() {
      clearTimeout(this.showTimer);
      clearTimeout(this.activeTimer);
      this.isActive = true;
      if (!this.isFocus) {
        this.keyword = '';
      }
      this.showTimer = setTimeout(() => {
        this.isShow = true;
      }, 200);
    },
    handlerMouseout() {
      clearTimeout(this.showTimer);
      clearTimeout(this.activeTimer);
      if (this.isFocus) {
        return false;
      }
      this.showTimer = setTimeout(() => {
        this.isShow = false;
      }, 100);
      this.activeTimer = setTimeout(() => {
        this.isActive = false;
      }, 400);
    },
    async goWebConsole(cluster) {
      const clusterId = cluster.cluster_id;
      const url = `${window.BCS_API_HOST}/bcsapi/v4/webconsole/projects/${this.projectId}/mgr/#cluster=${clusterId}`;

      this.keyword = '';
      this.isShow = false;
      if (this.terminalWins) {
        if (!this.terminalWins.closed) {
          this.terminalWins.postMessage({
            clusterId,
            clusterName: cluster.name,
          }, location.origin);
          this.terminalWins.focus();
        } else {
          this.terminalWins = window.open(url, '');
        }
      } else {
        this.terminalWins = window.open(url, '');
      }
    },
  },
};
</script>

<style scoped lang="postcss">
    @import "@/css/mixins/ellipsis.css";

    .biz-terminal {
        position: fixed;
        right: 10px;
        bottom: 10px;
        width: 200px;
        z-index: 1900;
        .bk-dropdown-content, .bk-dropdown-trigger {
            width: 200px;
        }
        .bk-dropdown-list {
            overflow: scroll;
            overflow-x: hidden;
            > li {
                width: 200px;
                a {
                    display: block;
                    vertical-align: middle;
                    width: 200px;
                    @mixin ellipsis 200px;
                }
            }
        }

        .no-result {
            line-height: 82px;
            text-align: center;
            font-size: 14px;
        }

        .search-box {
            padding: 10px;
            border-bottom: 1px solid #eee;
        }
    }
</style>
