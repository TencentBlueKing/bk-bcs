<!-- eslint-disable vue/multi-word-component-names -->
<!-- eslint-disable max-len -->
<template>
  <BcsContent hide-back :title="$t('nav.templateSet')">
    <div v-bkloading="{ isLoading: isLoading, opacity: 0.1 }">
      <template v-if="!isLoading">
        <Row class="mb-[16px]">
          <template #left>
            <div class="left">
              <bk-dropdown-menu
                :key="projectKind"
                ref="dropdown"
                v-authority="{
                  actionId: 'templateset_create',
                  permCtx: {
                    resource_type: 'project',
                    project_id: projectId
                  }
                }"
                trigger="click">
                <bk-button type="primary" icon-right="icon-angle-down" slot="dropdown-trigger">
                  <span class="text-[14px]"> {{$t('deploy.templateset.addTemplateSet')}}</span>
                </bk-button>
                <ul class="bk-dropdown-list" slot="dropdown-content">
                  <li>
                    <a href="javascript:;" @click.stop.prevent="addTemplate('k8sTemplatesetDeployment')">{{$t('dashboard.label.formMode')}}</a>
                  </li>
                  <li>
                    <a href="javascript:;" @click="addTemplate('K8sYamlTemplateset')">{{$t('dashboard.label.yamlMode')}}</a>
                  </li>
                </ul>
              </bk-dropdown-menu>
            </div>
          </template>
          <template #right>
            <div class="right">
              <p class="biz-tpl-desc" v-if="templatesetCount">{{$t('deploy.templateset.total')}} <strong>{{templatesetCount}}</strong> {{$t('units.suffix.units')}}</p>
              <div class="biz-search-input" style="width: 300px;">
                <bkbcs-input
                  right-icon="bk-icon icon-search"
                  clearable
                  :placeholder="$t('generic.placeholder.search')"
                  v-model="searchKeyword"
                  @enter="search"
                  @clear="clearSearch">
                </bkbcs-input>
              </div>
            </div>
          </template>
        </Row>
        <svg style="display: none;">
          <title>{{$t('deploy.templateset.icon')}}</title>
          <symbol id="biz-set-icon" viewBox="0 0 32 32">
            <path d="M6 3v3h-3v23h23v-3h3v-23h-23zM24 24v3h-19v-19h19v16zM27 24h-1v-18h-18v-1h19v19z"></path>
            <path d="M13.688 18.313h-6v6h6v-6z"></path>
            <path d="M21.313 10.688h-6v13.625h6v-13.625z"></path>
            <path d="M13.688 10.688h-6v6h6v-6z"></path>
          </symbol>
        </svg>

        <div class="biz-namespace">
          <table class="bk-table biz-templateset-table" id="templateset-table">
            <thead>
              <tr>
                <th class="template-name">{{$t('deploy.templateset.templateSetName')}}</th>
                <th class="template-type">{{$t('generic.label.kind')}}</th>
                <th class="template-container">{{$t('dashboard.workload.container.title')}} / {{$t('k8s.image')}}</th>
                <th class="template-action">{{$t('generic.label.action')}}</th>
              </tr>
            </thead>
          </table>

          <div id="templateset-container" ref="templatesetContainer">
            <table class="bk-table biz-templateset-table">
              <tbody>
                <template v-if="templateList.length">
                  <tr :key="templateIndex" v-for="(template, templateIndex) in templateList">
                    <td colspan="6">
                      <table class="biz-inner-table">
                        <tr>
                          <td class="data">
                            <div class="data-wrapper">
                              <a
                                href="javascript:void(0);" class="title" style="font-weight: normal;"
                                v-authority="{
                                  clickable: getAuthority('templateset_view', template.id),
                                  resourceName: template.name,
                                  actionId: 'templateset_view',
                                  disablePerms: true,
                                  permCtx: {
                                    project_id: projectId,
                                    template_id: template.id
                                  }
                                }" @click.stop.prevent="goTemplateIndex(template)">
                                {{template.name}}
                              </a>

                              <p class="vertion">{{$t('deploy.templateset.latestVersion')}}：{{template.latest_show_version || '--'}}</p>
                            </div>
                          </td>
                          <td class="type">
                            <div class="type-wrapper">
                              <span class="type-name">{{editMode[template.edit_mode]}}</span>
                            </div>
                          </td>
                          <td class="service">
                            <div class="service-wrapper">
                              <template v-if="template.images.length">
                                <div :key="imageIndex" v-for="(image, imageIndex) in template.images">
                                  <bcs-popover :content="image || '--'" placement="top">
                                    <span class="biz-text-wrapper" style="min-height: 16px;">{{image || '--'}}</span>
                                  </bcs-popover>
                                </div>
                              </template>
                              <template v-else>
                                -- / --
                              </template>
                            </div>
                          </td>
                          <td class="operate">
                            <div class="operate-wrapper">
                              <template>
                                <template v-if="template.latest_show_version_id === -1">
                                  <bcs-popover :content="$t('deploy.templateset.draftTemplateSet')" placement="top">
                                    <bk-button disabled="disabled" style="width: 124px;">
                                      {{$t('deploy.templateset.tempalte')}}
                                    </bk-button>
                                  </bcs-popover>
                                </template>
                                <template v-else>
                                  <bk-button
                                    v-authority="{
                                      clickable: getAuthority('templateset_instantiate', template.id),
                                      actionId: 'templateset_instantiate',
                                      resourceName: template.name,
                                      disablePerms: true,
                                      permCtx: {
                                        project_id: projectId,
                                        template_id: template.id
                                      }
                                    }" @click="goCreateInstance(template)" style="width: 124px;">
                                    {{$t('deploy.templateset.tempalte')}}
                                  </bk-button>
                                </template>
                              </template>

                              <PopoverSelector trigger="click" class="ml5" offset="0, 5">
                                <bk-button icon-right="icon-angle-down" class="min-w-[82px]">
                                  <span class="f14">{{$t('deploy.templateset.more')}}</span>
                                </bk-button>
                                <ul class="bk-dropdown-list" slot="content">
                                  <li
                                    class="bcs-dropdown-item"
                                    v-authority="{
                                      clickable: getAuthority('templateset_copy', template.id),
                                      actionId: 'templateset_copy',
                                      resourceName: template.name,
                                      disablePerms: true,
                                      permCtx: {
                                        project_id: projectId,
                                        template_id: template.id
                                      }
                                    }"
                                    v-if="template.edit_mode !== 'yaml'"
                                    @click="showCopy(template)">
                                    {{$t('deploy.templateset.copyTemplateSet')}}
                                  </li>
                                  <li
                                    class="bcs-dropdown-item"
                                    v-authority="{
                                      clickable: getAuthority('templateset_delete', template.id),
                                      actionId: 'templateset_delete',
                                      resourceName: template.name,
                                      disablePerms: true,
                                      permCtx: {
                                        project_id: projectId,
                                        template_id: template.id
                                      }
                                    }" @click="removeTemplate(template)">
                                    {{$t('deploy.templateset.deleteTemplateSet')}}
                                  </li>
                                  <li
                                    class="bcs-dropdown-item"
                                    v-if="template.edit_mode !== 'yaml'"
                                    v-authority="{
                                      clickable: getAuthority('templateset_instantiate', template.id),
                                      actionId: 'templateset_instantiate',
                                      resourceName: template.name,
                                      disablePerms: true,
                                      permCtx: {
                                        project_id: projectId,
                                        template_id: template.id
                                      }
                                    }"
                                    @click="showChooseDialog(template)">
                                    {{$t('deploy.templateset.deleteInstance')}}
                                  </li>
                                </ul>
                              </PopoverSelector>
                            </div>
                          </td>
                        </tr>
                      </table>
                    </td>
                  </tr>
                </template>
                <template v-else>
                  <tr>
                    <td colspan="6">
                      <div class="biz-guide-box" style="margin: 0 20px;">
                        <bcs-exception type="empty" scene="part"></bcs-exception>
                      </div>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>
        </div>
      </template>
      <p :class="['biz-more-btn f13', { actived: isScrollLoading }]" v-bkloading="{ isLoading: isScrollLoading, opacity: 1 }" v-show="pageConf.hasNext && templateList.length" @click="loadNextPage">{{$t('deploy.templateset.clickLoadMore')}}</p>
      <p class="biz-no-data" v-if="!pageConf.hasNext && templateList.length">{{$t('deploy.templateset.allDataLoaded')}}</p>
    </div>

    <bk-dialog
      :is-show.sync="copyDialogConf.isShow"
      :width="copyDialogConf.width"
      :close-icon="copyDialogConf.closeIcon"
      :ext-cls="'biz-config-templateset-copy-dialog'"
      :has-header="false"
      :quick-close="false">
      <!-- eslint-disable-next-line vue/no-useless-template-attributes -->
      <template slot="content" v-bkloading="{ isLoading: isCopying }">
        <div class="bk-form bk-form-vertical biz-instance-num-form mb20">
          <div class="bk-form-item">
            <label class="bk-label">
              {{$t('nav.templateSet')}}【{{copyDialogConf.title}}】{{$t('deploy.templateset.copyTo')}}：
            </label>
            <div class="bk-form-content">
              <bkbcs-input maxlength="30" :placeholder="$t('deploy.templateset.newTemplateSetName')" v-model="copyName" />
            </div>
          </div>
        </div>
      </template>
      <div slot="footer">
        <div class="bk-dialog-outer">
          <bk-button
            type="primary"
            :loading="isCopying"
            @click="confirmCopyTemplate">
            {{$t('generic.button.confirm')}}
          </bk-button>
          <bk-button
            type="button"
            :disabled="isCopying"
            @click="hideCopy">
            {{$t('generic.button.cancel')}}
          </bk-button>
        </div>
      </div>
    </bk-dialog>

    <bk-dialog
      :is-show.sync="delInstanceDialogConf.isShow"
      :width="delInstanceDialogConf.width"
      :title="delInstanceDialogConf.title"
      :close-icon="delInstanceDialogConf.closeIcon"
      :quick-close="false"
      :ext-cls="'biz-config-templateset-del-instance-dialog'"
      @cancel="delInstanceDialogConf.isShow = false">
      <!-- eslint-disable-next-line vue/no-useless-template-attributes -->
      <template slot="content" v-bkloading="{ isLoading: isDeleting }">
        <div class="content-inner">
          <div class="bk-form bk-form-vertical" style="margin-bottom: 20px;">
            <div class="bk-form-item">
              <label class="bk-label">
                {{$t('deploy.templateset.templatesetVer')}}
              </label>
              <div class="bk-form-content">
                <div class="bk-dropdown-box" style="width: 240px;" @click="fetchTemplatesetVerList">
                  <bk-selector
                    :placeholder="$t('generic.placeholder.select')"
                    :selected.sync="tplsetVerIndex"
                    :list="tplsetVerList"
                    @item-selected="changeTplset"
                    :is-loading="isLoadingTplsetVer"
                    :ext-cls="'ver-selector'">
                  </bk-selector>
                </div>
              </div>
            </div>
            <div class="bk-form-item">
              <label class="bk-label">
                {{$t('deploy.templateset.templateset')}}
              </label>
              <div class="bk-form-content">
                <bk-selector
                  :placeholder="$t('generic.placeholder.select')"
                  :selected.sync="tplIndex"
                  :list="tplList"
                  @item-selected="changeTpl"
                  :ext-cls="'ver-selector'">
                </bk-selector>
              </div>
            </div>
          </div>
          <template v-if="showNamespaceContainer && !candidateNamespaceList.length">
            <transition name="fadeDown">
              <div class="content-trigger-wrapper" style="text-align: center;">
                {{$t('deploy.templateset.noNamespaceData')}}
              </div>
            </transition>
          </template>
          <template v-else>
            <div tag="div" name="fadeDown">
              <div
                :key="index" class="content-trigger-wrapper"
                :class="item.isOpen ? 'open' : ''"
                v-for="(item, index) in candidateNamespaceList">
                <div class="content-trigger" @click="triggerHandler(item, index)">
                  <div class="left-area">
                    <div class="label">
                      <span :class="['biz-env-label mr5', { 'stag': item.environment !== 'prod', 'prod': item.environment === 'prod' }]">{{item.environment === 'prod' ? $t('cluster.tag.prod') : $t('cluster.tag.debug')}}</span>
                      {{item.cluster_name}}
                    </div>
                    <div class="checker-inner">
                      <a href="javascript:;" class="bk-text-button" @click.stop="selectAll(item, index)">{{$t('generic.button.selectAll')}}</a>
                      <a href="javascript:;" class="bk-text-button" @click.stop="selectInvert(item, index)">{{$t('deploy.templateset.reelection')}}</a>
                    </div>
                  </div>
                  <i v-if="item.isOpen" class="bcs-icon bcs-icon-angle-up trigger active"></i>
                  <i v-else class="bcs-icon bcs-icon-angle-down trigger"></i>
                </div>
                <div class="biz-namespace-wrapper" v-if="item.results.length" :style="{ display: item.isOpen ? '' : 'none' }">
                  <div class="namespace-inner">
                    <template v-for="(namespace, i) in item.results">
                      <div :key="i" v-if="namespace.isExist" class="candidate-namespace exist" style="position: relative;">
                        <bcs-popover :content="namespace.name" :delay="500" placement="bottom">
                          <div class="candidate-namespace-name">
                            <span>{{namespace.name}}</span>
                            <span class="icon" v-if="namespace.isExist"><i class="bcs-icon bcs-icon-check-1"></i></span>
                          </div>
                        </bcs-popover>
                      </div>
                      <div
                        :key="i" v-else class="candidate-namespace"
                        :class="namespace.isChoose ? 'active' : ''"
                        @click="selectNamespaceInDialog(index, namespace, i)" style="position: relative;">
                        <bcs-popover :content="namespace.name" :delay="500" placement="bottom">
                          <div class="candidate-namespace-name">
                            <span>{{namespace.name}}</span>
                            <span class="icon" v-if="namespace.isChoose"><i class="bcs-icon bcs-icon-check-1"></i></span>
                          </div>
                        </bcs-popover>
                      </div>
                    </template>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </template>
      <div slot="footer">
        <div class="bk-dialog-outer">
          <template v-if="isDeleting">
            <bk-button type="primary" disabled>
              {{$t('generic.status.deleting')}}
            </bk-button>
            <bk-button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
              {{$t('generic.button.cancel')}}
            </bk-button>
          </template>
          <template v-else>
            <bk-button
              type="primary" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary"
              @click="confirmDelInstance">
              {{$t('generic.button.submit')}}
            </bk-button>
            <bk-button type="button" @click="cancelDelInstance">
              {{$t('generic.button.cancel')}}
            </bk-button>
          </template>
        </div>
      </div>
    </bk-dialog>

    <bk-dialog
      :is-show.sync="delTemplateDialogConf.isShow"
      :width="delTemplateDialogConf.width"
      :close-icon="delTemplateDialogConf.closeIcon"
      :ext-cls="'biz-config-templateset-copy-dialog'"
      :has-header="false"
      :quick-close="false">
      <!-- eslint-disable-next-line vue/no-useless-template-attributes -->
      <template slot="content" style="padding: 0 20px;">
        <div style="color: #63656e; font-size: 20px">
          {{$t('nav.templateSet')}}【{{delTemplateDialogConf.title}}】{{$t('generic.label.version')}}：
        </div>
        <ul style="margin: 10px 0; font-size: 14px; color: #63656e;">
          <li v-for="(key, index) in Object.keys(delTemplateDialogConf.existVersion)" :key="index">
            <span>{{key}}：</span>
            <span>{{$t('deploy.templateset.with')}} {{delTemplateDialogConf.existVersion[key]}} {{$t('deploy.templateset.InstancesCount')}}</span>
          </li>
        </ul>
        <div class="mb20">
          {{$t('deploy.templateset.deleteInstanceBeforeTemplateSet')}}
        </div>
      </template>
      <div slot="footer">
        <div class="bk-dialog-outer">
          <bk-button
            type="primary" style="width: 96px;"
            @click="delTemplateConfirm">
            {{$t('deploy.templateset.deleteInstance')}}
          </bk-button>
          <bk-button type="button" @click="delTemplateCancel">
            {{$t('generic.button.cancel')}}
          </bk-button>
        </div>
      </div>
    </bk-dialog>
  </BcsContent>
</template>

<script>
import { catchErrorHandler } from '@/common/util';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import PopoverSelector from '@/components/popover-selector.vue';

export default {
  components: { BcsContent, Row, PopoverSelector },
  data() {
    return {
      fileImportIndex: 0,
      zipTooltipText: this.$t('deploy.templateset.importFromExportedZip'),
      webAnnotationsPerms: {},
      isLoading: true,
      isImportLoading: false,
      searchKeyword: '',
      templateList: [],
      templateListCache: [],
      linkAppList: [],
      curVersion: 0,
      editMode: {
        page_form: this.$t('dashboard.label.formMode'),
        yaml: this.$t('dashboard.label.yamlMode'),
      },
      resourceList: [
        {
          id: 'application',
          name: 'Application',
        },
        {
          id: 'deployment',
          name: 'Deployment',
        },
        {
          id: 'service',
          name: 'Service',
        },
        {
          id: 'configmap',
          name: 'Configmap',
        },
        {
          id: 'secret',
          name: 'Secret',
        },
      ],
      resourceIndex: -1,
      pageConf: {
        total: 0,
        totalPage: 1,
        pageSize: 10,
        hasNext: true,
        curPage: 1,
      },
      copyDialogConf: {
        isShow: false,
        width: 400,
        title: '',
        closeIcon: false,
      },
      copyName: '',
      curCopyTemplate: null,
      isCopying: false,
      delInstanceDialogConf: {
        isShow: false,
        width: 912,
        title: '',
        closeIcon: false,
      },
      candidateNamespaceList: [], // 弹层中的 namespace 集合
      namespaceListTmp: {}, // 在弹层中选择的 namespace 缓存
      showNamespaceContainer: false,
      curDelInstanceTemplate: null,
      tplsetVerList: [],
      tplsetVerIndex: -1,
      tplsetVerId: '',
      tplList: [],
      tplIndex: -1,
      tplId: '',
      tplCategory: '',
      isDeleting: false,
      delTemplateDialogConf: {
        isShow: false,
        width: 650,
        title: '',
        closeIcon: false,
        existVersion: {},
        template: {},
      },
      isLoadingTplsetVer: false,
      isScrollLoading: false,
      isListReload: false,
      templatesetCount: 0,
    };
  },
  computed: {
    projectId() {
      return this.$route.params.projectId;
    },
    projectCode() {
      return this.$route.params.projectCode;
    },
    projectKind() {
      return this.$store.state.curProject.kind;
    },
    curProject() {
      return this.$store.state.curProject;
    },
  },
  watch: {
    searchKeyword(newVal, oldVal) {
      // 如果删除，为空时触发搜索
      if (oldVal && !newVal) {
        this.search();
      }
    },
  },
  mounted() {
    this.getTemplateList();
    this.$nextTick(() => {
      window.scrollTo(0, 0);
      window.onscroll = () => {
        if (this.isLoading) {
          return false;
        }
        this.loadNextPage();
      };
    });
  },
  beforeRouteLeave(to, from, next) {
    window.onscroll = null;
    next();
  },
  beforeDestroy() {
    window.onscroll = null;
  },
  methods: {
    getAuthority(actionId, templateId) {
      return !!this.webAnnotationsPerms[templateId]?.[actionId];
    },

    /**
             * 加载下一页数据
             */
    loadNextPage() {
      if (this.isScrollLoading || !this.pageConf.hasNext || this.isListReload) {
        return false;
      }
      if (this.scrollBottom()) {
        this.isScrollLoading = true;

        // eslint-disable-next-line no-plusplus
        this.pageConf.curPage++;
        this.getTemplateList();
      }
    },

    scrollBottom() {
      // eslint-disable-next-line max-len
      return document.documentElement.clientHeight + window.scrollY >= (document.documentElement.scrollHeight || document.documentElement.clientHeight);
    },

    /**
             * 确认删除模板集
             * @param {Object} template 当前模板集对象
             */
    async removeTemplate(template) {
      try {
        // 先检测当前模板集是否存在实例
        const res = await this.$store.dispatch('templateset/getExistVersion', {
          projectId: this.projectId,
          templateId: template.id,
        });

        // 如果没有实例，可删除模板集，否则调用删除实例
        if (!Object.keys(res.data.exist_version || {}).length) {
          // eslint-disable-next-line @typescript-eslint/no-this-alias
          const me = this;
          me.$bkInfo({
            title: this.$t('generic.title.confirmDelete'),
            clsName: 'biz-remove-dialog',
            subHeader: new me.$createElement('p', {
              class: 'biz-confirm-desc',
            }, `${this.$t('deploy.templateset.confirmDeleteTemplateSet')}【${template.name}】?`),
            async confirmFn() {
              me.deleteTemplate(template);
            },
          });
        } else {
          this.delTemplateDialogConf.isShow = true;
          this.delTemplateDialogConf.title = template.name;
          this.delTemplateDialogConf.template = Object.assign({}, template);
          this.delTemplateDialogConf.existVersion = Object.assign({}, res.data.exist_version);
        }
      } catch (e) {
        catchErrorHandler(e, this);
      }
    },

    /**
             * 确认删除模板集
             * @param {Object} template 当前模板集对象
             */
    async deleteTemplate(template) {
      try {
        await this.$store.dispatch('templateset/removeTemplate', {
          projectId: this.projectId,
          templateId: template.id,
        });

        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.delete'),
        });
        const updateList = this.templateList.filter(item => item.id !== template.id);
        this.templateList.splice(0, this.templateList.length, ...updateList);
        this.templateListCache.splice(0, this.templateListCache.length, ...updateList);
        this.pageConf.total = this.pageConf.total - 1;
      } catch (e) {
        catchErrorHandler(e, this);
      }
    },

    /**
             * 创建实例
             * @param  {object} template 当前模板集对象
             */
    async goCreateInstance(template) {
      this.$router.push({
        name: 'instantiation',
        params: {
          projectId: this.projectId,
          templateId: template.id,
          curTemplate: template,
          projectCode: this.projectCode,
        },
      });
    },

    /**
             * 关闭 delTemplateDialog
             */
    delTemplateCancel() {
      this.delTemplateDialogConf.isShow = false;
      this.delTemplateDialogConf.title = '';
      this.delTemplateDialogConf.template = Object.assign({}, {});
      this.delTemplateDialogConf.existVersion = Object.assign({}, {});
    },

    /**
             * delTemplateDialog confirm
             */
    delTemplateConfirm() {
      const template = Object.assign({}, this.delTemplateDialogConf.template);
      this.delTemplateDialogConf.isShow = false;
      this.delTemplateDialogConf.title = '';
      this.delTemplateDialogConf.template = Object.assign({}, {});
      this.delTemplateDialogConf.existVersion = Object.assign({}, {});
      this.showChooseDialog(template);
    },

    /**
             * 显示复制弹层
             * @param {Object} template 当前 template
             */
    async showCopy(template) {
      this.curCopyTemplate = Object.assign({}, template);
      this.copyDialogConf.isShow = true;
      this.copyDialogConf.title = template.name;
    },

    /**
             * 模板集复制提交前检测
             */
    confirmCopyTemplate() {
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const me = this;

      const copyName = me.copyName.trim();

      if (!copyName) {
        me.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.enterCopyName'),
        });
        return;
      }

      if (copyName.toLowerCase() === me.curCopyTemplate.name.trim().toLowerCase()) {
        me.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.copyNameNotSame'),
        });
        return;
      }

      if (copyName.length > 30) {
        me.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.nameMax30Chars'),
        });
        return;
      }

      this.copyTemplate();
    },

    /**
             * 提交模板集复制
             */
    async copyTemplate() {
      const backup = [];
      backup.splice(0, backup.length, ...this.templateList);

      this.isCopying = true;
      try {
        await this.$store.dispatch('templateset/copyTemplate', {
          projectId: this.projectId,
          templateId: this.curCopyTemplate.id,
          name: this.copyName,
        });

        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.copy'),
        });

        this.templateList.splice(0, this.templateList.length, ...[]);
        this.templateListCache.splice(0, this.templateListCache.length, ...[]);
        this.isLoading = true;
        this.getTemplateList(true);
      } catch (e) {
        this.isLoading = false;
        this.templateList.splice(0, this.templateList.length, ...backup);
        this.templateListCache.splice(0, this.templateListCache.length, ...backup);
        catchErrorHandler(e, this);
      } finally {
        this.hideCopy();
        this.isCopying = false;
      }
    },

    /**
             * 关闭复制弹层
             */
    hideCopy() {
      this.curCopyTemplate = Object.assign({}, {});
      this.copyName = '';
      this.copyDialogConf.isShow = false;
    },

    /**
             * 显示选择命名空间弹层
             *
             * @param {Object} template 当前 template
             */
    async showChooseDialog(template) {
      // 清除弹层中的选中状态，不需要清除已选择的 ns 的状态
      this.clearCandidateNamespaceStatus();

      // 之前没选择过，那么展开第一个
      this.delInstanceDialogConf.title = `${this.$t('generic.button.delete')}【${template.name}】${this.$t('deploy.templateset.templateSetInstances')}`;
      this.delInstanceDialogConf.isShow = true;
      this.curDelInstanceTemplate = Object.assign({}, template);

      this.fetchTemplatesetVerList();
    },

    /**
             * 关闭选择命名空间弹层
             */
    cancelDelInstance() {
      this.tplsetVerList.splice(0, this.tplsetVerList.length, ...[]);
      this.tplsetVerIndex = -1;
      this.tplsetVerId = '';
      this.tplList.splice(0, this.tplList.length, ...[]);
      this.tplIndex = -1;
      this.tplId = '';
      this.tplCategory = '';
      this.candidateNamespaceList.splice(0, this.candidateNamespaceList.length, ...[]);
      this.showNamespaceContainer = false;
      this.delInstanceDialogConf.isShow = false;
      this.namespaceListTmp = Object.assign({}, {});
    },

    /**
             * 获取模板集版本
             */
    async fetchTemplatesetVerList() {
      try {
        this.isLoadingTplsetVer = true;
        this.tplsetVerList.splice(0, this.tplsetVerList.length, ...[]);
        const res = await this.$store.dispatch('templateset/getTemplatesetVerList', {
          projectId: this.projectId,
          templateId: this.curDelInstanceTemplate.id,
          hasFilter: true,
        });

        const list = res.data.results || [];
        list.forEach((item) => {
          this.tplsetVerList.push({
            id: item.id,
            name: item.version,
            version: item.version,
            template_id: item.template_id,
          });
        });
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        setTimeout(() => {
          this.isLoadingTplsetVer = false;
        }, 600);
      }
    },

    /**
             * 访问模板集首页
             * @param  {object} template 当前模板集对象
             */
    async goTemplateIndex(template) {
      if (this.projectKind === PROJECT_K8S || this.projectKind === PROJECT_TKE) {
        if (template.edit_mode === 'yaml') {
          this.$router.push({
            name: 'K8sYamlTemplateset',
            params: {
              projectId: this.projectId,
              projectCode: this.projectCode,
              templateId: template.id,
            },
          });
        } else {
          this.$router.push({
            name: 'k8sTemplatesetDeployment',
            params: {
              projectId: this.projectId,
              projectCode: this.projectCode,
              templateId: template.id,
            },
          });
        }
      } else {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.getProjectInfoFail'),
        });
      }
    },

    /**
             * 切换模板集下拉框
             * @param {number} index 索引
             * @param {Object} data 当前下拉框数据
             */
    async changeTplset(index, data) {
      try {
        const res = await this.$store.dispatch('templateset/getTemplateInsResourceById', {
          projectId: this.projectId,
          templateId: data.template_id,
          showVerName: data.name,
        });

        const list = [];
        Object.keys(res.data.data || {}).forEach((key) => {
          res.data.data[key].forEach((item) => {
            list.push({
              id: item.id,
              name: `${key}:${item.name}`,
            });
          });
        });

        if (list.length) {
          list.unshift({
            id: 0,
            name: this.$t('generic.label.total'),
          });
        }

        this.tplsetVerId = data.id;
        this.tplList.splice(0, this.tplList.length, ...list);
        this.tplIndex = -1;
        this.tplId = '';
        this.tplCategory = '';
        this.showNamespaceContainer = false;
        this.candidateNamespaceList.splice(0, this.candidateNamespaceList.length, ...[]);
      } catch (e) {
        catchErrorHandler(e, this);
      }
    },

    /**
             * 切换模板下拉框
             * @param {number} index 索引
             * @param {Object} data 当前下拉框数据
             */
    async changeTpl(index, data) {
      if (this.tplId === data.id) {
        return;
      }

      if (data.id === 0) {
        this.tplId = 0;
        this.tplCategory = 'ALL';
      } else {
        this.tplId = data.id;
        this.tplCategory = data.name.split(':')[0];
      }

      try {
        const res = await this.$store.dispatch('templateset/getNamespaceList4DelInstance', {
          projectId: this.projectId,
          tplMusterId: this.curDelInstanceTemplate.id,
          tplsetVerId: this.tplsetVerId,
          tplId: this.tplId === 0 ? this.tplList[1].id : this.tplId,
          category: this.tplCategory,
        });

        this.candidateNamespaceList.splice(0, this.candidateNamespaceList.length, ...[]);
        this.namespaceListTmp = Object.assign({}, {});

        const list = res.data;
        list.forEach((item, index) => {
          // 展开第一个
          this.candidateNamespaceList.push({ ...item, isOpen: index === 0 });
        });
        this.showNamespaceContainer = true;
      } catch (e) {
        catchErrorHandler(e, this);
      }
    },

    /**
             * 清除弹层中 namespace trigger 的展开以及 namespace 的选中
             */
    clearCandidateNamespaceStatus() {
      const list = this.candidateNamespaceList;
      list.forEach((item) => {
        item.isOpen = false;
        item.results.forEach((ns) => {
          ns.isChoose = false;
        });
      });

      this.candidateNamespaceList.splice(0, this.candidateNamespaceList.length, ...list);
      this.namespaceListTmp = Object.assign({}, {});
    },

    /**
             * 在弹层中选择命名空间
             * @param {number} index candidateNamespaceList 的索引
             * @param {Object} namespace 当前点击的这个 namespace
             * @param {number} i 当前点击的这个 namespace 在 item.results 的索引
             */
    selectNamespaceInDialog(index, namespace, i) {
      namespace.isChoose = !namespace.isChoose;
      this.$set(this.candidateNamespaceList[index].results, i, namespace);
      if (this.namespaceListTmp[`${namespace.env_type}_${namespace.id}`]) {
        delete this.namespaceListTmp[`${namespace.env_type}_${namespace.id}`];
      } else {
        this.namespaceListTmp[`${namespace.env_type}_${namespace.id}`] = {
          ...namespace,
          candidateIndex: index,
          index: i,
        };
      }
    },

    /**
             * 在弹层中全选命名空间
             * @param {Object} item 当前的 candidateNamespace 对象
             * @param {number} index 当前的 candidateNamespace 对象在 candidateNamespaceList 中的索引
             */
    selectAll(item, index) {
      this.collapseTrigger();
      item.results.forEach((ns, i) => {
        ns.isChoose = true;
        this.namespaceListTmp[`${ns.env_type}_${ns.id}`] = {
          ...ns,
          candidateIndex: index,
          index: i,
        };
      });
      item.isOpen = true;
      this.$set(this.candidateNamespaceList, index, item);
    },

    /**
             * 在弹层中反选命名空间
             * @param {Object} item 当前的 candidateNamespace 对象
             * @param {number} index 当前的 candidateNamespace 对象在 candidateNamespaceList 中的索引
             */
    selectInvert(item, index) {
      this.collapseTrigger();
      item.results.forEach((ns, i) => {
        ns.isChoose = !ns.isChoose;
        if (this.namespaceListTmp[`${ns.env_type}_${ns.id}`]) {
          delete this.namespaceListTmp[`${ns.env_type}_${ns.id}`];
        } else {
          this.namespaceListTmp[`${ns.env_type}_${ns.id}`] = {
            ...ns,
            candidateIndex: index,
            index: i,
          };
        }
      });
      item.isOpen = true;
      this.$set(this.candidateNamespaceList, index, item);
    },

    /**
             * 收起所有的 trigger
             */
    collapseTrigger() {
      const list = this.candidateNamespaceList;
      list.forEach((item) => {
        item.isOpen = false;
      });
      this.candidateNamespaceList.splice(0, this.candidateNamespaceList.length, ...list);
    },

    /**
             * 选择命名空间弹层 trigger 点击事件
             * @param {Object} item 当前 namespace 对象
             * @param {number} index 当前 namespace 对象的索引
             */
    triggerHandler(item, index) {
      this.collapseTrigger();
      item.isOpen = !item.isOpen;
      this.$set(this.candidateNamespaceList, index, item);
    },

    /**
             * 删除命名空间弹层确认
             */
    async confirmDelInstance() {
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const me = this;
      const list = Object.keys(this.namespaceListTmp);

      if (this.tplsetVerIndex === -1) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.pleaseSelectTemplateVersion'),
        });
        return;
      }
      if (this.tplIndex === -1) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.selectTemplate'),
        });
        return;
      }
      if (list.length === 0) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('dashboard.ns.validate.emptyNs'),
        });
        return;
      }

      this.$bkInfo({
        title: this.$t('deploy.templateset.confirmDeleteInstance'),
        async confirmFn() {
          me.deleteInstance();
        },
      });
    },

    /**
             * 删除实例
             */
    async deleteInstance() {
      const list = Object.keys(this.namespaceListTmp);

      const params = {
        projectId: this.projectId,
        tplMusterId: this.curDelInstanceTemplate.id,
        tplsetVerId: this.tplsetVerId,
        tplId: this.tplId,
        namespace_list: [],
        category: this.tplCategory,
      };

      list.forEach((key) => {
        params.namespace_list.push(this.namespaceListTmp[key].id);
      });

      let isRedirect = false;

      if (this.tplId === 0) {
        const idList = [];
        this.tplList.forEach((item, index) => {
          // 第 0 个是 all
          if (index !== 0) {
            const prefix = item.name.split(':')[0];
            idList.push({
              id: item.id,
              category: prefix,
            });
            if (prefix === 'application' || prefix === 'deployment') {
              isRedirect = true;
            }
          }
        });
        params.id_list = idList;
        params.category = 'all';
      }

      this.isDeleting = true;

      try {
        await this.$store.dispatch('templateset/delNamespaceInDelInstance', params);

        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.delete'),
        });

        if (this.tplCategory === 'application' || this.tplCategory === 'deployment' || isRedirect) {
          this.$router.push({
            name: 'deployments',
            params: {
              projectId: this.projectId,
              projectCode: this.projectCode,
              tplsetId: this.curDelInstanceTemplate.id,
            },
          });
        }
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        this.isDeleting = false;
        if (!isRedirect) {
          this.isLoading = true;
          this.templateList.splice(0, this.templateList.length, ...[]);
          this.templateListCache.splice(0, this.templateListCache.length, ...[]);
          this.cancelDelInstance();
          this.getTemplateList(true);
        }
      }
    },

    /**
             * 重置分页配置数据
             */
    resetPageConf() {
      this.pageConf.total = 0;
      this.pageConf.curPage = 1;
      this.pageConf.pageSize = 10;
      this.pageConf.hasNext = false;
    },

    /**
             * 搜索列表
             */
    search() {
      this.resetPageConf();
      this.getTemplateList(true);
    },

    // 获取模板集列表搜索参数
    getQueryString() {
      // 当前分页offset
      const curPageOffset = (this.pageConf.curPage - 1) * this.pageConf.pageSize;
      // 当前实际offset
      const curOffset = this.templateList.length;
      // 由于本地删除，实际offset有可能少于分页offset
      const offset = Math.min(curPageOffset, curOffset);

      if (!this.searchKeyword) {
        return `offset=${offset}&limit=${this.pageConf.pageSize}`;
      }
      return `offset=${offset}&limit=${this.pageConf.pageSize}&search=${this.searchKeyword}`;
    },

    getElementTop(element) {
      let actualTop = element.offsetTop;
      let current = element.offsetParent;

      while (current !== null) {
        actualTop += current.offsetTop;
        current = current.offsetParent;
      }
      return actualTop;
    },

    /**
             * 获取模板集列表
             * @params {boolean} isRelaod 是否重新加载数据
             */
    async getTemplateList(isReload) {
      let lastOffsetTop = 0;
      const { projectId } = this;
      const templatesetDoms = document.querySelectorAll('.biz-inner-table');

      // 清空数据，重置分页配置信息
      if (isReload) {
        this.resetPageConf();
        this.isListReload = true;
        window.scrollTo(0, 0);
      }

      if (templatesetDoms.length) {
        lastOffsetTop = this.getElementTop(templatesetDoms[templatesetDoms.length - 1]);
      }

      try {
        const queryString = this.getQueryString();
        const res = await this.$store.dispatch('templateset/getTemplateList', { projectId, queryString });
        const { data } = res;

        this.templatesetCount = data.count;
        // 清空数据，重置分页配置信息
        if (isReload) {
          this.templateList.splice(0, this.templateList.length);
        }

        data.results.forEach((item) => {
          const images = [];
          item.containers.forEach((item) => {
            const containerImage = `${item.name} / ${item.image}`;
            if (!images.includes(containerImage)) {
              images.push(containerImage);
            }
          });
          item.images = images;
          this.templateList.push(item);
        });

        this.webAnnotationsPerms = Object.assign({}, this.webAnnotationsPerms, res.web_annotations?.perms || {});
        this.pageConf.hasNext = data.has_next;
        this.pageConf.total = data.count;

        this.$nextTick(() => {
          if (!isReload && lastOffsetTop) {
            window.scrollTo(0, lastOffsetTop - 10); // 回滚到最上一页数据底部
          }
        });
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        this.isLoading = false;
        setTimeout(() => {
          this.isListReload = false;
          this.isScrollLoading = false;
        }, 500);
      }
    },

    /**
             * 清除搜索
             */
    clearSearch() {
      this.searchKeyword = '';
      this.search();
    },

    /**
             * 创建表单模板集
             */
    async addTemplate(type) {
      this.$router.push({
        name: type,
        params: {
          projectId: this.projectId,
          projectCode: this.projectCode,
          templateId: 0,
        },
      });
    },

    getImportFileList(zip) {
      for (const key in zip) {
        const file = zip[key];
        // 文件
        if (file.name) {
          this.importFileList.push(file);
        } else {
          this.getImportFileList(file, key);
        }
      }
    },

    async readFile(file) {
      if (!file) return;

      return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onerror = () => {
          reject(new Error('read file error'));
        };
        reader.onloadend = (event) => {
          resolve(event.target.result);
        };
        reader.readAsText(file);
      });
    },
  },
};
</script>
<style scoped>
    @import './templateset.css';
</style>
