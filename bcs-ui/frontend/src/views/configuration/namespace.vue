<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-config-namespace-title">
                {{$t('命名空间')}}
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper biz-namespace-loading" style="padding: 0;" v-bkloading="{ isLoading: isInitLoading, opacity: 0.1 }">
            <app-exception
                v-if="exceptionCode && !isInitLoading"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>
            <template v-if="!exceptionCode && !isInitLoading">
                <div class="biz-panel-header">
                    <div class="left">
                        <bk-button type="primary" @click.stop.prevent="showAddNamespace">
                            <i class="bcs-icon bcs-icon-plus"></i>
                            <span>{{$t('新建')}}</span>
                        </bk-button>
                        <bcs-popover v-if="!isSharedCluster" :content="$t('同步非本页面创建的命名空间数据')" placement="top">
                            <bk-button class="bk-button" @click.stop.prevent="syncNamespace">
                                <span>{{$t('同步命名空间')}}</span>
                            </bk-button>
                        </bcs-popover>
                        <span class="biz-tip ml10" style="vertical-align: middle;">{{$t('命名空间创建后不可更改')}}</span>
                    </div>
                    <div class="right">
                        <bk-data-searcher
                            ref="dataSearcher"
                            :search-key.sync="search"
                            :scope-list="searchScopeList"
                            :search-scope.sync="searchScope"
                            :cluster-fixed="!!curClusterId"
                            @search="fetchNamespaceList"
                            @refresh="refresh"
                            :key="isSharedCluster">
                        </bk-data-searcher>
                    </div>
                </div>

                <div class="biz-namespace">
                    <div class="biz-table-wrapper">
                        <bk-table
                            class="biz-namespace-table"
                            v-bkloading="{ isLoading: isPageLoading }"
                            :data="curPageData"
                            :page-params="pageConf"
                            @page-change="pageChange"
                            @page-limit-change="changePageSize">
                            <bk-table-column :label="$t('名称')" prop="name" :show-overflow-tooltip="true" :min-width="100">
                                <template slot-scope="{ row }">
                                    <span class="text">{{row.name}}</span>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('所属集群')" :show-overflow-tooltip="false" prop="cluster_name" :width="200">
                                <template slot-scope="{ row }">
                                    <bcs-popover :content="row.cluster_id || '--'" placement="top">
                                        <p class="biz-text-wrapper">{{row.cluster_name ? row.cluster_name : '--'}}</p>
                                    </bcs-popover>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('变量')" prop="ns_vars" :min-width="150" :show-overflow-tooltip="false">
                                <template slot-scope="{ row, $index }">
                                    <div style="position: relative;" v-if="row.ns_vars.length">
                                        <bcs-popover :delay="300" placement="left">
                                            <div class="labels-container" :class="row.isExpandLabels ? 'expand' : ''">
                                                <div class="labels-wrapper" :class="row.isExpandLabels ? 'expand' : ''" :ref="`${pageConf.curPage}-real${$index}`">
                                                    <div class="labels-inner" v-for="(label, labelIndex) in row.ns_vars" :key="labelIndex">
                                                        <span class="key">{{label.key}}</span>
                                                        <template v-if="label.value">
                                                            <span class="value">{{label.value}}</span>
                                                        </template>
                                                        <template v-else>
                                                            <span class="value">{{label.value}}</span>
                                                        </template>
                                                    </div>
                                                    <span v-if="row.showExpand" style="position: relative; top: 8px;">...</span>
                                                </div>
                                            </div>
                                            <template slot="content">
                                                <div class="labels-wrapper fake">
                                                    <div class="labels-inner" v-for="(label, labelIndex) in row.ns_vars" :key="labelIndex">
                                                        <div>
                                                            <span class="key">{{label.key}}</span>:
                                                            <span class="value">{{label.value}}</span>
                                                        </div>
                                                    </div>
                                                </div>
                                            </template>
                                        </bcs-popover>
                                    </div>
                                    <span v-else>--</span>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('操作')" prop="permissions" width="310" class-name="biz-table-action-column">
                                <template slot-scope="{ row }">
                                    <a href="javascript:void(0)" class="bk-text-button"
                                        @click="showEditNamespace(row, index)"
                                        v-authority="{
                                            clickable: web_annotations.perms[row.iam_ns_id]
                                                && web_annotations.perms[row.iam_ns_id].namespace_update,
                                            actionId: 'namespace_update',
                                            resourceName: row.name,
                                            disablePerms: true,
                                            permCtx: {
                                                project_id: projectId,
                                                cluster_id: row.cluster_id,
                                                name: row.name
                                            }
                                        }"
                                    >
                                        {{$t('设置变量值')}}
                                    </a>
                                    <a class="bk-text-button ml10"
                                        @click="showEditQuota(row, index)"
                                        v-authority="{
                                            clickable: web_annotations.perms[row.iam_ns_id]
                                                && web_annotations.perms[row.iam_ns_id].namespace_update,
                                            actionId: 'namespace_update',
                                            resourceName: row.name,
                                            disablePerms: true,
                                            permCtx: {
                                                project_id: projectId,
                                                cluster_id: row.cluster_id,
                                                name: row.name
                                            }
                                        }"
                                    >
                                        {{$t('配额管理')}}
                                    </a>
                                    <a href="javascript:void(0)" class="bk-text-button"
                                        @click="showDelNamespace(row, index)"
                                        v-authority="{
                                            clickable: web_annotations.perms[row.iam_ns_id]
                                                && web_annotations.perms[row.iam_ns_id].namespace_delete,
                                            actionId: 'namespace_delete',
                                            resourceName: row.name,
                                            disablePerms: true,
                                            permCtx: {
                                                project_id: projectId,
                                                cluster_id: row.cluster_id,
                                                name: row.name
                                            }
                                        }"
                                    >
                                        {{$t('删除')}}
                                    </a>
                                </template>
                            </bk-table-column>
                        </bk-table>
                    </div>
                </div>
            </template>
        </div>

        <bk-sideslider
            :is-show.sync="addNamespaceConf.isShow"
            :title="addNamespaceConf.title"
            :width="addNamespaceConf.width"
            :quick-close="false"
            class="biz-cluster-set-variable-sideslider"
            @hidden="hideAddNamespace">
            <div slot="content">
                <div class="wrapper" style="position: relative;">
                    <bcs-form form-type="vertical">
                        <bcs-form-item :label="$t('所属集群')" :required="true">
                            <bk-selector
                                :field-type="'cluster'"
                                :placeholder="$t('请选择集群')"
                                :setting-key="'cluster_id'"
                                :display-key="'name'"
                                :searchable="true"
                                :search-key="'name'"
                                :selected.sync="clusterId"
                                :list="clusterList"
                                :disabled="!!curClusterId"
                                @item-selected="chooseCluster">
                            </bk-selector>
                        </bcs-form-item>

                        <bcs-form-item :label="$t('名称')" :desc="namespaceNameTips" desc-type="icon" :required="true">
                            <bk-input v-if="!isSharedCluster" :placeholder="$t('请输入')" v-model="addNamespaceConf.namespaceName" maxlength="30" />
                            <div v-else class="namespace-name">
                                <span class="namespaceName-left">{{ $INTERNAL ? 'ieg-' + projectCode : projectCode }} -</span>
                                <span class="namespaceName-right">
                                    <bk-input :placeholder="$t('请输入')" v-model="addNamespaceConf.namespaceName" maxlength="30" />
                                </span>
                            </div>
                        </bcs-form-item>

                        <!-- 配额 start -->
                        <template v-if="curProject.kind !== 2">
                            <div class="quota-option">
                                <label class="bk-label label" style="width: 100%;">
                                    {{$t('配额设置')}}
                                    <bk-switcher v-if="!isSharedCluster" class="quota-switcher" size="small" :selected="showQuota" @change="toggleShowQuota" :key="showQuota"></bk-switcher>
                                </label>
                            </div>
                        </template>

                        <template v-if="showQuota">
                            <!-- 内存 -->
                            <bcs-form-item class="requestsMem-item" label="MEM" :required="true">
                                <div class="requestsMem-content">
                                    <bcs-slider v-model="quotaData.requestsMem" :min-value="1" :max-value="400" />
                                    <bcs-input
                                        v-model="quotaData.requestsMem"
                                        type="number"
                                        :min="1"
                                        :max="400"
                                        @blur="handleBlurRequestsMem">
                                    </bcs-input>
                                    G
                                </div>
                            </bcs-form-item>

                            <!-- CPU -->
                            <bcs-form-item class="requestsCpu-item" label="CPU" :required="true">
                                <div class="requestsCpu-content">
                                    <bcs-slider v-model="quotaData.requestsCpu" :min-value="1" :max-value="400" />
                                    <bcs-input
                                        v-model="quotaData.requestsCpu"
                                        type="number"
                                        :min="1"
                                        :max="400"
                                        @blur="handleBlurRequestsCpu">
                                    </bcs-input>
                                    核
                                </div>
                            </bcs-form-item>
                        </template>
                        <!-- 配额 end -->

                        <!-- 变量 start -->
                        <template v-if="addNamespaceConf.variableList && addNamespaceConf.variableList.length">
                            <bcs-form-item :label="$t('变量设置')">
                                <div class="biz-key-value-wrapper mb10">
                                    <div class="biz-key-value-item" v-for="(variable, index) in addNamespaceConf.variableList" :key="index">
                                        <bk-input style="width: 270px;" :disabled="true" v-model="variable.leftContent" />
                                        <span class="equals-sign">=</span>
                                        <bk-input style="width: 270px; margin-left: 35px;" :placeholder="$t('值')" v-model="variable.value"></bk-input>
                                    </div>
                                </div>
                            </bcs-form-item>
                        </template>
                        <!-- 变量 end -->

                        <div class="action-inner">
                            <bk-button type="primary"
                                :loading="addNamespaceConf.loading"
                                v-authority="{
                                    clickable: true,
                                    actionId: 'namespace_create',
                                    autoUpdatePerms: true,
                                    permCtx: {
                                        resource_type: 'cluster',
                                        project_id: projectId,
                                        cluster_id: clusterId
                                    }
                                }" @click="confirmAddNamespace">
                                {{$t('保存')}}
                            </bk-button>
                            <bk-button @click="hideAddNamespace" :disabled="addNamespaceConf.loading">
                                {{$t('取消')}}
                            </bk-button>
                        </div>
                    </bcs-form>
                </div>
            </div>
        </bk-sideslider>

        <bk-sideslider
            :is-show.sync="editNamespaceConf.isShow"
            :title="editNamespaceConf.title"
            :width="editNamespaceConf.width"
            :quick-close="false"
            class="biz-cluster-set-variable-sideslider"
            @hidden="hideEditNamespace">
            <div slot="content">
                <div class="wrapper" style="position: relative;">
                    <div class="bk-form bk-form-vertical set-label-form">
                        <div class="bk-form-item flex-item">
                            <div class="left">
                                <label class="bk-label label">{{$t('名称：')}}</label>
                            </div>
                            <div class="right" style="margin-left: 20px;">
                                <label class="bk-label label">{{$t('所属集群：')}}</label>
                            </div>
                        </div>
                        <div class="bk-form-item flex-item">
                            <div class="left">
                                <bk-input v-model="editNamespaceConf.namespaceName" :disabled="!editNamespaceConf.canEdit"></bk-input>
                            </div>
                            <div class="right" style="margin-left: 20px;">
                                <div class="cluster-wrapper">
                                    <bk-selector
                                        :field-type="'cluster'"
                                        :placeholder="$t('请选择')"
                                        :setting-key="'cluster_id'"
                                        :display-key="'name'"
                                        :searchable="true"
                                        :search-key="'name'"
                                        :selected.sync="clusterId"
                                        :list="clusterList"
                                        :disabled="!editNamespaceConf.canEdit"
                                        @item-selected="chooseCluster">
                                    </bk-selector>
                                </div>
                            </div>
                        </div>
                        <template v-if="editNamespaceConf.variableList && editNamespaceConf.variableList.length">
                            <div class="bk-form-item flex-item" style="margin-top: 20px;">
                                <div class="left">
                                    <label class="bk-label label">
                                        {{$t('变量：')}}
                                        <i18n path="（可通过 {action} 创建更多作用在命名空间的变量）" class="biz-tip fn">
                                            <button place="action" class="bk-text-button" @click="handleGoVar">{{$t('变量管理')}}</button>
                                        </i18n>
                                    </label>
                                </div>
                            </div>
                            <div class="bk-form-item">
                                <div class="bk-form-content">
                                    <div class="biz-key-value-wrapper mb10">
                                        <div class="biz-key-value-item" v-for="(variable, index) in editNamespaceConf.variableList" :key="index">
                                            <bk-input style="width: 270px;" :disabled="true" v-model="variable.leftContent" />
                                            <span class="equals-sign">=</span>
                                            <bk-input style="width: 270px; margin-left: 35px;" :placeholder="$t('值')" v-model="variable.value"></bk-input>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div class="action-inner">
                                <bk-button type="primary" :loading="editNamespaceConf.loading" @click="confirmEditNamespace">
                                    {{$t('保存')}}
                                </bk-button>
                                <bk-button @click="hideEditNamespace" :disabled="editNamespaceConf.loading">
                                    {{$t('取消')}}
                                </bk-button>
                            </div>
                        </template>
                        <template v-else>
                            <i18n path="该项目未设置作用在命名空间范围的环境变量，无法设置变量值，可前往 {action} 设置" class="biz-tip mt40" tag="div">
                                <router-link place="action" class="bk-text-button" :to="{ name: 'var', params: { projectCode: projectCode } }">{{$t('变量管理')}}</router-link>
                            </i18n>
                        </template>
                    </div>
                </div>
            </div>
        </bk-sideslider>

        <bk-sideslider
            :is-show.sync="editQuotaConf.isShow"
            :title="editQuotaConf.title"
            :width="editQuotaConf.width"
            :quick-close="false"
            class="biz-cluster-set-variable-sideslider"
            @hidden="hideEditQuota">
            <div slot="content" v-bkloading="{ isLoading: editQuotaConf.initLoading }">
                <div class="wrapper" style="position: relative;">
                    <div class="bk-form bk-form-vertical set-label-form">
                        <div class="bk-form-item flex-item">
                            <div class="left">
                                <label class="bk-label label">{{$t('名称')}}</label>
                            </div>
                            <div class="right" style="margin-left: 20px;">
                                <label class="bk-label label">{{$t('所属集群')}}</label>
                            </div>
                        </div>
                        <div class="bk-form-item flex-item">
                            <div class="left">
                                <bk-input :disabled="true" v-model="editQuotaConf.namespaceName"></bk-input>
                            </div>
                            <div class="right" style="margin-left: 20px;">
                                <div class="cluster-wrapper">
                                    <bk-selector
                                        :field-type="'cluster'"
                                        :placeholder="$t('请选择')"
                                        :setting-key="'cluster_id'"
                                        :display-key="'name'"
                                        :searchable="true"
                                        :search-key="'name'"
                                        :selected.sync="clusterId"
                                        :list="clusterList"
                                        :disabled="true">
                                    </bk-selector>
                                </div>
                            </div>
                        </div>
                        <div class="bk-form-item flex-item" style="margin: 30px 0;">
                            <div class="left">
                                <label class="bk-label label">
                                    {{$t('配额')}}
                                    <span class="quota-tip">{{$t('分配命名空间下容器可用的内存和 CPU 总量')}}</span>
                                </label>
                            </div>
                        </div>

                        <div class="bk-form-item requestsMem-item" style="margin-top: 32px;">
                            <div class="quota-label-tip">
                                <span class="title">MEM</span>
                            </div>
                            <div class="bk-form-content">
                                <div class="requestsMem-content">
                                    <bcs-slider v-model="quotaData.requestsMem" :min-value="0" :max-value="400" />
                                    <bcs-input
                                        v-model="quotaData.requestsMem"
                                        type="number"
                                        :min="0"
                                        :max="400"
                                        @blur="handleBlurRequestsMem">
                                    </bcs-input>
                                    G
                                </div>
                            </div>
                        </div>
                        <div class="bk-form-item requestsCpu-item" style="margin-top: 18px;">
                            <div class="quota-label-tip">
                                <span class="title">CPU</span>
                            </div>
                            <div class="bk-form-content">
                                <div class="requestsCpu-content">
                                    <bcs-slider v-model="quotaData.requestsCpu" :min-value="0" :max-value="400" />
                                    <bcs-input
                                        v-model="quotaData.requestsCpu"
                                        type="number"
                                        :min="0"
                                        :max="400"
                                        @blur="handleBlurRequestsCpu">
                                    </bcs-input>
                                    核
                                </div>
                            </div>
                        </div>
                        <div class="action-inner">
                            <bk-button type="primary" :loading="editQuotaConf.loading" @click="confirmEditQuota">
                                {{$t('保存')}}
                            </bk-button>
                            <bk-button @click="hideEditQuota" :disabled="editQuotaConf.loading">
                                {{$t('取消')}}
                            </bk-button>
                            <bk-button type="danger" @click="showDelQuota">
                                {{$t('删除')}}
                            </bk-button>
                        </div>
                    </div>
                </div>
            </div>
        </bk-sideslider>

        <bk-dialog
            :is-show.sync="delNamespaceDialogConf.isShow"
            :title="$t('删除命名空间')"
            :header-position="'left'"
            :width="delNamespaceDialogConf.width"
            :ext-cls="'biz-namespace-del-dialog'"
            :has-header="false"
            :quick-close="false"
            @cancel="delNamespaceDialogConf.isShow = false">
            <template slot="content" style="padding: 0 20px;">
                <div class="info">
                    {{$t('您确定要删除Namespace: {name}吗？', { name: delNamespaceDialogConf.ns.name })}}
                </div>
                <div style="color: red;">
                    {{$t('删除Namespace将销毁Namespace下的所有资源，销毁后所有数据将被清除且不可恢复，请提前备份好数据。')}}
                </div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <bk-button type="primary" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary"
                        @click="delNamespaceConfirm">
                        {{$t('删除')}}
                    </bk-button>
                    <bk-button class="bk-dialog-btn bk-dialog-btn-cancel" @click="delNamespaceCancel">
                        {{$t('取消')}}
                    </bk-button>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :title="$t('删除配额')"
            :is-show.sync="delQuotaDialogConf.isShow"
            :width="delQuotaDialogConf.width"
            :ext-cls="'biz-namespace-del-dialog'"
            :header-position="'left'"
            :has-header="false"
            :quick-close="false"
            @cancel="delQuotaDialogConf.isShow = false">
            <template slot="content" style="padding: 0 20px;">
                <div class="info">
                    {{$t('确定删除Namespace: {name}的配额？', { name: delQuotaDialogConf.ns.name })}}
                </div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <bk-button type="primary" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary"
                        @click="delQuotaConfirm">
                        {{$t('删除')}}
                    </bk-button>
                    <bk-button class="bk-dialog-btn bk-dialog-btn-cancel" @click="delQuotaCancel">
                        {{$t('取消')}}
                    </bk-button>
                </div>
            </div>
        </bk-dialog>
    </div>
</template>

<script>
    import { catchErrorHandler } from '@/common/util'
    import { mapGetters } from 'vuex'
    export default {
        data () {
            // 环境类型 list
            const envList = [
                {
                    id: 'dev',
                    name: 'dev'
                },
                {
                    id: 'test',
                    name: 'test'
                },
                {
                    id: 'prod',
                    name: 'prod'
                }
            ]
            return {
                isNamespaceAdd: false,
                envList: envList,
                addEnvIndex: -1,
                editEnvIndex: -1,
                clusterId: '',
                editClusterId: -1,
                isPageLoading: false,
                pageConf: {
                    total: 1,
                    totalPage: 1,
                    pageSize: 10,
                    curPage: 1,
                    show: true
                },
                namespaceList: [],
                // 缓存，用于搜索
                namespaceListTmp: [],
                curPageData: [],
                newName: '',
                editName: '',
                isInitLoading: true,
                search: '',
                addNamespaceConf: {
                    isShow: false,
                    title: this.$t('新建命名空间'),
                    width: 680,
                    variableList: [],
                    namespaceName: '',
                    loading: false
                },
                editNamespaceConf: {
                    isShow: false,
                    title: '',
                    width: 680,
                    variableList: [],
                    namespaceName: '',
                    canEdit: false,
                    ns: {},
                    loading: false
                },
                exceptionCode: null,
                bkMessageInstance: null,
                delNamespaceDialogConf: {
                    isShow: false,
                    width: 650,
                    title: '',
                    closeIcon: false,
                    ns: {}
                },
                showQuota: false,
                editQuotaConf: {
                    isShow: false,
                    title: '',
                    width: 680,
                    namespaceName: '',
                    ns: {},
                    loading: false,
                    initLoading: false
                },
                quotaData: {
                    limitsCpu: '400',
                    requestsCpu: '',
                    limitsMem: '400',
                    requestsMem: ''
                },
                // 数字输入框中允许输入的键盘按钮的 keyCode 集合
                validKeyCodeList4QuotaInput: [
                    48, 49, 50, 51, 52, 53, 54, 55, 56, 57, // 0-9
                    96, 97, 98, 99, 100, 101, 102, 103, 104, 105, // 0-9
                    8, // backspace
                    38, 40, 37, 39, // up down left right
                    46, // del
                    9, // tab
                    13 // enter
                ],
                delQuotaDialogConf: {
                    isShow: false,
                    width: 380,
                    title: '',
                    closeIcon: false,
                    ns: {}
                },
                areaIndex: -1,
                web_annotations: { perms: {} }
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            clusterList () {
                return this.$store.state.cluster.clusterList
            },
            searchScopeList () {
                const clusterList = this.clusterList
                const results = []
                if (clusterList.length) {
                    clusterList.forEach(item => {
                        results.push({
                            id: item.cluster_id,
                            name: item.name
                        })
                    })
                }

                return results
            },
            onlineProjectList () {
                return this.$store.state.sideMenu.onlineProjectList
            },
            isEn () {
                return this.$store.state.isEn
            },
            curProject () {
                return this.$store.state.curProject
            },
            isClusterDataReady () {
                return this.$store.state.cluster.isClusterDataReady
            },
            curClusterId () {
                return this.$store.state.curClusterId
            },
            namespaceNameTips () {
                if (this.$INTERNAL) {
                    return this.isSharedCluster ? this.$t('规则: ieg-项目英文名称-自定义名称') : ''
                } else {
                    return this.isSharedCluster ? this.$t('规则: 项目英文名称-自定义名称') : ''
                }
            },
            ...mapGetters('cluster', ['isSharedCluster'])
        },
        watch: {
            isClusterDataReady: {
                immediate: true,
                handler (val) {
                    if (val) {
                        if (this.searchScopeList.length) {
                            const clusterIds = this.searchScopeList.map(item => item.id)
                            // 使用当前缓存
                            if (sessionStorage['bcs-cluster'] && clusterIds.includes(sessionStorage['bcs-cluster'])) {
                                this.searchScope = sessionStorage['bcs-cluster']
                            } else {
                                this.searchScope = this.searchScopeList[0].id
                            }
                        }
                    }
                }
            },
            curClusterId () {
                this.searchScope = this.curClusterId
                this.clusterId = this.curClusterId
                this.fetchNamespaceList()
            }
        },
        async created () {
            await this.fetchNamespaceList()
        },
        destroyed () {
            this.bkMessageInstance && this.bkMessageInstance.close()
        },
        methods: {
            /**
             * 同步命名空间
             */
            async syncNamespace () {
                try {
                    await this.$store.dispatch('configuration/syncNamespace', {
                        projectId: this.projectId
                    })

                    // const me = this
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'success',
                        message: this.$t('同步命名空间任务已经启动，请稍后刷新'),
                        delay: 1000
                        // onClose: () => {
                        //     me.refresh()
                        // }
                    })
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
                    setTimeout(() => {
                        this.isInitLoading = false
                    }, 200)
                }
            },

            handleGoVar () {
                this.editNamespaceConf.isShow = false
                this.$router.push({
                    name: 'var',
                    params: {
                        projectCode: this.projectCode
                    }
                })
            },

            handleBlurRequestsMem (val) {
                this.quotaData.requestsMem = val
            },
            handleBlurRequestsCpu (val) {
                this.quotaData.requestsCpu = val
            },
            /**
             * 刷新列表
             */
            refresh () {
                this.pageConf.curPage = 1
                this.isPageLoading = true
                this.fetchNamespaceList()
            },

            /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
            changePageSize (pageSize) {
                this.pageConf.pageSize = pageSize
                this.pageConf.curPage = 1
                this.initPageConf()
                this.pageChange(this.pageConf.curPage)
            },

            /**
             * 搜索框清除事件
             */
            clearSearch () {
                this.search = ''
                this.handleSearch()
            },

            /**
             * 加载命名空间列表
             */
            async fetchNamespaceList () {
                try {
                    const res = await this.$store.dispatch('configuration/getNamespaceListByClusterId', {
                        projectId: this.projectId,
                        clusterId: this.searchScope
                    })
                    this.web_annotations = res.web_annotations || { perms: {} }

                    const list = []
                    res.data.forEach(item => {
                        item.isExpandLabels = false
                        // 是否显示标签的展开按钮
                        item.showExpand = false
                        list.push(item)
                    })

                    this.namespaceList.splice(0, this.namespaceList.length, ...list)
                    this.namespaceListTmp.splice(0, this.namespaceListTmp.length, ...list)
                    // this.initPageConf()
                    // this.curPageData = this.getDataByPage(this.pageConf.curPage)
                    this.handleSearch()

                    setTimeout(() => {
                        this.curPageData.forEach((item, index) => {
                            const real = this.$refs[`${this.pageConf.curPage}-real${index}`]
                            if (real) {
                                if (real.offsetHeight > 24 + 5) {
                                    item.showExpand = true
                                }
                            }
                        })
                    }, 100)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
                    setTimeout(() => {
                        this.isInitLoading = false
                    }, 200)
                }
            },

            /**
             * 初始化弹层翻页条
             */
            initPageConf () {
                const total = this.namespaceList.length
                this.pageConf.total = total
                this.pageConf.curPage = 1
                this.pageConf.totalPage = Math.ceil(total / this.pageConf.pageSize) || 1
            },

            /**
             * 翻页回调
             *
             * @param {number} page 当前页
             */
            pageChange (page) {
                this.pageConf.curPage = page
                const data = this.getDataByPage(page)
                this.curPageData.splice(0, this.curPageData.length, ...data)
                this.isPageLoading = true
                setTimeout(() => {
                    this.curPageData.forEach((item, index) => {
                        const real = this.$refs[`${this.pageConf.curPage}-real${index}`]
                        if (real) {
                            if (real.offsetHeight > 24 + 5) {
                                item.showExpand = true
                            }
                        }
                    })
                    this.isPageLoading = false
                }, 100)
            },

            /**
             * 获取当前这一页的数据
             *
             * @param {number} page 当前页
             *
             * @return {Array} 当前页数据
             */
            getDataByPage (page) {
                // 如果没有page，重置
                if (!page) {
                    this.pageConf.curPage = page = 1
                }
                let startIndex = (page - 1) * this.pageConf.pageSize
                let endIndex = page * this.pageConf.pageSize
                this.isPageLoading = true
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.namespaceList.length) {
                    endIndex = this.namespaceList.length
                }
                setTimeout(() => {
                    this.isPageLoading = false
                }, 200)
                return this.namespaceList.slice(startIndex, endIndex)
            },

            /**
             * 下拉框选择所属集群
             */
            chooseCluster (index, data) {
                this.showQuota = this.isSharedCluster
                const len = this.clusterList.length
                for (let i = len - 1; i >= 0; i--) {
                    if (String(this.clusterList[i].cluster_id) === String(data.cluster_id)) {
                        this.clusterId = data.cluster_id
                        break
                    }
                }
            },

            /**
             * 显示添加命名空间的 sideslider
             */
            async showAddNamespace () {
                this.showQuota = this.isSharedCluster
                this.addNamespaceConf.isShow = true
                this.clusterId = this.curClusterId ? this.curClusterId : ''

                try {
                    const res = await this.$store.dispatch('configuration/getNamespaceVariable', {
                        projectId: this.projectId,
                        namespaceId: 0
                    })
                    const variableList = []
                    ;(res.data || []).forEach(item => {
                        item.leftContent = `${item.name}(${item.key})`
                        variableList.push(item)
                    })

                    this.addNamespaceConf.variableList.splice(
                        0,
                        this.addNamespaceConf.variableList.length,
                        ...variableList
                    )
                } catch (e) {
                    console.error(e)
                } finally {
                    setTimeout(() => {
                        this.addNamespaceConf.loading = false
                    }, 300)
                }
            },

            /**
             * 添加命名空间 sideslder 取消按钮
             */
            hideAddNamespace () {
                this.addNamespaceConf.isShow = false
                this.addNamespaceConf.variableList.splice(0, this.addNamespaceConf.variableList.length, ...[])
                this.addNamespaceConf.namespaceName = ''
                this.clusterId = ''

                this.quotaData = Object.assign({}, {
                    limitsCpu: '400',
                    requestsCpu: '',
                    limitsMem: '400',
                    requestsMem: ''
                })
            },

            /**
             * 添加命名空间确认按钮
             */
            async confirmAddNamespace () {
                const namespaceName = this.addNamespaceConf.namespaceName
                if (!namespaceName.trim()) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请填写命名空间名称')
                    })
                    return
                }

                if (namespaceName.length < 2) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('命名空间名称不得小于2个字符')
                    })
                    return
                }

                if (!/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/g.test(namespaceName)) {
                    this.$bkMessage({
                        theme: 'error',
                        delay: 5000,
                        message: this.$t('命名空间名称只能包含小写字母、数字以及连字符(-)，连字符（-）后面必须接英文或者数字')
                    })
                    return
                }

                if (!this.clusterId) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        deplay: 5000,
                        message: this.$t('请选择所属集群')
                    })
                    return
                }

                const variableList = []
                const len = this.addNamespaceConf.variableList.length
                for (let i = 0; i < len; i++) {
                    const variable = this.addNamespaceConf.variableList[i]
                    variableList.push({
                        id: variable.id,
                        key: variable.key,
                        name: variable.name,
                        value: variable.value
                    })
                }

                try {
                    this.addNamespaceConf.loading = true
                    const params = {
                        projectId: this.projectId,
                        name: namespaceName,
                        cluster_id: this.clusterId,
                        ns_vars: variableList
                    }
                    if (this.showQuota) {
                        params.quota = {
                            'requests.cpu': this.quotaData.requestsCpu,
                            'requests.memory': this.quotaData.requestsMem + 'Gi',
                            'limits.cpu': this.quotaData.limitsCpu,
                            'limits.memory': this.quotaData.limitsMem + 'Gi'
                        }
                    }
                    await this.$store.dispatch('configuration/addNamespace', params)

                    this.hideAddNamespace()
                    setTimeout(() => {
                        this.fetchNamespaceList()
                    }, 300)
                } catch (e) {
                } finally {
                    this.addNamespaceConf.loading = false
                }
            },

            toggleShowQuota (v) {
                this.showQuota = v
                this.quotaData.requestsCpu = 1
            },

            /**
             * 文本块获取焦点事件回调，用于获取焦点时选中全部
             * input type=number 不支持 setSelectionRange
             *
             * @param {Object} e 事件对象
             */
            quotaInputFocusHandler (e) {
                e.stopPropagation()
                e.preventDefault()
                e.currentTarget.setSelectionRange(0, -1)
            },

            /**
             * custom-duration-input 文本框获 keydown 事件回调
             *
             * @param {Object} e 事件对象
             */
            quotaInputKeydownHandler (e) {
                const keyCode = e.keyCode

                // 键盘按下不允许的按钮
                if (this.validKeyCodeList4QuotaInput.indexOf(keyCode) < 0) {
                    e.stopPropagation()
                    e.preventDefault()
                    return false
                }
            },

            /**
             * 显示修改命名空间的 sideslider
             *
             * @param {Object} ns 当前 namespace 对象
             * @param {number} index 当前 namespace 对象的索引
             */
            async showEditNamespace (ns, index) {
                this.editNamespaceConf.isShow = true
                // this.editNamespaceConf.loading = true
                this.editNamespaceConf.namespaceName = this.isSharedCluster ? this.filterNamespace(ns.name) : ns.name
                this.editNamespaceConf.title = this.$t('修改命名空间：{nsName}', {
                    nsName: ns.name
                })
                this.editNamespaceConf.ns = Object.assign({}, ns)
                this.clusterId = ns.cluster_id

                // mesos 可以修改命名空间名称和所属集群
                this.editNamespaceConf.canEdit = false

                try {
                    const res = await this.$store.dispatch('configuration/getNamespaceVariable', {
                        projectId: this.projectId,
                        namespaceId: ns.id
                    })
                    const variableList = []
                    ;(res.data || []).forEach(item => {
                        item.leftContent = `${item.name}(${item.key})`
                        variableList.push(item)
                    })

                    this.editNamespaceConf.variableList.splice(
                        0,
                        this.editNamespaceConf.variableList.length,
                        ...variableList
                    )
                } catch (e) {
                    console.error(e)
                } finally {
                    setTimeout(() => {
                        this.editNamespaceConf.loading = false
                    }, 300)
                }
            },

            /**
             * 修改命名空间 sideslder 取消按钮
             */
            hideEditNamespace () {
                this.editNamespaceConf.isShow = false
                this.editNamespaceConf.variableList.splice(0, this.editNamespaceConf.variableList.length, ...[])
                this.editNamespaceConf.namespaceName = ''
                this.editNamespaceConf.title = ''
                this.editNamespaceConf.canEdit = false
                this.editNamespaceConf.ns = Object.assign({}, {})
                this.clusterId = ''

                this.quotaData = Object.assign({}, {
                    limitsCpu: '400',
                    requestsCpu: '',
                    limitsMem: '400',
                    requestsMem: ''
                })
            },

            /**
             * 修改命名空间确认按钮
             */
            async confirmEditNamespace () {
                const namespaceName = this.editNamespaceConf.namespaceName
                if (!namespaceName.trim()) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请填写命名空间名称')
                    })
                    return
                }

                if (namespaceName.length < 2) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('命名空间名称不得小于2个字符')
                    })
                    return
                }
                if (!/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/g.test(namespaceName)) {
                    this.$bkMessage({
                        theme: 'error',
                        delay: 5000,
                        message: this.$t('命名空间名称只能包含小写字母、数字以及连字符(-)，连字符（-）后面必须接英文或者数字')
                    })
                    return
                }

                if (!this.clusterId) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择所属集群')
                    })
                    return
                }

                const variableList = []
                const len = this.editNamespaceConf.variableList.length
                for (let i = 0; i < len; i++) {
                    const variable = this.editNamespaceConf.variableList[i]
                    variableList.push({
                        id: variable.id,
                        key: variable.key,
                        name: variable.name,
                        value: variable.value
                    })
                }

                try {
                    this.editNamespaceConf.loading = true
                    await this.$store.dispatch('configuration/editNamespace', {
                        projectId: this.projectId,
                        cluster_id: this.clusterId,
                        name: namespaceName,
                        namespaceId: this.editNamespaceConf.ns.id,
                        ns_vars: variableList
                    })

                    this.hideEditNamespace()
                    setTimeout(() => {
                        this.fetchNamespaceList()
                    }, 300)
                } catch (e) {
                } finally {
                    this.editNamespaceConf.loading = false
                }
            },

            /**
             * 显示修改配额的 sideslider
             *
             * @param {Object} ns 当前 namespace 对象
             * @param {number} index 当前 namespace 对象的索引
             */
            async showEditQuota (ns, index) {
                this.showQuotaData = ns
                this.editQuotaConf.isShow = true
                this.editQuotaConf.loading = true
                this.editQuotaConf.initLoading = true
                this.editQuotaConf.namespaceName = ns.name
                this.editQuotaConf.title = this.$t('配额管理：{nsName}', {
                    nsName: ns.name
                })
                this.editQuotaConf.ns = Object.assign({}, ns)
                this.clusterId = ns.cluster_id

                try {
                    const res = await this.$store.dispatch('configuration/getQuota', {
                        projectId: this.projectId,
                        namespaceName: ns.name,
                        clusterId: ns.cluster_id
                    })
                    const hard = res.data.quota.hard || {}
                    this.quotaData = Object.assign({}, {
                        limitsCpu: '400',
                        requestsCpu: hard['requests.cpu'] ? Number(hard['requests.cpu']) : 0,
                        limitsMem: '400',
                        requestsMem: hard['requests.memory'] ? Number(hard['requests.memory'].split('Gi')[0]) : 0
                    })
                } catch (e) {
                    console.error(e)
                } finally {
                    this.editQuotaConf.initLoading = false
                    this.editQuotaConf.loading = false
                }
            },

            /**
             * 修改配额 sideslder 取消按钮
             */
            hideEditQuota () {
                this.editQuotaConf.isShow = false
                this.editQuotaConf.namespaceName = ''
                this.editQuotaConf.title = ''
                this.editQuotaConf.ns = Object.assign({}, {})
                this.clusterId = ''

                this.quotaData = Object.assign({}, {
                    limitsCpu: '400',
                    requestsCpu: '',
                    limitsMem: '400',
                    requestsMem: ''
                })
            },

            /**
             * 修改配额确认按钮
             */
            async confirmEditQuota () {
                // if (!this.validQuota()) {
                //     return
                // }

                const namespaceName = this.editQuotaConf.namespaceName
                try {
                    this.editQuotaConf.loading = true
                    await this.$store.dispatch('configuration/editQuota', {
                        projectId: this.projectId,
                        clusterId: this.clusterId,
                        namespaceName: namespaceName,
                        data: {
                            quota: {
                                'requests.cpu': this.quotaData.requestsCpu,
                                'requests.memory': this.quotaData.requestsMem + 'Gi',
                                'limits.cpu': this.quotaData.limitsCpu,
                                'limits.memory': this.quotaData.limitsMem + 'Gi'
                            }
                        }
                    })

                    this.hideEditQuota()
                    setTimeout(() => {
                        this.fetchNamespaceList()
                    }, 300)
                } catch (e) {
                    console.log(e)
                } finally {
                    this.editQuotaConf.loading = false
                }
            },

            /**
             * 显示删除配额确认框
             *
             * @param {Object} ns 当前 namespace 对象
             * @param {number} index 当前 namespace 对象的索引
             */
            async showDelQuota (ns, index) {
                this.delQuotaDialogConf.isShow = true
                this.delQuotaDialogConf.ns = Object.assign({}, this.showQuotaData)
            },

            /**
             * 删除当前 namespace 的配额
             *
             * @param {Object} ns 当前 namespace 对象
             * @param {number} index 当前 namespace 对象的索引
             */
            async delQuotaConfirm () {
                try {
                    this.isPageLoading = true
                    this.delQuotaCancel()
                    await this.$store.dispatch('configuration/delQuota', {
                        projectId: this.projectId,
                        clusterId: this.delQuotaDialogConf.ns.cluster_id,
                        namespaceName: this.delQuotaDialogConf.ns.name
                    })
                    this.search = ''
                    this.refresh()
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'success',
                        message: this.$t('删除Namespace成功')
                    })
                    this.editQuotaConf.isShow = false
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 取消删除当前 namespace
             *
             * @param {Object} ns 当前 namespace 对象
             * @param {number} index 当前 namespace 对象的索引
             */
            delQuotaCancel () {
                this.delQuotaDialogConf.isShow = false
                setTimeout(() => {
                    this.delQuotaDialogConf.ns = Object.assign({}, {})
                }, 300)
            },

            /**
             * 显示删除 namespace 确认框
             *
             * @param {Object} ns 当前 namespace 对象
             * @param {number} index 当前 namespace 对象的索引
             */
            async showDelNamespace (ns, index) {
                this.delNamespaceDialogConf.isShow = true
                this.delNamespaceDialogConf.ns = Object.assign({}, ns)
            },

            /**
             * 删除当前 namespace
             *
             * @param {Object} ns 当前 namespace 对象
             * @param {number} index 当前 namespace 对象的索引
             */
            async delNamespaceConfirm () {
                try {
                    this.isPageLoading = true
                    this.delNamespaceCancel()
                    await this.$store.dispatch('configuration/delNamespace', {
                        projectId: this.projectId,
                        namespaceId: this.delNamespaceDialogConf.ns.id
                    })
                    this.search = ''
                    this.refresh()
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'success',
                        message: this.$t('删除Namespace成功')
                    })
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 取消删除当前 namespace
             *
             * @param {Object} ns 当前 namespace 对象
             * @param {number} index 当前 namespace 对象的索引
             */
            delNamespaceCancel () {
                this.delNamespaceDialogConf.isShow = false
                setTimeout(() => {
                    this.delNamespaceDialogConf.ns = Object.assign({}, {})
                }, 300)
            },

            /**
             * 搜索事件
             */
            handleSearch () {
                const search = String(this.search || '').trim().toLowerCase()
                let list = JSON.parse(JSON.stringify(this.namespaceListTmp))

                if (this.searchScope) {
                    list = list.filter(item => {
                        return item.cluster_id === this.searchScope
                    })
                }

                const results = list.filter(ns => {
                    // const envType = String(ns.env_type || '').toLowerCase()
                    // || envType.indexOf(search) > -1
                    const name = String(ns.name || '').toLowerCase()
                    const clusterName = String(ns.cluster_name || '').toLowerCase()

                    return name.indexOf(search) > -1
                        || clusterName.indexOf(search) > -1
                })
                // const beforeLen = this.namespaceListTmp.length
                // const afterLen = results.length
                this.namespaceList.splice(0, this.namespaceList.length, ...results)
                // this.pageConf.curPage = beforeLen !== afterLen ? 1 : this.pageConf.curPage
                this.initPageConf()
                this.curPageData = this.getDataByPage(this.pageConf.curPage)
            },
            filterNamespace (name) {
                const filterRule = this.projectCode + '-'
                return name.split(filterRule)[1]
            }
        }
    }
</script>

<style scoped>
    @import './namespace.css';
</style>
