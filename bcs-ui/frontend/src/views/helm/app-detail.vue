<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-helm-title">
                <a class="bcs-icon bcs-icon-arrows-left back" href="javascript:void(0);" @click="goToHelmIndex"></a>
                <span>{{curApp.name}}</span>
            </div>
        </div>

        <div class="biz-content-wrapper" v-bkloading="{ isLoading: updateInstanceLoading, zIndex: 100 }">
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
                            <svg class="logo" @click="gotoHelmTplDetail" style="cursor: pointer;">
                                <use xlink:href="#biz-set-icon"></use>
                            </svg>
                            <div class="title">{{curApp.name}}</div>
                            <!-- <p>
                                <a class="bk-text-button f12" href="javascript:void(0);" @click="gotoHelmTplDetail">{{$t('查看Chart详情')}}</a>
                            </p> -->
                            <div class="desc" :title="curApp.description">
                                <span>Chart：</span>
                                <a class="bk-text-button f12 ml5" href="javascript:void(0);" @click="gotoHelmTplDetail">{{curApp.chart_info.name || '--'}}</a>
                            </div>
                            <div class="desc" :title="curApp.description">
                                <span>{{$t('简介')}}：</span>
                                {{curApp.chart_info.description || '--'}}
                            </div>
                            <div class="desc">
                                <span>Notes：</span>
                                <bcs-button v-if="notesdialog.notes" text size="small" style="padding: 0;" @click="handleViewNodesDetails">{{$t('点击查看详情')}}</bcs-button>
                                <span v-else
                                    class="bk-primary bk-button-normal bk-button-text is-disabled"
                                    style="font-size: 12px; height: 26px; line-height: 26px;"
                                    v-bk-tooltips="getNotesTips">
                                    {{$t('点击查看详情')}}
                                </span>
                            </div>
                        </div>
                    </div>

                    <div class="right">
                        <div class="bk-collapse-item bk-collapse-item-active">
                            <div class="biz-item-header" style="cursor: default;">
                                {{$t('配置选项')}}
                            </div>
                            <div class="bk-collapse-item-content f13" style="padding: 15px;">
                                <div class="config-box">
                                    <div class="inner">
                                        <div class="inner-item">
                                            <label class="title">{{$t('名称')}}</label>
                                            <bkbcs-input :value="curApp.name" :readonly="true" />
                                        </div>

                                        <div class="inner-item">
                                            <label class="title">{{$t('版本')}}</label>

                                            <div>
                                                <bk-selector
                                                    :placeholder="$t('请选择')"
                                                    style="width: 215px;"
                                                    searchable
                                                    :selected.sync="tplVersionId"
                                                    :list="curAppVersions"
                                                    setting-key="version"
                                                    :disabled="isTplSynLoading"
                                                    display-key="version"
                                                    search-key="version"
                                                    @item-selected="handlerVersionChange">
                                                </bk-selector>
                                                <!-- <bkbcs-input
                                                    style="width: 215px;"
                                                    type="text"
                                                    :placeholder="$t('请选择')"
                                                    :value.sync="tplVersionId"
                                                    :is-select-mode="true"
                                                    :default-list="curAppVersions"
                                                    :setting-key="'version'"
                                                    :disabled="isTplSynLoading"
                                                    :display-key="'version'"
                                                    :search-key="'version'"
                                                    @item-selected="handlerVersionChange">
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
                                            <bkbcs-input :value="curClusterName" :disabled="true" />
                                        </div>
                                        <div class="inner-item">
                                            <label class="title">
                                                {{$t('命名空间')}}
                                                <span class="ml10 biz-error-tip" v-if="!isNamespaceMatch && !isNamespaceLoading">
                                                    （{{$t('此命名空间不存在')}}）
                                                </span>
                                            </label>
                                            <div>
                                                <bkbcs-input :value="curApp.namespace" :disabled="true" />
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <bk-tab type="border-card" active-name="chart" class="mt20">
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
                                <!-- <span class="f12 vm">{{isLocked ? '已锁定' : '已解锁'}}</span> -->
                                <span class="biz-tip vm">({{$t('默认锁定values内容为当前release')}}({{$t('版本：')}}<span v-bk-tooltips.top="curApp.chart_info.version" class="release-version">{{curApp.chart_info.version}}</span>){{$t('的内容，解除锁定后，加载为对应Chart中的values内容')}})</span>
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
                                    <template v-if="formData.questions">
                                        <bk-form-creater :form-data="formData" ref="bkFormCreater"></bk-form-creater>
                                    </template>
                                    <template v-else>
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
                                <button class="bk-text-button f12 mb10 pl0 mt10" @click.stop.prevent="toggleHign">
                                    {{$t('高级设置')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                                    <i style="font-size: 12px;cursor: pointer;"
                                        class="bcs-icon bcs-icon-info-circle"
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

                <div class="create-wrapper" v-if="!isNamespaceLoading && !isNamespaceMatch">
                    <bcs-popover :content="$t('所属命名空间不存在，不可操作')" placement="top">
                        <bk-button type="primary" :title="$t('更新')" :disabled="true">
                            {{$t('更新')}}
                        </bk-button>
                    </bcs-popover>
                    <bcs-popover :content="$t('所属命名空间不存在，不可操作')" placement="top">
                        <bk-button type="default" :title="$t('预览')" :disabled="true">
                            {{$t('预览')}}
                        </bk-button>
                    </bcs-popover>
                    <bk-button type="default" :title="$t('取消')" @click="goToHelmIndex">
                        {{$t('取消')}}
                    </bk-button>
                </div>

                <div class="create-wrapper" v-else>
                    <bk-button type="primary" :title="$t('更新')" @click="confirmUpdateApp" :disabled="isNamespaceLoading || !isNamespaceMatch">
                        {{$t('更新')}}
                    </bk-button>
                    <bk-button type="default" :title="$t('预览')" @click="showPreview" :disabled="isNamespaceLoading || !isNamespaceMatch">
                        {{$t('预览')}}
                    </bk-button>
                    <bk-button type="default" :title="$t('取消')" @click="goToHelmIndex">
                        {{$t('取消')}}
                    </bk-button>
                </div>
            </div>
        </div>

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
                                <bcs-tree
                                    :data="treeData"
                                    :node-key="'name'"
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
            :position="{ top: 80 }"
            :width="1100"
            :title="updateConfirmDialog.title"
            :close-icon="!updateInstanceLoading"
            :is-show.sync="updateConfirmDialog.isShow"
            @cancel="hideConfirmDialog">
            <template slot="content">
                <p class="biz-tip mb5 tl" style="color: #666;">{{$t('Helm Release参数发生如下变化，请确认后再点击“确定”更新')}}</p>
                <div class="difference-code" v-bkloading="{ isLoading: isDifferenceLoading }" v-if="isDifferenceLoading || difference" style="height: 400px;">
                    <div class="editor-header">
                        <div>当前版本</div>
                        <div>更新版本</div>
                    </div>

                    <div :class="['diff-editor-box', { 'editor-fullscreen': yamlDiffEditorOptions.fullScreen }]" style="position: relative;">
                        <!-- <div title="关闭全屏" class="fullscreen-close" v-if="yamlDiffEditorOptions.fullScreen" @click="cancelFullScreen">
                            <i class="bcs-icon bcs-icon-close"></i>
                        </div>
                        <div title="全屏" class="fullscreen-use" v-else @click="setFullScreen">
                            <i class="bcs-icon bcs-icon-full-screen"></i>
                        </div> -->
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
                    <bk-button type="primary" @click="hideErrorDialog">{{$t('知道了')}}</bk-button>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="codeDialogConf.isShow"
            :width="1100"
            :has-header="false"
            :has-footet="false"
            :title="codeDialogConf.title"
            @cancel="hideCodeDiffDialog">
            <template slot="content">
                <div class="code-diff-header">
                    <h3>{{$t('当前 Release 参数')}}：</h3>
                    <h3>{{$t('Chart 默认值')}}：</h3>
                </div>
                <div style="max-height: 500px; overflow: auto; position: relative; border: 1px solid #ddd; border-radius: 2px;">
                    <div :class="['diff-editor-box', { 'editor-fullscreen': yamlDiffEditorOptions.fullScreen }]" style="position: relative;">
                        <div :title="$t('关闭全屏')" class="fullscreen-close" v-if="yamlDiffEditorOptions.fullScreen" @click="cancelFullScreen">
                            <i class="bcs-icon bcs-icon-close"></i>
                        </div>
                        <div :title="$t('全屏')" class="fullscreen-use" v-else @click="setFullScreen">
                            <i class="bcs-icon bcs-icon-full-screen"></i>
                        </div>
                        <monaco-editor
                            ref="yamlEditor"
                            class="editor"
                            theme="monokai"
                            language="yaml"
                            :style="{ height: `${diffEditorHeight}px`, width: '100%' }"
                            v-model="curEditYaml"
                            :diff-editor="true"
                            :key="differenceKey"
                            :options="yamlDiffEditorOptions"
                            :original="instanceYamlValue">
                        </monaco-editor>
                    </div>
                </div>
            </template>
            <div slot="footer">
                <div class="biz-footer">
                    <bk-button type="primary" @click="hideCodeDiffDialog" class="mr5">{{$t('知道了')}}</bk-button>
                </div>
            </div>
        </bk-dialog>

        <bcs-dialog v-model="notesdialog.isShow"
            title="Notes"
            header-position="left"
            :width="750"
            @cancel="handleCloseNotes">
            <div class="notes-details">
                <pre class="notes-message">{{notesdialog.notes}}</pre>
                <bcs-button class="copy-button"
                    id="notes-message"
                    size="small"
                    type="default"
                    :data-clipboard-text="notesdialog.notes">
                    <i class="bcs-icon bcs-icon-clipboard mr5"></i>
                    {{$t('复制')}}
                </bcs-button>
            </div>
            <template slot="footer">
                <bcs-button type="primary" @click="handleCloseNotes">{{$t('关闭')}}</bcs-button>
            </template>
        </bcs-dialog>
    </div>
</template>

<script>
    import yamljs from 'js-yaml'
    import path2tree from '@/common/path2tree'
    import baseMixin from '@/mixins/helm/mixin-base'
    import { catchErrorHandler } from '@/common/util'
    import Clipboard from 'clipboard'
    import MonacoEditor from '@/components/monaco-editor/editor.vue'
    import resizer from '@/components/resize'

    export default {
        components: {
            MonacoEditor,
            resizer
        },
        mixins: [baseMixin],
        data () {
            return {
                tabChangeIndex: 0,
                tempProjectId: '',
                curTplReadme: '',
                curEditMode: 'yaml-mode',
                yamlEditor: null,
                yamlFile: '',
                curTplYaml: '',
                activeName: ['config'],
                collapseName: ['preview'],
                tplsetVerList: [],
                appPreviewList: [],
                isNamespaceLoading: true, // 命名空间加载中，如果没有命名空间，无法进行操作
                updateInstanceLoading: false,
                isDifferenceLoading: false,
                isQuestionsLoading: true,
                previewLoading: false,
                isNamespaceMatch: false, // 判断命名空间是否已经删除
                isRouterLeave: false,
                isTplSynLoading: false,
                curVersionId: -1,
                isAppVerLoading: true,
                instanceYamlValue: '', // 当前应用实例化后的配置
                instanceValueFileName: '', // 用户实例化选择的value文件名
                yamlDiffEditorOptions: {
                    readOnly: true,
                    fontSize: 14,
                    fullScreen: false
                },
                // previewList: [],
                difference: '',
                differenceKey: 0,
                curAppDifference: {
                    content: '',
                    originContent: ''
                },
                isChartVersionChange: false,
                appName: '',
                tplVersionId: -1,
                originReleaseVersion: '',
                originReleaseData: {},
                formData: {},
                fieldset: [],
                winHeight: 0,
                editor: null,
                errorDialogConf: {
                    title: '',
                    isShow: false,
                    message: '',
                    errorCode: 0
                },
                codeDialogConf: {
                    title: this.$t('和当前版本对比'),
                    isShow: false
                },
                // curVersionYaml: '',
                curEditYaml: '',
                curReourceFile: {
                    name: '',
                    value: ''
                },
                updateConfirmDialog: {
                    title: this.$t('确认更新'),
                    isShow: false,
                    width: '100%',
                    height: '100%',
                    lang: 'yaml',
                    closeIcon: true,
                    readOnly: true,
                    fullScreen: false,
                    values: [],
                    editors: []
                },
                curApp: {
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
                },
                treeData: [],
                editorConfig: {
                    width: '100%',
                    height: '100%',
                    lang: 'yaml',
                    readOnly: true,
                    fullScreen: false,
                    values: [],
                    editors: []
                },
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
                isSyncYamlLoading: false,
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
                curAppVersions: [],
                namespaceId: '',
                answers: {},
                namespaceList: [],
                appAction: {
                    create: this.$t('部署'),
                    noop: '',
                    update: this.$t('更新'),
                    rollback: this.$t('回滚'),
                    delete: this.$t('删除'),
                    destroy: this.$t('删除')
                },
                curValueFileList: [],
                curValueFile: 'values.yaml',
                isLocked: true,
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
                notesdialog: {
                    isShow: false,
                    notes: '',
                    clipboard: null
                },
                isNotesLoading: false,
                isHignPanelShow: true,
                hignSetupMap: [], // helm部署配置高级设置
                timeoutValue: 600,
                hignDesc: this.$t('设置Flags，如设置wait，输入格式为 --wait = true')
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
            diffEditorHeight () {
                return this.yamlDiffEditorOptions.fullScreen ? window.innerHeight : 315
            },
            curClusterName () {
                if (this.curApp.cluster_id !== undefined) {
                    const match = this.clusterList.find(item => item.cluster_id === this.curApp.cluster_id)
                    return match ? match.name : this.curApp.cluster_id
                }
                return ''
            },
            curLabelList () {
                const customs = this.curApp.release.customs
                const answers = {}
                const list = []
                customs.forEach(item => {
                    list.push({
                        key: item.name,
                        value: item.value
                    })
                    answers[item.name] = item.value
                })
                if (!list.length) {
                    list.push({
                        name: '',
                        value: ''
                    })
                }
                this.answers = answers
                return list
            },
            getNotesTips () {
                return this.isNotesLoading ? this.$t('加载中') : this.$t('Notes 为空')
            },
            clusterList () {
                return this.$store.state.cluster.clusterList
            }
        },
        watch: {
            tplVersionId () {
                this.tabChangeIndex++
            },
            isLocked () {
                this.initValuesFileData(this.curTplName, this.curTplFiles)
                this.tabChangeIndex++
            }
        },
        async mounted () {
            const appId = this.$route.params.appId
            this.curApp = await this.getAppById(appId)
            this.getAppVersions()
            this.getNamespaceList()
            this.winHeight = window.innerHeight
            await this.getNotes()
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
             * 显示yaml对比
             */
            showCodeDiffDialog () {
                this.differenceKey++
                this.curEditYaml = this.yamlEditor.getValue()
                this.codeDialogConf.isShow = true
            },

            /**
             * 隐藏yaml对比
             */
            hideCodeDiffDialog () {
                this.codeDialogConf.isShow = false
            },

            /**
             * 访问模板详情
             */
            gotoHelmTplDetail () {
                const route = this.$router.resolve({
                    name: 'helmTplDetail',
                    params: {
                        projectCode: this.projectCode,
                        tplId: this.curApp.chart
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
             * 获取文件详情
             * @param  {object} file 文件
             */
            getFileDetail (file) {
                if (file.hasOwnProperty('value')) {
                    this.curReourceFile = file
                }
            },

            /**
             * 返回Helm应用首页
             */
            goToHelmIndex () {
                this.$router.push({
                    name: 'helms',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode
                    }
                })
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
                    this.editYaml()
                } else {
                    this.saveYaml()
                }
            },

            /**
             * 编辑yamml
             */
            async editYaml () {
                let formData = []

                this.curEditMode = 'yaml-mode'
                this.isSyncYamlLoading = true
                // 将数据配置的数据和yaml的数据进行合并同步
                if (this.$refs.bkFormCreater) {
                    formData = this.$refs.bkFormCreater.getFormData()
                }
                if (this.curTplYaml) {
                    yamljs.load(this.curTplYaml)
                }

                this.yamlConfig.isShow = true
                try {
                    const res = await this.$store.dispatch('helm/syncJsonToYaml', {
                        json: formData,
                        yaml: this.curTplYaml
                    })
                    // this.curTplYaml = res.data.yaml.replace(/\'/ig, '\"')
                    this.curTplYaml = res.data.yaml
                    this.$nextTick(() => {
                        this.yamlEditor.gotoLine(0, 0)
                    })
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isSyncYamlLoading = false
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
                    // 通过load检测yaml是否合法
                    yamljs.load(yaml)
                } catch (err) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入合法的YAML')
                    })
                    return false
                }

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
                let formData = []
                let yamlData = {}

                try {
                    // 通过load检测yaml是否合法
                    yamljs.load(yaml)
                } catch (err) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入合法的YAML')
                    })
                    return false
                }

                // 同步到数据配置
                if (yaml) {
                    yamlData = yamljs.load(yaml)
                }
                if (this.$refs.bkFormCreater) {
                    formData = this.$refs.bkFormCreater.getFormData()
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
             * 获取应用
             * @param  {number} appId 应用ID
             * @return {object} result 应用
             */
            async getAppById (appId) {
                let result = {}
                const projectId = this.projectId

                this.isQuestionsLoading = true
                try {
                    const res = await this.$store.dispatch('helm/getAppById', { projectId, appId })
                    result = res.data
                    this.originReleaseData = result
                    this.setAppDetail()
                    this.isQuestionsLoading = false
                } catch (e) {
                    if (e.status === 404) {
                        this.goToHelmIndex()
                    } else {
                        catchErrorHandler(e, this)
                    }
                    this.isQuestionsLoading = false
                }

                return result
            },

            setAppDetail () {
                const files = this.originReleaseData.release.chartVersionSnapshot.files
                const tplName = this.originReleaseData.release.chartVersionSnapshot.name
                const questions = this.originReleaseData.release.chartVersionSnapshot.questions
                this.curTplReadme = files[`${tplName}/README.md`]
                this.curValueFile = this.originReleaseData.valuefile_name || 'values.file'
                this.curTplYaml = this.originReleaseData.valuefile
                this.instanceYamlValue = this.originReleaseData.valuefile // 保存当前应用实例化后的配置
                this.instanceValueFileName = this.originReleaseData.valuefile_name // 保存用户实例化时选择的文件名
                this.curTplName = tplName
                this.curTplFiles = files
                this.initValuesFileData(tplName, files, this.originReleaseData.valuefile_name)
                this.$nextTick(() => {
                    this.$refs.codeViewer && this.$refs.codeViewer.$ace && this.$refs.codeViewer.$ace.scrollToLine(1, true, true)
                })

                if (this.originReleaseData.cmd_flags && this.originReleaseData.cmd_flags.length) {
                    // 所有常用枚举项keys
                    const commonKeys = Object.keys(this.helmCommandParams)
                    this.originReleaseData.cmd_flags.forEach(item => {
                        const stringKey = Object.keys(item).join(',')
                        // 常用枚举项不包含则是用户自定义高级配置
                        if (!commonKeys.includes(stringKey)) {
                            const obj = {}
                            obj.key = stringKey
                            obj.value = item[stringKey]
                            this.hignSetupMap.push(obj)
                        } else {
                            if (stringKey === '--timeout') {
                                this.timeoutValue = item[stringKey].slice(0, item[stringKey].length - 1)
                            } else {
                                this.helmCommandParams[stringKey] = true
                            }
                        }
                    })
                }

                // 如果没有用户自定义helm配置, 默认添加一条空数据
                if (!this.hignSetupMap.length) {
                    const obj = {
                        key: '',
                        value: ''
                    }
                    this.hignSetupMap.push(obj)
                }

                if (questions.questions) {
                    questions.questions.forEach(question => {
                        this.fieldset = this.originReleaseData.release.answers
                        if (this.fieldset && this.fieldset.length) {
                            this.fieldset.forEach(item => {
                                if (question.variable === item.name) {
                                    question.default = item.value
                                }
                            })
                        }

                        if (question.subquestions) {
                            question.subquestions.forEach(subQuestion => {
                                if (this.fieldset && this.fieldset.length) {
                                    this.fieldset.forEach(item => {
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
             * 切换应用版本号回调
             * @param  {number} index 索引
             * @param  {object} data 版本对象
             */
            async handlerVersionChange (index, data) {
                if (data.version === this.originReleaseVersion) {
                    this.setAppDetail()
                    this.curVersionId = -1
                    return false
                }
                const projectId = this.projectId
                const appName = this.curApp.name
                const appId = this.curApp.id
                const clusterId = this.curApp.cluster_id
                const namespace = this.curApp.namespace
                const version = data.version
                const versionId = data.id

                // this.curVersionYaml = ''
                try {
                    const fnPath = this.$INTERNAL ? 'helm/getUpdateChartVersionDetail' : 'helm/getUpdateChartByVersion'
                    const res = await this.$store.dispatch(fnPath, {
                        projectId,
                        appId: this.$INTERNAL ? appName : appId,
                        version: this.$INTERNAL ? version : versionId,
                        clusterId: this.$INTERNAL ? clusterId : undefined,
                        namespace: this.$INTERNAL ? namespace : undefined
                    })

                    const files = res.data.data.files
                    const tplName = res.data.name
                    this.formData = res.data.data.questions

                    this.curTplName = tplName
                    this.curTplFiles = files
                    this.curVersionId = res.data.id
                    this.initValuesFileData(tplName, files)
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            initValuesFileData (tplName, files, valueFileName) {
                const list = []
                const regex = new RegExp(`^${tplName}\\/[\\w-]*values.(yaml|yml)$`)
                const bcsTplName = tplName + '/bcs-values'
                const bcsRegex = new RegExp(`^${bcsTplName}\\/[\\w-]*.(yaml|yml)$`)

                this.yamlFile = ''

                // 根据valueFileName判断是否第一次展示, curTplYaml显示实例化配置的内容
                if (valueFileName) {
                    this.curValueFile = valueFileName
                    // this.curVersionYaml = files[valueFileName]
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
                this.curValueFileList = list

                // 选择版本后（valueFileName为空）
                if (!valueFileName) {
                    // 如果锁定则用release values
                    if (this.isLocked) {
                        this.curTplYaml = this.instanceYamlValue
                        this.curValueFile = this.instanceValueFileName
                        console.log('locked usedefault')
                    } else {
                        // 选择版本后依然是原来的, 显示原来实例化时的配置内容
                        if (String(this.tplVersionId) === this.originReleaseVersion) {
                            this.curTplYaml = this.instanceYamlValue
                            this.curValueFile = this.instanceValueFileName
                            console.log('unlocked but no change usedefault')
                        } else {
                            const fileNames = Object.keys(files)
                            const matchName = fileNames.find(name => {
                                return name.endsWith(this.instanceValueFileName)
                            })
                            // 默认用原来的value文件
                            if (matchName) {
                                this.curTplYaml = files[matchName]
                                this.curValueFile = this.instanceValueFileName
                            } else {
                                const valueDefaultNames = `${tplName}/values.yaml`
                                this.curTplYaml = files[valueDefaultNames]
                                this.curValueFile = 'values.yaml'
                            }
                            console.log('unlocked and change usenew')
                        }
                    }
                }
            },

            /**
             * 修改value file
             */
            changeValueFile (index, data) {
                this.curValueFile = index
                // this.curVersionYaml = data.content
                this.curTplYaml = data.content
                this.yamlFile = ''

                // 没有选择过版本时, 如果切换为原来实例化的文件名，显示原来实例化时的配置内容
                if (String(this.tplVersionId) === this.originReleaseVersion && index === this.instanceValueFileName) {
                    this.curTplYaml = this.instanceYamlValue
                    console.log('unlocked && nochange use default')
                } else {
                    console.log('unlocked && valsuechange')
                }
            },

            /**
             * 设置预览文件
             * @param {array} files 文件
             */
            setPreviewList (files) {
                const list = []
                for (const key in files) {
                    list.push({
                        name: key,
                        value: files[key]
                    })
                }
                // this.previewList.splice(0, this.previewList.length, ...list)
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
                        this.isAppVerLoading = true
                        this.getAppVersions()
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
             * 获取应用版本列表
             * @param  {number} appId 应用ID
             */
            async getAppVersions () {
                this.curAppVersions = []
                try {
                    const projectId = this.projectId
                    const appId = this.$INTERNAL ? this.curApp.name : this.curApp.id
                    const clusterId = this.curApp.cluster_id
                    const namespace = this.curApp.namespace
                    if (this.$INTERNAL) {
                        const res = await this.$store.dispatch('helm/getUpdateVersionList', { projectId, clusterId, appId, namespace })
                        this.curAppVersions = res.data
                    } else {
                        const res = await this.$store.dispatch('helm/getUpdateVersions', { projectId, appId })
                        this.curAppVersions = res.data.results
                    }
                    if (this.curAppVersions.length) {
                        this.originReleaseVersion = this.curAppVersions[0].version
                        this.tplVersionId = this.curAppVersions[0].version
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isAppVerLoading = false
                }
            },

            /**
             * 获取命名空间列表
             */
            async getNamespaceList () {
                const projectId = this.projectId
                this.isNamespaceLoading = true

                try {
                    const res = await this.$store.dispatch('helm/getNamespaceList', {
                        projectId: projectId
                    })
                    const curNamespaceId = this.curApp.namespace_id
                    this.isNamespaceMatch = false

                    // this.clusterList = []
                    res.data.forEach(item => {
                        const obj = {}
                        const match = item.name.match(/^([\s\S]*)\(([\w-]*)\)/)
                        if (match && match.length > 2) {
                            obj.name = match[1]
                            obj.id = match[2]
                            item.id = match[2]
                        } else {
                            obj.name = item.name
                            obj.id = item.name
                        }
                        // this.clusterList.push(obj)

                        if (item.children) {
                            item.children.forEach(child => {
                                if (child.id === curNamespaceId) {
                                    this.isNamespaceMatch = true
                                }
                            })
                        }
                    })
                    this.namespaceList = res.data
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isNamespaceLoading = false
                }
            },

            /**
             * 显示确认更新弹窗
             */
            confirmUpdateApp () {
                if (this.curEditMode === 'yaml-mode' && !this.checkYaml()) {
                    return false
                }
                if (this.$refs.bkFormCreater && !(this.$refs.bkFormCreater.checkValid())) {
                    return false
                }
                this.isDifferenceLoading = true
                this.updateConfirmDialog.isShow = true
                this.getDifference()
            },

            /**
             * 获取应用参数
             * @return {object} params 应用参数
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
                    formData = this.$refs.bkFormCreater.getFormData()
                }

                const params = {
                    upgrade_verion: this.curVersionId,
                    answers: formData,
                    customs: customs,
                    cmd_flags: commands,
                    valuefile_name: this.curValueFile
                }

                params.valuefile = this.yamlFile || this.curTplYaml
                return params
            },

            /**
             * 隐藏确认更新弹窗
             */
            hideConfirmDialog () {
                if (this.updateInstanceLoading) {
                    return false
                }
                this.updateConfirmDialog.isShow = false
            },

            /**
             * 显示错误弹层
             * @param  {object} res ajax数据对象
             * @param  {string} title 错误提示
             * @param  {string} actionType 操作
             */
            showErrorDialog (res, title, actionType) {
                this.errorDialogConf.errorCode = res.code
                this.errorDialogConf.message = res.message || res.data.msg || res.statusText
                this.errorDialogConf.title = title
                this.errorDialogConf.isShow = true
                this.previewEditorConfig.isShow = false
                this.updateConfirmDialog.isShow = false

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
                const title = `${app.name}${this.appAction[app.transitioning_action]}${this.$t('失败')}`
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
                                message: `${app.name}${action}${this.$t('成功')}`
                            })
                            // 返回helm首页
                            setTimeout(() => {
                                this.$router.push({
                                    name: 'helms'
                                })
                            }, 200)
                        } else {
                            this.updateInstanceLoading = false
                            res.data.name = app.name || ''
                            this.showAppError(res.data)
                        }
                    }
                } catch (e) {
                    this.updateInstanceLoading = false
                    this.showErrorDialog(e, this.$t('操作失败'), 'reback')
                }
            },

            /**
             * 更新应用
             */
            async updateApp () {
                if (this.isDifferenceLoading || this.updateInstanceLoading) {
                    return false
                }

                const params = this.getAppParams()
                const projectId = this.projectId
                const appId = this.$route.params.appId

                this.errorDialogConf.isShow = false
                this.errorDialogConf.message = ''
                this.errorDialogConf.errorCode = 0

                this.updateInstanceLoading = true
                this.updateConfirmDialog.isShow = false

                try {
                    const res = await this.$store.dispatch('helm/updateApp', {
                        projectId,
                        appId,
                        params
                    })
                    this.checkAppStatus(res.data)
                } catch (e) {
                    this.showErrorDialog(e, this.$t('更新失败'), 'update')
                }
            },

            /**
             * 显示预览
             */
            async showPreview () {
                if (this.curEditMode === 'yaml-mode' && !this.checkYaml()) {
                    return false
                }
                if (this.$refs.bkFormCreater && !(this.$refs.bkFormCreater.checkValid())) {
                    return false
                }
                this.previewEditorConfig.isShow = true
                const params = this.getAppParams()
                const projectId = this.projectId
                const appId = this.$route.params.appId

                this.previewLoading = true
                this.appPreviewList = []
                this.difference = ''
                this.isChartVersionChange = false
                this.treeData = []

                try {
                    const res = await this.$store.dispatch('helm/previewApp', {
                        projectId,
                        appId,
                        params
                    })
                    for (const key in res.data.content) {
                        this.appPreviewList.push({
                            name: key,
                            value: res.data.content[key]
                        })
                    }
                    const tree = path2tree(this.appPreviewList)
                    this.treeData.push(tree)
                    this.difference = res.data.difference
                    this.curAppDifference.content = res.data.new_content
                    this.curAppDifference.originContent = res.data.old_content
                    this.isChartVersionChange = res.data.chart_version_changed
                    this.previewEditorConfig.value = res.data.notes
                    if (this.appPreviewList.length) {
                        this.curReourceFile = this.appPreviewList[0]
                    }
                } catch (e) {
                    this.showErrorDialog(e, this.$t('预览失败'), 'preview')
                    this.previewEditorConfig.value = ''
                } finally {
                    this.previewLoading = false
                }
            },

            /**
             * 获取版本对比
             */
            async getDifference () {
                const params = this.getAppParams()
                const projectId = this.projectId
                const appId = this.$route.params.appId
                this.isDifferenceLoading = true
                this.difference = ''
                this.isChartVersionChange = ''

                try {
                    const res = await this.$store.dispatch('helm/previewApp', {
                        projectId,
                        appId,
                        params
                    })
                    this.difference = res.data.difference
                    // for (const key in res.data.content) {
                    //     this.curAppDifference.content += res.data.content[key]
                    // }
                    this.curAppDifference.content = res.data.new_content
                    this.curAppDifference.originContent = res.data.old_content
                    this.differenceKey++
                    this.isChartVersionChange = res.data.chart_version_changed
                } catch (e) {
                    this.showErrorDialog(e, this.$t('Chart渲染失败'), 'preUpdate')
                    this.updateConfirmDialog.value = ''
                } finally {
                    this.isDifferenceLoading = false
                }
            },

            /**
             * 全屏
             */
            setFullScreen (index) {
                this.yamlDiffEditorOptions.fullScreen = true
                this.differenceKey++
            },

            /**
             * 取消全屏
             */
            cancelFullScreen () {
                this.yamlDiffEditorOptions.fullScreen = false
                this.differenceKey++
            },

            handleToggleLock () {
                this.isLocked = !this.isLocked
                this.initValuesFileData(this.curTplName, this.curTplFiles)
            },

            async getNotes () {
                this.isNotesLoading = true
                try {
                    const res = await this.$store.dispatch('helm/getNotes', {
                        projectId: this.projectId,
                        clusterId: this.curApp.cluster_id,
                        namespaceName: this.curApp.namespace,
                        releaseName: this.curApp.name
                    })
                    const data = res.data || {}
                    this.notesdialog.notes = data.notes || ''
                } catch (e) {
                    console.error(e)
                } finally {
                    this.isNotesLoading = false
                }
            },

            handleViewNodesDetails () {
                this.notesdialog.isShow = true
                if (this.notesdialog.notes) {
                    this.$nextTick(() => {
                        this.notesdialog.clipboard = new Clipboard('#notes-message')
                        this.notesdialog.clipboard.on('success', e => {
                            this.$bkMessage({
                                theme: 'success',
                                message: this.$t('复制成功')
                            })
                        })
                    })
                }
            },

            handleCloseNotes () {
                this.notesdialog.isShow = false
                if (this.notesdialog.clipboard && this.notesdialog.clipboard.off) {
                    this.notesdialog.clipboard.off('success')
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
            }
        }
    }
</script>

<style scoped>
    @import './app-detail.css';
</style>
