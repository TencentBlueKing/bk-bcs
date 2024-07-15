<!-- eslint-disable vue/multi-word-component-names -->
<!-- eslint-disable max-len -->
<template>
  <BcsContent hide-back title="Ingresses" :desc="$t('deploy.templateset.createFromTemplateOrHelmIngress')">
    <div v-bkloading="{ isLoading: isInitLoading }">
      <Row class="mb-[16px]">
        <div class="left" slot="left">
          <bk-button
            class="bk-button bk-default"
            v-if="curPageData.length"
            @click.stop.prevent="removeIngresses">
            <span>{{$t('generic.button.batchDelete')}}</span>
          </bk-button>
        </div>
        <div class="right" slot="right">
          <ClusterSelectComb
            :placeholder="$t('deploy.templateset.searchNameOrNamespaceEnter')"
            :search.sync="searchKeyword"
            :cluster-id.sync="searchScope"
            cluster-type="all"
            @search-change="searchIngress"
            @refresh="refresh" />
        </div>
      </Row>

      <div class="biz-resource">
        <div class="biz-table-wrapper">
          <bk-table
            :size="'medium'"
            :data="curPageData"
            :pagination="pageConf"
            v-bkloading="{ isLoading: isPageLoading && !isInitLoading, opacity: 1 }"
            @page-limit-change="handlePageLimitChange"
            @page-change="handlePageChange"
            @select="handlePageSelect"
            @select-all="handlePageSelectAll">
            <bk-table-column type="selection" width="60" :selectable="rowSelectable"></bk-table-column>
            <bk-table-column :label="$t('generic.label.name')" :show-overflow-tooltip="true" min-width="200">
              <template slot-scope="props">
                <a
                  href="javascript: void(0)"
                  class="bk-text-button biz-resource-title"
                  v-authority="{
                    clickable: webAnnotations.perms[props.row.iam_ns_id]
                      && webAnnotations.perms[props.row.iam_ns_id].namespace_scoped_view,
                    actionId: 'namespace_scoped_view',
                    resourceName: props.row.namespace,
                    disablePerms: true,
                    permCtx: {
                      project_id: projectId,
                      cluster_id: props.row.cluster_id,
                      name: props.row.namespace
                    }
                  }"
                  @click.stop.prevent="showIngressDetail(props.row, index)"
                >{{props.row.resourceName}}</a>
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('generic.label.cluster1')" min-width="150">
              <template slot-scope="props">
                <bcs-popover :content="props.row.cluster_id || '--'" placement="top">
                  <p class="biz-text-wrapper">{{curCluster ? curCluster.clusterName : '--'}}</p>
                </bcs-popover>
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('k8s.namespace')" min-width="130">
              <template slot-scope="props">
                {{props.row.namespace}}
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('deploy.templateset.sources')" min-width="130">
              <template slot-scope="props">
                {{props.row.source_type ? props.row.source_type : '--'}}
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('cluster.labels.createdAt')" min-width="160">
              <template slot-scope="props">
                {{props.row.createTime ? formatDate(props.row.createTime) : '--'}}
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('cluster.labels.updatedAt')" min-width="160">
              <template slot-scope="props">
                {{props.row.updateTime ? formatDate(props.row.updateTime) : '--'}}
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('generic.label.updator')" min-width="100">
              <template slot-scope="props">
                {{props.row.updator || '--'}}
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('generic.label.action')" width="150">
              <template slot-scope="props">
                <a
                  v-if="props.row.can_update"
                  href="javascript:void(0);"
                  class="bk-text-button"
                  v-authority="{
                    clickable: webAnnotations.perms[props.row.iam_ns_id]
                      && webAnnotations.perms[props.row.iam_ns_id].namespace_scoped_update,
                    actionId: 'namespace_scoped_update',
                    resourceName: props.row.namespace,
                    disablePerms: true,
                    permCtx: {
                      project_id: projectId,
                      cluster_id: props.row.cluster_id,
                      name: props.row.namespace
                    }
                  }"
                  @click="showIngressEditDialog(props.row)"
                >{{$t('generic.button.update')}}</a>
                <bcs-popover :content="props.row.can_update_msg" v-else placement="left">
                  <a href="javascript:void(0);" class="bk-text-button is-disabled">{{$t('generic.button.update')}}</a>
                </bcs-popover>
                <a
                  v-if="props.row.can_delete"
                  v-authority="{
                    clickable: webAnnotations.perms[props.row.iam_ns_id]
                      && webAnnotations.perms[props.row.iam_ns_id].namespace_scoped_delete,
                    actionId: 'namespace_scoped_delete',
                    resourceName: props.row.namespace,
                    disablePerms: true,
                    permCtx: {
                      project_id: projectId,
                      cluster_id: props.row.cluster_id,
                      name: props.row.namespace
                    }
                  }"
                  @click.stop="removeIngress(props.row)"
                  class="bk-text-button ml10"
                >{{$t('generic.button.delete')}}</a>
                <bcs-popover :content="props.row.can_delete_msg || $t('deploy.templateset.cannotDelete')" v-else placement="left">
                  <span class="bk-text-button is-disabled ml10">{{$t('generic.button.delete')}}</span>
                </bcs-popover>
              </template>
            </bk-table-column>
            <template #empty>
              <BcsEmptyTableStatus :type="searchKeyword ? 'search-empty' : 'empty'" @clear="handleClearSearchData" />
            </template>
          </bk-table>
        </div>
      </div>

      <bk-sideslider
        v-if="curIngress"
        :quick-close="true"
        :is-show.sync="ingressSlider.isShow"
        :title="ingressSlider.title"
        :width="800">
        <div class="pt20 pr30 pb20 pl30" slot="content">
          <label class="biz-title">{{$t('deploy.templateset.hostList')}}（spec.tls）</label>
          <table class="bk-table biz-data-table has-table-bordered biz-special-bk-table">
            <thead>
              <tr>
                <th style="width: 270px;">{{$t('deploy.templateset.hostName')}}</th>
                <th>SecretName</th>
              </tr>
            </thead>
            <tbody>
              <template v-if="curIngress.tls.length">
                <tr v-for="(rule, index) in curIngress.tls" :key="index">
                  <td>{{rule.host || '--'}}</td>
                  <td>{{rule.secretName || '--'}}</td>
                </tr>
              </template>
              <template v-else>
                <tr>
                  <td colspan="2"><bcs-exception type="empty" scene="part"></bcs-exception></td>
                </tr>
              </template>
            </tbody>
          </table>

          <label class="biz-title">{{$t('generic.label.rule')}}（spec.rules）</label>
          <table class="bk-table biz-data-table has-table-bordered biz-special-bk-table">
            <thead>
              <tr>
                <th style="width: 200px;">{{$t('deploy.templateset.hostName')}}</th>
                <th style="width: 150px;">{{$t('deploy.templateset.path')}}</th>
                <th>{{$t('deploy.templateset.serviceName')}}</th>
                <th style="width: 100px;">{{$t('deploy.templateset.servicePort')}}</th>
              </tr>
            </thead>
            <tbody>
              <template v-if="curIngress.rules.length">
                <tr v-for="(rule, index) in curIngress.rules" :key="index">
                  <td>{{rule.host || '--'}}</td>
                  <td>{{rule.path || '--'}}</td>
                  <td>{{rule.serviceName || '--'}}</td>
                  <td>{{rule.servicePort || '--'}}</td>
                </tr>
              </template>
              <template v-else>
                <tr>
                  <td colspan="4"><bcs-exception type="empty" scene="part"></bcs-exception></td>
                </tr>
              </template>
            </tbody>
          </table>

          <div class="actions">
            <bk-button class="show-labels-btn bk-button bk-button-small bk-primary">{{$t('deploy.templateset.displayLabel')}}</bk-button>
          </div>

          <div class="point-box">
            <template v-if="curIngress.labels.length">
              <ul class="key-list">
                <li v-for="(label, index) in curIngress.labels" :key="index">
                  <span class="key">{{label[0]}}</span>
                  <span class="value">{{label[1] || '--'}}</span>
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
        :is-show.sync="ingressEditSlider.isShow"
        :title="ingressEditSlider.title"
        :width="1020"
        @hidden="handleCancelUpdate">
        <div slot="content">
          <div class="bk-form biz-configuration-form pt20 pb20 pl10 pr20">
            <div class="bk-form-item">
              <div class="bk-form-item">
                <div class="bk-form-content" style="margin-left: 0;">
                  <div class="bk-form-item is-required">
                    <label class="bk-label" style="width: 130px;">{{$t('generic.label.name')}}：</label>
                    <div class="bk-form-content" style="margin-left: 130px;">
                      <bk-input
                        :disabled="true"
                        style="width: 310px;"
                        v-model="curEditedIngress.config.metadata.name"
                        maxlength="64"
                        name="applicationName" />
                    </div>
                  </div>
                </div>
              </div>

              <div class="bk-form-item">
                <div class="bk-form-content" style="margin-left: 130px;">
                  <button :class="['bk-text-button f12 mb10 pl0', { 'rotate': isTlsPanelShow }]" @click.stop.prevent="toggleTlsPanel">
                    {{$t('deploy.templateset.TLSsettings')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                  </button>
                  <button :class="['bk-text-button f12 mb10 pl0', { 'rotate': isPanelShow }]" @click.stop.prevent="togglePanel">
                    {{$t('deploy.templateset.moreSettings')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                  </button>
                </div>
              </div>

              <div class="bk-form-item mt0" v-show="isTlsPanelShow">
                <div class="bk-form-content" style="margin-left: 130px;">
                  <bk-tab :type="'fill'" :active-name="'tls'" :size="'small'">
                    <bk-tab-panel name="tls" title="TLS">
                      <div class="p20">
                        <table class="biz-simple-table">
                          <tbody>
                            <tr v-for="(computer, index) in curEditedIngress.config.spec.tls" :key="index">
                              <td>
                                <bkbcs-input
                                  type="text"
                                  :placeholder="$t('deploy.templateset.hostnamesCommaSeparated')"
                                  style="width: 310px;"
                                  :value.sync="computer.hosts"
                                  :list="varList"
                                >
                                </bkbcs-input>
                              </td>
                              <td>
                                <bkbcs-input
                                  type="text"
                                  :placeholder="$t('deploy.templateset.enterCertificate')"
                                  style="width: 350px;"
                                  :value.sync="computer.secretName"
                                >
                                </bkbcs-input>
                              </td>
                              <td>
                                <bk-button class="action-btn ml5" @click.stop.prevent="addTls">
                                  <i class="bcs-icon bcs-icon-plus"></i>
                                </bk-button>
                                <bk-button class="action-btn" v-if="curEditedIngress.config.spec.tls.length > 1" @click.stop.prevent="removeTls(index, computer)">
                                  <i class="bcs-icon bcs-icon-minus"></i>
                                </bk-button>
                              </td>
                            </tr>
                          </tbody>
                        </table>
                      </div>
                    </bk-tab-panel>
                  </bk-tab>
                </div>
              </div>

              <div class="bk-form-item mt0" v-show="isPanelShow">
                <div class="bk-form-content" style="margin-left: 130px;">
                  <bk-tab :type="'fill'" :active-name="'remark'" :size="'small'">
                    <bk-tab-panel name="remark" :title="$t('k8s.annotation')">
                      <div class="biz-tab-wrapper m20">
                        <bk-keyer :key-list.sync="curRemarkList" :var-list="varList" ref="remarkKeyer"></bk-keyer>
                      </div>
                    </bk-tab-panel>
                    <bk-tab-panel name="label" :title="$t('k8s.label')">
                      <div class="biz-tab-wrapper m20">
                        <bk-keyer :key-list.sync="curLabelList" :var-list="varList" ref="labelKeyer"></bk-keyer>
                      </div>
                    </bk-tab-panel>
                  </bk-tab>
                </div>
              </div>

              <!-- part2 start -->
              <div class="biz-part-header">
                <div class="bk-button-group">
                  <div class="item" v-for="(rule, index) in curEditedIngress.config.spec.rules" :key="index">
                    <bk-button :class="['bk-button bk-default is-outline', { 'is-selected': curRuleIndex === index }]" @click.stop="setCurRule(rule, index)">
                      {{rule.host || $t('deploy.templateset.unnamed')}}
                    </bk-button>
                    <span class="bcs-icon bcs-icon-close-circle" @click.stop="removeRule(index)" v-if="curEditedIngress.config.spec.rules.length > 1"></span>
                  </div>
                  <bcs-popover ref="containerTooltip" :content="$t('deploy.templateset.addRule')" placement="top">
                    <bk-button type="button" class="bk-button bk-default is-outline is-icon" @click.stop.prevent="addLocalRule">
                      <i class="bcs-icon bcs-icon-plus"></i>
                    </bk-button>
                  </bcs-popover>
                </div>
              </div>

              <div class="bk-form biz-configuration-form pb15">
                <div class="biz-span">
                  <span class="title">{{$t('generic.title.basicInfo')}}</span>
                </div>
                <div class="bk-form-item is-required">
                  <label class="bk-label" style="width: 130px;">{{$t('deploy.templateset.virtualHostName')}}：</label>
                  <div class="bk-form-content" style="margin-left: 130px;">
                    <bk-input :placeholder="$t('generic.placeholder.input')" style="width: 310px;" v-model="curRule.host" name="ruleName" />
                  </div>
                </div>
                <div class="bk-form-item">
                  <label class="bk-label" style="width: 130px;">{{$t('deploy.templateset.pathGroup')}}：</label>
                  <div class="bk-form-content" style="margin-left: 130px;">
                    <table class="biz-simple-table">
                      <tbody>
                        <tr v-for="(pathRule, index) of curRule.http.paths" :key="index">
                          <td>
                            <bkbcs-input
                              type="text"
                              :placeholder="$t('deploy.templateset.path')"
                              style="width: 310px;"
                              :value.sync="pathRule.path"
                              :list="varList"
                            >
                            </bkbcs-input>
                          </td>
                          <td style="text-align: center;">
                            <i class="bcs-icon bcs-icon-arrows-right"></i>
                          </td>
                          <td>
                            <bk-selector
                              style="width: 180px;"
                              :placeholder="$t('deploy.templateset._serviceName')"
                              :setting-key="'_name'"
                              :display-key="'_name'"
                              :selected.sync="pathRule.backend.serviceName"
                              :list="linkServices || []"
                              @item-selected="handlerSelectService(pathRule)">
                            </bk-selector>
                          </td>
                          <td>
                            <bk-selector
                              style="width: 180px;"
                              :placeholder="$t('deploy.helm.port')"
                              :setting-key="'_id'"
                              :display-key="'_name'"
                              :selected.sync="pathRule.backend.servicePort"
                              :list="linkServices[pathRule.backend.serviceName] || []">
                            </bk-selector>
                          </td>
                          <td>
                            <bk-button class="action-btn ml5" @click.stop.prevent="addRulePath">
                              <i class="bcs-icon bcs-icon-plus"></i>
                            </bk-button>
                            <bk-button class="action-btn" v-if="curRule.http.paths.length > 1" @click.stop.prevent="removeRulePath(pathRule, index)">
                              <i class="bcs-icon bcs-icon-minus"></i>
                            </bk-button>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                    <p class="biz-tip">{{$t('deploy.templateset.sameHostnameMultiplePathsHint')}}</p>
                  </div>
                </div>
              </div>

              <div class="bk-form-item mt25" style="margin-left: 130px;">
                <bk-button type="primary" :loading="isDetailSaving" @click.stop.prevent="saveIngressDetail">{{$t('deploy.templateset.saveAndUpdate')}}</bk-button>
                <bk-button :loading="isDetailSaving" @click.stop.prevent="handleCancelUpdate">{{$t('generic.button.cancel')}}</bk-button>
              </div>

            </div>
          </div>
        </div>
      </bk-sideslider>

      <bk-dialog
        :title="$t('generic.title.confirmDelete')"
        :header-position="'left'"
        :is-show="batchDialogConfig.isShow"
        :width="600"
        :has-header="false"
        :quick-close="false"
        @confirm="deleteIngresses(batchDialogConfig.data)"
        @cancel="batchDialogConfig.isShow = false">
        <template slot="content">
          <div class="biz-batch-wrapper">
            <p class="batch-title mt10 f14">{{$t('deploy.templateset.confirmDeleteIngress')}}</p>
            <ul class="batch-list">
              <li v-for="(item, index) of batchDialogConfig.list" :key="index">{{item}}</li>
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
import bkKeyer from '@/components/keyer';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import ingressParams from '@/json/k8s-ingress.json';
import ruleParams from '@/json/k8s-ingress-rule.json';

export default {
  components: {
    bkKeyer,
    BcsContent,
    Row,
    ClusterSelectComb,
  },
  data() {
    return {
      formatDate,
      isInitLoading: true,
      isPageLoading: false,
      searchKeyword: '',
      searchScope: '',
      curPageData: [],
      curIngress: null,
      curEditedIngress: ingressParams,
      isPanelShow: false,
      isTlsPanelShow: false,
      isDetailSaving: false,
      pageConf: {
        count: 1,
        totalPage: 1,
        limit: 10,
        current: 1,
        show: true,
      },
      ingressSlider: {
        title: '',
        isShow: false,
      },
      batchDialogConfig: {
        isShow: false,
        list: [],
        data: [],
      },
      curRuleIndex: 0,
      curRule: ingressParams.config.spec.rules[0],
      curIngressName: '',
      alreadySelectedNums: 0,
      ingressEditSlider: {
        title: '',
        isShow: false,
      },
      linkServices: [],
      ingressSelectedList: [],
      webAnnotations: { perms: {} },
    };
  },
  computed: {
    isEn() {
      return this.$store.state.isEn;
    },
    curProject() {
      return this.$store.state.curProject;
    },
    searchScopeList() {
      const { clusterList } = this.$store.state.cluster;
      const results = clusterList.map(item => ({
        id: item.cluster_id,
        name: item.name,
      }));

      return results;
    },
    isCheckCurPageAll() {
      if (this.curPageData.length) {
        const list = this.curPageData;
        const selectList = list.filter(item => item.isChecked === true);
        const canSelectList = list.filter(item => item.can_delete);
        if (selectList.length && (selectList.length === canSelectList.length)) {
          return true;
        }
        return false;
      }
      return false;
    },
    projectId() {
      return this.$route.params.projectId;
    },
    ingressList() {
      const list = this.$store.state.resource.ingressList;
      list.forEach((item) => {
        item.isChecked = false;
      });
      return JSON.parse(JSON.stringify(list));
    },
    isClusterDataReady() {
      return this.$store.state.cluster.isClusterDataReady;
    },
    varList() {
      const list = this.$store.state.variable.varList.map((item) => {
        item._id = item.key;
        item._name = item.key;
        return item;
      });
      return list;
    },
    curLabelList() {
      const list = [];
      const { labels } = this.curEditedIngress.config.metadata;
      for (const [key, value] of Object.entries(labels)) {
        list.push({
          key,
          value,
        });
      }
      if (!list.length) {
        list.push({
          key: '',
          value: '',
        });
      }
      return list;
    },
    curRemarkList() {
      const list = [];
      const { annotations } = this.curEditedIngress.config.metadata;
      for (const [key, value] of Object.entries(annotations)) {
        list.push({
          key,
          value,
        });
      }
      if (!list.length) {
        list.push({
          key: '',
          value: '',
        });
      }
      return list;
    },
    curClusterId() {
      return this.$store.getters.curClusterId;
    },
    curCluster() {
      const list = this.$store.state.cluster.clusterList || [];
      return list.find(item => item.clusterID === this.searchScope);
    },
  },
  watch: {
    searchScope() {
      this.getIngressList();
    },
  },
  created() {
    this.initPageConf();
  },
  methods: {
    /**
             * 刷新列表
             */
    refresh() {
      this.pageConf.current = 1;
      this.isPageLoading = true;
      this.getIngressList();
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
      this.handlePageChange();
    },

    /**
             * 确认批量删除
             */
    async removeIngresses() {
      const data = [];
      const names = [];

      this.ingressSelectedList.forEach((item) => {
        data.push({
          cluster_id: item.cluster_id,
          namespace: item.namespace,
          name: item.name,
        });
        names.push(`${item.cluster_id} / ${item.namespace} / ${item.resourceName}`);
      });

      if (!data.length) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.selectIngressToDelete'),
        });
        return false;
      }

      this.batchDialogConfig.list = names;
      this.batchDialogConfig.data = data;
      this.batchDialogConfig.isShow = true;
    },

    /**
             * 批量删除
             * @param  {object} data ingresses
             */
    async deleteIngresses(data) {
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const me = this;
      const { projectId } = this;

      this.batchDialogConfig.isShow = false;
      this.isPageLoading = true;
      try {
        await this.$store.dispatch('resource/deleteIngresses', { projectId, data });
        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.delete'),
        });
        // 稍晚一点加载数据，接口不一定立即清除
        setTimeout(() => {
          me.getIngressList();
        }, 500);
      } catch (e) {
        // 4004，已经被删除过，但接口不能立即清除，再重新拉数据，防止重复删除
        if (e.code === 4004) {
          me.isPageLoading = true;
          setTimeout(() => {
            me.getIngressList();
          }, 500);
        } else {
          this.isPageLoading = false;
        }
        catchErrorHandler(e, this);
      }
    },

    /**
             * 确认删除ingress
             * @param  {object} ingress ingress
             */
    async removeIngress(ingress) {
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const me = this;
      me.$bkInfo({
        title: me.$t('generic.title.confirmDelete'),
        clsName: 'biz-remove-dialog max-size',
        content: me.$createElement('p', {
          class: 'biz-confirm-desc',
        }, `${this.$t('deploy.templateset.confirmRemoveIngress')}【${ingress.cluster_id} / ${ingress.namespace} / ${ingress.name}】？`),
        confirmFn() {
          me.deleteIngress(ingress);
        },
      });
    },

    /**
             * 删除ingress
             * @param  {object} ingress ingress
             */
    async deleteIngress(ingress) {
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const me = this;
      const { projectId } = me;
      const clusterId = ingress.cluster_id;
      const { namespace } = ingress;
      const { name } = ingress;

      this.isPageLoading = true;
      try {
        await this.$store.dispatch('resource/deleteIngress', {
          projectId,
          clusterId,
          namespace,
          name,
        });
        me.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.delete'),
        });

        // 稍晚一点加载数据，接口不一定立即清除
        setTimeout(() => {
          me.getIngressList();
        }, 500);
      } catch (e) {
        this.isPageLoading = false;
        catchErrorHandler(e, this);
      }
    },

    /**
             * 显示ingress详情
             * @param  {object} ingress object
             * @param  {number} index 索引
             */
    showIngressDetail(ingress) {
      this.ingressSlider.title = ingress.resourceName;
      this.curIngress = ingress;
      this.ingressSlider.isShow = true;
    },

    /**
             * 清除选择，在分页改变时触发
             */
    clearSelectIngress() {
      this.curPageData.forEach((item) => {
        item.isChecked = false;
      });
    },

    /**
             * 获取Ingresslist
             */
    async getIngressList() {
      const { projectId } = this;
      const params = {
        cluster_id: this.searchScope,
      };
      try {
        this.isPageLoading = true;
        const res = await this.$store.dispatch('resource/getIngressList', {
          projectId,
          params,
        });
        this.webAnnotations = res.web_annotations || { perms: {} };

        this.initPageConf();
        this.curPageData = this.getDataByPage(this.pageConf.current);

        // 如果有搜索关键字，继续显示过滤后的结果
        if (this.searchKeyword) {
          this.searchIngress();
        }
      } catch (e) {
        catchErrorHandler(e, this);
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
      this.searchIngress();
    },

    /**
             * 搜索Ingress
             */
    searchIngress() {
      const keyword = this.searchKeyword.trim();
      const keyList = ['resourceName', 'namespace', 'cluster_name'];
      let list = JSON.parse(JSON.stringify(this.$store.state.resource.ingressList));
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

      this.ingressList.splice(0, this.ingressList.length, ...results);
      this.pageConf.current = 1;
      this.initPageConf();
      this.curPageData = this.getDataByPage(this.pageConf.current);
    },

    /**
             * 初始化分页配置
             */
    initPageConf() {
      const total = this.ingressList.length;
      this.pageConf.count = total;
      this.pageConf.current = 1;
      this.pageConf.totalPage = Math.ceil(total / this.pageConf.limit);
    },

    /**
             * 重新加载当面页数据
             * @return {[type]} [description]
             */
    reloadCurPage() {
      this.initPageConf();
      this.curPageData = this.getDataByPage(this.pageConf.current);
    },

    /**
             * 获取分页数据
             * @param  {number} page 第几页
             * @return {object} data 数据
             */
    getDataByPage(page) {
      if (page < 1) {
        // eslint-disable-next-line no-multi-assign
        this.pageConf.current = page = 1;
      }
      let startIndex = (page - 1) * this.pageConf.limit;
      let endIndex = page * this.pageConf.limit;
      this.isPageLoading = true;
      if (startIndex < 0) {
        startIndex = 0;
      }
      if (endIndex > this.ingressList.length) {
        endIndex = this.ingressList.length;
      }
      setTimeout(() => {
        this.isPageLoading = false;
      }, 200);
      this.ingressSelectedList = [];
      return this.ingressList.slice(startIndex, endIndex);
    },

    /**
             * 页数改变回调
             * @param  {number} page 第几页
             */
    handlePageChange(page = 1) {
      this.pageConf.current = page;

      const data = this.getDataByPage(page);
      this.curPageData = data;
    },

    /**
             * 每行的多选框点击事件
             */
    rowClick() {
      this.$nextTick(() => {
        this.alreadySelectedNums = this.ingressList.filter(item => item.isChecked).length;
      });
    },

    async showIngressEditDialog(ingress) {
      // eslint-disable-next-line no-prototype-builtins
      if (!ingress.data.spec.hasOwnProperty('tls')) {
        ingress.data.spec.tls = [
          {
            hosts: '',
            secretName: '',
          },
        ];
      } else if (JSON.stringify(ingress.data.spec.tls) === '[{}]') {
        ingress.data.spec.tls = [
          {
            hosts: '',
            secretName: '',
          },
        ];
      }
      const ingressClone = JSON.parse(JSON.stringify(ingress));
      ingressClone.data.spec.tls.forEach((item) => {
        if (item.hosts?.join) {
          item.hosts = item.hosts.join(',');
        }
      });
      this.curEditedIngress = ingressClone;
      this.curEditedIngress.config = ingressClone.data;
      this.ingressEditSlider.title = ingress.name;
      delete this.curEditedIngress.data;

      if (this.curEditedIngress.config.spec.rules.length) {
        // 初始化数据放在后面使用报错
        const rule = Object.assign({
          http: {
            paths: [
              {
                backend: {
                  serviceName: '',
                  servicePort: '',
                },
                path: '',
              },
            ],
          },
        }, this.curEditedIngress.config.spec.rules[0]);
        this.setCurRule(rule, 0);
      } else {
        this.addLocalRule();
      }
      this.getServiceList(ingress.cluster_id, ingress.namespace_id);
      this.ingressEditSlider.isShow = true;
    },

    togglePanel() {
      this.isTlsPanelShow = false;
      this.isPanelShow = !this.isPanelShow;
    },
    toggleTlsPanel() {
      this.isPanelShow = false;
      this.isTlsPanelShow = !this.isTlsPanelShow;
    },
    goCertList() {
      if (this.certListUrl) {
        window.open(this.certListUrl);
      }
    },
    addTls() {
      this.curEditedIngress.config.spec.tls.push({
        hosts: '',
        secretName: '',
      });
    },
    removeTls(index) {
      this.curEditedIngress.config.spec.tls.splice(index, 1);
    },
    setCurRule(rule, index) {
      this.curRule = rule;
      this.curRuleIndex = index;
    },
    removeRule(index) {
      const { rules } = this.curEditedIngress.config.spec;
      rules.splice(index, 1);
      if (this.curRuleIndex === index) {
        this.curRuleIndex = 0;
      } else if (this.curRuleIndex !== 0) {
        this.curRuleIndex = this.curRuleIndex - 1;
      }

      this.curRule = rules[this.curRuleIndex];
    },
    addLocalRule() {
      const rule = JSON.parse(JSON.stringify(ruleParams));
      const { rules } = this.curEditedIngress.config.spec;
      const index = rules.length;
      rule.host = `rule-${index + 1}`;
      rules.push(rule);
      this.setCurRule(rule, index);
    },
    addRulePath() {
      const params = {
        backend: {
          serviceName: '',
          servicePort: '',
        },
        path: '',
      };

      this.curRule.http.paths.push(params);
    },
    removeRulePath(pathRule, index) {
      this.curRule.http.paths.splice(index, 1);
    },
    handlerSelectService(pathRule) {
      pathRule.backend.servicePort = '';
    },
    async initServices(version) {
      const { projectId } = this;
      await this.$store.dispatch('k8sTemplate/getServicesByVersion', { projectId, version });
    },
    /**
             * 获取service列表
             */
    async getServiceList(clusterId, namespaceId) {
      const { projectId } = this;
      const params = {
        cluster_id: clusterId,
      };

      try {
        const res = await this.$store.dispatch('network/getServiceList', {
          projectId,
          params,
        });

        const serviceList = res.data.filter(service => service.namespace_id === namespaceId).map((service) => {
          const ports = service.data.spec.ports || [];
          return {
            _name: service.resourceName,
            service_name: service.resourceName,
            service_ports: ports,
          };
        });
        serviceList.forEach((service) => {
          serviceList[service.service_name] = [];
          service.service_ports.forEach((item) => {
            serviceList[service.service_name].push({
              _id: item.port,
              _name: item.port,
            });
          });
        });
        this.linkServices = serviceList;
      } catch (e) {
        catchErrorHandler(e, this);
      }
    },

    checkData() {
      const ingress = this.curEditedIngress;
      const nameReg = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;
      const pathReg = /\/((?!\.)[\w\d\-./~]+)*/;
      let megPrefix = '';

      for (const rule of ingress.config.spec.rules) {
        // 检查rule
        if (!rule.host) {
          megPrefix += this.$t('deploy.templateset.ruleLabel');
          this.$bkMessage({
            theme: 'error',
            message: megPrefix + this.$t('deploy.templateset.hostnameNotEmpty'),
          });
          return false;
        }

        if (!nameReg.test(rule.host)) {
          megPrefix += this.$t('deploy.templateset.ruleHostname');
          this.$bkMessage({
            theme: 'error',
            message: megPrefix + this.$t('deploy.templateset.nameErrorAlphanumericDash'),
            delay: 8000,
          });
          return false;
        }

        const paths = rule.http?.paths || [];

        for (const path of paths) {
          if (!path.path) {
            megPrefix += this.$t('deploy.templateset.pathGroupInHost', { host: rule.host });
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset.enterPath'),
              delay: 8000,
            });
            return false;
          }

          if (path.path && !pathReg.test(path.path)) {
            megPrefix += this.$t('deploy.templateset.pathGroupInHost', { host: rule.host });
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset.pathIncorrect'),
              delay: 8000,
            });
            return false;
          }

          if (!path.backend.serviceName) {
            megPrefix += this.$t('deploy.templateset.pathGroupInHost', { host: rule.host });
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset.associateService'),
              delay: 8000,
            });
            return false;
          }

          if (!path.backend.servicePort) {
            megPrefix += this.$t('deploy.templateset.pathGroupInHost', { host: rule.host });
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset._associateServicePort'),
              delay: 8000,
            });
            return false;
          }

          // eslint-disable-next-line no-prototype-builtins
          if (path.backend.serviceName && !this.linkServices.hasOwnProperty(path.backend.serviceName)) {
            megPrefix += this.$t('deploy.templateset.pathGroupInHost', { host: rule.host });
            this.$bkMessage({
              theme: 'error',
              message: megPrefix + this.$t('deploy.templateset.associatedServiceNotExist', { serviceName: path.backend.serviceName }),
              delay: 8000,
            });
            return false;
          }
        }
      }
      return true;
    },

    formatData() {
      const params = JSON.parse(JSON.stringify(this.curEditedIngress));
      delete params.config.metadata.resourceVersion;
      delete params.config.metadata.selfLink;
      delete params.config.metadata.uid;

      params.config.metadata.annotations = this.$refs.remarkKeyer.getKeyObject();
      params.config.metadata.labels = this.$refs.labelKeyer.getKeyObject();

      // 如果不是变量，转为数组形式
      // eslint-disable-next-line no-useless-escape
      const varReg = /\{\{([^\{\}]+)?\}\}/g;
      params.config.spec.tls.forEach((item) => {
        if (!varReg.test(item.hosts)) {
          item.hosts = item.hosts.split(',');
        }
      });
      // 设置当前rules
      params.config.spec.rules = this.curEditedIngress.config.spec.rules.map(item => JSON.parse(JSON.stringify(item)));
      return params;
    },

    /**
             * 保存service
             */
    async saveIngressDetail() {
      if (this.checkData()) {
        const data = this.formatData();
        const { projectId } = this;
        const clusterId = this.curEditedIngress.cluster_id;
        const { namespace } = this.curEditedIngress;
        const ingressId = this.curEditedIngress.config.metadata.name;

        if (this.isDetailSaving) {
          return false;
        }

        this.isDetailSaving = true;

        try {
          await this.$store.dispatch('resource/saveIngressDetail', {
            projectId,
            clusterId,
            namespace,
            ingressId,
            data,
          });

          this.$bkMessage({
            theme: 'success',
            message: this.$t('generic.msg.success.save'),
            hasCloseIcon: true,
            delay: 3000,
          });
          this.getIngressList();
          this.handleCancelUpdate();
        } catch (e) {
          catchErrorHandler(e, this);
        } finally {
          this.isDetailSaving = false;
        }
      }
    },

    /**
             * 单选
             * @param {array} selection 已经选中的行数
             * @param {object} row 当前选中的行
             */
    handlePageSelect(selection) {
      this.ingressSelectedList = selection;
    },

    /**
             * 全选
             */
    handlePageSelectAll(selection) {
      this.ingressSelectedList = selection;
    },

    handleCancelUpdate() {
      this.ingressEditSlider.isShow = false;
    },

    handlerSelectCert(computer, index, data) {
      computer.certType = data.certType;
    },
    rowSelectable(row) {
      return row.can_delete
                    && this.webAnnotations.perms[row.iam_ns_id]?.namespace_scoped_delete;
    },
    handleClearSearchData() {
      this.searchKeyword = '';
      this.searchIngress();
    },
  },
};
</script>

<style scoped>
    @import '../../ingress.css';
</style>
