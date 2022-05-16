<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-helm-title">
                <a class="bcs-icon bcs-icon-arrows-left back" href="javascript:void(0);" @click="goTplList"></a>
                <span>{{ chartName }}</span>
            </div>
            <bcs-steps ext-cls="update-app-steps"
                :controllable="controllableSteps.controllable"
                :steps="controllableSteps.steps"
                :cur-step.sync="controllableSteps.curStep"
                @step-changed="stepChanged">
            </bcs-steps>
        </div>
        <div class="biz-content-wrapper" v-bkloading="{ isLoading: createInstanceLoading, zIndex: 100 }">
            <div class="step-content">
                <!-- 基本信息 -->
                <div class="basic-content" v-show="controllableSteps.curStep === 1">
                    <div class="content-item">
                        <svg style="display: none;">
                            <title>{{$t('模板集默认图标')}}</title>
                            <symbol id="biz-set-icon" viewBox="0 0 32 32">
                                <path d="M6 3v3h-3v23h23v-3h3v-23h-23zM24 24v3h-19v-19h19v16zM27 24h-1v-18h-18v-1h19v19z"></path>
                                <path d="M13.688 18.313h-6v6h6v-6z"></path>
                                <path d="M21.313 10.688h-6v13.625h6v-13.625z"></path>
                                <path d="M13.688 10.688h-6v6h6v-6z"></path>
                            </symbol>
                        </svg>
                        <div class="logo-wrapper" v-if="curTpl.icon && isImage(curTpl.icon)">
                            <img :src="curTpl.icon" style="width: 60px;">
                        </div>
                        <svg class="logo-wrapper" v-else>
                            <use xlink:href="#biz-set-icon"></use>
                        </svg>
                        
                        <span class="basic-wrapper">
                            <div class="appName">{{ curTpl.name }}</div>
                            <div class="desc">{{ $t('简介: ') }} <span class="">{{ curTpl.description || '--' }}</span></div>
                        </span>
                    </div>
                    <div class="basic-box">
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

                                <div style="display: flex;">
                                    <bk-selector
                                        :placeholder="$t('请选择')"
                                        style="display: inline-block; vertical-align: middle;"
                                        searchable
                                        :selected.sync="tplsetVerIndex"
                                        :list="curTplVersions"
                                        :disabled="isTplSynLoading"
                                        setting-key="version"
                                        display-key="version"
                                        search-key="version"
                                        @item-selected="getTplDetail">
                                    </bk-selector>

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
                                <label class="title">{{$t('所属集群')}}{{ globalClusterId }}</label>
                                <bk-selector
                                    :placeholder="$t('请选择')"
                                    :searchable="true"
                                    :selected.sync="curClusterId"
                                    field-type="cluster"
                                    :list="clusterList"
                                    setting-key="cluster_id"
                                    display-key="name"
                                    :disabled="!!globalClusterId">
                                </bk-selector>
                            </div>
                            <div class="inner-item">
                                <label class="title">{{$t('命名空间')}}</label>
                                <div style="display: flex;align-items: center;">
                                    <bcs-select style="width: 100%;"
                                        searchable
                                        :clearable="false"
                                        v-model="namespaceName">
                                        <bcs-option v-for="(item, index) in namespaceList"
                                            :key="item.id"
                                            :id="item.name"
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
                                    <i v-bk-tooltips.top="$t('如果Chart中已经配置命名空间，则会使用Chart中的命名空间，会导致不匹配等问题;建议Chart中不要配置命名空间')" class="bcs-icon bcs-icon-question-circle f14 ml5"></i>
                                </div>
                                <p class="biz-tip pt10" id="cluster-info" style="clear: both;" v-if="clusterInfo" v-html="clusterInfo"></p>
                            </div>
                        </div>

                        <div class="desc-inner">
                            <label class="title">{{$t('描述')}}</label>
                            <bcs-input type="textarea" :rows="3"></bcs-input>
                        </div>

                        <div class="readme-inner">
                            <label class="title">{{$t('自述')}}</label>
                            <div class="readme-content">
                                Rancher Alerting Drivers This chart enables ability to capture backups of the Rancher application and restore from these backups. This chart can be used to migrate Rancher from one Kubernetes cluster to a different Kubernetes cluster.For more information on how to use the feature,
                                refer to our docs. This chart installs one or more Alertmanager Webhook Receiver Integrations (i.e. Drivers).
                            </div>
                        </div>
                    </div>
                </div>

                <!-- 配置信息 -->
                <div class="config-content" v-show="controllableSteps.curStep === 2">
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
                                        setting-key="name"
                                        display-key="name"
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
                                    type="fill"
                                    size="small"
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
                                        <!-- <template v-if="formData.questions">
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
                                        </template> -->
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
                                    <button class="bk-text-button f12 mb10 pl0 mt10" @click.stop.prevent="isHignPanelShow = !isHignPanelShow">
                                        {{$t('高级设置')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                                        <i style="font-size: 12px; cursor: pointer;"
                                            class="bcs-icon bcs-icon-info-circle ml5"
                                            v-bk-tooltips.top="$t('设置Flags，如设置wait，输入格式为 --wait = true')" />
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
                </div>
            </div>
        </div>
        <div class="biz-footer-actions">
            <bcs-button theme="primary" v-if="controllableSteps.curStep === 2" :loading="createInstanceLoading" @click="controllableSteps.curStep = 1">{{ $t('上一步') }}</bcs-button>
            <bcs-button theme="primary" v-else :disabled="!tplsetVerIndex" :loading="createInstanceLoading" @click="controllableSteps.curStep = 2">{{ $t('下一步') }}</bcs-button>
            <bcs-button theme="primary" v-if="controllableSteps.curStep === 2" :loading="createInstanceLoading" @click="createApp">{{$t('部署')}}</bcs-button>
            <bcs-button @click="showPreview" :loading="createInstanceLoading">{{$t('预览')}} </bcs-button>
            <bcs-button @click="goBack" :loading="createInstanceLoading">{{$t('取消')}}</bcs-button>
        </div>
        
        <!-- 预览 -->
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
                                    node-key="id"
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
                <bcs-exception v-else type="empty" scene="part"></bcs-exception>
            </div>
        </bk-sideslider>

        <!-- 错误弹框 -->
        <bk-dialog
            :is-show.sync="errorDialogConf.isShow"
            :width="750"
            :has-footet="false"
            :title="errorDialogConf.title"
            @cancel="errorDialogConf.isShow = false">
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
                    <bk-button type="primary" @click="errorDialogConf.isShow = false">{{$t('知道了')}}</bk-button>
                </div>
            </div>
        </bk-dialog>
    </div>
</template>

<script>
    import { onMounted, ref, reactive, computed, watch, nextTick } from "@vue/composition-api"
    import baseMixin from '@/mixins/helm/mixin-base'
    import resizer from '@/components/resize'
    import MarkdownIt from 'markdown-it'
    import Clipboard from 'clipboard'
    import yamljs from 'js-yaml'
    import path2tree from '@/common/path2tree'

    export default {
        name: 'tplDeploy',
        components: {
            resizer
        },
        mixins: [baseMixin],
        setup (props, ctx) {
            const { $i18n, $route, $router, $store, $bkMessage } = ctx.root
            const appName = ref('')
            const clusterInfo = ref('')
            const namespaceName = ref('')
            const curClusterId = ref('')
            const tplsetVerIndex = ref('')
            const clipboardInstance = ref()
            const createInstanceLoading = ref(false)
            const isTplSynLoading = ref(false)
            const namespaceList = ref([])
            const winHeight = ref(0)
            const curTplVersions = ref([])
            const curTpl = ref({})
            const treeData = ref([])
            const tplPreviewList = ref([])
            const previewInstanceLoading = ref(false)
            const curValueFile = ref('values.yaml')
            const webAnnotations = ref({
                perms: {}
            })
            const curReourceFile = reactive({
                name: '',
                value: ''
            })
            const yamlEditor = ref(null)
            const isHignPanelShow = ref(false)
            const timeoutValue = ref(600)
            const hignSetupMap = ref([
                {
                    key: '',
                    value: ''
                }
            ])
            const editorConfig = ref({
                width: '100%',
                height: '100%',
                lang: 'yaml',
                readOnly: true,
                fullScreen: false,
                values: [],
                editors: []
            })
            const commandList = ref([
                {
                    id: '--skip-crds',
                    disabled: false,
                    desc: $i18n.t('忽略CRD')
                },
                {
                    id: '--wait-for-jobs',
                    disabled: false,
                    desc: $i18n.t('等待所有Jobs完成')
                },
                {
                    id: '--wait',
                    disabled: false,
                    desc: $i18n.t('等待所有Pod，PVC处于ready状态')
                }
            ])
            const helmCommandParams = ref({
                '--skip-crds': false,
                '--wait-for-jobs': false,
                '--wait': false,
                '--timeout': false
            })
            
            const curEditMode = ref('')
            const isSyncYamlLoading = ref(false)
            const bkFormCreater = ref(null)
            const isQuestionsLoading = ref(false)
            const formData = ref({})
            const curVersionData = ref({})
            const curValueFileList = ref([])
            const curTplYaml = ref('')
            const yamlFile = ref('')

            const yamlConfig = reactive({
                isShow: false,
                title: $i18n.t('预览'),
                width: '100%',
                height: '700',
                lang: 'yaml',
                readOnly: false,
                fullScreen: false,
                value: '',
                editors: []
            })
            const previewEditorConfig = reactive({
                isShow: false,
                title: $i18n.t('预览'),
                width: '100%',
                height: '100%',
                lang: 'yaml',
                readOnly: true,
                fullScreen: false,
                value: '',
                editors: []
            })

            const errorDialogConf = reactive({
                title: '',
                isShow: false,
                message: '',
                errorCode: 0
            })
            const controllableSteps = reactive({
                controllable: false,
                steps: [
                    { title: '基本信息', icon: 1 },
                    { title: '配置信息', icon: 2 }
                ],
                curStep: 1
            })

            const projectId = computed(() => {
                return $route.params.projectId
            })
            const projectCode = computed(() => {
                return $route.params.projectCode
            })
            const chartName = computed(() => {
                return $route.params.chartName
            })
            const tplList = computed(() => {
                return $store.state.helm.tplList
            })
            const globalClusterId = computed(() => {
                return $store.state.curClusterId
            })
            const username = computed(() => {
                return $store.state.user.username
            })
            const tplId = computed(() => {
                return $route.params.tplId
            })
            const clusterList = computed(() => {
                return $store.state.cluster.clusterList
            })

            watch(globalClusterId, (val) => {
                curClusterId.value = val
            }, {
                immediate: true
            })

            watch(curClusterId, () => {
                getNamespaceList(tplId.value)
            })

            /**
             * 获取命名集群和空间列表
             */
            const getNamespaceList = async (chartId) => {
                if (!curClusterId.value) return

                const res = await $store.dispatch(
                    'helm/getNamespaceList',
                    {
                        projectId: projectId.value,
                        params: {
                            chart_id: chartId,
                            cluster_id: curClusterId.value
                        }
                    }
                ).catch(() => false)

                if (!res) return

                namespaceList.value = res.data
                webAnnotations.value = res.web_annotations || { perms: {} }
            }

            /**
             * 点击修改步骤
             * @param {Boolean} val 当前步骤
             */
            const stepChanged = (val) => {
                controllableSteps.curStep = val
            }

            /**
             * 获取模板版本列表
             */
            const getTplVersions = async () => {
                const name = $route.params.chartName
                const res = await $store.dispatch('helm/getChartVersions', {
                    $projectId: projectCode.value,
                    $repository: projectCode.value,
                    $name: name,
                    page: 1,
                    size: 1500,
                    operator: username.value
                }).catch(() => false)

                if (!res) return
                curTplVersions.value = res.data
            }

            /**
             * 获取模板
             * @param  {number} id 模板ID
             * @return {object} result 模板
             */
            const getTplByName = async (name) => {
                let list = tplList.value

                // 如果没有缓存，获取远程数据
                if (!list.length) {
                    const res = await $store.dispatch('helm/getTplList', {
                        $projectId: projectCode.value,
                        $repository: projectCode.value,
                        page: 1,
                        size: 1500,
                        operator: username.value
                    })
                    list = res.data
                }
                const result = list.find(item => item.name === name)
                return result || {}
            }

            /**
             * 根据版本号获取模板详情
             * @param  {number} index 索引
             * @param  {object} data 数据
             */
            const getTplDetail = async (version, data) => {
                const list = []
                const name = curTplVersions.value.find(item => item.version === version).name
                isQuestionsLoading.value = true

                const res = await $store.dispatch('helm/getChartByVersion', {
                    $projectId: projectCode.value,
                    $repository: projectCode.value,
                    $name: name,
                    $version: version,
                    $operator: username.value
                })
                isQuestionsLoading.value = false

                curVersionData.value = res
                const tplName = res.name
                const files = res.contents
                const bcsTplName = res.name + '/bcs-values'
                const regex = new RegExp(`^${tplName}\\/[\\w-]*values.(yaml|yml)$`)
                const bcsRegex = new RegExp(`^${bcsTplName}\\/[\\w-]*.(yaml|yml)$`)
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
                curValueFileList.value.splice(0, curValueFileList.value.length, ...list)
                curTplYaml.value = files[`${tplName}/values.yaml`].content
                yamlFile.value = files[`${tplName}/values.yaml`].content
                editYaml()
            }

            /**
             * 编辑yamml
             */
            const editYaml = async () => {
                curEditMode.value = 'yaml-mode'
                let formData = []

                isSyncYamlLoading.value = true
                // 将数据配置的数据和yaml的数据进行合并同步
                if (bkFormCreater.value) {
                    formData = bkFormCreater.value.getFormData()
                }

                yamlConfig.isShow = true

                const res = await $store.dispatch('helm/syncJsonToYaml', {
                    json: formData,
                    yaml: curTplYaml.value
                }).catch(() => false)
                isSyncYamlLoading.value = false
                if (!res) return

                curTplYaml.value = res.data.yaml
            }

            /**
             * 获取集群信息
             * @param  {number} index 索引
             * @param  {object} data 集群
             */
            const getClusterInfo = async (index, data) => {
                const clusterId = data.cluster_id

                clusterInfo.value = ''

                const res = await $store.dispatch('helm/getClusterInfo', {
                    projectId: projectId.value,
                    clusterId
                }).catch(() => false)

                if (!res) return

                const note = res.data.note
                const md = new MarkdownIt({
                    linkify: false
                })

                clusterInfo.value = md.render(note)
                nextTick(() => {
                    // 处理链接情况
                    const markdownDom = document.getElementById('cluster-info')
                    markdownDom.querySelectorAll('a').forEach(item => {
                        item.target = '__blank'
                        item.className = 'bk-text-button'
                    })
                })
            }

            /**
             * 修改value file
             */
            const changeValueFile = (index, data) => {
                curValueFile.value = index
                curTplYaml.value = data.content.content
                yamlFile.value = data.content.content
                editYaml()
            }

            /**
             * 返回chart 模版列表
             */
            const goTplList = () => {
                $router.push({
                    name: 'helmTplList',
                    params: {
                        projectCode: projectCode.value
                    }
                })
            }

            const goNamespaceList = () => {
                $router.push({
                    name: 'namespace',
                    params: {
                        projectId: projectId.value,
                        projectCode: projectCode.value
                    }
                })
            }

            /**
             * 编辑器初始化成功回调
             * @param  {object} editor ace
             */
            const editorInit = (editor) => {
                yamlEditor.value = editor
            }

            const handleHignkeyChange = (val, index) => {
                if (val.length > 1 && val.slice(0, 2) !== '--') {
                    const obj = {
                        errorKeyTip: $i18n.t('参数Key必须由 -- 字符开头')
                    }
                    Object.assign(hignSetupMap.value[index], obj)
                } else {
                    delete hignSetupMap.value[index].errorKeyTip
                }
            }

            const addHign = () => {
                const hignList = []
                hignList.splice(0, hignList.length, ...hignSetupMap.value)
                hignList.push({ key: '', value: '' })
                hignSetupMap.value.splice(0, hignSetupMap.value.length, ...hignList)
            }

            const delHign = (index) => {
                if (!index) {
                    // 只剩一行时,置空数据
                    hignSetupMap.value[0].key = ''
                    hignSetupMap.value[0].value = ''
                    hignSetupMap.value[0].errorKeyTip = ''
                } else {
                    const hignList = []
                    hignList.splice(0, hignList.length, ...hignSetupMap.value)
                    hignList.splice(index, 1)
                    hignSetupMap.value.splice(0, hignSetupMap.value.length, ...hignList)
                }
            }

            /**
             * 创建应用
             */
            const createApp = async () => {
                if (curEditMode.value === 'yaml-mode' && !checkYaml()) {
                    return false
                }
                const data = getAppParams()
                if (!checkFormData(data)) {
                    return false
                }

                errorDialogConf.isShow = false
                errorDialogConf.message = ''
                errorDialogConf.errorCode = 0
                createInstanceLoading.value = true

                const res = await $store.dispatch('helm/createApp', {
                    $lusterId: curClusterId.value,
                    $namespace: namespaceName.value,
                    $name: appName.value,
                    ...data
                }).catch((e) => {
                    showErrorDialog(e, $i18n.t('部署失败'), 'createApp')
                })
                
                if (res && res.data.status === 'deployed') {
                    $bkMessage({
                        theme: 'success',
                        message: `${res.data.name}${$i18n.t('部署成功')}`
                    })
                    // 返回helm首页
                    setTimeout(() => {
                        $router.push({
                            name: 'helms'
                        })
                    }, 200)
                }
            }

            /**
             * 检查数据
             * @param  {object} data 实例化数据
             * @return {boolean} true/false
             */
            const checkFormData = (data) => {
                const nameReg = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/
                if (bkFormCreater.value) {
                    if (!bkFormCreater.value.checkValid()) {
                        return false
                    }
                }
                if (!data.name) {
                    $bkMessage({
                        theme: 'error',
                        message: $i18n.t('请输入Release名称')
                    })
                    return false
                }

                if (!nameReg.test(data.name)) {
                    $bkMessage({
                        theme: 'error',
                        message: $i18n.t('Release名称只能由小写字母数字或者-组成')
                    })
                    return false
                }

                if (!data.version) {
                    $bkMessage({
                        theme: 'error',
                        message: $i18n.t('请选择版本')
                    })
                    return false
                }

                if (!data.clusterID) {
                    $bkMessage({
                        theme: 'error',
                        message: $i18n.t('请选择所属集群')
                    })
                    return false
                }

                if (!data.namespace) {
                    $bkMessage({
                        theme: 'error',
                        message: $i18n.t('请选择命名空间')
                    })
                    return false
                }

                return true
            }

            /**
             * 获取实例化参数
             * @return {object} data 实例化参数
             */
            const getAppParams = () => {
                const commands = []

                for (const key in helmCommandParams.value) {
                    if (helmCommandParams.value[key]) {
                        commands.push(`${key}=true`)
                    }
                }
                if (timeoutValue.value) {
                    commands.push(`--timeout=${timeoutValue.value}s`)
                }

                // helm高级设置参数
                hignSetupMap.value.forEach(item => {
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

                if (curEditMode.value === 'yaml-mode') {
                    saveYaml()
                }

                // if (bkFormCreater.value) {
                //     if (bkFormCreater.value.checkValid()) {
                //         formData = bkFormCreater.value.getFormData()
                //     }
                // }
                const data = {
                    name: appName.value,
                    namespace: namespaceName.value,
                    clusterID: curClusterId.value,
                    projectID: projectCode.value,
                    repository: projectCode.value,
                    chart: chartName.value,
                    version: curVersionData.value.version,
                    operator: username.value,
                    args: commands,
                    values: [yamlFile.value]
                }
                return data
            }

            /**
             * 保存yaml
             */
            const saveYaml = () => {
                if (!checkYaml()) {
                    return false
                }
                const editor = yamlEditor.value
                const yaml = editor.getValue()
                let yamlData = {}

                try {
                    yamlData = yamljs.load(yaml)
                } catch (err) {
                    $bkMessage({
                        theme: 'error',
                        message: $i18n.t('请输入合法的YAML')
                    })
                    return false
                }

                // 同步表单到yaml数据配置
                if (yaml) {
                    yamlData = yamljs.load(yaml)
                }
                if (bkFormCreater.value) {
                    const formData = bkFormCreater.value.getFormData()
                    formData.forEach(formItem => {
                        const path = formItem.name
                        if (hasProperty(yamlData, path)) {
                            formItem.value = getProperty(yamlData, path)
                        }
                    })
                    setFormData(formData)
                }
                yamlFile.value = yaml
                yamlConfig.isShow = false
            }

            /**
             * 检查yaml
             */
            const checkYaml = () => {
                const editor = yamlEditor.value
                const yaml = editor.getValue()

                if (!yaml) {
                    $bkMessage({
                        theme: 'error',
                        message: $i18n.t('请输入YAML')
                    })
                    return false
                }

                try {
                    yamljs.load(yaml)
                } catch (err) {
                    $bkMessage({
                        theme: 'error',
                        message: $i18n.t('请输入合法的YAML')
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
            }

            /**
             * 根据path（eg: a.b.c）获取对象属性
             * @param {object} obj 对象
             * @param {string} path  路径
             * @return {string number...} value
             */
            const getProperty = (obj, path) => {
                const paths = path.split('.')
                let temp = obj
                if (paths.length) {
                    for (const item of paths) {
                        if (temp.hasOwnProperty(item)) {
                            temp = temp[item]
                        } else {
                            return undefined
                        }
                    }
                    return temp
                }
                return undefined
            }

            /**
             * 根据path（eg: a.b.c）判断对象属性是否存在
             * @param {object} obj 对象
             * @param {string} path  路径
             * @return {boolean} true/false
             */
            const hasProperty = (obj, path) => {
                const paths = path.split('.')
                let temp = obj
                const pathLength = paths.length
                if (pathLength) {
                    for (let i = 0; i < pathLength; i++) {
                        const item = paths[i]
                        if (temp.hasOwnProperty(item)) {
                            temp = temp[item]
                        } else {
                            return false
                        }
                    }
                    return true
                }
                return false
            }

            /**
             * 设置formCreater的值
             * @param {array} fieldset 字段数据
             */
            const setFormData = (fieldset) => {
                const questions = JSON.parse(JSON.stringify(formData.value))
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
                formData.value = questions
            }

            /**
             * 显示错误弹层
             * @param  {object} res ajax数据对象
             * @param  {string} title 错误提示
             * @param  {string} actionType 操作
             */
            const showErrorDialog = (res, title, actionType) => {
                // 先检查集群是否注册到 BKE server。未注册则返回 code: 40031
                errorDialogConf.errorCode = res.code
                createInstanceLoading.value = false
                errorDialogConf.message = res.response.data.message || ''
                errorDialogConf.isShow = true
                errorDialogConf.title = title
                errorDialogConf.actionType = actionType
                previewEditorConfig.isShow = false

                if (clipboardInstance.value && clipboardInstance.value.off) {
                    clipboardInstance.value.off('success')
                }
                if (errorDialogConf.message) {
                    nextTick(() => {
                        clipboardInstance.value = new Clipboard('#error-copy-btn')
                        clipboardInstance.value.on('success', e => {
                            $bkMessage({
                                theme: 'success',
                                message: $i18n.t('复制成功')
                            })
                        })
                    })
                }
            }

            /**
             * 显示预览
             */
            const showPreview = async () => {
                if (curEditMode.value === 'yaml-mode' && !checkYaml()) {
                    return false
                }
                const data = getAppParams()

                if (!checkFormData(data)) {
                    return false
                }

                previewEditorConfig.isShow = true
                previewInstanceLoading.value = true
                tplPreviewList.value = []
                treeData.value = []

                try {
                    const res = await $store.dispatch('helm/previewCreateApp', {
                        projectId: projectId.value,
                        data
                    })
                    previewEditorConfig.value = res.data.notes
                    for (const key in res.data.content) {
                        tplPreviewList.value.push({
                            name: key,
                            value: res.data.content[key]
                        })
                    }

                    const tree = path2tree(tplPreviewList.value)
                    treeData.value.push(tree)
                    if (tplPreviewList.value.length) {
                        curReourceFile.value = tplPreviewList.value[0]
                    }
                } catch (e) {
                    showErrorDialog(e, $i18n.t('预览失败'), 'previewApp')
                } finally {
                    previewInstanceLoading.value = false
                }
            }

            /**
             * 同步仓库
             */
            const syncHelmTpl = async () => {
                if (isTplSynLoading.value) {
                    return false
                }

                isTplSynLoading.value = true
                try {
                    await $store.dispatch('helm/syncHelmTpl', { projectId: projectId.value })

                    $bkMessage({
                        theme: 'success',
                        message: $i18n.t('同步成功')
                    })

                    setTimeout(() => {
                        getTplVersions()
                        tplsetVerIndex.value = ''
                    }, 1000)
                } catch (e) {
                    console.err(e)
                } finally {
                    setTimeout(() => {
                        isTplSynLoading.value = false
                    }, 1000)
                }
            }

            onMounted(async () => {
                curTpl.value = await getTplByName(chartName.value)
                getTplVersions()
                getNamespaceList(tplId.value)
                winHeight.value = window.innerHeight
            })

            return {
                createInstanceLoading,
                chartName,
                clusterInfo,
                curClusterId,
                namespaceName,
                projectId,
                webAnnotations,
                curTpl,
                appName,
                winHeight,
                tplsetVerIndex,
                curEditMode,
                curTplYaml,
                isSyncYamlLoading,
                isTplSynLoading,
                bkFormCreater,
                yamlConfig,
                isHignPanelShow,
                timeoutValue,
                hignSetupMap,
                commandList,
                editorConfig,
                curReourceFile,
                previewInstanceLoading,
                treeData,
                tplPreviewList,
                helmCommandParams,
                curValueFile,
                curValueFileList,
                curTplVersions,
                previewEditorConfig,
                errorDialogConf,
                controllableSteps,
                tplList,
                globalClusterId,
                clusterList,
                namespaceList,
                formData,
                stepChanged,
                getTplDetail,
                goTplList,
                getClusterInfo,
                goNamespaceList,
                changeValueFile,
                editorInit,
                addHign,
                delHign,
                handleHignkeyChange,
                createApp,
                showPreview,
                syncHelmTpl
            }
        }
    }
</script>

<style lang="postcss" scoped>
    @import '../common.css';
    .biz-helm-title {
        display: inline-block;
        height: 60px;
        line-height: 60px;
        font-size: 16px;
        margin-left: 20px;
    }
    .update-app-steps {
        width: 300px;
        position: absolute;
        right: 26px;
        top: 18px;
    }
    .step-content {
        background-color: #fff;
        padding: 24px 32px;
    }
    .biz-footer-actions {
        position: fixed;
        bottom: 0;
        width: 100%;
        height: 60px;
        background-color: #fff;
        padding: 14px 24px;
        z-index: 399;
    }
    .basic-content {
        .content-item {
            width: 800px;
            display: flex;
            .logo-wrapper {
                display: inline-block;
                width: 60px;
                height: 60px;
                margin-right: 24px;
                vertical-align: middle;
                fill: #ebf0f5;
            }
            .basic-wrapper {
                display: inline-block;
                .appName {
                    margin-top: 6px;
                    font-size: 16px;
                    color: #313238;
                }
                .desc {
                    margin-top: 10px;
                    color: #313238;
                    font-size: 12px;
                    span {
                        color: #979ba5;
                    }
                }
            }
        }
        .basic-box {
            margin-top: 24px;
            .inner {
                display: -webkit-box;
                width: 783px;
            }
            .inner-item {
                width: 360px;
                margin-right: 64px;
                margin-bottom: 24px;
            }
            .title {
                display: inline-block;
                margin-bottom: 6px;
                font-size: 12px;
            }
            .desc-inner {
                width: 783px;
                margin-bottom: 24px;
            }
            .readme-inner {
                .readme-content {
                    width: 783px;
                    margin-bottom: 20px;
                    background: #f0f1f5;
                    border: 1px solid #dcdee5;
                    border-radius: 2px;
                    font-size: 12px;
                    padding: 14px 8px;
                }
            }
        }
    }

    .config-content {
        .value-file-wrapper {
            padding: 15px 16px;
            border: 1px solid #dcdee5;
            border-radius: 2px;
            background: #f9fbfd;
            position: relative;
            bottom: -1px;
            font-size: 13px;
        }
        .action-btn {
            width: auto;
            padding: 0;
            height: 30px;
            text-align: center;
            display: inline-block;
            border: none;
            background: transparent;
            outline: none;
            margin-left: 5px;

            .bk-icon {
                width: 24px;
                height: 24px;
                line-height: 24px;
                border-radius: 50%;
                vertical-align: middle;
                color: #999999;
                font-size: 24px;
                display: inline-block;

                &:hover {
                    color: $primaryColor;
                    border-color: $primaryColor;
                }
            }
        }
    }
    .difference-code {
        height: 350px;
    }

    .editor-header {
        display: flex;
        display: flex;
        background: #eee;
        border-radius: 2px 2px 0 0;

        > div {
            padding: 5px;
            width: 50%;
        }
    }
    .biz-error-message {
        white-space: pre-line;
        text-align: left;
        max-height: 200px;
        overflow: auto;
        margin: 0;
    }
    .biz-message {
        margin-bottom: 0;

        h3 {
            text-align: left;
            font-size: 14px;
        }

        p {
            text-align: left;
            font-size: 13px;
        }
    }
    .biz-footer {
        text-align: right;
        padding-right: 15px;
    }
    #error-copy-btn {
        position: absolute;
        bottom: 5px;
        right: 5px;
    }
    .helm-set-panel {
        padding: 20px;
        li {
            font-size: 14px;
        }
    }
    .error-key {
        font-size: 12px;
        color: red
    }
</style>
