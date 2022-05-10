<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-configuration-instantiation-title">
                <i class="bcs-icon bcs-icon-arrows-left back" @click="goTemplateset(false)"></i>
                <span @click="refreshCurRouter">{{$t('模板实例化')}}</span>
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper pt0">
            <app-exception v-if="exceptionCode" :type="exceptionCode.code" :text="exceptionCode.msg"></app-exception>
            <div v-else class="biz-configuration-instantiation-wrapper">
                <div class="biz-tip mt20 mb15">{{$t('模板实例化操作即平台通过用户配置的模板，生成对应的资源YAML文件，并将它们下发到指定集群的命名空间下。资源创建成功后，可在"应用"和"网络"中查看资源实例详情。')}}</div>
                <div class="biz-configuration-instantiation-header">
                    <div class="left">
                        <svg style="display: none;">
                            <title>{{$t('模板集默认图标')}}</title>
                            <symbol id="biz-set-icon" viewBox="0 0 32 32">
                                <path d="M6 3v3h-3v23h23v-3h3v-23h-23zM24 24v3h-19v-19h19v16zM27 24h-1v-18h-18v-1h19v19z"></path>
                                <path d="M13.688 18.313h-6v6h6v-6z"></path>
                                <path d="M21.313 10.688h-6v13.625h6v-13.625z"></path>
                                <path d="M13.688 10.688h-6v6h6v-6z"></path>
                            </symbol>
                        </svg>
                        <div class="info">
                            <svg class="logo"><use xlink:href="#biz-set-icon"></use></svg>
                            <div class="title" :title="curTemplate.name">{{curTemplate.name || '--'}}</div>
                            <div class="creater" :title="curTemplate.creator">{{$t('创建人')}}：{{curTemplate.creator || '--'}}</div>
                        </div>
                        <div class="desc" :title="curTemplate.desc">
                            <span>{{$t('简介')}}：</span>
                            {{curTemplate.desc || '--'}}
                        </div>
                    </div>
                    <div class="right">
                        <div class="top">
                            <div class="inner">
                                <div class="inner-item" style="width: 180px;">
                                    <label class="title">{{$t('模板集版本')}}</label>
                                    <bk-selector
                                        :placeholder="$t('请选择')"
                                        :selected.sync="tplsetVerIndex"
                                        :list="tplsetVerList"
                                        :setting-key="'show_version_id'"
                                        @item-selected="changeTplset">
                                    </bk-selector>
                                    <label class="tip">{{$t('创建后启动实例')}}</label>
                                </div>
                                <div class="inner-item" style="width: 280px;">
                                    <label class="title">{{$t('模板')}}</label>
                                    <bk-selector :placeholder="$t('请选择要实例化的模板')"
                                        :searchable="true"
                                        :selected.sync="tplIndex"
                                        :setting-key="'settingKey'"
                                        :list="tplList"
                                        :multi-select="true"
                                        @item-selected="multiSelect">
                                    </bk-selector>
                                    <a href="javascript:void(0);" class="bk-text-button select-all-tpl is-disabled" v-if="tplList.length === 0">
                                        {{$t('全选模板')}}
                                    </a>
                                    <a href="javascript:void(0);" class="bk-text-button select-all-tpl" v-else @click="selectAllTpl">
                                        {{isSelectAllTpl ? $t('清空全选') : $t('全选模板')}}
                                    </a>
                                </div>
                            </div>
                        </div>
                        <div class="bottom">
                            <div class="inner">
                                <bk-button type="default" @click="showChooseDialog">
                                    {{$t('选择运行的集群及命名空间')}}
                                </bk-button>
                                <div class="selected-namespace-wrapper biz-configuration-instantiation-dialog">
                                    <div class="content-inner m0 pl0 pb0">
                                        <div :key="index" class="content-trigger-wrapper open" v-for="(cluster, index) in selectedNamespaceCluster" style="cursor: default;">
                                            <div class="content-trigger">
                                                <div class="left-area" style="border-right: none; width: auto;">
                                                    <div class="label">
                                                        <template v-if="cluster.environment !== 'prod'">
                                                            <span class="biz-env-label mr5 stag">{{$t('测试')}}</span>
                                                        </template>
                                                        <template v-else>
                                                            <span class="biz-env-label mr5 prod">{{$t('正式')}}</span>
                                                        </template>
                                                        <span class="biz-text-wrapper">{{cluster.cluster_name}}</span>
                                                        <span class="choose-num">{{cluster.namespaceList.length}} {{ isEn ? '' : 个}}</span>
                                                    </div>
                                                </div>
                                            </div>
                                            <div class="biz-namespace-wrapper pt20">
                                                <template v-for="(item, itemIndex) in selectedNamespaceList">
                                                    <div :key="itemIndex"
                                                        class="selected-namespace-item"
                                                        :class="item.isSelected ? 'active' : ''"
                                                        v-if="item.cluster_id === cluster.cluster_id"
                                                        @click="previewNamespace(item, itemIndex)">
                                                        <div class="selected-namespace-item-name" :class="item.isSelected ? 'active' : ''">
                                                            {{item.name}}
                                                        </div>
                                                    </div>
                                                </template>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                <transition name="fade">
                    <div class="biz-configuration-instantiation-content" v-if="previewShow"
                        v-bkloading="{ isLoading: previewLoading }">
                        <div class="header">
                            {{previewTitle}}
                        </div>
                        <div class="content">
                            <div class="form-wrapper">
                                <template v-if="invalidNsList.length && invalidNsList.indexOf(previewNs.name) > -1">
                                    <div style="text-align: center; margin: 0 auto; margin-bottom: 20px; margin-top: -20px;">
                                        {{$t('命令空间')}} [{{previewNs.name}}] {{$t('没有相关联的 LoadBalance，请先到 网络 -> LoadBalance 页面关联')}}
                                    </div>
                                </template>
                                <template v-else>
                                    <div :key="index" class="form-item" v-for="(lb, index) in lbServiceListInPage">
                                        <div class="left" v-if="lb[0]">
                                            <label class="form-label">
                                                {{lb[0].key}}：
                                            </label>
                                            <div class="form-item-inner">
                                                <bk-selector
                                                    :placeholder="$t('请选择')"
                                                    :ext-cls="'dropdown'"
                                                    :selected.sync="lbSelectData[previewNs.name][lb[0].key]"
                                                    :list="lb[0].value"
                                                    :setting-key="'lb_id'"
                                                    :display-key="'lb_name'"
                                                    @item-selected="changeLbServiceSelect">
                                                </bk-selector>
                                            </div>
                                        </div>
                                        <div class="left" v-if="lb[1]">
                                            <label class="form-label">
                                                {{lb[1].key}}：
                                            </label>
                                            <div class="form-item-inner">
                                                <bk-selector
                                                    :placeholder="$t('请选择')"
                                                    :ext-cls="'dropdown'"
                                                    :selected.sync="lbSelectData[previewNs.name][lb[1].key]"
                                                    :list="lb[1].value"
                                                    :setting-key="'lb_id'"
                                                    :display-key="'lb_name'"
                                                    @item-selected="changeLbServiceSelect">
                                                </bk-selector>
                                            </div>
                                        </div>
                                    </div>
                                    <template v-if="previewNs.variableList && previewNs.variableList.length && !isYamlMode">
                                        <div class="form-item">
                                            <label class="form-label">
                                                {{$t('变量')}}：
                                            </label>
                                            <div class="form-item-inner">
                                                <div class="biz-key-value-item" v-for="(variable, index) in previewNs.variableList" :key="index">
                                                    <input type="text" class="bk-form-input" disabled :value="`${variable.name}(${variable.key})`">
                                                    <span class="equals-sign">=</span>
                                                    <input type="text" class="bk-form-input right" :placeholder="$t('值')"
                                                        v-model="variable.value" @keyup="variableValChange" />
                                                </div>
                                            </div>
                                        </div>
                                    </template>
                                    <div class="form-item" v-if="previewList.length">
                                        <label class="form-label">
                                            {{$t('预览')}}：
                                        </label>
                                        <div class="form-item-inner" style="width: 100px">
                                            <bk-tab
                                                class="biz-scroll-tab biz-tab-container"
                                                :type="'fill'"
                                                :size="'small'"
                                                :active-name="previewList[0].name"
                                                :key="previewList.length"
                                                @tab-changed="tabChange">
                                                <bk-tab-panel :key="index" :name="item.name" :title="item.title" :tag="item.tag" v-for="(item, index) in previewList">
                                                    <template slot="label">
                                                        <div class="biz-tab-label">
                                                            <span class="bk-panel-title">{{item.title}}</span>
                                                            <div class="bk-panel-label">
                                                                <div class="bk-panel-tag">{{item.tag}}</div>
                                                            </div>
                                                        </div>
                                                    </template>
                                                    <div class="biz-code-wrapper">
                                                        <div class="build-code-fullscreen" :title="$t('全屏')"
                                                            @click="setFullScreen(index)">
                                                            <i class="bcs-icon bcs-icon-full-screen"></i>
                                                        </div>
                                                        <ace
                                                            :value="editorConfig.values[index]"
                                                            :width="editorConfig.width"
                                                            :height="editorConfig.height"
                                                            :lang="editorConfig.lang"
                                                            :read-only="editorConfig.readOnly"
                                                            :full-screen="editorConfig.fullScreen"
                                                            @change-annotation="changeAnnotation(index, ...arguments)"
                                                            @init="editorInitAfter">
                                                        </ace>
                                                    </div>
                                                </bk-tab-panel>
                                            </bk-tab>
                                        </div>
                                    </div>
                                </template>
                            </div>
                            <div class="biz-tip-box" v-if="previewErrorMessage" style="margin-bottom: 40px;">
                                <div class="wrapper danger">
                                    <i class="bcs-icon bcs-icon-exclamation-circle-shape"></i>
                                    {{previewErrorMessage}}
                                </div>
                            </div>
                        </div>
                    </div>
                </transition>
                <div class="create-wrapper">
                    <bk-button type="primary" :title="$t('创建')" @click="createInstance">
                        {{$t('创建')}}
                    </bk-button>
                    <bk-button type="default" :title="$t('取消')" @click="goTemplateset(true)">
                        {{$t('取消')}}
                    </bk-button>
                </div>
            </div>
        </div>

        <bk-dialog
            :is-show.sync="dialogConf.isShow"
            :width="dialogConf.width"
            :title="dialogConf.title"
            :close-icon="dialogConf.closeIcon"
            :confirm="$t('提交')"
            :cancel="$t('取消')"
            :ext-cls="'biz-configuration-instantiation-dialog'"
            :quick-close="false"
            @confirm="confirmSelect"
            @cancel="dialogConf.isShow = false">
            <template slot="content">
                <div class="content-inner" :style="{ 'max-height': '420px', 'overflow': 'auto' }">
                    <div class="namespace-types">
                        <span class="bk-outline"><i class="bcs-icon bcs-icon-circle-shape"></i>{{$t('未实例化过')}}</span>
                        <span class="bk-default"><i class="bcs-icon bcs-icon-circle-shape"></i>{{$t('已实例化过')}}</span>
                    </div>
                    <div :key="index" class="content-trigger-wrapper" :class="item.isOpen ? 'open' : ''" v-for="(item, index) in candidateNamespaceList" v-show="!curClusterId || (curClusterId && item.cluster_id === curClusterId)">
                        <div class="content-trigger" @click="triggerHandler(item, index)">
                            <div class="left-area" style="border-right: none;">
                                <div class="label">
                                    <span :class="['biz-env-label mr5', { 'stag': item.environment !== 'prod', 'prod': item.environment === 'prod' }]">{{item.environment_name}}</span>
                                    <span class="biz-text-wrapper" style="max-width: 300px;">{{item.name}}</span>
                                    <span class="choose-num">{{$t('已经选择')}} {{item.results.filter(ns => ns.isChoose).length}} {{$t('个1')}}</span>
                                </div>
                                <div class="checker-inner" v-if="!isYamlMode">
                                    <a href="javascript:;" class="bk-text-button" @click.stop="selectAll(item, index)">{{$t('全选')}}</a>
                                    <a href="javascript:;" class="bk-text-button" @click.stop="selectInvert(item, index)">{{$t('反选')}}</a>
                                </div>
                            </div>
                            <i v-if="item.isOpen" class="bcs-icon bcs-icon-angle-up trigger active" style="border-left: 1px solid #eee;"></i>
                            <i v-else class="bcs-icon bcs-icon-angle-down trigger" style="border-left: 1px solid #eee;"></i>
                        </div>
                        <div class="biz-namespace-wrapper" :style="{ display: item.isOpen ? '' : 'none' }">
                            <div class="namespace-inner">
                                <template v-for="(namespace, i) in item.results">
                                    <div :key="i" v-if="namespace.isExist" class="candidate-namespace exist">
                                        <bcs-popover :content="namespace.message" :delay="500" placement="bottom">
                                            <div class="candidate-namespace-name">
                                                <span>{{namespace.name}}</span>
                                                <span class="icon" v-if="namespace.isExist"><i class="bcs-icon bcs-icon-check-1"></i></span>
                                            </div>
                                        </bcs-popover>
                                    </div>
                                    <div :key="i" v-else class="candidate-namespace"
                                        :title="namespace.name"
                                        :class="namespace.isChoose ? 'active' : ''"
                                        @click="selectNamespaceInDialog(index, namespace, i)">
                                        <bcs-popover :content="namespace.name" :delay="500" placement="bottom">
                                            <div class="candidate-namespace-name">
                                                <span>{{namespace.name}}</span>
                                                <span class="icon" v-if="namespace.isChoose"><i class="bcs-icon bcs-icon-check-1"></i></span>
                                            </div>
                                        </bcs-popover>
                                    </div>
                                </template>
                                <div class="candidate-namespace add-namespace" :title="$t('新增命名空间')" v-if="!isSharedCluster">
                                    <bcs-popover ref="addNamespaceNode" theme="light" :delay="120000" placement="top-end" ext-cls="add-namespace-popover" :controlled="true" @on-show="showAddNamespace(index)">
                                        <div class="candidate-namespace-name" @click="triggerAddNamespace(index)">
                                            <img src="@/images/plus.svg" class="add-btn" />
                                        </div>
                                        <template slot="content">
                                            <div class="title">{{$t('新增命名空间')}}</div>
                                            <input type="text" ref="addNamespaceInputNode" :placeholder="$t('输入名称')" class="bk-form-input ns-name" v-model="namespaceName" v-if="dialogConf.loading" disabled />
                                            <input type="text" ref="addNamespaceInputNode" :placeholder="$t('输入名称')" class="bk-form-input ns-name" v-model="namespaceName" v-else />
                                            <a href="javascript:;" class="bk-text-button link disabled" v-if="dialogConf.loading">
                                                {{$t('更多设置')}}
                                                <img src="@/images/link-disabled.svg" />
                                            </a>
                                            <a href="javascript:;" class="bk-text-button link" @click="goNamespace" v-else>
                                                {{$t('更多设置')}}
                                                <img src="@/images/link.svg" />
                                            </a>
                                            <div class="operate">
                                                <a href="javascript:;" class="bk-text-button disabled" v-if="dialogConf.loading">{{$t('保存中...')}}</a>
                                                <a href="javascript:;" class="bk-text-button" v-else @click="addNamespace(item, index)">{{$t('保存')}}</a>
                                                <a href="javascript:;" class="bk-text-button disabled" v-if="dialogConf.loading" @click="cancelNamespace">{{$t('取消')}}</a>
                                                <a href="javascript:;" class="bk-text-button" v-else @click="cancelNamespace">{{$t('取消')}}</a>
                                            </div>
                                        </template>
                                    </bcs-popover>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </template>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="goNamespaceDialogConf.isShow"
            :width="goNamespaceDialogConf.width"
            :title="goNamespaceDialogConf.title"
            :quick-close="false"
            @cancel="goNamespaceDialogConf.isShow = false">
            <template slot="content">
                <div style="text-align: center; padding-bottom: 15px;">
                    {{$t('您没有可用的命名空间，请创建或申请已有命名空间使用')}}
                </div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <bk-button type="primary" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary"
                        @click="goNamespace" style="width: 110px;">
                        {{$t('创建或申请')}}
                    </bk-button>
                    <bk-button type="primary" class="bk-dialog-btn bk-dialog-btn-cancel" @click="hideNamesapceDialog">
                        {{$t('取消')}}
                    </bk-button>
                </div>
            </div>
        </bk-dialog>

        <div :title="$t('关闭全屏')" @click="cancelFullScreen" class="biz-configuration-instantiation-cancel-fullscreen"
            v-if="editorConfig.fullScreen">
            <i class="bcs-icon bcs-icon-close"></i>
        </div>
    </div>
</template>

<script>
    import yamljs from 'js-yaml'
    import ace from '@/components/ace-editor'
    import { catchErrorHandler } from '@/common/util'
    import { mapGetters } from 'vuex'

    const ARR = [
        'Application',
        'Deployment',
        'Service',
        'ConfigMap',
        'Secret',
        'DaemonSet',
        'Job',
        'StatefulSet',
        'Ingress'
    ]

    const ABBR_ARR = [
        'app',
        'dep',
        'svc',
        'cm',
        'srt',
        'ds',
        'job',
        'sts',
        'Ing'
    ]

    const toString = Object.prototype.toString

    export default {
        components: {
            ace
        },
        data () {
            return {
                tplList: [],
                // 存放给的是模板的 id 集合
                tplIndex: [],
                // 存放选择的yaml模板资源
                yamlTplResource: [],
                tplsetVerList: [],
                // 存放给的是模板集的 show_version_id
                tplsetVerIndex: -1,
                // 存放给的是模板集的 id
                tplsetVerId: -1,
                dialogConf: {
                    isShow: false,
                    width: 912,
                    // width: 895,
                    title: this.$t('选择运行的集群及命名空间'),
                    closeIcon: false,
                    loading: false
                },
                existList: [],
                curTemplateTmp: {},
                // 弹层中的 namespace 集合
                candidateNamespaceList: [],
                // 在弹层中选择的 namespace 缓存
                namespaceListTmp: {},
                // 弹层点击确定后把 namespaceListTmp 赋值给 selectedNamespaceList，用于显示在页面上
                selectedNamespaceList: [],
                selectedNamespaceCluster: {},
                previewTitle: '',
                previewList: [],
                previewLoading: false,
                previewShow: false,
                previewNs: null,
                lbServiceList: [],
                lbServiceListInPage: [],
                // 映射 lb 的数据，下拉框里面的是 id
                lbServiceListInPageTmp: {},
                lbSelectData: {},
                // 没有 lb 信息的 ns 的集合
                invalidNsList: [],
                // 创建后启动实例多选框
                // isStart: true,
                createInstanceLoading: false,
                variableMap: {},
                bkMessageInstance: null,
                exceptionCode: null,
                editorConfig: {
                    width: '100%',
                    height: '100%',
                    lang: 'json',
                    readOnly: true,
                    fullScreen: false,
                    values: [],
                    editors: []
                },
                // 模板多选下拉框选择的值，提交给后端的
                instanceEntity: {},
                projectId: '',
                // aceAnnotationErrorMsg: '',
                goNamespaceDialogConf: {
                    isShow: false,
                    width: 500,
                    title: this.$t('没有命名空间'),
                    closeIcon: true,
                    hasFooter: false,
                    hasHeader: false
                },
                curProject: null,
                isSelectAllTpl: false,
                previewErrorMessage: '',
                namespaceName: '',
                PROJECT_MESOS: window.PROJECT_MESOS
            }
        },
        computed: {
            projectCode () {
                return this.$route.params.projectCode
            },
            templateId () {
                return this.$route.params.templateId
            },
            // 从应用列表跳转过来才会有 category
            category () {
                return this.$route.params.category
            },
            // 从应用列表跳转过来才会有 tmplAppId
            tmplAppId () {
                return this.$route.params.tmplAppId
            },
            // 从应用列表跳转过来才会有 tmplAppName
            tmplAppName () {
                return this.$route.params.tmplAppName
            },
            curTemplate: {
                get () {
                    if (this.$route.params.curTemplate) {
                        this.curTemplateTmp = this.$route.params.curTemplate
                    }
                    return this.curTemplateTmp
                },
                set (val) {
                    this.curTemplateTmp = val
                    return this.curTemplateTmp
                }
            },
            isYamlMode () {
                return this.curTemplate.edit_mode === 'yaml'
            },
            onlineProjectList () {
                return this.$store.state.sideMenu.onlineProjectList
            },
            curShowVersionId () {
                return this.$route.params.curShowVersionId
            },
            curClusterId () {
                return this.$store.state.curClusterId
            },
            curProjectId () {
                return this.$store.state.curProjectId
            },
            isEn () {
                return this.$store.state.isEn
            },
            ...mapGetters('cluster', ['isSharedCluster'])
        },
        created () {
            // router > localStorage > onlineProjectList[0]
            const len = this.onlineProjectList.length
            if (len) {
                this.projectId = this.$route.params.projectId
                    || this.curProjectId
                    || this.onlineProjectList[0].project_id

                for (let i = 0; i < len; i++) {
                    const project = this.onlineProjectList[i]
                    if (project.project_id === this.projectId) {
                        this.curProject = Object.assign({}, project)
                        break
                    }
                }
                // k8s
                if (this.curProject.kind === PROJECT_K8S || this.curProject.kind === PROJECT_TKE) {
                    this.editorConfig.lang = 'yaml'
                } else { // mesos
                    this.editorConfig.lang = 'json'
                }
                if (Object.keys(this.curTemplate).length === 0) {
                    this.fetchTemplate()
                }
                this.fetchTemplatesetVerList()
                this.fetchNamespaceList()
            }
        },
        methods: {
            /**
             * 模板集的排序，顺序依次为：
             * Application/app
             * Deployment/dep
             * Service/svc
             * ConfigMap/cm
             * Secret/srt
             * DaemonSet/ds
             * Job/job
             * StatefulSet/sts
             * Ingress/Ing
             *
             * @param {Array} list 要排序的数组
             * @param {string} key 要排序的 key
             * @param {boolean} isAbbr 是否是简写
             *
             * @return {Array} 结果
             */
            sortTplType (list, key, isAbbr) {
                const arr = isAbbr ? ABBR_ARR : ARR
                return list.sort((a, b) => arr.indexOf(a[key]) - arr.indexOf(b[key]) >= 0)
            },

            /**
             * ace 编辑器 annotation change 回调
             *
             * @param {number} index 当前是哪个 ace editor 的索引
             * @param {Array} annotations annotations 数据
             */
            changeAnnotation (index, annotations) {
                this.editorConfig.editors[index].gotoLine(annotations[0].row + 1, annotations[0].column, true)
            },

            /**
             *  编辑器初始化之后的回调函数
             *  @param editor - 编辑器对象
             */
            editorInitAfter (editor) {
                this.editorConfig.editors.push(editor)
                setTimeout(() => {
                    editor.resize(true)
                    editor.gotoLine(0, 0, true)
                }, 100)
            },

            /**
             * 刷新当前 router
             */
            refreshCurRouter () {
                typeof this.$parent.$parent.refreshRouterView === 'function' && this.$parent.$parent.refreshRouterView()
            },

            /**
             * 获取当前 template 的信息。名称、创建人、简介等信息
             */
            async fetchTemplate () {
                try {
                    const res = await this.$store.dispatch('configuration/getTemplateById', {
                        projectId: this.projectId,
                        templateId: this.templateId
                    })
                    this.curTemplate = Object.assign({}, res.data)
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 加载所有命名空间列表，不在点击选择命名空间按钮点击事件中加载，而是提前加载
             * 是为了防止选择命名空间弹层异步加载数据时高度变化
             * 获取到 candidateNamespaceList，弹层中所有的命名空间集合
             */
            async fetchNamespaceList () {
                try {
                    const res = await this.$store.dispatch('configuration/getAllNamespaceList', {
                        projectId: this.projectId,
                        group_by: 'cluster_name',
                        perm_can_use: 1,
                        with_perms: false
                    })

                    const list = res.data

                    list.forEach(item => {
                        this.candidateNamespaceList.push({ ...item, isOpen: true })
                    })
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 获取模板集版本。模板集版本下拉框数据
             */
            async fetchTemplatesetVerList () {
                try {
                    const res = await this.$store.dispatch('configuration/getTemplatesetVerList', {
                        projectId: this.projectId,
                        templateId: this.templateId
                    })
                    const list = res.data.results || []
                    list.forEach(item => {
                        this.tplsetVerList.push({
                            id: item.id,
                            name: item.version,
                            show_version_id: item.show_version_id,
                            show_version_name: item.show_version_name
                        })
                    })

                    // 如果带上版本ID，则默认选中相应版本
                    if (this.curShowVersionId) {
                        this.tplsetVerIndex = this.curShowVersionId

                        // 根据已经选中的版本来查找资源列表
                        this.tplsetVerList.forEach((item, index) => {
                            if (item.show_version_id === this.tplsetVerIndex) {
                                this.changeTplset(item.show_version_id, item)
                            }
                        })
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            async setFormTemplateList (settingKey, data) {
                try {
                    const action = 'configuration/getTemplateListById'
                    const params = {
                        projectId: this.projectId,
                        tplVerId: data.id
                    }
                    const res = await this.$store.dispatch(action, params)
                    const tplList = []
                    const tplIndex = []
                    const tplData = res.data.data || {}

                    Object.keys(tplData).forEach(key => {
                        const obj = {
                            name: key,
                            children: []
                        }
                        tplData[key].forEach(item => {
                            // 默认全选中
                            // tplIndex.push(key + '-' + item.id)
                            obj.children.push({
                                id: item.id,
                                settingKey: key + '-' + item.id + '-' + settingKey,
                                name: item.name,
                                type: key
                            })
                        })
                        tplList.push(obj)
                    })

                    this.sortTplType(tplList, 'name')

                    this.tplIndex.splice(0, this.tplIndex.length, ...tplIndex)
                    this.tplList.splice(0, this.tplList.length, ...tplList)
                    this.instanceEntity = Object.assign({}, tplData)
                } catch (e) {
                    console.error(e)
                }
            },

            async setYamlTemplateList (settingKey, data) {
                try {
                    const action = 'k8sTemplate/getYamlTemplateDetailByVersion'
                    const params = {
                        projectId: this.projectId,
                        templateId: this.templateId,
                        versionId: data.show_version_id,
                        withFileContent: false
                    }
                    const res = await this.$store.dispatch(action, params)
                    const tplList = []
                    const tplIndex = []
                    const yamlTplResource = []
                    const tplData = res.data.data || {}
                    res.data.template_files.forEach(resource => {
                        const key = resource.resource_name
                        const obj = {
                            name: key,
                            children: []
                        }

                        resource.files.forEach(file => {
                            obj.children.push({
                                id: file.id,
                                settingKey: key + '-' + file.id + '-' + settingKey,
                                name: file.name,
                                type: key
                            })
                        })

                        tplList.push(obj)
                    })

                    this.sortTplType(tplList, 'name')

                    this.tplIndex.splice(0, this.tplIndex.length, ...tplIndex)
                    this.yamlTplResource.splice(0, this.yamlTplResource.length, ...yamlTplResource)
                    this.tplList.splice(0, this.tplList.length, ...tplList)
                    this.instanceEntity = Object.assign({}, tplData)
                } catch (e) {
                    console.error(e)
                }
            },

            /**
             * 切换模板集下拉框，获取模板下拉框的数据
             *
             * @param {number} settingKey setting-key
             * @param {Object} data 当前下拉框数据
             */
            async changeTplset (settingKey, data) {
                this.tplIndex.splice(0, this.tplIndex.length, ...[])
                this.yamlTplResource.splice(0, this.yamlTplResource.length, ...[])
                this.tplList.splice(0, this.tplList.length, ...[])
                this.tplsetVerId = data.id + '-' + data.show_version_id
                this.clearCandidateNamespaceStatus()
                this.clearNamespaceStatus()

                this.previewTitle = ''
                this.previewNs = Object.assign({}, {})
                this.previewList.splice(0, this.previewList.length, ...[])
                this.editorConfig.editors.splice(0, this.editorConfig.editors.length, ...[])
                this.editorConfig.values.splice(0, this.editorConfig.values.length, ...[])
                this.previewShow = false

                // 清空已经选择的 namespace
                this.selectedNamespaceList.splice(0, this.selectedNamespaceList.length, ...[])
                this.selectedNamespaceCluster = {}

                // 表单模式的模板集和yaml模式的模板集作区分
                try {
                    if (this.isYamlMode) {
                        await this.setYamlTemplateList(settingKey, data)
                    } else {
                        await this.setFormTemplateList(settingKey, data)
                    }
                    this.isSelectAllTpl = false
                } catch (e) {
                    console.error(e)
                }
            },

            /**
             * 模板 多选下拉框选择事件
             *
             * @param {Array} index 索引的数组
             * @param {Array} data 选择的数据对象的数组
             */
            multiSelect (index, data) {
                const ret = {}
                data.forEach(item => {
                    if (!ret[item.type]) {
                        ret[item.type] = []
                    }
                    ret[item.type].push({
                        id: item.id,
                        name: item.name
                    })
                })

                let count = 0
                this.tplList.forEach(item => {
                    count += (item.children || []).length
                })
                this.isSelectAllTpl = data.length === count

                this.instanceEntity = Object.assign({}, ret)

                this.clearCandidateNamespaceStatus()
                this.clearNamespaceStatus()
                this.previewTitle = ''
                this.previewNs = Object.assign({}, {})
                this.previewList.splice(0, this.previewList.length, ...[])
                this.editorConfig.editors.splice(0, this.editorConfig.editors.length, ...[])
                this.editorConfig.values.splice(0, this.editorConfig.values.length, ...[])
                this.previewShow = false

                // 清空已经选择的 namespace
                this.selectedNamespaceList.splice(0, this.selectedNamespaceList.length, ...[])
                this.yamlTplResource.splice(0, this.yamlTplResource.length, ...[])
                this.selectedNamespaceCluster = {}

                // yaml模式
                if (this.isYamlMode) {
                    data.forEach(item => {
                        let resource = this.yamlTplResource.find(resource => resource.resource_name === item.type)

                        if (!resource) {
                            resource = {
                                resource_name: item.type,
                                files: []
                            }
                            this.yamlTplResource.push(resource)
                        }

                        resource.files.push({
                            id: item.id,
                            name: item.name
                        })
                    })
                }
            },

            /**
             * 全选模板
             */
            selectAllTpl () {
                if (this.isSelectAllTpl) {
                    this.isSelectAllTpl = false
                    this.tplIndex = [...[]]
                    this.yamlTplResource = [...[]]
                    this.instanceEntity = Object.assign({}, {})
                    return
                }

                const ret = {}
                const tplIndex = []
                const yamlTplResource = []
                this.tplList.forEach(item => {
                    if (!ret[item.name]) {
                        ret[item.name] = []
                    }

                    const yamlResource = {
                        resource_name: item.name,
                        files: []
                    }
                    item.children.forEach(child => {
                        ret[item.name].push({
                            id: child.id,
                            name: child.name
                        })
                        tplIndex.push(child.settingKey)
                        yamlResource.files.push({
                            id: child.id,
                            name: child.name
                        })
                    })
                    yamlTplResource.push(yamlResource)
                })

                this.tplIndex = [...tplIndex]
                this.yamlTplResource = [...yamlTplResource]
                this.instanceEntity = Object.assign({}, ret)

                this.clearCandidateNamespaceStatus()
                this.clearNamespaceStatus()
                this.previewTitle = ''
                this.previewNs = Object.assign({}, {})
                this.previewList.splice(0, this.previewList.length, ...[])
                this.editorConfig.editors.splice(0, this.editorConfig.editors.length, ...[])
                this.editorConfig.values.splice(0, this.editorConfig.values.length, ...[])
                this.previewShow = false

                // 清空已经选择的 namespace
                this.selectedNamespaceList.splice(0, this.selectedNamespaceList.length, ...[])
                this.selectedNamespaceCluster = {}
                this.isSelectAllTpl = true
            },

            /**
             * 根据模板集 id 获取已经被使用过的 namespace
             */
            async fetchExistNamespace () {
                try {
                    const res = await this.$store.dispatch('configuration/getExistNamespace', {
                        projectId: this.projectId,
                        tplVerId: this.tplsetVerId.split('-')[0],
                        instanceEntity: this.instanceEntity
                    })
                    const existNamespaceList = res.data.ns_resources || {}
                    const list = []
                    const existList = Object.keys(existNamespaceList)
                    list.splice(0, 0, ...this.candidateNamespaceList)
                    list.forEach(item => {
                        (item.results || []).forEach(ns => {
                            if (existList.indexOf(String(ns.id)) > -1) {
                                const message = existNamespaceList[ns.id].join(', ')
                                ns.isExist = true
                                ns.message = `${this.$t('命名空间')}【${ns.name}】${this.$t('下存在')} ${message.replace(/K8s/ig, '')}${this.$t('类型的同名实例')}, ${this.$t('无法再次实例化')}`
                            } else {
                                ns.isExist = false
                            }
                        })
                        item.results.sort((cur, next) => {
                            // boolean（强制转数整数）相减，isEist为true排后面
                            return cur.isExist - next.isExist
                        })
                        item.isOpen = true
                    })
                    this.existList.splice(0, this.existList.length, ...existList)
                    this.candidateNamespaceList.splice(0, this.candidateNamespaceList.length, ...list)
                    this.candidateNamespaceList = this.isSharedCluster ? this.candidateNamespaceList.filter(i => i.is_shared) : this.candidateNamespaceList.filter(i => !i.is_shared)
                } catch (e) {
                    console.error(e)
                } finally {
                    setTimeout(() => {
                        this.dialogConf.loading = false
                    }, 300)
                }
            },

            /**
             * 显示选择命名空间弹层
             */
            async showChooseDialog () {
                if (this.tplsetVerId === -1) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择模板集版本')
                    })
                    return
                }

                if (!this.tplIndex.length) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择要实例化的模板')
                    })
                    return
                }

                const { candidateNamespaceList, selectedNamespaceList } = this

                if (!candidateNamespaceList.length) {
                    this.goNamespaceDialogConf.isShow = true
                    return
                }

                this.dialogConf.isShow = true
                this.dialogConf.loading = true

                await this.fetchExistNamespace()

                // 清除弹层中的选中状态，不需要清除已选择的 ns 的状态
                this.clearCandidateNamespaceStatus()

                // 默认展开所有
                candidateNamespaceList.forEach(item => {
                    item.isOpen = true
                })
                if (selectedNamespaceList.length) {
                    selectedNamespaceList.forEach(ns => {
                        candidateNamespaceList[ns.candidateIndex].isOpen = true
                        candidateNamespaceList[ns.candidateIndex].results[ns.index].isChoose = true
                        this.$set(candidateNamespaceList, ns.candidateIndex, candidateNamespaceList[ns.candidateIndex])

                        this.namespaceListTmp[`${ns.env_type}_${ns.id}`] = {
                            ...ns,
                            candidateIndex: ns.candidateIndex,
                            index: ns.index
                        }
                    })
                }
            },

            /**
             * 清除弹层中 namespace trigger 的展开以及 namespace 的选中
             */
            clearCandidateNamespaceStatus () {
                const list = this.candidateNamespaceList
                list.forEach(item => {
                    item.isOpen = false
                    item.results.forEach(ns => {
                        ns.isChoose = false
                    })
                })

                this.candidateNamespaceList.splice(0, this.candidateNamespaceList.length, ...list)
                this.namespaceListTmp = {}
            },

            /**
             * 清除 selectedNamespaceList 中的选中状态
             */
            clearNamespaceStatus () {
                const selectedNamespaceList = this.selectedNamespaceList
                this.selectedNamespaceCluster = {}
                selectedNamespaceList.forEach(ns => {
                    ns.isSelected = false
                    if (!this.selectedNamespaceCluster[ns.cluster_id]) {
                        this.selectedNamespaceCluster[ns.cluster_id] = {
                            cluster_id: ns.cluster_id,
                            environment: ns.environment,
                            cluster_name: ns.cluster_name,
                            namespaceList: [ns]
                        }
                    } else {
                        this.selectedNamespaceCluster[ns.cluster_id].namespaceList.push(ns)
                    }
                })
                this.selectedNamespaceList.splice(0, this.selectedNamespaceList.length, ...selectedNamespaceList)
            },

            /**
             * 收起所有的 trigger
             */
            collapseTrigger () {
                const list = this.candidateNamespaceList
                list.forEach(item => {
                    item.isOpen = false
                })
                this.candidateNamespaceList.splice(0, this.candidateNamespaceList.length, ...list)
            },

            /**
             * 选择命名空间弹层 trigger 点击事件
             *
             * @param {Object} item 当前 namespace 对象
             * @param {number} index 当前 namespace 对象的索引
             */
            triggerHandler (item, index) {
                // 展开时滚动条回到顶部
                // document.querySelectorAll('.namespace-wrapper .namespace-inner')[1].scrollTop = 0
                // this.collapseTrigger()
                item.isOpen = !item.isOpen
                this.$set(this.candidateNamespaceList, index, item)

                this.cancelNamespace()
            },

            /**
             * 在弹层中选择命名空间
             *
             * @param {number} index candidateNamespaceList 的索引
             * @param {Object} namespace 当前点击的这个 namespace
             * @param {number} i 当前点击的这个 namespace 在 item.results 的索引
             */
            selectNamespaceInDialog (index, namespace, i) {
                // yaml模式单选
                if (this.isYamlMode) {
                    this.candidateNamespaceList.forEach(cluster => {
                        cluster.results.forEach(namespace => {
                            if (namespace.isChoose) {
                                namespace.isChoose = false
                            }
                        })
                    })
                    this.namespaceListTmp = {}
                }
                namespace.isChoose = !namespace.isChoose

                this.$set(this.candidateNamespaceList[index].results, i, namespace)

                if (this.namespaceListTmp[`${namespace.env_type}_${namespace.id}`]) {
                    delete this.namespaceListTmp[`${namespace.env_type}_${namespace.id}`]
                } else {
                    this.namespaceListTmp[`${namespace.env_type}_${namespace.id}`] = {
                        ...namespace,
                        candidateIndex: index,
                        index: i
                    }
                }
            },

            /**
             * 在弹层中全选命名空间
             *
             * @param {Object} item 当前的 candidateNamespace 对象
             * @param {number} index 当前的 candidateNamespace 对象在 candidateNamespaceList 中的索引
             */
            selectAll (item, index) {
                this.collapseTrigger()
                item.results.forEach((ns, i) => {
                    if (this.existList.indexOf(ns.id) < 0) {
                        ns.isChoose = true
                        this.namespaceListTmp[`${ns.env_type}_${ns.id}`] = {
                            ...ns,
                            candidateIndex: index,
                            index: i
                        }
                    }
                })
                item.isOpen = true
                this.$set(this.candidateNamespaceList, index, item)
            },

            /**
             * 在弹层中反选命名空间
             *
             * @param {Object} item 当前的 candidateNamespace 对象
             * @param {number} index 当前的 candidateNamespace 对象在 candidateNamespaceList 中的索引
             */
            selectInvert (item, index) {
                this.collapseTrigger()
                item.results.forEach((ns, i) => {
                    if (this.existList.indexOf(ns.id) < 0) {
                        ns.isChoose = !ns.isChoose
                        if (this.namespaceListTmp[`${ns.env_type}_${ns.id}`]) {
                            delete this.namespaceListTmp[`${ns.env_type}_${ns.id}`]
                        } else {
                            this.namespaceListTmp[`${ns.env_type}_${ns.id}`] = {
                                ...ns,
                                candidateIndex: index,
                                index: i
                            }
                        }
                    }
                })
                item.isOpen = true
                this.$set(this.candidateNamespaceList, index, item)
            },

            /**
             * 选择命名空间弹层确认
             */
            async confirmSelect () {
                const list = Object.keys(this.namespaceListTmp)
                if (list.length === 0) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择命名空间')
                    })
                    this.dialogConf.isShow = false
                    return
                }

                const namespaces = []
                list.forEach(item => {
                    namespaces.push(item.split('_')[1])
                })

                this.dialogConf.loading = true
                this.previewErrorMessage = ''
                try {
                    const res = await this.$store.dispatch('configuration/getLbVariable', {
                        projectId: this.projectId,
                        tplVerId: this.tplsetVerId.split('-')[0],
                        namespaces: namespaces.join(','),
                        instanceEntity: this.instanceEntity
                    })
                    this.lbServiceList.splice(0, this.lbServiceList.length, ...(res.data.lb_services || []))

                    // 前一次的 lbSelectData
                    const prevLbSelectData = Object.assign({}, this.lbSelectData)
                    this.lbSelectData = Object.assign({}, {})

                    const lbServiceListLen = this.lbServiceList.length

                    const variableList = res.data.variable_dict || {}

                    const tmp = []
                    list.forEach(key => {
                        const obj = Object.assign({}, this.namespaceListTmp[key])
                        obj.variableList = variableList[key.split('_')[1]] || []

                        const lbSelectDataKey = obj.name
                        this.lbSelectData[lbSelectDataKey] = prevLbSelectData[lbSelectDataKey] || {}

                        for (let i = 0; i < lbServiceListLen; i++) {
                            const serviceName = this.lbServiceList[i].name
                            if (this.lbSelectData[lbSelectDataKey][serviceName] === null
                                || this.lbSelectData[lbSelectDataKey][serviceName] === undefined
                            ) {
                                this.lbSelectData[lbSelectDataKey][serviceName] = -1
                            }
                        }

                        tmp.push(obj)
                    })

                    this.selectedNamespaceList.splice(0, this.selectedNamespaceList.length, ...tmp)
                    this.dialogConf.isShow = false

                    this.invalidNsList.splice(0, this.invalidNsList.length, ...[])
                    this.lbServiceListInPage.splice(0, this.lbServiceListInPage.length, ...[])

                    // 点击选择命名空间弹层确认时，如果之前在 selectedNamespaceList 有选中，那么保持这个选中，同时强制发送
                    // 这个选中的 ns 的 previewNamespace 请求，如果没有选中，那么默认选中 selectedNamespaceList 第一个，
                    // 同时强制发送这个 ns 的 previewNamespace 请求
                    const alreadySelected = this.selectedNamespaceList.filter(item => item.isSelected)[0]
                    this.previewTitle = ''
                    this.previewNs = Object.assign({}, {})
                    this.previewList.splice(0, this.previewList.length, ...[])
                    this.editorConfig.editors.splice(0, this.editorConfig.editors.length, ...[])
                    this.editorConfig.values.splice(0, this.editorConfig.values.length, ...[])
                    this.previewShow = false

                    this.previewNamespace(alreadySelected || this.selectedNamespaceList[0], 0, true)
                } catch (e) {
                    console.error(e)
                } finally {
                    setTimeout(() => {
                        this.dialogConf.loading = false
                        this.dialogConf.isShow = false
                    }, 300)
                }
            },

            checkParams () {
                if (this.tplsetVerId === -1) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择模板集版本')
                    })
                    return false
                }

                if (!this.tplIndex.length) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择要实例化的模板')
                    })
                    return false
                }
                return true
            },

            resetNamespacePreview () {
                this.previewList.splice(0, this.previewList.length, ...[])
                this.previewNs = Object.assign({}, {})
                this.editorConfig.editors.splice(0, this.editorConfig.editors.length, ...[])
                this.editorConfig.values.splice(0, this.editorConfig.values.length, ...[])

                this.previewLoading = true
                this.previewShow = true

                this.previewErrorMessage = ''
            },

            previewNamespace (ns, index, forceSelect) {
                if (this.isYamlMode) {
                    this.previewNamespaceByYaml(ns, index)
                } else {
                    this.previewNamespaceByForm(ns, index, forceSelect)
                }
            },

            /**
             * 点击已选择的 namespace，查看yaml模式的
             *
             * @param {Object} ns 当前点击的 namespace 对象
             * @param {number} index 当前点击的 namespace 对象的索引
             */
            async previewNamespaceByYaml (ns, index) {
                if (!this.checkParams()) return false

                this.resetNamespacePreview(ns)
                this.clearNamespaceStatus()

                ns.isSelected = !ns.isSelected
                this.$set(this.selectedNamespaceList, index, ns)

                try {
                    const res = await this.$store.dispatch('k8sTemplate/createYamlTemplateReleases', {
                        projectId: this.projectId,
                        templateId: this.templateId,
                        versionId: this.tplsetVerIndex,
                        data: {
                            is_preview: true,
                            namespace_id: ns.id,
                            template_files: this.yamlTplResource
                        }
                    })
                    const list = []

                    // 标签，仅用于yaml模式
                    const tags = {
                        Deployment: 'deploy',
                        StatefulSet: 'sts',
                        DaemonSet: 'ds',
                        Ingress: 'ing',
                        Job: 'job',
                        Service: 'svc',
                        ConfigMap: 'cm',
                        Secret: 'secret',
                        HPA: 'hpa',
                        StorageClass: 'sc',
                        ServiceAccount: 'sa',
                        PodDisruptionBudget: 'pdb',
                        PersistentVolumeClaim: 'pvc',
                        PersistentVolume: 'pv',
                        ClusterRoleBinding: 'clusterrolebinding',
                        ClusterRole: 'clusterrole'
                    }
                    res.data.forEach(resource => {
                        const key = resource.resource_name
                        resource.files.forEach(file => {
                            list.push({
                                name: `${key}:${file.name}`, // 防止不同资源名称相同冲突
                                title: file.name,
                                tag: tags[key] || key,
                                content: file.content,
                                originalContent: file.content
                            })
                        })
                    })

                    this.previewTitle = `${ns.cluster_name} / ${ns.name} ${this.$t('的详细配置')}`
                    this.previewNs = Object.assign({}, ns)

                    this.sortTplType(list, 'tag', true)

                    this.previewList.splice(0, this.previewList.length, ...list)
                    this.previewList.forEach((preview, index) => {
                        this.editorConfig.values[index] = preview.content
                    })
                    setTimeout(() => {
                        // 这里触发一次 change 为了让初始值也显示
                        this.variableValChange()
                        this.editorConfig.editors[0] && this.editorConfig.editors[0].resize(true)
                        this.editorConfig.editors[0] && this.editorConfig.editors[0].gotoLine(0, 0, true)
                    }, 100)
                } catch (e) {
                    console.error('error: ' + e)
                    const errorMsg = e.message || e.data.msg || e.statusText
                    this.previewErrorMessage = errorMsg
                } finally {
                    this.previewLoading = false
                }
            },

            /**
             * 点击已选择的 namespace，查看form模式的
             *
             * @param {Object} ns 当前点击的 namespace 对象
             * @param {number} index 当前点击的 namespace 对象的索引
             */
            async previewNamespaceByForm (ns, index, forceSelect) {
                if (!this.checkParams()) return false

                this.resetNamespacePreview()
                this.clearNamespaceStatus()
                const variableInfo = {
                    [`${ns.id}`]: {}
                }
                ns.variableList.forEach(variable => {
                    variableInfo[`${ns.id}`][variable.key] = variable.value
                })

                ns.isSelected = !ns.isSelected
                this.$set(this.selectedNamespaceList, index, ns)

                try {
                    if (this.lbServiceList.length) {
                        const lbRes = await this.$store.dispatch('configuration/getLbInfo', {
                            projectId: this.projectId,
                            clusterId: ns.cluster_id
                        })

                        const lbData = lbRes.data || []

                        // 当前这个 ns 没有 lb 信息，不需要展示变量了，也不能 preview 和提交
                        if (!lbData.length) {
                            this.previewTitle = `${ns.cluster_name} / ${ns.name} ${this.$t('的详细配置')}`
                            this.previewNs = Object.assign({}, ns)
                            if (this.invalidNsList.indexOf(ns.name) < 0) {
                                this.invalidNsList.splice(0, this.invalidNsList.length, ...[].concat(ns.name))
                            }
                            return
                        }

                        const i = this.invalidNsList.indexOf(ns.name)
                        if (i !== -1) {
                            this.invalidNsList.splice(i, 1)
                        }

                        const lbServiceListInPage = []
                        const lbServiceListInPageTmp = {}

                        let segment = []
                        const lbServiceListLen = this.lbServiceList.length
                        for (let i = 0; i < lbServiceListLen; i++) {
                            const key = this.lbServiceList[i].name
                            if (!lbServiceListInPageTmp[key]) {
                                lbServiceListInPageTmp[key] = {}
                            }
                            const tmp = []
                            lbData.forEach((item, index) => {
                                tmp.push({
                                    service: key,
                                    ...item
                                })
                                lbServiceListInPageTmp[key][item.lb_id] = item.lb_name
                            })
                            segment.push({
                                key: key,
                                value: tmp
                            })

                            if (i % 2 !== 0) {
                                segment = []
                            } else {
                                lbServiceListInPage.push(segment)
                            }
                        }

                        this.lbServiceListInPage.splice(0, this.lbServiceListInPage.length, ...lbServiceListInPage)
                        this.lbServiceListInPageTmp = Object.assign({}, lbServiceListInPageTmp)

                        this.previewTitle = `${ns.cluster_name} / ${ns.name} ${this.$t('的详细配置')}`
                        this.previewNs = Object.assign({}, ns)

                        if (!this.checkCurNamespacePreview(ns)) {
                            this.previewList.splice(0, this.previewList.length, ...[])
                            this.editorConfig.editors.splice(0, this.editorConfig.editors.length, ...[])
                            this.editorConfig.values.splice(0, this.editorConfig.values.length, ...[])
                            return
                        }

                        this.preview4Lb(this.previewNs)
                    } else {
                        const curTplsetVer = this.tplsetVerList.filter(
                            tplsetVer => tplsetVer.id + '-' + tplsetVer.show_version_id === this.tplsetVerId
                        )[0] || {}

                        const res = await this.$store.dispatch('configuration/previewNamespace', {
                            projectId: this.projectId,
                            namespace: ns.id,
                            version_id: curTplsetVer.id,
                            show_version_id: curTplsetVer.show_version_id,
                            show_version_name: curTplsetVer.show_version_name,
                            instance_entity: this.instanceEntity,
                            variable_info: variableInfo
                        })

                        const list = []
                        const data = res.data || {}
                        Object.keys(data).forEach(key => {
                            data[key].forEach(item => {
                                const content = JSON.stringify(item.config, null, 4)
                                list.push({
                                    name: `${key}:${item.name}`, // 防止不同资源名称相同冲突
                                    title: item.name,
                                    tag: key,
                                    content: content,
                                    originalContent: content
                                })
                            })
                        })
                        this.previewTitle = `${ns.cluster_name} / ${ns.name} ${this.$t('的详细配置')}`
                        this.previewNs = Object.assign({}, ns)

                        this.sortTplType(list, 'tag', true)

                        this.previewList.splice(0, this.previewList.length, ...list)
                        this.previewList.forEach((preview, index) => {
                            this.editorConfig.values[index] = this.editorConfig.lang === 'yaml'
                                ? yamljs.dump(JSON.parse(preview.content, null, 4))
                                : preview.content
                        })
                        setTimeout(() => {
                            // 这里触发一次 change 为了让初始值也显示
                            this.variableValChange()
                            this.editorConfig.editors[0] && this.editorConfig.editors[0].resize(true)
                            this.editorConfig.editors[0] && this.editorConfig.editors[0].gotoLine(0, 0, true)
                        }, 100)
                    }
                } catch (e) {
                    console.error('error: ' + e)
                    const errorMsg = e.message || e.data.msg || e.statusText
                    this.previewErrorMessage = errorMsg
                } finally {
                    this.previewLoading = false
                }
            },

            /**
             * 解析 json
             */
            analysisJSON (data, originalData, path, keyMap) {
                if (toString.call(data) === '[object Array]') {
                    data.forEach((item, index) => {
                        const type = toString.call(item)
                        if (type === '[object Array]' || type === '[object Object]') {
                            this.analysisJSON(item, originalData, path.concat(index), keyMap)
                        } else {
                            const reg = new RegExp(`\\{\\{\\s*${Object.keys(keyMap).join('|')}\\s*\\}\\}`)
                            if (reg.test(item)) {
                                const copyPath = path.splice(0, path.length, ...path)
                                let s = Object.assign({}, originalData)
                                while (copyPath.length) {
                                    s = s[copyPath.shift()]
                                }
                                s[index] = keyMap[item.replace(/\{\{\s*/, '').replace(/\s*\}\}/, '')]
                            }
                        }
                    })
                } else if (toString.call(data) === '[object Object]') {
                    const keys = Object.keys(data)
                    keys.forEach(k => {
                        const type = toString.call(data[k])
                        if (type === '[object Array]' || type === '[object Object]') {
                            this.analysisJSON(data[k], originalData, path.concat(k), keyMap)
                        } else {
                            const intReg = /\{\{\s*([^{{]*)\s*\|toInt\}\}/g
                            if (intReg.test(data[k])) {
                                const ret = data[k].replace(intReg, (c, d) => {
                                    if (keyMap[d] !== undefined && keyMap[d] !== null) {
                                        if (String(keyMap[d]).trim() === '') {
                                            return c
                                        }
                                        return keyMap[d]
                                    }
                                })

                                if (!isNaN(ret)) {
                                    data[k] = parseFloat(ret)
                                } else {
                                    data[k] = ret
                                }
                            } else {
                                const reg = /\{\{\s*([^{{]*)\s*\}\}/g
                                if (reg.test(data[k])) {
                                    data[k] = data[k].replace(reg, (c, d) => {
                                        if (keyMap[d] !== undefined && keyMap[d] !== null) {
                                            if (String(keyMap[d]).trim() === '') {
                                                return c
                                            }
                                            return keyMap[d]
                                        }
                                    })
                                }
                            }
                        }
                    })
                }
            },

            /**
             * 变量修改的回调事件
             */
            variableValChange () {
                if (!this.previewNs.variableList || !this.previewNs.variableList.length) {
                    return
                }

                const scrollTopList = []
                const previewList = []
                previewList.splice(0, 0, ...this.previewList)

                const keys = {}
                this.previewNs.variableList.forEach(variable => {
                    keys[variable.key] = variable.value
                })

                const values = []
                previewList.forEach((preview, previewIndex) => {
                    const data = JSON.parse(preview.content)
                    const originalData = JSON.parse(preview.content)
                    this.analysisJSON(data, originalData, [], keys)

                    scrollTopList.push(
                        this.editorConfig.editors[previewIndex]
                            && this.editorConfig.editors[previewIndex].session.getScrollTop()
                    )

                    if (this.editorConfig.lang === 'yaml') {
                        const str = yamljs.dump(JSON.parse(JSON.stringify(data)))
                        values.push(str)
                    } else {
                        values.push(JSON.stringify(data, null, 4))
                    }
                })

                this.editorConfig.values.splice(0, this.editorConfig.values.length, ...values)
                this.previewList.splice(0, this.previewList.length, ...previewList)
                this.$nextTick(() => {
                    this.previewList.forEach((preview, previewIndex) => {
                        this.editorConfig.editors[previewIndex]
                            && this.editorConfig.editors[previewIndex].session.setScrollTop(scrollTopList[previewIndex])
                    })
                })
            },

            // variableValChange () {
            //     if (!this.previewNs.variableList || !this.previewNs.variableList.length) {
            //         return
            //     }
            //     const replacements = {}
            //     this.previewNs.variableList.forEach(variable => {
            //         replacements[`{{${variable.key}}}`] = variable.value
            //         replacements[`"{{${variable.key}}}}"`] =
            //             variable.value === null || variable.value === undefined || variable.value === ''
            //                 ? '""'
            //                 : isNaN(variable.value) ? `"{{${variable.key}}}}"` : variable.value
            //     })
            //     const regex = new RegExp(Object.keys(replacements).map(escape).join('|'), 'g')

            //     const scrollTopList = []
            //     const previewList = []
            //     previewList.splice(0, 0, ...this.previewList)

            //     const values = []
            //     previewList.forEach((preview, previewIndex) => {
            //         scrollTopList.push(this.editorConfig.editors[previewIndex].session.getScrollTop())
            //         preview.content = preview.originalContent.replace(regex, $0 => replacements[$0])

            //         values.push(this.editorConfig.lang === 'yaml'
            //             ? yamljs.dump(JSON.parse(preview.content, null, 4))
            //             : preview.content)
            //     })
            //     this.editorConfig.values.splice(0, this.editorConfig.values.length, ...values)
            //     this.previewList.splice(0, this.previewList.length, ...previewList)
            //     this.$nextTick(() => {
            //         this.previewList.forEach((preview, previewIndex) => {
            //             this.editorConfig.editors[previewIndex].session.setScrollTop(scrollTopList[previewIndex])
            //         })
            //     })
            // },

            /**
             * lbservice 下拉框改变事件
             *
             * @param {number} index 索引
             * @param {Object} data 索引对应的对象数据
             */
            async changeLbServiceSelect (index, data) {
                const lbSelectData = Object.assign({}, this.lbSelectData)
                lbSelectData[this.previewNs.name][data.service] = index
                this.lbSelectData = Object.assign({}, lbSelectData)

                if (!this.checkCurNamespacePreview(this.previewNs)) {
                    return
                }

                this.preview4Lb(this.previewNs)
            },

            /**
             * 判断当前的 ns 是否可以发送 preview
             * 只有当 当前的 ns 的下拉框全部选择时才可以发送
             *
             * @param {Object} ns 当前 ns
             */
            checkCurNamespacePreview (ns) {
                let canPreview = true
                const keys = Object.keys(this.lbSelectData[ns.name])
                for (let i = 0, len = keys.length; i < len; i++) {
                    if (this.lbSelectData[ns.name][keys[i]] === -1) {
                        canPreview = false
                        break
                    }
                }
                return canPreview
            },

            /**
             * 带有 lb 的 ns 的 preview
             *
             * @param {Object} ns 当前 ns
             */
            async preview4Lb (ns) {
                const curTplsetVer = this.tplsetVerList.filter(
                    tplsetVer => tplsetVer.id + '-' + tplsetVer.show_version_id === this.tplsetVerId
                )[0] || {}

                const variableInfo = {
                    [`${ns.id}`]: {}
                }
                ns.variableList.forEach(variable => {
                    variableInfo[`${ns.id}`][variable.key] = variable.value
                })

                const params = {
                    projectId: this.projectId,
                    namespace: ns.id,
                    version_id: curTplsetVer.id,
                    show_version_id: curTplsetVer.show_version_id,
                    show_version_name: curTplsetVer.show_version_name,
                    instance_entity: this.instanceEntity,
                    variable_info: variableInfo
                }

                const lbInfo = Object.assign({}, this.lbSelectData[ns.name])
                Object.keys(lbInfo).forEach(key => {
                    lbInfo[key] = this.lbServiceListInPageTmp[key][lbInfo[key]]
                })
                params.lb_info = {
                    [`${ns.id}`]: lbInfo
                }

                this.previewLoading = true

                try {
                    const res = await this.$store.dispatch('configuration/previewNamespace', params)

                    const list = []
                    const data = res.data || {}
                    Object.keys(data).forEach(key => {
                        data[key].forEach(item => {
                            const content = JSON.stringify(item.config, null, 4)
                            list.push({
                                tag: key,
                                name: item.name,
                                title: item.name,
                                content: content,
                                originalContent: content
                            })
                        })
                    })

                    this.sortTplType(list, 'tag', true)

                    this.previewList.splice(0, this.previewList.length, ...list)
                    this.previewList.forEach((preview, index) => {
                        this.editorConfig.values[index] = this.editorConfig.lang === 'yaml'
                            ? yamljs.dump(JSON.parse(preview.content, null, 4))
                            : preview.content
                    })

                    setTimeout(() => {
                        // 这里触发一次 change 为了让初始值也显示
                        this.variableValChange()
                        this.editorConfig.editors[0].resize(true)
                        this.editorConfig.editors[0].gotoLine(0, 0, true)
                    }, 100)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.previewLoading = false
                }
            },

            /**
             * 选项卡切换事件
             *
             * @param {string} name tab 名称
             * @param {index} index 索引，标识当前是第几个 tab 的 codeContent
             */
            tabChange (name, index) {
                // this.aceAnnotationErrorMsg = ''
                this.$nextTick(() => {
                    const curEditor = this.editorConfig.editors[index]
                    curEditor.resize(true)
                    const annotations = curEditor.getSession().$annotations
                    if (annotations && annotations.length) {
                        curEditor.gotoLine(annotations[0].row + 1, annotations[0].column, true)
                    }
                })
            },

            /**
             * ace editor 全屏
             *
             * @param {index} index 索引，标识当前是第几个 tab 的 ace editor
             */
            setFullScreen (index) {
                this.editorConfig.fullScreen = true
            },

            /**
             * 取消全屏
             */
            cancelFullScreen () {
                this.editorConfig.fullScreen = false
            },

            /**
             * 创建模板实例化
             */
            createInstance () {
                if (this.isYamlMode) {
                    this.createInstanceByYaml()
                } else {
                    this.createInstanceByForm()
                }
            },

            checkInstanceParams () {
                if (this.tplsetVerId === -1) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择模板集版本')
                    })
                    return
                }

                if (!this.tplIndex.length) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择要实例化的模板')
                    })
                    return
                }

                if (!this.selectedNamespaceList.length) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择命名空间')
                    })
                    return
                }

                return true
            },

            /**
             * 创建yaml模式模板实例化
             */
            async createInstanceByYaml () {
                if (!this.checkInstanceParams()) return false
                const me = this
                const ns = me.selectedNamespaceList[0]
                me.$bkInfo({
                    title: '',
                    content: me.$createElement('p', me.$t('请注意：命名空间{name}下的同名资源将会被覆盖更新，确认创建？', { name: ns.name })),
                    async confirmFn () {
                        me.createInstanceLoading = true
                        try {
                            await me.$store.dispatch('k8sTemplate/createYamlTemplateReleases', {
                                projectId: me.projectId,
                                templateId: me.templateId,
                                versionId: me.tplsetVerIndex,
                                data: {
                                    is_preview: false,
                                    namespace_id: ns.id,
                                    template_files: me.yamlTplResource
                                }
                            })

                            // 1、如果仅包含测试环境，跳转到应用列表页默认搜索测试环境集群
                            // 2、如果仅包含正式环境，跳转到应用列表页默认搜索正式环境集群
                            // 3、如果包含测试环境和正式环境，则按目前逻辑，默认搜索测试环境集群
                            // 只要有一个不是 prod 环境就应该是测试，否则全部都是 prod 就应该是正式
                            const hasNoProd = !!me.selectedNamespaceList.filter(
                                item => item.environment !== 'prod'
                            )[0]
                            me.gotoDeployments(hasNoProd)
                        } catch (e) {
                            console.log(e)
                        } finally {
                            me.createInstanceLoading = false
                        }
                    }
                })
            },

            /**
             * 创建form模式模板实例化
             */
            async createInstanceByForm () {
                if (!this.checkInstanceParams()) return false

                if (this.invalidNsList.length) {
                    this.$bkMessage({
                        theme: 'error',
                        message: `${this.$t('命令空间')} [${this.invalidNsList.join(',')}] ${this.$t('没有相关联的 LoadBalance，请先到 网络 -> LoadBalance 页面关联')}`
                    })
                    return
                }

                const curTplsetVer = this.tplsetVerList.filter(
                    tplsetVer => tplsetVer.id + '-' + tplsetVer.show_version_id === this.tplsetVerId
                )[0] || {}

                const variableInfo = {}

                let canCreate = true

                let params = {}
                const namespaces = []
                if (this.lbServiceList.length) {
                    if (this.checkLbServiceSelect()) {
                        params = {
                            projectId: this.projectId,
                            version_id: curTplsetVer.id,
                            show_version_id: curTplsetVer.show_version_id,
                            show_version_name: curTplsetVer.show_version_name,
                            instance_entity: this.instanceEntity,
                            lb_info: {},
                            is_start: true
                        }

                        this.selectedNamespaceList.forEach(item => {
                            namespaces.push(item.id)
                            variableInfo[`${item.id}`] = {}
                            item.variableList.forEach(variable => {
                                variableInfo[`${item.id}`][variable.key] = variable.value
                            })

                            const lbInfo = Object.assign({}, this.lbSelectData[item.name])
                            Object.keys(lbInfo).forEach(key => {
                                lbInfo[key] = this.lbServiceListInPageTmp[key][lbInfo[key]]
                            })
                            params.lb_info[item.id] = lbInfo
                        })

                        params.namespaces = namespaces.join(',')
                        params.variable_info = variableInfo
                    } else {
                        canCreate = false
                    }
                } else {
                    this.selectedNamespaceList.forEach(item => {
                        namespaces.push(item.id)
                        variableInfo[`${item.id}`] = {}
                        item.variableList.forEach(variable => {
                            variableInfo[`${item.id}`][variable.key] = variable.value
                        })
                    })
                    params = {
                        projectId: this.projectId,
                        namespaces: namespaces.join(','),
                        version_id: curTplsetVer.id,
                        show_version_id: curTplsetVer.show_version_id,
                        show_version_name: curTplsetVer.show_version_name,
                        instance_entity: this.instanceEntity,
                        variable_info: variableInfo,
                        is_start: true
                    }
                }

                if (!canCreate) {
                    return
                }

                const me = this
                me.$bkInfo({
                    title: me.$t('确认创建'),
                    content: me.$createElement('p', this.$t('确定要进行创建操作？')),
                    async confirmFn () {
                        me.createInstanceLoading = true
                        try {
                            await me.$store.dispatch('configuration/createInstance', params)

                            // 1、如果仅包含测试环境，跳转到应用列表页默认搜索测试环境集群
                            // 2、如果仅包含正式环境，跳转到应用列表页默认搜索正式环境集群
                            // 3、如果包含测试环境和正式环境，则按目前逻辑，默认搜索测试环境集群
                            // 只要有一个不是 prod 环境就应该是测试，否则全部都是 prod 就应该是正式
                            const hasNoProd = !!me.selectedNamespaceList.filter(
                                item => item.environment !== 'prod'
                            )[0]
                            me.gotoDeployments(hasNoProd)
                        } catch (e) {
                            console.log(e)
                        } finally {
                            me.createInstanceLoading = false
                        }
                    }
                })
            },

            gotoDeployments (hasNoProd) {
                if (this.isSharedCluster) {
                    const route = this.$router.resolve({ name: 'dashboardWorkload' })
                    window.location.href = route.href
                } else {
                    this.$router.push({
                        name: 'deployments',
                        params: {
                            isProdCluster: !hasNoProd,
                            projectId: this.projectId,
                            projectCode: this.projectCode,
                            tplsetId: this.templateId
                        }
                    })
                }
            },

            /**
             * 检测所有 ns 的 lbservice 下拉框是否都选择了，都选了以后才能 create
             */
            checkLbServiceSelect () {
                const keys = Object.keys(this.lbSelectData)
                const len = keys.length
                let outloop = true
                let ret = true
                for (let i = 0; i < len; i++) {
                    if (!outloop) {
                        break
                    }
                    const key = keys[i]
                    const serviceObj = this.lbSelectData[key]
                    const serviceObjKeyLen = Object.keys(serviceObj).length
                    for (let j = 0; j < serviceObjKeyLen; j++) {
                        const k = Object.keys(serviceObj)[j]
                        // console.error(key, serviceObj, k, serviceObj[k])
                        if (serviceObj[k] === -1) {
                            this.$bkMessage({
                                theme: 'error',
                                message: this.$t('请选择{key}命名空间的{k}')
                            })
                            outloop = false
                            ret = false
                            break
                        }
                    }
                }
                return ret
            },

            /**
             * 返回模板集列表
             *
             * @param {boolean} needConfirm 是否需要 confirm 提示
             */
            goTemplateset (needConfirm) {
                if (needConfirm) {
                    const me = this
                    const h = me.$createElement
                    me.$bkInfo({
                        title: this.$t('确认取消'),
                        content: h('p', this.$t('确定要取消实例化操作？')),
                        async confirmFn () {
                            me.$router.push({
                                name: 'templateset',
                                params: {
                                    projectId: this.projectId,
                                    projectCode: this.projectCode
                                }
                            })
                        }
                    })
                } else {
                    window.history.go(-1)
                }
            },

            /**
             * 去命名空间页面
             */
            goNamespace () {
                this.$router.push({
                    name: 'namespace',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode
                    }
                })
            },
            hideNamesapceDialog () {
                this.goNamespaceDialogConf.isShow = false
            },

            /**
             * 触发 快速添加命名空间
             *
             * @param {string} paramName paramDesc
             *
             * @return {string} returnDesc
             */
            triggerAddNamespace (index) {
                this.namespaceName = ''
                this.$refs.addNamespaceNode.forEach(vnode => {
                    vnode.instance.hide()
                })

                const vnode = this.$refs.addNamespaceNode[index]
                vnode.instance.show()
            },

            /**
             * 快速添加命名空间确认
             *
             * @param {Object} item 当前集群对象
             * @param {number} index 当前集群对象的索引
             */
            async addNamespace (item, index) {
                this.dialogConf.loading = true
                try {
                    const clusterId = item.cluster_id
                    if (!this.namespaceName.trim()) {
                        this.bkMessageInstance && this.bkMessageInstance.close()
                        this.bkMessageInstance = this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请填写命名空间名称')
                        })
                        return
                    }

                    if (this.namespaceName.length < 2) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('命名空间名称不得小于2个字符')
                        })
                        return
                    }

                    if (!/^[a-z][a-z0-9-]+$/g.test(this.namespaceName)) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('命名空间名称只能包含小写字母、数字以及连字符(-)，且不能以数字开头')
                        })
                        return
                    }

                    if (!clusterId) {
                        this.bkMessageInstance && this.bkMessageInstance.close()
                        this.bkMessageInstance = this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请选择所属集群')
                        })
                        return
                    }

                    const addedRes = await this.$store.dispatch('configuration/addNamespace', {
                        projectId: this.projectId,
                        name: this.namespaceName,
                        cluster_id: clusterId
                    })

                    const res = await this.$store.dispatch('configuration/getAllNamespaceList', {
                        projectId: this.projectId,
                        group_by: 'cluster_name',
                        perm_can_use: 1,
                        with_perms: false
                    })

                    const resList = res.data
                    const resCluster = resList.find(cluster => cluster.cluster_id === item.cluster_id)
                    if (resCluster) {
                        const resNamespaces = resCluster.results
                        const itemNamespaces = item.results
                        resNamespaces.forEach(ns => {
                            const inItemNamespaces = itemNamespaces.find(
                                itemNs => itemNs.id === ns.id && itemNs.name === ns.name
                            )
                            if (inItemNamespaces) {
                                ns.isChoose = inItemNamespaces.isChoose
                                ns.isExist = inItemNamespaces.isExist
                            }
                        })
                        this.$set(this.candidateNamespaceList, index, Object.assign(resCluster, {
                            isOpen: this.candidateNamespaceList[index].isOpen
                        }))

                        this.selectNamespaceInDialog(index, addedRes.data, 0)
                    }
                    this.cancelNamespace()
                } catch (e) {
                    catchErrorHandler(e, this)
                    this.$nextTick(() => {
                        this.$refs.addNamespaceInputNode[index] && this.$refs.addNamespaceInputNode[index].focus()
                    })
                } finally {
                    this.dialogConf.loading = false
                }
            },

            /**
             * 快速添加命名空间 tooltip 弹出回调函数
             */
            showAddNamespace (index) {
                this.$nextTick(() => {
                    this.$refs.addNamespaceInputNode[index] && this.$refs.addNamespaceInputNode[index].focus()
                })
            },

            /**
             * 快速添加命名空间取消
             */
            cancelNamespace () {
                this.$nextTick(() => {
                    this.namespaceName = ''
                    this.$refs.addNamespaceNode.forEach(vnode => {
                        vnode.instance.hide()
                    })
                })
            }
        }
    }
</script>

<style scoped>
    @import './instantiation.css';
</style>
