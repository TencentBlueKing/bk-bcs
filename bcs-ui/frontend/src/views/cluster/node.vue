<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-cluster-node-title">
                <i class="bcs-icon bcs-icon-arrows-left back" @click="goIndex" v-if="!curClusterId"></i>
                <template v-if="exceptionCode && exceptionCode.code !== 4005"><span>{{$t('返回')}}</span></template>
                <template v-else>
                    <template v-if="curClusterInPage.cluster_id">
                        <span @click="refreshCurRouter">{{curClusterInPage.name}}</span>
                        <span style="font-size: 12px; color: #c3cdd7;cursor:default;margin-left: 10px;">
                            （{{curClusterInPage.cluster_id}}）
                        </span>
                    </template>
                </template>
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper">
            <app-exception
                v-if="exceptionCode"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>
            <div v-if="!exceptionCode" class="biz-cluster-node-wrapper">
                <div class="biz-cluster-tab-header">
                    <span class="header-item" @click="goOverview">
                        <i class="bcs-icon bcs-icon-bar-chart"></i>{{$t('总览')}}
                    </span>
                    <span class="header-item active">
                        <i class="bcs-icon bcs-icon-list"></i>{{$t('节点管理')}}
                    </span>
                    <span class="header-item" @click="goInfo">
                        <i class="icon-cc icon-cc-machine"></i>{{$t('集群信息')}}
                    </span>
                </div>
                <div class="biz-cluster-tab-content">
                    <bcs-alert type="info" class="biz-cluster-node-tip"
                        :title="['k8s', 'tke'].includes(curClusterInPage.type)
                            ? $t('集群就绪后，您可以创建命名空间、推送项目镜像到仓库，然后通过服务配置模板集或使用Helm部署服务')
                            : $t('集群就绪后，您可以创建命名空间、推送项目镜像到仓库，然后通过服务配置模板集部署服务')">
                    </bcs-alert>
                    <div class="biz-cluster-node-content">
                        <div class="biz-cluster-node-header">
                            <span
                                :disabled="curClusterInPage.state !== 'existing'"
                                v-bk-tooltips="{
                                    content: $t('自有集群不支持通过平台添加节点')
                                }">
                                <bk-button type="primary"
                                    :disabled="curClusterInPage.state === 'existing'"
                                    @click.stop="openDialog">
                                    <i class="bcs-icon bcs-icon-plus"></i>
                                    <span>{{$t('添加节点')}}</span>
                                </bk-button>
                            </span>
                            <template v-if="$INTERNAL && curClusterInPage.providerType === 'tke'">
                                <apply-host theme="primary" style="display: inline-block;" :cluster-id="clusterId" :is-backfill="true" />
                            </template>
                            <bcs-popover v-if="!allowBatch" :content="dontAllowBatchMsg" placement="top">
                                <bk-dropdown-menu :align="'center'" ref="toggleFilterDropdownMenu" class="batch-operate-dropdown" :disabled="true">
                                    <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button batch-operate" :class="!allowBatch ? 'disabled' : ''">
                                        <span class="label">{{$t('批量操作')}}</span>
                                        <i class="bcs-icon bcs-icon-angle-down dropdown-menu-angle-down"></i>
                                    </a>
                                </bk-dropdown-menu>
                            </bcs-popover>
                            <bk-dropdown-menu v-else :align="'center'" ref="toggleFilterDropdownMenu" class="batch-operate-dropdown">
                                <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button batch-operate" :class="!allowBatch ? 'disabled' : ''">
                                    <span class="label">{{$t('批量操作')}}</span>
                                    <i class="bcs-icon bcs-icon-angle-down dropdown-menu-angle-down"></i>
                                </a>
                                <ul class="bk-dropdown-list" slot="dropdown-content">
                                    <li>
                                        <a class="action" href="javascript:void(0)" @click="batchOperate('1')">{{$t('允许调度')}}</a>
                                    </li>
                                    <li>
                                        <a class="action" href="javascript:void(0)" @click="batchOperate('2')">{{$t('停止调度')}}</a>
                                    </li>
                                    <li>
                                        <a class="action" href="javascript:void(0)" @click="batchOperate('3')">{{$t('删除')}}</a>
                                    </li>
                                    <li>
                                        <a class="action" href="javascript:void(0)" @click="exportNode">{{$t('导出')}}</a>
                                    </li>
                                    <li>
                                        <a class="action" href="javascript:void(0)" @click="batchOperate('4')">{{$t('重新添加')}}</a>
                                        <!-- <a href="javascript:void(0)" v-else class="action disabled" :title="$t('所选节点均处于初始化失败状态时才允许此操作')">{{$t('重新添加')}}</a> -->
                                    </li>
                                </ul>
                            </bk-dropdown-menu>
                            <bk-dropdown-menu :align="'left'" ref="copyIpDropdownMenu" class="copy-ip-dropdown">
                                <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button copy-ip-btn">
                                    <span class="label">{{$t('复制IP')}}</span>
                                    <i class="bcs-icon bcs-icon-angle-down dropdown-menu-angle-down"></i>
                                </a>
                                <ul class="bk-dropdown-list" slot="dropdown-content">
                                    <li>
                                        <a href="javascript:void(0)" @click="copyIp('selected')" class="selected" :class="!allowBatch ? 'disabled' : ''">{{$t('复制所选IP')}}</a>
                                    </li>
                                    <li>
                                        <a href="javascript:void(0)" @click="copyIp('cur-page')" class="cur-page">{{$t('复制当前页IP')}}</a>
                                    </li>
                                    <li>
                                        <a href="javascript:void(0)" @click="copyIp('all')" class="all">{{$t('复制所有IP')}}</a>
                                    </li>
                                </ul>
                            </bk-dropdown-menu>
                            <div class="biz-searcher-wrapper">
                                <node-searcher :cluster-id="clusterId" :project-id="projectId" ref="searcher"
                                    :params="ipSearchParams" @search="searchNodeList"></node-searcher>
                            </div>
                            <span class="close-wrapper">
                                <template v-if="$refs.searcher && $refs.searcher.searchParams && $refs.searcher.searchParams.length">
                                    <bk-button class="bk-button bk-default is-outline is-icon" :title="$t('清除')" style="border: 1px solid #c4c6cc;" @click="clearSearchParams">
                                        <i class="bcs-icon bcs-icon-close"></i>
                                    </bk-button>
                                </template>
                                <template v-else>
                                    <bk-button class="bk-button bk-default is-outline is-icon" style="border-color: #c4c6cc;">
                                    </bk-button>
                                </template>
                            </span>

                            <span class="refresh-wrapper">
                                <bk-button theme="default" v-bk-tooltips="$t('重置')" @click="refresh">
                                    <i class="bcs-icon bcs-icon-refresh"></i>
                                </bk-button>
                            </span>
                        </div>
                        <div class="biz-cluster-node-table-wrapper" v-bkloading="{ isLoading: isPageLoading, zIndex: 500 }">
                            <table class="bk-table has-table-hover biz-table" :style="{ borderBottomWidth: nodeList.length ? '1px' : 0 }">
                                <thead>
                                    <tr>
                                        <!-- k8s 10 列、tke 9 列 -->
                                        <th style="width: 3%; text-align: center; padding: 0; padding-left: 20px;">
                                            <bk-checkbox name="check-all-node" v-model="isCheckCurPageAllNode" @change="checkAllNode(...arguments)" />
                                        </th>
                                        <th style="width: 10%; padding-left: 10px;">{{$t('主机名/IP')}}</th>
                                        <th style="width: 8%;">{{$t('状态')}}</th>
                                        <th style="width: 8%;">{{$t('容器数量')}}</th>
                                        <th style="width: 8%;">{{$t('Pod数量')}}</th>
                                        <th style="width: 10%;">
                                            CPU
                                            <div class="biz-table-sort">
                                                <span class="sort-direction asc"
                                                    :class="sortIdx === 'cpu_summary' ? 'active' : ''"
                                                    :title="sortIdx === 'cpu_summary' ? $t('取消') : $t('升序')"
                                                    @click="sortNodeList('cpu_summary', 'asc', 'cpu_summary')"></span>
                                                <span class="sort-direction desc"
                                                    :class="sortIdx === '-cpu_summary' ? 'active' : ''"
                                                    :title="sortIdx === '-cpu_summary' ? $t('取消') : $t('降序')"
                                                    @click="sortNodeList('cpu_summary', 'desc', '-cpu_summary')"></span>
                                            </div>
                                        </th>
                                        <th style="width: 10%;">
                                            {{$t('内存')}}
                                            <div class="biz-table-sort">
                                                <span class="sort-direction asc"
                                                    :class="sortIdx === 'mem' ? 'active' : ''"
                                                    :title="sortIdx === 'mem' ? $t('取消') : $t('升序')"
                                                    @click="sortNodeList('mem', 'asc', 'mem')"></span>
                                                <span class="sort-direction desc"
                                                    :class="sortIdx === '-mem' ? 'active' : ''"
                                                    :title="sortIdx === '-mem' ? $t('取消') : $t('降序')"
                                                    @click="sortNodeList('mem', 'desc', '-mem')"></span>
                                            </div>
                                        </th>
                                        <th style="width: 10%;">
                                            {{$t('磁盘')}}
                                            <div class="biz-table-sort">
                                                <span class="sort-direction asc "
                                                    :class="sortIdx === 'disk' ? 'active' : ''"
                                                    :title="sortIdx === 'disk' ? $t('取消') : $t('升序')"
                                                    @click="sortNodeList('disk', 'asc', 'disk')"></span>
                                                <span class="sort-direction desc"
                                                    :class="sortIdx === '-disk' ? 'active' : ''"
                                                    :title="sortIdx === '-disk' ? $t('取消') : $t('降序')"
                                                    @click="sortNodeList('disk', 'desc', '-disk')"></span>
                                            </div>
                                        </th>
                                        <template>
                                            <th style="width: 9%;">
                                                {{$t('磁盘IO')}}
                                                <div class="biz-table-sort">
                                                    <span class="sort-direction asc "
                                                        :class="sortIdx === 'io' ? 'active' : ''"
                                                        :title="sortIdx === 'io' ? $t('取消') : $t('升序')"
                                                        @click="sortNodeList('io', 'asc', 'io')"></span>
                                                    <span class="sort-direction desc"
                                                        :class="sortIdx === '-io' ? 'active' : ''"
                                                        :title="sortIdx === '-io' ? $t('取消') : $t('降序')"
                                                        @click="sortNodeList('io', 'desc', '-io')"></span>
                                                </div>
                                            </th>
                                        </template>
                                        <th style="width: 28%; text-align: left;"><span>{{$t('操作')}}</span></th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <template v-if="nodeList.length">
                                        <tr v-for="(node, index) in curNodeList" :key="index">
                                            <!-- 全选 -->
                                            <td style="width: 3%; text-align: center; padding: 0; padding-left: 20px;">
                                                <!-- <label class="bk-form-checkbox" style="margin-top: 6px;">
                                                    <input type="checkbox" name="check-node" v-model="node.isChecked" @click="checkNode(node, index)" />
                                                </label> -->
                                                <bk-checkbox name="check-node" v-model="node.isChecked" @change="checkNode(node, ...arguments)" />
                                            </td>
                                            <!-- 节点IP -->
                                            <td class="pl10">
                                                <bk-button
                                                    :disabled="['INITIALIZATION', 'DELETING'].includes(node.status)"
                                                    text
                                                    @click="goNodeOverview(node)"
                                                >
                                                    {{ node.inner_ip }}
                                                </bk-button>
                                            </td>
                                            <!-- 节点状态 -->
                                            <td>
                                                <loading-cell :style="{ left: 0 }"
                                                    :ext-cls="['bk-spin-loading-mini', 'bk-spin-loading-danger']"
                                                    v-if="['INITIALIZATION', 'DELETING'].includes(node.status)"
                                                ></loading-cell>
                                                <StatusIcon :status="node.status" :status-color-map="nodeStatusColorMap" v-else>
                                                    {{ statusMap[node.status.toLowerCase()] }}
                                                </StatusIcon>
                                            </td>
                                            <!-- 容器数量 -->
                                            <td>{{node.containerCount || ''}}</td>
                                            <!-- POD数量 -->
                                            <td>
                                                {{node.podCount || ''}}
                                            </td>
                                            <!-- CPU -->
                                            <td>
                                                <ring-cell :percent="node.cpuMetric" :fill-color="'#3ede78'"
                                                    v-if="node.cpuMetric !== null && node.cpuMetric !== undefined"
                                                ></ring-cell>
                                                <loading-cell v-else></loading-cell>
                                            </td>
                                            <!-- 内存 -->
                                            <td>
                                                <ring-cell :percent="node.memMetric" :fill-color="'#3a84ff'"
                                                    v-if="node.memMetric !== null && node.memMetric !== undefined"
                                                ></ring-cell>
                                                <loading-cell v-else></loading-cell>
                                            </td>
                                            <!-- 磁盘 -->
                                            <td>
                                                <ring-cell :percent="node.diskMetric" :fill-color="'#853cff'"
                                                    v-if="node.diskMetric !== null && node.diskMetric !== undefined"
                                                ></ring-cell>
                                                <loading-cell v-else></loading-cell>
                                            </td>
                                            <!-- 磁盘IO -->
                                            <td>
                                                <ring-cell :percent="node.diskioMetric" :fill-color="'#853cff'"
                                                    v-if="node.diskioMetric !== null && node.diskioMetric !== undefined"
                                                ></ring-cell>
                                                <loading-cell v-else></loading-cell>
                                            </td>
                                            <!-- 节点操作 -->
                                            <td class="node-option">
                                                <bk-button text @click="stopNode(node, index)" v-if="node.status === 'RUNNING'">
                                                    {{ $t('停止调度') }}
                                                </bk-button>
                                                <bk-button text
                                                    v-if="['INITIALIZATION', 'DELETING', 'REMOVE-FAILURE', 'ADD-FAILURE'].includes(node.status)"
                                                    @click="handleShowLog(node)"
                                                >
                                                    {{$t('查看日志')}}
                                                </bk-button>
                                                <template v-if="node.status === 'REMOVABLE'">
                                                    <bk-button text @click="enableNode(node, index)">
                                                        {{ $t('允许调度') }}
                                                    </bk-button>
                                                    <bk-button text class="ml10" @click="schedulerNode(node, index)">{{ $t('pod迁移') }}</bk-button>
                                                </template>
                                                <bk-button text class="ml10" v-if="['REMOVE-FAILURE', 'ADD-FAILURE', 'REMOVABLE'].includes(node.status)"
                                                    @click="showDelNode(node, index)">
                                                    {{ $t('删除') }}
                                                </bk-button>
                                                <bk-button text class="ml10"
                                                    v-if="['REMOVE-FAILURE', 'ADD-FAILURE'].includes(node.status)"
                                                    @click="handleRetry(node, index)"
                                                >{{ $t('重试') }}</bk-button>
                                            </td>
                                        </tr>
                                    </template>
                                    <template v-else>
                                        <tr class="no-hover">
                                            <td colspan="9">
                                                <div class="bk-message-box">
                                                    <bcs-exception type="empty" scene="part"></bcs-exception>
                                                </div>
                                            </td>
                                        </tr>
                                    </template>
                                </tbody>
                            </table>
                            <div class="bk-table-footer" v-if="nodeListPageConf.total">
                                <bk-pagination
                                    :location="'left'"
                                    :show-limit="true"
                                    :current="nodeListPageConf.curPage"
                                    :count="nodeListPageConf.total"
                                    :limit="nodeListPageConf.pageSize"
                                    @limit-change="changePageSize"
                                    @change="nodeListPageChange">
                                </bk-pagination>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <IpSelector v-model="showIpSelector" @confirm="chooseServer"></IpSelector>

        <bk-sideslider
            :is-show.sync="logSideDialogConf.isShow"
            :title="logSideDialogConf.title"
            :width="640"
            @hidden="closeLog"
            :quick-close="true">
            <div slot="content">
                <div class="log-wrapper">
                    <bk-table :data="taskData">
                        <bk-table-column :label="$t('步骤')" prop="taskName"></bk-table-column>
                        <bk-table-column :label="$t('状态')" prop="status">
                            <template #default="{ row }">
                                <div class="log-wrapper-status" v-if="row.status === 'RUNNING'">
                                    <loading-cell :style="{ left: 0, margin: 0 }"
                                        :ext-cls="['bk-spin-loading-mini', 'bk-spin-loading-danger']"></loading-cell>
                                    <span class="ml5">{{ $t('运行中') }}</span>
                                </div>
                                <StatusIcon :status="row.status" :status-color-map="statusColorMap" v-else>
                                    {{ taskStatusTextMap[row.status.toLowerCase()] }}
                                </StatusIcon>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('内容')" prop="message"></bk-table-column>
                    </bk-table>
                </div>
            </div>
        </bk-sideslider>

        <tip-dialog
            ref="nodeNoticeDialog"
            icon="bcs-icon bcs-icon-exclamation-triangle"
            :title="$t('添加节点')"
            :sub-title="$t('此操作需要对你的主机进行如下操作，请知悉：')"
            :check-list="nodeNoticeList"
            :is-confirming="isCreating"
            :confirm-btn-text="$t('确定，添加节点')"
            :cancel-btn-text="$t('我再想想')"
            :confirm-loading="nodeNoticeLoading"
            :confirm-callback="saveNode">
        </tip-dialog>

        <tip-dialog
            ref="removeNodeDialog"
            icon="bcs-icon bcs-icon-exclamation-triangle"
            :show-close="false"
            :sub-title="$t('此操作无法撤回，请确认：')"
            :tips="$t('注意: 节点状态以集群中的状态为准；点击【删除】后，节点状态可能会仍然处于不可调度')"
            :check-list="deleteNodeNoticeList"
            :confirm-btn-text="$t('确定')"
            :confirming-btn-text="$t('删除中...')"
            :canceling-btn-text="$t('取消')"
            :confirm-callback="confirmDelNode"
            :cancel-callback="cancelDelNode">
        </tip-dialog>

        <tip-dialog
            ref="forceRemoveNodeDialog"
            icon="bcs-icon bcs-icon-exclamation-triangle"
            :show-close="false"
            :sub-title="$t('此操作无法撤回，请确认：')"
            :tips="$t('注意: 节点状态以集群中的状态为准；点击【删除】后，节点状态可能会仍然处于不可调度')"
            :check-list="deleteNodeNoticeList"
            :confirm-btn-text="$t('确定')"
            :confirming-btn-text="$t('删除中...')"
            :canceling-btn-text="$t('取消')"
            :confirm-callback="confirmForceRemoveNode"
            :cancel-callback="cancelForceRemoveNode">
        </tip-dialog>

        <tip-dialog
            ref="faultRemoveDialog"
            icon="bcs-icon bcs-icon-exclamation-triangle"
            :show-close="false"
            :sub-title="$t('此操作无法撤回，请确认：') "
            :check-list="faultRemoveoticeList"
            :confirm-btn-text="$t('确定')"
            :confirming-btn-text="$t('删除中...')"
            :canceling-btn-text="$t('取消')"
            :confirm-callback="confirmFaultRemove"
            :cancel-callback="cancelFaultRemove">
        </tip-dialog>

        <tip-dialog
            ref="recordRemoveDialog"
            icon="bcs-icon bcs-icon-exclamation-triangle"
            :has-sub-title="false"
            :show-close="false"
            :check-list="recordRemoveNoticeList"
            :confirm-btn-text="$t('确定')"
            :confirming-btn-text="$t('删除中...')"
            :canceling-btn-text="$t('取消')"
            :confirm-callback="confirmRecordRemove"
            :cancel-callback="cancelRecordRemove">
        </tip-dialog>

        <bk-dialog
            :is-show.sync="reInitializationDialogConf.isShow"
            :width="reInitializationDialogConf.width"
            :title="reInitializationDialogConf.title"
            :close-icon="reInitializationDialogConf.closeIcon"
            :ext-cls="'biz-node-re-initialization-dialog'"
            :quick-close="false">
            <template slot="content">
                <div>{{reInitializationDialogConf.content}}</div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <bk-button type="primary" disabled>
                            {{$t('初始化中...')}}
                        </bk-button>
                        <bk-button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                    <template v-else>
                        <bk-button type="primary" @click="reInitializationConfirm">
                            {{$t('确定')}}
                        </bk-button>
                        <bk-button type="button" @click="reInitializationCancel">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="reDelDialogConf.isShow"
            :width="reDelDialogConf.width"
            :title="reDelDialogConf.title"
            :close-icon="reDelDialogConf.closeIcon"
            :ext-cls="'biz-node-re-initialization-dialog'"
            :quick-close="false">
            <template slot="content">
                <div>{{reDelDialogConf.content}}</div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <bk-button type="primary" disabled>
                            {{$t('删除中...')}}
                        </bk-button>
                        <bk-button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                    <template v-else>
                        <bk-button type="primary" @click="reDelConfirm">
                            {{$t('确定')}}
                        </bk-button>
                        <bk-button type="button" @click="reDelCancel">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="delDialogConf.isShow"
            :width="delDialogConf.width"
            :title="delDialogConf.title"
            :close-icon="delDialogConf.closeIcon"
            :ext-cls="'biz-node-re-initialization-dialog'"
            :quick-close="false">
            <template slot="content">
                <div>{{delDialogConf.content}}</div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <bk-button type="primary" disabled>
                            {{$t('删除中...')}}
                        </bk-button>
                        <bk-button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                    <template v-else>
                        <bk-button type="primary" @click="delConfirm">
                            {{$t('确定')}}
                        </bk-button>
                        <bk-button type="button" @click="delCancel">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="enableDialogConf.isShow"
            :width="enableDialogConf.width"
            :title="enableDialogConf.title"
            :close-icon="enableDialogConf.closeIcon"
            :ext-cls="'biz-node-re-initialization-dialog'"
            :quick-close="false">
            <template slot="content">
                <div>{{enableDialogConf.content}}</div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <bk-button type="primary" disabled>
                            {{$t('启用中...')}}
                        </bk-button>
                        <bk-button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                    <template v-else>
                        <bk-button type="primary" @click="enableConfirm">
                            {{$t('确定')}}
                        </bk-button>
                        <bk-button type="button" @click="enableCancel">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="stopDialogConf.isShow"
            :width="stopDialogConf.width"
            :title="stopDialogConf.title"
            :ext-cls="'biz-node-re-initialization-dialog'"
            :quick-close="false"
            @cancel="stopDialogConf.isShow = false">
            <template slot="content">
                <div :class="{ 'stopDialog-content': true, 'font-content': !this.$INTERNAL }">{{stopDialogConf.content}}</div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <bk-button type="primary" disabled>
                            {{$t('停用中...')}}
                        </bk-button>
                        <bk-button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                    <template v-else>
                        <bk-button type="primary" @click="stopConfirm">
                            {{$t('确定')}}
                        </bk-button>
                        <bk-button type="button" @click="stopCancel">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="removeDialogConf.isShow"
            :width="removeDialogConf.width"
            :title="removeDialogConf.title"
            :close-icon="removeDialogConf.closeIcon"
            :ext-cls="'biz-node-re-initialization-dialog'"
            :quick-close="false">
            <template slot="content">
                <div>{{removeDialogConf.content}}</div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <bk-button type="primary" disabled>
                            {{$t('删除中...')}}
                        </bk-button>
                        <bk-button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                    <template v-else>
                        <bk-button type="primary" @click="removeConfirm">
                            {{$t('确定')}}
                        </bk-button>
                        <bk-button type="button" @click="removeCancel">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="schedulerDialogConf.isShow"
            :width="schedulerDialogConf.width"
            :title="schedulerDialogConf.title"
            :close-icon="schedulerDialogConf.closeIcon"
            :ext-cls="'biz-node-re-initialization-dialog'"
            :quick-close="false">
            <template slot="content">
                <div>{{schedulerDialogConf.content}}</div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <bk-button type="primary" disabled>
                            {{$t('迁移中...')}}
                        </bk-button>
                        <bk-button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                    <template v-else>
                        <bk-button type="primary" @click="schedulerConfirm">
                            {{$t('确定')}}
                        </bk-button>
                        <bk-button type="button" @click="schedulerCancel">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="batchDialogConf.isShow"
            :width="batchDialogConf.width"
            :title="batchDialogConf.title"
            :close-icon="batchDialogConf.closeIcon"
            :ext-cls="'biz-node-re-initialization-dialog'"
            :quick-close="false">
            <template slot="content">
                <i18n path="确定要对{len}个节点进行{operate}操作？" tag="div">
                    <span place="len" class="len">{{batchDialogConf.len}}</span>
                    <span place="operate" class="operate">{{batchDialogConf.operate}}</span>
                </i18n>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <bk-button type="primary" disabled>
                            {{$t('操作中...')}}
                        </bk-button>
                        <bk-button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                    <template v-else>
                        <bk-button type="primary" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary"
                            @click="batchConfirm">
                            {{$t('确定')}}
                        </bk-button>
                        <bk-button type="button" @click="batchCancel">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                </div>
            </div>
        </bk-dialog>

    </div>
</template>

<script>
    import applyPerm from '@/mixins/apply-perm'
    import tipDialog from '@/components/tip-dialog'
    import RingCell from './ring-cell'
    import LoadingCell from './loading-cell'
    import mixin from '@/views/cluster/mixin-node'
    import nodeSearcher from '@/views/cluster/searcher'
    import ApplyHost from './apply-host.vue'
    import IpSelector from '@/components/ip-selector/selector-dialog.vue'
    import StatusIcon from '@/views/dashboard/common/status-icon.tsx'

    export default {
        components: {
            RingCell,
            LoadingCell,
            tipDialog,
            nodeSearcher,
            ApplyHost,
            IpSelector,
            StatusIcon
        },
        mixins: [applyPerm, mixin],
        data () {
            return {
                nodeNoticeList: [
                    {
                        id: 1,
                        text: this.$t('操作系统初始化'),
                        isChecked: true
                    },
                    {
                        id: 2,
                        text: this.$t('安装容器服务相关的组件'),
                        isChecked: true
                    }
                ],
                statusMap: {
                    initialization: this.$t('初始化中'),
                    running: this.$t('正常'),
                    deleting: this.$t('删除中'),
                    'add-failure': this.$t('上架失败'),
                    'remove-failure': this.$t('下架失败'),
                    removable: this.$t('不可调度'),
                    notready: this.$t('不正常'),
                    unknown: this.$t('未知状态')
                },
                nodeStatusColorMap: {
                    initialization: 'blue',
                    running: 'green',
                    deleting: 'blue',
                    'add-failure': 'red',
                    'remove-failure': 'red',
                    removable: '',
                    notready: 'red',
                    unknown: ''
                },
                taskStatusTextMap: {
                    initialzing: this.$t('初始化中'),
                    running: this.$t('运行中'),
                    success: this.$t('成功'),
                    failure: this.$t('失败'),
                    timeout: this.$t('超时'),
                    notstarted: this.$t('未执行')
                },
                statusColorMap: {
                    initialzing: 'blue',
                    running: 'blue',
                    success: 'green',
                    failure: 'red',
                    timeout: 'red',
                    notstarted: 'blue'
                },
                taskData: [],
                taskTimer: null
            }
        },
        computed: {
            curClusterId () {
                return this.$store.state.curClusterId
            }
        },
        methods: {
            async handleShowLog (node) {
                if (!node?.permissions?.edit) {
                    await this.$store.dispatch('getResourcePermissions', {
                        project_id: this.projectId,
                        policy_code: 'view',
                        resource_code: this.curClusterInPage.cluster_id,
                        resource_name: this.curClusterInPage.name,
                        resource_type: `cluster_${this.curClusterInPage?.environment === 'prod' ? 'prod' : 'test'}`
                    })
                }
                this.logSideDialogConf.isShow = true
                this.logSideDialogConf.title = node.inner_ip

                this.handleGetTaskData(node)
            },
            async handleGetTaskData (node) {
                const res = await this.$store.dispatch('clustermanager/taskList', {
                    clusterID: this.clusterId,
                    projectID: this.projectId,
                    nodeIP: node.inner_ip
                })
                const latestTask = res.latestTask
                const steps = latestTask?.stepSequence || []
                this.taskData = steps.map(step => {
                    return latestTask?.steps[step]
                })
                if (['RUNNING', 'INITIALZING'].includes(latestTask?.status)) {
                    this.taskTimer = setTimeout(() => {
                        this.handleGetTaskData(node)
                    }, 5000)
                } else {
                    clearTimeout(this.taskTimer)
                    this.taskTimer = null
                }
            }
        }
    }
</script>

<style lang="postcss">
.tippy-tooltip.create-node-selector-theme {
    border: 1px solid #dde4eb;
    padding: 0;
    .create-node-selector-list a {
        display: block;
        height: 32px;
        line-height: 33px;
        padding: 0 16px;
        color: #63656e;
        font-size: 12px;
        font-size: 12px;
        text-decoration: none;
        white-space: nowrap;
        &:hover {
            background-color: #eaf3ff;
            color: #3a84ff;
        }
        .bcs-icon {
            position: relative;
            margin-right: 0;
            margin-left: 5px;
            vertical-align: middle;
            color: #2dcb56;
            top: -1px;
        }
    }
}
</style>

<style scoped lang="postcss">
    @import './node.css';
    .server-tip {
        float: left;
        line-height: 17px;
        font-size: 12px;
        text-align: left;
        padding: 13px 0 13px 20px;
        margin-left: 20px;

        li {
            list-style: circle;
        }
    }
    .stopDialog-title {
        font-size: 16px;
    }
    .stopDialog-content {
        color: #ff5656;
    }
    .font-content {
        color: #63656e !important;
    }
    .log-wrapper {
        padding: 20px 30px 0 30px;
        &-status {
            display: flex;
            align-items: center;
        }
    }
    /* .bk-dialog-footer .bk-dialog-outer button {
        margin-top: 30px;
    } */
</style>
