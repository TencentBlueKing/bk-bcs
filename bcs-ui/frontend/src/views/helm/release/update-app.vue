<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-helm-title">
                <a class="bcs-icon bcs-icon-arrows-left back" href="javascript:void(0);" @click="goToHelmIndex"></a>
                <span>{{ $t('升级') }}{{ appName }}</span>
            </div>
            <bcs-steps ext-cls="update-app-steps"
                :controllable="controllableSteps.controllable"
                :steps="controllableSteps.steps"
                :cur-step.sync="controllableSteps.curStep"
                @step-changed="stepChanged">
            </bcs-steps>
        </div>
        <div class="biz-content-wrapper" v-bkloading="{ isLoading: updateInstanceLoading, zIndex: 100 }">
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
                        <svg class="logo-wapper">
                            <use xlink:href="#biz-set-icon"></use>
                        </svg>
                        <span class="basic-wapper">
                            <div class="appName">{{ curApp.name }}</div>
                            <!-- <div class="desc">{{ $t('简介: ') }} <span class="">{{ curApp.chart_info.description || '--' }}</span></div> -->
                        </span>
                    </div>
                    <div class="basic-box">
                        <div class="inner">
                            <div class="inner-item">
                                <label class="title">{{$t('名称')}}</label>
                                <bkbcs-input :value="curApp.name" :disabled="true" />
                            </div>

                            <div class="inner-item">
                                <label class="title">{{$t('版本')}}</label>

                                <div style="display: flex;">
                                    <bk-selector
                                        :placeholder="$t('请选择')"
                                        searchable
                                        :selected.sync="tplVersionId"
                                        :list="curAppVersions"
                                        setting-key="version"
                                        :disabled="isTplSynLoading"
                                        display-key="version"
                                        search-key="version"
                                        @item-selected="handlerVersionChange">
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
                                <label class="title">{{$t('所属集群')}}</label>
                                <bkbcs-input :value="curClusterName" :disabled="true" />
                            </div>
                            <div class="inner-item">
                                <label class="title">
                                    {{$t('命名空间')}}
                                    <span class="ml10 biz-error-tip" v-if="!isNamespaceMatch">
                                        （{{$t('此命名空间不存在')}}）
                                    </span>
                                </label>
                                <div>
                                    <bkbcs-input :value="curApp.namespace" :disabled="true" />
                                </div>
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
                    <bk-tab type="border-card" active-name="chart">
                        <bk-tab-panel name="chart" :title="$t('Chart配置选项')">
                            <div slot="content" class="mt10" style="min-height: 180px;">
                                <section class="value-file-wrapper">
                                    {{$t('Values文件：')}}
                                    <bk-selector
                                        style="width: 200px;"
                                        :placeholder="$t('请选择')"
                                        :searchable="true"
                                        :selected.sync="curValueFile"
                                        :list="curValueFileList"
                                        :setting-key="'name'"
                                        :display-key="'name'"
                                        :disabled="isLocked"
                                        @item-selected="changeValueFile">
                                    </bk-selector>

                                    <bcs-popover placement="top">
                                        <span class="bk-badge" style="margin-left: 3px;">
                                            <i class="bcs-icon bcs-icon-question-circle f10"></i>
                                        </span>
                                        <div slot="content">
                                            <p>{{ $t('Values文件包含两类:') }}</p>
                                            <p>{{ $t('- 以values.yaml结尾，例如xxx-values.yaml文件') }}</p>
                                            <p>{{ $t('- bcs-values目录下的文件') }}</p>
                                        </div>
                                    </bcs-popover>

                                    <div style="display: inline-block;">
                                        <bk-checkbox class="ml10 mr5" v-model="isLocked">{{isLocked ? '已锁定' : '已解锁'}}</bk-checkbox>
                                    </div>
                                    <!-- <span class="biz-tip vm">({{$t('默认锁定values内容为当前release')}}({{$t('版本：')}}<span v-bk-tooltips.top="curApp.chart_info.version" class="release-version">{{curApp.chart_info.version}}</span>){{$t('的内容，解除锁定后，加载为对应Chart中的values内容')}})</span> -->
                                </section>
                                <bk-tab
                                    :type="'fill'"
                                    :size="'small'"
                                    :key="tabChangeIndex"
                                    :active-name.sync="curEditMode"
                                    class="biz-tab-container"
                                    @tab-changed="helmModeChangeHandler">
                                    <bk-tab-panel name="yaml-mode" :title="$t('YAML模式')">
                                        <template slot="tag">
                                            <span
                                                class="bcs-icon bcs-icon-circle-shape biz-danger-text v-bk"
                                                style="font-size: 10px;"
                                                v-bk-tooltips.top="$t('Release参数与选中的Chart Version中values.yaml有区别')"
                                                v-if="String(tplVersionId) !== originReleaseVersion && !isLocked">
                                            </span>
                                        </template>
                                        <div style="width: 100%; min-height: 600px; overflow: hidden;">
                                            <p class="biz-tip m15" style="color: #63656E;">
                                                <i class="bcs-icon bcs-icon-info-circle biz-warning-text mr5"></i>
                                                {{$t('YAML初始值为创建时Chart中values.yaml内容，后续更新部署以该YAML内容为准，内容最终通过`--values`选项传递给`helm template`命令')}}
                                            </p>
                                            <div v-if="String(tplVersionId) !== originReleaseVersion && !isLocked" class="f14 mb15 ml15" style="color: #63656E;">
                                                <i class="bcs-icon bcs-icon-eye biz-warning-text mr5"></i>
                                                {{$t('您更改了Chart版本，')}}<span class="bk-text-button" @click="showCodeDiffDialog">{{$t('点击查看')}}</span> Helm Release参数与选中的Chart Version中values.yaml区别
                                            </div>
                                            <div v-bkloading="{ isLoading: isSyncYamlLoading, color: '#272822' }">
                                                <ace
                                                    ref="codeViewer"
                                                    :value="curTplYaml"
                                                    :width="yamlConfig.width"
                                                    :height="yamlConfig.height"
                                                    :lang="yamlConfig.lang"
                                                    :read-only="yamlConfig.readOnly"
                                                    :full-screen="yamlConfig.fullScreen"
                                                    :key="curValueFile"
                                                    @init="editorInit">
                                                </ace>
                                                
                                            </div>
                                        </div>
                                    </bk-tab-panel>
                                    <bk-tab-panel name="form-mode" :title="$t('表单模式')">
                                        <p class="biz-tip p15" style="color: #63656E;">
                                            <i class="bcs-icon bcs-icon-info-circle biz-warning-text mr5"></i>{{$t('表单根据Chart中questions.yaml生成，表单修改后的数据会自动同步给YAML模式')}}
                                        </p>
                                        <template>
                                            <bk-form-creater :form-data="formData" ref="bkFormCreater"></bk-form-creater>
                                        </template>
                                        <template>
                                            <div class="biz-guard-box" v-if="!isQuestionsLoading">
                                                <span>{{$t('您可以参考')}}
                                                    <a class="bk-text-button" :href="PROJECT_CONFIG.doc.questionsYaml" target="_blank">{{$t('指引')}}</a>
                                                    {{$t('通过表单模式配置您的Helm Release 参数')}}，
                                                </span>
                                                <span>{{$t('也可以通过')}}<a href="javascript:void(0)" class="bk-text-button" @click="editYaml">{{$t('YAML模式')}}</a>{{$t('直接修改Helm Release参数')}}
                                                </span>
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
                                    <button class="bk-text-button f12 mb10 pl0 mt10" @click.stop.prevent="isHignPanelShow = !isHignPanelShow">
                                        {{$t('高级设置')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                                        <i style="font-size: 12px;cursor: pointer;"
                                            class="bcs-icon bcs-icon-info-circle"
                                            v-bk-tooltips.top="{ content: $t('设置Flags，如设置wait，输入格式为 --wait = true') }" />
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
            <bcs-button theme="primary" v-if="controllableSteps.curStep === 2" @click="controllableSteps.curStep = 1">{{ $t('上一步') }}</bcs-button>
            <bcs-button theme="primary" v-else @click="controllableSteps.curStep = 2" :disabled="updateInstanceLoading">{{ $t('下一步') }}</bcs-button>

            <template v-if="!isNamespaceLoading && !isNamespaceMatch && controllableSteps.curStep === 2">
                <bcs-popover :content="$t('所属命名空间不存在，不可操作')" placement="top">
                    <bcs-button theme="primary" :disabled="true">
                        {{$t('升级')}}
                    </bcs-button>
                </bcs-popover>
                <bcs-popover :content="$t('所属命名空间不存在，不可操作')" placement="top">
                    <bcs-button :disabled="true">
                        {{$t('预览')}}
                    </bcs-button>
                </bcs-popover>
            </template>
            <template v-else>
                <bcs-button
                    theme="primary"
                    v-if="controllableSteps.curStep === 2"
                    :loading="updateInstanceLoading"
                    :disabled="isNamespaceLoading && !isNamespaceMatch"
                    @click="confirmUpdateApp"
                >
                    {{ $t('升级') }}
                </bcs-button>
                <bcs-button
                    v-if="controllableSteps.curStep === 2"
                    :loading="updateInstanceLoading"
                    :disabled="isNamespaceLoading && !isNamespaceMatch"
                    @click="showPreview"
                >
                    {{ $t('预览') }}
                </bcs-button>
            </template>
            
            <bcs-button @click="goToHelmIndex" :loading="updateInstanceLoading">{{ $t('取消') }}</bcs-button>
        </div>
        
        <!-- 升级弹框 -->
        <bk-dialog
            :position="{ top: 80 }"
            :width="1100"
            :title="updateConfirmDialog.title"
            :close-icon="!updateInstanceLoading"
            :is-show.sync="updateConfirmDialog.isShow">
            <template slot="content">
                <p class="biz-tip mb5 tl" style="color: #666;">{{$t('Helm Release参数发生如下变化，请确认后再点击“确定”更新')}}</p>
                <div class="difference-code" v-bkloading="{ isLoading: isDifferenceLoading }" v-if="isDifferenceLoading || difference" style="height: 400px;">
                    <div class="editor-header">
                        <div>当前版本</div>
                        <div>更新版本</div>
                    </div>

                    <div :class="['diff-editor-box', { 'editor-fullscreen': yamlDiffEditorOptions.fullScreen }]" style="position: relative;">
                        <monaco-editor
                            ref="yamlEditor"
                            class="editor"
                            theme="monokai"
                            language="yaml"
                            :style="{ height: `${diffEditorHeight}px`, width: '100%' }"
                            v-model="curAppDifference.content"
                            :diff-editor="true"
                            :key="differenceKey"
                            :options="yamlDiffEditorOptions"
                            :original="curAppDifference.originContent">
                        </monaco-editor>
                    </div>
                </div>
                <div class="difference-code" v-bkloading="{ isLoading: isDifferenceLoading }" v-else style="height: 400px;">
                    <ace
                        :value="$t('本次更新没有内容变化')"
                        :width="updateConfirmDialog.width"
                        :height="updateConfirmDialog.height"
                        :lang="updateConfirmDialog.lang"
                        :read-only="updateConfirmDialog.readOnly"
                        :full-screen="updateConfirmDialog.fullScreen">
                    </ace>
                </div>
                <p class="biz-tip mt15 tl biz-warning" v-if="isChartVersionChange">{{$t('温馨提示：Helm Chart 版本已更改，请检查是否需要同步容器服务上Release 的参数')}}</p>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template>
                        <bk-button
                            theme="primary"
                            :loading="updateInstanceLoading"
                            :disabled="isDifferenceLoading"
                            @click="updateApp">
                            {{$t('确定')}}
                        </bk-button>
                        <bk-button
                            @click="hideConfirmDialog">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                </div>
            </div>
        </bk-dialog>

        <!-- 错误弹框 -->
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
                <div class="biz-message" v-else>
                    <h3>{{$t('您可以')}}：</h3>
                    <p>1、{{$t('更新Helm Chart，并推送到项目Chart仓库')}}</p>
                    <p>2、{{$t('重新更新')}}</p>
                </div>
            </template>
            <div slot="footer">
                <div class="biz-footer">
                    <bk-button type="primary" @click="errorDialogConf.isShow = false">{{$t('知道了')}}</bk-button>
                </div>
            </div>
        </bk-dialog>

        <!-- 预览 -->
        <bk-sideslider
            :is-show.sync="previewEditorConfig.isShow"
            :title="previewEditorConfig.title"
            :quick-close="true"
            :width="900">
            <div slot="content" :style="{ height: `${winHeight - 70}px` }" v-bkloading="{ isLoading: previewLoading }">
                <template v-if="appPreviewList.length">
                    <div class="biz-resource-wrapper" style="height: 100%;">
                        <resizer :class="['resize-layout fl']"
                            direction="right"
                            :handler-offset="3"
                            :min="250"
                            :max="400">
                            <div class="tree-box">
                                <bk-tree
                                    :data="treeData"
                                    :node-key="'name'"
                                    :has-border="true"
                                    @on-click="getFileDetail">
                                </bk-tree>
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
    </div>
</template>

<script lang="ts">
    import { ref, reactive, onMounted, computed, nextTick, watch } from '@vue/composition-api'
    import yamljs from 'js-yaml'
    import path2tree from '@/common/path2tree'
    import Clipboard from 'clipboard'
    import baseMixin from '@/mixins/helm/mixin-base'
    import resizer from '@/components/resize'
    import MonacoEditor from '@/components/monaco-editor/editor.vue'

    export default {
        name: 'updateApp',
        components: {
            resizer,
            MonacoEditor
        },
        mixins: [baseMixin],
        setup (props, ctx) {
            const { $router, $route, $store, $INTERNAL, $bkMessage, $i18n } = ctx.root
            const isQuestionsLoading = ref(false)
            const originReleaseData = ref<any>({})
            const tabChangeIndex = ref(-1)
            const curValueFile = ref('values.yaml')
            const curTplYaml = ref<any>('')
            const instanceYamlValue = ref('') // 当前应用实例化后的配置
            const instanceValueFileName = ref('') // 用户实例化选择的value文件名
            const curTplName = ref('')
            const curTplFiles = ref()
            const codeViewer = ref<any>(null)
            const timeoutValue = ref(600)
            const fieldset = ref([])
            const formData = ref({})
            const yamlFile = ref('')
            const curValueFileList = ref<any>([])
            const isLocked = ref(true)
            const tplVersionId = ref(-1)
            const originReleaseVersion = ref('')
            const winHeight = ref(0)
            const curEditMode = ref('yaml-mode')
            const isHignPanelShow = ref(true)
            const hignSetupMap = ref<any>([]) // helm部署配置高级设置
            const difference = ref('')
            const isDifferenceLoading = ref(false)
            const updateInstanceLoading = ref(false)
            const clipboardInstance = ref()
            const appPreviewList = ref<any>([])
            const previewLoading = ref(false)
            const treeData = ref<any>([])
            const curReourceFile = ref({
                name: '',
                value: ''
            })
            const errorDialogConf = reactive({
                title: '',
                isShow: false,
                message: '',
                errorCode: 0
            })
            const yamlDiffEditorOptions = reactive({
                readOnly: true,
                fontSize: 14,
                fullScreen: false
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

            const updateConfirmDialog = reactive({
                title: $i18n.t('确认升级'),
                isShow: false,
                width: '100%',
                height: '100%',
                lang: 'yaml',
                closeIcon: true,
                readOnly: true,
                fullScreen: false,
                values: [],
                editors: []
            })

            const controllableSteps = reactive({
                controllable: true,
                steps: [
                    { title: '基本信息', icon: 1 },
                    { title: '配置信息', icon: 2 }
                ],
                curStep: 1
            })

            const helmCommandParams = reactive({
                '--skip-crds': false,
                '--wait-for-jobs': false,
                '--wait': false,
                '--timeout': false
            })

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

            const curApp = ref<any>({
                cluster_id: '',
                created: '',
                chart_info: {
                    description: ''
                },
                namespace_id: '',
                release: {
                    id: '',
                    customs: [],
                    answers: {}
                }
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

            const editorConfig = reactive({
                width: '100%',
                height: '100%',
                lang: 'yaml',
                readOnly: true,
                fullScreen: false,
                values: [],
                editors: []
            })
            const appAction = reactive({
                create: $i18n.t('部署'),
                noop: '',
                update: $i18n.t('升级'),
                rollback: $i18n.t('回滚'),
                delete: $i18n.t('删除'),
                destroy: $i18n.t('删除')
            })
            const isChartVersionChange = ref(false)
            const differenceKey = ref(0)
            const curAppVersions = ref<any>([])
            const isNamespaceLoading = ref(false)
            const isNamespaceMatch = ref(true) // 判断命名空间是否已经删除
            const isTplSynLoading = ref(false)
            const curVersionId = ref(-1)
            const isSyncYamlLoading = ref(false)
            const bkFormCreater = ref<any>(null)
            const yamlEditor = ref<any>(null)
            const answers = ref({})
            const curAppDifference = reactive({
                content: '',
                originContent: ''
            })

            const username = computed(() => {
                return $store.state.user.username
            })
            const projectId = computed(() => {
                return $route.params.projectId
            })
            const projectCode = computed(() => {
                return $route.params.projectCode
            })

            const appName = computed(() => {
                return $route.params.appName
            })

            const clusterList = computed(() => {
                return $store.state.cluster.clusterList
            })

            const curClusterName = computed(() => {
                if (curApp.value.cluster_id !== undefined) {
                    const match = clusterList.value.find(item => item.cluster_id === curApp.value.cluster_id)
                    return match ? match.name : curApp.value.cluster_id
                }
                return ''
            })

            const diffEditorHeight = computed(() => {
                return yamlDiffEditorOptions.fullScreen ? window.innerHeight : 315
            })

            watch(isLocked, (val) => {
                initValuesFileData(curTplName.value, curTplFiles.value, '')
                tabChangeIndex.value++
            })

            /**
             * 返回Helm应用首页
             */
            const goToHelmIndex = () => {
                $router.push({
                    name: 'helms',
                    params: {
                        projectId: projectId.value,
                        projectCode: projectCode.value
                    }
                })
            }
            
            /**
             * 获取应用
             * @param  {number} appId 应用ID
             * @return {object} result 应用
             */
            // const getAppById = async (appId) => {
            //     let result = {}

            //     isQuestionsLoading.value = true
            //     const res = await $store.dispatch('helm/getAppById', {
            //         projectId: projectId.value,
            //         appId
            //     }).catch((e) => {
            //         if (e.status === 404) {
            //             goToHelmIndex()
            //         }
            //         return false
            //     })
            //     isQuestionsLoading.value = false
            //     if (!res) return

            //     result = res.data
            //     originReleaseData.value = result
            //     setAppDetail()
            //     isQuestionsLoading.value = false

            //     return result
            // }

            const setAppDetail = () => {
                const files = originReleaseData.value['release'].chartVersionSnapshot.files
                const tplName = originReleaseData.value['release'].chartVersionSnapshot.name
                const questions = originReleaseData.value['release'].chartVersionSnapshot.questions
                curValueFile.value = originReleaseData.value['valuefile_name'] || 'values.file'
                curTplYaml.value = originReleaseData.value['valuefile']
                instanceYamlValue.value = originReleaseData.value['valuefile'] // 保存当前应用实例化后的配置
                instanceValueFileName.value = originReleaseData.value['valuefile_name'] // 保存用户实例化时选择的文件名
                curTplName.value = tplName
                curTplFiles.value = files
                initValuesFileData(tplName, files, originReleaseData.value['valuefile_name'])

                nextTick(() => {
                    codeViewer.value && codeViewer.value.$ace && codeViewer.value.$ace.scrollToLine(1, true, true)
                })

                if (originReleaseData.value['cmd_flags'] && originReleaseData.value['cmd_flags'].length) {
                    // 所有常用枚举项keys
                    const commonKeys = Object.keys(helmCommandParams)
                    originReleaseData.value['cmd_flags'].forEach(item => {
                        const stringKey = Object.keys(item).join(',')
                        // 常用枚举项不包含则是用户自定义高级配置
                        if (!commonKeys.includes(stringKey)) {
                            const obj = {}
                            obj['key'] = stringKey
                            obj['value'] = item[stringKey]
                            hignSetupMap.value.push(obj)
                        } else {
                            if (stringKey === '--timeout') {
                                timeoutValue.value = item[stringKey].slice(0, item[stringKey].length - 1)
                            } else {
                                helmCommandParams[stringKey] = true
                            }
                        }
                    })
                }

                // 如果没有用户自定义helm配置, 默认添加一条空数据
                if (!hignSetupMap.value.length) {
                    const obj = {
                        key: '',
                        value: ''
                    }
                    hignSetupMap.value.push(obj)
                }

                if (questions.questions) {
                    questions.questions.forEach(question => {
                        fieldset.value = originReleaseData.value.release.answers
                        if (fieldset.value && fieldset.value.length) {
                            fieldset.value.forEach(item => {
                                if (question.variable === item['name']) {
                                    question.default = item['value']
                                }
                            })
                        }

                        if (question['subquestions']) {
                            question['subquestions'].forEach(subQuestion => {
                                if (fieldset.value && fieldset.value.length) {
                                    fieldset.value.forEach(item => {
                                        if (subQuestion.variable === item['name']) {
                                            subQuestion.default = item['value']
                                        }
                                    })
                                }
                            })
                        }
                    })
                }
                formData.value = questions
            }

            const initValuesFileData = (tplName, files, valueFileName) => {
                const list: Array<any> = []
                const regex = new RegExp(`^${tplName}\\/[\\w-]*values.(yaml|yml)$`)
                const bcsTplName = tplName + '/bcs-values'
                const bcsRegex = new RegExp(`^${bcsTplName}\\/[\\w-]*.(yaml|yml)$`)

                yamlFile.value = ''

                // 根据valueFileName判断是否第一次展示, curTplYaml显示实例化配置的内容
                if (valueFileName) {
                    curValueFile.value = valueFileName
                }
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
                curValueFileList.value = list

                // 选择版本后（valueFileName为空）
                if (!valueFileName) {
                    // 如果锁定则用release values
                    if (isLocked.value) {
                        curTplYaml.value = instanceYamlValue.value
                        curValueFile.value = instanceValueFileName.value
                        console.log('locked usedefault')
                    } else {
                        // 选择版本后依然是原来的, 显示原来实例化时的配置内容
                        if (String(tplVersionId.value) === originReleaseVersion.value) {
                            curTplYaml.value = instanceYamlValue.value
                            curValueFile.value = instanceValueFileName.value
                            console.log('unlocked but no change usedefault')
                        } else {
                            const fileNames = Object.keys(files)
                            const matchName = fileNames.find(name => {
                                return name.endsWith(instanceValueFileName.value)
                            })
                            // 默认用原来的value文件
                            if (matchName) {
                                curTplYaml.value = files[matchName]
                                curValueFile.value = instanceValueFileName.value
                            } else {
                                const valueDefaultNames = `${tplName}/values.yaml`
                                curTplYaml.value = files[valueDefaultNames]
                                curValueFile.value = 'values.yaml'
                            }
                            console.log('unlocked and change usenew')
                        }
                    }
                }
            }
                        
            /**
             * 获取应用版本列表
             * @param  {number} appId 应用ID
             */
            const getAppVersions = async () => {
                curAppVersions.value = []

                const name = $route.params.name
                const res = await $store.dispatch('helm/getChartVersions', {
                    projectId: projectCode.value,
                    repository: projectCode.value,
                    name,
                    params: {
                        page: 1,
                        size: 1500,
                        operator: username.value
                    }
                }).catch(() => false)

                if (!res) return
                curAppVersions.value = res.data

                if (curAppVersions.value.length) {
                    originReleaseVersion.value = curAppVersions.value[0].version
                    tplVersionId.value = curAppVersions.value[0].version
                }
            }

            /**
             * 获取命名空间列表
             */
            const getNamespaceList = async () => {
                isNamespaceLoading.value = true

                const res = await $store.dispatch('helm/getNamespaceList', {
                    projectId: projectId.value,
                    params: {
                        cluster_id: $route.params.clusterId
                    }
                }).catch(() => false)
                isNamespaceLoading.value = false

                if (!res) return
                const curNamespace = curApp.value.namespace
                isNamespaceMatch.value = (res.data || []).some(item => item.name === curNamespace)
            }

            /**
             * 同步仓库
             */
            const syncHelmTpl = async () => {
                if (isTplSynLoading.value) {
                    return false
                }
                isTplSynLoading.value = true
                const res = await $store.dispatch('helm/syncHelmTpl', { projectId: projectId.value }).catch(() => false)
                setTimeout(() => {
                    isTplSynLoading.value = false
                }, 1000)

                if (!res) return

                $bkMessage({
                    theme: 'success',
                    message: $i18n.t('同步成功')
                })

                setTimeout(() => {
                    getAppVersions()
                }, 1000)
            }

            /**
             * 切换应用版本号回调
             * @param  {number} index 索引
             * @param  {object} data 版本对象
             */
            const handlerVersionChange = async (index, data) => {
                tabChangeIndex.value++
                if (data.version === originReleaseVersion.value) {
                    setAppDetail()
                    curVersionId.value = -1
                    return false
                }
                const appName = curApp.value.name
                const appId = curApp.value.id
                const clusterId = curApp.value.cluster_id
                const namespace = curApp.value.namespace
                const version = data.version
                const versionId = data.id

                const fnPath = $INTERNAL ? 'helm/getUpdateChartVersionDetail' : 'helm/getUpdateChartByVersion'
                const res = await $store.dispatch(fnPath, {
                    projectId: projectId.value,
                    appId: $INTERNAL ? appName : appId,
                    version: $INTERNAL ? version : versionId,
                    clusterId: $INTERNAL ? clusterId : undefined,
                    namespace: $INTERNAL ? namespace : undefined
                }).catch(() => false)

                if (!res) return

                const files = res.data.data.files
                const tplName = res.data.name
                formData.value = res.data.data.questions

                curTplName.value = tplName
                curTplFiles.value = files
                curVersionId.value = res.data.id
                initValuesFileData(tplName, files, '')
            }

            /**
             * 修改value file
             */
            const changeValueFile = (index, data) => {
                curValueFile.value = index
                // curVersionYaml = data.content
                curTplYaml.value = data.content
                yamlFile.value = ''
                // 没有选择过版本时, 如果切换为原来实例化的文件名，显示原来实例化时的配置内容
                if (String(tplVersionId.value) === originReleaseVersion.value && index === instanceValueFileName) {
                    curTplYaml.value = instanceYamlValue
                    console.log('unlocked && nochange use default')
                } else {
                    console.log('unlocked && valsuechange')
                }
            }

            /**
             * 编辑模式变化回调
             */
            const helmModeChangeHandler = (name) => {
                if (name === 'yaml-mode') {
                    editYaml()
                } else {
                    saveYaml()
                }
            }

            /**
             * 编辑yamml
             */
            const editYaml = async () => {
                let formData = []

                curEditMode.value = 'yaml-mode'
                isSyncYamlLoading.value = true
                // 将数据配置的数据和yaml的数据进行合并同步
                if (bkFormCreater.value) {
                    formData = bkFormCreater.value.getFormData()
                }
                if (curTplYaml.value) {
                    yamljs.load(curTplYaml.value)
                }

                yamlConfig.isShow = true
                const res = await $store.dispatch('helm/syncJsonToYaml', {
                    json: formData,
                    yaml: curTplYaml.value
                }).catch(() => false)

                isSyncYamlLoading.value = false

                if (!res) return

                curTplYaml.value = res.data.yaml
                nextTick(() => {
                    yamlEditor.value.gotoLine(0, 0)
                })
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
                    // 通过load检测yaml是否合法
                    yamljs.load(yaml)
                } catch (err) {
                    $bkMessage({
                        theme: 'error',
                        message: $i18n.t('请输入合法的YAML')
                    })
                    return false
                }

                const annot = editor.getSession().getAnnotations()
                if (annot && annot.length) {
                    editor.gotoLine(annot[0].row, annot[0].column, true)
                    return false
                }
                return true
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
                let formData = []
                let yamlData = {}

                try {
                    // 通过load检测yaml是否合法
                    yamljs.load(yaml)
                } catch (err) {
                    $bkMessage({
                        theme: 'error',
                        message: $i18n.t('请输入合法的YAML')
                    })
                    return false
                }

                // 同步到数据配置
                if (yaml) {
                    yamlData = yamljs.load(yaml)
                }
                if (bkFormCreater.value) {
                    formData = bkFormCreater.value.getFormData()
                    formData.forEach(formItem => {
                        const path = formItem['name']
                        if (hasProperty(yamlData, path)) {
                            formItem['value'] = getProperty(yamlData, path) as never
                        }
                    })
                    setFormData(formData)
                }

                yamlFile.value = yaml
                yamlConfig.isShow = false
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
             * 编辑器初始化成功回调
             * @param  {object} editor ace
             */
            const editorInit = (editor) => {
                yamlEditor.value = editor
            }

            const addHign = () => {
                const hignList: Array<any> = []
                hignList.splice(0, hignList.length, ...hignSetupMap.value)
                hignList.push({ key: '', value: '' })
                console.log(hignSetupMap.value, '')
                hignSetupMap.value.splice(0, hignSetupMap.value.length, ...hignList)
            }

            const delHign = (index) => {
                if (!index) {
                    // 只剩一行时,置空数据
                    hignSetupMap.value[0].key = ''
                    hignSetupMap.value[0].value = ''
                    hignSetupMap.value[0].errorKeyTip = ''
                } else {
                    const hignList: Array<any> = []
                    hignList.splice(0, hignList.length, ...hignSetupMap.value)
                    hignList.splice(index, 1)
                    hignSetupMap.value.splice(0, hignSetupMap.value.length, ...hignList)
                }
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
            
            /**
             * 显示确认更新弹窗
             */
            const confirmUpdateApp = () => {
                if (curEditMode.value === 'yaml-mode' && !checkYaml()) {
                    return false
                }
                if (bkFormCreater.value && !(bkFormCreater.value.checkValid())) {
                    return false
                }
                isDifferenceLoading.value = true
                updateConfirmDialog.isShow = true
                getDifference()
            }

            /**
             * 获取版本对比
             */
            const getDifference = async () => {
                const params = getAppParams()
                const appId = $route.params.appId
                isDifferenceLoading.value = true
                difference.value = ''
                isChartVersionChange.value = false

                try {
                    const res = await $store.dispatch('helm/previewApp', {
                        projectId: projectId.value,
                        appId,
                        params
                    })
                    difference.value = res.data.difference
                    curAppDifference.content = res.data.new_content
                    curAppDifference.originContent = res.data.old_content
                    differenceKey.value++
                    isChartVersionChange.value = res.data.chart_version_changed
                } catch (e) {
                    updateConfirmDialog.isShow = false
                    showErrorDialog(e.response, $i18n.t('Chart渲染失败'), 'preUpdate')
                } finally {
                    isDifferenceLoading.value = false
                }
            }

            /**
             * 获取应用参数
             * @return {object} params 应用参数
             */
            const getAppParams = () => {
                let formData = []
                const customs: Array<any> = []
                const commands = []
                for (const key in answers.value) {
                    customs.push({
                        name: key,
                        value: answers.value[key],
                        type: 'string'
                    })
                }

                for (const key in helmCommandParams) {
                    if (helmCommandParams[key]) {
                        const obj = {}
                        obj[key] = true
                        commands.push(obj as never)
                    }
                }
                if (timeoutValue.value) {
                    const obj = {}
                    obj['--timeout'] = timeoutValue.value + 's'
                    commands.push(obj as never)
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
                        commands.push(obj as never)
                    }
                })

                if (curEditMode.value === 'yaml-mode') {
                    saveYaml()
                }

                if (bkFormCreater.value) {
                    formData = bkFormCreater.value.getFormData()
                }

                const params = {
                    upgrade_verion: curVersionId.value,
                    answers: formData,
                    customs: customs,
                    cmd_flags: commands,
                    valuefile_name: curValueFile.value,
                    valuefile: yamlFile.value || curTplYaml.value
                }

                return params
            }
            
            /**
             * 显示错误弹层
             * @param  {object} res ajax数据对象
             * @param  {string} title 错误提示
             * @param  {string} actionType 操作
             */
            const showErrorDialog = (res, title, actionType) => {
                errorDialogConf.errorCode = res.code
                errorDialogConf.message = res.message || res.data.msg || res.statusText
                errorDialogConf.title = title
                errorDialogConf.isShow = true
                previewEditorConfig.isShow = false
                updateConfirmDialog.isShow = false

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
             * 更新应用
             */
            const updateApp = async () => {
                if (isDifferenceLoading.value || updateInstanceLoading.value) {
                    return false
                }

                const params = getAppParams()
                const appId = $route.params.appId

                errorDialogConf.isShow = false
                errorDialogConf.message = ''
                errorDialogConf.errorCode = 0

                updateInstanceLoading.value = true
                updateConfirmDialog.isShow = false

                try {
                    const res = await $store.dispatch('helm/updateApp', {
                        projectId: projectId.value,
                        appId,
                        params
                    })
                    checkAppStatus(res.data)
                } catch (e) {
                    showErrorDialog(e, $i18n.t('升级失败'), 'update')
                }
            }
            
            /**
             * 查看app状态，包括创建、更新、回滚、删除
             * @param  {object} app 应用对象
             */
            const checkAppStatus = async (app) => {
                const appId = app.id

                try {
                    const res = await $store.dispatch('helm/checkAppStatus', { projectId: projectId.value, appId })
                    const action = appAction[res.data.transitioning_action]

                    if (res.data.transitioning_on) {
                        setTimeout(() => {
                            checkAppStatus(app)
                        }, 2000)
                    } else {
                        if (res.data.transitioning_result) {
                            $bkMessage({
                                theme: 'success',
                                message: `${app.name}${action}${$i18n.t('成功')}`
                            })
                            // 返回helm首页
                            setTimeout(() => {
                                $router.push({
                                    name: 'helms'
                                })
                            }, 200)
                        } else {
                            updateInstanceLoading.value = false
                            res.data.name = app.name || ''
                            showAppError(res.data)
                        }
                    }
                } catch (e) {
                    updateInstanceLoading.value = false
                    showErrorDialog(e, $i18n.t('操作失败'), 'reback')
                }
            }

            /**
             * 展示App异常信息
             * @param  {object} app 应用对象
             */
            const showAppError = (app) => {
                let actionType = ''
                const res = {
                    code: 500,
                    message: ''
                }

                res.message = app.transitioning_message
                actionType = app.transitioning_action
                const title = `${app.name}${appAction[app.transitioning_action]}${$i18n.t('失败')}`
                showErrorDialog(res, title, actionType)
            }

            /**
             * 点击修改步骤
             * @param {Boolean} val 当前步骤
             */
            const stepChanged = (val) => {
                controllableSteps.curStep = val
            }

            /**
             * 隐藏确认更新弹窗
             */
            const hideConfirmDialog = () => {
                if (updateInstanceLoading.value) {
                    return false
                }
                updateConfirmDialog.isShow = false
            }
            
            /**
             * 显示预览
             */
            const showPreview = async () => {
                if (curEditMode.value === 'yaml-mode' && !checkYaml()) {
                    return false
                }
                if (bkFormCreater.value && !bkFormCreater.value.checkValid()) {
                    return false
                }
                previewEditorConfig.isShow = true
                const params = getAppParams()
                const appId = $route.params.appId

                previewLoading.value = true
                appPreviewList.value = []
                difference.value = ''
                isChartVersionChange.value = false
                treeData.value = []

                try {
                    const res = await $store.dispatch('helm/previewApp', {
                        projectId: projectId.value,
                        appId,
                        params
                    })
                    for (const key in res.data.content) {
                        appPreviewList.value.push({
                            name: key,
                            value: res.data.content[key]
                        })
                    }
                    const tree = path2tree(appPreviewList.value)
                    treeData.value.push(tree)
                    difference.value = res.data.difference
                    curAppDifference.content = res.data.new_content
                    curAppDifference.originContent = res.data.old_content
                    isChartVersionChange.value = res.data.chart_version_changed
                    previewEditorConfig.value = res.data.notes
                    if (appPreviewList.value.length) {
                        curReourceFile.value.value = appPreviewList[0]
                    }
                } catch (e) {
                    showErrorDialog(e, $i18n.t('预览失败'), 'preview')
                    previewEditorConfig.value = ''
                } finally {
                    previewLoading.value = false
                }
            }

            /**
             * 获取文件详情
             * @param  {object} file 文件
             */
            const getFileDetail = (file) => {
                if (file.hasOwnProperty('value')) {
                    curReourceFile.value = file
                }
            }

            const fetchReleaseDetail = async () => {
                isQuestionsLoading.value = true
                updateInstanceLoading.value = true
                const { clusterId, namespace, name } = $route.params
                const res = await $store.dispatch('helm/getReleaseDetail', {
                    $clusterId: clusterId,
                    $namespace: namespace,
                    $name: name
                })
                updateInstanceLoading.value = false
                isQuestionsLoading.value = true

                curApp.value = res
                originReleaseData.value = res
                setAppDetail()
            }

            onMounted(async () => {
                await fetchReleaseDetail()
                await getAppVersions()
                await getNamespaceList()
                winHeight.value = window.innerHeight
            })

            return {
                curApp,
                winHeight,
                tplVersionId,
                curAppVersions,
                curClusterName,
                curValueFileList,
                curValueFile,
                isLocked,
                tabChangeIndex,
                appName,
                formData,
                curEditMode,
                curTplYaml,
                yamlConfig,
                commandList,
                helmCommandParams,
                isHignPanelShow,
                hignSetupMap,
                timeoutValue,
                isTplSynLoading,
                isNamespaceMatch,
                updateConfirmDialog,
                updateInstanceLoading,
                isDifferenceLoading,
                difference,
                isNamespaceLoading,
                yamlDiffEditorOptions,
                isQuestionsLoading,
                goToHelmIndex,
                controllableSteps,
                stepChanged,
                syncHelmTpl,
                handlerVersionChange,
                changeValueFile,
                helmModeChangeHandler,
                errorDialogConf,
                editorInit,
                curAppDifference,
                diffEditorHeight,
                differenceKey,
                editYaml,
                codeViewer,
                addHign,
                delHign,
                handleHignkeyChange,
                confirmUpdateApp,
                hideConfirmDialog,
                updateApp,
                originReleaseVersion,
                previewEditorConfig,
                editorConfig,
                appPreviewList,
                previewLoading,
                isChartVersionChange,
                showPreview,
                curReourceFile,
                treeData,
                getFileDetail
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
        z-index: 20;
    }

    .basic-content {
        .content-item {
            width: 800px;
            display: flex;
            .logo-wapper {
                display: inline-block;
                width: 60px;
                height: 60px;
                margin-right: 24px;
                vertical-align: middle;
                fill: #ebf0f5;
            }
            .basic-wapper {
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
</style>
