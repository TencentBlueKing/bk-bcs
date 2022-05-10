<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-service-title">
                Service
                <span class="biz-tip ml10">{{$t('请通过模板集或Helm创建Service')}}</span>
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper" style="padding: 0;" v-bkloading="{ isLoading: isInitLoading, opacity: 0.1 }">
            <app-exception
                v-if="exceptionCode && !isInitLoading"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>

            <template v-if="!exceptionCode && !isInitLoading">
                <div class="biz-panel-header">
                    <div class="left">
                        <bk-button @click.stop.prevent="removeServices" v-if="curPageData.length">
                            <span>{{$t('批量删除')}}</span>
                        </bk-button>
                        <bk-button style="opacity: 0">
                            <span>.</span>
                        </bk-button>
                    </div>
                    <div class="right">
                        <bk-data-searcher
                            :placeholder="$t('输入关键字，按Enter搜索')"
                            :scope-list="searchScopeList"
                            :search-key.sync="searchKeyword"
                            :search-scope.sync="searchScope"
                            :cluster-fixed="!!curClusterId"
                            @search="getServiceList"
                            @refresh="refresh">
                        </bk-data-searcher>
                    </div>
                </div>

                <div class="biz-service">
                    <div class="biz-table-wrapper" style="border: none;">
                        <bk-table
                            :size="'medium'"
                            :data="curPageData"
                            :pagination="pageConf"
                            v-bkloading="{ isLoading: isPageLoading && !isInitLoading }"
                            @page-limit-change="handlePageLimitChange"
                            @page-change="handlePageChange"
                            @select="handlePageSelect"
                            @select-all="handlePageSelectAll">
                            <bk-table-column type="selection" width="60" :selectable="rowSelectable"></bk-table-column>
                            <bk-table-column :label="$t('名称')" :show-overflow-tooltip="true" min-width="200">
                                <template slot-scope="props">
                                    <a href="javascript: void(0)"
                                        class="bk-text-button"
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
                                        @click.stop.prevent="showServiceDetail(props.row, index)"
                                    >{{props.row.resourceName ? props.row.resourceName : '--'}}</a>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('所属集群/命名空间')" min-width="200">
                                <template slot-scope="props">
                                    <div class="lh20">
                                        <span :class="['cluster-namespace-source', { en: isEn }]">{{$t('所属集群: ')}}</span>
                                        <bcs-popover :content="props.row.clusterId || '--'" placement="top">
                                            <div class="cluster-name">{{curCluster ? curCluster.clusterName : '--'}}</div>
                                        </bcs-popover>
                                    </div>
                                    <div class="lh20">
                                        <span :class="['cluster-namespace-source', { en: isEn }]">{{$t('命名空间: ')}}</span>
                                        {{props.row.namespace ? props.row.namespace : '--'}}
                                    </div>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('类型')" min-width="150">
                                <template slot-scope="props">
                                    {{props.row.data.spec.type}}
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('集群内访问')" min-width="200" :show-overflow-tooltip="false">
                                <template slot-scope="props">
                                    <bcs-popover>
                                        <template v-for="(internal, internalIndex) in props.row.accessInternal">
                                            <div :key="internalIndex" v-if="internalIndex < 2">
                                                {{internal}}
                                            </div>
                                        </template>
                                        <div v-if="props.row.accessExternal.length > 2">
                                            ...
                                        </div>
                                        <template slot="content">
                                            <div v-for="(internal, internalIndex) in props.row.accessInternal" :key="internalIndex">
                                                {{internal}}
                                            </div>
                                        </template>
                                    </bcs-popover>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('集群外访问')" min-width="150" :show-overflow-tooltip="false">
                                <template slot-scope="props">
                                    <bcs-popover>
                                        <template v-for="(external, externalIndex) in props.row.accessExternal">
                                            <div :key="externalIndex" v-if="externalIndex < 2">
                                                {{external}}
                                            </div>
                                        </template>
                                        <div v-if="props.row.accessExternal.length > 2">
                                            ...
                                        </div>
                                        <template slot="content">
                                            <div v-for="(external, externalIndex) in props.row.accessExternal" :key="externalIndex">
                                                {{external}}
                                            </div>
                                        </template>
                                    </bcs-popover>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('操作')" width="240">
                                <template slot-scope="props">
                                    <template v-if="props.row.can_update">
                                        <a href="javascript:void(0);"
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
                                            @click="showUpdateServicePanel(props.row)"
                                        >{{$t('更新')}}</a>
                                    </template>
                                    <template v-else>
                                        <bcs-popover :content="props.row.can_update_msg" placement="left">
                                            <a href="javascript:void(0);" class="bk-text-button is-disabled">{{$t('更新')}}</a>
                                        </bcs-popover>
                                    </template>
                                    <template v-if="props.row.can_delete">
                                        <a href="javascript:void(0);"
                                            class="bk-text-button ml15"
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
                                            @click="removeService(props.row)"
                                        >{{$t('删除')}}</a>
                                    </template>
                                    <template v-else>
                                        <bcs-popover :content="props.row.can_delete_msg" placement="left" style="margin-left: 15px;">
                                            <a href="javascript:void(0);" class="bk-text-button is-disabled">{{$t('删除')}}</a>
                                        </bcs-popover>
                                    </template>
                                    <bcs-button text class="ml15 cl5-router"
                                        :disabled="props.row.access_info.external.CL5"
                                        v-if="$INTERNAL"
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
                                        @click="showCreateCL5(props.row)">
                                        {{$t('CL5路由')}}
                                    </bcs-button>
                                </template>
                            </bk-table-column>
                        </bk-table>
                    </div>
                </div>
            </template>

            <bk-sideslider
                v-if="curServiceDetail"
                :quick-close="false"
                :is-show.sync="updateServiceSliderConf.isShow"
                :title="updateServiceSliderConf.title"
                :width="'700'">
                <div class="p30" slot="content" v-bkloading="{ isLoading: isDetailLoading }" style="overflow: hidden;">
                    <div class="bk-form bk-form-vertical">
                        <div class="bk-form-item">
                            <label class="bk-label">{{$t('名称')}}：</label>
                            <div class="bk-form-content">
                                <bk-input
                                    :placeholder="$t('请输入')"
                                    v-model="curServiceDetail.name"
                                    disabled style="width: 600px;" />
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label">{{$t('关联应用')}}：</label>
                            <div class="bk-form-content" style="width: 600px;">
                                <bk-selector
                                    :placeholder="$t('请选择要关联的Application')"
                                    :setting-key="'deploy_tag'"
                                    :multi-select="true"
                                    :display-key="'deploy_name'"
                                    :selected.sync="curServiceDetail.config.webCache.link_app"
                                    :list="applicationList"
                                    :init-prevent-trigger="true"
                                    :is-link="true"
                                    @item-selected="selectAppsHandler">
                                </bk-selector>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label">{{$t('关联标签')}}：</label>
                            <div class="bk-form-content key-tip-wrapper" style="width: 705px;">
                                <ul class="key-list" v-if="appLabels.length && !isLabelsLoading">
                                    <li v-for="(label, index) in appLabels" @click="selectLabel(label)" :key="index">
                                        <span class="key">
                                            <bk-checkbox class="pb0 pt0" name="linkapp" :value="label.isSelected"></bk-checkbox>
                                        </span>
                                        <span class="value">{{label.key}}:{{label.value}}</span>
                                    </li>
                                </ul>
                                <div v-else-if="!isLabelsLoading" class="key-tip">
                                    {{curServiceDetail.config.webCache.link_app ? $t('关联的应用没有公共的标签（注：Key、Value都相同的标签为公共标签）') : $t('请先关联应用')}}
                                </div>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label">{{$t('Service类型')}}：</label>
                            <div class="bk-form-content">
                                <bk-selector
                                    style="width: 360px;"
                                    :disabled="true"
                                    :placeholder="$t('请选择')"
                                    :setting-key="'id'"
                                    :display-key="'name'"
                                    :selected.sync="curServiceDetail.config.spec.type"
                                    :list="serviceTypeList">
                                </bk-selector>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label">{{$t('端口映射')}}：</label>
                            <div class="bk-form-content">
                                <div class="biz-keys-list mt10">
                                    <template v-if="curServiceDetail.config.webCache.link_app.length">
                                        <template v-if="appPortList.length && curServiceDetail.config.spec.ports.length">
                                            <table class="biz-simple-table">
                                                <thead>
                                                    <tr>
                                                        <th style="width: 100px;">{{$t('端口名称')}}</th>
                                                        <th style="width: 100px;">{{$t('端口')}}</th>
                                                        <th style="width: 120px;">{{$t('协议')}}</th>
                                                        <th style="width: 120px;">{{$t('目标端口')}}</th>
                                                        <th style="width: 100px;">NodePort</th>
                                                        <th></th>
                                                    </tr>
                                                </thead>
                                                <tbody>
                                                    <tr v-for="(port, index) in curServiceDetail.config.spec.ports" :key="index">
                                                        <td>
                                                            <bk-input class="list-input" :placeholder="$t('请输入')" v-model="port.name" />
                                                        </td>
                                                        <td>
                                                            <bk-number-input
                                                                :value.sync="port.port"
                                                                :min="1"
                                                                :max="65535"
                                                                :hide-operation="true"
                                                                :ex-style="{ 'width': '100px' }"
                                                                :placeholder="$t('请输入')">
                                                            </bk-number-input>
                                                        </td>
                                                        <td>
                                                            <bk-selector
                                                                :placeholder="$t('协议')"
                                                                :setting-key="'id'"
                                                                :allow-clear="true"
                                                                :selected.sync="port.protocol"
                                                                :list="protocolList">
                                                            </bk-selector>
                                                        </td>
                                                        <td>
                                                            <!-- {{appPortList}} -->
                                                            <bk-selector
                                                                :placeholder="$t('请输入')"
                                                                :setting-key="'name'"
                                                                :display-key="'name'"
                                                                :selected.sync="port.targetPort"
                                                                :allow-clear="true"
                                                                :filter-list="curServicePortList"
                                                                :is-link="true"
                                                                :init-prevent-trigger="true"
                                                                :list="appPortList"
                                                                @clear="clearPortHandler(port)">
                                                            </bk-selector>
                                                        </td>
                                                        <td>
                                                            <bk-number-input
                                                                :disabled="curServiceDetail.config.spec.type !== 'NodePort'"
                                                                :value.sync="port.nodePort"
                                                                :min="0"
                                                                :max="32767"
                                                                :hide-operation="true"
                                                                :ex-style="{ 'width': '100px' }"
                                                                :placeholder="$t('请输入')">
                                                            </bk-number-input>
                                                        </td>
                                                        <td>
                                                            <bk-button class="action-btn" @click.stop.prevent="addPort" v-show="curServiceDetail.config.spec.ports.length < appPortList.length">
                                                                <i class="bcs-icon bcs-icon-plus"></i>
                                                            </bk-button>
                                                            <bk-button class="action-btn" @click.stop.prevent="removePort(port, index)" v-show="curServiceDetail.config.spec.ports.length > 1">
                                                                <i class="bcs-icon bcs-icon-minus"></i>
                                                            </bk-button>
                                                        </td>
                                                    </tr>
                                                </tbody>
                                            </table>
                                        </template>
                                        <template v-else>
                                            <p class="mt5">
                                                <router-link :to="{ name: 'k8sTemplatesetDeployment', params: { templateId: curServiceDetail.template_id } }" class="bk-text-button">{{$t('点此')}}</router-link>{{$t('去模板集配置端口映射信息')}}
                                            </p>
                                        </template>
                                    </template>
                                    <template v-else>
                                        <p class="mt5">{{$t('请关联相应的应用')}}</p>
                                    </template>
                                </div>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label">{{$t('标签管理')}}：</label>
                            <div class="bk-form-content">
                                <bk-keyer :key-list.sync="curLabelList" ref="labelKeyer" @change="updateLabelList"></bk-keyer>
                            </div>
                        </div>

                        <div class="bk-form-item">
                            <label class="bk-label">{{$t('注解管理')}}：</label>
                            <div class="bk-form-content">
                                <bk-keyer :key-list.sync="curRemarkList" ref="labelKeyer" @change="updateRemarkList"></bk-keyer>
                            </div>
                        </div>

                        <div class="bk-form-item mt25">
                            <bk-button type="primary" :loading="isDetailSaving" @click.stop.prevent="saveServiceDetail">{{$t('保存并更新')}}</bk-button>
                            <bk-button :disabled="isDetailSaving" @click.stop.prevent="hideServiceSlider">{{$t('取消')}}</bk-button>
                        </div>
                    </div>
                </div>
            </bk-sideslider>

            <bk-sideslider
                v-if="curService"
                :quick-close="true"
                :is-show.sync="serviceSlider.isShow"
                :title="serviceSlider.title"
                :width="'800'">
                <div class="p30" slot="content" v-bkloading="{ isLoading: isEndpointLoading }">
                    <p class="data-title">
                        {{$t('基础信息')}}
                    </p>
                    <div class="biz-metadata-box">
                        <div class="data-item" style="width: 260px;">
                            <p class="key">{{$t('选择器')}}：</p>
                            <p class="value" v-bk-tooltips="{ direction: 'top', content: selector }">
                                {{selector ? selector : '--'}}
                            </p>
                        </div>
                        <div class="data-item">
                            <p class="key">{{$t('类型')}}：</p>
                            <p class="value">{{curService.data.spec.type}}</p>
                        </div>
                        <div class="data-item">
                            <p class="key">Service IP：</p>
                            <p class="value">{{curService.data.spec && curService.data.spec.clusterIP ? curService.data.spec.clusterIP : '--'}}</p>
                        </div>
                        <div class="data-item">
                            <p class="key">Service Domain：</p>
                            <p class="value">--</p>
                        </div>
                    </div>
                    <p class="data-title">
                        {{$t('端口映射')}}
                    </p>
                    <table class="bk-table biz-data-table">
                        <thead>
                            <tr>
                                <th>{{$t('端口名称')}}</th>
                                <th>{{$t('端口')}}</th>
                                <th>{{$t('协议')}}</th>
                                <th>{{$t('目标端口')}}</th>
                                <th>NodePort</th>
                            </tr>
                        </thead>
                        <tbody>
                            <template v-if="curService.data.spec.ports && curService.data.spec.ports.length">
                                <tr v-for="(port, index) in curService.data.spec.ports" :key="index">
                                    <td>{{port.name ? port.name : '--'}}</td>
                                    <td>{{port.port ? port.port : '--'}}</td>
                                    <td>
                                        {{port.protocol ? port.protocol : '--'}}
                                    </td>
                                    <td>
                                        {{port.targetPort ? port.targetPort : '--'}}
                                    </td>
                                    <td>
                                        {{port.nodePort ? port.nodePort : '--'}}
                                    </td>
                                </tr>
                            </template>
                            <template v-else>
                                <tr>
                                    <td colspan="5">
                                        <div class="biz-app-list">
                                            <div class="bk-message-box" style="min-height: auto;">
                                                <bcs-exception type="empty" scene="part"></bcs-exception>
                                            </div>
                                        </div>
                                    </td>
                                </tr>
                            </template>
                        </tbody>
                    </table>

                    <p class="data-title">
                        Endpoints
                    </p>
                    <table class="bk-table biz-data-table">
                        <thead>
                            <tr>
                                <th>{{$t('名称')}}</th>
                                <th>Pod IP</th>
                                <th>Node IP</th>
                            </tr>
                        </thead>
                        <tbody>
                            <template v-if="endpoints.length">
                                <tr v-for="(point, index) in endpoints" :key="index">
                                    <td>{{point.targetRef && point.targetRef.name ? point.targetRef.name : '--'}}</td>
                                    <td>
                                        {{point.ip ? point.ip : '--'}}
                                    </td>
                                    <td>
                                        {{point.nodeName ? getNodeIpByName(point.nodeName) : '--'}}
                                    </td>
                                </tr>
                            </template>

                            <template v-else>
                                <tr>
                                    <td colspan="3">
                                        <div class="biz-app-list">
                                            <div class="bk-message-box" style="min-height: auto;">
                                                <bcs-exception type="empty" scene="part"></bcs-exception>
                                            </div>
                                        </div>
                                    </td>
                                </tr>
                            </template>
                        </tbody>
                    </table>

                    <p class="data-title">
                        {{$t('其他信息')}}
                    </p>
                    <div class="biz-metadata-box">
                        <div class="data-item" style="width: 260px;">
                            <p class="key">{{$t('创建时间')}}：</p>
                            <p class="value">
                                {{curService.createTime ? formatDate(curService.createTime) : '--'}}
                            </p>
                        </div>
                        <div class="data-item">
                            <p class="key">{{$t('更新时间')}}：</p>
                            <p class="value">
                                {{curService.updateTime ? formatDate(curService.updateTime) : '--'}}
                            </p>
                        </div>
                        <div class="data-item">
                            <p class="key">{{$t('更新人')}}：</p>
                            <p class="value">{{curService.updator ? curService.updator : '--'}}</p>
                        </div>
                        <div class="data-item">
                            <p class="key">{{$t('来源')}}：</p>
                            <p class="value">{{curService.source_type ? curService.source_type : '--'}}</p>
                        </div>
                    </div>

                    <div class="data-title">
                        {{$t('标签')}}
                    </div>
                    <div class="point-box">
                        <template v-if="labelList.length">
                            <ul class="key-list">
                                <li v-for="(label, index) in labelList" :key="index">
                                    <span class="key">{{label[0]}}</span>
                                    <span class="value">{{label[1]}}</span>
                                </li>
                            </ul>
                        </template>
                        <template v-else>
                            <div class="bk-message-box" style="min-height: auto;">
                                <bcs-exception type="empty" scene="part"></bcs-exception>
                            </div>
                        </template>
                    </div>
                </div>
            </bk-sideslider>

            <bk-dialog
                :title="$t('确认删除')"
                :is-show="batchDialogConfig.isShow"
                :width="600"
                :has-header="false"
                :quick-close="false"
                @confirm="deleteServices(batchDialogConfig.data)"
                @cancel="batchDialogConfig.isShow = false">
                <template slot="content">
                    <div class="biz-batch-wrapper">
                        <p class="batch-title">{{$t('确定要删除以下Service？')}}</p>
                        <ul class="batch-list">
                            <li v-for="(item, index) of batchDialogConfig.list" :key="index">{{item}}</li>
                        </ul>
                    </div>
                </template>
            </bk-dialog>

            <bk-dialog
                class="createcl5-dialog"
                :title="$t('新建CL5')"
                :is-show="createCL5DialogConf.isShow"
                :width="500"
                :quick-close="false"
                @cancel="createCL5DialogConf.isShow = false">
                <template slot="content">
                    <div class="biz-batch-wrapper">
                        <bcs-alert type="info" class="mb15">
                            <div slot="title">
                                <i18n path="具体使用方法参考{article}">
                                    <a place="article" :href="PROJECT_CONFIG.doc.cl5" target="_blank" class="bk-text-button ml5">{{$t('文章')}}</a>
                                </i18n>
                            </div>
                        </bcs-alert>
                        <bk-form ref="formRef" :label-width="labelWidth">
                            <bk-form-item :label="$t('是否有sid')">
                                <bk-radio-group v-model="createCL5DialogConf.isHasSid">
                                    <bk-radio value="yes">{{$t('是')}}</bk-radio>
                                    <bk-radio value="no">{{$t('否')}}</bk-radio>
                                </bk-radio-group>
                            </bk-form-item>
                            <bk-form-item :label="createCL5DialogConf.isHasSid === 'yes' ? 'sid' : $t('业务模块')">
                                <div style="display: flex;">
                                    <bk-input maxlength="60"
                                        :placeholder="createCL5DialogConf.isHasSid === 'yes' ? $t('请输入sid') : $t('请输入业务模块')"
                                        v-model="createCL5DialogConf.bid4" />
                                    <bcs-popover v-if="createCL5DialogConf.isHasSid === 'no'" :content="$t('Node节点所属的CMDB三级业务模块ID')" placement="right">
                                        <i class="bcs-icon bcs-icon-info-circle" style="margin-left: 5px;"></i>
                                    </bcs-popover>
                                </div>
                            </bk-form-item>
                            <bk-form-item :label="$t('责任人')">
                                <span>{{userInfo.username}}</span>
                            </bk-form-item>
                        </bk-form>
                    </div>
                </template>
                <div slot="footer">
                    <template>
                        <div class="bk-dialog-outer" style="line-height: 56px; height: 56px;">
                            <div style="float: right;">
                                <bk-button type="primary" :loading="isCreatingCL5" @click="createCl5">
                                    {{$t('确定')}}
                                </bk-button>
                                <bk-button type="button" :disabled="isCreatingCL5" @click="hideCreateCL5">
                                    {{$t('取消')}}
                                </bk-button>
                            </div>
                        </div>
                    </template>
                </div>
            </bk-dialog>
        </div>
    </div>
</template>

<script>
    import mixin from './mixin'
    import { catchErrorHandler, formatDate } from '@/common/util'
    import useFormLabel from '@/common/use-form-label'

    export default {
        mixins: [mixin],
        data () {
            return {
                formatDate: formatDate,
                isInitLoading: true,
                isPageLoading: false,
                exceptionCode: null,
                isEndpointLoading: true,
                searchKeyword: '',
                searchScope: '',
                curPageData: [],
                appLabels: [],
                appIds: [],
                curService: null,
                curVersion: 0,
                curServiceDetail: null,
                isDetailLoading: true,
                curTotalPercent: 0,
                isLoadBalanceEdited: false,
                isDetailSaving: false,
                isLabelsLoading: true,
                statusTimer: [],
                pageConf: {
                    total: 0,
                    totalPage: 1,
                    pageSize: 10,
                    curPage: 1,
                    show: true,
                    current: 1,
                    count: 18,
                    limit: 10
                },
                isWeightError: true,
                serviceSlider: {
                    title: '',
                    isShow: false
                },
                updateServiceSliderConf: {
                    title: this.$t('更新Service'),
                    isShow: false
                },
                isLabelsShow: false,
                serviceTypeList: [
                    {
                        id: 'ClusterIP',
                        name: 'ClusterIP'
                    },
                    {
                        id: 'NodePort',
                        name: 'NodePort'
                    },
                    {
                        id: 'LoadBalancer',
                        name: 'LoadBalancer'
                    }
                ],
                serviceType: -1,
                ports: [
                    {
                        name: '',
                        protocol: 'http',
                        nodePort: '',
                        domainName: '',
                        path: '',
                        servicePort: ''
                    }
                ],
                protocolIndex: -1,
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
                appPortList: [],
                algorithmList: [
                    {
                        id: 'roundrobin',
                        name: 'roundrobin'
                    },
                    {
                        id: 'source',
                        name: 'source'
                    },
                    {
                        id: 'leastconn',
                        name: 'leastconn'
                    }
                ],
                curServicePortMap: [],
                algorithmIndex: -1,
                loadBalanceList: [],
                createCL5DialogConf: {
                    isShow: false,
                    bid4: '',
                    service: {},
                    isHasSid: 'no'
                },
                isCreatingCL5: false,
                serviceSelectedList: [],
                webAnnotations: { perms: {} }
            }
        },
        computed: {
            isEn () {
                return this.$store.state.isEn
            },
            applicationList () {
                return this.$store.state.k8sTemplate.linkApplications
            },
            curRemarkList () {
                const list = []
                const annotations = this.curServiceDetail?.config.metadata.annotations || {}
                // 如果有缓存直接使用
                // if (this.curServiceDetail.config.webCache && this.curServiceDetail.config.webCache.remarkList) {
                //     return this.curServiceDetail.config.webCache.remarkList
                // }
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
            curServicePortList () {
                const results = []
                const ports = this.curServiceDetail?.config.spec.ports || []
                ports.forEach(item => {
                    results.push(item.targetPort)
                })
                return results
            },
            isClusterDataReady () {
                return this.$store.state.cluster.isClusterDataReady
            },
            curClusterId () {
                return this.$store.state.curClusterId
            },
            userInfo () {
                return this.$store.state.user
            },
            curCluster () {
                const list = this.$store.state.cluster.clusterList || []
                return list.find(item => item.clusterID === this.searchScope)
            }
        },
        watch: {
            isClusterDataReady: {
                immediate: true,
                handler (val) {
                    if (val) {
                        setTimeout(() => {
                            if (this.searchScopeList.length) {
                                const clusterIds = this.searchScopeList.map(item => item.id)
                                // 使用当前缓存
                                if (sessionStorage['bcs-cluster'] && clusterIds.includes(sessionStorage['bcs-cluster'])) {
                                    this.searchScope = sessionStorage['bcs-cluster']
                                } else {
                                    this.searchScope = this.searchScopeList[0].id
                                }
                            }
                            this.getServiceList()
                        }, 1000)
                    }
                }
            }
            // curClusterId () {
            //     this.searchScope = this.curClusterId
            //     this.getServiceList()
            // }
        },
        created () {
            this.initPageConf()
        },
        mounted () {
            const { initFormLabelWidth } = useFormLabel()
            this.labelWidth = initFormLabelWidth(this.$refs.formRef)
        },
        methods: {
            /**
             * 选择标签
             * @param  {object} labels 标签
             */
            selectLabel (label) {
                label.isSelected = !label.isSelected
                this.updateLabels()
            },

            updateLabels () {
                this.curServiceDetail.config.webCache.link_labels = []
                this.curServiceDetail.config.spec.selector = {}
                this.appLabels.forEach(label => {
                    if (label.isSelected) {
                        this.curServiceDetail.config.webCache.link_labels.push(label.id)
                        this.curServiceDetail.config.spec.selector[label.key] = label.value
                    }
                })
            },

            /**
             * 通过nodeName获取nodeIP eg: ip-10-235-46-109-n-bcs-k8s-15007
             * @param {string} nodeName
             * @return {string} nodeIP
             */
            getNodeIpByName (nodeName) {
                const reg = /ip-(.+)-n\w*/
                const match = nodeName.match(reg)
                let ip = '--'
                if (match && match[1]) {
                    ip = match[1].replace(/-/g, '.')
                }
                return ip
            },

            /**
             * service兼容处理
             * @param  {object} service sertice
             */
            formatService (service) {
                // 数据兼容处理
                if (!service.app_id) {
                    service.app_id = {}
                }

                if (!service.config.webCache) {
                    service.config.webCache = {
                        link_app: [],
                        link_labels: []
                    }
                }

                if (!service.config.webCache.link_app.length) {
                    service.config.webCache.link_app = []
                    service.deploy_tag_list.forEach(item => {
                        service.config.webCache.link_app.push(item)
                    })
                }

                if (!service.config.spec.ports) {
                    service.config.spec.ports = []
                }
                if (service.config.spec.ports && !service.config.spec.ports.length) {
                    service.config.spec.ports.push({
                        'id': '',
                        'name': '',
                        'port': '',
                        'protocol': 'TCP',
                        'targetPort': '',
                        'nodePort': ''
                    })
                } else {
                    service.config.spec.ports.forEach(port => {
                        if (!port.nodePort) {
                            port.nodePort = ''
                        }
                    })
                }
            },

            /**
             * 显示更新侧面板
             * @param  {object} service service
             */
            async showUpdateServicePanel (service) {
                const projectId = this.projectId
                const namespace = service.namespace
                const clusterId = service.clusterId
                const serviceId = service.resourceName
                this.isDetailLoading = true
                this.curService = service
                this.updateServiceSliderConf.isShow = true

                try {
                    const res = await this.$store.dispatch('network/getServiceDetail', {
                        projectId,
                        namespace,
                        clusterId,
                        serviceId
                    })
                    const service = res.data.service[0]
                    const appIds = []

                    this.formatService(service)
                    service.config.webCache.link_app.forEach(item => {
                        appIds.push(item)
                    })

                    this.isLoadBalanceEdited = service.config.isLinkLoadBalance
                    this.curServiceDetail = service
                    this.curVersion = service.version
                    this.getApplications(this.curVersion)
                    this.updateServiceSliderConf.isShow = true
                    this.isDetailLoading = false
                    this.getPorts(appIds)
                    this.getLabels(appIds)
                } catch (e) {
                    catchErrorHandler(e, this)
                    this.updateServiceSliderConf.isShow = false
                } finally {
                    this.isDetailLoading = false
                }
            },

            /**
             * 单选
             * @param {array} selection 已经选中的行数
             * @param {object} row 当前选中的行
             */
            handlePageSelect (selection, row) {
                this.serviceSelectedList = selection
            },

            /**
             * 全选
             */
            handlePageSelectAll (selection, row) {
                this.serviceSelectedList = selection
            },

            /**
             * 清空端口数据
             * @param  {object} port port
             */
            clearPortHandler (port) {
                port.targetPort = ''
            },

            /**
             * 添加端口映射
             */
            addPort () {
                const ports = this.curServiceDetail.config.spec.ports
                const port = {
                    'id': '',
                    'name': '',
                    'port': '',
                    'protocol': 'TCP',
                    'targetPort': '',
                    'nodePort': ''
                }
                ports.push(port)
            },

            /**
             * 删除端口映射
             * @param  {object} port  端口
             * @param  {number} index 索引
             */
            removePort (port, index) {
                const ports = this.curServiceDetail.config.spec.ports
                ports.splice(index, 1)
            },

            /**
             * 关联应用回调处理
             * @param  {number} index 应用索引
             * @param  {object} data  应用
             */
            selectAppsHandler (index, data) {
                this.getPorts(index)
                this.getLabels(index)
            },

            /**
             * 获取Labels
             * @param  {array} apps 应用ids
             */
            async getLabels (apps) {
                this.isLabelsLoading = true
                const projectId = this.projectId
                const version = this.curVersion

                try {
                    const res = await this.$store.dispatch('k8sTemplate/getLabelsByDeployments', {
                        projectId,
                        version,
                        apps
                    })

                    const labels = []
                    for (const key in res.data) {
                        const params = {
                            id: key + ':' + res.data[key],
                            key: key,
                            value: res.data[key],
                            isSelected: false
                        }
                        console.log('this.curServiceDetail.config.webCache', this.curServiceDetail.config.webCache)
                        if (this.curServiceDetail.config.webCache.link_labels && this.curServiceDetail.config.webCache.link_labels.includes(params.id)) {
                            params.isSelected = true
                        }
                        labels.push(params)
                    }
                    this.appLabels.splice(0, this.appLabels.length, ...labels)
                    this.updateLabels()
                } catch (e) {
                    catchErrorHandler(e, this)
                    this.appLabels.splice(0, this.appLabels.length)
                } finally {
                    this.isLabelsLoading = false
                }
            },

            /**
             * 获取应用的端口列表
             * @param  {array} apps 应用id数组
             */
            async getPorts (apps) {
                const projectId = this.projectId
                const version = this.curVersion

                try {
                    const res = await this.$store.dispatch('k8sTemplate/getPortsByDeployments', {
                        projectId,
                        version,
                        apps
                    })

                    const ports = res.data.filter(item => {
                        return item.name
                    })
                    const keys = []
                    let results = []
                    ports.forEach(port => {
                        keys.push(port.id)
                    })
                    this.appPortList.splice(0, this.appPortList.length, ...ports)

                    // 重新拉取后，再匹配用户已经选择的端口是否在列表中
                    results = this.curServiceDetail.config.spec.ports.filter(item => {
                        if (!item.id) {
                            return true
                        } else if (keys.includes(item.id)) {
                            return true
                        } else {
                            return false
                        }
                    })
                    this.curServiceDetail.config.spec.ports.splice(0, this.curServiceDetail.config.spec.ports.length, ...results)

                    // 如果没数据，default：添加一项
                    if (!this.curServiceDetail.config.spec.ports.length) {
                        this.addPort()
                    }
                } catch (e) {
                    this.curServiceDetail.config.spec.ports.splice(0, this.appPortList.length)
                }
            },

            /**
             * 获取LB列表
             * @param  {number} namespace 命名空间ID
             */
            async getLoadBalanceList (namespace) {
                const projectId = this.projectId
                this.loadBalanceList = []
                if (!namespace) return

                try {
                    const res = await this.$store.dispatch('network/getLoadBalanceByNamespace', { projectId, namespace })
                    this.loadBalanceList = res.data
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 获取Applications
             * @param  {number} version 版本号
             */
            async getApplications (version) {
                const projectId = this.projectId

                try {
                    await this.$store.dispatch('k8sTemplate/getAppsByVersion', { projectId, version })
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 显示service详情侧面板
             * @param  {object} service service
             * @param  {number} index 索引
             */
            async showServiceDetail (service, index) {
                this.serviceSlider.title = service.resourceName
                this.curService = service
                this.serviceSlider.isShow = true
                this.getServiceData(service)
            },

            /**
             * 获取service列表
             */
            async getServiceList () {
                const projectId = this.projectId
                const params = {
                    cluster_id: this.searchScope
                }

                try {
                    this.isPageLoading = true
                    const res = await this.$store.dispatch('network/getServiceList', {
                        projectId,
                        params
                    })
                    this.webAnnotations = res.web_annotations

                    setTimeout(() => {
                        this.clearSelectServices()
                        this.initPageConf()
                        this.curPageData = this.getDataByPage(this.pageConf.current)

                        // 如果有搜索关键字，继续显示过滤后的结果
                        if (this.searchKeyword) {
                            this.searchService()
                        }
                    }, 200)
                } catch (e) {
                    // catchErrorHandler(e, this)
                } finally {
                    // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
                    setTimeout(() => {
                        this.isPageLoading = false
                        this.isInitLoading = false
                    }, 400)
                }
            },

            /**
             * 获取service
             * @param  {object} service service
             * @param  {number} index 索引
             */
            async getServiceData (service) {
                const projectId = this.projectId
                const clusterId = service.clusterId
                const namespace = service.namespace
                const name = service.resourceName

                this.isEndpointLoading = true

                try {
                    const res = await this.$store.dispatch('network/getEndpoints', {
                        projectId,
                        clusterId,
                        namespace,
                        name
                    })

                    if (res.data.length) {
                        const service = res.data[0]
                        this.curService.data.metadata.labels = service.data.metadata.labels
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isEndpointLoading = false
                }
            },

            /**
             * 清空搜索
             */
            clearSearch () {
                this.searchKeyword = ''
                this.searchService()
            },

            /**
             * 搜索service
             */
            searchService () {
                const keyword = this.searchKeyword.trim()
                const keyList = ['resourceName', 'namespace']
                const results = []
                let list = JSON.parse(JSON.stringify(this.$store.state.network.serviceList))

                if (this.searchScope) {
                    list = list.filter(item => {
                        return item.clusterId === this.searchScope
                    })
                }

                list.forEach(item => {
                    item.isChecked = false
                    for (const key of keyList) {
                        if (item[key].indexOf(keyword) > -1) {
                            results.push(item)
                            return true
                        }
                    }
                })
                this.serviceList.splice(0, this.serviceList.length, ...results)
                this.pageConf.current = 1
                this.initPageConf()
                this.curPageData = this.getDataByPage(this.pageConf.current)
            },

            /**
             * 初始化分页配置
             */
            initPageConf () {
                const total = this.serviceList.length
                this.pageConf.count = total
                this.pageConf.current = 1
                this.pageConf.totalPage = Math.ceil(total / this.pageConf.limit)
            },

            /**
             * 重新加载当面页数据
             * @return {[type]} [description]
             */
            reloadCurPage () {
                this.initPageConf()
                this.curPageData = this.getDataByPage(this.pageConf.current)
            },

            /**
             * 获取分页数据
             * @param  {number} page 第几页
             * @return {object} data 数据
             */
            getDataByPage (page) {
                let startIndex = (page - 1) * this.pageConf.limit
                let endIndex = page * this.pageConf.limit
                this.isPageLoading = true
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.serviceList.length) {
                    endIndex = this.serviceList.length
                }
                setTimeout(() => {
                    this.isPageLoading = false
                }, 200)
                this.serviceSelectedList = []
                return this.serviceList.slice(startIndex, endIndex)
            },

            /**
             * 页数改变回调
             * @param  {number} page 第几页
             */
            handlePageChange (page = 1) {
                this.pageConf.current = page
                const data = this.getDataByPage(page)
                this.curPageData = data
            },

            /**
             * 更新Label列表
             * @param  {array} list 列表 [{key, value}...]
             * @param  {object} data 对象 {key:value...}
             */
            updateLabelList (list, data) {
                if (!this.curServiceDetail.config.webCache) {
                    this.curServiceDetail.config.webCache = {}
                }
                this.curServiceDetail.config.webCache.labelList = list
            },

            /**
             * 更新备注列表
             * @param  {array} list 列表 [{key, value}...]
             * @param  {object} data 对象 {key:value...}
             */
            updateRemarkList (list, data) {
                if (!this.curServiceDetail.config.webCache) {
                    this.curServiceDetail.config.webCache = {}
                }
                this.curServiceDetail.config.webCache.remarkList = list
            },

            /**
             * 检查提交的数据
             * @return {boolean} true/false 是否合法
             */
            checkData () {
                const service = this.curServiceDetail
                let megPrefix = ''
                if (service.config.spec.type === 'NodePort' && !service.deploy_tag_list.length) {
                    this.$bkMessage({
                        theme: 'error',
                        message: megPrefix + this.$t('请选择要关联的应用'),
                        delay: 3000
                    })
                    return false
                }

                // 端口映射检查
                const ports = service.config.spec.ports
                const serviceIp = service.config.spec.clusterIP
                if (serviceIp !== 'None') {
                    let hasPort = false
                    for (const item of ports) {
                        if (item.name && item.port && item.targetPort) {
                            hasPort = true
                        }
                    }
                    if (!hasPort) {
                        megPrefix += `${this.$t('端口映射')}：`
                        this.$bkMessage({
                            theme: 'error',
                            delay: 8000,
                            message: megPrefix + this.$t('ClusterIP不为None时，请配置相应的端口映射')
                        })
                        return false
                    }
                }

                for (const item of ports) {
                    if (item.name || item.port || item.targetPort) {
                        if (item.name && !/^[a-z]{1}[a-z0-9-]{0,29}$/.test(item.name)) {
                            megPrefix += `${this.$t('端口映射')}：`
                            this.$bkMessage({
                                theme: 'error',
                                delay: 8000,
                                message: megPrefix + this.$t('端口名称以小写字母开头，只能包含：小写字母、数字、连字符(-)，长度小于30个字符')
                            })
                            return false
                        }
                        if (!item.port) {
                            megPrefix += `${this.$t('端口映射')}：`
                            this.$bkMessage({
                                theme: 'error',
                                delay: 5000,
                                message: megPrefix + this.$t('端口不能为空')
                            })
                            return false
                        }
                        if (!item.protocol) {
                            megPrefix += `${this.$t('端口映射')}：`
                            this.$bkMessage({
                                theme: 'error',
                                delay: 5000,
                                message: megPrefix + this.$t('请选择协议')
                            })
                            return false
                        }
                        if (item.nodePort || item.nodePort === 0) {
                            if (item.nodePort < 30000 || item.nodePort > 32767) {
                                megPrefix += `${this.$t('端口映射')}：`
                                this.$bkMessage({
                                    theme: 'error',
                                    delay: 5000,
                                    message: megPrefix + this.$t('NodePort的范围为30000-32767')
                                })
                                return false
                            }
                        }
                    }
                }

                return true
            },

            /**
             * 格式化数据
             */
            formatData () {
                const params = JSON.parse(JSON.stringify(this.curServiceDetail))
                if (params.config.webCache.labelList) {
                    const keys = {}
                    params.config.webCache.labelList.forEach(item => {
                        if (item.key) {
                            keys[item.key] = item.value
                        }
                    })
                    params.config.metadata.labels = keys
                }
                if (params.config.webCache.remarkList) {
                    const keys = {}
                    params.config.webCache.remarkList.forEach(item => {
                        if (item.key) {
                            keys[item.key] = item.value
                        }
                    })
                    params.config.metadata.annotations = keys
                }
                return params
            },

            /**
             * 保存service
             */
            async saveServiceDetail () {
                if (this.checkData()) {
                    const data = this.formatData()
                    const projectId = this.projectId
                    const clusterId = this.curServiceDetail.cluster_id
                    const namespace = this.curServiceDetail.namespace
                    const serviceId = this.curServiceDetail.name

                    if (this.isDetailSaving) {
                        return false
                    }

                    if (!data.config.spec.clusterIP) {
                        data.config.spec.clusterIP = []
                    }

                    this.isDetailSaving = true

                    try {
                        await this.$store.dispatch('network/saveServiceDetail', {
                            projectId,
                            clusterId,
                            namespace,
                            serviceId,
                            data
                        })

                        this.$bkMessage({
                            theme: 'success',
                            message: this.$t('操作成功'),
                            hasCloseIcon: true,
                            delay: 3000
                        })
                        this.getServiceList()
                        this.updateServiceSliderConf.isShow = false
                        this.isDetailSaving = false
                    } catch (e) {
                        catchErrorHandler(e, this)
                    } finally {
                        this.isDetailSaving = false
                    }
                }
            },

            /**
             * 隐藏service侧面板
             */
            hideServiceSlider () {
                this.updateServiceSliderConf.isShow = false
            },

            /**
             * 显示创建 CL5 路由弹框
             */
            showCreateCL5 (service) {
                this.createCL5DialogConf.isShow = true
                this.createCL5DialogConf.service = Object.assign({}, service)
            },

            /**
             * 隐藏创建 CL5 路由弹框
             */
            hideCreateCL5 () {
                this.createCL5DialogConf.isShow = false
                this.createCL5DialogConf.bid4 = ''
                this.createCL5DialogConf.service = Object.assign({}, {})
                this.createCL5DialogConf.isHasSid = 'no'
            },

            /**
             * 创建 CL5 路由
             */
            async createCl5 () {
                const bid4 = this.createCL5DialogConf.bid4.trim()
                if (!bid4) {
                    this.$bkMessage({
                        theme: 'error',
                        delay: 3000,
                        message: this.$t('请填写业务模块')
                    })
                    return
                }

                const params = {
                    projectId: this.projectId,
                    manifest: this.createCL5DialogConf.service.data,
                    clusterId: this.createCL5DialogConf.service.clusterId,
                    crdKind: 'TkexCl5Config'
                }

                if (this.createCL5DialogConf.isHasSid === 'yes') {
                    params.sid = this.createCL5DialogConf.bid4
                } else {
                    params.bid4 = this.createCL5DialogConf.bid4
                }

                try {
                    this.isCreatingCL5 = true
                    await this.$store.dispatch('network/createCl5', params)
                    this.hideCreateCL5()
                    this.$bkMessage({
                        theme: 'success',
                        delay: 1000,
                        message: this.$t('新建CL5成功')
                    })
                    setTimeout(() => {
                        this.searchService()
                    }, 300)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isCreatingCL5 = false
                }
            },

            rowSelectable (row, index) {
                return row.can_delete
                    && this.webAnnotations.perms[row.iam_ns_id]
                    && this.webAnnotations.perms[row.iam_ns_id].namespace_scoped_delete
            }
        }
    }
</script>

<style scoped>
  @import '../service.css';
</style>
