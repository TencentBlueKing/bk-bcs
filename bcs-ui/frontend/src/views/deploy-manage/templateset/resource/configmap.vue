<!-- eslint-disable vue/multi-word-component-names -->
<!-- eslint-disable max-len -->
<template>
  <BcsContent hide-back title="ConfigMaps" :desc="$t('deploy.templateset.createFromTemplateOrHelmConfigMap')">
    <div v-bkloading="{ isLoading: isInitLoading }">

      <Row class="mb-[16px]">
        <div class="left" slot="left">
          <bk-button @click.stop.prevent="removeConfigmaps" v-if="curPageData.length">
            <span>{{$t('generic.button.batchDelete')}}</span>
          </bk-button>
        </div>
        <div class="right" slot="right">
          <ClusterSelectComb
            :placeholder="$t('deploy.templateset.searchNameOrNamespaceEnter')"
            :search.sync="searchKeyword"
            :cluster-id.sync="searchScope"
            cluster-type="all"
            @search-change="searchConfigmap"
            @refresh="refresh" />
        </div>
      </Row>

      <div class="biz-resource biz-table-wrapper">
        <bk-table
          v-bkloading="{ isLoading: isPageLoading && !isInitLoading }"
          class="biz-resource-table"
          :data="curPageData"
          :page-params="pageConf"
          @page-change="pageChangeHandler"
          @page-limit-change="changePageSize"
          @select="handlePageSelect"
          @select-all="handlePageSelectAll">
          <bk-table-column type="selection" width="60" :selectable="rowSelectable" />
          <bk-table-column :label="$t('generic.label.name')" prop="name" :show-overflow-tooltip="true" min-width="150">
            <template slot-scope="{ row }">
              <a
                class="bk-text-button biz-resource-title biz-text-wrapper"
                href="javascript: void(0)"
                v-authority="{
                  clickable: webAnnotations.perms[row.iam_ns_id]
                    && webAnnotations.perms[row.iam_ns_id].namespace_scoped_view,
                  actionId: 'namespace_scoped_view',
                  resourceName: row.namespace,
                  disablePerms: true,
                  permCtx: {
                    project_id: projectId,
                    cluster_id: row.cluster_id,
                    name: row.namespace
                  }
                }"
                @click.stop.prevent="showConfigmapDetail(row)">
                {{row.resourceName}}
              </a>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.label.cluster1')" prop="cluster_name" min-width="150">
            <template slot-scope="{ row }">
              <bcs-popover :content="row.cluster_id || '--'" placement="top">
                <p class="biz-text-wrapper">{{curSelectedCluster.name || '--'}}</p>
              </bcs-popover>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('k8s.namespace')" prop="namespace" min-width="100" />
          <bk-table-column :label="$t('deploy.templateset.sources')" prop="source_type" min-width="100">
            <template slot-scope="{ row }">
              {{ row.source_type || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('cluster.labels.createdAt')" prop="createTime" min-width="150">
            <template slot-scope="{ row }">
              {{ formatDate(row.createTime) || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('cluster.labels.updatedAt')" prop="update_time" min-width="150">
            <template slot-scope="{ row }">
              {{ formatDate(row.update_time) || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.label.updator')" prop="updator">
            <template slot-scope="{ row }">
              {{row.updator || '--'}}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('generic.label.action')" prop="permissions" min-width="150">
            <template slot-scope="{ row }">
              <li>
                <span
                  v-if="row.can_update"
                  v-authority="{
                    clickable: webAnnotations.perms[row.iam_ns_id]
                      && webAnnotations.perms[row.iam_ns_id].namespace_scoped_update,
                    actionId: 'namespace_scoped_update',
                    resourceName: row.namespace,
                    disablePerms: true,
                    permCtx: {
                      project_id: projectId,
                      cluster_id: row.cluster_id,
                      name: row.namespace
                    }
                  }"
                  @click.stop="updateConfigmap(row)"
                  class="biz-operate"
                >{{$t('generic.button.update')}}</span>
                <bcs-popover :content="row.can_update_msg" v-else placement="left">
                  <span class="biz-not-operate">{{$t('generic.button.update')}}</span>
                </bcs-popover>
                <span
                  v-if="row.can_delete"
                  v-authority="{
                    clickable: webAnnotations.perms[row.iam_ns_id]
                      && webAnnotations.perms[row.iam_ns_id].namespace_scoped_delete,
                    actionId: 'namespace_scoped_delete',
                    resourceName: row.namespace,
                    disablePerms: true,
                    permCtx: {
                      project_id: projectId,
                      cluster_id: row.cluster_id,
                      name: row.namespace
                    }
                  }"
                  @click.stop="removeConfigmap(row)"
                  class="biz-operate"
                >{{$t('generic.button.delete')}}</span>
                <bcs-popover :content="row.can_delete_msg || $t('deploy.templateset.cannotDelete')" v-else placement="left">
                  <span class="biz-not-operate">{{$t('generic.button.delete')}}</span>
                </bcs-popover>
              </li>
            </template>
          </bk-table-column>
          <template #empty>
            <BcsEmptyTableStatus :type="searchKeyword ? 'search-empty' : 'empty'" @clear="handleClearSearchData" />
          </template>
        </bk-table>
      </div>

      <bk-sideslider
        v-if="curConfigmap"
        :quick-close="true"
        :is-show.sync="configmapSlider.isShow"
        :title="configmapSlider.title"
        :width="640">
        <div class="p30" slot="content">
          <table class="bk-table biz-data-table has-table-bordered">
            <thead>
              <tr>
                <th style="width: 270px;">{{$t('generic.label.key')}}</th>
                <th>{{$t('generic.label.value')}}<a href="javascript:void(0)" v-if="curConfigmapKeyList.length" class="bk-text-button display-text-btn" @click.stop.prevent="showKeyValue">{{isShowKeyValue ? $t('deploy.templateset.hide') : $t('deploy.templateset.plaintextDisplay')}}</a></th>
              </tr>
            </thead>
            <tbody>
              <template v-if="curConfigmapKeyList.length">
                <tr v-for="item in curConfigmapKeyList" :key="item.key">
                  <td>{{item.key}}</td>
                  <td>
                    <div class="key-box-wrapper">
                      <template v-if="isShowKeyValue">
                        <div :class="['key-box', { 'expanded': item.isExpanded }]">{{item.value}}</div>
                        <a href="javascript: void(0);" class="expand-btn" v-if="item.value.length > 40" @click="item.isExpanded = !item.isExpanded">{{item.isExpanded ? $t('deploy.templateset.collapse') : $t('deploy.templateset.expand')}}</a>
                      </template>
                      <template v-else>
                        ******
                      </template>
                    </div>
                  </td>
                </tr>
              </template>
              <template v-else>
                <tr>
                  <td colspan="2"><bcs-exception type="empty" scene="part"></bcs-exception></td>
                </tr>
              </template>
            </tbody>
          </table>

          <div class="actions">
            <bk-button class="show-labels-btn bk-button bk-button-small bk-primary">{{$t('deploy.templateset.displayLabel')}}</bk-button>
          </div>

          <div class="point-box">
            <template v-if="curConfigmap.labels.length">
              <ul class="key-list">
                <li v-for="(label, index) in curConfigmap.labels" :key="index">
                  <span class="key">{{label[0]}}</span>
                  <span class="value">{{label[1]}}</span>
                </li>
              </ul>
            </template>
            <template v-else>
              <bcs-exception type="empty" scene="part"></bcs-exception>
            </template>
          </div>
        </div>
      </bk-sideslider>

      <bk-sideslider
        :is-show.sync="addSlider.isShow"
        :title="addSlider.title"
        :width="800"
        :quick-close="false">
        <div class="p30 bk-resource-configmap" slot="content">
          <div v-bkloading="{ isLoading: isUpdateLoading }">
            <div class="bk-form-item">
              <div class="bk-form-item" style="margin-bottom: 20px;">
                <label class="bk-label">{{$t('generic.label.name')}}：</label>
                <div class="bk-form-content" style="margin-left: 105px;">
                  <bk-input
                    name="configmapName"
                    disabled
                    style="min-width: 310px;"
                    v-model="curConfigmapName" />
                </div>
              </div>
              <label class="bk-label">{{$t('generic.label.key')}}：</label>
              <div class="bk-form-content" style="margin-left: 105px;">
                <div class="biz-list-operation">
                  <div class="item" v-for="(data, index) in configmapKeyList" :key="index">
                    <bk-button
                      v-show="!data.isEdit"
                      style="width: 120px;"
                      :class="['bk-button bcs-button-ellipsis', { 'bk-primary': curKeyIndex === index }]"
                      @click.stop.prevent="setCurKey(data, index)">
                      {{data.key || $t('deploy.templateset.unnamed')}}
                    </bk-button>
                    <bkbcs-input
                      v-show="data.isEdit"
                      :ref="`bcs-input-${index}`"
                      :key="index"
                      type="text"
                      placeholder=""
                      style="width: 120px;"
                      :value.sync="data.key"
                      :list="varList"
                      @blur="setKey(data, index)">
                    </bkbcs-input>
                    <span class="bcs-icon bcs-icon-edit" v-show="!data.isEdit" @click.stop.prevent="editKey(data, index)"></span>
                    <span class="bcs-icon bcs-icon-close" v-show="!data.isEdit" @click.stop.prevent="removeKey(data, index)"></span>
                  </div>
                  <bcs-popover ref="keyTooltip" :content="$t('deploy.templateset.addKey')" placement="top">
                    <bk-button class="bk-button bk-default is-outline is-icon" @click.stop.prevent="addKey">
                      <i class="bcs-icon bcs-icon-plus"></i>
                    </bk-button>
                  </bcs-popover>
                </div>
              </div>
            </div>
            <template v-if="curKeyParams">
              <div class="bk-form-item" style="margin-top: 13px;">
                <label class="bk-label">{{$t('generic.label.value')}}：</label>
                <div class="bk-form-content" style="margin-left: 105px;" v-if="curProject.kind === PROJECT_K8S">
                  <div v-full-screen="{ css: 'color: #333' }" class="h-[200px]">
                    <textarea
                      class="bk-form-textarea h-full"
                      v-model="curKeyParams.content"
                      :placeholder="$t('deploy.templateset.enterKeyName') + curKeyParams.key + $t('deploy.templateset.keyContent')">
                  </textarea>
                  </div>
                </div>
                <div class="bk-form-content" style="margin-left: 105px;" v-else>
                  <div v-full-screen="{ css: 'color: #333' }" class="h-[200px]">
                    <textarea
                      class="bk-form-textarea h-full"
                      v-model="curKeyParams.content"
                      :placeholder="$t('deploy.templateset.enterKeyName') + curKeyParams.key + $t('deploy.templateset.keyContent')"
                      v-if="curKeyParams.type === 'file'">
                  </textarea>
                    <textarea
                      class="bk-form-textarea h-full"
                      v-model="curKeyParams.content"
                      :placeholder="$t('deploy.templateset.enterOnlineFileUrl')"
                      v-else>
                  </textarea>
                  </div>
                </div>
              </div>
            </template>
            <div class="action-inner" style="margin-top: 20px; margin-left: 105px;">
              <bk-button type="primary" :loading="isSaveBtnLoading" @click="submitUpdateConfigmap">
                {{$t('generic.button.save')}}
              </bk-button>
              <bk-button :disabled="isSaveBtnLoading" @click="cancleUpdateConfigmap">
                {{$t('generic.button.cancel')}}
              </bk-button>
            </div>
          </div>
        </div>
      </bk-sideslider>

      <bk-dialog
        :is-show="batchDialogConfig.isShow"
        :width="600"
        :has-header="false"
        :quick-close="false"
        :title="$t('generic.title.confirmDelete')"
        @confirm="deleteConfigmaps(batchDialogConfig.data)"
        @cancel="batchDialogConfig.isShow = false">
        <template slot="content">
          <div class="biz-batch-wrapper">
            <p class="batch-title">{{$t('deploy.templateset.confirmDeleteConfigMap')}}</p>
            <ul class="batch-list">
              <li v-for="(item, index) of batchDialogConfig.list" :key="index" :title="item">{{item}}</li>
            </ul>
          </div>
        </template>
      </bk-dialog>
    </div>
  </BcsContent>
</template>

<script>
import { catchErrorHandler, formatDate } from '@/common/util';
import ClusterSelectComb from '@/components/cluster-selector/cluster-select-comb.vue';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import fullScreen from '@/directives/full-screen';

export default {
  directives: {
    'full-screen': fullScreen,
  },
  components: { BcsContent, Row, ClusterSelectComb },
  data() {
    return {
      formatDate,
      isInitLoading: true,
      isPageLoading: false,
      searchKeyword: '',
      searchScope: '',
      curPageData: [],
      curConfigmap: null,
      isShowKeyValue: false,
      isSaveBtnLoading: false,
      pageConf: {
        total: 1,
        totalPage: 1,
        pageSize: 10,
        curPage: 1,
        show: true,
      },
      batchDialogConfig: {
        isShow: false,
        list: [],
        data: [],
      },
      configmapSlider: {
        title: '',
        isShow: false,
      },
      addSlider: {
        title: '',
        isShow: false,
      },
      configmapKeyList: [],
      curKeyIndex: 0,
      curKeyParams: null,
      curConfigmapName: '',
      namespaceId: 0,
      instanceId: 0,
      clusterId: '',
      namespace: '',
      isUpdateLoading: false,
      configmapTimer: null,
      isBatchRemoving: false,
      curSelectedData: [],
      alreadySelectedNums: 0,
      curConfigmapKeyList: [],
      comfigSelectedList: [],
      webAnnotations: { perms: {} },
      PROJECT_K8S: window.PROJECT_K8S,
    };
  },
  computed: {
    isEn() {
      return this.$store.state.isEn;
    },
    projectId() {
      return this.$route.params.projectId;
    },
    configmapList() {
      const list = this.$store.state.resource.configmapList;
      list.forEach((item) => {
        item.isChecked = false;
      });
      return JSON.parse(JSON.stringify(list));
    },
    varList() {
      return this.$store.state.variable.varList;
    },
    searchScopeList() {
      const { clusterList } = this.$store.state.cluster;
      const results = clusterList.map(item => ({
        id: item.cluster_id,
        name: item.name,
      }));

      return results;
    },
    isClusterDataReady() {
      return this.$store.state.cluster.isClusterDataReady;
    },
    curClusterId() {
      return this.$store.getters.curClusterId;
    },
    curSelectedCluster() {
      return this.searchScopeList.find(item => item.id === this.searchScope) || {};
    },
    curProject() {
      return this.$store.state.curProject;
    },
  },
  watch: {
    searchScope() {
      this.getConfigmapList();
    },
  },
  created() {
    this.initPageConf();
    // this.getConfigmapList()
  },
  methods: {
    /**
             * 刷新列表
             */
    refresh() {
      this.pageConf.curPage = 1;
      this.isPageLoading = true;
      this.getConfigmapList();
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
             * 清空选择
             */
    clearSelectConfigmaps() {
      this.curPageData.forEach((item) => {
        item.isChecked = false;
      });
    },

    /**
             * 确认删除configmap
             */
    async removeConfigmaps() {
      const data = [];
      const names = [];
      if (this.comfigSelectedList.length) {
        this.comfigSelectedList.forEach((item) => {
          data.push({
            cluster_id: item.cluster_id,
            namespace: item.namespace,
            name: item.name,
          });
          names.push(`${item.cluster_id} / ${item.namespace} / ${item.resourceName}`);
        });
      }
      if (!data.length) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.selectConfigMapToDelete'),
        });
        return false;
      }

      this.batchDialogConfig.list = names;
      this.batchDialogConfig.data = data;
      this.batchDialogConfig.isShow = true;
    },

    /**
             * 删除configmap
             * @param  {Object} data configmap
             */
    async deleteConfigmaps(data) {
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const me = this;
      const { projectId } = this;

      this.batchDialogConfig.isShow = false;
      this.isPageLoading = true;
      try {
        await this.$store.dispatch('resource/deleteConfigmaps', {
          projectId,
          data,
        });

        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.delete'),
        });
        // 稍晚一点加载数据，接口不一定立即清除
        setTimeout(() => {
          me.getConfigmapList();
        }, 500);
      } catch (e) {
        // 4004，已经被删除过，但接口不能立即清除，再重新拉数据，防止重复删除
        if (e.code === 4004) {
          setTimeout(() => {
            me.getConfigmapList();
          }, 500);
        } else {
          me.isPageLoading = false;
        }
        catchErrorHandler(e, this);
      }
    },

    /**
             * 更新configmap
             * @param  {Object} configmap configmap
             */
    async updateConfigmap(configmap) {
      this.addSlider.isShow = true;
      this.isUpdateLoading = true;
      this.addSlider.title = `${this.$t('generic.button.update')}${configmap.name}`;
      this.curConfigmapName = configmap.name;
      this.namespaceId = configmap.namespace_id;
      this.instanceId = configmap.instance_id;
      this.namespace = configmap.namespace;
      this.clusterId = configmap.cluster_id;

      try {
        const res = await this.$store.dispatch('resource/updateSelectConfigmap', {
          projectId: this.projectId,
          namespace: this.namespace,
          name: this.curConfigmapName,
          clusterId: this.clusterId,
        });
        const configmapObj = res.data[0] || {};
        this.initKeyList(configmapObj);
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        this.isUpdateLoading = false;
      }
    },

    /**
             * 删除configmap前的确认
             * @param  {Object} configmap configmap
             * @return {[type]}
             */
    async removeConfigmap(configmap) {
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const me = this;
      me.$bkInfo({
        title: me.$t('generic.title.confirmDelete'),
        clsName: 'biz-remove-dialog max-size',
        content: me.$createElement('p', {
          class: 'biz-confirm-desc',
        }, `${this.$t('deploy.templateset.confirmRemoveConfigMap')}【${configmap.cluster_id} / ${configmap.namespace} / ${configmap.name}】？`),
        confirmFn() {
          me.deleteConfigmap(configmap);
        },
      });
    },

    /**
             * 删除configmap
             * @param {Object} configmap configmap
             */
    async deleteConfigmap(configmap) {
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const me = this;
      const { projectId } = me;
      const clusterId = configmap.cluster_id;
      const { namespace } = configmap;
      const { name } = configmap;

      this.isPageLoading = true;
      try {
        await this.$store.dispatch('resource/deleteConfigmap', { projectId, clusterId, namespace, name });
        me.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.delete'),
        });
        // 稍晚一点加载数据，接口不一定立即清除
        setTimeout(() => {
          me.getConfigmapList();
        }, 500);
      } catch (e) {
        catchErrorHandler(e, this);
        this.isPageLoading = false;
      }
    },

    /**
             * 向服务器提交configmap更新数据
             */
    async submitUpdateConfigmap() {
      this.isSaveBtnLoading = true;
      const enity = {};
      enity.namespace_id = this.namespaceId;
      enity.instance_id = this.instanceId;
      enity.config = {};
      const oName = {
        name: this.curConfigmapName,
      };
      enity.config.metadata = oName;
      const keyList = [];
      const oKey = {};

      const k8sList = this.configmapKeyList;
      const k8sLength = k8sList.length;
      for (let i = 0; i < k8sLength; i++) {
        const item = k8sList[i];
        keyList.push(item.key);
        oKey[item.key] = item.content;
      }
      const aKey = keyList.sort();
      for (let i = 0; i < aKey.length; i++) {
        if (aKey[i] === aKey[i + 1]) {
          this.bkMessageInstance = this.$bkMessage({
            theme: 'error',
            message: `${this.$t('generic.label.key')}【${aKey[i]}】${this.$t('deploy.templateset.duplicate')}`,
          });
          return;
        }
      }
      enity.config.data = oKey;

      try {
        await this.$store.dispatch('resource/updateSingleConfigmap', {
          projectId: this.projectId,
          clusterId: this.clusterId,
          namespace: this.namespace,
          name: this.curConfigmapName,
          data: enity,
        });
        this.isPageLoading = true;
        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.update'),
        });
        this.getConfigmapList();
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        this.isSaveBtnLoading = false;
        this.cancleUpdateConfigmap();
      }
    },

    /**
             * 取消更新configmap
             */
    cancleUpdateConfigmap() {
      // 数据清空或恢复默认值
      this.addSlider.isShow = false;
      this.isUpdateLoading = false;
      this.configmapKeyList.splice(0, this.configmapKeyList.length, ...[]);
      this.curKeyIndex = 0;
      this.namespaceId = 0;
      this.instanceId = 0;
      this.curKeyParams = null;
      this.curConfigmapName = '';
      this.namespace = '';
      this.clusterId = '';
    },

    /**
             * 添加key
             */
    addKey() {
      const index = this.configmapKeyList.length + 1;
      this.configmapKeyList.push({
        key: `key-${index}`,
        isEdit: false,
        content: '',
      });
      this.curKeyParams = this.configmapKeyList[index - 1];
      this.curKeyIndex = index - 1;
      this.$refs.keyTooltip.visible = false;
    },

    /**
             * 设置当前key
             * @param {Object} data 当前key数据
             * @param {number} index 索引
             */
    setKey(data, index) {
      if (data.key === '') {
        data.key = `key-${index + 1}`;
      } else {
        const nameReg = /^[a-zA-Z]{1}[a-zA-Z0-9-_.]{0,254}$/;
        // eslint-disable-next-line no-useless-escape
        const varReg = /\{\{([^\{\}]+)?\}\}/g;

        if (!nameReg.test(data.key.replace(varReg, 'key'))) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('deploy.templateset.msg.labelKey'),
            delay: 5000,
          });
          return false;
        }

        const keyObj = {};
        for (const item of this.configmapKeyList) {
          if (!keyObj[item.key]) {
            keyObj[item.key] = true;
          } else {
            this.$bkMessage({
              theme: 'error',
              message: this.$t('deploy.templateset.keyNotDuplicate'),
              delay: 5000,
            });
            data.isEdit = false;
            return false;
          }
        }
      }
      this.curKeyParams = this.configmapKeyList[index];
      this.curKeyIndex = index;
      data.isEdit = false;
    },

    /**
             * 删除key
             * @param  {Object} data data
             * @param  {Number} index 索引
             */
    removeKey(data, index) {
      if (this.curKeyIndex > index) {
        this.curKeyIndex = this.curKeyIndex - 1;
      } else if (this.curKeyIndex === index) {
        this.curKeyIndex = 0;
      }
      this.configmapKeyList.splice(index, 1);
      this.curKeyParams = this.configmapKeyList[this.curKeyIndex];
    },

    /**
             * 编辑key
             * @param  {Object} data data
             * @param  {Number} index 索引
             */
    editKey(data, index) {
      setTimeout(() => {
        this.$refs[`bcs-input-${index}`][0].$el.children[0].focus();
        this.setCurKey(data, index);
      }, 200);
      // this.$refs[`bsc-input-${index}`]
      data.isEdit = true;
    },

    /**
             * 选择当前key
             * @param  {Object} data data
             * @param  {Number} index 索引
             */
    setCurKey(data, index) {
      this.curKeyParams = data;
      this.curKeyIndex = index;
    },

    /**
             * 编辑key
             * @param  {Object} data data
             * @param  {Number} index 索引
             */
    initKeyList(configmap) {
      const list = [];
      const k8sConfigmapData = configmap.data.data;
      for (const [key, value] of Object.entries(k8sConfigmapData)) {
        list.push({
          key,
          isEdit: false,
          content: value,
        });
      }
      this.curKeyIndex = 0;
      if (list.length) {
        this.curKeyParams = list[0];
      } else {
        this.curKeyParams = null;
      }
      this.configmapKeyList.splice(0, this.configmapKeyList.length, ...list);
    },

    /**
             * 展示configmap详情
             * @param  {Object} configmap configmap
             * @param  {Number} index 索引
             */
    async showConfigmapDetail(configmap) {
      this.configmapSlider.title = configmap.resourceName;
      this.curConfigmap = configmap;
      this.configmapSlider.isShow = true;
      this.updateConfigmapKeyList();
    },

    updateConfigmapKeyList() {
      if (this.curConfigmap) {
        const results = [];

        const data = this.curConfigmap.data.data || {};

        const keys = Object.keys(data);
        keys.forEach((item) => {
          results.push({
            isExpanded: false,
            key: item,
            value: data[item],
          });
        });

        this.curConfigmapKeyList = results;
      } else {
        this.curConfigmapKeyList = [];
      }
    },

    showKeyValue() {
      this.isShowKeyValue = !this.isShowKeyValue;
    },

    /**
             * 加载configmap列表数据
             */
    async getConfigmapList() {
      const { projectId } = this;
      const params = {
        cluster_id: this.searchScope,
      };
      try {
        this.isPageLoading = true;
        const res = await this.$store.dispatch('resource/getConfigmapList', {
          projectId,
          params,
        });
        this.webAnnotations = res.web_annotations || { perms: {} };
        this.initPageConf();
        this.curPageData = this.getDataByPage(this.pageConf.curPage);
        // 如果有搜索关键字，继续显示过滤后的结果
        if (this.searchKeyword) {
          this.searchConfigmap();
        }
      } catch (e) {
        catchErrorHandler(e, this);
        clearTimeout(this.configmapTimer);
        this.configmapTimer = null;
      } finally {
        // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
        setTimeout(() => {
          this.isPageLoading = false;
          this.isInitLoading = false;
        }, 200);
      }
    },

    /**
             * 清除搜索
             */
    clearSearch() {
      this.searchKeyword = '';
      this.searchConfigmap();
    },

    /**
             * 搜索configmap
             */
    searchConfigmap() {
      const keyword = this.searchKeyword.trim();
      const keyList = ['resourceName', 'namespace', 'cluster_id'];
      let list = JSON.parse(JSON.stringify(this.$store.state.resource.configmapList));
      let results = [];
      this.pageConf.curPage = 1;
      this.isPageLoading = true;

      if (this.searchScope) {
        list = list.filter(item => item.cluster_id === this.searchScope);
      }

      if (keyword) {
        clearTimeout(this.configmapTimer);

        list.forEach((item) => {
          item.isChecked = false;
          for (const key of keyList) {
            if (item[key].indexOf(keyword) > -1) {
              results.push(item);
              return true;
            }
          }
        });
      } else {
        results = list;
      }
      this.configmapList.splice(0, this.configmapList.length, ...results);
      this.initPageConf();
      this.curPageData = this.getDataByPage(this.pageConf.curPage);
    },

    /**
             * 初始化分页配置
             */
    initPageConf() {
      const total = this.configmapList.length;
      this.pageConf.total = total;
      this.pageConf.curPage = 1;
      this.pageConf.totalPage = Math.ceil(total / this.pageConf.pageSize);
    },

    /**
             * 重新加载当面页数据
             * @return {[type]} [description]
             */
    reloadCurPage() {
      this.initPageConf();
      this.curPageData = this.getDataByPage(this.pageConf.curPage);
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
      if (endIndex > this.configmapList.length) {
        endIndex = this.configmapList.length;
      }
      setTimeout(() => {
        this.isPageLoading = false;
      }, 200);
      this.comfigSelectedList = [];
      return this.configmapList.slice(startIndex, endIndex);
    },

    /**
             * 单选
             * @param {array} selection 已经选中的行数
             * @param {object} row 当前选中的行
             */
    handlePageSelect(selection) {
      this.comfigSelectedList = selection;
    },

    /**
             * 全选
             */
    handlePageSelectAll(selection) {
      this.comfigSelectedList = selection;
    },

    /**
             * 页数改变回调
             * @param  {number} page 第几页
             */
    pageChangeHandler(page = 1) {
      this.pageConf.curPage = page;
      if (this.configmapTimer) {
        this.getConfigmapList();
      } else {
        const data = this.getDataByPage(page);
        this.curPageData = data;
      }
    },

    /**
             * 每行的多选框点击事件
             */
    rowClick() {
      this.$nextTick(() => {
        this.alreadySelectedNums = this.configmapList.filter(item => item.isChecked).length;
      });
    },

    rowSelectable(row) {
      return row.can_delete
                    && this.webAnnotations.perms[row.iam_ns_id]?.namespace_scoped_delete;
    },
    handleClearSearchData() {
      this.searchKeyword = '';
      this.searchConfigmap();
    },
  },
};
</script>

<style scoped>
    @import './configmap.css';
    .bk-spin-loading  {
        position: absolute;
        top: 28px;
        left: 48px;
    }
</style>
