<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-topbar-title">
                Ingress
                <span class="biz-tip ml10">{{$t('请通过模板集或Helm创建Ingress')}}</span>
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper p0" v-bkloading="{ isLoading: isInitLoading, opacity: 0.1 }">
            <template v-if="!isInitLoading">
                <div class="biz-panel-header">
                    <div class="left">
                        <bk-button
                            class="bk-button bk-default"
                            v-if="curPageData.length"
                            @click.stop.prevent="removeIngresses">
                            <span>{{$t('批量删除')}}</span>
                        </bk-button>
                    </div>
                    <div class="right">
                        <bk-data-searcher
                            :scope-list="searchScopeList"
                            :search-key.sync="searchKeyword"
                            :search-scope.sync="searchScope"
                            :cluster-fixed="!!curClusterId"
                            @search="getIngressList"
                            @refresh="refresh">
                        </bk-data-searcher>
                    </div>
                </div>

                <div class="biz-resource">
                    <div class="biz-table-wrapper">
                        <bk-table
                            :size="'medium'"
                            :data="curPageData"
                            :pagination="pageConf"
                            v-bkloading="{ isLoading: isPageLoading && !isInitLoading, opacity: 1 }"
                            @page-limit-change="handlePageLimitChange"
                            @page-change="handlePageChange"
                            @select="handlePageSelect"
                            @select-all="handlePageSelectAll">
                            <bk-table-column type="selection" width="60" :selectable="rowSelectable"></bk-table-column>
                            <bk-table-column :label="$t('名称')" :show-overflow-tooltip="true" min-width="200">
                                <template slot-scope="props">
                                    <a href="javascript: void(0)"
                                        class="bk-text-button biz-resource-title"
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
                                        @click.stop.prevent="showIngressDetail(props.row, index)"
                                    >{{props.row.resourceName}}</a>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('所属集群')" min-width="150">
                                <template slot-scope="props">
                                    <bcs-popover :content="props.row.cluster_id || '--'" placement="top">
                                        <p class="biz-text-wrapper">{{curCluster ? curCluster.clusterName : '--'}}</p>
                                    </bcs-popover>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('命名空间')" min-width="130">
                                <template slot-scope="props">
                                    {{props.row.namespace}}
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('来源')" min-width="130">
                                <template slot-scope="props">
                                    {{props.row.source_type ? props.row.source_type : '--'}}
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('创建时间')" min-width="160">
                                <template slot-scope="props">
                                    {{props.row.createTime ? formatDate(props.row.createTime) : '--'}}
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('更新时间')" min-width="160">
                                <template slot-scope="props">
                                    {{props.row.updateTime ? formatDate(props.row.updateTime) : '--'}}
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('更新人')" min-width="100">
                                <template slot-scope="props">
                                    {{props.row.updator || '--'}}
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('操作')" width="150">
                                <template slot-scope="props">
                                    <a v-if="props.row.can_update"
                                        href="javascript:void(0);"
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
                                        @click="showIngressEditDialog(props.row)"
                                    >{{$t('更新')}}</a>
                                    <bcs-popover :content="props.row.can_update_msg" v-else placement="left">
                                        <a href="javascript:void(0);" class="bk-text-button is-disabled">{{$t('更新')}}</a>
                                    </bcs-popover>
                                    <a v-if="props.row.can_delete"
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
                                        @click.stop="removeIngress(props.row)"
                                        class="bk-text-button ml10"
                                    >{{$t('删除')}}</a>
                                    <bcs-popover :content="props.row.can_delete_msg || $t('不可删除')" v-else placement="left">
                                        <span class="bk-text-button is-disabled ml10">{{$t('删除')}}</span>
                                    </bcs-popover>
                                </template>
                            </bk-table-column>
                        </bk-table>
                    </div>
                </div>
            </template>

            <bk-sideslider
                v-if="curIngress"
                :quick-close="true"
                :is-show.sync="ingressSlider.isShow"
                :title="ingressSlider.title"
                :width="'800'">
                <div class="pt20 pr30 pb20 pl30" slot="content">
                    <label class="biz-title">{{$t('主机列表')}}（spec.tls）</label>
                    <table class="bk-table biz-data-table has-table-bordered biz-special-bk-table">
                        <thead>
                            <tr>
                                <th style="width: 270px;">{{$t('主机名')}}</th>
                                <th>SecretName</th>
                            </tr>
                        </thead>
                        <tbody>
                            <template v-if="curIngress.tls.length">
                                <tr v-for="(rule, index) in curIngress.tls" :key="index">
                                    <td>{{rule.host || '--'}}</td>
                                    <td>{{rule.secretName || '--'}}</td>
                                </tr>
                            </template>
                            <template v-else>
                                <tr>
                                    <td colspan="2"><bcs-exception type="empty" scene="part"></bcs-exception></td>
                                </tr>
                            </template>
                        </tbody>
                    </table>

                    <label class="biz-title">{{$t('规则')}}（spec.rules）</label>
                    <table class="bk-table biz-data-table has-table-bordered biz-special-bk-table">
                        <thead>
                            <tr>
                                <th style="width: 200px;">{{$t('主机名')}}</th>
                                <th style="width: 150px;">{{$t('路径')}}</th>
                                <th>{{$t('服务名称')}}</th>
                                <th style="width: 100px;">{{$t('服务端口')}}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <template v-if="curIngress.rules.length">
                                <tr v-for="(rule, index) in curIngress.rules" :key="index">
                                    <td>{{rule.host || '--'}}</td>
                                    <td>{{rule.path || '--'}}</td>
                                    <td>{{rule.serviceName || '--'}}</td>
                                    <td>{{rule.servicePort || '--'}}</td>
                                </tr>
                            </template>
                            <template v-else>
                                <tr>
                                    <td colspan="4"><bcs-exception type="empty" scene="part"></bcs-exception></td>
                                </tr>
                            </template>
                        </tbody>
                    </table>

                    <div class="actions">
                        <bk-button class="show-labels-btn bk-button bk-button-small bk-primary">{{$t('显示标签')}}</bk-button>
                    </div>

                    <div class="point-box">
                        <template v-if="curIngress.labels.length">
                            <ul class="key-list">
                                <li v-for="(label, index) in curIngress.labels" :key="index">
                                    <span class="key">{{label[0]}}</span>
                                    <span class="value">{{label[1] || '--'}}</span>
                                </li>
                            </ul>
                        </template>
                        <template v-else>
                            <bcs-exception type="empty" scene="part"></bcs-exception>
                        </template>
                    </div>
                </div>
            </bk-sideslider>

            <bk-sideslider
                :is-show.sync="ingressEditSlider.isShow"
                :title="ingressEditSlider.title"
                :width="'1020'"
                @hidden="handleCancelUpdate">
                <div slot="content">
                    <div class="bk-form biz-configuration-form pt20 pb20 pl10 pr20">
                        <div class="bk-form-item">
                            <div class="bk-form-item">
                                <div class="bk-form-content" style="margin-left: 0;">
                                    <div class="bk-form-item is-required">
                                        <label class="bk-label" style="width: 130px;">{{$t('名称')}}：</label>
                                        <div class="bk-form-content" style="margin-left: 130px;">
                                            <bk-input
                                                :disabled="true"
                                                style="width: 310px;"
                                                v-model="curEditedIngress.config.metadata.name"
                                                maxlength="64"
                                                name="applicationName" />
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div class="bk-form-item">
                                <div class="bk-form-content" style="margin-left: 130px;">
                                    <button :class="['bk-text-button f12 mb10 pl0', { 'rotate': isTlsPanelShow }]" @click.stop.prevent="toggleTlsPanel">
                                        {{$t('TLS设置')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                                    </button>
                                    <button :class="['bk-text-button f12 mb10 pl0', { 'rotate': isPanelShow }]" @click.stop.prevent="togglePanel">
                                        {{$t('更多设置')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                                    </button>
                                </div>
                            </div>

                            <div class="bk-form-item mt0" v-show="isTlsPanelShow">
                                <div class="bk-form-content" style="margin-left: 130px;">
                                    <bk-tab :type="'fill'" :active-name="'tls'" :size="'small'">
                                        <bk-tab-panel name="tls" title="TLS">
                                            <div class="p20">
                                                <table class="biz-simple-table">
                                                    <tbody>
                                                        <tr v-for="(computer, index) in curEditedIngress.config.spec.tls" :key="index">
                                                            <td>
                                                                <bkbcs-input
                                                                    type="text"
                                                                    :placeholder="$t('主机名，多个用英文逗号分隔')"
                                                                    style="width: 310px;"
                                                                    :value.sync="computer.hosts"
                                                                    :list="varList"
                                                                >
                                                                </bkbcs-input>
                                                            </td>
                                                            <td>
                                                                <bkbcs-input
                                                                    type="text"
                                                                    :placeholder="$t('请输入证书')"
                                                                    style="width: 350px;"
                                                                    :value.sync="computer.secretName"
                                                                >
                                                                </bkbcs-input>
                                                            </td>
                                                            <td>
                                                                <bk-button class="action-btn ml5" @click.stop.prevent="addTls">
                                                                    <i class="bcs-icon bcs-icon-plus"></i>
                                                                </bk-button>
                                                                <bk-button class="action-btn" v-if="curEditedIngress.config.spec.tls.length > 1" @click.stop.prevent="removeTls(index, computer)">
                                                                    <i class="bcs-icon bcs-icon-minus"></i>
                                                                </bk-button>
                                                            </td>
                                                        </tr>
                                                    </tbody>
                                                </table>
                                            </div>
                                        </bk-tab-panel>
                                    </bk-tab>
                                </div>
                            </div>

                            <div class="bk-form-item mt0" v-show="isPanelShow">
                                <div class="bk-form-content" style="margin-left: 130px;">
                                    <bk-tab :type="'fill'" :active-name="'remark'" :size="'small'">
                                        <bk-tab-panel name="remark" :title="$t('注解')">
                                            <div class="biz-tab-wrapper m20">
                                                <bk-keyer :key-list.sync="curRemarkList" :var-list="varList" ref="remarkKeyer"></bk-keyer>
                                            </div>
                                        </bk-tab-panel>
                                        <bk-tab-panel name="label" :title="$t('标签')">
                                            <div class="biz-tab-wrapper m20">
                                                <bk-keyer :key-list.sync="curLabelList" :var-list="varList" ref="labelKeyer"></bk-keyer>
                                            </div>
                                        </bk-tab-panel>
                                    </bk-tab>
                                </div>
                            </div>

                            <!-- part2 start -->
                            <div class="biz-part-header">
                                <div class="bk-button-group">
                                    <div class="item" v-for="(rule, index) in curEditedIngress.config.spec.rules" :key="index">
                                        <bk-button :class="['bk-button bk-default is-outline', { 'is-selected': curRuleIndex === index }]" @click.stop="setCurRule(rule, index)">
                                            {{rule.host || $t('未命名')}}
                                        </bk-button>
                                        <span class="bcs-icon bcs-icon-close-circle" @click.stop="removeRule(index)" v-if="curEditedIngress.config.spec.rules.length > 1"></span>
                                    </div>
                                    <bcs-popover ref="containerTooltip" :content="curEditedIngress.config.spec.rules.length >= 5 ? $t('最多添加5个') : $t('添加Rule')" placement="top">
                                        <bk-button type="button" class="bk-button bk-default is-outline is-icon" :disabled="curEditedIngress.config.spec.rules.length >= 5 " @click.stop.prevent="addLocalRule">
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
                                    <label class="bk-label" style="width: 130px;">{{$t('虚拟主机名')}}：</label>
                                    <div class="bk-form-content" style="margin-left: 130px;">
                                        <bk-input :placeholder="$t('请输入')" style="width: 310px;" v-model="curRule.host" name="ruleName" />
                                    </div>
                                </div>
                                <div class="bk-form-item">
                                    <label class="bk-label" style="width: 130px;">{{$t('路径组')}}：</label>
                                    <div class="bk-form-content" style="margin-left: 130px;">
                                        <table class="biz-simple-table">
                                            <tbody>
                                                <tr v-for="(pathRule, index) of curRule.http.paths" :key="index">
                                                    <td>
                                                        <bkbcs-input
                                                            type="text"
                                                            :placeholder="$t('路径')"
                                                            style="width: 310px;"
                                                            :value.sync="pathRule.path"
                                                            :list="varList"
                                                        >
                                                        </bkbcs-input>
                                                    </td>
                                                    <td style="text-align: center;">
                                                        <i class="bcs-icon bcs-icon-arrows-right"></i>
                                                    </td>
                                                    <td>
                                                        <bk-selector
                                                            style="width: 180px;"
                                                            :placeholder="$t('Service名称')"
                                                            :disabled="isLoadBalanceEdited"
                                                            :setting-key="'_name'"
                                                            :display-key="'_name'"
                                                            :selected.sync="pathRule.backend.serviceName"
                                                            :list="linkServices || []"
                                                            @item-selected="handlerSelectService(pathRule)">
                                                        </bk-selector>
                                                    </td>
                                                    <td>
                                                        <bk-selector
                                                            style="width: 180px;"
                                                            :placeholder="$t('端口')"
                                                            :disabled="isLoadBalanceEdited"
                                                            :setting-key="'_id'"
                                                            :display-key="'_name'"
                                                            :selected.sync="pathRule.backend.servicePort"
                                                            :list="linkServices[pathRule.backend.serviceName] || []">
                                                        </bk-selector>
                                                    </td>
                                                    <td>
                                                        <bk-button class="action-btn ml5" @click.stop.prevent="addRulePath">
                                                            <i class="bcs-icon bcs-icon-plus"></i>
                                                        </bk-button>
                                                        <bk-button class="action-btn" v-if="curRule.http.paths.length > 1" @click.stop.prevent="removeRulePath(pathRule, index)">
                                                            <i class="bcs-icon bcs-icon-minus"></i>
                                                        </bk-button>
                                                    </td>
                                                </tr>
                                            </tbody>
                                        </table>
                                        <p class="biz-tip">{{$t('提示：同一个虚拟主机名可以有多个路径')}}</p>
                                    </div>
                                </div>
                            </div>

                            <div class="bk-form-item mt25" style="margin-left: 130px;">
                                <bk-button type="primary" :loading="isDetailSaving" @click.stop.prevent="saveIngressDetail">{{$t('保存并更新')}}</bk-button>
                                <bk-button :loading="isDetailSaving" @click.stop.prevent="handleCancelUpdate">{{$t('取消')}}</bk-button>
                            </div>

                        </div>
                    </div>
                </div>
            </bk-sideslider>

            <bk-dialog
                :title="$t('确认删除')"
                :header-position="'left'"
                :is-show="batchDialogConfig.isShow"
                :width="600"
                :has-header="false"
                :quick-close="false"
                @confirm="deleteIngresses(batchDialogConfig.data)"
                @cancel="batchDialogConfig.isShow = false">
                <template slot="content">
                    <div class="biz-batch-wrapper">
                        <p class="batch-title mt10 f14">{{$t('确定要删除以下Ingress？')}}</p>
                        <ul class="batch-list">
                            <li v-for="(item, index) of batchDialogConfig.list" :key="index">{{item}}</li>
                        </ul>
                    </div>
                </template>
            </bk-dialog>
        </div>
    </div>
</template>

<script>
    import { catchErrorHandler, formatDate } from '@/common/util'
    import ingressParams from '@/json/k8s-ingress.json'
    import ruleParams from '@/json/k8s-ingress-rule.json'
    import bkKeyer from '@/components/keyer'

    export default {
        components: {
            bkKeyer
        },
        data () {
            return {
                formatDate: formatDate,
                isInitLoading: true,
                isPageLoading: false,
                searchKeyword: '',
                searchScope: '',
                curPageData: [],
                curIngress: null,
                curEditedIngress: ingressParams,
                isPanelShow: false,
                isTlsPanelShow: false,
                isDetailSaving: false,
                pageConf: {
                    count: 1,
                    totalPage: 1,
                    limit: 10,
                    current: 1,
                    show: true
                },
                ingressSlider: {
                    title: '',
                    isShow: false
                },
                batchDialogConfig: {
                    isShow: false,
                    list: [],
                    data: []
                },
                curRuleIndex: 0,
                curRule: ingressParams.config.spec.rules[0],
                curIngressName: '',
                alreadySelectedNums: 0,
                ingressEditSlider: {
                    title: '',
                    isShow: false
                },
                linkServices: [],
                ingressSelectedList: [],
                webAnnotations: { perms: {} }
            }
        },
        computed: {
            isEn () {
                return this.$store.state.isEn
            },
            curProject () {
                return this.$store.state.curProject
            },
            searchScopeList () {
                const clusterList = this.$store.state.cluster.clusterList
                const results = clusterList.map(item => {
                    return {
                        id: item.cluster_id,
                        name: item.name
                    }
                })

                return results
            },
            isCheckCurPageAll () {
                if (this.curPageData.length) {
                    const list = this.curPageData
                    const selectList = list.filter((item) => {
                        return item.isChecked === true
                    })
                    const canSelectList = list.filter((item) => {
                        return item.can_delete
                    })
                    if (selectList.length && (selectList.length === canSelectList.length)) {
                        return true
                    } else {
                        return false
                    }
                } else {
                    return false
                }
            },
            projectId () {
                return this.$route.params.projectId
            },
            ingressList () {
                const list = this.$store.state.resource.ingressList
                list.forEach(item => {
                    item.isChecked = false
                })
                return JSON.parse(JSON.stringify(list))
            },
            isClusterDataReady () {
                return this.$store.state.cluster.isClusterDataReady
            },
            varList () {
                const list = this.$store.state.variable.varList.map(item => {
                    item._id = item.key
                    item._name = item.key
                    return item
                })
                return list
            },
            curLabelList () {
                const list = []
                const labels = this.curEditedIngress.config.metadata.labels
                for (const [key, value] of Object.entries(labels)) {
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
            curRemarkList () {
                const list = []
                const annotations = this.curEditedIngress.config.metadata.annotations
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
            curClusterId () {
                return this.$store.state.curClusterId
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

                            this.getIngressList()
                        }, 1000)
                    }
                }
            },
            curClusterId () {
                this.searchScope = this.curClusterId
                this.getIngressList()
            }
        },
        created () {
            this.initPageConf()
        },
        methods: {
            /**
             * 刷新列表
             */
            refresh () {
                this.pageConf.current = 1
                this.isPageLoading = true
                this.getIngressList()
            },

            /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
            handlePageLimitChange (pageSize) {
                this.pageConf.limit = pageSize
                this.pageConf.current = 1
                this.initPageConf()
                this.handlePageChange()
            },

            /**
             * 确认批量删除
             */
            async removeIngresses () {
                const data = []
                const names = []

                this.ingressSelectedList.forEach(item => {
                    data.push({
                        cluster_id: item.cluster_id,
                        namespace: item.namespace,
                        name: item.name
                    })
                    names.push(`${item.cluster_name} / ${item.namespace} / ${item.resourceName}`)
                })

                if (!data.length) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择要删除的Ingress')
                    })
                    return false
                }

                this.batchDialogConfig.list = names
                this.batchDialogConfig.data = data
                this.batchDialogConfig.isShow = true
            },

            /**
             * 批量删除
             * @param  {object} data ingresses
             */
            async deleteIngresses (data) {
                const me = this
                const projectId = this.projectId

                this.batchDialogConfig.isShow = false
                this.isPageLoading = true
                try {
                    await this.$store.dispatch('resource/deleteIngresses', { projectId, data })
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('删除成功')
                    })
                    // 稍晚一点加载数据，接口不一定立即清除
                    setTimeout(() => {
                        me.getIngressList()
                    }, 500)
                } catch (e) {
                    // 4004，已经被删除过，但接口不能立即清除，再重新拉数据，防止重复删除
                    if (e.code === 4004) {
                        me.isPageLoading = true
                        setTimeout(() => {
                            me.getIngressList()
                        }, 500)
                    } else {
                        this.isPageLoading = false
                    }
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 确认删除ingress
             * @param  {object} ingress ingress
             */
            async removeIngress (ingress) {
                const me = this
                me.$bkInfo({
                    title: me.$t('确认删除'),
                    clsName: 'biz-remove-dialog max-size',
                    content: me.$createElement('p', {
                        class: 'biz-confirm-desc'
                    }, `${this.$t('确定要删除Ingress')}【${ingress.cluster_name} / ${ingress.namespace} / ${ingress.name}】？`),
                    confirmFn () {
                        me.deleteIngress(ingress)
                    }
                })
            },

            /**
             * 删除ingress
             * @param  {object} ingress ingress
             */
            async deleteIngress (ingress) {
                const me = this
                const projectId = me.projectId
                const clusterId = ingress.cluster_id
                const namespace = ingress.namespace
                const name = ingress.name

                this.isPageLoading = true
                try {
                    await this.$store.dispatch('resource/deleteIngress', {
                        projectId,
                        clusterId,
                        namespace,
                        name
                    })
                    me.$bkMessage({
                        theme: 'success',
                        message: this.$t('删除成功')
                    })

                    // 稍晚一点加载数据，接口不一定立即清除
                    setTimeout(() => {
                        me.getIngressList()
                    }, 500)
                } catch (e) {
                    this.isPageLoading = false
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 显示ingress详情
             * @param  {object} ingress object
             * @param  {number} index 索引
             */
            showIngressDetail (ingress, index) {
                this.ingressSlider.title = ingress.resourceName
                this.curIngress = ingress
                this.ingressSlider.isShow = true
            },

            /**
             * 清除选择，在分页改变时触发
             */
            clearSelectIngress () {
                this.curPageData.forEach((item) => {
                    item.isChecked = false
                })
            },

            /**
             * 获取Ingresslist
             */
            async getIngressList () {
                const projectId = this.projectId
                const params = {
                    cluster_id: this.searchScope
                }
                try {
                    this.isPageLoading = true
                    const res = await this.$store.dispatch('resource/getIngressList', {
                        projectId,
                        params
                    })
                    this.webAnnotations = res.web_annotations || { perms: {} }

                    this.initPageConf()
                    this.curPageData = this.getDataByPage(this.pageConf.current)

                    // 如果有搜索关键字，继续显示过滤后的结果
                    if (this.searchKeyword) {
                        this.searchIngress()
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
                    setTimeout(() => {
                        this.isPageLoading = false
                        this.isInitLoading = false
                    }, 200)
                }
            },

            /**
             * 清除搜索
             */
            clearSearch () {
                this.searchKeyword = ''
                this.searchIngress()
            },

            /**
             * 搜索Ingress
             */
            searchIngress () {
                const keyword = this.searchKeyword.trim()
                const keyList = ['resourceName', 'namespace', 'cluster_name']
                let list = JSON.parse(JSON.stringify(this.$store.state.resource.ingressList))
                const results = []

                if (this.searchScope) {
                    list = list.filter(item => {
                        return item.cluster_id === this.searchScope
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

                this.ingressList.splice(0, this.ingressList.length, ...results)
                this.pageConf.current = 1
                this.initPageConf()
                this.curPageData = this.getDataByPage(this.pageConf.current)
            },

            /**
             * 初始化分页配置
             */
            initPageConf () {
                const total = this.ingressList.length
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
                if (page < 1) {
                    this.pageConf.current = page = 1
                }
                let startIndex = (page - 1) * this.pageConf.limit
                let endIndex = page * this.pageConf.limit
                this.isPageLoading = true
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.ingressList.length) {
                    endIndex = this.ingressList.length
                }
                setTimeout(() => {
                    this.isPageLoading = false
                }, 200)
                this.ingressSelectedList = []
                return this.ingressList.slice(startIndex, endIndex)
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
             * 每行的多选框点击事件
             */
            rowClick () {
                this.$nextTick(() => {
                    this.alreadySelectedNums = this.ingressList.filter(item => item.isChecked).length
                })
            },

            async showIngressEditDialog (ingress) {
                if (!ingress.data.spec.hasOwnProperty('tls')) {
                    ingress.data.spec.tls = [
                        {
                            hosts: '',
                            secretName: ''
                        }
                    ]
                } else if (JSON.stringify(ingress.data.spec.tls) === '[{}]') {
                    ingress.data.spec.tls = [
                        {
                            hosts: '',
                            secretName: ''
                        }
                    ]
                }
                const ingressClone = JSON.parse(JSON.stringify(ingress))
                ingressClone.data.spec.tls.forEach(item => {
                    if (item.hosts && item.hosts.join) {
                        item.hosts = item.hosts.join(',')
                    }
                })
                this.curEditedIngress = ingressClone
                this.curEditedIngress.config = ingressClone.data
                this.ingressEditSlider.title = ingress.name
                delete this.curEditedIngress.data

                if (this.curEditedIngress.config.spec.rules.length) {
                    // 初始化数据放在后面使用报错
                    const rule = Object.assign({
                        http: {
                            paths: [
                                {
                                    backend: {
                                        serviceName: '',
                                        servicePort: ''
                                    },
                                    path: ''
                                }
                            ]
                        }
                    }, this.curEditedIngress.config.spec.rules[0])
                    this.setCurRule(rule, 0)
                } else {
                    this.addLocalRule()
                }
                this.getServiceList(ingress.cluster_id, ingress.namespace_id)
                this.ingressEditSlider.isShow = true
            },

            togglePanel () {
                this.isTlsPanelShow = false
                this.isPanelShow = !this.isPanelShow
            },
            toggleTlsPanel () {
                this.isPanelShow = false
                this.isTlsPanelShow = !this.isTlsPanelShow
            },
            goCertList () {
                if (this.certListUrl) {
                    window.open(this.certListUrl)
                }
            },
            addTls () {
                this.curEditedIngress.config.spec.tls.push({
                    hosts: '',
                    secretName: ''
                })
            },
            removeTls (index, curTls) {
                this.curEditedIngress.config.spec.tls.splice(index, 1)
            },
            setCurRule (rule, index) {
                this.curRule = rule
                this.curRuleIndex = index
            },
            removeRule (index) {
                const rules = this.curEditedIngress.config.spec.rules
                rules.splice(index, 1)
                if (this.curRuleIndex === index) {
                    this.curRuleIndex = 0
                } else if (this.curRuleIndex !== 0) {
                    this.curRuleIndex = this.curRuleIndex - 1
                }

                this.curRule = rules[this.curRuleIndex]
            },
            addLocalRule () {
                const rule = JSON.parse(JSON.stringify(ruleParams))
                const rules = this.curEditedIngress.config.spec.rules
                const index = rules.length
                rule.host = 'rule-' + (index + 1)
                rules.push(rule)
                this.setCurRule(rule, index)
            },
            addRulePath () {
                const params = {
                    backend: {
                        serviceName: '',
                        servicePort: ''
                    },
                    path: ''
                }

                this.curRule.http.paths.push(params)
            },
            removeRulePath (pathRule, index) {
                this.curRule.http.paths.splice(index, 1)
            },
            handlerSelectService (pathRule) {
                pathRule.backend.servicePort = ''
            },
            async initServices (version) {
                const projectId = this.projectId
                await this.$store.dispatch('k8sTemplate/getServicesByVersion', { projectId, version })
            },
            /**
             * 获取service列表
             */
            async getServiceList (clusterId, namespaceId) {
                const projectId = this.projectId
                const params = {
                    cluster_id: clusterId
                }

                try {
                    const res = await this.$store.dispatch('network/getServiceList', {
                        projectId,
                        params
                    })

                    const serviceList = res.data.filter(service => {
                        return service.namespace_id === namespaceId
                    }).map(service => {
                        const ports = service.data.spec.ports || []
                        return {
                            _name: service.resourceName,
                            service_name: service.resourceName,
                            service_ports: ports
                        }
                    })
                    serviceList.forEach(service => {
                        serviceList[service.service_name] = []
                        service.service_ports.forEach(item => {
                            serviceList[service.service_name].push({
                                _id: item.port,
                                _name: item.port
                            })
                        })
                    })
                    this.linkServices = serviceList
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            checkData () {
                const ingress = this.curEditedIngress
                const nameReg = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/
                const pathReg = /\/((?!\.)[\w\d\-./~]+)*/
                let megPrefix = ''

                for (const rule of ingress.config.spec.rules) {
                    // 检查rule
                    if (!rule.host) {
                        megPrefix += this.$t('规则：')
                        this.$bkMessage({
                            theme: 'error',
                            message: megPrefix + this.$t('主机名不能为空')
                        })
                        return false
                    }

                    if (!nameReg.test(rule.host)) {
                        megPrefix += this.$t('规则主机名：')
                        this.$bkMessage({
                            theme: 'error',
                            message: megPrefix + this.$t('名称错误，只能包含：小写字母、数字、连字符(-)，首字母必须是字母'),
                            delay: 8000
                        })
                        return false
                    }

                    const paths = rule.http?.paths || []

                    for (const path of paths) {
                        if (!path.path) {
                            megPrefix += this.$t('{host}中路径组：', { host: rule.host })
                            this.$bkMessage({
                                theme: 'error',
                                message: megPrefix + this.$t('请填写路径！'),
                                delay: 8000
                            })
                            return false
                        }

                        if (path.path && !pathReg.test(path.path)) {
                            megPrefix += this.$t('{host}中路径组：', { host: rule.host })
                            this.$bkMessage({
                                theme: 'error',
                                message: megPrefix + this.$t('路径不正确'),
                                delay: 8000
                            })
                            return false
                        }

                        if (!path.backend.serviceName) {
                            megPrefix += this.$t('{host}中路径组：', { host: rule.host })
                            this.$bkMessage({
                                theme: 'error',
                                message: megPrefix + this.$t('请关联服务！'),
                                delay: 8000
                            })
                            return false
                        }

                        if (!path.backend.servicePort) {
                            megPrefix += this.$t('{host}中路径组：', { host: rule.host })
                            this.$bkMessage({
                                theme: 'error',
                                message: megPrefix + this.$t('请关联服务端口！'),
                                delay: 8000
                            })
                            return false
                        }

                        if (path.backend.serviceName && !this.linkServices.hasOwnProperty(path.backend.serviceName)) {
                            megPrefix += this.$t('{host}中路径组：', { host: rule.host })
                            this.$bkMessage({
                                theme: 'error',
                                message: megPrefix + this.$t('关联的Service【{serviceName}】不存在，请重新绑定', { serviceName: path.backend.serviceName }),
                                delay: 8000
                            })
                            return false
                        }
                    }
                }
                return true
            },

            formatData () {
                const params = JSON.parse(JSON.stringify(this.curEditedIngress))
                delete params.config.metadata.resourceVersion
                delete params.config.metadata.selfLink
                delete params.config.metadata.uid

                params.config.metadata.annotations = this.$refs.remarkKeyer.getKeyObject()
                params.config.metadata.labels = this.$refs.labelKeyer.getKeyObject()

                // 如果不是变量，转为数组形式
                const varReg = /\{\{([^\{\}]+)?\}\}/g
                params.config.spec.tls.forEach(item => {
                    if (!varReg.test(item.hosts)) {
                        item.hosts = item.hosts.split(',')
                    }
                })
                // 设置当前rules
                params.config.spec.rules = [JSON.parse(JSON.stringify(this.curRule))]
                return params
            },

            /**
             * 保存service
             */
            async saveIngressDetail () {
                if (this.checkData()) {
                    const data = this.formatData()
                    const projectId = this.projectId
                    const clusterId = this.curEditedIngress.cluster_id
                    const namespace = this.curEditedIngress.namespace
                    const ingressId = this.curEditedIngress.config.metadata.name

                    if (this.isDetailSaving) {
                        return false
                    }

                    this.isDetailSaving = true

                    try {
                        await this.$store.dispatch('resource/saveIngressDetail', {
                            projectId,
                            clusterId,
                            namespace,
                            ingressId,
                            data
                        })

                        this.$bkMessage({
                            theme: 'success',
                            message: this.$t('保存成功'),
                            hasCloseIcon: true,
                            delay: 3000
                        })
                        this.getIngressList()
                        this.handleCancelUpdate()
                    } catch (e) {
                        catchErrorHandler(e, this)
                    } finally {
                        this.isDetailSaving = false
                    }
                }
            },

            /**
             * 单选
             * @param {array} selection 已经选中的行数
             * @param {object} row 当前选中的行
             */
            handlePageSelect (selection, row) {
                this.ingressSelectedList = selection
            },

            /**
             * 全选
             */
            handlePageSelectAll (selection, row) {
                this.ingressSelectedList = selection
            },

            handleCancelUpdate () {
                this.ingressEditSlider.isShow = false
            },

            handlerSelectCert (computer, index, data) {
                computer.certType = data.certType
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
    @import '../../ingress.css';
</style>
