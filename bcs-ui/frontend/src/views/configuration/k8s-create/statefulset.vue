<template>
    <div class="biz-content">
        <biz-header ref="commonHeader"
            @exception="exceptionHandler"
            @saveStatefulsetSuccess="saveStatefulsetSuccess"
            @switchVersion="initResource"
            @exmportToYaml="exportToYaml">
        </biz-header>
        <template>
            <div class="biz-content-wrapper biz-confignation-wrapper" v-bkloading="{ isLoading: isTemplateSaving }">
                <app-exception
                    v-if="exceptionCode && !isDataLoading"
                    :type="exceptionCode.code"
                    :text="exceptionCode.msg">
                </app-exception>
                <div class="biz-tab-box" v-else v-show="!isDataLoading">
                    <biz-tabs @tab-change="tabResource" ref="commonTab"></biz-tabs>
                    <div class="biz-tab-content" v-bkloading="{ isLoading: isTabChanging }">
                        <bk-alert type="info" class="mb20">
                            <div slot="title">
                                {{$t('StatefulSet是k8s中标准的有状态服务，区别于Deployment，它是一个给Pod提供唯一标志的控制器，可以保证部署和扩展的顺序')}}，
                                <a class="bk-text-button" :href="PROJECT_CONFIG.doc.k8sStatefulset" target="_blank">{{$t('详情查看文档')}}</a>
                            </div>
                        </bk-alert>
                        <template v-if="!statefulsets.length">
                            <div class="biz-guide-box mt0">
                                <bk-button type="primary" @click.stop.prevent="addLocalApplication">
                                    <i class="bcs-icon bcs-icon-plus"></i>
                                    <span style="margin-left: 0;">{{$t('添加')}}StatefulSet</span>
                                </bk-button>
                            </div>
                        </template>
                        <template v-else>
                            <div class="biz-configuration-topbar">
                                <div class="biz-list-operation">
                                    <div class="item" v-for="(application, index) in statefulsets" :key="application.id">
                                        <bk-button :class="['bk-button', { 'bk-primary': curApplication.id === application.id }]" @click.stop="setCurApplication(application, index)">
                                            {{(application && application.config.metadata.name) || $t('未命名')}}
                                            <span class="biz-update-dot" v-show="application.isEdited"></span>
                                        </bk-button>
                                        <span class="bcs-icon bcs-icon-close" @click.stop="removeApplication(application, index)"></span>
                                    </div>

                                    <bcs-popover ref="applicationTooltip" :content="$t('添加StatefulSet')" placement="top">
                                        <bk-button class="bk-button bk-default is-outline is-icon" @click.stop="addLocalApplication">
                                            <i class="bcs-icon bcs-icon-plus"></i>
                                        </bk-button>
                                    </bcs-popover>
                                </div>
                            </div>
                            <div class="biz-configuration-content" style="position: relative;">
                                <!-- part1 start -->
                                <div class="bk-form biz-configuration-form">
                                    <a href="javascript:void(0);" class="bk-text-button from-json-btn" @click.stop.prevent="showJsonPanel">{{$t('导入YAML')}}</a>

                                    <bk-sideslider
                                        :is-show.sync="toJsonDialogConf.isShow"
                                        :title="toJsonDialogConf.title"
                                        :width="toJsonDialogConf.width"
                                        :quick-close="false"
                                        class="biz-app-container-tojson-sideslider"
                                        @hidden="closeToJson">
                                        <div slot="content" style="position: relative;">
                                            <div class="biz-log-box" :style="{ height: `${winHeight - 60}px` }" v-bkloading="{ isLoading: toJsonDialogConf.loading }">
                                                <bk-button class="bk-button bk-primary save-json-btn" @click.stop.prevent="saveApplicationJson">{{$t('导入')}}</bk-button>
                                                <bk-button class="bk-button bk-default hide-json-btn" @click.stop.prevent="hideApplicationJson">{{$t('取消')}}</bk-button>
                                                <ace
                                                    :value="editorConfig.value"
                                                    :width="editorConfig.width"
                                                    :height="editorConfig.height"
                                                    :lang="editorConfig.lang"
                                                    :read-only="editorConfig.readOnly"
                                                    :full-screen="editorConfig.fullScreen"
                                                    @init="editorInitAfter">
                                                </ace>
                                            </div>
                                        </div>
                                    </bk-sideslider>

                                    <div class="bk-form-item">
                                        <div class="bk-form-item">
                                            <div class="bk-form-content" style="margin-left: 0;">
                                                <div class="bk-form-inline-item is-required">
                                                    <label class="bk-label" style="width: 140px;">{{$t('应用名称')}}：</label>
                                                    <div class="bk-form-content" style="margin-left: 140px;">
                                                        <div class="bk-form-input-group">
                                                            <input type="text" :class="['bk-form-input',{ 'is-danger': errors.has('applicationName') }]" :placeholder="$t('请输入64个字符以内')" style="width: 310px;" v-model="curApplication.config.metadata.name" maxlength="64" name="applicationName" v-validate="{ required: true, regex: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/ }">
                                                        </div>
                                                    </div>
                                                </div>
                                                <div class="bk-form-inline-item is-required">
                                                    <label class="bk-label" style="width: 140px;">{{$t('实例数量')}}：</label>
                                                    <div class="bk-form-content" style="margin-left: 140px;">
                                                        <div class="bk-form-input-group">
                                                            <bkbcs-input
                                                                type="number"
                                                                :placeholder="$t('请输入')"
                                                                style="width: 310px;"
                                                                :min="0"
                                                                :value.sync="curApplication.config.spec.replicas"
                                                                :list="varList"
                                                            >
                                                            </bkbcs-input>
                                                        </div>
                                                    </div>
                                                </div>
                                                <div class="bk-form-tip is-danger" style="margin-left: 140px;" v-if="errors.has('applicationName')">
                                                    <p class="bk-tip-text">{{$t('名称必填，以小写字母或数字开头和结尾，只能包含：小写字母、数字、连字符(-)、点(.)')}}</p>
                                                </div>
                                            </div>
                                        </div>
                                    </div>

                                    <div class="bk-form-item" v-if="curApplication.service_tag">
                                        <label class="bk-label" style="width: 140px;">{{$t('关联Service')}}：</label>
                                        <div class="bk-form-content" style="margin-left: 140px;">
                                            <div class="bk-dropdown-box" style="width: 310px;" @click="reloadServices">
                                                <!-- <input type="text" class="bk-form-input" :value="linkServiceName" disabled> -->
                                                <bk-selector
                                                    :placeholder="$t('请选择关联的Service')"
                                                    :setting-key="'service_tag'"
                                                    :display-key="'service_name'"
                                                    :selected.sync="curApplication.service_tag"
                                                    :list="serviceList"
                                                    :prevent-init-trigger="'true'"
                                                    :is-loading="isLoadingServices"
                                                    :disabled="true"
                                                >
                                                </bk-selector>
                                            </div>
                                        </div>
                                    </div>

                                    <div class="bk-form-item">
                                        <label class="bk-label" style="width: 140px;">{{$t('重要级别')}}：</label>
                                        <div class="bk-form-content" style="margin-left: 140px;">
                                            <bk-radio-group v-model="curApplication.config.monitorLevel">
                                                <bk-radio class="mr20" :value="'important'">{{$t('重要')}}</bk-radio>
                                                <bk-radio class="mr20" :value="'general'">{{$t('一般')}}</bk-radio>
                                                <bk-radio :value="'unimportant'">{{$t('不重要')}}</bk-radio>
                                            </bk-radio-group>
                                        </div>
                                    </div>

                                    <div class="bk-form-item">
                                        <label class="bk-label" style="width: 140px;">{{$t('描述')}}：</label>
                                        <div class="bk-form-content" style="margin-left: 140px;">
                                            <textarea class="bk-form-textarea" :placeholder="$t('请输入256个字符以内')" v-model="curApplication.desc" maxlength="256"></textarea>
                                        </div>
                                    </div>

                                    <div class="bk-form-item is-required">
                                        <label class="bk-label" style="width: 140px;">{{$t('标签')}}：</label>
                                        <div class="bk-form-content" style="margin-left: 140px;">
                                            <bk-keyer
                                                :key-list.sync="curLabelList"
                                                :var-list="varList"
                                                ref="labelKeyer"
                                                @change="updateApplicationLabel"
                                                :is-link-to-selector="true"
                                                :is-tip-change="isSelectorChange"
                                                :tip="$t('k8S使用选择器(spec.selector.matchLabels)关联资源，实例化后选择器的值不可变')"
                                                :can-disabled="true">
                                                <slot>
                                                    <p class="biz-tip" style="line-height: 1;">{{$t('小提示：K8S使用选择器（spec.selector.matchLabels）关联资源，必填，且实例化后选择器的值不可变')}}</p>
                                                </slot>
                                            </bk-keyer>
                                        </div>
                                    </div>

                                    <div class="bk-form-item">
                                        <div class="bk-form-content" style="margin-left: 140px;">
                                            <button :class="['bk-text-button f12 mb10 pl0', { 'rotate': isMorePanelShow }]" @click.stop.prevent="toggleMore">
                                                {{$t('更多设置')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                                            </button>

                                            <button :class="['bk-text-button f12 mb10 pl0', { 'rotate': isPodPanelShow }]" @click.stop.prevent="togglePod">
                                                {{$t('Pod模板设置')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                                            </button>
                                        </div>
                                        <bk-tab :type="'fill'" :active-name="'tab1'" :size="'small'" v-show="isMorePanelShow" style="margin-left: 140px;">

                                            <bk-tab-panel name="tab1" :title="$t('更新策略')">
                                                <div class="bk-form m20">
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 155px;">{{$t('类型')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 155px;">
                                                            <bk-radio-group v-model="curApplication.config.spec.updateStrategy.type">
                                                                <bk-radio :value="'OnDelete'">
                                                                    OnDelete
                                                                </bk-radio>
                                                                <bk-radio :value="'RollingUpdate'">
                                                                    RollingUpdate
                                                                </bk-radio>
                                                            </bk-radio-group>
                                                        </div>
                                                    </div>

                                                    <div class="bk-form-item" v-show="curApplication.config.spec.updateStrategy.type === 'RollingUpdate'">
                                                        <label class="bk-label" style="width: 155px;">Partition：</label>
                                                        <div class="bk-form-content" style="margin-left: 155px;">
                                                            <bkbcs-input
                                                                type="number"
                                                                :placeholder="$t('请输入')"
                                                                style="width: 250px;"
                                                                :min="0"
                                                                :max="curApplication.config.spec.replicas - 1"
                                                                :value.sync="curApplication.config.spec.updateStrategy.rollingUpdate.partition"
                                                                :list="varList"
                                                            >
                                                            </bkbcs-input>
                                                        </div>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>
                                            <bk-tab-panel name="tab2" :title="$t('Pod管理策略')">
                                                <div class="bk-form m20">
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 210px;">PodManagementPolicy：</label>
                                                        <div class="bk-form-content" style="margin-left: 210px;">
                                                            <bk-radio-group v-model="curApplication.config.spec.podManagementPolicy">
                                                                <bk-radio :value="'OrderedReady'">OrderedReady</bk-radio>
                                                                <bk-radio :value="'Parallel'">Parallel</bk-radio>
                                                            </bk-radio-group>
                                                        </div>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>
                                            <!-- <bk-tab-panel name="tab3" title="卷模板">
                                                <div class="bk-form m20">
                                                    <table class="biz-simple-table" style="width: 720px;">
                                                        <thead>
                                                            <tr>
                                                                <th style="width: 150px;">挂载名</th>
                                                                <th style="width: 187px;">StorageClassName</th>
                                                                <th>大小</th>
                                                                <th style="width: 187px;">访问模式</th>
                                                                <th style="width: 73px;"></th>
                                                            </tr>
                                                        </thead>
                                                        <tbody>
                                                            <tr v-for="(vol, index) in curApplication.config.spec.volumeClaimTemplates" :key="index">
                                                                <td>
                                                                    <bkbcs-input
                                                                        type="text"
                                                                        :placeholder="$t('请输入')"
                                                                        :value.sync="vol.metadata.name"
                                                                        :list="varList"
                                                                    >
                                                                    </bkbcs-input>
                                                                </td>
                                                                <td>
                                                                    <bk-selector
                                                                        :placeholder="$t('请选择')"
                                                                        :setting-key="'id'"
                                                                        :selected.sync="vol.spec.storageClassName"
                                                                        :list="[]">
                                                                    </bk-selector>
                                                                </td>
                                                                <td>
                                                                    <div class="bk-form-input-group">
                                                                        <bk-number-input
                                                                            :value.sync="vol.spec.resources.requests.storage"
                                                                            :min="0"
                                                                            :hide-operation="true"
                                                                            :ex-style="{ 'width': '80px' }"
                                                                            :placeholder="'请输入'">
                                                                        </bk-number-input>
                                                                        <span class="input-group-addon">
                                                                            Gi
                                                                        </span>
                                                                    </div>
                                                                </td>
                                                                <td>
                                                                    <bk-selector
                                                                        :placeholder="$t('请选择')"
                                                                        :multi-select="true"
                                                                        :setting-key="'id'"
                                                                        :selected.sync="vol.spec.accessModes"
                                                                        :list="volVisitList">
                                                                    </bk-selector>
                                                                </td>
                                                                <td>
                                                                    <div class="action-box">
                                                                        <bk-button class="action-btn ml5" @click.stop.prevent="addVolTpl()">
                                                                            <i class="bcs-icon bcs-icon-plus"></i>
                                                                        </bk-button>
                                                                        <bk-button class="action-btn" @click.stop.prevent="removeVolTpl(vol, index)" v-show="curApplication.config.spec.volumeClaimTemplates.length > 1">
                                                                            <i class="bcs-icon bcs-icon-minus"></i>
                                                                        </bk-button>
                                                                    </div>
                                                                </td>
                                                            </tr>
                                                        </tbody>
                                                    </table>
                                                </div>
                                            </bk-tab-panel> -->
                                            <!-- <bk-tab-panel name="tab4" :title="$t('Metric信息')">
                                                <div class="bk-form m20">
                                                    <div class="bk-form-item">
                                                        <div class="bk-form-content" style="margin-left: 0;">
                                                            <bk-checkbox style="margin-left: 15px;"
                                                                name="mdtric-policy"
                                                                v-model="curApplication.config.webCache.isMetric">
                                                                {{$t('启用Metric数据采集')}}
                                                            </bk-checkbox>
                                                            <router-link class="bk-text-button ml10" style="vertical-align: middle;" :to="{ name: 'metricManage', params: { projectCode, projectId } }">{{$t('管理Metric')}}</router-link>
                                                        </div>
                                                    </div>
                                                    <transition name="fade">
                                                        <div class="bk-form-item" style="margin-top: 15px;" v-if="curApplication.config.webCache.isMetric">
                                                            <div class="bk-form-content" style="margin-left: 15px;">
                                                                <div class="bk-dropdown-box" style="width: 310px; display: block;">
                                                                    <bk-selector
                                                                        :placeholder="$t('Metric数列（多选）')"
                                                                        field-type="metric"
                                                                        :setting-key="'id'"
                                                                        :display-key="'name'"
                                                                        :multi-select="true"
                                                                        :searchable="true"
                                                                        :selected.sync="curApplication.config.webCache.metricIdList"
                                                                        :list="metricList">
                                                                    </bk-selector>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </transition>
                                                </div>
                                            </bk-tab-panel> -->
                                            <bk-tab-panel name="ta5" :title="'HostAliases'">
                                                <div class="bk-form m20">
                                                    <table class="biz-simple-table" style="width: 720px;" v-if="curApplication.config.webCache.hostAliasesCache && curApplication.config.webCache.hostAliasesCache.length">
                                                        <thead>
                                                            <tr>
                                                                <th style="width: 200px;">IP</th>
                                                                <th style="width: 400px;">HostNames</th>
                                                                <th></th>
                                                            </tr>
                                                        </thead>
                                                        <tbody>
                                                            <tr v-for="(hostAlias, index) in curApplication.config.webCache.hostAliasesCache" :key="index">
                                                                <td>
                                                                    <bkbcs-input
                                                                        type="text"
                                                                        :placeholder="$t('请输入')"
                                                                        :value.sync="hostAlias.ip"
                                                                        :list="varList">
                                                                    </bkbcs-input>
                                                                </td>
                                                                <td>
                                                                    <bkbcs-input
                                                                        type="text"
                                                                        :placeholder="$t('请输入，多个HostName以英文分号(;)分隔')"
                                                                        :value.sync="hostAlias.hostnames"
                                                                        :list="varList">
                                                                    </bkbcs-input>
                                                                </td>
                                                                <td>
                                                                    <div class="action-box">
                                                                        <bk-button class="action-btn ml5" @click.stop.prevent="addHostAlias">
                                                                            <i class="bcs-icon bcs-icon-plus"></i>
                                                                        </bk-button>
                                                                        <bk-button class="action-btn" @click.stop.prevent="removeHostAlias(hostAlias, index)">
                                                                            <i class="bcs-icon bcs-icon-minus"></i>
                                                                        </bk-button>
                                                                    </div>
                                                                </td>
                                                            </tr>
                                                        </tbody>
                                                    </table>
                                                    <div class="tc p40" v-else>
                                                        <bk-button type="primary" @click="addHostAlias">
                                                            <i class="bcs-icon bcs-icon-plus f12" style="top: -2px;"></i>
                                                            {{$t('添加 HostAlias')}}
                                                        </bk-button>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>
                                        </bk-tab>

                                        <bk-tab :type="'fill'" :active-name="'tab2'" :size="'small'" v-show="isPodPanelShow" style="margin-left: 105px;">
                                            <bk-tab-panel name="tab2" :title="$t('注解')">
                                                <div class="bk-form m20">
                                                    <bk-keyer :key-list.sync="curRemarkList" :var-list="varList" ref="remarkKeyer" @change="updateApplicationRemark"></bk-keyer>
                                                </div>
                                            </bk-tab-panel>
                                            <bk-tab-panel name="tab3" :title="$t('Restart策略')">
                                                <div class="bk-form m20">
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 105px;">{{$t('重启策略')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 105px;">
                                                            <bk-radio-group v-model="curApplication.config.spec.template.spec.restartPolicy">
                                                                <bk-radio :value="policy" v-for="(policy, index) in restartPolicy" :key="index">
                                                                    {{policy}}
                                                                </bk-radio>
                                                            </bk-radio-group>
                                                        </div>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>
                                            <bk-tab-panel name="tab4" :title="$t('Kill策略')">
                                                <div class="bk-form m20">
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 105px;">{{$t('宽期限')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 105px;">
                                                            <div class="bk-form-input-group">
                                                                <bkbcs-input
                                                                    type="number"
                                                                    :placeholder="$t('请输入')"
                                                                    style="width: 80px;"
                                                                    :min="0"
                                                                    :value.sync="curApplication.config.spec.template.spec.terminationGracePeriodSeconds"
                                                                    :list="varList"
                                                                >
                                                                </bkbcs-input>
                                                                <span class="input-group-addon">
                                                                    {{$t('秒')}}
                                                                </span>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="tab5" :title="$t('调度约束')">
                                                <div class="bk-form m20">
                                                    <p class="title mb5">NodeSelector</p>
                                                    <bk-keyer :key-list.sync="curConstraintLabelList" :var-list="varList" ref="nodeSelectorKeyer" @change="updateNodeSelectorList"></bk-keyer>
                                                    <div class="mb5 mt10">
                                                        <span class="title">{{$t('亲和性约束')}}</span>
                                                        <bk-checkbox class="ml10" name="image-get" v-model="curApplication.config.webCache.isUserConstraint">{{$t('启用')}}</bk-checkbox>
                                                    </div>

                                                    <div style="height: 300px;" v-if="curApplication.config.webCache.isUserConstraint">
                                                        <ace
                                                            :value="curApplication.config.webCache.affinityYaml"
                                                            :width="yamlEditorConfig.width"
                                                            :height="yamlEditorConfig.height"
                                                            :lang="yamlEditorConfig.lang"
                                                            :read-only="yamlEditorConfig.readOnly"
                                                            :full-screen="yamlEditorConfig.fullScreen"
                                                            @init="yamlEditorInitAfter"
                                                            @input="yamlEditorInput"
                                                            @blur="yamlEditorBlur">
                                                        </ace>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="tab6" :title="$t('网络')">
                                                <div class="bk-form m20">
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 105px;">{{$t('网络策略')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 105px;">
                                                            <bk-selector
                                                                style="width: 300px;"
                                                                :placeholder="$t('请选择')"
                                                                :setting-key="'id'"
                                                                :display-key="'name'"
                                                                :selected.sync="curApplication.config.spec.template.spec.hostNetwork"
                                                                :list="netStrategyList"
                                                                @item-selected="changeDNSPolicy">
                                                            </bk-selector>
                                                        </div>
                                                    </div>
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 105px;">{{$t('DNS策略')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 105px;">
                                                            <bk-selector
                                                                style="width: 300px;"
                                                                :placeholder="$t('请选择')"
                                                                :setting-key="'id'"
                                                                :display-key="'name'"
                                                                :selected.sync="curApplication.config.spec.template.spec.dnsPolicy"
                                                                :list="dnsStrategyList">
                                                            </bk-selector>
                                                        </div>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="ta7" :title="$t('卷')">
                                                <div class="bk-form m20">
                                                    <table class="biz-simple-table" v-if="curApplication.config.webCache.volumes.length">
                                                        <thead>
                                                            <tr>
                                                                <th style="width: 200px;">{{$t('类型')}}</th>
                                                                <th style="width: 220px;">{{$t('挂载名')}}</th>
                                                                <th>{{$t('挂载源')}}</th>
                                                                <th style="width: 100px;"></th>
                                                            </tr>
                                                        </thead>
                                                        <tbody>
                                                            <tr v-for="(volume, index) in curApplication.config.webCache.volumes" :key="index">
                                                                <td>
                                                                    <bk-selector
                                                                        :placeholder="$t('类型')"
                                                                        :setting-key="'id'"
                                                                        :selected.sync="volume.type"
                                                                        :list="volumeTypeList">
                                                                    </bk-selector>
                                                                </td>
                                                                <td>
                                                                    <bkbcs-input
                                                                        type="text"
                                                                        :placeholder="$t('请输入')"
                                                                        :value.sync="volume.name"
                                                                        :list="varList"
                                                                    >
                                                                    </bkbcs-input>
                                                                </td>
                                                                <td>
                                                                    <template v-if="volume.type === 'emptyDir'">
                                                                        <bkbcs-input value="{}" :disabled="true" />
                                                                    </template>
                                                                    <template v-if="volume.type === 'emptyDir(Memory)'">
                                                                        <div class="source-flex-box">
                                                                            <bkbcs-input value="Memory" :disabled="true" style="width: 75px;" />
                                                                            <div class="bk-form-input-group">
                                                                                <bkbcs-input
                                                                                    type="number"
                                                                                    placeholder="sizeLimit"
                                                                                    :min="0"
                                                                                    :value.sync="volume.source"
                                                                                    :list="varList"
                                                                                >
                                                                                </bkbcs-input>
                                                                                <span class="input-group-addon">
                                                                                    Gi
                                                                                </span>
                                                                            </div>
                                                                        </div>
                                                                    </template>
                                                                    <template v-else-if="volume.type === 'persistentVolumeClaim'">
                                                                        <bk-selector
                                                                            placeholder="PVC List"
                                                                            :setting-key="'id'"
                                                                            :searchable="true"
                                                                            :selected.sync="volume.source"
                                                                            :list="[]">
                                                                        </bk-selector>
                                                                    </template>
                                                                    <template v-else-if="volume.type === 'hostPath'">
                                                                        <bkbcs-input v-model="volume.source" :placeholder="$t('请输入')" />
                                                                    </template>
                                                                    <template v-else-if="volume.type === 'configMap'">
                                                                        <bk-selector
                                                                            placeholder="Configmap List"
                                                                            :setting-key="'id'"
                                                                            :display-key="'name'"
                                                                            :searchable="true"
                                                                            :selected.sync="volume.source"
                                                                            :list="volumeConfigmapAllList">
                                                                        </bk-selector>
                                                                    </template>
                                                                    <template v-else-if="volume.type === 'secret'">
                                                                        <bk-selector
                                                                            placeholder="Secret List"
                                                                            :setting-key="'name'"
                                                                            :display-key="'name'"
                                                                            :searchable="true"
                                                                            :selected.sync="volume.source"
                                                                            :list="volumeSecretList">
                                                                        </bk-selector>
                                                                    </template>
                                                                </td>
                                                                <td>
                                                                    <div class="action-box">
                                                                        <bk-button class="action-btn ml5" @click.stop.prevent="addVolumn()">
                                                                            <i class="bcs-icon bcs-icon-plus"></i>
                                                                        </bk-button>
                                                                        <bk-button class="action-btn" @click.stop.prevent="removeVolumn(volume, index)">
                                                                            <i class="bcs-icon bcs-icon-minus"></i>
                                                                        </bk-button>
                                                                    </div>
                                                                </td>
                                                            </tr>
                                                        </tbody>
                                                    </table>
                                                    <div class="tc p40" v-else>
                                                        <bk-button type="primary" @click="addVolumn">
                                                            <i class="bcs-icon bcs-icon-plus f12" style="top: -2px;"></i>
                                                            {{$t('添加卷')}}
                                                        </bk-button>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="tab8" :title="$t('日志采集')">
                                                <div class="bk-form p20">
                                                    <div class="biz-expand-panel">
                                                        <div class="panel">
                                                            <div class="header">
                                                                <span class="topic">{{$t('标准日志')}}</span>
                                                            </div>
                                                            <div class="bk-form-item content">
                                                                <ul>
                                                                    <li>
                                                                        <bk-checkbox name="type" :value="true" :disabled="true">{{$t('标准输出：包含容器Stdout日志')}}</bk-checkbox>
                                                                    </li>
                                                                </ul>
                                                            </div>
                                                        </div>
                                                        <div class="panel mt0">
                                                            <div class="header" style="border-top: 1px solid #dfe0e5;">
                                                                <div class="topic">
                                                                    {{$t('附加日志标签')}}
                                                                    <bcs-popover :content="$t('附加的日志标签会以KV的形式追加到采集日志中')" placement="top">
                                                                        <span class="bk-badge">
                                                                            <i class="bcs-icon bcs-icon-question-circle"></i>
                                                                        </span>
                                                                    </bcs-popover>
                                                                </div>
                                                            </div>
                                                            <div class="bk-form-item content">
                                                                <bk-keyer
                                                                    :key-list.sync="curLogLabelList"
                                                                    :var-list="varList"
                                                                    @change="updateApplicationLogLabel">
                                                                </bk-keyer>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="tab9" :title="$t('镜像凭证')">
                                                <div class="bk-form m20">
                                                    <bk-keyer
                                                        :data-key="'name'"
                                                        :key-list.sync="curImageSecretList"
                                                        :var-list="varList"
                                                        :key-input-width="170"
                                                        :value-input-width="450"
                                                        :tip="$t('提示：实际对应配置的imagePullSecrets字段')"
                                                        @change="updateApplicationImageSecrets">
                                                    </bk-keyer>
                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="tab10" :title="$t('服务帐户')">
                                                <div class="bk-form m20">
                                                    <div class="biz-equal-inputer">
                                                        <div class="inputer-content">
                                                            <bkbcs-input
                                                                type="text"
                                                                style="width: 170px;"
                                                                value="serviceAccountName"
                                                                :disabled="true">
                                                            </bkbcs-input>
                                                            <span class="operator">=</span>
                                                            <bkbcs-input
                                                                type="text"
                                                                style="width: 450px;"
                                                                :placeholder="$t('值')"
                                                                :value.sync="curApplication.config.spec.template.spec.serviceAccountName">
                                                            </bkbcs-input>
                                                        </div>
                                                        <p class="biz-tip mt5">{{$t('提示：创建 Pod 时，如果没有指定服务账户，Pod 会被指定成对应命名空间中的 default 服务账户')}}</p>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>
                                        </bk-tab>
                                    </div>
                                </div>
                                <!-- part1 end -->

                                <!-- part2 start -->
                                <div class="biz-part-header">
                                    <div class="bk-button-group">
                                        <div class="item" v-for="(container, index) in curApplication.config.spec.template.spec.allContainers" :key="index">
                                            <bk-button :class="['bk-button bk-default is-outline', { 'is-selected': curContainerIndex === index }]" @click.stop="setCurContainer(container, index)">
                                                {{container.name || $t('未命名')}}
                                            </bk-button>
                                            <span class="bcs-icon bcs-icon-close-circle" @click.stop="removeContainer(index)" v-if="curApplication.config.spec.template.spec.allContainers.length > 1"></span>
                                        </div>
                                        <bcs-popover ref="containerTooltip" :content="$t('添加Container')" placement="top">
                                            <bk-button type="button" class="bk-button bk-default is-outline is-icon" @click.stop.prevent="addLocalContainer">
                                                <i class="bcs-icon bcs-icon-plus"></i>
                                            </bk-button>
                                        </bcs-popover>
                                    </div>
                                </div>

                                <div class="bk-form biz-configuration-form pb15">
                                    <div class="biz-span">
                                        <span class="title">{{$t('基础信息')}}</span>
                                    </div>
                                    <div class="bk-form-item is-required">
                                        <div class="bk-form-content" style="margin-left: 0">
                                            <div class="bk-form-inline-item is-required">
                                                <label class="bk-label" style="width: 140px;">{{$t('容器名称')}}：</label>
                                                <div class="bk-form-content" style="margin-left: 140px;">
                                                    <input type="text" :class="['bk-form-input', { 'is-danger': errors.has('containerName') }]" :placeholder="$t('请输入64个字符以内')" style="width: 310px;" v-model="curContainer.name" maxlength="64" name="containerName" v-validate="{ required: true, regex: /^[a-z]{1}[a-z0-9-]{0,63}$/ }">
                                                </div>
                                            </div>

                                            <div class="bk-form-inline-item">
                                                <label class="bk-label" style="width: 105px;">{{$t('类型')}}：</label>
                                                <div class="bk-form-content" style="margin-left: 105px;">
                                                    <bk-radio-group v-model="curContainer.webCache.containerType">
                                                        <bk-radio :value="'container'">
                                                            Container
                                                            <i class="bcs-icon bcs-icon-question-circle ml5" v-bk-tooltips="$t('应用Container')"></i>
                                                        </bk-radio>
                                                        <bk-radio :value="'initContainer'">
                                                            InitContainer
                                                            <i class="bcs-icon bcs-icon-question-circle ml5" v-bk-tooltips="$t('用于在启动应用Container之前，启动一个或多个“初始化”容器，完成应用Container所需的预置条件')"></i>
                                                        </bk-radio>
                                                    </bk-radio-group>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                    <div class="bk-form-item">
                                        <label class="bk-label" style="width: 140px;">{{$t('描述')}}：</label>
                                        <div class="bk-form-content" style="margin-left: 140px;">
                                            <textarea name="" id="" cols="30" rows="10" class="bk-form-textarea" :placeholder="$t('请输入256个字符以内')" v-model="curContainer.webCache.desc" maxlength="256"></textarea>
                                        </div>
                                    </div>
                                    <div class="bk-form-item is-required">
                                        <label class="bk-label" style="width: 140px;">{{$t('镜像及版本')}}：</label>
                                        <div class="bk-form-content" style="margin-left: 140px;">
                                            <div class="mb10">
                                                <span @click="handleChangeImageMode">
                                                    <bk-switcher
                                                        :selected="curContainer.webCache.isImageCustomed"
                                                        size="small"
                                                        :key="curContainer.name">
                                                    </bk-switcher>
                                                </span>
                                                <span class="vm">{{$t('使用自定义镜像')}}</span>
                                                <span class="biz-tip vm">({{$t('启用后允许直接填写镜像信息')}})</span>
                                            </div>
                                            <template v-if="curContainer.webCache.isImageCustomed">
                                                <bkbcs-input
                                                    type="text"
                                                    style="width: 325px;"
                                                    :placeholder="$t('镜像')"
                                                    :value.sync="curContainer.webCache.imageName"
                                                    @change="handleImageCustom">
                                                </bkbcs-input>
                                                <bkbcs-input
                                                    type="text"
                                                    style="width: 250px;"
                                                    :placeholder="$t('版本号1')"
                                                    :value.sync="curContainer.imageVersion"
                                                    @change="handleImageCustom">
                                                </bkbcs-input>
                                            </template>
                                            <template v-else>
                                                <div class="bk-dropdown-box" style="width: 380px;">
                                                    <bk-combox
                                                        style="width: 325px;"
                                                        type="text"
                                                        :placeholder="$t('镜像')"
                                                        :key="renderImageIndex"
                                                        :display-key="'_name'"
                                                        :setting-key="'_id'"
                                                        :search-key="'_name'"
                                                        :value.sync="curContainer.webCache.imageName"
                                                        :list="varList"
                                                        :is-link="true"
                                                        :is-select-mode="true"
                                                        :default-list="imageList"
                                                        @item-selected="changeImage(...arguments, curContainer)">
                                                    </bk-combox>

                                                    <bk-button
                                                        style="min-width: 20px;"
                                                        class="bk-button bk-default is-outline is-icon"
                                                        v-bk-tooltips.top="$t('刷新镜像列表')"
                                                        @click="initImageList">
                                                        <div class="bk-spin-loading bk-spin-loading-mini bk-spin-loading-default" style="margin-top: -3px;" v-if="isLoadingImageList">
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

                                                <div class="bk-dropdown-box" style="width: 250px;">
                                                    <bk-combox
                                                        type="text"
                                                        :placeholder="$t('版本号1')"
                                                        :display-key="'_name'"
                                                        :setting-key="'_id'"
                                                        :search-key="'_name'"
                                                        :value.sync="curContainer.imageVersion"
                                                        :list="varList"
                                                        :is-select-mode="true"
                                                        :default-list="imageVersionList"
                                                        :disabled="!curContainer.webCache.imageName"
                                                        @item-selected="setImageVersion">
                                                    </bk-combox>
                                                </div>
                                            </template>

                                            <bk-checkbox
                                                class="ml10"
                                                name="image-get"
                                                :true-value="'Always'"
                                                :false-value="'IfNotPresent'"
                                                v-model="curContainer.imagePullPolicy">
                                                {{$t('总是在创建之前拉取镜像')}}
                                            </bk-checkbox>

                                            <p class="biz-tip mt5" v-if="!isLoadingImageList && !imageList.length">{{$t('提示：项目镜像不存在，')}}
                                                <router-link class="bk-text-button" :to="{ name: 'projectImage', params: { projectCode, projectId } }">{{$t('去创建')}}</router-link>
                                            </p>
                                        </div>
                                    </div>

                                    <div class="biz-span">
                                        <span class="title">{{$t('端口映射')}}</span>
                                    </div>

                                    <div class="bk-form-item">
                                        <div class="bk-form-content" style="margin-left: 140px;">
                                            <table class="biz-simple-table">
                                                <thead>
                                                    <tr>
                                                        <th style="width: 330px;">{{$t('名称')}}</th>
                                                        <th style="width: 135px;">{{$t('协议')}}</th>
                                                        <th style="width: 135px;">{{$t('容器端口')}}</th>
                                                        <th></th>
                                                    </tr>
                                                </thead>
                                                <tbody>
                                                    <tr v-for="(port, index) in curContainer.ports" :key="index">
                                                        <td>
                                                            <bkbcs-input
                                                                type="text"
                                                                :placeholder="$t('名称')"
                                                                style="width: 325px;"
                                                                maxlength="255"
                                                                :value.sync="port.name"
                                                                :list="varList"
                                                            >
                                                            </bkbcs-input>
                                                        </td>
                                                        <td>
                                                            <bk-selector
                                                                :placeholder="$t('协议')"
                                                                :setting-key="'id'"
                                                                :selected.sync="port.protocol"
                                                                :list="protocolList">
                                                            </bk-selector>
                                                        </td>
                                                        <td>
                                                            <bkbcs-input
                                                                type="number"
                                                                placeholder="1-65535"
                                                                style="width: 135px;"
                                                                :value.sync="port.containerPort"
                                                                :min="1"
                                                                :max="65535"
                                                                :list="varList"
                                                            >
                                                            </bkbcs-input>
                                                        </td>
                                                        <td>
                                                            <bk-button class="action-btn ml5" @click.stop.prevent="addPort">
                                                                <i class="bcs-icon bcs-icon-plus"></i>
                                                            </bk-button>
                                                            <bk-button class="action-btn" v-if="curContainer.ports.length > 1" @click.stop.prevent="removePort(port, index)">
                                                                <i class="bcs-icon bcs-icon-minus"></i>
                                                            </bk-button>
                                                        </td>
                                                    </tr>
                                                </tbody>
                                            </table>
                                            <p class="biz-tip">{{$t('提示：容器端口是容器内部的Port。在配置Service的端口映射时，通过"目标端口"进行关联，从而暴露服务')}}</p>
                                        </div>
                                    </div>

                                    <div class="biz-span">
                                        <div class="title">
                                            <button :class="['bk-text-button', { 'rotate': isPartBShow }]" @click.stop.prevent="togglePartB">
                                                {{$t('更多设置')}}<i class="bcs-icon bcs-icon-angle-double-down f12 ml5 mb10 fb"></i>
                                            </button>
                                        </div>
                                    </div>

                                    <div style="margin-left: 140px;" v-show="isPartBShow">
                                        <bk-tab :type="'fill'" :active-name="'tab1'" :size="'small'">
                                            <bk-tab-panel name="tab1" :title="$t('命令')">
                                                <div class="bk-form m20">
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 140px;">{{$t('启动命令')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 140px;">
                                                            <bkbcs-input
                                                                type="text"
                                                                :placeholder="$t('例如/bin/bash')"
                                                                :value.sync="curContainer.command"
                                                                :list="varList"
                                                            >
                                                            </bkbcs-input>
                                                        </div>
                                                    </div>
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 140px;">{{$t('命令参数')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 140px;">
                                                            <bkbcs-input
                                                                type="text"
                                                                :placeholder="$t('多个参数用空格分隔，例如&quot;-c&quot;  &quot;while true; do echo hello; sleep 10;done&quot;')"
                                                                :value.sync="curContainer.args"
                                                                :list="varList"
                                                            >
                                                            </bkbcs-input>
                                                        </div>
                                                    </div>
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 140px;">{{$t('工作目录')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 140px;">
                                                            <bkbcs-input
                                                                type="text"
                                                                :placeholder="$t('例如{path}', { path: '/mywork' })"
                                                                :value.sync="curContainer.workingDir"
                                                                :list="varList"
                                                            >
                                                            </bkbcs-input>
                                                        </div>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="tab2" :title="$t('挂载卷')">
                                                <div class="bk-form m20">
                                                    <template v-if="curMountVolumes.length">
                                                        <template v-if="curContainer.volumeMounts.length">
                                                            <p class="biz-tip mb10">
                                                                {{$t('请先在"Pod模板设置" -> "卷"中设置')}}</p>
                                                            <table class="biz-simple-table">
                                                                <thead>
                                                                    <tr>
                                                                        <th style="width: 200px;">{{$t('卷')}}</th>
                                                                        <th style="width: 300px;">{{$t('容器目录')}}</th>
                                                                        <th style="width: 200px;">{{$t('子目录')}}</th>
                                                                        <th style="width: 70px;"></th>
                                                                        <th></th>
                                                                    </tr>
                                                                </thead>
                                                                <tbody>
                                                                    <tr v-for="(volumeItem, index) in curContainer.volumeMounts" :key="index">
                                                                        <td>
                                                                            <bk-selector
                                                                                :placeholder="$t('请选择')"
                                                                                :setting-key="'name'"
                                                                                :display-key="'name'"
                                                                                :allow-clear="true"
                                                                                :selected.sync="volumeItem.name"
                                                                                :list="curMountVolumes"
                                                                                @item-selected="selectVolumeType(volumeItem)">
                                                                            </bk-selector>
                                                                        </td>
                                                                        <td>
                                                                            <bkbcs-input
                                                                                type="text"
                                                                                placeholder="MountPath"
                                                                                maxlength="512"
                                                                                :value.sync="volumeItem.mountPath"
                                                                                :list="varList"
                                                                            >
                                                                            </bkbcs-input>
                                                                        </td>
                                                                        <td>
                                                                            <bkbcs-input
                                                                                type="text"
                                                                                placeholder="SubPath"
                                                                                maxlength="200"
                                                                                :value.sync="volumeItem.subPath"
                                                                                :list="varList">
                                                                            </bkbcs-input>
                                                                        </td>
                                                                        <td>
                                                                            <div class="biz-input-wrapper">
                                                                                <bk-checkbox v-model="volumeItem.readOnly">{{$t('只读')}}</bk-checkbox>
                                                                            </div>
                                                                        </td>
                                                                        <div class="action-box">
                                                                            <bk-button class="action-btn ml5" @click.stop.prevent="addMountVolumn()">
                                                                                <i class="bcs-icon bcs-icon-plus"></i>
                                                                            </bk-button>
                                                                            <bk-button class="action-btn" @click.stop.prevent="removeMountVolumn(volumeItem, index)">
                                                                                <i class="bcs-icon bcs-icon-minus"></i>
                                                                            </bk-button>
                                                                        </div>
                                                                    </tr>
                                                                </tbody>
                                                            </table>
                                                        </template>
                                                        <div class="tc p40" v-else>
                                                            <bk-button type="primary" @click="addMountVolumn">
                                                                <i class="bcs-icon bcs-icon-plus f12" style="top: -2px;"></i>
                                                                {{$t('添加挂载卷')}}
                                                            </bk-button>
                                                        </div>
                                                    </template>
                                                    <div v-else class="tc p30">
                                                        {{$t('请先在"Pod模板设置" -> "卷"中设置')}}
                                                    </div>
                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="tab3" :title="$t('环境变量')">
                                                <div class="bk-form m20">
                                                    <table class="biz-simple-table" style="width: 690px;">
                                                        <thead>
                                                            <tr>
                                                                <th style="width: 160px;">{{$t('类型')}}</th>
                                                                <th style="width: 220px;">{{$t('变量键')}}</th>
                                                                <th style="width: 220px;">{{$t('变量值')}}</th>
                                                                <th></th>
                                                            </tr>
                                                        </thead>
                                                        <tbody>
                                                            <tr v-for="(env, index) in curContainer.webCache.env_list" :key="index">
                                                                <td>
                                                                    <bk-selector
                                                                        :placeholder="$t('类型')"
                                                                        :setting-key="'id'"
                                                                        :selected.sync="env.type"
                                                                        :list="mountTypeList">
                                                                    </bk-selector>
                                                                </td>
                                                                <td v-if="['valueFrom', 'custom', 'configmapKey', 'secretKey'].includes(env.type)">
                                                                    <bkbcs-input
                                                                        type="text"
                                                                        :placeholder="$t('请输入')"
                                                                        :value.sync="env.key"
                                                                        :list="varList"
                                                                    >
                                                                    </bkbcs-input>
                                                                </td>
                                                                <td :colspan="['valueFrom', 'custom', 'configmapKey', 'secretKey'].includes(env.type) ? 1 : 2">
                                                                    <template v-if="['valueFrom', 'custom'].includes(env.type)">
                                                                        <bkbcs-input
                                                                            type="text"
                                                                            :placeholder="$t('例如{path}', { path: '/metadata/name' })"
                                                                            :value.sync="env.value"
                                                                            :list="varList"
                                                                        >
                                                                        </bkbcs-input>
                                                                    </template>
                                                                    <template v-else-if="['configmapKey'].includes(env.type)">
                                                                        <bk-selector
                                                                            :placeholder="$t('请选择')"
                                                                            :setting-key="'id'"
                                                                            :selected.sync="env.value"
                                                                            :list="configmapKeyList"
                                                                            @item-selected="updateEnvItem(...arguments, env)">
                                                                        </bk-selector>
                                                                    </template>
                                                                    <template v-else-if="['secretKey'].includes(env.type)">
                                                                        <bk-selector
                                                                            :placeholder="$t('请选择')"
                                                                            :setting-key="'id'"
                                                                            :selected.sync="env.value"
                                                                            :list="secretKeyList"
                                                                            @item-selected="updateEnvItem(...arguments, env)">
                                                                        </bk-selector>
                                                                    </template>
                                                                    <template v-else-if="['configmapFile'].includes(env.type)">
                                                                        <bk-selector
                                                                            :placeholder="$t('ConfigMap列表')"
                                                                            :setting-key="'name'"
                                                                            :selected.sync="env.value"
                                                                            :list="volumeConfigmapList">
                                                                        </bk-selector>
                                                                    </template>
                                                                    <template v-else-if="['secretFile'].includes(env.type)">
                                                                        <bk-selector
                                                                            :placeholder="$t('Secret列表')"
                                                                            :setting-key="'name'"
                                                                            :selected.sync="env.value"
                                                                            :list="volumeSecretList">
                                                                        </bk-selector>
                                                                    </template>
                                                                </td>
                                                                <td>
                                                                    <div class="action-box">
                                                                        <bk-button class="action-btn ml5" @click.stop.prevent="addEnv()">
                                                                            <i class="bcs-icon bcs-icon-plus"></i>
                                                                        </bk-button>
                                                                        <bk-button class="action-btn" @click.stop.prevent="removeEnv(env, index)" v-show="curContainer.webCache.env_list.length > 1">
                                                                            <i class="bcs-icon bcs-icon-minus"></i>
                                                                        </bk-button>
                                                                    </div>
                                                                </td>
                                                            </tr>
                                                        </tbody>
                                                    </table>
                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="tab4" :title="$t('资源限制')">
                                                <div class="bk-form m20">
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 105px;">{{$t('特权')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 105px;">
                                                            <bk-checkbox v-model="curContainer.securityContext.privileged">{{$t('可完全访问母机资源')}}</bk-checkbox>
                                                        </div>
                                                    </div>
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 105px;">CPU：</label>
                                                        <div class="bk-form-content" style="margin-left: 105px;">
                                                            <div class="bk-form-input-group mr5">
                                                                <span class="input-group-addon is-left">
                                                                    requests
                                                                </span>
                                                                <bkbcs-input
                                                                    type="number"
                                                                    style="width: 100px;"
                                                                    :min="0"
                                                                    :max="curContainer.resources.limits.cpu ? curContainer.resources.limits.cpu : Number.MAX_VALUE"
                                                                    :placeholder="$t('请输入')"
                                                                    :value.sync="curContainer.resources.requests.cpu"
                                                                    :list="varList"
                                                                >
                                                                </bkbcs-input>
                                                                <span class="input-group-addon">
                                                                    m
                                                                </span>
                                                            </div>
                                                            <bcs-popover :content="$t('设置CPU requests，1000m CPU=1核 CPU')" placement="top">
                                                                <span class="bk-badge">
                                                                    <i class="bcs-icon bcs-icon-question-circle"></i>
                                                                </span>
                                                            </bcs-popover>

                                                            <div class="bk-form-input-group ml20 mr5">
                                                                <span class="input-group-addon is-left">
                                                                    limits
                                                                </span>
                                                                <bkbcs-input
                                                                    type="number"
                                                                    style="width: 100px;"
                                                                    :min="0"
                                                                    :placeholder="$t('请输入')"
                                                                    :value.sync="curContainer.resources.limits.cpu"
                                                                    :list="varList">
                                                                </bkbcs-input>
                                                                <span class="input-group-addon">
                                                                    m
                                                                </span>
                                                            </div>
                                                            <bcs-popover :content="$t('设置CPU limits，1000m CPU=1核 CPU')" placement="top">
                                                                <span class="bk-badge">
                                                                    <i class="bcs-icon bcs-icon-question-circle"></i>
                                                                </span>
                                                            </bcs-popover>
                                                        </div>
                                                    </div>
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 105px;">{{$t('内存')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 105px;">
                                                            <div class="bk-form-input-group mr5">
                                                                <span class="input-group-addon is-left">
                                                                    requests
                                                                </span>
                                                                <bkbcs-input
                                                                    type="number"
                                                                    style="width: 100px;"
                                                                    :min="0"
                                                                    :max="curContainer.resources.limits.memory ? curContainer.resources.limits.memory : Number.MAX_VALUE"
                                                                    :placeholder="$t('请输入')"
                                                                    :value.sync="curContainer.resources.requests.memory"
                                                                    :list="varList"
                                                                >
                                                                </bkbcs-input>
                                                                <span class="input-group-addon">
                                                                    Mi
                                                                </span>
                                                            </div>
                                                            <bcs-popover :content="$t('设置内存requests')" placement="top">
                                                                <span class="bk-badge">
                                                                    <i class="bcs-icon bcs-icon-question-circle"></i>
                                                                </span>
                                                            </bcs-popover>

                                                            <div class="bk-form-input-group ml20 mr5">
                                                                <span class="input-group-addon is-left">
                                                                    limits
                                                                </span>
                                                                <bkbcs-input
                                                                    type="number"
                                                                    style="width: 100px;"
                                                                    :min="0"
                                                                    :placeholder="$t('请输入')"
                                                                    :value.sync="curContainer.resources.limits.memory"
                                                                    :list="varList">
                                                                </bkbcs-input>
                                                                <span class="input-group-addon">
                                                                    Mi
                                                                </span>
                                                            </div>
                                                            <bcs-popover :content="$t('设置内存limits')" placement="top">
                                                                <span class="bk-badge">
                                                                    <i class="bcs-icon bcs-icon-question-circle"></i>
                                                                </span>
                                                            </bcs-popover>
                                                        </div>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="tab5" :title="$t('健康检查')">
                                                <div class="bk-form m20">
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 120px;">{{$t('类型')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 120px">
                                                            <div class="bk-dropdown-box" style="width: 250px;">
                                                                <bk-selector
                                                                    :placeholder="$t('请选择')"
                                                                    :setting-key="'id'"
                                                                    :display-key="'name'"
                                                                    :selected.sync="curContainer.webCache.livenessProbeType"
                                                                    :list="healthCheckTypes">
                                                                </bk-selector>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <div class="bk-form-item" v-show="curContainer.webCache.livenessProbeType && curContainer.webCache.livenessProbeType !== 'EXEC'">
                                                        <label class="bk-label" style="width: 120px;">{{$t('端口名称')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 120px;">
                                                            <div class="bk-dropdown-box" style="width: 250px;">
                                                                <bk-selector
                                                                    :placeholder="$t('请选择')"
                                                                    :setting-key="'name'"
                                                                    :display-key="'name'"
                                                                    :selected="livenessProbePortName"
                                                                    :list="portList"
                                                                    @item-selected="livenessProbePortNameSelect">
                                                                </bk-selector>
                                                            </div>
                                                            <bcs-popover placement="right">
                                                                <i class="bcs-icon bcs-icon-question-circle ml5" style="vertical-align: middle; cursor: pointer;"></i>
                                                                <div slot="content">
                                                                    {{$t('引用端口映射中的端口设置')}}
                                                                </div>
                                                            </bcs-popover>
                                                            <p class="biz-guard-tip bk-default mt5" v-if="!portList.length">{{$t('请先配置完整的端口映射')}}</p>
                                                        </div>
                                                    </div>

                                                    <div class="bk-form-item" v-show="curContainer.webCache.livenessProbeType && (curContainer.webCache.livenessProbeType === 'HTTP')">
                                                        <div class="bk-form-content" style="margin-left: 0">
                                                            <div class="bk-form-inline-item">
                                                                <label class="bk-label" style="width: 120px;">{{$t('请求路径')}}：</label>
                                                                <div class="bk-form-content" style="margin-left: 120px;">
                                                                    <bkbcs-input
                                                                        type="text"
                                                                        style="width: 521px;"
                                                                        :placeholder="$t('例如{path}', { path: '/healthcheck' })"
                                                                        :value.sync="curContainer.livenessProbe.httpGet.path"
                                                                        :list="varList"
                                                                    >
                                                                    </bkbcs-input>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <div class="bk-form-item" v-show="curContainer.webCache.livenessProbeType && curContainer.webCache.livenessProbeType === 'EXEC'">
                                                        <div class="bk-form-content" style="margin-left: 0">
                                                            <div class="bk-form-inline-item">
                                                                <label class="bk-label" style="width: 120px;">{{$t('检查命令')}}：</label>
                                                                <div class="bk-form-content" style="margin-left: 120px;">
                                                                    <bkbcs-input
                                                                        type="text"
                                                                        style="width: 521px;"
                                                                        :placeholder="$t('例如/tmp/check.sh，多个命令用空格分隔')"
                                                                        :value.sync="curContainer.livenessProbe.exec.command"
                                                                        :list="varList"
                                                                    >
                                                                    </bkbcs-input>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <div class="bk-form-item" v-show="curContainer.webCache.livenessProbeType && curContainer.webCache.livenessProbeType === 'HTTP'">
                                                        <div class="bk-form-content" style="margin-left: 0">
                                                            <div class="bk-form-inline-item">
                                                                <label class="bk-label" style="width: 120px;">{{$t('设置Header')}}：</label>
                                                                <div class="bk-form-content" style="margin-left: 120px;">
                                                                    <bk-keyer ref="livenessProbeHeaderKeyer" :key-list.sync="livenessProbeHeaders" :var-list="varList" @change="updateLivenessHeader"></bk-keyer>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <template>
                                                        <bk-button :class="['bk-text-button mt10 f12 mb10', { 'rotate': isPartCShow }]" style="margin-left: 114px;" @click.stop.prevent="togglePartC">
                                                            {{$t('高级设置')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                                                        </bk-button>
                                                        <div v-show="isPartCShow">
                                                            <div class="bk-form-item">
                                                                <div class="bk-form-content" style="margin-left: 0">
                                                                    <div class="bk-form-inline-item">
                                                                        <label class="bk-label" style="width: 120px;">{{$t('初始化超时')}}：</label>
                                                                        <div class="bk-form-content" style="margin-left: 120px;">
                                                                            <div class="bk-form-input-group">
                                                                                <bkbcs-input
                                                                                    type="number"
                                                                                    style="width: 70px;"
                                                                                    :placeholder="$t('请输入')"
                                                                                    :min="1"
                                                                                    :value.sync="curContainer.livenessProbe.initialDelaySeconds"
                                                                                    :list="varList"
                                                                                >
                                                                                </bkbcs-input>
                                                                                <span class="input-group-addon">
                                                                                    {{$t('秒')}}
                                                                                </span>
                                                                            </div>
                                                                        </div>
                                                                    </div>

                                                                    <div class="bk-form-inline-item">
                                                                        <label class="bk-label" style="width: 140px;">{{$t('检查间隔')}}：</label>
                                                                        <div class="bk-form-content" style="margin-left: 140px;">
                                                                            <div class="bk-form-input-group">
                                                                                <bkbcs-input
                                                                                    type="number"
                                                                                    style="width: 70px;"
                                                                                    :placeholder="$t('请输入')"
                                                                                    :min="1"
                                                                                    :value.sync="curContainer.livenessProbe.periodSeconds"
                                                                                    :list="varList"
                                                                                >
                                                                                </bkbcs-input>
                                                                                <span class="input-group-addon">
                                                                                    {{$t('秒')}}
                                                                                </span>
                                                                            </div>
                                                                        </div>
                                                                    </div>
                                                                    <div class="bk-form-inline-item">
                                                                        <label class="bk-label" style="width: 140px;">{{$t('检查超时')}}：</label>
                                                                        <div class="bk-form-content" style="margin-left: 140px;">
                                                                            <div class="bk-form-input-group">
                                                                                <bkbcs-input
                                                                                    type="number"
                                                                                    style="width: 70px;"
                                                                                    :placeholder="$t('请输入')"
                                                                                    :min="1"
                                                                                    :value.sync="curContainer.livenessProbe.timeoutSeconds"
                                                                                    :list="varList"
                                                                                >
                                                                                </bkbcs-input>
                                                                                <span class="input-group-addon">
                                                                                    {{$t('秒')}}
                                                                                </span>
                                                                            </div>
                                                                        </div>
                                                                    </div>
                                                                </div>
                                                            </div>

                                                            <div class="bk-form-item">
                                                                <div class="bk-form-content" style="margin-left: 0">
                                                                    <div class="bk-form-inline-item">
                                                                        <label class="bk-label" style="width: 120px;">{{$t('不健康阈值')}}：</label>
                                                                        <div class="bk-form-content" style="margin-left: 120px;">
                                                                            <div class="bk-form-input-group">
                                                                                <bkbcs-input
                                                                                    type="number"
                                                                                    style="width: 70px;"
                                                                                    :min="1"
                                                                                    :placeholder="$t('请输入')"
                                                                                    :value.sync="curContainer.livenessProbe.failureThreshold"
                                                                                    :list="varList"
                                                                                >
                                                                                </bkbcs-input>
                                                                                <span class="input-group-addon">
                                                                                    {{$t('次失败')}}
                                                                                </span>
                                                                            </div>
                                                                        </div>
                                                                    </div>

                                                                    <div class="bk-form-inline-item">
                                                                        <label class="bk-label" style="width: 140px;">{{$t('健康阈值')}}：</label>
                                                                        <div class="bk-form-content" style="margin-left: 140px;">
                                                                            <div class="bk-form-input-group">
                                                                                <bkbcs-input
                                                                                    type="number"
                                                                                    style="width: 70px;"
                                                                                    :min="1"
                                                                                    :placeholder="$t('请输入')"
                                                                                    :value.sync="curContainer.livenessProbe.successThreshold"
                                                                                    :list="varList"
                                                                                >
                                                                                </bkbcs-input>
                                                                                <span class="input-group-addon">
                                                                                    {{$t('次成功')}}
                                                                                </span>
                                                                            </div>
                                                                        </div>
                                                                    </div>

                                                                </div>
                                                            </div>
                                                        </div>
                                                    </template>

                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="tab5-1" :title="$t('就绪检查')">
                                                <div class="bk-form m20">
                                                    <div class="bk-form-item">
                                                        <label class="bk-label" style="width: 120px;">{{$t('类型')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 120px">
                                                            <div class="bk-dropdown-box" style="width: 250px;">
                                                                <bk-selector
                                                                    :placeholder="$t('请选择')"
                                                                    :setting-key="'id'"
                                                                    :display-key="'name'"
                                                                    :selected.sync="curContainer.webCache.readinessProbeType"
                                                                    :list="healthCheckTypes">
                                                                </bk-selector>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <div class="bk-form-item" v-show="curContainer.webCache.readinessProbeType && curContainer.webCache.readinessProbeType !== 'EXEC'">
                                                        <label class="bk-label" style="width: 120px;">{{$t('端口名称')}}：</label>
                                                        <div class="bk-form-content" style="margin-left: 120px;">
                                                            <div class="bk-dropdown-box" style="width: 250px;">
                                                                <bk-selector
                                                                    :placeholder="$t('请选择')"
                                                                    :setting-key="'name'"
                                                                    :display-key="'name'"
                                                                    :selected="readinessProbePortName"
                                                                    :list="portList"
                                                                    @item-selected="readinessProbePortNameSelect">
                                                                </bk-selector>
                                                            </div>
                                                            <bcs-popover placement="right">
                                                                <i class="bcs-icon bcs-icon-question-circle ml5" style="vertical-align: middle; cursor: pointer;"></i>
                                                                <div slot="content">
                                                                    {{$t('引用端口映射中的端口设置')}}
                                                                </div>
                                                            </bcs-popover>
                                                            <p class="biz-guard-tip bk-default mt5" v-if="!portList.length">{{$t('请先配置完整的端口映射')}}</p>
                                                        </div>
                                                    </div>

                                                    <div class="bk-form-item" v-show="curContainer.webCache.readinessProbeType && (curContainer.webCache.readinessProbeType === 'HTTP')">
                                                        <div class="bk-form-content" style="margin-left: 0">
                                                            <div class="bk-form-inline-item">
                                                                <label class="bk-label" style="width: 120px;">{{$t('请求路径')}}：</label>
                                                                <div class="bk-form-content" style="margin-left: 120px;">
                                                                    <bkbcs-input
                                                                        type="text"
                                                                        style="width: 521px;"
                                                                        :placeholder="$t('例如{path}', { path: '/healthcheck' })"
                                                                        :value.sync="curContainer.readinessProbe.httpGet.path"
                                                                        :list="varList"
                                                                    >
                                                                    </bkbcs-input>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <div class="bk-form-item" v-show="curContainer.webCache.readinessProbeType && curContainer.webCache.readinessProbeType === 'EXEC'">
                                                        <div class="bk-form-content" style="margin-left: 0">
                                                            <div class="bk-form-inline-item">
                                                                <label class="bk-label" style="width: 120px;">{{$t('检查命令')}}：</label>
                                                                <div class="bk-form-content" style="margin-left: 120px;">
                                                                    <bkbcs-input
                                                                        type="text"
                                                                        style="width: 521px;"
                                                                        :placeholder="$t('例如{path}', { path: '/tmp/check.sh' })"
                                                                        :value.sync="curContainer.readinessProbe.exec.command"
                                                                        :list="varList"
                                                                    >
                                                                    </bkbcs-input>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <div class="bk-form-item" v-show="curContainer.webCache.readinessProbeType && curContainer.webCache.readinessProbeType === 'HTTP'">
                                                        <div class="bk-form-content" style="margin-left: 0">
                                                            <div class="bk-form-inline-item">
                                                                <label class="bk-label" style="width: 120px;">{{$t('设置Header')}}：</label>
                                                                <div class="bk-form-content" style="margin-left: 120px;">
                                                                    <bk-keyer ref="readinessProbeHeaderKeyer" :key-list.sync="readinessProbeHeaders" :var-list="varList" @change="updateReadinessHeader"></bk-keyer>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <template>
                                                        <bk-button :class="['bk-text-button mt10 f12 mb10', { 'rotate': isPartCShow }]" style="margin-left: 114px;" @click.stop.prevent="togglePartC">
                                                            {{$t('高级设置')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                                                        </bk-button>
                                                        <div v-show="isPartCShow">
                                                            <div class="bk-form-item">
                                                                <div class="bk-form-content" style="margin-left: 0">
                                                                    <div class="bk-form-inline-item">
                                                                        <label class="bk-label" style="width: 120px;">{{$t('初始化超时')}}：</label>
                                                                        <div class="bk-form-content" style="margin-left: 120px;">
                                                                            <div class="bk-form-input-group">
                                                                                <bkbcs-input
                                                                                    type="number"
                                                                                    style="width: 70px;"
                                                                                    :placeholder="$t('请输入')"
                                                                                    :min="1"
                                                                                    :value.sync="curContainer.readinessProbe.initialDelaySeconds"
                                                                                    :list="varList"
                                                                                >
                                                                                </bkbcs-input>
                                                                                <span class="input-group-addon">
                                                                                    {{$t('秒')}}
                                                                                </span>
                                                                            </div>
                                                                        </div>
                                                                    </div>

                                                                    <div class="bk-form-inline-item">
                                                                        <label class="bk-label" style="width: 140px;">{{$t('检查间隔')}}：</label>
                                                                        <div class="bk-form-content" style="margin-left: 140px;">
                                                                            <div class="bk-form-input-group">
                                                                                <bkbcs-input
                                                                                    type="number"
                                                                                    style="width: 70px;"
                                                                                    :placeholder="$t('请输入')"
                                                                                    :min="1"
                                                                                    :value.sync="curContainer.readinessProbe.periodSeconds"
                                                                                    :list="varList"
                                                                                >
                                                                                </bkbcs-input>
                                                                                <span class="input-group-addon">
                                                                                    {{$t('秒')}}
                                                                                </span>
                                                                            </div>
                                                                        </div>
                                                                    </div>
                                                                    <div class="bk-form-inline-item">
                                                                        <label class="bk-label" style="width: 140px;">{{$t('检查超时')}}：</label>
                                                                        <div class="bk-form-content" style="margin-left: 140px;">
                                                                            <div class="bk-form-input-group">
                                                                                <bkbcs-input
                                                                                    type="number"
                                                                                    style="width: 70px;"
                                                                                    :placeholder="$t('请输入')"
                                                                                    :min="1"
                                                                                    :value.sync="curContainer.readinessProbe.timeoutSeconds"
                                                                                    :list="varList"
                                                                                >
                                                                                </bkbcs-input>
                                                                                <span class="input-group-addon">
                                                                                    {{$t('秒')}}
                                                                                </span>
                                                                            </div>
                                                                        </div>
                                                                    </div>
                                                                </div>
                                                            </div>

                                                            <div class="bk-form-item">
                                                                <div class="bk-form-content" style="margin-left: 0">
                                                                    <div class="bk-form-inline-item">
                                                                        <label class="bk-label" style="width: 120px;">{{$t('不健康阈值')}}：</label>
                                                                        <div class="bk-form-content" style="margin-left: 120px;">
                                                                            <div class="bk-form-input-group">
                                                                                <bkbcs-input
                                                                                    type="number"
                                                                                    style="width: 70px;"
                                                                                    :placeholder="$t('请输入')"
                                                                                    :min="1"
                                                                                    :value.sync="curContainer.readinessProbe.failureThreshold"
                                                                                    :list="varList"
                                                                                >
                                                                                </bkbcs-input>
                                                                                <span class="input-group-addon">
                                                                                    {{$t('次失败')}}
                                                                                </span>
                                                                            </div>
                                                                        </div>
                                                                    </div>

                                                                    <div class="bk-form-inline-item">
                                                                        <label class="bk-label" style="width: 140px;">{{$t('健康阈值')}}：</label>
                                                                        <div class="bk-form-content" style="margin-left: 140px;">
                                                                            <div class="bk-form-input-group">
                                                                                <bkbcs-input
                                                                                    type="number"
                                                                                    style="width: 70px;"
                                                                                    :placeholder="$t('请输入')"
                                                                                    :min="1"
                                                                                    :value.sync="curContainer.readinessProbe.successThreshold"
                                                                                    :list="varList"
                                                                                >
                                                                                </bkbcs-input>
                                                                                <span class="input-group-addon">
                                                                                    {{$t('次成功')}}
                                                                                </span>
                                                                            </div>
                                                                        </div>
                                                                    </div>

                                                                </div>
                                                            </div>
                                                        </div>
                                                    </template>

                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="tab6" :title="$t('非标准日志采集')">
                                                <div class="bk-form m20">
                                                    <div class="bk-form-item">
                                                        <div class="bk-form-content" style="margin-left: 20px">
                                                            <div class="bk-keyer">
                                                                <div class="biz-keys-list mb10">
                                                                    <div class="biz-key-item" v-for="(logItem, index) in curContainer.webCache.logListCache" :key="index">
                                                                        <bkbcs-input
                                                                            type="text"
                                                                            style="width: 360px;"
                                                                            :placeholder="$t('请输入容器中自定义采集的日志绝对路径')"
                                                                            :value.sync="logItem.value"
                                                                            :list="varList"
                                                                        >
                                                                        </bkbcs-input>

                                                                        <bk-button class="action-btn ml5" @click.stop.prevent="addLog">
                                                                            <i class="bcs-icon bcs-icon-plus"></i>
                                                                        </bk-button>
                                                                        <bk-button class="action-btn" v-if="curContainer.webCache.logListCache.length > 1" @click.stop.prevent="removeLog(logItem, index)">
                                                                            <i class="bcs-icon bcs-icon-minus"></i>
                                                                        </bk-button>
                                                                    </div>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>

                                            <bk-tab-panel name="tab7" :title="$t('生命周期')">
                                                <div class="bk-form m20">
                                                    <div class="bk-form-item">
                                                        <div class="bk-form-content" style="margin-left: 20px">
                                                            <div class="bk-form-item">
                                                                <label class="bk-label" style="width: 108px;">{{$t('停止前执行')}}：</label>
                                                                <div class="bk-form-content" style="margin-left: 108px;">
                                                                    <bkbcs-input
                                                                        type="text"
                                                                        :placeholder="$t('多个命令用空格分隔，例如/bin/bash &quot;-c&quot;  &quot;while true; do echo hello; sleep 10;done&quot;')"
                                                                        :value.sync="curContainer.lifecycle.preStop.exec.command"
                                                                        :list="varList"
                                                                    >
                                                                    </bkbcs-input>
                                                                </div>
                                                            </div>
                                                            <div class="bk-form-item">
                                                                <label class="bk-label" style="width: 108px;">{{$t('启动后执行')}}：</label>
                                                                <div class="bk-form-content" style="margin-left: 108px;">
                                                                    <bkbcs-input
                                                                        type="text"
                                                                        :placeholder="$t('多个命令用空格分隔，例如/bin/bash &quot;-c&quot;  &quot;while true; do echo hello; sleep 10;done&quot;')"
                                                                        :value.sync="curContainer.lifecycle.postStart.exec.command"
                                                                        :list="varList"
                                                                    >
                                                                    </bkbcs-input>
                                                                    <!-- <bcs-popover content="多个命令用空格分隔" placement="top">
                                                                        <span class="bk-badge ml5">
                                                                            <i class="bcs-icon bcs-icon-question-circle"></i>
                                                                        </span>
                                                                    </bcs-popover> -->
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>
                                            </bk-tab-panel>
                                        </bk-tab>
                                    </div>

                                </div>
                                <div class="operation-area mt30 mb50" style="margin-left: 105px;">
                                </div>
                            </div>
                        </template>
                    </div>
                </div>
            </div>
        </template>
    </div>
</template>

<script>
    import bkKeyer from '@/components/keyer'
    import ace from '@/components/ace-editor'
    import header from './header.vue'
    import tabs from './tabs.vue'
    import _ from 'lodash'
    import yamljs from 'js-yaml'
    import mixinBase from '@/mixins/configuration/mixin-base'
    import k8sBase from '@/mixins/configuration/k8s-base'

    import applicationParams from '@/json/k8s-statefulset.json'
    import containerParams from '@/json/k8s-container.json'

    export default {
        components: {
            ace,
            'bk-keyer': bkKeyer,
            'biz-header': header,
            'biz-tabs': tabs
        },
        mixins: [mixinBase, k8sBase],
        data () {
            return {
                isTabChanging: false,
                renderVersionIndex: 0,
                renderImageIndex: 0,
                curDesc: '',
                curImageData: {},
                winHeight: 0,
                exceptionCode: null,
                isDataLoading: true,
                isDataSaveing: false,
                isPartAShow: false, // 第一个更多设置
                isPartBShow: false, // 第二个更多设置
                isPartCShow: false, // 第三个更多设置
                imageIndex: -1,
                versionIndex: -1,
                appJsonValidator: null,
                isEditName: false,
                isEditDesc: false,
                appParamKeys: [],
                keyList: [],
                yamlContainerWebcache: [],
                curApplicationLinkLabels: [],
                isLoadingImageList: false,
                isLoadingServices: true,
                toJsonDialogConf: {
                    isShow: false,
                    title: '',
                    timer: null,
                    width: 800,
                    loading: false
                },
                editorConfig: {
                    width: '100%',
                    height: '100%',
                    lang: 'yaml',
                    readOnly: false,
                    fullScreen: false,
                    value: '',
                    editor: null
                },
                yamlEditorConfig: {
                    width: '100%',
                    height: '100%',
                    lang: 'yaml',
                    readOnly: false,
                    fullScreen: false,
                    value: '',
                    editor: null
                },
                volVisitList: [
                    {
                        id: 'ReadWriteOnce',
                        name: 'ReadWriteOnce'
                    },
                    {
                        id: 'ReadOnlyMany',
                        name: 'ReadOnlyMany'
                    },
                    {
                        id: 'ReadWriteMany',
                        name: 'ReadWriteMany'
                    }
                ],
                netList: [
                    {
                        id: 'HOST',
                        name: 'HOST'
                    },
                    {
                        id: 'BRIDGE',
                        name: 'BRIDGE'
                    },
                    {
                        id: 'NONE',
                        name: 'NONE'
                    },
                    {
                        id: 'USER',
                        name: 'USER'
                    },
                    {
                        id: 'CUSTOM',
                        name: this.$t('自定义')
                    }
                ],
                constraintNameList: [
                    {
                        id: 'hostname',
                        name: 'Hostname'
                    },
                    {
                        id: 'InnerIP',
                        name: 'InnerIP'
                    }
                ],
                operatorList: [
                    {
                        id: 'CLUSTER',
                        name: 'CLUSTER'
                    },
                    {
                        id: 'GROUPBY',
                        name: 'GROUPBY'
                    },
                    {
                        id: 'LIKE',
                        name: 'LIKE'
                    },
                    {
                        id: 'UNLIKE',
                        name: 'UNLIKE'
                    },
                    {
                        id: 'UNIQUE',
                        name: 'UNIQUE'
                    },
                    {
                        id: 'MAXPER',
                        name: 'MAXPER'
                    }
                ],
                livenessProbeHeaders: [],
                readinessProbeHeaders: [],
                protocolList: [
                    {
                        id: 'TCP',
                        name: 'TCP'
                    },
                    {
                        id: 'UDP',
                        name: 'UDP'
                    }
                ],
                mountTypeList: [
                    {
                        id: 'custom',
                        name: this.$t('自定义')
                    },
                    {
                        id: 'valueFrom',
                        name: 'ValueFrom'
                    },
                    {
                        id: 'configmapKey',
                        name: this.$t('ConfigMap单键')
                    },
                    {
                        id: 'configmapFile',
                        name: this.$t('ConfigMap文件')
                    },
                    {
                        id: 'secretKey',
                        name: this.$t('Secret单键')
                    },
                    {
                        id: 'secretFile',
                        name: this.$t('Secret文件')
                    }
                ],
                volumeTypeList: [
                    {
                        id: 'emptyDir',
                        name: 'emptyDir'
                    },
                    {
                        id: 'emptyDir(Memory)',
                        name: 'emptyDir(Memory)'
                    },
                    // {
                    //     id: 'persistentVolumeClaim',
                    //     name: 'persistentVolumeClaim'
                    // },
                    {
                        id: 'hostPath',
                        name: 'hostPath'
                    },
                    {
                        id: 'configMap',
                        name: 'configMap'
                    },
                    {
                        id: 'secret',
                        name: 'secret'
                    }
                ],
                metricIndex: [],
                configmapList: [],
                configmapKeyList: [],
                secretKeyList: [],
                secretList: [],
                volumeConfigmapList: [],
                volumeSecretList: [],
                curApplicationId: 0,
                curApplication: applicationParams,
                curContainerIndex: 0,
                curContainer: applicationParams.config.spec.template.spec.allContainers[0],
                isAlwayCheckImage: false,
                editTemplate: {
                    name: '',
                    desc: ''
                },
                imageList: [],
                imageVersionList: [],
                restartPolicy: ['Always', 'OnFailure', 'Never'],
                healthCheckTypes: [
                    {
                        id: 'HTTP',
                        name: 'HTTP'
                    },
                    {
                        id: 'TCP',
                        name: 'TCP'
                    },
                    {
                        id: 'EXEC',
                        name: 'EXEC'
                    }
                ],
                logList: [
                    {
                        value: ''
                    }
                ],
                isMorePanelShow: false,
                isPodPanelShow: false,
                strategy: 'Cluster',
                netStrategyList: [
                    {
                        id: 0,
                        name: 'Cluster'
                    },
                    {
                        id: 1,
                        name: 'Host'
                    }
                ],
                curApplicationCache: null
            }
        },
        computed: {
            volumeConfigmapAllList () {
                const list = [...this.volumeConfigmapList]
                this.existConfigmapList.forEach(item => {
                    list.push({
                        id: `${item.name}:${item.cluster_id}:${item.namespace}`,
                        name: `${item.name} (${item.cluster_name}-${item.namespace})`,
                        cluster_name: item.cluster_name,
                        cluster_id: item.cluster_id,
                        namespace: item.namespace
                    })
                })
                return list
            },
            existConfigmapList () {
                return this.$store.state.k8sTemplate.existConfigmapList
            },
            isSelectorChange () {
                const curSelector = {}
                const list = this.curApplication.config.webCache.labelListCache
                list.forEach(item => {
                    if (item.isSelector) {
                        curSelector[item.key] = item.value
                    }
                })

                if (!this.curApplication.cache) {
                    return false
                }
                if (JSON.stringify(this.curApplication.cache.config.spec.selector.matchLabels) === '{}') {
                    return false
                }

                const selectorCache = this.curApplication.cache.config.spec.selector.matchLabels
                if (JSON.stringify(selectorCache) !== JSON.stringify(curSelector)) {
                    return true
                } else {
                    return false
                }
            },
            varList () {
                const list = this.$store.state.variable.varList
                list.forEach(item => {
                    item._id = item.key
                    item._name = item.key
                })
                return list
            },
            dnsStrategyList () {
                const netType = this.curApplication.config.spec.template.spec.hostNetwork
                let list

                if (netType === 0) {
                    list = [
                        {
                            id: 'ClusterFirst',
                            name: 'ClusterFirst'
                        },
                        {
                            id: 'Default',
                            name: 'Default'
                        },
                        {
                            id: 'None',
                            name: 'None'
                        }
                    ]
                } else {
                    list = [
                        {
                            id: 'ClusterFirstWithHostNet',
                            name: 'ClusterFirstWithHostNet'
                        },
                        {
                            id: 'Default',
                            name: 'Default'
                        },
                        {
                            id: 'None',
                            name: 'None'
                        }
                    ]
                }
                return list
            },
            serviceList () {
                return this.$store.state.k8sTemplate.linkServices
            },
            linkServiceName () {
                if (this.curApplication.service_tag) {
                    const service = this.serviceList.find(item => {
                        return item.service_tag === this.curApplication.service_tag
                    })
                    return service ? service.service_name : ''
                } else {
                    return ''
                }
            },
            curMountVolumes () {
                const results = this.curApplication.config.webCache.volumes.filter(item => {
                    return item.name
                })
                return results
            },
            metricList () {
                return this.$store.state.k8sTemplate.metricList
            },
            versionList () {
                const list = this.$store.state.k8sTemplate.versionList
                return list
            },
            isTemplateSaving () {
                return this.$store.state.k8sTemplate.isTemplateSaving
            },
            curTemplate () {
                return this.$store.state.k8sTemplate.curTemplate
            },
            deployments () {
                return this.$store.state.k8sTemplate.deployments
            },
            services () {
                return this.$store.state.k8sTemplate.services
            },
            configmaps () {
                return this.$store.state.k8sTemplate.configmaps
            },
            secrets () {
                return this.$store.state.k8sTemplate.secrets
            },
            daemonsets () {
                return this.$store.state.k8sTemplate.daemonsets
            },
            jobs () {
                return this.$store.state.k8sTemplate.jobs
            },
            statefulsets () {
                return this.$store.state.k8sTemplate.statefulsets
            },
            livenessProbePortName () {
                const healthParams = this.curContainer.livenessProbe
                const type = this.curContainer.webCache.livenessProbeType
                if (type === 'HTTP') {
                    return healthParams.httpGet.port
                } else if (type === 'TCP') {
                    return healthParams.tcpSocket.port
                } else {
                    return ''
                }
            },
            readinessProbePortName () {
                const healthParams = this.curContainer.readinessProbe
                const type = this.curContainer.webCache.readinessProbeType
                if (type === 'HTTP') {
                    return healthParams.httpGet.port
                } else if (type === 'TCP') {
                    return healthParams.tcpSocket.port
                } else {
                    return ''
                }
            },
            curVersion () {
                return this.$store.state.k8sTemplate.curVersion
            },
            templateId () {
                return this.$route.params.templateId
            },
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            portList () {
                let results = []
                const ports = this.curContainer.ports

                if (ports && ports.length) {
                    results = ports.filter(port => {
                        return port.name && port.containerPort
                    })
                    return results
                } else {
                    return []
                }
            },
            curConstraintLabelList () {
                const keyList = []
                const nodes = this.curApplication.config.spec.template.spec.nodeSelector
                // 如果有缓存直接使用
                if (this.curApplication.config.webCache && this.curApplication.config.webCache.nodeSelectorList) {
                    return this.curApplication.config.webCache.nodeSelectorList
                }
                for (const [key, value] of Object.entries(nodes)) {
                    keyList.push({
                        key: key,
                        value: value
                    })
                }
                if (!keyList.length) {
                    keyList.push({
                        key: '',
                        value: ''
                    })
                }
                return keyList
            },
            curLabelList () {
                const keyList = []
                const labels = this.curApplication.config.spec.template.metadata.labels
                const selector = this.curApplication.config.spec.selector.matchLabels
                const linkLabels = this.curApplicationLinkLabels

                // 如果有缓存直接使用
                if (this.curApplication.config.webCache && this.curApplication.config.webCache.labelListCache) {
                    this.curApplication.config.webCache.labelListCache.forEach(item => {
                        const params = {
                            key: item.key,
                            value: item.value,
                            isSelector: item.isSelector,
                            disabled: item.disabled
                        }
                        keyList.push(params)
                    })
                    for (const params of keyList) {
                        const key = params.key
                        const value = params.value
                        for (const label of linkLabels) {
                            if (label.key === key && label.value === value) {
                                params.disabled = true
                                params.linkMessage = label.linkMessage
                            }
                        }
                    }
                    return keyList
                } else {
                    for (const [key, value] of Object.entries(labels)) {
                        const params = {
                            key: key,
                            value: value,
                            isSelector: false,
                            disabled: false,
                            linkMessage: ''
                        }
                        keyList.push(params)
                    }
                }
                for (const params of keyList) {
                    const key = params.key
                    const value = params.value

                    for (const [mKey, mValue] of Object.entries(selector)) {
                        if (mKey === key && mValue === value) {
                            params.isSelector = true
                        }
                    }
                    for (const label of linkLabels) {
                        if (label.key === key && label.value === value) {
                            params.disabled = true
                            params.linkMessage = label.linkMessage
                        }
                    }
                }
                if (!keyList.length) {
                    keyList.push({
                        key: '',
                        value: '',
                        isSelector: false,
                        disabled: false
                    })
                }
                return keyList
            },
            curLogLabelList () {
                const keyList = []
                const labels = this.curApplication.config.customLogLabel
                // 如果有缓存直接使用
                if (this.curApplication.config.webCache && this.curApplication.config.webCache.logLabelListCache) {
                    return this.curApplication.config.webCache.logLabelListCache
                }
                for (const [key, value] of Object.entries(labels)) {
                    keyList.push({
                        key: key,
                        value: value
                    })
                }
                if (!keyList.length) {
                    keyList.push({
                        key: '',
                        value: ''
                    })
                }
                return keyList
            },
            curImageSecretList () {
                const list = []
                if (this.curApplication.config.spec.template.spec.imagePullSecrets) {
                    const secrets = this.curApplication.config.spec.template.spec.imagePullSecrets
                    secrets.forEach(item => {
                        list.push({
                            key: 'name',
                            value: item.name
                        })
                    })
                }
                if (!list.length) {
                    list.push({
                        key: 'name',
                        value: ''
                    })
                }
                return list
            },
            curRemarkList () {
                const list = []
                // 如果有缓存直接使用
                if (this.curApplication.config.webCache && this.curApplication.config.webCache.remarkListCache) {
                    return this.curApplication.config.webCache.remarkListCache
                }
                const annotations = this.curApplication.config.spec.template.metadata.annotations
                for (const [key, value] of Object.entries(annotations)) {
                    list.push({
                        key: key,
                        value: value
                    })
                }
                if (!list.length) {
                    list.push({
                        key: '',
                        value: ''
                    })
                }
                return list
            },
            curEnvList () {
                const list = []
                const envs = this.curContainer.env
                envs.forEach(env => {
                    for (const [key, value] of Object.entries(env)) {
                        list.push({
                            key: key,
                            value: value
                        })
                    }
                })
                return list
            }
        },
        watch: {
            'curApplication.config.metadata.name' (val) {
                console.log('va', val)
                this.curApplication.config.webCache.labelListCache.forEach(item => {
                    if (item.key === 'k8s-app') {
                        item.value = val
                    }
                })
            },
            'services' () {
                if (this.curVersion) {
                    this.initServices(this.curVersion)
                }
            },
            'curContainer' () {
                if (this.curContainer.imagePullPolicy === 'Always') {
                    this.isAlwayCheckImage = true
                } else {
                    this.isAlwayCheckImage = false
                }

                if (!this.curContainer.ports.length) {
                    this.addPort()
                }
                // else {
                //     this.curContainer.ports.forEach(item => {
                //         const projectId = this.projectId
                //         const version = this.curVersion
                //         const portId = item.id
                //         if (portId) {
                //             this.$store.dispatch('k8sTemplate/checkPortIsLink', { projectId, version, portId }).then(res => {
                //                 item.isLink = ''
                //             }, res => {
                //                 const message = res.message || res.data.data || ''
                //                 const msg = message.split(',')[0]
                //                 if (msg) {
                //                     item.isLink = msg + this.$t('，不能修改协议')
                //                 } else {
                //                     item.isLink = ''
                //                 }
                //             })
                //         } else {
                //             item.isLink = ''
                //         }
                //     })
                // }

                // if (!this.curContainer.volumeMounts.length) {
                //     const volumes = this.curContainer.volumeMounts
                //     volumes.push({
                //         'name': '',
                //         'mountPath': '',
                //         'subPath': '',
                //         'readOnly': false
                //     })
                // }
            },
            'curApplication' () {
                this.curContainerIndex = 0
                const container = this.curApplication.config.spec.template.spec.allContainers[0]
                this.setCurContainer(container, 0)
            },
            'curVersion' (val) {
                this.initVolumeConfigmaps()
                this.initVloumeSelectets()
            }
        },
        // async beforeRouteLeave (to, form, next) {
        //     // 修改模板集信息
        //     await this.$refs.commonHeader.saveTemplate()
        //     next()
        // },
        mounted () {
            this.isDataLoading = true
            this.$refs.commonHeader.initTemplate((data) => {
                this.initResource(data)
                this.isDataLoading = false
            })
            this.winHeight = window.innerHeight
            this.initImageList()
            this.initVolumeConfigmaps()
            this.initVloumeSelectets()
            // this.initMetricList()

            const Validator = require('jsonschema').Validator
            this.appJsonValidator = new Validator()
        },
        methods: {
            updateEnvItem (index, data, env) {
                env.keyCache = data.keyCache
                env.nameCache = data.nameCache
            },
            reloadServices () {
                if (this.curVersion) {
                    this.isLoadingServices = true
                    this.initServices(this.curVersion)
                }
            },
            initServices (version) {
                const projectId = this.projectId
                this.linkAppVersion = version
                this.$store.dispatch('k8sTemplate/getServicesByVersion', { projectId, version }).then(res => {
                    this.isLoadingServices = false
                }, res => {
                    const message = res.message
                    this.$bkMessage({
                        theme: 'error',
                        message: message,
                        hasCloseIcon: true,
                        delay: '10000'
                    })
                })
            },
            removeVolTpl (item, index) {
                const volumes = this.curApplication.config.spec.volumeClaimTemplates
                volumes.splice(index, 1)
            },
            addVolTpl () {
                const volumes = this.curApplication.config.spec.volumeClaimTemplates
                volumes.push({
                    metadata: {
                        name: ''
                    },
                    spec: {
                        accessModes: [],
                        storageClassName: '',
                        resources: {
                            requests: {
                                storage: 1
                            }
                        }

                    }
                })
            },
            updateLivenessHeader (list, data) {
                const result = []
                list.forEach(item => {
                    const params = {
                        name: item.key,
                        value: item.value
                    }
                    result.push(params)
                })
                this.curContainer.livenessProbe.httpGet.httpHeaders = result
            },
            updateReadinessHeader (list, data) {
                const result = []
                list.forEach(item => {
                    const params = {
                        name: item.key,
                        value: item.value
                    }
                    result.push(params)
                })
                this.curContainer.readinessProbe.httpGet.httpHeaders = result
            },
            changeDNSPolicy (item, data) {
                if (item === 0) {
                    this.curApplication.config.spec.template.spec.dnsPolicy = 'ClusterFirst'
                } else {
                    this.curApplication.config.spec.template.spec.dnsPolicy = 'ClusterFirstWithHostNet'
                }
            },
            getAppParamsKeys (obj, result) {
                for (const key in obj) {
                    if (key === 'nodeSelector') continue
                    if (key === 'annotations') continue
                    if (key === 'volumes') continue
                    if (key === 'affinity') continue
                    if (key === 'labels') continue
                    if (key === 'selector') continue
                    if (Object.prototype.toString.call(obj) === '[object Array]') {
                        this.getAppParamsKeys(obj[key], result)
                    } else if (Object.prototype.toString.call(obj) === '[object Object]') {
                        if (!result.includes(key)) {
                            result.push(key)
                        }
                        this.getAppParamsKeys(obj[key], result)
                    }
                }
            },
            checkJson (jsonObj) {
                const editor = this.editorConfig.editor
                const appParams = applicationParams.config
                const appParamKeys = [
                    'id',
                    'containerPort',
                    'hostPort',
                    'name',
                    'protocol',
                    'isLink',
                    'isDisabled',
                    'env',
                    'secrets',
                    'configmaps',
                    'logPathList',
                    'valueFrom',
                    'configMapRef',
                    'configMapKeyRef',
                    'secretRef',
                    'secretKeyRef',
                    'serviceAccountName',
                    'fieldRef',
                    'fieldPath',
                    'envFrom'
                ]
                const jsonParamKeys = []
                this.getAppParamsKeys(appParams, appParamKeys)
                this.getAppParamsKeys(jsonObj, jsonParamKeys)
                // application查看无效字段
                for (const key of jsonParamKeys) {
                    if (!appParamKeys.includes(key)) {
                        this.$bkMessage({
                            theme: 'error',
                            message: `${key}${this.$t('为无效字段')}`
                        })
                        const match = editor.find(`${key}`)
                        if (match) {
                            editor.moveCursorTo(match.end.row, match.end.column)
                        }
                        return false
                    }
                }

                return true
            },
            formatJson (jsonObj) {
                // 标签
                const keyList = []
                const labels = jsonObj.spec.template.metadata.labels
                const selector = jsonObj.spec.selector.matchLabels
                const linkLabels = this.curApplicationLinkLabels
                const hostAliases = jsonObj.spec.template.spec.hostAliases

                for (const [key, value] of Object.entries(labels)) {
                    const params = {
                        key: key,
                        value: value,
                        isSelector: false,
                        disabled: false
                    }
                    keyList.push(params)
                }
                for (const params of keyList) {
                    const key = params.key
                    const value = params.value

                    for (const [mKey, mValue] of Object.entries(selector)) {
                        if (mKey === key && mValue === value) {
                            params.isSelector = true
                        }
                    }
                    for (const label of linkLabels) {
                        if (label.key === key && label.value === value) {
                            params.disabled = true
                        }
                    }
                }
                if (!keyList.length) {
                    keyList.push({
                        key: '',
                        value: '',
                        isSelector: false,
                        disabled: false
                    })
                }
                jsonObj.webCache.labelListCache = keyList

                // hostAliases
                const hostAliasesCache = []
                if (hostAliases) {
                    for (const hostAlias of hostAliases) {
                        hostAliasesCache.push({
                            ip: hostAlias.ip,
                            hostnames: hostAlias.hostnames.join(';')
                        })
                    }
                    jsonObj.webCache.hostAliasesCache = hostAliasesCache
                }

                // 日志标签
                const logLabels = jsonObj.customLogLabel
                const logLabelList = []
                for (const [key, value] of Object.entries(logLabels)) {
                    logLabelList.push({
                        key: key,
                        value: value
                    })
                }
                if (!logLabelList.length) {
                    logLabelList.push({
                        key: '',
                        value: ''
                    })
                }
                jsonObj.webCache.logLabelListCache = logLabelList

                // 注解
                const remarkList = []
                const annotations = jsonObj.spec.template.metadata.annotations
                for (const [key, value] of Object.entries(annotations)) {
                    remarkList.push({
                        key: key,
                        value: value
                    })
                }
                if (!remarkList.length) {
                    remarkList.push({
                        key: '',
                        value: ''
                    })
                }

                jsonObj.webCache.remarkListCache = remarkList

                // 亲和性约束
                const affinity = jsonObj.spec.template.spec.affinity
                if (affinity && JSON.stringify(affinity) !== '{}') {
                    const yamlStr = yamljs.dump(jsonObj.spec.template.spec.affinity, { indent: 2 })
                    jsonObj.webCache.affinityYaml = yamlStr
                    jsonObj.webCache.isUserConstraint = true
                } else {
                    jsonObj.spec.template.spec.affinity = {}
                    jsonObj.webCache.affinityYaml = ''
                    jsonObj.webCache.isUserConstraint = false
                }

                // 调度约束
                const nodeSelector = jsonObj.spec.template.spec.nodeSelector
                const nodeSelectorList = jsonObj.webCache.nodeSelectorList = []
                for (const [key, value] of Object.entries(nodeSelector)) {
                    nodeSelectorList.push({
                        key: key,
                        value: value
                    })
                }
                if (!nodeSelectorList.length) {
                    nodeSelectorList.push({
                        key: '',
                        value: ''
                    })
                }

                // Metric信息 (合并原数据)
                jsonObj.webCache.isMetric = this.curApplicationCache.webCache.isMetric
                jsonObj.webCache.metricIdList = this.curApplicationCache.webCache.metricIdList

                // 挂载卷
                const volumes = jsonObj.spec.template.spec.volumes
                let volumesCache = jsonObj.webCache.volumes

                if (volumes && volumes.length) {
                    volumesCache = []
                    volumes.forEach(volume => {
                        if (volume.hasOwnProperty('emptyDir')) {
                            if (volume.emptyDir.medium) {
                                volumesCache.push({
                                    type: 'emptyDir(Memory)',
                                    name: volume.name,
                                    source: volume.emptyDir.sizeLimit.replace('Gi', '')
                                })
                            } else {
                                volumesCache.push({
                                    type: 'emptyDir',
                                    name: volume.name,
                                    source: ''
                                })
                            }
                        } else if (volume.hasOwnProperty('persistentVolumeClaim')) {
                            volumesCache.push({
                                type: 'persistentVolumeClaim',
                                name: volume.name,
                                source: volume.persistentVolumeClaim.claimName
                            })
                        } else if (volume.hasOwnProperty('hostPath')) {
                            const volumeItem = {
                                type: 'hostPath',
                                name: volume.name,
                                source: volume.hostPath.path
                            }
                            if (volume.hostPath.type) {
                                volumeItem.hostType = volume.hostPath.type
                            }
                            volumesCache.push(volumeItem)
                        } else if (volume.hasOwnProperty('configMap')) {
                            volumesCache.push({
                                type: 'configMap',
                                name: volume.name,
                                source: volume.configMap.name
                            })
                        } else if (volume.hasOwnProperty('secret')) {
                            volumesCache.push({
                                type: 'secret',
                                name: volume.name,
                                source: volume.secret.secretName
                            })
                        }
                    })
                }
                jsonObj.webCache.volumes = JSON.parse(JSON.stringify(volumesCache))
                // container env
                jsonObj.spec.template.spec.allContainers = []
                const containers = jsonObj.spec.template.spec.containers
                const initContainers = jsonObj.spec.template.spec.initContainers
                containers.forEach((container, index) => {
                    this.formatContainerJosn(container, index)
                    container.webCache.containerType = 'container'
                    jsonObj.spec.template.spec.allContainers.push(container)
                })
                initContainers.forEach((container, index) => {
                    this.formatContainerJosn(container, index)
                    container.webCache.containerType = 'initContainer'
                    jsonObj.spec.template.spec.allContainers.push(container)
                })
                return jsonObj
            },
            formatContainerJosn (container, index) {
                if (!container.webCache) {
                    container.webCache = {}
                }

                // 兼容原数据webcache
                container.webCache.imageName = container.imageName
                delete container.imageName
                container.webCache.env_list = []

                this.curApplicationCache.spec.template.spec.allContainers.forEach(containerCache => {
                    if (containerCache.name === container.name) {
                        // 描述
                        container.webCache.desc = containerCache.webCache.desc
                        // 合并非标准日志采集
                        container.webCache.logListCache = containerCache.webCache.logListCache
                    }
                })

                // 环境变量
                if ((container.env && container.env.length) || (container.envFrom && container.envFrom.length)) {
                    const envs = container.env || []
                    const envFroms = container.envFrom || []
                    envs.forEach(item => {
                        // valuefrom
                        if (item.valueFrom && item.valueFrom.fieldRef) {
                            container.webCache.env_list.push({
                                type: 'valueFrom',
                                key: item.name,
                                value: item.valueFrom.fieldRef.fieldPath
                            })
                            return false
                        }

                        // configMap单键
                        if (item.valueFrom && item.valueFrom.configMapKeyRef) {
                            container.webCache.env_list.push({
                                type: 'configmapKey',
                                key: item.name,
                                nameCache: item.valueFrom.configMapKeyRef.name,
                                keyCache: item.valueFrom.configMapKeyRef.key,
                                value: `${item.valueFrom.configMapKeyRef.name}.${item.valueFrom.configMapKeyRef.key}`
                            })
                            return false
                        }

                        // secret单键
                        if (item.valueFrom && item.valueFrom.secretKeyRef) {
                            container.webCache.env_list.push({
                                type: 'secretKey',
                                key: item.name,
                                nameCache: item.valueFrom.secretKeyRef.name,
                                keyCache: item.valueFrom.secretKeyRef.key,
                                value: `${item.valueFrom.secretKeyRef.name}.${item.valueFrom.secretKeyRef.key}`
                            })
                            return false
                        }

                        // 自定义
                        container.webCache.env_list.push({
                            type: 'custom',
                            key: item.name,
                            value: item.value
                        })
                    })

                    envFroms.forEach(item => {
                        // configMap文件
                        if (item.configMapRef) {
                            container.webCache.env_list.push({
                                type: 'configmapFile',
                                key: '',
                                value: item.configMapRef.name
                            })
                            return false
                        }

                        // secret文件
                        if (item.secretRef) {
                            container.webCache.env_list.push({
                                type: 'secretFile',
                                key: '',
                                value: item.secretRef.name
                            })
                            return false
                        }
                    })
                }

                if (!container.webCache.env_list.length) {
                    container.webCache.env_list.push({
                        type: 'custom',
                        key: '',
                        value: ''
                    })
                }

                // volumeMounts
                if (container.volumeMounts.length) {
                    container.volumeMounts.forEach(volume => {
                        volume.readOnly = false
                    })
                }

                // 镜像自定义
                container.webCache.isImageCustomed = !container.image.startsWith(`${DEVOPS_ARTIFACTORY_HOST}`)

                // volumeMounts
                if (Array.isArray(container.args)) {
                    container.args = container.args.join(' ')
                }

                // 端口
                if (container.ports) {
                    const ports = container.ports
                    ports.forEach((item, index) => {
                        item.isLink = false
                        if (!item.id) {
                            item.id = +new Date() + index
                        }
                    })
                }

                // 资源限制
                const resources = container.resources
                if (resources.limits.cpu && resources.limits.cpu.replace) {
                    resources.limits.cpu = Number(resources.limits.cpu.replace('m', ''))
                }
                if (resources.limits.memory && resources.limits.memory.replace) {
                    resources.limits.memory = Number(resources.limits.memory.replace('Mi', ''))
                }
                if (resources.requests.cpu && resources.requests.cpu.replace) {
                    resources.requests.cpu = Number(resources.requests.cpu.replace('m', ''))
                }
                if (resources.requests.memory && resources.requests.memory.replace) {
                    resources.requests.memory = Number(resources.requests.memory.replace('Mi', ''))
                }
            },
            hideApplicationJson () {
                this.toJsonDialogConf.isShow = false
            },
            saveApplicationJson () {
                const editor = this.editorConfig.editor
                const yaml = editor.getValue()
                const cParams = containerParams
                let appObj = null
                if (!yaml) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入YAML')
                    })
                    return false
                }

                try {
                    appObj = yamljs.load(yaml)
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

                if (appObj.spec.template.spec.containers) {
                    const containers = appObj.spec.template.spec.containers
                    const containerCopys = []
                    containers.forEach(container => {
                        const copy = _.merge({}, cParams, container)
                        containerCopys.push(copy)
                    })
                    containers.splice(0, containers.length, ...containerCopys)
                }

                if (appObj.spec.template.spec.initContainers) {
                    const initContainers = appObj.spec.template.spec.initContainers
                    const containerCopys = []
                    initContainers.forEach(container => {
                        const copy = _.merge({}, cParams, container)
                        containerCopys.push(copy)
                    })
                    initContainers.splice(0, initContainers.length, ...containerCopys)
                }

                const newConfObj = _.merge({}, applicationParams.config, appObj)
                const jsonFromat = this.formatJson(newConfObj)
                this.curApplication.config = jsonFromat
                this.curApplication.desc = this.curDesc
                this.toJsonDialogConf.isShow = false
                if (this.curApplication.config.spec.template.spec.allContainers.length) {
                    const container = this.curApplication.config.spec.template.spec.allContainers[0]
                    this.setCurContainer(container, 0)
                }
            },
            showJsonPanel () {
                this.toJsonDialogConf.title = this.curApplication.config.metadata.name + '.yaml'
                const appConfig = JSON.parse(JSON.stringify(this.curApplication.config))
                const webCache = appConfig.webCache

                this.curDesc = this.curApplication.desc
                // 在处理yaml导入时，保存一份原数据，方便对导入的数据进行合并处理
                this.curApplicationCache = JSON.parse(JSON.stringify(this.curApplication.config))

                // 标签
                if (webCache && webCache.labelListCache) {
                    const labelKeyList = this.tranListToObject(webCache.labelListCache)
                    appConfig.spec.template.metadata.labels = labelKeyList
                    appConfig.spec.selector.matchLabels = {}
                    webCache.labelListCache.forEach(item => {
                        if (item.isSelector && item.key && item.value) {
                            appConfig.spec.selector.matchLabels[item.key] = item.value
                        }
                    })
                }

                // 日志标签
                if (webCache && webCache.logLabelListCache) {
                    const labelKeyList = this.tranListToObject(webCache.logLabelListCache)
                    appConfig.customLogLabel = labelKeyList
                }

                // HostAliases
                if (webCache && webCache.hostAliasesCache) {
                    appConfig.spec.template.spec.hostAliases = []
                    webCache.hostAliasesCache.forEach(item => {
                        appConfig.spec.template.spec.hostAliases.push({
                            ip: item.ip,
                            hostnames: item.hostnames.replace(/ /g, '').split(';')
                        })
                    })
                }

                // 注解
                if (webCache && webCache.remarkListCache) {
                    const remarkKeyList = this.tranListToObject(webCache.remarkListCache)
                    appConfig.spec.template.metadata.annotations = remarkKeyList
                }

                // 调度约束
                if (webCache.nodeSelectorList) {
                    const nodeSelector = appConfig.spec.template.spec.nodeSelector = {}
                    const nodeSelectorList = webCache.nodeSelectorList
                    nodeSelectorList.forEach(item => {
                        nodeSelector[item.key] = item.value
                    })
                }

                // 亲和性约束
                if (webCache.isUserConstraint) {
                    try {
                        const yamlCode = webCache.affinityYamlCache || webCache.affinityYaml
                        webCache.affinityYaml = yamlCode
                        const json = yamljs.load(yamlCode)
                        if (json) {
                            appConfig.spec.template.spec.affinity = json
                        } else {
                            appConfig.spec.template.spec.affinity = {}
                        }
                    } catch (err) {
                        // error
                    }
                } else {
                    appConfig.spec.template.spec.affinity = {}
                }

                if (webCache && webCache.volumes) {
                    const cacheColumes = webCache.volumes
                    const volumes = []
                    cacheColumes.forEach(volume => {
                        if ((volume.name && volume.source) || (volume.name && volume.type === 'emptyDir')) {
                            switch (volume.type) {
                                case 'emptyDir':
                                    volumes.push({
                                        name: volume.name,
                                        emptyDir: {}
                                    })
                                    break

                                case 'persistentVolumeClaim':
                                    volumes.push({
                                        name: volume.name,
                                        persistentVolumeClaim: {
                                            claimName: volume.source
                                        }
                                    })
                                    break

                                case 'hostPath':
                                    const item = {
                                        name: volume.name,
                                        hostPath: {
                                            path: volume.source
                                        }
                                    }
                                    if (volume.hostType) {
                                        item.hostPath.type = volume.hostType
                                    }
                                    volumes.push(item)
                                    break

                                case 'configMap':
                                    // 针对已经存的configmap处理
                                    let volumeSource = volume.source
                                    if (volume.is_exist) {
                                        volumeSource = volume.source.split(':')[0]
                                    }
                                    volumes.push({
                                        name: volume.name,
                                        configMap: {
                                            name: volumeSource
                                        }
                                    })
                                    break

                                case 'secret':
                                    volumes.push({
                                        name: volume.name,
                                        secret: {
                                            secretName: volume.source
                                        }
                                    })
                                    break

                                case 'emptyDir(Memory)':
                                    volumes.push({
                                        name: volume.name,
                                        emptyDir: {
                                            medium: 'Memory',
                                            sizeLimit: `${volume.source}Gi`
                                        }
                                    })
                                    break
                            }
                        }
                    })

                    appConfig.spec.template.spec.volumes = volumes
                }
                delete appConfig.webCache

                // container
                appConfig.spec.template.spec.containers = []
                appConfig.spec.template.spec.initContainers = []

                appConfig.spec.template.spec.allContainers.forEach(container => {
                    container.imageName = container.webCache.imageName
                    this.yamlContainerWebcache.push(JSON.parse(JSON.stringify(container.webCache)))

                    container.env = []
                    container.envFrom = []
                    // 环境变量
                    if (container.webCache && container.webCache.env_list) {
                        const envs = container.webCache.env_list
                        envs.forEach(env => {
                            // valuefrom
                            if (env.type === 'valueFrom') {
                                container.env.push({
                                    name: env.key,
                                    valueFrom: {
                                        fieldRef: {
                                            fieldPath: env.value
                                        }
                                    }
                                })
                                return false
                            }

                            // configMap单键
                            if (env.type === 'configmapKey') {
                                container.env.push({
                                    name: env.key,
                                    valueFrom: {
                                        configMapKeyRef: {
                                            name: env.nameCache,
                                            key: env.keyCache
                                        }
                                    }
                                })
                                return false
                            }

                            // configMap文件
                            if (env.type === 'configmapFile') {
                                container.envFrom.push({
                                    configMapRef: {
                                        name: env.value
                                    }
                                })
                                return false
                            }

                            // secret单键
                            if (env.type === 'secretKey') {
                                container.env.push({
                                    name: env.key,
                                    valueFrom: {
                                        secretKeyRef: {
                                            name: env.nameCache,
                                            key: env.keyCache
                                        }
                                    }
                                })
                                return false
                            }

                            // secret文件
                            if (env.type === 'secretFile') {
                                container.envFrom.push({
                                    secretRef: {
                                        name: env.value
                                    }
                                })
                                return false
                            }

                            // 自定义
                            if (env.key) {
                                container.env.push({
                                    name: env.key,
                                    value: env.value
                                })
                            }
                        })
                    }

                    if (!container.webCache.env_list.length) {
                        container.webCache.env_list.push({
                            type: 'custom',
                            key: '',
                            value: ''
                        })
                    }
                    if (container.webCache.containerType === 'initContainer') {
                        appConfig.spec.template.spec.initContainers.push(container)
                    } else {
                        appConfig.spec.template.spec.containers.push(container)
                    }
                    delete container.webCache
                })
                delete appConfig.spec.template.spec.allContainers

                const yamlStr = yamljs.dump(appConfig, { indent: 2 })
                this.editorConfig.value = yamlStr
                this.toJsonDialogConf.isShow = true
            },
            editorInitAfter (editor) {
                this.editorConfig.editor = editor
                this.editorConfig.editor.setStyle('biz-app-container-tojson-ace')
            },
            yamlEditorInitAfter (editor) {
                this.yamlEditorConfig.editor = editor
                if (this.curApplication.config.webCache.affinityYaml) {
                    editor.setValue(this.curApplication.config.webCache.affinityYaml)
                }
            },
            yamlEditorInput (val) {
                this.curApplication.config.webCache.affinityYamlCache = val
            },
            yamlEditorBlur (val) {
                this.curApplication.config.webCache.affinityYaml = val
            },
            setFullScreen () {
                this.editorConfig.fullScreen = !this.editorConfig.fullScreen
            },
            cancelFullScreen () {
                this.editorConfig.fullScreen = false
            },
            closeToJson () {
                this.toJsonDialogConf.isShow = false
                this.toJsonDialogConf.title = ''
                this.editorConfig.value = ''
                this.copyContent = ''
            },
            initResource (data) {
                const version = data.latest_version_id || data.version
                if (data.statefulsets && data.statefulsets.length) {
                    this.setCurApplication(data.statefulsets[0], 0)
                } else if (data.statefulset && data.statefulset.length) {
                    this.setCurApplication(data.statefulset[0], 0)
                }
                if (version) {
                    this.initServices(version)
                } else {
                    this.isLoadingServices = false
                }
            },
            exportToYaml (data) {
                this.$router.push({
                    name: 'K8sYamlTemplateset',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        templateId: 0
                    },
                    query: {
                        action: 'export'
                    }
                })
            },
            async tabResource (type, target) {
                this.isTabChanging = true
                await this.$refs.commonHeader.saveTemplate()
                await this.$refs.commonHeader.autoSaveResource(type)
                this.$refs.commonTab.goResource(target)
            },
            exceptionHandler (exceptionCode) {
                this.isDataLoading = false
                this.exceptionCode = exceptionCode
            },
            livenessProbePortNameSelect (selected, data) {
                const healthParams = this.curContainer.livenessProbe
                const type = this.curContainer.webCache.livenessProbeType
                if (type === 'HTTP') {
                    healthParams.httpGet.port = selected
                } else if (type === 'TCP') {
                    healthParams.tcpSocket.port = selected
                }
            },
            readinessProbePortNameSelect (selected, data) {
                const healthParams = this.curContainer.readinessProbe
                const type = this.curContainer.webCache.readinessProbeType
                if (type === 'HTTP') {
                    healthParams.httpGet.port = selected
                } else if (type === 'TCP') {
                    healthParams.tcpSocket.port = selected
                }
            },
            toggleRouter (target) {
                this.$router.push({
                    name: target,
                    params: {
                        projectId: this.projectId,
                        templateId: this.templateId
                    }
                })
            },
            changeImagePullPolicy () {
                // 判断改变前的状态
                if (!this.isAlwayCheckImage) {
                    this.curContainer.imagePullPolicy = 'Always'
                } else {
                    this.curContainer.imagePullPolicy = 'IfNotPresent'
                }
            },
            addLocalApplication () {
                const application = JSON.parse(JSON.stringify(applicationParams))
                const index = this.statefulsets.length
                const now = +new Date()
                const applicationName = 'statefulset-' + (index + 1)
                const containerName = 'container-1'

                application.id = 'local_' + now
                application.isEdited = true

                application.config.metadata.name = applicationName
                application.config.spec.template.spec.allContainers[0].name = containerName
                this.statefulsets.push(application)
                this.setCurApplication(application, index)

                // 标签添加默认选择器
                const defaultLabels = [{
                    disabled: true,
                    isSelector: true,
                    key: 'k8s-app',
                    value: applicationName
                }]
                const defaultLabelObject = {
                    'APP': applicationName
                }
                this.updateApplicationLabel(defaultLabels, defaultLabelObject)
            },
            setCurApplication (application, index) {
                this.renderImageIndex++
                this.curApplication = application
                this.curApplicationId = application.id
                this.initLinkLabels()

                clearInterval(this.compareTimer)
                clearTimeout(this.setTimer)
                this.setTimer = setTimeout(() => {
                    if (!this.curApplication.cache) {
                        this.curApplication.cache = JSON.parse(JSON.stringify(application))
                    }
                    this.watchChange()
                }, 500)
            },
            watchChange () {
                this.compareTimer = setInterval(() => {
                    const appCopy = JSON.parse(JSON.stringify(this.curApplication))
                    const cacheCopy = JSON.parse(JSON.stringify(this.curApplication.cache))
                    // 删除无用属性
                    delete appCopy.isEdited
                    delete appCopy.cache
                    delete appCopy.id
                    delete appCopy.config.spec.template.spec.containers
                    delete appCopy.config.spec.template.spec.initContainers
                    appCopy.config.spec.template.spec.allContainers.forEach(item => {
                        if (item.ports.length === 1) {
                            const port = item.ports[0]
                            if (port.containerPort === '' && port.name === '') {
                                item.ports = []
                            }
                        }

                        item.ports.forEach(port => {
                            delete port.isLink
                        })

                        if (item.volumeMounts.length === 1) {
                            const volumn = item.volumeMounts[0]
                            if (volumn.name === '' && volumn.mountPath === '' && volumn.readOnly === false) {
                                item.volumeMounts = []
                            }
                        }
                    })

                    delete cacheCopy.isEdited
                    delete cacheCopy.cache
                    delete cacheCopy.id
                    delete cacheCopy.config.spec.template.spec.containers
                    delete cacheCopy.config.spec.template.spec.initContainers
                    cacheCopy.config.spec.template.spec.allContainers.forEach(item => {
                        if (item.ports.length === 1) {
                            const port = item.ports[0]
                            if (port.containerPort === '' && port.name === '') {
                                item.ports = []
                            }
                        }

                        item.ports.forEach(port => {
                            delete port.isLink
                        })

                        if (item.volumeMounts.length === 1) {
                            const volumn = item.volumeMounts[0]
                            if (volumn.name === '' && volumn.mountPath === '' && volumn.readOnly === false) {
                                item.volumeMounts = []
                            }
                        }
                    })

                    const appStr = JSON.stringify(appCopy)
                    const cacheStr = JSON.stringify(cacheCopy)

                    if (String(this.curApplication.id).indexOf('local_') > -1) {
                        this.curApplication.isEdited = true
                    } else if (appStr !== cacheStr) {
                        this.curApplication.isEdited = true
                    } else {
                        this.curApplication.isEdited = false
                    }
                }, 1000)
            },
            getProbeHeaderList (headers) {
                const list = []
                if (headers.forEach) {
                    headers.forEach(item => {
                        list.push({
                            key: item.name || '',
                            value: item.value || ''
                        })
                    })
                }
                if (!list.length) {
                    list.push({
                        key: '',
                        value: ''
                    })
                }
                return list
            },
            /**
             * 把上一个容器的参数重置
             */
            resetPreContainerParams () {
                this.imageVersionList = []
            },
            /**
             * 切换container
             * @param {object} container container
             */
            setCurContainer (container, index) {
                // 利用setTimeout事件来先让当前容器的blur事件执行完才切换
                setTimeout(() => {
                    // 切换container
                    // this.resetPreContainerParams()
                    container.ports.forEach(port => {
                        if (!port.protocol) {
                            port.protocol = 'TCP'
                        }
                    })
                    this.renderImageIndex++
                    this.curContainer = container
                    this.curContainerIndex = index

                    this.livenessProbeHeaders = this.getProbeHeaderList(this.curContainer.livenessProbe.httpGet.httpHeaders)
                    this.readinessProbeHeaders = this.getProbeHeaderList(this.curContainer.readinessProbe.httpGet.httpHeaders)

                    const volumesNames = this.curApplication.config.webCache.volumes.map(item => item.name)
                    const tmp = this.curContainer.volumeMounts.filter(item => {
                        return volumesNames.includes(item.name)
                    })
                    this.curContainer.volumeMounts = tmp
                }, 300)
            },
            removeContainer (index) {
                const containers = this.curApplication.config.spec.template.spec.allContainers
                containers.splice(index, 1)
                if (this.curContainerIndex === index) {
                    this.curContainerIndex = 0
                } else if (this.curContainerIndex > index) {
                    this.curContainerIndex = this.curContainerIndex - 1
                }
                this.curContainer = containers[this.curContainerIndex]
            },
            addLocalContainer () {
                // let container = Object.assign({}, containerParams)
                const container = JSON.parse(JSON.stringify(containerParams))
                const containers = this.curApplication.config.spec.template.spec.allContainers
                const index = containers.length
                container.name = 'container-' + (index + 1)
                containers.push(container)
                this.setCurContainer(container, index)
                this.$refs.containerTooltip.visible = false
            },
            removeLocalApplication (application, index) {
                // 是否删除当前项
                if (this.curApplication.id === application.id) {
                    if (index === 0 && this.statefulsets[index + 1]) {
                        this.setCurApplication(this.statefulsets[index + 1])
                    } else if (this.statefulsets[0]) {
                        this.setCurApplication(this.statefulsets[0])
                    }
                }
                this.statefulsets.splice(index, 1)
            },
            removeApplication (application, index) {
                const self = this
                const projectId = this.projectId
                const version = this.curVersion
                const id = application.id
                this.$bkInfo({
                    title: this.$t('确认删除'),
                    content: this.$createElement('p', { style: { 'text-align': 'left' } }, `${this.$t('删除StatefulSet')}：${application.config.metadata.name || this.$t('未命名')}`),
                    confirmFn () {
                        if (id.indexOf && id.indexOf('local_') > -1) {
                            self.removeLocalApplication(application, index)
                        } else {
                            self.$store.dispatch('k8sTemplate/removeStatefulset', { id, version, projectId }).then(res => {
                                const data = res.data
                                self.removeLocalApplication(application, index)

                                if (data.version) {
                                    self.$store.commit('k8sTemplate/updateCurVersion', data.version)
                                    self.$store.commit('k8sTemplate/updateBindVersion', true)
                                }
                            }, res => {
                                const message = res.message
                                self.$bkMessage({
                                    theme: 'error',
                                    message: message
                                })
                            })
                        }
                    }
                })
            },
            togglePartA () {
                this.isPartAShow = !this.isPartAShow
            },
            togglePartB () {
                this.isPartBShow = !this.isPartBShow
            },
            togglePartC () {
                this.isPartCShow = !this.isPartCShow
            },
            toggleMore () {
                this.isMorePanelShow = !this.isMorePanelShow
                this.isPodPanelShow = false
            },
            togglePod () {
                this.isPodPanelShow = !this.isPodPanelShow
                this.isMorePanelShow = false
            },
            saveStatefulsetSuccess (params) {
                this.statefulsets.forEach(item => {
                    if (params.responseData.id === item.id || params.preId === item.id) {
                        item.cache = JSON.parse(JSON.stringify(item))
                    }
                })
                if (params.responseData.id === this.curApplication.id || params.preId === this.curApplication.config.id) {
                    this.updateLocalData(params.resource)
                }
            },
            updateLocalData (data) {
                if (data.id) {
                    this.curApplication.config.id = data.id
                    this.curApplicationId = data.id
                }
                if (data.version) {
                    this.$store.commit('k8sTemplate/updateCurVersion', data.version)
                }

                this.$store.commit('k8sTemplate/updateStatefulsets', this.statefulsets)
                setTimeout(() => {
                    this.statefulsets.forEach(item => {
                        if (item.id === data.id) {
                            this.setCurApplication(item)
                        }
                    })
                }, 500)
            },
            createFirstApplication (data) {
                const templateId = this.templateId
                const projectId = this.projectId
                this.$store.dispatch('k8sTemplate/addFirstApplication', { projectId, templateId, data }).then(res => {
                    const data = res.data
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('数据保存成功')
                    })
                    this.updateLocalData(data)
                    this.isDataSaveing = false
                    if (templateId === 0 || templateId === '0') {
                        this.$router.push({
                            name: 'mesosTemplatesetApplication',
                            params: {
                                projectId: this.projectId,
                                templateId: data.template_id
                            }
                        })
                    }
                }, res => {
                    const message = res.message
                    this.$bkMessage({
                        theme: 'error',
                        message: message,
                        hasCloseIcon: true,
                        delay: '10000'
                    })
                    this.isDataSaveing = false
                })
            },
            updateApplication (data) {
                const version = this.curVersion
                const projectId = this.projectId
                const applicationId = this.curApplicationId
                this.$store.dispatch('k8sTemplate/updateApplication', { projectId, version, data, applicationId }).then(res => {
                    const data = res.data
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('数据保存成功')
                    })
                    this.updateLocalData(data)
                    this.isDataSaveing = false
                }, res => {
                    const message = res.message
                    this.$bkMessage({
                        theme: 'error',
                        message: message
                    })
                    this.isDataSaveing = false
                })
            },
            createApplication (data) {
                const version = this.curVersion
                const projectId = this.projectId
                this.$store.dispatch('k8sTemplate/addApplication', { projectId, version, data }).then(res => {
                    const data = res.data
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('数据保存成功')
                    })

                    this.updateLocalData(data)
                    this.isDataSaveing = false
                }, res => {
                    const message = res.message
                    this.$bkMessage({
                        theme: 'error',
                        message: message
                    })
                    this.isDataSaveing = false
                })
            },
            removeVolumn (item, index) {
                const allContainers = this.curApplication.config.spec.template.spec.allContainers
                const volumes = this.curApplication.config.webCache.volumes

                let matchItem
                for (const container of allContainers) {
                    matchItem = container.volumeMounts.find(volumeMount => {
                        return volumeMount.name && (volumeMount.name === item.name)
                    })
                    if (matchItem) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请先删除{name}中挂载卷已经关联项', { name: container.name || this.$('容器') })
                        })
                        return false
                    }
                }

                if (!matchItem) {
                    volumes.splice(index, 1)
                }
            },
            addVolumn () {
                const volumes = this.curApplication.config.webCache.volumes
                volumes.push({
                    type: 'emptyDir',
                    name: '',
                    source: ''
                })
            },
            removeMountVolumn (item, index) {
                const volumes = this.curContainer.volumeMounts
                volumes.splice(index, 1)
            },
            addMountVolumn () {
                const volumes = this.curContainer.volumeMounts
                volumes.push({
                    'name': '',
                    'mountPath': '',
                    'subPath': '',
                    'readOnly': false
                })
            },
            removeEnv (item, index) {
                const envList = this.curContainer.webCache.env_list
                envList.splice(index, 1)
            },
            addEnv () {
                const envList = this.curContainer.webCache.env_list
                envList.push({
                    'type': 'custom',
                    'key': '',
                    'value': ''
                })
            },
            pasteKey (item, event) {
                const cache = item.key
                this.paste(event)
                item.key = cache
                setTimeout(() => {
                    item.key = cache
                }, 0)
            },
            paste (event) {
                const clipboard = event.clipboardData
                const text = clipboard.getData('Text')
                const envList = this.curContainer.webCache.env_list
                if (text) {
                    const items = text.split('\n')
                    items.forEach(item => {
                        const arr = item.split('=')
                        envList.push({
                            type: 'custom',
                            key: arr[0],
                            value: arr[1]
                        })
                    })
                }
                setTimeout(() => {
                    this.formatEnvListData()
                }, 10)

                return false
            },
            formatEnvListData () {
                // 去掉空值
                if (this.curContainer.webCache.env_list.length) {
                    const results = []
                    const keyObj = {}
                    const length = this.curContainer.webCache.env_list.length
                    this.curContainer.webCache.env_list.forEach((item, i) => {
                        if (item.key || item.value) {
                            if (!keyObj[item.key]) {
                                results.push(item)
                                keyObj[item.key] = true
                            }
                        }
                    })
                    const patchLength = results.length - length
                    if (patchLength > 0) {
                        for (let i = 0; i < patchLength; i++) {
                            results.push({
                                type: 'custom',
                                key: '',
                                value: ''
                            })
                        }
                    }
                    this.curContainer.webCache.env_list.splice(0, this.curContainer.webCache.env_list.length, ...results)
                }
            },
            getVolumeNameList (type) {
                if (type === 'configmap') {
                    return this.configmapList
                } else if (type === 'secret') {
                    return this.secretList
                }
            },
            getVolumeSourceList (type, name) {
                if (!name) return []
                if (type === 'configmap') {
                    const list = this.configmapList
                    for (const item of list) {
                        if (item.name === name) {
                            return item.childList
                        }
                    }
                    return []
                } else if (type === 'secret') {
                    const list = this.secretList
                    for (const item of list) {
                        if (item.name === name) {
                            return item.childList
                        }
                    }
                    return []
                }
                return []
            },
            selectOperate (data) {
                const operate = data.operate
                if (operate === 'UNIQUE') {
                    data.type = 0
                    data.arg_value = ''
                }
            },
            updateNodeSelectorList (list, data) {
                if (!this.curApplication.config.webCache) {
                    this.curApplication.config.webCache = {}
                }
                this.curApplication.config.webCache.nodeSelectorList = list
            },
            updateApplicationRemark (list, data) {
                if (!this.curApplication.config.webCache) {
                    this.curApplication.config.webCache = {}
                }
                this.curApplication.config.webCache.remarkListCache = list
            },
            updateApplicationImageSecrets (list, data) {
                const secrets = []
                list.forEach(item => {
                    secrets.push({
                        name: item.value
                    })
                })
                this.curApplication.config.spec.template.spec.imagePullSecrets = secrets
            },
            updateApplicationLabel (list, data) {
                if (!this.curApplication.config.webCache) {
                    this.curApplication.config.webCache = {}
                }
                this.curApplication.config.webCache.labelListCache = list
            },
            updateApplicationLogLabel (list, data) {
                if (!this.curApplication.config.webCache) {
                    this.curApplication.config.webCache = {}
                }
                this.curApplication.config.customLogLabel = data
                this.curApplication.config.webCache.logLabelListCache = list
            },
            formatData () {
                const params = JSON.parse(JSON.stringify(this.curApplication))
                params.template = {
                    name: this.curTemplate.name,
                    desc: this.curTemplate.desc
                }
                delete params.isEdited
                // 键值转换
                const remarkKeyList = this.$refs.remarkKeyer.getKeyObject()
                const labelKeyList = this.$refs.labelKeyer.getKeyObject()

                params.metadata.labels = labelKeyList
                params.metadata.annotations = remarkKeyList

                // 转换调度约束
                const constraint = params.constraint.intersectionItem
                constraint.forEach(item => {
                    const data = item.unionData[0]
                    const operate = data.operate
                    switch (operate) {
                        case 'UNIQUE':
                            delete data.type
                            delete data.set
                            delete data.text
                            break
                        case 'MAXPER':
                            data.type = 3
                            data.text = {
                                'value': data.arg_value
                            }
                            delete data.set
                            break
                        case 'CLUSTER':
                            data.type = 4
                            if (data.arg_value.trim().length) {
                                data.set = {
                                    'item': data.arg_value.split('|')
                                }
                            } else {
                                data.set = {
                                    'item': []
                                }
                            }

                            delete data.text
                            break
                        case 'GROUPBY':
                            data.type = 4
                            if (data.arg_value.trim().length) {
                                data.set = {
                                    'item': data.arg_value.split('|')
                                }
                            } else {
                                data.set = {
                                    'item': []
                                }
                            }
                            delete data.text
                            break
                        case 'LIKE':
                            if (data.arg_value.indexOf('|') > -1) {
                                data.type = 4
                                if (data.arg_value.trim().length) {
                                    data.set = {
                                        'item': data.arg_value.split('|')
                                    }
                                } else {
                                    data.set = {
                                        'item': []
                                    }
                                }
                                delete data.text
                            } else {
                                data.type = 3
                                data.text = {
                                    'value': data.arg_value
                                }
                                delete data.set
                            }
                            break
                        case 'UNLIKE':
                            if (data.arg_value.indexOf('|') > -1) {
                                data.type = 4
                                if (data.arg_value.trim().length) {
                                    data.set = {
                                        'item': data.arg_value.split('|')
                                    }
                                } else {
                                    data.set = {
                                        'item': []
                                    }
                                }
                                delete data.text
                            } else {
                                data.type = 3
                                data.text = {
                                    'value': data.arg_value
                                }
                                delete data.set
                            }
                            break
                    }
                })

                // 转换命令参数和环境变量
                const containers = params.spec.template.spec.allContainers
                containers.forEach(container => {
                    if (container.args_text.trim().length) {
                        container.args = container.args_text.split(' ')
                    } else {
                        container.args = []
                    }

                    container.resources.limits.cpu = parseFloat(container.resources.limits.cpu)

                    // docker参数
                    const parameterList = container.parameter_list
                    container.parameters = []
                    parameterList.forEach(param => {
                        if (param.key && param.value) {
                            container.parameters.push(param)
                        }
                    })

                    // 端口
                    const ports = container.ports
                    const validatePorts = []
                    ports.forEach(item => {
                        if (item.containerPort && (item.hostPort !== undefined) && item.name && item.protocol) {
                            validatePorts.push({
                                id: item.id,
                                containerPort: item.containerPort,
                                hostPort: item.hostPort,
                                protocol: item.protocol,
                                name: item.name
                            })
                        }
                    })
                    container.ports = validatePorts

                    // volumes
                    const volumes = container.volumes
                    let validateVolumes = []
                    validateVolumes = volumes.filter(item => {
                        return item.volume.hostPath && item.volume.mountPath && item.name
                    })
                    container.volumes = validateVolumes

                    // logpath
                    const paths = []
                    const logList = container.logListCache
                    logList.forEach(item => {
                        if (item.value) {
                            paths.push(item.value)
                        }
                    })
                    container.logPathList = paths
                })
                return params
            },
            saveApplication () {
                if (!this.checkData()) {
                    return false
                }
                if (this.isDataSaveing) {
                    return false
                } else {
                    this.isDataSaveing = true
                }
                const data = this.formatData()
                if (this.curVersion) {
                    if (this.curApplicationId.indexOf && this.curApplicationId.indexOf('local') > -1) {
                        this.createApplication(data)
                    } else {
                        this.updateApplication(data)
                    }
                } else {
                    this.createFirstApplication(data)
                }
            },
            initImageList () {
                if (this.isLoadingImageList) return false
                this.isLoadingImageList = true
                const projectId = this.projectId
                this.$store.dispatch('k8sTemplate/getImageList', { projectId }).then(res => {
                    const data = res.data
                    setTimeout(() => {
                        data.forEach(item => {
                            item._id = item.value
                            item._name = item.name
                        })
                        this.imageList.splice(0, this.imageList.length, ...data)
                        this.$store.commit('k8sTemplate/updateImageList', this.imageList)
                        this.isLoadingImageList = false
                    }, 1000)
                }, res => {
                    const message = res.message
                    this.$bkMessage({
                        theme: 'error',
                        message: message,
                        delay: 10000
                    })
                    this.isLoadingImageList = false
                })
            },
            handleImageCustom () {
                setTimeout(() => {
                    const imageName = this.curContainer.webCache.imageName
                    const imageVersion = this.curContainer.imageVersion
                    if (imageName && imageVersion) {
                        this.curContainer.image = `${imageName}:${imageVersion}`
                    } else {
                        this.curContainer.image = ''
                    }
                }, 100)
            },

            handleVersionCustom () {
                this.$nextTick(() => {
                    const versionName = this.curContainer.imageVersion
                    const matcher = this.imageVersionList.find(version => version._name === versionName)
                    if (matcher) {
                        this.setImageVersion(matcher._id, matcher)
                    } else {
                        const imageName = this.curContainer.webCache.imageName
                        const version = this.curContainer.imageVersion

                        // curImageData有值，表示是通过选择
                        if (JSON.stringify(this.curImageData) !== '{}') {
                            if (this.curImageData.is_pub !== undefined) {
                                this.curContainer.image = `${DEVOPS_ARTIFACTORY_HOST}/${imageName}:${version}`
                                console.log('镜像是下拉，版本是自定义', this.curContainer.image)
                            } else {
                                this.curContainer.image = `${DEVOPS_ARTIFACTORY_HOST}/paas/${this.projectCode}/${imageName}:${version}`
                                console.log('镜像是变量，版本是自定义', this.curContainer.image)
                            }
                        } else {
                            this.curContainer.image = `${imageName}:${version}`
                            console.log('镜像和版本都是自定义', this.curContainer.image)
                        }
                    }
                })
            },

            handleChangeImageMode () {
                this.curContainer.webCache.isImageCustomed = !this.curContainer.webCache.isImageCustomed
                // 清空原来值
                this.curContainer.webCache.imageName = ''
                this.curContainer.image = ''
                this.curContainer.imageName = ''
                this.curContainer.imageVersion = ''
            },

            changeImage (value, data, isInitTrigger) {
                const projectId = this.projectId
                const imageId = data.value
                const isPub = data.is_pub
                this.curImageData = data
                // 如果不是输入变量
                if (isPub !== undefined) {
                    this.$store.dispatch('k8sTemplate/getImageVertionList', { projectId, imageId, isPub }).then(res => {
                        const data = res.data
                        data.forEach(item => {
                            item._id = item.text
                            item._name = item.text
                        })

                        this.imageVersionList.splice(0, this.imageVersionList.length, ...data)

                        // 非首次关联触发，默认选择第一项或清空
                        if (isInitTrigger) return

                        if (this.imageVersionList.length) {
                            const imageInfo = this.imageVersionList[0]

                            this.curContainer.image = imageInfo.value
                            this.curContainer.imageVersion = imageInfo.text
                        } else {
                            this.curContainer.image = ''
                            this.curContainer.imageVersion = ''
                        }
                    }, res => {
                        this.curContainer.image = ''
                        this.curContainer.imageVersion = ''
                        const message = res.message
                        this.$bkMessage({
                            theme: 'error',
                            message: message
                        })
                    })
                } else if (!isInitTrigger) {
                    this.imageVersionList = []
                    this.curContainer.image = ''
                    this.curContainer.imageVersion = ''
                }
            },

            setImageVersion (value, data) {
                // 镜像和版本都是通过下拉选择
                const projectCode = this.projectCode
                // curImageData不是空对象
                if (JSON.stringify(this.curImageData) !== '{}') {
                    if (data.text && data.value) {
                        this.curContainer.imageVersion = data.text
                        this.curContainer.image = data.value
                    } else if (this.curImageData.is_pub !== undefined) {
                        // 镜像是下拉，版本是变量
                        // image = imageBase + imageName + ':' + imageVersion
                        const imageName = this.curContainer.webCache.imageName
                        this.curContainer.imageVersion = value
                        this.curContainer.image = `${DEVOPS_ARTIFACTORY_HOST}/${imageName}:${value}`
                    } else {
                        // 镜像和版本是变量
                        // image = imageBase +  'paas/' + projectCode + '/' + imageName + ':' + imageVersion
                        const imageName = this.curContainer.webCache.imageName
                        this.curContainer.imageVersion = value
                        this.curContainer.image = `${DEVOPS_ARTIFACTORY_HOST}/paas/${projectCode}/${imageName}:${value}`
                    }
                }
            },
            addPort () {
                const id = +new Date()
                const params = {
                    id: id,
                    containerPort: '',
                    protocol: 'TCP',
                    name: '',
                    isLink: false
                }

                this.curContainer.ports.push(params)
            },
            addLog () {
                this.curContainer.webCache.logListCache.push({
                    value: ''
                })
            },
            removeLog (log, index) {
                this.curContainer.webCache.logListCache.splice(index, 1)
            },
            changeProtocol (item) {
                const projectId = this.projectId
                const version = this.curVersion
                const portId = item.id
                this.$store.dispatch('k8sTemplate/checkPortIsLink', { projectId, version, portId }).then(res => {
                }, res => {
                    const message = res.message || res.data.data
                    const msg = message.split(',')[0]
                    this.$bkMessage({
                        theme: 'error',
                        message: msg + this.$t('，不能修改协议')
                    })
                })
            },
            removePort (item, index) {
                const projectId = this.projectId
                const version = this.curVersion
                const portId = item.id
                this.$store.dispatch('k8sTemplate/checkPortIsLink', { projectId, version, portId }).then(res => {
                    this.curContainer.ports.splice(index, 1)
                })
            },
            selectVolumeType (volumeItem) {
                // volumeItem.name = ''
                // volumeItem.mountPath = ''
                // let data = Object.assign([], this.curContainer.volumeMounts)
                // this.curContainer.volumeMounts.splice(0, this.curContainer.volumeMounts.length, ...data)
            },
            setVolumeName (volumeItem) {
                volumeItem.volume.hostPath = ''
            },
            initVolumeConfigmaps () {
                const version = this.curVersion
                if (!version) {
                    return false
                }
                const projectId = this.projectId

                this.$store.dispatch('k8sTemplate/getConfigmaps', { projectId, version }).then(res => {
                    const data = res.data
                    const keyList = []
                    data.forEach(item => {
                        const list = []
                        const name = item.name
                        const keys = item.keys
                        item.id = item.name
                        keys.forEach(key => {
                            const params = {
                                id: name + '.' + key,
                                name: name + '.' + key,
                                nameCache: name,
                                keyCache: key
                            }
                            list.push(params)
                            keyList.push(params)
                        })
                        item.childList = list
                    })
                    this.volumeConfigmapList = data
                    this.configmapKeyList.splice(0, this.configmapKeyList.length, ...keyList)
                    this.configmapList.splice(0, this.configmapList.length, ...data)
                }, res => {
                    const message = res.message
                    this.$bkMessage({
                        theme: 'error',
                        message: message
                    })
                })
            },
            initVloumeSelectets () {
                const version = this.curVersion
                if (!version) {
                    return false
                }
                const projectId = this.projectId

                this.$store.dispatch('k8sTemplate/getSecrets', { projectId, version }).then(res => {
                    const data = res.data
                    const keyList = []
                    data.forEach(item => {
                        const list = []
                        const name = item.name
                        const keys = item.keys
                        keys.forEach(key => {
                            const params = {
                                id: name + '.' + key,
                                name: name + '.' + key,
                                nameCache: name,
                                keyCache: key
                            }
                            list.push(params)
                            keyList.push(params)
                        })

                        item.childList = list
                    })
                    this.volumeSecretList = data
                    this.secretKeyList.splice(0, this.secretKeyList.length, ...keyList)
                    this.secretList.splice(0, this.secretList.length, ...data)
                }, res => {
                    const message = res.message
                    this.$bkMessage({
                        theme: 'error',
                        message: message
                    })
                })
            },
            initMetricList () {
                const projectId = this.projectId
                this.$store.dispatch('k8sTemplate/getMetricList', projectId)
            },
            initLinkLabels () {
                const projectId = this.projectId
                const versionId = this.curVersion
                this.$store.dispatch('k8sTemplate/getApplicationLinkLabels', { projectId, versionId }).then(res => {
                    const data = res.data
                    for (const key in data) {
                        const keys = key.split(':')
                        if (keys.length >= 2) {
                            this.curApplicationLinkLabels = [
                                {
                                    key: keys[0],
                                    value: keys[1],
                                    linkMessage: this.$t('标签 ({key}) 已经被Service ({service}) 关联，使用该标签的应用会被关联的Service导流', {
                                        key: key,
                                        service: data[key].join('；')
                                    })
                                }
                            ]
                        }
                    }
                }, res => {
                    this.curApplicationLinkLabels = []
                })
            },
            addHostAlias () {
                this.curApplication.config.webCache.hostAliasesCache.push({
                    ip: '',
                    hostnames: ''
                })
            },

            removeHostAlias (item, index) {
                this.curApplication.config.webCache.hostAliasesCache.splice(index, 1)
            }
        }
    }
</script>

<style scoped>
    @import './statefulset.css';
</style>
