<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-cluster-node-title">
                <i class="bk-icon icon-arrows-left back" @click="goIndex"></i>
                <template v-if="exceptionCode"><span>{{$t('返回')}}</span></template>
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
                v-if="exceptionCode && !getClusterLoading"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>
            <div v-if="!exceptionCode && !getClusterLoading" class="biz-cluster-node-wrapper">
                <div class="biz-cluster-tab-header">
                    <div class="header-item" @click="goOverview">
                        <i class="bk-icon icon-bar-chart"></i>{{$t('总览')}}
                    </div>
                    <div class="header-item active">
                        <i class="bk-icon icon-list"></i>{{$t('节点管理')}}
                    </div>
                    <div class="header-item" @click="goInfo">
                        <i class="icon-cc icon-cc-machine"></i>{{$t('集群信息')}}
                    </div>
                </div>
                <div class="biz-cluster-tab-content" v-bkloading="{ isLoading: isInitLoading, opacity: 1 }" :style="{ height: isInitLoading ? '300px' : 'auto' }">
                    <div class="biz-cluster-node-tip" v-if="curClusterInPage.type === 'k8s' || curProject.kind === 3">{{$t('集群就绪后，您可以创建命名空间、推送项目镜像到仓库，然后通过服务配置模板集或使用Helm部署服务')}}</div>
                    <div class="biz-cluster-node-tip" v-else>{{$t('集群就绪后，您可以创建命名空间、推送项目镜像到仓库，然后通过服务配置模板集部署服务')}}</div>
                    <div class="biz-cluster-node-content" v-if="!isInitLoading">
                        <div class="biz-cluster-node-header">
                            <bk-tooltip v-if="curClusterInPage.state === 'existing'" :content="$t('自有集群不支持通过平台添加节点')" placement="top">
                                <button class="bk-button bk-primary is-disabled">
                                    <i class="bk-icon icon-plus"></i>
                                    <span>{{$t('添加节点')}}</span>
                                </button>
                            </bk-tooltip>
                            <button v-else class="bk-button bk-primary" @click.stop="openDialog">
                                <i class="bk-icon icon-plus"></i>
                                <span>{{$t('添加节点')}}</span>
                            </button>
                            <bk-tooltip v-if="!allowBatch" :content="dontAllowBatchMsg" placement="top">
                                <bk-dropdown-menu :align="'center'" ref="toggleFilterDropdownMenu" class="batch-operate-dropdown" :disabled="true">
                                    <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button batch-operate" :class="!allowBatch ? 'disabled' : ''">
                                        <span class="label">{{$t('批量操作')}}</span>
                                        <i class="bk-icon icon-angle-down dropdown-menu-angle-down"></i>
                                    </a>
                                </bk-dropdown-menu>
                            </bk-tooltip>
                            <bk-dropdown-menu v-else :align="'center'" ref="toggleFilterDropdownMenu" class="batch-operate-dropdown">
                                <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button batch-operate" :class="!allowBatch ? 'disabled' : ''">
                                    <span class="label">{{$t('批量操作')}}</span>
                                    <i class="bk-icon icon-angle-down dropdown-menu-angle-down"></i>
                                </a>
                                <ul class="bk-dropdown-list" slot="dropdown-content">
                                    <li>
                                        <a class="action" href="javascript:void(0)" @click="batchOperate('1')">{{$t('允许调度')}}</a>
                                    </li>
                                    <li>
                                        <a class="action" href="javascript:void(0)" @click="batchOperate('2')">{{$t('停止调度')}}</a>
                                    </li>
                                    <li v-if="curClusterInPage.state !== 'existing'">
                                        <a class="action" href="javascript:void(0)" @click="batchOperate('3')">{{$t('删除')}}</a>
                                    </li>
                                    <li>
                                        <a class="action" href="javascript:void(0)" @click="exportNode">{{$t('导出')}}</a>
                                    </li>
                                    <li>
                                        <a class="action" href="javascript:void(0)" @click="batchOperate('4')" v-if="isBatchReInstall">{{$t('重新添加')}}</a>
                                        <a href="javascript:void(0)" v-else class="action disabled" :title="$t('所选节点均处于初始化失败状态时才允许此操作')">{{$t('重新添加')}}</a>
                                    </li>
                                </ul>
                            </bk-dropdown-menu>
                            <bk-dropdown-menu :align="'left'" ref="copyIpDropdownMenu" class="copy-ip-dropdown">
                                <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button copy-ip-btn">
                                    <span class="label">{{$t('复制IP')}}</span>
                                    <i class="bk-icon icon-angle-down dropdown-menu-angle-down"></i>
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
                                    <button class="bk-button bk-default is-outline is-icon" :title="$t('清除')" @click="clearSearchParams">
                                        <i class="bk-icon icon-close"></i>
                                    </button>
                                </template>
                                <template v-else>
                                    <button class="bk-button bk-default is-outline is-icon">
                                    </button>
                                </template>
                            </span>

                            <span class="refresh-wrapper">
                                <bk-tooltip :content="$t('重置')" :transfer="true" :placement="'top-end'">
                                    <button class="bk-button bk-default is-outline is-icon" :title="$t('重置')" @click="refresh">
                                        <i class="bk-icon icon-refresh"></i>
                                    </button>
                                </bk-tooltip>
                            </span>
                        </div>
                        <div class="biz-cluster-node-table-wrapper" v-bkloading="{ isLoading: isPageLoading, opacity: 1, zIndex: 500 }">
                            <table class="bk-table has-table-hover biz-table" :style="{ borderBottomWidth: nodeList.length ? '1px' : 0 }">
                                <thead>
                                    <tr>
                                        <th style="width: 3%; text-align: center; padding: 0; padding-left: 20px;">
                                            <label class="bk-form-checkbox">
                                                <input type="checkbox" name="check-all-node" v-model="isCheckCurPageAllNode" @click="checkAllNode($event)" />
                                            </label>
                                        </th>
                                        <th style="width: 12%; padding-left: 10px;">{{$t('主机名/IP')}}</th>
                                        <th style="width: 8%;">{{$t('状态')}}</th>
                                        <th style="width: 8%;">{{$t('容器数量')}}</th>
                                        <template v-if="curClusterInPage.type === 'k8s'">
                                            <th style="width: 8%;">{{$t('Pod数量')}}</th>
                                        </template>
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
                                        <th style="width: 28%; text-align: left;"><span>{{$t('操作')}}</span></th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <template v-if="nodeList.length">
                                        <tr v-for="(node, index) in nodeList" :key="index">
                                            <td style="width: 3%; text-align: center; padding: 0; padding-left: 20px;">
                                                <label class="bk-form-checkbox">
                                                    <input type="checkbox" name="check-node" v-model="node.isChecked" @click="checkNode(node, index)" />
                                                </label>
                                            </td>
                                            <!--
                                                初始化中: initializing, so_initializing, initial_checking, uninitialized
                                                删除中: removing
                                                操作: 查看日志
                                            -->
                                            <template v-if="ingStatus.includes(node.status)">
                                                <td style="padding-left: 10px;">
                                                    {{node.inner_ip}}
                                                </td>
                                                <td>
                                                    <div class="biz-status-node"><loading-cell :style="{ left: 0 }" :ext-cls="['bk-spin-loading-mini', 'bk-spin-loading-danger']"></loading-cell></div>
                                                    {{node.status === 'initializing' || node.status === 'so_initializing' || node.status === 'initial_checking' ? $t('初始化中') : $t('删除中')}}
                                                </td>
                                                <td></td>
                                                <td></td>
                                                <td></td>
                                                <td></td>
                                                <td></td>
                                                <td v-if="curClusterInPage.type === 'k8s'"></td>
                                                <td style="text-align: left;">
                                                    <a href="javascript:void(0);" class="bk-text-button" @click.stop="showLog(node)">{{$t('查看日志')}}</a>
                                                    <bk-dropdown-menu class="dropdown-menu ml10" :align="'center'" ref="dropdown">
                                                        <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button">
                                                            {{$t('更多')}}
                                                            <i class="bk-icon icon-angle-down dropdown-menu-angle-down"></i>
                                                        </a>
                                                        <ul class="bk-dropdown-list" slot="dropdown-content">
                                                            <li>
                                                                <a href="javascript:void(0);" class="bk-text-button" @click.stop="showRecordRemove(node, index)">
                                                                    {{$t('仅移除记录')}}
                                                                </a>
                                                            </li>
                                                        </ul>
                                                    </bk-dropdown-menu>
                                                </td>
                                            </template>

                                            <!--
                                                初始化失败: initial_failed, so_init_failed, check_failed, bke_failed, schedule_failed
                                                操作: 查看日志，删除，重试（重试初始化）

                                                删除失败: delete_failed
                                                操作: 查看日志，删除

                                                删除失败: remove_failed
                                                操作: 查看日志，重试（重试删除）
                                            -->
                                            <template v-if="failStatus.includes(node.status)">
                                                <td style="padding-left: 10px;">
                                                    <a href="javascript:void(0)" class="bk-text-button" @click="goNodeOverview(node)">{{node.inner_ip}}</a>
                                                </td>
                                                <td>
                                                    <div class="biz-status-node"><i class="node danger"></i></div>
                                                    {{node.status === 'initial_failed' || node.status === 'so_init_failed' || node.status === 'check_failed' || node.status === 'bke_failed' || node.status === 'schedule_failed' ? $t('初始化失败') : $t('删除失败')}}
                                                </td>
                                                <td>{{node.containerCount}}</td>
                                                <template v-if="curClusterInPage.type === 'k8s'">
                                                    <td>{{node.podCount}}</td>
                                                </template>
                                                <template>
                                                    <td v-if="node.cpuMetric !== null && node.cpuMetric !== undefined"><ring-cell :percent="node.cpuMetric" :fill-color="'#3ede78'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                    <td v-if="node.memMetric !== null && node.memMetric !== undefined"><ring-cell :percent="node.memMetric" :fill-color="'#3c96ff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                    <td v-if="node.diskMetric !== null && node.diskMetric !== undefined"><ring-cell :percent="node.diskMetric" :fill-color="'#853cff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                </template>

                                                <template>
                                                    <td v-if="node.diskioMetric !== null && node.diskioMetric !== undefined"><ring-cell :percent="node.diskioMetric" :fill-color="'#853cff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                </template>

                                                <td style="text-align: left;">
                                                    <a href="javascript:void(0);" class="bk-text-button" @click.stop="showLog(node)">{{$t('查看日志')}}</a>
                                                    <template v-if="node.status === 'delete_failed'">
                                                        <a href="javascript:void(0);" class="bk-text-button" @click.stop="delFailedNode(node, index)">{{$t('删除')}}</a>
                                                        <a href="javascript:void(0);" class="bk-text-button" @click.stop="showFaultRemove(node, index)">{{$t('故障移除')}}</a>
                                                    </template>
                                                    <template v-else-if="node.status === 'remove_failed'">
                                                        <a href="javascript:void(0);" class="bk-text-button" @click.stop="reTryDel(node, index)">{{$t('重试')}}</a>
                                                    </template>
                                                    <template v-else>
                                                        <a href="javascript:void(0);" class="bk-text-button" @click.stop="showDelNode(node, index)">{{$t('删除')}}</a>
                                                        <a href="javascript:void(0);" class="bk-text-button" @click.stop="reInitializationNode(node, index)">{{$t('重试')}}</a>
                                                    </template>
                                                    <bk-dropdown-menu class="dropdown-menu ml10" :align="'center'" ref="dropdown">
                                                        <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button">
                                                            {{$t('更多')}}
                                                            <i class="bk-icon icon-angle-down dropdown-menu-angle-down"></i>
                                                        </a>
                                                        <ul class="bk-dropdown-list" slot="dropdown-content">
                                                            <li>
                                                                <a href="javascript:void(0);" class="bk-text-button" @click.stop="showRecordRemove(node, index)">
                                                                    {{$t('仅移除记录')}}
                                                                </a>
                                                            </li>
                                                        </ul>
                                                    </bk-dropdown-menu>
                                                </td>
                                            </template>

                                            <!--
                                                不可调度: to_removed
                                                操作: 允许调度（启用），删除（没有这个操作，和之前一样，置灰显示 tooltip），强制删除
                                            -->
                                            <template v-if="node.status === 'to_removed'">
                                                <td style="padding-left: 10px;">
                                                    <a href="javascript:void(0)" class="bk-text-button" @click="goNodeOverview(node)">{{node.inner_ip}}</a>
                                                </td>
                                                <td>
                                                    <div class="biz-status-node"><i class="node warning"></i></div>
                                                    {{$t('不可调度')}}
                                                </td>
                                                <td>{{node.containerCount}}</td>
                                                <template v-if="curClusterInPage.type === 'k8s'">
                                                    <td>{{node.podCount}}</td>
                                                </template>
                                                <template>
                                                    <td v-if="node.cpuMetric !== null && node.cpuMetric !== undefined"><ring-cell :percent="node.cpuMetric" :fill-color="'#3ede78'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                    <td v-if="node.memMetric !== null && node.memMetric !== undefined"><ring-cell :percent="node.memMetric" :fill-color="'#3c96ff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                    <td v-if="node.diskMetric !== null && node.diskMetric !== undefined"><ring-cell :percent="node.diskMetric" :fill-color="'#853cff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                </template>

                                                <template>
                                                    <td v-if="node.diskioMetric !== null && node.diskioMetric !== undefined"><ring-cell :percent="node.diskioMetric" :fill-color="'#853cff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                </template>
                                                <td>
                                                    <a href="javascript:void(0);" class="bk-text-button" @click.stop="enableNode(node, index)">{{$t('允许调度')}}</a>
                                                    <template v-if="curClusterInPage.state === 'existing'">
                                                        <bk-tooltip style="margin: 0 15px;" :content="$t('导入的集群不允许删除操作')" placement="top-end">
                                                            <a href="javascript:void(0);" class="bk-text-button is-disabled">{{$t('删除')}}</a>
                                                        </bk-tooltip>
                                                        <bk-tooltip style="margin-right: 15px;" :content="$t('导入的集群不允许强制删除操作')" placement="top-end">
                                                            <a href="javascript:void(0);" class="bk-text-button is-disabled">{{$t('强制删除')}}</a>
                                                        </bk-tooltip>
                                                    </template>
                                                    <template v-else>
                                                        <bk-tooltip style="margin: 0 15px;" :content="$t('请确保该节点已经没有运行中的容器')" placement="top-end">
                                                            <a href="javascript:void(0);" class="bk-text-button is-disabled">{{$t('删除')}}</a>
                                                        </bk-tooltip>
                                                        <a href="javascript:void(0);" class="bk-text-button" style="margin-right: 15px;" @click.stop="showForceDelNode(node, index)">{{$t('强制删除')}}</a>
                                                    </template>
                                                    <bk-dropdown-menu class="dropdown-menu" :align="'center'" ref="dropdown">
                                                        <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button">
                                                            {{$t('更多')}}
                                                            <i class="bk-icon icon-angle-down dropdown-menu-angle-down"></i>
                                                        </a>
                                                        <ul class="bk-dropdown-list" slot="dropdown-content">
                                                            <li>
                                                                <a href="javascript:void(0);" class="bk-text-button" @click.stop="schedulerNode(node, index)">
                                                                    {{(curClusterInPage.type === 'k8s' || curProject.kind === 3) ? $t('pod迁移') : $t('taskgroup迁移')}}
                                                                </a>
                                                            </li>
                                                            <li>
                                                                <a href="javascript:void(0);" class="bk-text-button" @click.stop="showRecordRemove(node, index)">
                                                                    {{$t('仅移除记录')}}
                                                                </a>
                                                            </li>
                                                        </ul>
                                                    </bk-dropdown-menu>
                                                </td>
                                            </template>

                                            <!--
                                                不可调度: removable
                                                操作: 允许调度（启用），删除，强制删除
                                            -->
                                            <template v-if="node.status === 'removable'">
                                                <td style="padding-left: 10px;">
                                                    <a href="javascript:void(0)" class="bk-text-button" @click="goNodeOverview(node)">{{node.inner_ip}}</a>
                                                </td>
                                                <td>
                                                    <div class="biz-status-node"><i class="node warning"></i></div>
                                                    {{$t('不可调度')}}
                                                </td>
                                                <td>{{node.containerCount}}</td>
                                                <template v-if="curClusterInPage.type === 'k8s'">
                                                    <td>{{node.podCount}}</td>
                                                </template>
                                                <template>
                                                    <td v-if="node.cpuMetric !== null && node.cpuMetric !== undefined"><ring-cell :percent="node.cpuMetric" :fill-color="'#3ede78'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                    <td v-if="node.memMetric !== null && node.memMetric !== undefined"><ring-cell :percent="node.memMetric" :fill-color="'#3c96ff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                    <td v-if="node.diskMetric !== null && node.diskMetric !== undefined"><ring-cell :percent="node.diskMetric" :fill-color="'#853cff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                </template>

                                                <template>
                                                    <td v-if="node.diskioMetric !== null && node.diskioMetric !== undefined"><ring-cell :percent="node.diskioMetric" :fill-color="'#853cff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                </template>
                                                <td style="text-align: left;">
                                                    <a href="javascript:void(0);" class="bk-text-button" @click.stop="enableNode(node, index)">{{$t('允许调度')}}</a>
                                                    <template v-if="curClusterInPage.state === 'existing'">
                                                        <bk-tooltip style="margin: 0 15px;" :content="$t('导入的集群不允许删除操作')" placement="top-end">
                                                            <a href="javascript:void(0);" class="bk-text-button is-disabled">{{$t('删除')}}</a>
                                                        </bk-tooltip>
                                                        <bk-tooltip style="margin-right: 15px;" :content="$t('导入的集群不允许强制删除操作')" placement="top-end">
                                                            <a href="javascript:void(0);" class="bk-text-button is-disabled">{{$t('强制删除')}}</a>
                                                        </bk-tooltip>
                                                    </template>
                                                    <template v-else>
                                                        <a href="javascript:void(0);" class="bk-text-button" @click.stop="showDelNode(node, index)">{{$t('删除')}}</a>
                                                        <a href="javascript:void(0);" class="bk-text-button" style="margin-right: 15px;" @click.stop="showForceDelNode(node, index)">{{$t('强制删除')}}</a>
                                                    </template>
                                                    <bk-dropdown-menu class="dropdown-menu" :align="'center'" ref="dropdown">
                                                        <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button">
                                                            {{$t('更多')}}
                                                            <i class="bk-icon icon-angle-down dropdown-menu-angle-down"></i>
                                                        </a>
                                                        <ul class="bk-dropdown-list" slot="dropdown-content">
                                                            <li>
                                                                <a href="javascript:void(0);" class="bk-text-button" @click.stop="schedulerNode(node, index)">
                                                                    {{(curClusterInPage.type === 'k8s' || curProject.kind === 3) ? $t('pod迁移') : $t('taskgroup迁移')}}
                                                                </a>
                                                            </li>
                                                            <li>
                                                                <a href="javascript:void(0);" class="bk-text-button" @click.stop="showRecordRemove(node, index)">
                                                                    {{$t('仅移除记录')}}
                                                                </a>
                                                            </li>
                                                        </ul>
                                                    </bk-dropdown-menu>
                                                </td>
                                            </template>

                                            <!--
                                                不正常: not_ready
                                                操作: 删除，强制删除
                                            -->
                                            <template v-if="node.status === 'not_ready'">
                                                <td style="padding-left: 10px;">
                                                    <a href="javascript:void(0)" class="bk-text-button" @click="goNodeOverview(node)">{{node.inner_ip}}</a>
                                                </td>
                                                <td>
                                                    <div class="biz-status-node"><i class="node danger"></i></div>
                                                    {{$t('不正常')}}
                                                </td>
                                                <td>{{node.containerCount}}</td>
                                                <template v-if="curClusterInPage.type === 'k8s'">
                                                    <td>{{node.podCount}}</td>
                                                </template>
                                                <template>
                                                    <td v-if="node.cpuMetric !== null && node.cpuMetric !== undefined"><ring-cell :percent="node.cpuMetric" :fill-color="'#3ede78'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                    <td v-if="node.memMetric !== null && node.memMetric !== undefined"><ring-cell :percent="node.memMetric" :fill-color="'#3c96ff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                    <td v-if="node.diskMetric !== null && node.diskMetric !== undefined"><ring-cell :percent="node.diskMetric" :fill-color="'#853cff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                </template>

                                                <template>
                                                    <td v-if="node.diskioMetric !== null && node.diskioMetric !== undefined"><ring-cell :percent="node.diskioMetric" :fill-color="'#853cff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                </template>

                                                <td style="text-align: left;">
                                                    <!-- <a href="javascript:void(0);" class="bk-text-button" @click.stop="showDelNode(node, index)">{{$t('删除')}}</a>
                                                    <a href="javascript:void(0);" class="bk-text-button" @click.stop="showForceDelNode(node, index)">{{$t('强制删除')}}</a> -->
                                                    <a href="javascript:void(0);" class="bk-text-button" @click.stop="showFaultRemove(node, index)">{{$t('故障移除')}}</a>
                                                    <bk-dropdown-menu class="dropdown-menu ml10" :align="'center'" ref="dropdown">
                                                        <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button">
                                                            {{$t('更多')}}
                                                            <i class="bk-icon icon-angle-down dropdown-menu-angle-down"></i>
                                                        </a>
                                                        <ul class="bk-dropdown-list" slot="dropdown-content">
                                                            <li>
                                                                <a href="javascript:void(0);" class="bk-text-button" @click.stop="showRecordRemove(node, index)">
                                                                    {{$t('仅移除记录')}}
                                                                </a>
                                                            </li>
                                                        </ul>
                                                    </bk-dropdown-menu>
                                                </td>
                                            </template>

                                            <!--
                                                正常: normal
                                                操作: 停止调度（停止分配）

                                                不正常: unnormal
                                                操作: 停止调度（停止分配）
                                            -->
                                            <template v-if="node.status === 'normal' || node.status === 'unnormal'">
                                                <td style="padding-left: 10px;">
                                                    <a href="javascript:void(0)" class="bk-text-button" @click="goNodeOverview(node)">{{node.inner_ip}}</a>
                                                </td>
                                                <td v-if="node.status === 'normal'">
                                                    <div class="biz-status-node"><i class="node success"></i></div>
                                                    {{$t('正常')}}
                                                </td>
                                                <td v-else>
                                                    <div class="biz-status-node"><i class="node danger"></i></div>
                                                    {{$t('不正常')}}
                                                </td>
                                                <td>{{node.containerCount}}</td>
                                                <template v-if="curClusterInPage.type === 'k8s'">
                                                    <td>{{node.podCount}}</td>
                                                </template>
                                                <template>
                                                    <td v-if="node.cpuMetric !== null && node.cpuMetric !== undefined"><ring-cell :percent="node.cpuMetric" :fill-color="'#3ede78'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                    <td v-if="node.memMetric !== null && node.memMetric !== undefined"><ring-cell :percent="node.memMetric" :fill-color="'#3c96ff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                    <td v-if="node.diskMetric !== null && node.diskMetric !== undefined"><ring-cell :percent="node.diskMetric" :fill-color="'#853cff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                    <td v-if="node.diskioMetric !== null && node.diskioMetric !== undefined"><ring-cell :percent="node.diskioMetric" :fill-color="'#853cff'"></ring-cell></td>
                                                    <td v-else><loading-cell></loading-cell></td>
                                                </template>

                                                <td style="text-align: left;">
                                                    <a href="javascript:void(0);" class="bk-text-button" @click.stop="stopNode(node, index)">{{$t('停止调度')}}</a>
                                                    <bk-dropdown-menu class="dropdown-menu ml10" :align="'center'" ref="dropdown">
                                                        <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button">
                                                            {{$t('更多')}}
                                                            <i class="bk-icon icon-angle-down dropdown-menu-angle-down"></i>
                                                        </a>
                                                        <ul class="bk-dropdown-list" slot="dropdown-content">
                                                            <li>
                                                                <a href="javascript:void(0);" class="bk-text-button" @click.stop="showRecordRemove(node, index)">
                                                                    {{$t('仅移除记录')}}
                                                                </a>
                                                            </li>
                                                        </ul>
                                                    </bk-dropdown-menu>
                                                </td>
                                            </template>
                                        </tr>
                                    </template>
                                    <template v-else>
                                        <tr class="no-hover">
                                            <td colspan="9">
                                                <div class="bk-message-box">
                                                    <p class="message empty-message">{{$t('无数据')}}</p>
                                                </div>
                                            </td>
                                        </tr>
                                    </template>
                                </tbody>
                            </table>
                        </div>
                        <div class="bk-table-footer" v-if="nodeListPageConf.total">
                            <div class="biz-page-wrapper" style="margin: 5px 0;">
                                <bk-page-counter
                                    :is-en="isEn"
                                    :total="nodeListPageConf.total"
                                    :page-size="nodeListPageConf.pageSize"
                                    @change="changePageSize">
                                </bk-page-counter>
                                <bk-paging
                                    style="float: right;"
                                    :cur-page.sync="nodeListPageConf.curPage"
                                    :total-page="nodeListPageConf.totalPage"
                                    @page-change="nodeListPageChange">
                                </bk-paging>
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
            @hidden="closeLog"
            :quick-close="true">
            <div class="p20" slot="content">
                <template v-if="logEndState === 'none'">
                    <div style="margin: 0 0 5px 0; text-align: center;">
                        {{$t('暂无日志信息')}}
                    </div>
                </template>
                <template v-else>
                    <div class="biz-log-box">
                        <div class="operation-item" v-for="(op, index) in logList" :key="index">
                            <p class="log-message title">
                                {{op.prefix_message}}
                            </p>
                            <template v-if="op.log.node_tasks">
                                <p class="log-message item" v-for="(task, taskIndex) in op.log.node_tasks" :key="taskIndex">
                                    <template v-if="op.prefix_message.indexOf($t('前置检查')) > -1">
                                        === <span>{{task.name}}</span> start ===
                                        <br />
                                    </template>
                                    <template v-else>
                                        {{task.name}} -
                                    </template>
                                    <span v-if="task.state.toLowerCase() === 'failure'" class="biz-danger-text">
                                        {{task.state}}
                                    </span>
                                    <span v-else-if="task.state.toLowerCase() === 'success'" class="biz-success-text">
                                        {{task.state}}
                                    </span>
                                    <span v-else-if="task.state.toLowerCase() === 'running'" class="biz-warning-text">
                                        {{task.state}}
                                    </span>
                                    <span v-else v-html="formatLog(task.state)">
                                    </span>
                                </p>
                            </template>
                            <div v-if="op.status.toLowerCase() === 'success'" style="margin: 0 0 5px 0; color: #34d97b; font-size: 14px; font-weight: 700; margin-left: 20px;">
                                {{$t('操作成功')}}
                            </div>
                            <div v-else-if="op.status.toLowerCase() === 'failed'" style="margin: 0 0 5px 0; color: #e64d34; font-size: 14px; font-weight: 700; margin-left: 20px;">
                                {{$t('操作失败')}}<span style="margin-left: 10px;" v-if="op.taskUrl"><a :href="op.taskUrl" class="bk-text-button" target="_blank">{{$t('查看详情')}}</a></span>
                            </div>
                            <div style="margin: 10px 0px 5px 13px; font-size: 10px;" v-else>
                                <div class="bk-spin-loading bk-spin-loading-small bk-spin-loading-primary">
                                    <div class="rotate rotate1"></div>
                                    <div class="rotate rotate2"></div>
                                    <div class="rotate rotate3"></div>
                                    <div class="rotate rotate4"></div>
                                    <div class="rotate rotate5"></div>
                                    <div class="rotate rotate6"></div>
                                    <div class="rotate rotate7"></div>
                                    <div class="rotate rotate8"></div>
                                </div>
                                {{$t('正在加载中...')}}
                            </div>
                        </div>
                    </div>
                </template>
            </div>
        </bk-sideslider>

        <tip-dialog
            ref="nodeNoticeDialog"
            icon="bk-icon icon-exclamation-triangle"
            :title="$t('添加节点')"
            :sub-title="$t('此操作需要对你的主机进行如下操作，请知悉：')"
            :check-list="nodeNoticeList"
            :confirm-btn-text="$t('确定，添加节点')"
            :cancel-btn-text="$t('我再想想')"
            :confirm-callback="saveNode">
        </tip-dialog>

        <tip-dialog
            ref="removeNodeDialog"
            icon="bk-icon icon-exclamation-triangle"
            :show-close="false"
            :sub-title="$t('此操作无法撤回，请确认：')"
            :check-list="deleteNodeNoticeList"
            :confirm-btn-text="$t('确定')"
            :confirming-btn-text="$t('删除中...')"
            :canceling-btn-text="$t('取消')"
            :confirm-callback="confirmDelNode"
            :cancel-callback="cancelDelNode">
        </tip-dialog>

        <tip-dialog
            ref="forceRemoveNodeDialog"
            icon="bk-icon icon-exclamation-triangle"
            :show-close="false"
            :sub-title="$t('此操作无法撤回，请确认：')"
            :check-list="deleteNodeNoticeList"
            :confirm-btn-text="$t('确定')"
            :confirming-btn-text="$t('删除中...')"
            :canceling-btn-text="$t('取消')"
            :confirm-callback="confirmForceRemoveNode"
            :cancel-callback="cancelForceRemoveNode">
        </tip-dialog>

        <tip-dialog
            ref="faultRemoveDialog"
            icon="bk-icon icon-exclamation-triangle"
            :show-close="false"
            sub-title="此操作无法撤回，请确认： "
            :check-list="faultRemoveoticeList"
            :confirm-btn-text="$t('确定')"
            :confirming-btn-text="$t('删除中...')"
            :canceling-btn-text="$t('取消')"
            :confirm-callback="confirmFaultRemove"
            :cancel-callback="cancelFaultRemove">
        </tip-dialog>

        <tip-dialog
            ref="recordRemoveDialog"
            icon="bk-icon icon-exclamation-triangle"
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
            <template slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary disabled">
                            {{$t('初始化中...')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </button>
                    </template>
                    <template v-else>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary" @click="reInitializationConfirm">
                            {{$t('确定')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel" @click="reInitializationCancel">
                            {{$t('取消')}}
                        </button>
                    </template>
                </div>
            </template>
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
            <template slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary disabled">
                            {{$t('删除中...')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </button>
                    </template>
                    <template v-else>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary" @click="reDelConfirm">
                            {{$t('确定')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel" @click="reDelCancel">
                            {{$t('取消')}}
                        </button>
                    </template>
                </div>
            </template>
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
            <template slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary disabled">
                            {{$t('删除中...')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </button>
                    </template>
                    <template v-else>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary" @click="delConfirm">
                            {{$t('确定')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel" @click="delCancel">
                            {{$t('取消')}}
                        </button>
                    </template>
                </div>
            </template>
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
            <template slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary disabled">
                            {{$t('启用中...')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </button>
                    </template>
                    <template v-else>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary" @click="enableConfirm">
                            {{$t('确定')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel" @click="enableCancel">
                            {{$t('取消')}}
                        </button>
                    </template>
                </div>
            </template>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="stopDialogConf.isShow"
            :width="stopDialogConf.width"
            :title="stopDialogConf.title"
            :close-icon="stopDialogConf.closeIcon"
            :ext-cls="'biz-node-re-initialization-dialog'"
            :quick-close="false">
            <template slot="content">
                <div>{{stopDialogConf.content}}</div>
            </template>
            <template slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary disabled">
                            {{$t('停用中...')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </button>
                    </template>
                    <template v-else>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary" @click="stopConfirm">
                            {{$t('确定')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel" @click="stopCancel">
                            {{$t('取消')}}
                        </button>
                    </template>
                </div>
            </template>
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
            <template slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary disabled">
                            {{$t('删除中...')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </button>
                    </template>
                    <template v-else>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary" @click="removeConfirm">
                            {{$t('确定')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel" @click="removeCancel">
                            {{$t('取消')}}
                        </button>
                    </template>
                </div>
            </template>
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
            <template slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary disabled">
                            {{$t('迁移中...')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </button>
                    </template>
                    <template v-else>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary" @click="schedulerConfirm">
                            {{$t('确定')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel" @click="schedulerCancel">
                            {{$t('取消')}}
                        </button>
                    </template>
                </div>
            </template>
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
            <template slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isUpdating">
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary disabled">
                            {{$t('操作中...')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
                            {{$t('取消')}}
                        </button>
                    </template>
                    <template v-else>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary"
                            @click="batchConfirm">
                            {{$t('确定')}}
                        </button>
                        <button type="button" class="bk-dialog-btn bk-dialog-btn-cancel" @click="batchCancel">
                            {{$t('取消')}}
                        </button>
                    </template>
                </div>
            </template>
        </bk-dialog>

    </div>
</template>

<script>
    import applyPerm from '@open/mixins/apply-perm'
    import tipDialog from '@open/components/tip-dialog'
    import RingCell from './ring-cell'
    import LoadingCell from './loading-cell'
    import mixin from './mixin-node'
    import nodeSearcher from './searcher'
    import IpSelector from '@/components/ip-selector/selector-dialog.vue'

    export default {
        components: {
            RingCell,
            LoadingCell,
            tipDialog,
            nodeSearcher,
            IpSelector
        },
        mixins: [applyPerm, mixin],
        data () {
            return {
                nodeNoticeList: [
                    {
                        id: 2,
                        text: this.$t('按照规则修改主机名'),
                        isChecked: true
                    },
                    {
                        id: 3,
                        text: this.$t('安装容器服务相关的组件'),
                        isChecked: true
                    }
                ]
            }
        }
    }
</script>

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
    /* .bk-dialog-footer .bk-dialog-outer button {
        margin-top: 30px;
    } */
</style>
