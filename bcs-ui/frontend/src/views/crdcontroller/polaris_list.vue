<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-crd-instance-title">
                <a href="javascript:void(0);" class="bcs-icon bcs-icon-arrows-left back" @click="goBack"></a>
                Polaris
                <span class="biz-tip ml10">({{$t('集群名称')}}：{{ clusterName }})</span>
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
                        <bk-button type="primary" @click.stop.prevent="createPolarisRules">
                            <i class="bcs-icon bcs-icon-plus" style="top: -1px;"></i>
                            <span>{{$t('新建规则')}}</span>
                        </bk-button>
                    </div>
                    <div class="right search-wrapper">
                        <div class="left">
                            <bk-selector
                                style="width: 180px;"
                                :searchable="true"
                                :placeholder="$t('Polaris命名空间')"
                                :selected.sync="searchParams.polaris_ns"
                                :list="polarisNameSpaceList"
                                :setting-key="'name'"
                                :display-key="'name'"
                                :allow-clear="true">
                            </bk-selector>
                            <bkbcs-input
                                style="width: 180px;"
                                :placeholder="$t('Polaris服务名')"
                                :value.sync="searchParams.polaris_name">
                            </bkbcs-input>
                        </div>
                        <div class="left">
                            <bk-button type="primary" :title="$t('查询')" icon="search" @click="handleSearch">
                                {{$t('查询')}}
                            </bk-button>
                        </div>
                    </div>
                </div>
                <div class="biz-crd-instance">
                    <div class="biz-table-wrapper">
                        <bk-table
                            class="biz-namespace-table"
                            v-bkloading="{ isLoading: isPageLoading && !isInitLoading }"
                            :size="'medium'"
                            :data="curPageData"
                            :pagination="pageConf"
                            @page-change="handlePageChange"
                            @page-limit-change="handlePageSizeChange">
                            <bk-table-column :label="$t('规则名')" prop="name" :show-overflow-tooltip="true" min-width="120">
                                <template slot-scope="{ row }">
                                    <p class="polaris-cell-item">{{ row.name }}</p>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('集群/命名空间')" min-width="220">
                                <template slot-scope="{ row }">
                                    <div>
                                        <p class="polaris-cell-item" style="padding-bottom: 5px;" :title="clusterName">{{ $t('所属集群') }}：{{ clusterName }}</p>
                                        <p class="polaris-cell-item" :title="row.namespace">{{ $t('命名空间') }}：{{ row.namespace }}</p>
                                    </div>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('Polaris信息')" min-width="180" :show-overflow-tooltip="true">
                                <template slot-scope="{ row }">
                                    <div>
                                        <p class="polaris-cell-item" style="padding-bottom: 5px;">
                                            <span class="label">{{ $t('名称') }}：</span>
                                            <span>{{ row.crd_data.polaris.name }}</span>
                                        </p>
                                        <p class="polaris-cell-item">
                                            <span class="label">{{ $t('命名空间') }}：</span>
                                            <span>{{ row.crd_data.polaris.namespace }}</span>
                                        </p>
                                    </div>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('关联Service')" min-width="160" :show-overflow-tooltip="true">
                                <template slot-scope="{ row }">
                                    <div v-for="(service, index) in row.crd_data.services" :key="index" style="padding: 5px 0;">
                                        <span style="display: inline-block; position: relative; bottom: 37px;">-</span>
                                        <span style="display: inline-block;">
                                            <p class="polaris-cell-item">name: {{ service.name }}</p>
                                            <p class="polaris-cell-item">port: <span style="padding-left: 8px;">{{ service.port }}</span></p>
                                            <p class="polaris-cell-item">direct: {{ service.direct === 'true' ? `${$t('是')}` : `${$t('否')}` }}</p>
                                        </span>
                                    </div>
                                </template>
                            </bk-table-column>
                            <bk-table-column label="ip: port weight" min-width="180" :show-overflow-tooltip="true">
                                <template slot-scope="{ row }">
                                    <div v-if="row.status && row.status.syncStatus && row.status.syncStatus.lastRemoteInstances">
                                        <div v-for="(remote, remoteIndex) in row.status.syncStatus.lastRemoteInstances" :key="remoteIndex" style="padding: 5px 0;">
                                            <p class="polaris-cell-item">{{ remote.ip || '--' }}: {{ remote.port || '--' }} {{ remote.weight || '--' }}</p>
                                        </div>
                                    </div>
                                    <div v-else>--</div>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('状态')" width="270">
                                <template slot-scope="{ row }">
                                    <div v-if="row.status && row.status.syncStatus" style="padding: 5px 0;">
                                        <p class="polaris-cell-item" style="padding-bottom: 5px;" :title="row.status.syncStatus.state">{{ $t('同步状态') }}：{{ row.status.syncStatus.state || '--' }}</p>
                                        <p class="polaris-cell-item" style="padding-bottom: 5px;" :title="row.status.syncStatus.lastSyncLatencyomitempty">{{ $t('同步耗时') }}：{{ row.status.syncStatus.lastSyncLatencyomitempty || '--' }}</p>
                                        <p class="polaris-cell-item" :title="row.status.syncStatus.lastSyncTime">{{ $t('最后同步时间') }}：{{ row.status.syncStatus.lastSyncTime || '--' }}</p>
                                    </div>
                                    <div v-else>--</div>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('操作记录')" min-width="240">
                                <template slot-scope="{ row }">
                                    <p class="polaris-cell-item" style="padding-bottom: 5px;">{{ $t('更新人') }}：<span style="padding-left: 14px;">{{ row.operator || '--' }}</span></p>
                                    <p class="polaris-cell-item" :title="row.updated">{{ $t('更新时间') }}：{{row.updated || '--'}}</p>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('操作')" min-width="140">
                                <template slot-scope="{ row }">
                                    <a href="javascript:void(0);" class="bk-text-button" @click="editCrdInstance(row)">{{$t('更新')}}</a>
                                    <a href="javascript:void(0);" class="bk-text-button" @click="removeCrdInstance(row)">{{$t('删除')}}</a>
                                </template>
                            </bk-table-column>
                        </bk-table>
                    </div>
                </div>
            </template>
        </div>
        <!-- 新建/更新 -->
        <bk-sideslider
            :quick-close="false"
            :is-show.sync="crdInstanceSlider.isShow"
            :title="crdInstanceSlider.title"
            :width="800">
            <div class="p30" slot="content">
                <div class="bk-form bk-form-vertical">
                    <div class="bk-form-item">
                        <div class="bk-form-content">
                            <div class="bk-form-inline-item is-required" style="width: 320px;">
                                <label class="bk-label">{{$t('规则名')}}：</label>
                                <div class="bk-form-content">
                                    <bkbcs-input
                                        :placeholder="$t('请输入')"
                                        :disabled="isReadonly"
                                        :value.sync="curCrdInstance.name">
                                    </bkbcs-input>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="bk-form-item">
                        <div class="bk-form-content">
                            <div class="bk-form-inline-item is-required" style="width: 320px;">
                                <label class="bk-label">{{$t('所属集群')}}：</label>
                                <div class="bk-form-content">
                                    <bkbcs-input
                                        :value.sync="clusterName"
                                        :disabled="true">
                                    </bkbcs-input>
                                </div>
                            </div>

                            <div class="bk-form-inline-item is-required" style="width: 320px; margin-left: 35px;">
                                <label class="bk-label">{{$t('命名空间')}}：</label>
                                <div class="bk-form-content">
                                    <bk-selector
                                        :searchable="true"
                                        :placeholder="$t('请选择')"
                                        :disabled="isReadonly"
                                        :selected.sync="curCrdInstance.namespace"
                                        :list="nameSpaceList"
                                        @item-selected="handleNamespaceSelect">
                                    </bk-selector>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="bk-form-item">
                        <label class="bk-label">{{$t('Polaris信息')}}：</label>
                    </div>
                    <div class="bk-form-item">
                        <div class="bk-form-content">
                            <div class="polaris-wrapper polaris-info">
                                <div class="bk-form-inline-item is-required" style="width: 299px;">
                                    <label class="bk-label">{{$t('名称')}}：</label>
                                    <div class="bk-form-content">
                                        <bkbcs-input
                                            :value.sync="curCrdInstance.polaris.name"
                                            :disabled="isReadonly"
                                            :placeholder="$t('允许数字、英文字母、.、-、_, 限制128个字符')">
                                        </bkbcs-input>
                                    </div>
                                </div>
                                <div class="bk-form-inline-item is-required" style="width: 320px; margin-left: 35px;">
                                    <label class="bk-label">{{$t('命名空间')}}：</label>
                                    <div class="bk-form-content">
                                        <bk-selector
                                            :searchable="true"
                                            :placeholder="$t('请选择')"
                                            :selected.sync="curCrdInstance.polaris.namespace"
                                            :disabled="isReadonly"
                                            :list="polarisNameSpaceList"
                                            @item-selected="handlePolarisNamespaceSelect">
                                        </bk-selector>
                                    </div>
                                </div>
                                <div class="bk-form-inline-item" style="width: 299px; margin-top: 30px; margin-right: 35px;">
                                    <bk-checkbox v-model="isTokenExist" :disabled="isReadonly" name="cluster-classify-checkbox">
                                        {{$t('Polaris服务是否已存在')}}
                                    </bk-checkbox>
                                </div>
                                <div v-if="isTokenExist" class="bk-form-inline-item" style="width: 350px; margin-top: 10px; height: 64px;">
                                    <label class="bk-label">token：</label>
                                    <div class="bk-form-content token">
                                        <bkbcs-input
                                            class="basic-input"
                                            :placeholder="$t('请输入')"
                                            :disabled="isReadonly"
                                            :value.sync="curCrdInstance.polaris.token">
                                        </bkbcs-input>
                                        <i class="bcs-icon bcs-icon-question-circle token-icon ml10" v-bk-tooltips.top="$t('如服务已存在则必填，若不存在平台会自动申请服务并创建')" />
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="bk-form-item">
                        <label class="bk-label">{{$t('关联Service')}}：</label>
                    </div>
                    <div class="bk-form-item">
                        <div class="bk-form-content">
                            <div class="polaris-wrapper">
                                <section class="polaris-inner-wrapper mb10" v-for="(service, index) in curCrdInstance.services" :key="index">
                                    <div class="bk-form-inline-item is-required" style="width: 284px;">
                                        <label class="bk-label">{{$t('服务名')}}：</label>
                                        <div class="bk-form-content">
                                            <bkbcs-input
                                                :placeholder="$t('请输入')"
                                                :value.sync="service.name">
                                            </bkbcs-input>
                                        </div>
                                    </div>
                                    <div class="bk-form-inline-item is-required" style="width: 319px; margin-left: 35px;">
                                        <label class="bk-label">{{$t('端口')}}：</label>
                                        <div class="bk-form-content">
                                            <bkbcs-input
                                                :placeholder="$t('请输入')"
                                                :value.sync="service.port">
                                            </bkbcs-input>
                                        </div>
                                    </div>
                                    <div class="bk-form-inline-item is-required" style="width: 284px; margin-top: 30px;">
                                        <div class="bk-form-content">
                                            <bk-checkbox v-model="service.direct" name="cluster-classify-checkbox">
                                                {{$t('是否直连Pod')}}
                                            </bk-checkbox>
                                            <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.top="$t('NodePort模式不勾选')" />
                                        </div>
                                    </div>
                                    <div class="bk-form-inline-item is-required" style="width: 319px; margin-top: 10px; margin-left: 35px;">
                                        <label class="bk-label">{{$t('权重')}}：</label>
                                        <div class="bk-form-content">
                                            <bkbcs-input
                                                :placeholder="$t('请输入')"
                                                :value.sync="service.weight">
                                            </bkbcs-input>
                                        </div>
                                    </div>

                                    <i class="bcs-icon bcs-icon-close polaris-close" @click="removeServiceMap(index)" v-if="curCrdInstance.services.length > 1"></i>
                                </section>

                                <bk-button class="polaris-block-btn mt10" @click="addServiceMap">
                                    <i class="bcs-icon bcs-icon-plus"></i>
                                    {{$t('点击增加')}}
                                </bk-button>
                            </div>
                        </div>
                    </div>

                    <div class="bk-form-item mt25">
                        <bk-button type="primary" :loading="isDataSaveing" @click.stop.prevent="saveCrdInstance">{{curCrdInstance.crd_id ? $t('更新') : $t('创建')}}</bk-button>
                        <bk-button @click.stop.prevent="hideCrdInstanceSlider" :disabled="isDataSaveing">{{$t('取消')}}</bk-button>
                    </div>
                </div>
            </div>
        </bk-sideslider>
    </div>
</template>

<script>
    import { defineComponent, reactive, toRefs, computed, onMounted, watch } from '@vue/composition-api'
    export default defineComponent({
        name: 'crdcontrollerPolarisInstances',
        components: {
        },
        setup (props, ctx) {
            const { $store, $route, $router, $i18n } = ctx.root
            const crdInstanceList = computed(() => Object.assign([], $store.state.crdcontroller.crdInstanceList))
            const clusterList = computed(() => $store.state.cluster.clusterList)
            const curProject = computed(() => $store.state.curProject).value
            const projectId = computed(() => $route.params.projectId).value
            const clusterId = computed(() => $route.params.clusterId).value
            const clusterName = computed(() => {
                const cluster = clusterList.value.find(item => {
                    return item.cluster_id === clusterId
                })
                return cluster ? cluster.name : ''
            })

            const state = reactive({
                crdKind: 'PolarisConfig',
                isInitLoading: true,
                isDataSaveing: false,
                isPageLoading: false,
                exceptionCode: null,
                isTokenExist: false,
                isReadonly: false,
                pageConf: {
                    count: 0,
                    totalPage: 1,
                    limit: 5,
                    current: 1,
                    show: true
                },
                nameSpaceList: [],
                polarisNameSpaceList: [
                    {
                        id: 'Production',
                        name: 'Production'
                    },
                    {
                        id: 'Pre-release',
                        name: 'Pre-release'
                    },
                    {
                        id: 'Test',
                        name: 'Test'
                    },
                    {
                        id: 'Development',
                        name: 'Development'
                    }
                ],
                curPageData: [],
                crdInstanceSlider: {
                    title: $i18n.t('新建'),
                    isShow: false
                },
                curCrdInstance: {
                    'name': '',
                    'namespace': '',
                    'polaris': {
                        'name': '',
                        'namespace': '',
                        'token': ''
                    },
                    'services': [
                        {
                            'name': '',
                            'port': '',
                            'direct': true,
                            'weight': ''
                        }
                    ]
                },
                searchParams: {
                    polaris_name: '',
                    polaris_ns: ''
                },
                appTypes: [
                    {
                        id: 'polaris_name',
                        name: $i18n.t('Polaris服务名')
                    }
                ]
            })

            watch(crdInstanceList, async () => {
                state.curPageData = await getDataByPage(state.pageConf.current)
            })

            onMounted(() => {
                getCrdInstanceList()
                getNameSpaceList()
            })

            const goBack = () => {
                $router.push({
                    name: 'dbCrdcontroller',
                    params: {
                        projectId: projectId
                    }
                })
            }

            /**
             * 新建规则
             */
            const createPolarisRules = () => {
                state.curCrdInstance = {
                    'name': '',
                    'namespace': '',
                    'polaris': {
                        'name': '',
                        'namespace': '',
                        'token': ''
                    },
                    'services': [{
                        'name': '',
                        'port': '',
                        'direct': true,
                        'weight': ''
                    }]
                }
                state.crdInstanceSlider.title = $i18n.t('新建')
                state.isTokenExist = false
                state.isReadonly = false
                state.crdInstanceSlider.isShow = true
            }

            /**
             * 搜索列表
             */
            const handleSearch = () => {
                state.pageConf.current = 1
                state.isPageLoading = true
                getCrdInstanceList()
            }

            /**
             * 加载数据
             */
            const getCrdInstanceList = async () => {
                const crdKind = state.crdKind
                const params = {}

                if (state.searchParams.polaris_name) {
                    params.polaris_name = state.searchParams.polaris_name
                }
                if (state.searchParams.polaris_ns) {
                    params.polaris_ns = state.searchParams.polaris_ns
                }

                const res = await $store.dispatch('crdcontroller/getCrdInstanceList', {
                    projectId,
                    clusterId,
                    crdKind,
                    params
                })
                // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
                setTimeout(() => {
                    state.isPageLoading = false
                    state.isInitLoading = false
                }, 200)

                if (!res) return
                initPageConf()
                state.curPageData = getDataByPage(state.pageConf.current)
            }

            /**
             * 初始化分页配置
             */
            const initPageConf = () => {
                const total = crdInstanceList.value.length
                state.pageConf.count = total
                state.pageConf.totalPage = Math.ceil(total / state.pageConf.limit)
                if (state.pageConf.current > state.pageConf.totalPage) {
                    state.pageConf.current = state.pageConf.totalPage
                }
            }

            /**
             * 获取页数据
             * @param  {number} page 页
             * @return {object} data lb
             */
            const getDataByPage = (page) => {
                // 如果没有page，重置
                if (!page) {
                    state.pageConf.current = page = 1
                }
                let startIndex = (page - 1) * state.pageConf.limit
                let endIndex = page * state.pageConf.limit
                // state.isPageLoading = true
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > crdInstanceList.value.length) {
                    endIndex = crdInstanceList.value.length
                }
                state.isPageLoading = false
                return crdInstanceList.value.slice(startIndex, endIndex)
            }

            /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
            const handlePageSizeChange = (pageSize) => {
                state.pageConf.limit = pageSize
                state.pageConf.current = 1
                initPageConf()
                handlePageChange()
            }

            /**
             * 分页改变回调
             * @param  {number} page 页
             */
            const handlePageChange = (page = 1) => {
                state.isPageLoading = true
                state.pageConf.current = page
                const data = getDataByPage(page)
                state.curPageData = JSON.parse(JSON.stringify(data))
            }

            /**
             * 获取命名空间列表
             */
            const getNameSpaceList = async () => {
                const res = await $store.dispatch('crdcontroller/getNameSpaceListByCluster', { projectId, clusterId }).catch(() => false)
                if (!res) return
                const list = res.data
                list.forEach(item => {
                    item.isSelected = false
                    item.id = item.name
                })
                state.nameSpaceList.splice(0, state.nameSpaceList.length, ...list)
            }

            /**
             * 增加一个关联Service
             */
            const addServiceMap = () => {
                const params = {
                    'name': '',
                    'port': '',
                    'direct': true,
                    'weight': ''
                }
                state.curCrdInstance.services.push(params)
            }

            /**
             * 移除一个关联Service
             */
            const removeServiceMap = (index) => {
                state.curCrdInstance.services.splice(index, 1)
            }

            /**
             * 隐藏侧面板
             */
            const hideCrdInstanceSlider = () => {
                state.crdInstanceSlider.isShow = false
            }

            /**
             * 保存新建/更新
             */
            const actionCrdInstance = async (params) => {
                let url = ''
                if (state.curCrdInstance.id > 0) {
                    url = 'crdcontroller/updateCrdInstance'
                    state.isReadonly = true
                } else {
                    url = 'crdcontroller/addCrdInstance'
                    state.isReadonly = false
                }

                const data = JSON.parse(JSON.stringify(params))
                data.services.forEach(item => {
                    item.direct = String(item.direct)
                })

                const crdKind = state.crdKind
                state.isDataSaveing = true

                const result = await $store.dispatch(url, { projectId, clusterId, crdKind, data }).catch(() => false)
                state.isDataSaveing = false

                if (!result) return

                ctx.root.$bkMessage({
                    theme: 'success',
                    message: $i18n.t('数据保存成功')
                })
                getCrdInstanceList()
                hideCrdInstanceSlider()
            }

            /**
             * 保存 / 更新
             */
            const saveCrdInstance = () => {
                const params = {
                    ...state.curCrdInstance
                }
                if (checkData() && !state.isDataSaveing) {
                    actionCrdInstance(params)
                }
            }

            /**
             * 编辑
             * @param  {object} crdInstance crdInstance
             * @param  {number} index 索引
             */
            const editCrdInstance = async (crdInstance) => {
                const crdKind = state.crdKind
                const crdId = crdInstance.id
                const res = await $store.dispatch('crdcontroller/getCrdInstanceDetail', {
                    crdKind,
                    projectId,
                    clusterId,
                    crdId
                }).catch(() => false)
                state.crdInstanceSlider.isShow = true
                state.isReadonly = true
                if (!res) return
                res.data.crd_data.services.forEach(item => {
                    item.direct === 'true' ? item.direct = true : item.direct = false
                })
                state.curCrdInstance = res.data.crd_data
                state.curCrdInstance.crd_id = crdId
                state.crdInstanceSlider.title = $i18n.t('编辑')
            }

            /**
             * 删除
             * @param  {object} crdInstance crdInstance
             * @param  {number} index 索引
             */
            const removeCrdInstance = async (crdInstance, index) => {
                const crdKind = state.crdKind
                const crdId = crdInstance.id

                ctx.root.$bkInfo({
                    title: $i18n.t('确认删除'),
                    clsName: 'biz-remove-dialog',
                    content: ctx.root.$createElement('p', {
                        class: 'biz-confirm-desc'
                    }, `${$i18n.t('确定要删除')}【${crdInstance.name}】？`),
                    async confirmFn () {
                        state.isPageLoading = true
                        const res = await $store.dispatch('crdcontroller/deleteCrdInstance', { projectId, clusterId, crdKind, crdId }).catch(() => false)
                        state.isPageLoading = false

                        if (!res) return

                        ctx.root.$bkMessage({
                            theme: 'success',
                            message: $i18n.t('删除成功')
                        })
                        getCrdInstanceList()
                    }
                })
            }

            /**
             * 检查提交的数据
             * @return {boolean} true/false 是否合法
             */
            const checkData = () => {
                if (!state.curCrdInstance.name) {
                    ctx.root.$bkMessage({
                        theme: 'error',
                        message: $i18n.t('请输入规则名'),
                        delay: 5000
                    })
                    return false
                }

                if (state.curCrdInstance.namespace === '') {
                    ctx.root.$bkMessage({
                        theme: 'error',
                        message: $i18n.t('请选择命名空间')
                    })
                    return false
                }

                if (!state.curCrdInstance.polaris.name) {
                    ctx.root.$bkMessage({
                        theme: 'error',
                        message: $i18n.t('请输入Polaris信息名'),
                        delay: 5000
                    })
                    return false
                }

                if (state.curCrdInstance.polaris.name && !/^[\w-.:]{1,128}$/.test(state.curCrdInstance.polaris.name)) {
                    ctx.root.$bkMessage({
                        theme: 'error',
                        message: $i18n.t('Polaris信息名只允许数字、英文字母、.、-、_, 限制128个字符'),
                        delay: 5000
                    })
                    return false
                }

                if (!state.curCrdInstance.polaris.namespace) {
                    ctx.root.$bkMessage({
                        theme: 'error',
                        message: $i18n.t('请选择Polaris命名空间'),
                        delay: 5000
                    })
                    return false
                }

                if (state.curCrdInstance.services.length) {
                    let status = true
                    state.curCrdInstance.services.forEach(i => {
                        if (!i.name) {
                            ctx.root.$bkMessage({
                                theme: 'error',
                                message: $i18n.t('请输入关联Sercice服务名'),
                                delay: 5000
                            })
                            status = false
                        }
                        if (!i.port) {
                            ctx.root.$bkMessage({
                                theme: 'error',
                                message: $i18n.t('请输入关联Sercice的端口(整数类型)'),
                                delay: 5000
                            })
                            status = false
                        }
                        if (i.port && i.port < 0) {
                            ctx.root.$bkMessage({
                                theme: 'error',
                                message: $i18n.t('关联Sercice的端口不能为负数'),
                                delay: 5000
                            })
                            status = false
                        }
                        if (!i.weight) {
                            ctx.root.$bkMessage({
                                theme: 'error',
                                message: $i18n.t('请输入关联Sercice的权重(整数类型)'),
                                delay: 5000
                            })
                            status = false
                        }
                        if (i.port && i.weight < 0) {
                            ctx.root.$bkMessage({
                                theme: 'error',
                                message: $i18n.t('关联Sercice的权重不能为负数'),
                                delay: 5000
                            })
                            status = false
                        }
                    })
                    return status
                }
                return true
            }

            return {
                ...toRefs(state),
                projectId,
                crdInstanceList,
                curProject,
                clusterId,
                clusterName,
                goBack,
                handleSearch,
                handlePageChange,
                handlePageSizeChange,
                createPolarisRules,
                hideCrdInstanceSlider,
                editCrdInstance,
                removeCrdInstance,
                saveCrdInstance,
                addServiceMap,
                removeServiceMap
            }
        }
    })
</script>

<style scoped>
    @import './polaris_list.css';
</style>
