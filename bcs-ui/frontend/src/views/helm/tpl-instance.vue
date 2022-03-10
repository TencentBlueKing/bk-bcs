<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-helm-title">
                <a class="bcs-icon bcs-icon-arrows-left back" @click="goTplList"></a>
                <span>{{$t('Chart部署')}}</span>
            </div>
        </div>

        <div class="biz-content-wrapper" v-bkloading="{ isLoading: createInstanceLoading, zIndex: 100 }">
            <div>
                <div class="biz-helm-header">
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
                            <div class="logo-wrapper" v-if="curTpl.icon && isImage(curTpl.icon)" @click="gotoHelmTplDetail">
                                <img :src="curTpl.icon" style="width: 100px;">
                            </div>
                            <svg class="logo" v-else>
                                <use xlink:href="#biz-set-icon"></use>
                            </svg>

                            <div class="title">{{curTpl.name}}</div>
                            <p>
                                <a class="bk-text-button f12" href="javascript:void(0);" @click="gotoHelmTplDetail">{{$t('查看Chart详情')}}</a>
                            </p>
                            <div class="desc" :title="curTpl.description">
                                <span>{{$t('简介')}}：</span>
                                {{curTpl.description || '--'}}
                            </div>
                        </div>
                    </div>

                    <div class="right">
                        <div class="bk-collapse biz-collapse" style="border-top: none;">
                            <div class="bk-collapse-item bk-collapse-item-active">
                                <div class="biz-item-header" style="cursor: default; color: #737987;">
                                    {{$t('配置选项')}}
                                </div>
                                <div class="bk-collapse-item-content" style="padding: 15px;">
                                    <div class="config-box">
                                        <div class="inner">
                                            <div class="inner-item">
                                                <label class="title">
                                                    {{$t('名称')}}
                                                    <bcs-popover :content="$t('Release名称只能由小写字母数字或者-组成')" placement="top">
                                                        <span class="bk-badge">
                                                            <i class="bcs-icon bcs-icon-question-circle f14"></i>
                                                        </span>
                                                    </bcs-popover>
                                                </label>
                                                <bkbcs-input v-model="appName" :placeholder="$t('请输入Release名称')" />
                                            </div>

                                            <div class="inner-item">
                                                <label class="title">{{$t('版本')}}</label>
                                                <div>
                                                    <bk-selector
                                                        :placeholder="$t('请选择')"
                                                        style="width: 210px; display: inline-block; vertical-align: middle;"
                                                        searchable
                                                        :selected.sync="tplsetVerIndex"
                                                        :list="curTplVersions"
                                                        :setting-key="'version'"
                                                        :disabled="isTplSynLoading"
                                                        :display-key="'version'"
                                                        search-key="version"
                                                        @item-selected="getTplDetail">
                                                    </bk-selector>
                                                    <!-- <bkbcs-input
                                                        style="width: 210px; display: inline-block; vertical-align: middle;"
                                                        type="text"
                                                        :placeholder="$t('请选择')"
                                                        :value.sync="tplsetVerIndex"
                                                        :is-select-mode="true"
                                                        :default-list="curTplVersions"
                                                        :setting-key="'version'"
                                                        :disabled="isTplSynLoading"
                                                        :display-key="'version'"
                                                        :search-key="'version'"
                                                        @item-selected="getTplDetail">
                                                    </bkbcs-input> -->
                                                    <bk-button class="bk-button bk-default is-outline is-icon" v-bk-tooltips.top="$t('同步仓库')" @click="syncHelmTpl">
                                                        <div class="bk-spin-loading bk-spin-loading-mini bk-spin-loading-default" style="margin-top: -3px;" v-if="isTplSynLoading">
                                                            <div class="rotate rotate1"></div>
                                                            <div class="rotate rotate2"></div>
                                                            <div class="rotate rotate3"></div>
                                                            <div class="rotate rotate4"></div>
                                                            <div class="rotate rotate5"></div>
                                                            <div class="rotate rotate6"></div>
                                                            <div class="rotate rotate7"></div>
                                                            <div class="rotate rotate8"></div>
                                                        </div>
                                                        <i class="bcs-icon bcs-icon-refresh" v-else></i>
                                                    </bk-button>
                                                </div>
                                            </div>
                                        </div>
                                        <div class="inner">
                                            <div class="inner-item">
                                                <label class="title">{{$t('所属集群')}}</label>
                                                <bk-selector
                                                    style="width: 268px;"
                                                    :placeholder="$t('请选择')"
                                                    :searchable="true"
                                                    :selected.sync="curClusterId"
                                                    :field-type="'cluster'"
                                                    :list="clusterList"
                                                    :setting-key="'cluster_id'"
                                                    :display-key="'name'"
                                                    :disabled="!!globalClusterId">
                                                </bk-selector>
                                            </div>
                                            <div class="inner-item">
                                                <label class="title">{{$t('命名空间')}}</label>
                                                <div style="display: flex;align-items: center;">
                                                    <bcs-select style="width: 248px;"
                                                        searchable
                                                        :clearable="false"
                                                        v-model="namespaceId">
                                                        <bcs-option v-for="(item, index) in namespaceList"
                                                            :key="item.id"
                                                            :id="item.id"
                                                            :name="item.name"
                                                            v-authority="{
                                                                clickable: webAnnotations.perms[item.iam_ns_id]
                                                                    && webAnnotations.perms[item.iam_ns_id].namespace_scoped_use,
                                                                actionId: 'namespace_scoped_use',
                                                                resourceName: item.name,
                                                                disablePerms: true,
                                                                permCtx: {
                                                                    project_id: projectId,
                                                                    cluster_id: item.cluster_id,
                                                                    name: item.name
                                                                }
                                                            }"
                                                            @click.native="getClusterInfo(index, item)">
                                                        </bcs-option>
                                                        <div slot="extension" style="cursor: pointer;"
                                                            @click="goNamespaceList">
                                                            <i class="bcs-icon bcs-icon-apps"></i>
                                                            <span style="font-size: 14px">{{$t('命名空间列表')}}</span>
                                                        </div>
                                                    </bcs-select>
                                                    <!-- <bk-selector
                                                        style="width: 248px;"
                                                        :placeholder="$t('请选择')"
                                                        :searchable="true"
                                                        :selected.sync="namespaceId"
                                                        :field-type="'namespace'"
                                                        :list="namespaceList"
                                                        :setting-key="'id'"
                                                        :display-key="'name'"
                                                        @item-selected="getClusterInfo">
                                                    </bk-selector> -->
                                                    <i v-bk-tooltips.top="$t('如果Chart中已经配置命名空间，则会使用Chart中的命名空间，会导致不匹配等问题;建议Chart中不要配置命名空间')" class="bcs-icon bcs-icon-question-circle f14 ml5"></i>
                                                </div>
                                            </div>
                                            <p class="biz-tip pt10" id="cluster-info" style="clear: both;" v-if="clusterInfo" v-html="clusterInfo"></p>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                <template v-if="tplsetVerIndex">
                    <bk-tab type="border-card" active-name="chart" class="mt20">
                        <bk-tab-panel name="chart" :title="$t('Chart配置选项')">
                            <div slot="content" style="min-height: 180px;">
                                <section class="value-file-wrapper">
                                    {{$t('Values文件：')}}
                                    <bk-selector
                                        style="width: 300px;"
                                        :placeholder="$t('请选择')"
                                        :searchable="true"
                                        :selected.sync="curValueFile"
                                        :list="curValueFileList"
                                        :setting-key="'name'"
                                        :display-key="'name'"
                                        @item-selected="changeValueFile">
                                    </bk-selector>
                                    <bcs-popover placement="top">
                                        <span class="bk-badge ml5">
                                            <i class="bcs-icon bcs-icon-question-circle f14"></i>
                                        </span>
                                        <div slot="content">
                                            <p>{{ $t('Values文件包含两类:') }}</p>
                                            <p>{{ $t('- 以values.yaml结尾，例如xxx-values.yaml文件') }}</p>
                                            <p>{{ $t('- bcs-values目录下的文件') }}</p>
                                        </div>
                                    </bcs-popover>
                                </section>
                                <bk-tab
                                    :type="'fill'"
                                    :size="'small'"
                                    :active-name.sync="curEditMode"
                                    class="biz-tab-container"
                                    @tab-changed="helmModeChangeHandler">
                                    <bk-tab-panel name="yaml-mode" :title="$t('YAML模式')">
                                        <div style="width: 100%; min-height: 600px;">
                                            <p class="biz-tip p15" style="color: #63656E; overflow: hidden;">
                                                <i class="bcs-icon bcs-icon-info-circle biz-warning-text mr5"></i>
                                                {{$t('YAML初始值为创建时Chart中values.yaml内容，后续更新部署以该YAML内容为准，YAML内容最终通过`--values`选项传递给`helm template`命令')}}
                                            </p>
                                            <div v-bkloading="{ isLoading: isSyncYamlLoading, color: '#272822' }">
                                                <ace
                                                    v-if="curEditMode === 'yaml-mode'"
                                                    :value="curTplYaml"
                                                    :width="yamlConfig.width"
                                                    :height="yamlConfig.height"
                                                    :lang="yamlConfig.lang"
                                                    :read-only="yamlConfig.readOnly"
                                                    :full-screen="yamlConfig.fullScreen"
                                                    @init="editorInit">
                                                </ace>
                                            </div>
                                        </div>
                                    </bk-tab-panel>
                                    <bk-tab-panel name="form-mode" :title="$t('表单模式')">
                                        <p class="biz-tip p15" style="color: #63656E;">
                                            <i class="bcs-icon bcs-icon-info-circle biz-warning-text mr5"></i>{{$t('表单根据Chart中questions.yaml生成，表单修改后的数据会自动同步给YAML模式')}}
                                        </p>
                                        <template v-if="formData.questions">
                                            <bk-form-creater :form-data="formData" ref="bkFormCreater"></bk-form-creater>
                                        </template>
                                        <template v-else>
                                            <div class="biz-guard-box" v-if="!isQuestionsLoading">
                                                <span>{{$t('您可以参考')}}
                                                    <a class="bk-text-button" :href="PROJECT_CONFIG.doc.questionsYaml" target="_blank">{{$t('指引')}}</a>
                                                    {{$t('通过表单模式配置您的Helm Release 参数')}}，
                                                </span>
                                                <span>{{$t('也可以通过')}}<a href="javascript:void(0)" class="bk-text-button" @click="editYaml"></a>{{$t('直接修改Helm Release参数')}}</span>
                                            </div>
                                        </template>
                                    </bk-tab-panel>
                                </bk-tab>
                            </div>
                        </bk-tab-panel>
                        <bk-tab-panel name="helm" :title="$t('Helm部署选项')">
                            <div class="helm-set-panel">
                                <ul class="mt10">
                                    <!-- 常用枚举项 -->
                                    <li v-for="command of commandList" :key="command.id">
                                        <bk-checkbox
                                            class="mr5"
                                            v-model="helmCommandParams[command.id]" />
                                        <span class="mb5" style="display: inline-block;">
                                            {{command.desc}}
                                            <i style="font-size: 12px;cursor: pointer;"
                                                class="bcs-icon bcs-icon-info-circle"
                                                v-bk-tooltips.top="command.id" />
                                        </span>
                                    </li>
                                    <li class="mt10">
                                        <div style="margin-bottom:4px;">
                                            {{ $t('超时时间') }}
                                            <i style="font-size: 12px;cursor: pointer;"
                                                class="bcs-icon bcs-icon-info-circle"
                                                v-bk-tooltips.top="'--timeout'" />
                                        </div>
                                        <bk-input
                                            v-model="timeoutValue"
                                            placeholder="500"
                                            style="width: 200px; margin-right:4px;" />
                                        <span>{{ $t('秒') }}</span>
                                    </li>
                                    <!-- 高级选项 -->
                                    <button class="bk-text-button f12 mb10 pl0 mt10" @click.stop.prevent="toggleHign">
                                        {{$t('高级设置')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                                        <i style="font-size: 12px; cursor: pointer;"
                                            class="bcs-icon bcs-icon-info-circle ml5"
                                            v-bk-tooltips.top="hignDesc" />
                                    </button>
                                    <div v-show="isHignPanelShow">
                                        <div class="biz-key-value-wrapper mb10">
                                            <li class="biz-key-value-item mb10" v-for="(item, index) in hignSetupMap" :key="index">
                                                <bk-input style="width: 280px;" v-model="item.key" @change="handleHignkeyChange(item.key ,index)" />
                                                <span class="equals-sign">=</span>
                                                <bk-input style="width: 280px;" :placeholder="$t('值')" v-model="item.value" />
                                                <button class="action-btn" @click.stop.prevent>
                                                    <i class="bk-icon icon-plus-circle mr5" v-if="index === 0" @click.stop.prevent="addHign"></i>
                                                    <i class="bk-icon icon-minus-circle" @click.stop.prevent="delHign(index)"></i>
                                                </button>
                                                <p class="error-key" v-if="item.errorKeyTip">{{ item.errorKeyTip }}</p>
                                            </li>
                                        </div>
                                    </div>
                                </ul>
                            </div>
                        </bk-tab-panel>
                    </bk-tab>
                </template>

                <div class="create-wrapper">
                    <bk-button type="primary" :title="$t('部署')" @click="createApp">
                        {{$t('部署')}}
                    </bk-button>
                    <bk-button type="default" :title="$t('预览')" @click="showPreview">
                        {{$t('预览')}}
                    </bk-button>
                    <bk-button type="default" :title="$t('取消')" @click="goBack">
                        {{$t('取消')}}
                    </bk-button>
                </div>
            </div>
        </div>

        <bk-sideslider
            :is-show.sync="previewEditorConfig.isShow"
            :title="previewEditorConfig.title"
            :quick-close="true"
            :width="1000">
            <div slot="content" :style="{ height: `${winHeight - 70}px` }" v-bkloading="{ isLoading: previewInstanceLoading }">
                <template v-if="tplPreviewList.length">
                    <div class="biz-resource-wrapper" style="height: 100%; flex: 1;">
                        <resizer :class="['resize-layout fl']"
                            direction="right"
                            :handler-offset="3"
                            :min="250"
                            :max="400">
                            <div class="tree-box">
                                <bcs-tree
                                    ref="tree1"
                                    :data="treeData"
                                    :node-key="'id'"
                                    :has-border="true"
                                    @on-click="getFileDetail">
                                </bcs-tree>
                            </div>
                        </resizer>

                        <div class="resource-box">
                            <div class="biz-code-wrapper" style="height: 100%;">
                                <ace
                                    :value="curReourceFile.value"
                                    :width="editorConfig.width"
                                    :height="editorConfig.height"
                                    :lang="editorConfig.lang"
                                    :read-only="editorConfig.readOnly"
                                    :full-screen="editorConfig.fullScreen">
                                </ace>
                            </div>
                        </div>
                    </div>
                </template>
                <bcs-exception type="empty" scene="part" v-else></bcs-exception>
            </div>
        </bk-sideslider>

        <bk-dialog
            :is-show.sync="errorDialogConf.isShow"
            :width="750"
            :has-footet="false"
            :title="errorDialogConf.title"
            @cancel="hideErrorDialog">
            <template slot="content">
                <div class="bk-intro bk-danger pb30 mb15" v-if="errorDialogConf.message" style="position: relative;">
                    <pre class="biz-error-message">
                        {{errorDialogConf.message}}
                    </pre>
                    <bk-button size="small" type="default" id="error-copy-btn" :data-clipboard-text="errorDialogConf.message"><i class="bcs-icon bcs-icon-clipboard mr5"></i>{{$t('复制')}}</bk-button>
                </div>
                <div class="biz-message" v-if="errorDialogConf.errorCode === 40031">
                    <h3>{{$t('您需要')}}：</h3>
                    <p>1、{{$t('在集群页面，启用Helm')}}</p>
                    <p>2、{{$t('或者联系')}}【<a :href="PROJECT_CONFIG.doc.contact" class="bk-text-button">{{$t('蓝鲸容器助手')}}</a>】</p>
                </div>
                <div class="biz-message" v-else-if="errorDialogConf.actionType === 'previewApp'">
                    <h3>{{$t('您可以')}}：</h3>
                    <p>1、{{$t('检查Helm Chart是否存在语法错误')}}</p>
                    <p>2、{{$t('前往Helm Release列表页面，更新Helm Release')}}</p>
                </div>
                <div class="biz-message" v-else>
                    <h3>{{$t('您可以')}}：</h3>
                    <p>1、{{$t('更新Helm Chart，并推送到项目Chart仓库')}}</p>
                    <p>2、{{$t('前往Helm Release列表页面，更新Helm Release')}}</p>
                </div>
            </template>
            <div slot="footer">
                <div class="biz-footer">
                    <bk-button type="primary" @click="hideErrorDialog">{{$t('知道了')}}</bk-button>
                </div>
            </div>
        </bk-dialog>
    </div>
</template>

<script>
    import MarkdownIt from 'markdown-it'
    import yamljs from 'js-yaml'
    import path2tree from '@/common/path2tree'
    import baseMixin from '@/mixins/helm/mixin-base'
    import { catchErrorHandler } from '@/common/util'
    import Clipboard from 'clipboard'
    import resizer from '@/components/resize'

    export default {
        components: {
            resizer
        },
        mixins: [baseMixin],
        data () {
            return {
                clusterInfo: '',
                curEditMode: '',
                curTplReadme: '',
                yamlEditor: null,
                yamlFile: '',
                curTplYaml: '',
                curVersionData: {},
                activeName: ['config'],
                collapseName: ['var'],
                tplsetVerList: [],
                formData: {},
                createInstanceLoading: false,
                curValueFileList: [],
                tplPreviewList: [],
                difference: '',
                previewInstanceLoading: true,
                isQuestionsLoading: false,
                isSyncYamlLoading: true,
                isTplVerLoading: false,
                isRouterLeave: false,
                appName: '',
                winHeight: 0,
                curClusterId: '',
                editor: null,
                curTpl: {
                    data: {
                        name: ''
                    }
                },
                errorDialogConf: {
                    title: '',
                    isShow: false,
                    message: '',
                    errorCode: 0
                },
                curProjectId: '',
                previewEditorConfig: {
                    isShow: false,
                    title: this.$t('预览'),
                    width: '100%',
                    height: '100%',
                    lang: 'yaml',
                    readOnly: true,
                    fullScreen: false,
                    value: '',
                    editors: []
                },
                initedList: [],
                yamlConfig: {
                    isShow: false,
                    title: this.$t('预览'),
                    width: '100%',
                    height: '700',
                    lang: 'yaml',
                    readOnly: false,
                    fullScreen: false,
                    value: '',
                    editors: []
                },
                curReourceFile: {
                    name: '',
                    value: ''
                },
                editorConfig: {
                    width: '100%',
                    height: '100%',
                    lang: 'yaml',
                    readOnly: true,
                    fullScreen: false,
                    values: [],
                    editors: []
                },
                curTplVersions: [],
                tplsetVerIndex: '',
                namespaceId: '',
                answers: {},
                treeData: [],
                namespaceList: [],
                isTplSynLoading: false,
                curLabelList: [
                    {
                        key: '',
                        value: ''
                    }
                ],
                appAction: {
                    create: this.$t('部署'),
                    noop: '',
                    update: this.$t('更新'),
                    rollback: this.$t('回滚'),
                    delete: this.$t('删除'),
                    destroy: this.$t('删除')
                },
                curValueFile: 'values.yaml',
                isShowCommandParams: false,
                commandList: [
                    {
                        id: '--skip-crds',
                        disabled: false,
                        desc: this.$t('忽略CRD')
                    },
                    {
                        id: '--wait-for-jobs',
                        disabled: false,
                        desc: this.$t('等待所有Jobs完成')
                    },
                    {
                        id: '--wait',
                        disabled: false,
                        desc: this.$t('等待所有Pod，PVC处于ready状态')
                    }
                ],
                helmCommandParams: {
                    '--skip-crds': false,
                    '--wait-for-jobs': false,
                    '--wait': false,
                    '--timeout': false
                },
                timeoutValue: 600,
                isHignPanelShow: false,
                hignSetupMap: [
                    {
                        key: '',
                        value: ''
                    }
                ],
                hignDesc: this.$t('设置Flags，如设置wait，输入格式为 --wait = true'),
                webAnnotations: { perms: {} }
            }
        },
        computed: {
            curProject () {
                return this.$store.state.curProject
            },
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            tplList () {
                return this.$store.state.helm.tplList
            },
            globalClusterId () {
                return this.$store.state.curClusterId
            },
            clusterList () {
                return this.$store.state.cluster.clusterList
            }
        },
        watch: {
            globalClusterId: {
                handler (newVal) {
                    this.curClusterId = newVal
                    this.namespaceId = ''
                },
                immediate: true
            },
            curClusterId () {
                this.getNamespaceList(this.$route.params.tplId)
            }
        },
        async mounted () {
            const tplId = this.$route.params.tplId
            this.isRouterLeave = false
            this.curTpl = await this.getTplById(tplId)
            this.appName = ''
            this.getTplVersions()
            this.getNamespaceList(tplId)
            this.winHeight = window.innerHeight
        },
        beforeRouteLeave (to, from, next) {
            this.isRouterLeave = true
            next()
        },
        beforeDestroy () {
            this.isRouterLeave = true
        },
        methods: {
            /**
             * 返回chart 模版列表
             */
            goTplList () {
                const projectCode = this.$route.params.projectCode
                this.$router.push({
                    name: 'helmTplList',
                    params: {
                        projectCode: projectCode
                    }
                })
            },

            /**
             * 访问模板详情
             */
            gotoHelmTplDetail () {
                const route = this.$router.resolve({
                    name: 'helmTplDetail',
                    params: {
                        projectCode: this.projectCode,
                        tplId: this.$route.params.tplId
                    }
                })
                window.open(route.href, '_blank')
            },

            /**
             * 隐藏错误弹窗
             */
            hideErrorDialog () {
                this.errorDialogConf.isShow = false
            },

            /**
             * 获取集群信息
             * @param  {number} index 索引
             * @param  {object} data 集群
             */
            async getClusterInfo (index, data) {
                const clusterId = data.cluster_id
                const projectId = this.projectId

                this.clusterInfo = ''

                try {
                    const res = await this.$store.dispatch('helm/getClusterInfo', { projectId, clusterId })
                    const clusterInfo = res.data.note
                    const md = new MarkdownIt({
                        linkify: false
                    })

                    this.clusterInfo = md.render(clusterInfo)
                    this.$nextTick(() => {
                        // 处理链接情况
                        const markdownDom = document.getElementById('cluster-info')
                        markdownDom.querySelectorAll('a').forEach(item => {
                            item.target = '_blank'
                            item.className = 'bk-text-button'
                        })
                    })
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 获取文件详情
             * @param  {object} file 文件
             */
            getFileDetail (file) {
                if (file.hasOwnProperty('value')) {
                    this.curReourceFile = file
                }
            },

            /**
             * 从本地选择yaml文件
             */
            selectYaml (event) {
                const file = event.target.files[0]
                if (file) {
                    const fileReader = new FileReader()
                    fileReader.onload = (e) => {
                        this.curTplYaml = e.target.result
                        this.yamlConfig.isShow = true
                    }
                    fileReader.readAsText(file)
                }
            },

            /**
             * 编辑模式变化回调
             */
            helmModeChangeHandler (name) {
                if (name === 'yaml-mode') {
                    this.curTplYaml && this.editYaml()
                } else {
                    this.saveYaml()
                }
            },

            /**
             * 编辑yamml
             */
            async editYaml () {
                this.curEditMode = 'yaml-mode'
                let formData = []

                this.isSyncYamlLoading = true
                // 将数据配置的数据和yaml的数据进行合并同步
                if (this.$refs.bkFormCreater) {
                    formData = this.$refs.bkFormCreater.getFormData()
                }

                this.yamlConfig.isShow = true

                try {
                    const res = await this.$store.dispatch('helm/syncJsonToYaml', {
                        json: formData,
                        yaml: this.curTplYaml
                    })
                    this.curTplYaml = res.data.yaml
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    setTimeout(() => {
                        this.isSyncYamlLoading = false
                    }, 500)
                }
            },

            /**
             * 检查yaml
             */
            checkYaml () {
                const editor = this.yamlEditor
                const yaml = editor.getValue()

                if (!yaml) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入YAML')
                    })
                    return false
                }

                try {
                    yamljs.load(yaml)
                } catch (err) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入合法的YAML')
                    })
                    return false
                }

                // 显示错误提示
                const annot = editor.getSession().getAnnotations()
                if (annot && annot.length) {
                    editor.gotoLine(annot[0].row, annot[0].column, true)
                    return false
                }
                return true
            },

            /**
             * 保存yaml
             */
            saveYaml () {
                if (!this.checkYaml()) {
                    return false
                }
                const editor = this.yamlEditor
                const yaml = editor.getValue()
                let yamlData = {}

                try {
                    yamlData = yamljs.load(yaml)
                } catch (err) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入合法的YAML')
                    })
                    return false
                }

                // 同步表单到yaml数据配置
                if (yaml) {
                    yamlData = yamljs.load(yaml)
                }
                if (this.$refs.bkFormCreater) {
                    const formData = this.$refs.bkFormCreater.getFormData()
                    formData.forEach(formItem => {
                        const path = formItem.name
                        if (this.hasProperty(yamlData, path)) {
                            formItem.value = this.getProperty(yamlData, path)
                        }
                    })
                    this.setFormData(formData)
                }
                this.yamlFile = yaml
                this.yamlConfig.isShow = false
            },

            /**
             * 设置formCreater的值
             * @param {array} fieldset 字段数据
             */
            setFormData (fieldset) {
                const questions = JSON.parse(JSON.stringify(this.formData))
                if (questions.questions) {
                    questions.questions.forEach(question => {
                        if (fieldset && fieldset.length) {
                            fieldset.forEach(item => {
                                if (question.variable === item.name) {
                                    question.default = item.value
                                }
                            })
                        }

                        if (question.subquestions) {
                            question.subquestions.forEach(subQuestion => {
                                if (fieldset && fieldset.length) {
                                    fieldset.forEach(item => {
                                        if (subQuestion.variable === item.name) {
                                            subQuestion.default = item.value
                                        }
                                    })
                                }
                            })
                        }
                    })
                }
                this.formData = questions
            },

            /**
             * 隐藏yaml编辑
             */
            hideYaml () {
                this.yamlConfig.isShow = false
            },

            /**
             * 编辑器初始化成功回调
             * @param  {object} editor ace
             */
            editorInit (editor) {
                this.yamlEditor = editor
            },

            /**
             * 获取模板
             * @param  {number} id 模板ID
             * @return {object} result 模板
             */
            async getTplById (id) {
                let list = this.tplList

                // 如果没有缓存，获取远程数据
                if (!list.length) {
                    try {
                        const projectId = this.projectId
                        const res = await this.$store.dispatch('helm/asyncGetTplList', projectId)
                        list = res.data
                    } catch (e) {
                        catchErrorHandler(e, this)
                    }
                }

                const result = list.find(item => item.id === Number(id))
                return result || {}
            },

            /**
             * 根据版本号获取模板详情
             * @param  {number} index 索引
             * @param  {object} data 数据
             */
            async getTplDetail (index, data) {
                const list = []
                const projectId = this.projectId
                const version = index
                const versionId = this.curTplVersions.find(item => item.version === index).id
                const isPublic = this.curTpl.repository.name === 'public-repo'

                this.isQuestionsLoading = true
                try {
                    const fnPath = this.$INTERNAL ? 'helm/getChartVersionDetail' : 'helm/getChartByVersion'
                    const res = await this.$store.dispatch(fnPath, {
                        projectId,
                        chartId: this.$INTERNAL ? this.curTpl.name : this.curTpl.id,
                        version: this.$INTERNAL ? version : versionId,
                        isPublic
                    })
                    const tplData = res.data
                    const files = res.data.data.files
                    const tplName = tplData.name
                    const bcsTplName = tplData.name + '/bcs-values'
                    const regex = new RegExp(`^${tplName}\\/[\\w-]*values.(yaml|yml)$`)
                    const bcsRegex = new RegExp(`^${bcsTplName}\\/[\\w-]*.(yaml|yml)$`)

                    this.formData = res.data.data.questions
                    this.curVersionData = tplData

                    for (const key in files) {
                        if (bcsRegex.test(key)) {
                            const catalog = key.split('/')
                            const fileName = catalog[catalog.length - 2] + '/' + catalog[catalog.length - 1]
                            list.push({
                                name: fileName,
                                content: files[key]
                            })
                        }
                        if (regex.test(key)) {
                            const catalog = key.split('/')
                            const fileName = catalog[catalog.length - 1]
                            list.push({
                                name: fileName,
                                content: files[key]
                            })
                        }
                    }
                    this.curValueFileList.splice(0, this.curValueFileList.length, ...list)
                    this.curTplReadme = files[`${tplName}/README.md`]
                    this.curTplYaml = files[`${tplName}/values.yaml`]
                    this.yamlFile = files[`${tplName}/values.yaml`]
                    this.editYaml()
                    this.curTpl.description = res.data.data.description
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isQuestionsLoading = false
                }
            },

            /**
             * 修改value file
             */
            changeValueFile (index, data) {
                this.curValueFile = index
                this.curTplYaml = data.content
                this.yamlFile = data.content
                this.editYaml()
            },

            /**
             * 同步仓库
             */
            async syncHelmTpl () {
                if (this.isTplSynLoading) {
                    return false
                }

                this.isTplSynLoading = true
                try {
                    await this.$store.dispatch('helm/syncHelmTpl', { projectId: this.projectId })

                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('同步成功')
                    })

                    setTimeout(() => {
                        this.isTplVerLoading = true
                        this.getTplVersions()
                        this.tplsetVerIndex = ''
                    }, 1000)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    setTimeout(() => {
                        this.isTplSynLoading = false
                    }, 1000)
                }
            },

            /**
             * 获取模板版本列表
             */
            async getTplVersions () {
                const projectId = this.projectId
                try {
                    if (this.$INTERNAL) {
                        // 内部版本
                        const tplId = this.curTpl.name
                        const isPublic = this.curTpl.repository.name === 'public-repo'
                        const res = await this.$store.dispatch('helm/getTplVersionList', { projectId, tplId, isPublic })
                        this.curTplVersions = res.data
                    } else {
                        // 外部版本
                        const tplId = this.curTpl.id
                        const res = await this.$store.dispatch('helm/getTplVersions', {
                            projectId,
                            tplId
                        })
                        this.curTplVersions = res.data.results || []
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    setTimeout(() => {
                        this.isTplVerLoading = false
                    }, 1000)
                }
            },

            /**
             * 获取命名集群和空间列表
             */
            async getNamespaceList (chartId) {
                if (!this.curClusterId) return
                const projectId = this.projectId

                try {
                    const res = await this.$store.dispatch(
                        'helm/getNamespaceList',
                        {
                            projectId,
                            params: {
                                chart_id: chartId,
                                cluster_id: this.curClusterId
                            }
                        }
                    )
                    this.namespaceList = res.data
                    this.webAnnotations = res.web_annotations || { perms: {} }
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 检查数据
             * @param  {object} data 实例化数据
             * @return {boolean} true/false
             */
            checkFormData (data) {
                const nameReg = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/
                if (this.$refs.bkFormCreater) {
                    if (!this.$refs.bkFormCreater.checkValid()) {
                        return false
                    }
                }
                if (!data.name) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入Release名称')
                    })
                    return false
                }

                if (!nameReg.test(data.name)) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('Release名称只能由小写字母数字或者-组成')
                    })
                    return false
                }

                if (!data.chart_version) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择版本')
                    })
                    return false
                }

                if (!data.cluster_id) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择所属集群')
                    })
                    return false
                }

                if (!data.namespace_info) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择命名空间')
                    })
                    return false
                }

                return true
            },

            /**
             * 获取实例化参数
             * @return {object} data 实例化参数
             */
            getAppParams () {
                let formData = []
                const customs = []
                const commands = []
                for (const key in this.answers) {
                    customs.push({
                        name: key,
                        value: this.answers[key],
                        type: 'string'
                    })
                }

                for (const key in this.helmCommandParams) {
                    if (this.helmCommandParams[key]) {
                        const obj = {}
                        obj[key] = true
                        commands.push(obj)
                    }
                }
                if (this.timeoutValue) {
                    const obj = {}
                    obj['--timeout'] = this.timeoutValue + 's'
                    commands.push(obj)
                }

                // helm高级设置参数
                this.hignSetupMap.forEach(item => {
                    if (item.key.length) {
                        const obj = {}
                        let value
                        // 如果输入的值是（'false', 'False', 'True', 'true'）
                        // 则转成布尔值
                        if (['False', 'false'].includes(item.value)) {
                            value = false
                        } else if (['True', 'true'].includes(item.value)) {
                            value = true
                        } else {
                            value = item.value
                        }
                        obj[item.key] = value
                        commands.push(obj)
                    }
                })

                if (this.curEditMode === 'yaml-mode') {
                    this.saveYaml()
                }

                if (this.$refs.bkFormCreater) {
                    if (this.$refs.bkFormCreater.checkValid()) {
                        formData = this.$refs.bkFormCreater.getFormData()
                    }
                }
                const data = {
                    name: this.appName,
                    namespace_info: this.namespaceId,
                    cluster_id: this.curClusterId,
                    chart_version: this.curVersionData.id,
                    answers: formData,
                    customs: customs,
                    cmd_flags: commands,
                    valuefile_name: this.curValueFile,
                    valuefile: this.yamlFile
                }
                return data
            },

            /**
             * 显示错误弹层
             * @param  {object} res ajax数据对象
             * @param  {string} title 错误提示
             * @param  {string} actionType 操作
             */
            showErrorDialog (res, title, actionType) {
                // 先检查集群是否注册到 BKE server。未注册则返回 code: 40031
                this.errorDialogConf.errorCode = res.code
                this.createInstanceLoading = false
                this.errorDialogConf.message = res.message || res.data.msg || res.statusText
                this.errorDialogConf.isShow = true
                this.previewEditorConfig.isShow = false
                this.errorDialogConf.title = title
                this.errorDialogConf.actionType = actionType

                if (this.clipboardInstance && this.clipboardInstance.off) {
                    this.clipboardInstance.off('success')
                }
                if (this.errorDialogConf.message) {
                    this.$nextTick(() => {
                        this.clipboardInstance = new Clipboard('#error-copy-btn')
                        this.clipboardInstance.on('success', e => {
                            this.$bkMessage({
                                theme: 'success',
                                message: this.$t('复制成功')
                            })
                            this.isVarPanelShow = false
                        })
                    })
                }
            },

            /**
             * 展示App异常信息
             * @param  {object} app 应用对象
             */
            showAppError (app) {
                let actionType = ''
                const res = {
                    code: 500,
                    message: ''
                }

                res.message = app.transitioning_message
                actionType = app.transitioning_action
                const title = `${app.name}${this.appAction[app.transitioning_action]}${this.$t("失败")}`
                this.showErrorDialog(res, title, actionType)
            },

            /**
             * 查看app状态，包括创建、更新、回滚、删除
             * @param  {object} app 应用对象
             */
            async checkAppStatus (app) {
                const projectId = this.projectId
                const appId = app.id

                if (this.isRouterLeave) {
                    return false
                }

                try {
                    const res = await this.$store.dispatch('helm/checkAppStatus', { projectId, appId })
                    const action = this.appAction[res.data.transitioning_action]

                    if (res.data.transitioning_on) {
                        setTimeout(() => {
                            this.checkAppStatus(app)
                        }, 2000)
                    } else {
                        if (res.data.transitioning_result) {
                            this.$bkMessage({
                                theme: 'success',
                                message: `${app.name}${action}${this.$t('成功2')}`
                            })
                            // 返回helm首页
                            setTimeout(() => {
                                this.$router.push({
                                    name: 'helms'
                                })
                            }, 200)
                        } else {
                            this.createInstanceLoading = false
                            res.data.name = app.name || ''
                            this.showAppError(res.data)
                        }
                    }
                } catch (e) {
                    this.createInstanceLoading = false
                    this.showErrorDialog(e, this.$t('操作失败'), 'reback')
                }
            },

            /**
             * 创建应用
             */
            async createApp () {
                if (this.curEditMode === 'yaml-mode' && !this.checkYaml()) {
                    return false
                }
                const projectId = this.projectId
                const data = this.getAppParams()
                if (!this.checkFormData(data)) {
                    return false
                }

                this.errorDialogConf.isShow = false
                this.errorDialogConf.message = ''
                this.errorDialogConf.errorCode = 0
                this.createInstanceLoading = true

                try {
                    const res = await this.$store.dispatch('helm/createApp', { projectId, data })
                    this.checkAppStatus(res.data)
                } catch (e) {
                    this.showErrorDialog(e, this.$t('部署失败'), 'createApp')
                }
            },

            /**
             * 显示预览
             */
            async showPreview () {
                if (this.curEditMode === 'yaml-mode' && !this.checkYaml()) {
                    return false
                }
                const projectId = this.projectId
                const data = this.getAppParams()

                if (!this.checkFormData(data)) {
                    return false
                }

                this.previewEditorConfig.isShow = true
                this.previewInstanceLoading = true
                this.tplPreviewList = []
                this.difference = ''
                this.treeData = []

                try {
                    const res = await this.$store.dispatch('helm/previewCreateApp', { projectId, data })
                    this.previewEditorConfig.value = res.data.notes
                    for (const key in res.data.content) {
                        this.tplPreviewList.push({
                            name: key,
                            value: res.data.content[key]
                        })
                    }

                    const tree = path2tree(this.tplPreviewList)
                    this.treeData.push(tree)
                    this.difference = res.data.difference
                    if (this.tplPreviewList.length) {
                        this.curReourceFile = this.tplPreviewList[0]
                    }
                } catch (e) {
                    this.showErrorDialog(e, this.$t('预览失败'), 'previewApp')
                } finally {
                    this.previewInstanceLoading = false
                }
            },

            toggleHign () {
                this.isHignPanelShow = !this.isHignPanelShow
            },

            addHign () {
                const hignList = []
                hignList.splice(0, hignList.length, ...this.hignSetupMap)
                hignList.push({ key: '', value: '' })
                this.hignSetupMap.splice(0, this.hignSetupMap.length, ...hignList)
            },

            delHign (index) {
                if (!index) {
                    // 只剩一行时,置空数据
                    this.hignSetupMap[0].key = ''
                    this.hignSetupMap[0].value = ''
                    this.hignSetupMap[0].errorKeyTip = ''
                } else {
                    const hignList = []
                    hignList.splice(0, hignList.length, ...this.hignSetupMap)
                    hignList.splice(index, 1)
                    this.hignSetupMap.splice(0, this.hignSetupMap.length, ...hignList)
                }
            },

            handleHignkeyChange (val, index) {
                if (val.length > 1 && val.slice(0, 2) !== '--') {
                    const obj = {
                        errorKeyTip: this.$t('参数Key必须由 -- 字符开头')
                    }
                    Object.assign(this.hignSetupMap[index], obj)
                } else {
                    delete this.hignSetupMap[index].errorKeyTip
                }
            },
            goNamespaceList () {
                this.$router.push({
                    name: 'namespace',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode
                    }
                })
            }
        }
    }
</script>

<style scoped>
    @import './common.css';
    @import './tpl-instance.css';
</style>
