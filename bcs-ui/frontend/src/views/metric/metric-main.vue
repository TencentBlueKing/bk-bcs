<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-app-title">
                <!-- Metric管理{{$t('test', { vari1: 1, vari2: 2 })}} -->
                {{$t('Metric管理')}}
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
                <div class="biz-panel-header biz-metric-manage-create" style="padding: 27px 30px 22px 20px;">
                    <div class="left">
                        <bk-button type="primary" :title="$t('新建Metric')" @click="showCreateMetric">
                            <i class="bcs-icon bcs-icon-plus"></i>
                            <span class="text">{{$t('新建Metric')}}</span>
                        </bk-button>
                    </div>
                    <div class="right">
                        <bk-data-searcher
                            :placeholder="$t('输入名称，按Enter搜索')"
                            :search-key.sync="searchKeyWord"
                            @search="searchMetric"
                            @refresh="refresh">
                        </bk-data-searcher>
                    </div>
                </div>
                <div class="biz-table-wrapper">
                    <bk-table
                        :size="'medium'"
                        :data="curPageData"
                        :pagination="pageConf"
                        v-bkloading="{ isLoading: isPageLoading && !isInitLoading }"
                        @page-limit-change="handlePageSizeChange"
                        @page-change="handlePageChange">
                        <bk-table-column :label="$t('名称')" :show-overflow-tooltip="true" min-width="160">
                            <template slot-scope="props">
                                {{props.row.name || '--'}}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('端口')" :show-overflow-tooltip="true" min-width="100">
                            <template slot-scope="props">
                                {{props.row.port || '--'}}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('URI')" :show-overflow-tooltip="true" min-width="250">
                            <template slot-scope="props">
                                {{props.row.uri || '--'}}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('采集频率(秒/次)')" :show-overflow-tooltip="true" min-width="130">
                            <template slot-scope="props">
                                {{props.row.frequency || '--'}}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('操作')" :show-overflow-tooltip="true" width="330">
                            <template slot-scope="props">
                                <a href="javascript:void(0);" class="bk-text-button" @click="checkMetricInstance(props.row)">{{$t('查看实例')}}</a>
                                <template v-if="!props.row.status || props.row.status === 'normal'">
                                    <a href="javascript:void(0);" class="bk-text-button" @click="pauseAndResume(props.row, 'pause', [])">{{$t('暂停')}}</a>
                                    <a href="javascript:void(0);" class="bk-text-button" @click="editMetric(props.row)">{{$t('更新')}}</a>
                                </template>
                                <template v-else>
                                    <a href="javascript:void(0);" class="bk-text-button" @click="pauseAndResume(props.row, 'resume', [])">{{$t('恢复')}}</a>
                                </template>
                                <a href="javascript:void(0);" class="bk-text-button" @click="deleteMetric(props.row)">{{$t('删除')}}</a>
                                <!-- 数据平台不能直接跳转到字段设置页面，先去掉 -->
                                <!-- <a class="bk-text-button" href="javascript:void(0)" @click="go(props.row, props.row.uri_fields_info)" target="_blank">字段设置</a> -->
                                <a class="bk-text-button" href="javascript:void(0)" @click="go(props.row, props.row.uri_data_clean)" target="_blank">{{$t('数据清洗')}}</a>
                            </template>
                        </bk-table-column>
                    </bk-table>
                </div>
            </template>
        </div>

        <bk-sideslider
            :is-show.sync="createMetricConf.isShow"
            :title="createMetricConf.title"
            :width="createMetricConf.width"
            :quick-close="false"
            class="biz-metric-manage-create-sideslider"
            @hidden="hideCreateMetric">
            <div slot="content">
                <div class="wrapper" style="position: relative;">
                    <form class="bk-form bk-form-vertical create-form">
                        <div class="bk-form-item flex-item">
                            <div class="left">
                                <label class="bk-label label">{{$t('名称')}}：<span class="red">*</span></label>
                                <div class="bk-form-content">
                                    <bk-input style="width: 270px;" v-model="createParams.name" :placeholder="$t('请输入')" maxlength="253" />
                                </div>
                            </div>
                            <div class="right">
                                <label class="bk-label label">{{$t('端口')}}：<span class="red">*</span></label>
                                <div class="bk-form-content">
                                    <bk-number-input
                                        class="text-input-half"
                                        :value.sync="createParams.port"
                                        :min="1"
                                        :max="65535"
                                        :debounce-timer="0"
                                        :placeholder="$t('请输入')">
                                    </bk-number-input>
                                </div>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label label">
                                URI：<span class="red">*</span>
                            </label>
                            <div class="bk-form-content">
                                <bk-input v-model="createParams.url" :placeholder="$t('请输入')" />
                            </div>
                        </div>
                        <div class="bk-form-item flex-item">
                            <div class="left">
                                <label class="bk-label label">{{$t('采集频率')}}：<span class="red">*</span></label>
                                <div class="bk-form-content">
                                    <bk-number-input
                                        class="text-input-half has-suffix"
                                        :value.sync="createParams.frequency"
                                        :min="0"
                                        :max="999999999"
                                        :debounce-timer="0"
                                        :placeholder="$t('请输入')">
                                    </bk-number-input>
                                    <span class="suffix">
                                        {{$t('秒/次')}}
                                    </span>
                                </div>
                            </div>
                            <div class="right">
                                <label class="bk-label label">{{$t('单次采集超时时间（秒）')}}：<span class="red">*</span></label>
                                <div class="bk-form-content">
                                    <bk-number-input
                                        class="text-input-half has-suffix"
                                        :value.sync="createParams.timeout"
                                        :min="0"
                                        :debounce-timer="0"
                                        :placeholder="$t('请输入')">
                                    </bk-number-input>
                                    <span class="suffix">
                                        {{$t('秒')}}
                                    </span>
                                </div>
                            </div>
                        </div>
                        <div class="bk-form-item prometheus-item">
                            <div class="prometheus-header">
                                <label class="bk-label label" style="width: 300px;">{{$t('Prometheus格式设置')}}</label>
                                <bk-checkbox class="mt5" name="metric-type" v-model="createParams.metricType"></bk-checkbox>
                            </div>

                            <div class="prometheus-keys" v-show="createParams.metricType">
                                <label class="bk-label label">
                                    {{$t('附加数据')}}：
                                </label>
                                <bk-keyer
                                    class="prometheus-keylist"
                                    ref="constKeyer"
                                    :key-placeholder="$t('键')"
                                    :value-placeholder="$t('值')"
                                    :tip="$t('小提示：同时粘贴多行“键=值”的文本会自动添加多行记录')"
                                    :key-list.sync="createParams.constLabels"
                                ></bk-keyer>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label label">
                                Http Header：
                            </label>
                            <div class="bk-form-content">
                                <bk-keyer
                                    class="http-header"
                                    ref="labelKeyer"
                                    :key-list.sync="createParams.httpHeader"
                                    :key-placeholder="$t('键')"
                                    :value-placeholder="$t('值')"
                                    :tip="$t('小提示：同时粘贴多行“键=值”的文本会自动添加多行记录')"
                                ></bk-keyer>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label label">Http Method：</label>
                            <div class="bk-form-content scroll-order-form-item">
                                <bk-radio-group v-model="createParams.httpMethod">
                                    <bk-radio value="GET">GET</bk-radio>
                                    <bk-radio value="POST">POST</bk-radio>
                                </bk-radio-group>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label label">
                                {{$t('Http参数')}}：
                            </label>
                            <div class="bk-form-content">
                                <template v-if="createParams.httpMethod === 'GET'">
                                    <bk-keyer
                                        class="http-header"
                                        ref="labelKeyer"
                                        :key-list.sync="createParams.httpBodyGet"
                                        :key-placeholder="$t('键')"
                                        :value-placeholder="$t('值')"
                                        :tip="$t('小提示：同时粘贴多行“键=值”的文本会自动添加多行记录')"
                                    ></bk-keyer>
                                </template>
                                <template v-else>
                                    <textarea v-model="createParams.httpBodyPost" class="bk-form-textarea" :placeholder="$t('请输入')"></textarea>
                                </template>
                            </div>
                        </div>
                        <div class="action-inner">
                            <bk-button type="primary" :loading="isCreatingOrEditing" @click="confirmCreateMetric">
                                {{$t('创建')}}
                            </bk-button>
                            <bk-button type="button" :diasbled="isCreatingOrEditing" @click="hideCreateMetric">
                                {{$t('取消')}}
                            </bk-button>
                        </div>
                    </form>
                </div>
            </div>
        </bk-sideslider>

        <bk-sideslider
            :is-show.sync="editMetricConf.isShow"
            :title="editMetricConf.title"
            :width="editMetricConf.width"
            :quick-close="false"
            class="biz-metric-manage-create-sideslider"
            @hidden="hideEditMetric">
            <div slot="content">
                <div class="wrapper" style="position: relative;">
                    <form class="bk-form bk-form-vertical create-form">
                        <div class="bk-form-item flex-item">
                            <div class="left">
                                <label class="bk-label label">{{$t('名称')}}：<span class="red">*</span></label>
                                <div class="bk-form-content">
                                    <bk-input :disabled="true" v-model="editParams.name" style="width: 270px;" :placeholder="$t('请输入')" maxlength="32" />
                                </div>
                            </div>
                            <div class="right">
                                <label class="bk-label label">{{$t('端口')}}：<span class="red">*</span></label>
                                <div class="bk-form-content">
                                    <bk-number-input
                                        class="text-input-half"
                                        :value.sync="editParams.port"
                                        :min="1"
                                        :max="65535"
                                        :debounce-timer="0"
                                        :placeholder="$t('请输入')">
                                    </bk-number-input>
                                </div>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label label">
                                URI：<span class="red">*</span>
                            </label>
                            <div class="bk-form-content">
                                <bk-input v-model="editParams.url" :placeholder="$t('请输入')" />
                            </div>
                        </div>
                        <div class="bk-form-item flex-item">
                            <div class="left">
                                <label class="bk-label label">{{$t('采集频率')}}：<span class="red">*</span></label>
                                <div class="bk-form-content">
                                    <bk-number-input
                                        class="text-input-half has-suffix"
                                        :value.sync="editParams.frequency"
                                        :min="0"
                                        :max="999999999"
                                        :debounce-timer="0"
                                        :placeholder="$t('请输入')">
                                    </bk-number-input>
                                    <span class="suffix">
                                        {{$t('秒/次')}}
                                    </span>
                                </div>
                            </div>
                            <div class="right">
                                <label class="bk-label label">{{$t('单次采集超时时间（秒）')}}：<span class="red">*</span></label>
                                <div class="bk-form-content">
                                    <bk-number-input
                                        class="text-input-half has-suffix"
                                        :value.sync="editParams.timeout"
                                        :min="0"
                                        :debounce-timer="0"
                                        :placeholder="$t('请输入')">
                                    </bk-number-input>
                                    <span class="suffix">
                                        {{$t('秒')}}
                                    </span>
                                </div>
                            </div>
                        </div>

                        <div class="bk-form-item prometheus-item">
                            <div class="prometheus-header">
                                <label class="bk-label label" style="width: 300px;">{{$t('Prometheus格式设置')}}</label>
                                <bk-checkbox class="mt5" name="metric-type" v-model="editParams.metricType" :disabled="true" v-bk-tooltips.left="$t('在创建的时候已经按{metricType}类型在数据平台申请dataid，不能更改', { metricType: editParams.metricType ? 'Prometheus' : $t('普通') })"></bk-checkbox>
                            </div>

                            <div class="prometheus-keys" v-show="editParams.metricType">
                                <label class="bk-label label">
                                    {{$t('附加数据')}}：
                                </label>
                                <bk-keyer
                                    class="prometheus-keylist"
                                    ref="constKeyer"
                                    :key-list.sync="editParams.constLabels"
                                    :key-placeholder="$t('键')"
                                    :value-placeholder="$t('值')"
                                    :tip="$t('小提示：同时粘贴多行“键=值”的文本会自动添加多行记录')"
                                ></bk-keyer>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label label">
                                Http Header：
                            </label>
                            <div class="bk-form-content">
                                <bk-keyer
                                    class="http-header"
                                    ref="labelKeyer"
                                    :key-list.sync="editParams.httpHeader"
                                    :key-placeholder="$t('键')"
                                    :value-placeholder="$t('值')"
                                    :tip="$t('小提示：同时粘贴多行“键=值”的文本会自动添加多行记录')"
                                ></bk-keyer>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label label">Http Method：</label>
                            <div class="bk-form-content scroll-order-form-item">
                                <bk-radio-group v-model="editParams.httpMethod">
                                    <bk-radio value="GET">GET</bk-radio>
                                    <bk-radio value="POST">POST</bk-radio>
                                </bk-radio-group>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label label">
                                {{$t('Http参数')}}：
                            </label>
                            <div class="bk-form-content">
                                <template v-if="editParams.httpMethod === 'GET'">
                                    <bk-keyer
                                        class="http-header"
                                        ref="labelKeyer"
                                        :key-list.sync="editParams.httpBodyGet"
                                        :key-placeholder="$t('键')"
                                        :value-placeholder="$t('值')"
                                        :tip="$t('小提示：同时粘贴多行“键=值”的文本会自动添加多行记录')"
                                    ></bk-keyer>
                                </template>
                                <template v-else>
                                    <textarea v-model="editParams.httpBodyPost" class="bk-form-textarea" :placeholder="$t('请输入')"></textarea>
                                </template>
                            </div>
                        </div>
                        <div class="action-inner">
                            <bk-button type="primary" :loading="isCreatingOrEditing" @click="confirmEditMetric">
                                {{$t('更新')}}
                            </bk-button>
                            <bk-button type="button" :disabled="isCreatingOrEditing" @click="hideEditMetric">
                                {{$t('取消')}}
                            </bk-button>
                        </div>
                    </form>
                </div>
            </div>
        </bk-sideslider>

        <bk-dialog
            :is-show.sync="instanceDialogConf.isShow"
            :width="instanceDialogConf.width"
            :content="instanceDialogConf.content"
            :has-header="instanceDialogConf.hasHeader"
            :has-footer="false"
            :close-icon="true"
            @cancel="hideInstanceDialog"
            :ext-cls="'biz-metric-manage-dialog'">
            <template slot="content">
                <div style="margin: -20px;">
                    <div class="instance-title">
                        {{curInstanceMetric.name}}{{$t('实例')}}
                    </div>
                    <div style="min-height: 100px;" v-bkloading="{ isLoading: isMetricInstanceLoading }">
                        <table class="bk-table has-table-hover biz-table biz-metric-instance-table" :style="{ borderBottomWidth: curMetricInstancePageData.length ? '1px' : 0 }" v-show="!isMetricInstanceLoading">
                            <thead>
                                <tr>
                                    <th style="padding-left: 30px;">{{$t('关联命名空间')}}</th>
                                    <th>{{$t('关联应用')}}</th>
                                    <th style="width: 150px;">{{$t('应用类型')}}</th>
                                </tr>
                            </thead>
                            <tbody>
                                <template v-if="curMetricInstancePageData.length">
                                    <tr v-for="(instance, index) in curMetricInstancePageData" :key="index">
                                        <td style="padding-left: 30px;">
                                            {{instance.namespace}}
                                        </td>
                                        <td>
                                            {{instance.name}}
                                        </td>
                                        <td>
                                            {{instance.category}}
                                        </td>
                                    </tr>
                                </template>
                                <template v-else>
                                    <tr>
                                        <td colspan="3">
                                            <div class="bk-message-box no-data">
                                                <bcs-exception type="empty" scene="part"></bcs-exception>
                                            </div>
                                        </td>
                                    </tr>
                                </template>
                            </tbody>
                        </table>
                    </div>
                    <div class="biz-page-box">
                        <bk-pagination
                            :show-limit="false"
                            :current.sync="metricInstancePageConf.curPage"
                            :count.sync="metricInstancePageConf.total"
                            :limit="metricInstancePageConf.pageSize"
                            @change="metricInstancePageChange">
                        </bk-pagination>
                    </div>
                </div>
            </template>
        </bk-dialog>
    </div>
</template>

<script>
    import bkKeyer from '@/components/keyer'

    export default {
        components: {
            'bk-keyer': bkKeyer
        },
        data () {
            return {
                winHeight: 0,
                searchKeyWord: '',
                isInitLoading: true,
                isPageLoading: false,
                exceptionCode: null,
                dataList: [],
                dataListTmp: [],
                curPageData: [],
                metricInstancePageConf: {
                    totalPage: 1,
                    total: 1,
                    pageSize: 5,
                    curPage: 1,
                    show: true
                },
                isMetricInstanceLoading: true,
                curMetricInstancePageData: [],
                instanceDialogConf: {
                    isShow: false,
                    width: 690,
                    hasHeader: false,
                    closeIcon: false
                },
                pageConf: {
                    // 总数
                    count: 0,
                    // 总页数
                    totalPage: 1,
                    // 每页多少条
                    limit: 5,
                    // 当前页
                    current: 1,
                    // 是否显示翻页条
                    show: false
                },
                createMetricConf: {
                    isShow: false,
                    title: this.$t('新建Metric'),
                    timer: null,
                    width: 650,
                    loading: false
                },
                // 创建的参数
                createParams: {
                    name: '',
                    port: '',
                    url: '',
                    timeout: 30,
                    metricType: false,
                    frequency: 60,
                    httpMethod: 'GET',
                    httpHeader: [{ key: '', value: '' }],
                    httpBodyGet: [{ key: '', value: '' }],
                    constLabels: [{ key: '', value: '' }],
                    httpBodyPost: ''
                },
                // 编辑的参数
                editMetricConf: {
                    isShow: false,
                    title: this.$t('新建Metric'),
                    timer: null,
                    width: 650,
                    loading: false
                },
                curInstanceMetric: {
                    name: ''
                },
                // 编辑的参数
                editParams: {
                    curMetric: null,
                    name: '',
                    port: '',
                    url: '',
                    metricType: false,
                    timeout: 30,
                    frequency: 60,
                    httpMethod: 'GET',
                    httpHeader: [],
                    httpBodyGet: [],
                    constLabels: [],
                    httpBodyPost: ''
                },
                isCreatingOrEditing: false,
                creatingOrEditingStr: ''
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            isEn () {
                return this.$store.state.isEn
            }
        },
        mounted () {
            this.winHeight = window.innerHeight
            this.fetchData()
        },
        methods: {
            /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
            handlePageSizeChange (pageSize) {
                this.pageConf.limit = pageSize
                this.pageConf.current = 1
                this.initPageConf()
                this.handlePageChange()
            },

            /**
             * 搜索框清除事件
             */
            clearSearch () {
                this.searchKeyWord = ''
                this.searchMetric(true)
            },

            /**
             * 重置添加 metric 的参数
             */
            resetCreateParams () {
                this.createParams = Object.assign({}, {
                    name: '',
                    port: '',
                    url: '',
                    metricType: false,
                    timeout: 30,
                    frequency: 60,
                    httpMethod: 'GET',
                    httpHeader: [{ key: '', value: '' }],
                    httpBodyGet: [{ key: '', value: '' }],
                    constLabels: [{ key: '', value: '' }],
                    httpBodyPost: ''
                })
            },

            /**
             * 搜索
             *
             * @param {boolean} resetPage 是否重置 curPage 为 1
             * @param {Boolean} notLoading 是否不需要 loading
             */
            searchMetric (resetPage, notLoading = false) {
                let results = []
                if (this.searchKeyWord === '') {
                    this.dataList.splice(0, this.dataList.length, ...this.dataListTmp)
                } else {
                    results = this.dataListTmp.filter(m => {
                        return m.name.indexOf(this.searchKeyWord) > -1
                    })
                    this.dataList.splice(0, this.dataList.length, ...results)
                }
                if (resetPage) {
                    this.pageConf.current = 1
                }
                this.initPageConf()
                this.curPageData = this.getDataByPage(this.pageConf.current, notLoading)
            },

            /**
             * 获取 metric 列表数据
             */
            async fetchData () {
                try {
                    const res = await this.$store.dispatch('metric/getMetricList', {
                        projectId: this.projectId
                    })
                    this.dataList.splice(0, this.dataList.length, ...(res.data || []))
                    this.dataListTmp.splice(0, this.dataListTmp.length, ...(res.data || []))
                    this.initPageConf()
                    this.curPageData = this.getDataByPage(this.pageConf.current)
                } catch (e) {
                    console.error(e)
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
                const total = this.dataList.length
                if (total <= this.pageConf.limit) {
                    this.pageConf.show = false
                } else {
                    this.pageConf.show = true
                }
                this.pageConf.count = total
                this.pageConf.totalPage = Math.ceil(total / this.pageConf.limit) || 1
            },

            /**
             * 翻页回调
             *
             * @param {number} page 当前页
             */
            handlePageChange (page = 1) {
                this.pageConf.current = page
                const data = this.getDataByPage(page)
                this.curPageData.splice(0, this.curPageData.length, ...data)
            },

            /**
             * 获取当前这一页的数据
             *
             * @param {number} page 当前页
             * @param {Boolean} notLoading 是否不需要 loading
             *
             * @return {Array} 当前页数据
             */
            getDataByPage (page, notLoading = false) {
                // 如果没有page，重置
                if (!page) {
                    this.pageConf.current = page = 1
                }
                this.isPageLoading = !notLoading
                let startIndex = (page - 1) * this.pageConf.limit
                let endIndex = page * this.pageConf.limit
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.dataList.length) {
                    endIndex = this.dataList.length
                }
                setTimeout(() => {
                    this.isPageLoading = false
                }, 200)
                return this.dataList.slice(startIndex, endIndex)
            },

            /**
             * 手动刷新表格数据
             */
            refresh () {
                this.pageConf.current = 1
                this.searchKeyWord = ''
                this.fetchData()
            },

            /**
             * 创建 Metric 确定按钮
             */
            async confirmCreateMetric () {
                const me = this
                const name = me.createParams.name.trim()
                const port = me.createParams.port
                const url = me.createParams.url.trim()
                const timeout = me.createParams.timeout
                const frequency = me.createParams.frequency
                const httpMethod = me.createParams.httpMethod.trim()

                if (!name) {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入名称')
                    })
                    return
                }

                if (name.length < 3) {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('名称不得小于三个字符')
                    })
                    return
                }

                if (url.length < 2) {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('URI不得小于两个字符')
                    })
                    return
                }

                if (port === null || port === undefined || port === '') {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入端口')
                    })
                    return
                }

                if (!url) {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入URI')
                    })
                    return
                }

                if (frequency === null || frequency === undefined || frequency === '') {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入采集频率')
                    })
                    return
                }

                if (timeout === '' || timeout === null || timeout === undefined) {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入单次采集超时时间')
                    })
                    return
                }

                const params = {
                    projectId: me.projectId,
                    name: name,
                    port: port,
                    uri: url,
                    metric_type: '',
                    timeout: timeout,
                    frequency: frequency,
                    const_labels: {},
                    http_method: httpMethod
                }

                if (me.createParams.metricType) {
                    params.metric_type = 'prometheus'
                }

                const headers = {}
                me.createParams.httpHeader.forEach(item => {
                    if (item.key) {
                        headers[item.key] = item.value
                    }
                })
                if (Object.keys(headers).length) {
                    params.http_headers = headers
                }

                const constLabels = {}
                me.createParams.constLabels.forEach(item => {
                    if (item.key) {
                        constLabels[item.key] = item.value
                    }
                })
                if (Object.keys(constLabels).length && params.metric_type) {
                    params.const_labels = constLabels
                }

                if (httpMethod === 'POST') {
                    if (me.createParams.httpBodyPost.trim()) {
                        params.http_body = me.createParams.httpBodyPost
                    }
                } else {
                    const bodys = {}
                    me.createParams.httpBodyGet.forEach(item => {
                        if (item.key) {
                            bodys[item.key] = item.value
                        }
                    })
                    if (Object.keys(bodys).length) {
                        params.http_body = JSON.stringify(bodys)
                    }
                }

                try {
                    me.isCreatingOrEditing = true
                    me.creatingOrEditingStr = this.$t('创建Metric中，请稍候...')
                    await me.$store.dispatch('metric/createMetric', params)

                    const res = await me.$store.dispatch('metric/getMetricList', {
                        projectId: me.projectId
                    })
                    me.dataList.splice(0, me.dataList.length, ...(res.data || []))
                    me.dataListTmp.splice(0, me.dataListTmp.length, ...(res.data || []))
                    me.pageConf.current = 1
                    me.searchKeyWord = ''
                    me.initPageConf()
                    // me.curPageData = me.getDataByPage(me.pageConf.current, true)

                    me.searchMetric(false, true)
                    me.hideCreateMetric()
                    me.isCreatingOrEditing = false
                    me.creatingOrEditingStr = ''
                } catch (e) {
                    console.error(e)
                    me.isCreatingOrEditing = false
                    me.creatingOrEditingStr = ''
                }
            },

            /**
             * 显示创建 metric sideslider
             */
            async showCreateMetric () {
                this.resetCreateParams()
                this.createMetricConf.isShow = true
            },

            /**
             * 隐藏创建 metric sideslider
             */
            hideCreateMetric () {
                this.createMetricConf.isShow = false
                this.isCreatingOrEditing = false
                this.creatingOrEditingStr = ''
            },

            /**
             * 跳转到 字段设置或数据清洗 页面
             *
             * @param {Object} metric 当前 metric 对象
             * @param {string} url 要跳转的 url
             */
            async go (metric, url) {
                window.open(url)
            },

            /**
             * 显示编辑 metric sideslider
             *
             * @param {Object} metric 当前 metric 对象
             */
            async editMetric (metric) {
                this.editMetricConf.isShow = true
                this.editMetricConf.title = this.$t(`更新【{metricName}】`, { metricName: metric.name })

                this.editParams.curMetric = metric
                this.editParams.name = metric.name
                this.editParams.port = metric.port
                this.editParams.url = metric.uri
                this.editParams.timeout = metric.timeout
                this.editParams.metricType = !!metric.metric_type
                this.editParams.frequency = metric.frequency
                this.editParams.httpMethod = metric.http_method
                const headersKeyList = Object.keys(metric.http_headers)
                if (headersKeyList.length) {
                    headersKeyList.forEach(key => {
                        this.editParams.httpHeader.push({
                            key: key,
                            value: metric.http_headers[key]
                        })
                    })
                } else {
                    this.editParams.httpHeader.push({
                        key: '',
                        value: ''
                    })
                }

                const constLabelsKeys = Object.keys(metric.const_labels)
                if (constLabelsKeys.length) {
                    constLabelsKeys.forEach(key => {
                        this.editParams.constLabels.push({
                            key: key,
                            value: metric.const_labels[key]
                        })
                    })
                } else {
                    this.editParams.constLabels.push({
                        key: '',
                        value: ''
                    })
                }

                if (this.editParams.httpMethod === 'POST') {
                    this.editParams.httpBodyPost = metric.http_body
                    this.editParams.httpBodyGet.push({
                        key: '',
                        value: ''
                    })
                } else {
                    const bodyKeyList = Object.keys(metric.http_body)
                    if (bodyKeyList.length) {
                        bodyKeyList.forEach(key => {
                            this.editParams.httpBodyGet.push({
                                key: key,
                                value: metric.http_body[key]
                            })
                        })
                    } else {
                        this.editParams.httpBodyGet.push({
                            key: '',
                            value: ''
                        })
                    }
                }
            },

            /**
             * 编辑 Metric 确定按钮
             */
            async confirmEditMetric () {
                const me = this
                const name = me.editParams.name.trim()
                const port = me.editParams.port
                const url = me.editParams.url.trim()
                const timeout = me.editParams.timeout
                const frequency = me.editParams.frequency
                const httpMethod = me.editParams.httpMethod.trim()

                if (!name) {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入名称')
                    })
                    return
                }

                if (name.length < 3) {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('名称不得小于三个字符')
                    })
                    return
                }

                if (port === null || port === undefined || port === '') {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入端口')
                    })
                    return
                }

                if (!url) {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入URI')
                    })
                    return
                }

                if (frequency === null || frequency === undefined || frequency === '') {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入采集频率')
                    })
                    return
                }

                if (timeout === null || timeout === undefined || timeout === '') {
                    me.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入单次采集超时时间')
                    })
                    return
                }

                const params = {
                    projectId: me.projectId,
                    metricId: me.editParams.curMetric.id,
                    name: name,
                    port: port,
                    uri: url,
                    metric_type: '',
                    timeout: timeout,
                    const_labels: {},
                    frequency: frequency,
                    http_method: httpMethod
                }

                if (me.editParams.metricType) {
                    params.metric_type = 'prometheus'
                }

                const constLabels = {}
                me.editParams.constLabels.forEach(item => {
                    if (item.key) {
                        constLabels[item.key] = item.value
                    }
                })
                if (Object.keys(constLabels).length && params.metric_type) {
                    params.const_labels = constLabels
                }

                if (me.editParams.httpHeader.length === 1
                    && me.editParams.httpHeader[0].key === ''
                    && me.editParams.httpHeader[0].value === ''
                ) {
                    params.http_headers = ''
                } else {
                    const headers = {}
                    me.editParams.httpHeader.forEach(item => {
                        if (item.key) {
                            headers[item.key] = item.value
                        }
                    })
                    if (Object.keys(headers).length) {
                        params.http_headers = headers
                    }
                }

                if (httpMethod === 'POST') {
                    // if (me.editParams.httpBodyPost.trim()) {
                    //     params.http_body = me.editParams.httpBodyPost
                    // }
                    params.http_body = me.editParams.httpBodyPost.trim() || ''
                } else {
                    if (me.editParams.httpBodyGet.length === 1
                        && me.editParams.httpBodyGet[0].key === ''
                        && me.editParams.httpBodyGet[0].value === ''
                    ) {
                        params.http_body = '{}'
                    } else {
                        const bodys = {}
                        me.editParams.httpBodyGet.forEach(item => {
                            if (item.key) {
                                bodys[item.key] = item.value
                            }
                        })
                        if (Object.keys(bodys).length) {
                            params.http_body = JSON.stringify(bodys)
                        }
                    }
                }

                try {
                    me.isCreatingOrEditing = true
                    me.creatingOrEditingStr = this.$t('更新Metric中，请稍候...')
                    await me.$store.dispatch('metric/editMetric', params)

                    const res = await me.$store.dispatch('metric/getMetricList', {
                        projectId: me.projectId
                    })
                    me.dataList.splice(0, me.dataList.length, ...(res.data || []))
                    me.dataListTmp.splice(0, me.dataListTmp.length, ...(res.data || []))
                    me.initPageConf()
                    me.searchMetric(false, true)
                    me.hideEditMetric()
                    me.isCreatingOrEditing = false
                    me.creatingOrEditingStr = ''
                } catch (e) {
                    console.error(e)
                    me.isCreatingOrEditing = false
                    me.creatingOrEditingStr = ''
                }
            },

            /**
             * 隐藏创建 metric sideslider
             */
            hideEditMetric () {
                this.editParams = Object.assign({}, {
                    curMetric: null,
                    name: '',
                    port: '',
                    url: '',
                    timeout: 30,
                    metricType: false,
                    frequency: 60,
                    httpMethod: 'GET',
                    httpHeader: [],
                    httpBodyGet: [],
                    constLabels: [],
                    httpBodyPost: ''
                })
                this.editMetricConf.isShow = false
            },

            initMetricInstancePageConf () {
                const total = this.metricInstanceList.length
                this.metricInstancePageConf.totalPage = Math.ceil(total / this.metricInstancePageConf.pageSize)
                this.metricInstancePageChange.total = total
            },
            reloadMetricInstanceCurPage () {
                this.initMetricInstancePageConf()
                if (this.metricInstancePageConf.curPage > this.metricInstancePageConf.totalPage) {
                    this.metricInstancePageConf.curPage = this.metricInstancePageConf.totalPage
                }
                this.curMetricInstancePageData = this.getDataByPage(this.metricInstancePageConf.curPage)
            },
            getMetricInstanceDataByPage (page) {
                let startIndex = (page - 1) * this.metricInstancePageConf.pageSize
                let endIndex = page * this.metricInstancePageConf.pageSize
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.metricInstanceList.length) {
                    endIndex = this.metricInstanceList.length
                }
                const data = this.metricInstanceList.slice(startIndex, endIndex)
                return data
            },
            metricInstancePageChange (page) {
                this.metricInstancePageConf.curPage = page
                const data = this.getMetricInstanceDataByPage(page)
                this.curMetricInstancePageData = JSON.parse(JSON.stringify(data))
            },

            hideInstanceDialog () {
                this.instanceDialogConf.isShow = false
            },

            /**
             * 查看 metric 实例
             *
             * @param {Object} metric 当前 metric 对象
             */
            async checkMetricInstance (metric) {
                this.curInstanceMetric = metric
                this.isMetricInstanceLoading = true
                this.instanceDialogConf.isShow = true
                try {
                    const res = await this.$store.dispatch('metric/checkMetricInstance', {
                        projectId: this.projectId,
                        metricId: metric.id
                    })
                    this.metricInstanceList = res.data
                    this.metricInstancePageConf.curPage = 1
                    this.initMetricInstancePageConf()
                    this.curMetricInstancePageData = this.getMetricInstanceDataByPage(this.metricInstancePageConf.curPage)
                    this.isMetricInstanceLoading = false
                } catch (e) {
                    console.error(e)
                }
            },

            /**
             * 删除 metric
             *
             * @param {Object} metric 当前 metric 对象
             */
            async deleteMetric (metric) {
                const me = this
                me.$bkInfo({
                    title: this.$t('确认删除'),
                    clsName: 'biz-remove-dialog',
                    content: me.$createElement('p', {
                        class: 'biz-confirm-desc'
                    }, `${this.$t('确定要删除Metric')}【${metric.name}】？`),
                    async confirmFn () {
                        try {
                            me.$bkLoading({
                                title: me.$createElement('span', me.$t('删除Metric中，请稍候'))
                            })

                            await me.$store.dispatch('metric/deleteMetric', {
                                projectId: me.projectId,
                                metricId: metric.id
                            })

                            const res = await me.$store.dispatch('metric/getMetricList', {
                                projectId: me.projectId
                            })

                            me.dataList.splice(0, me.dataList.length, ...(res.data || []))
                            me.dataListTmp.splice(0, me.dataListTmp.length, ...(res.data || []))
                            me.pageConf.current = 1
                            me.searchKeyWord = ''
                            me.initPageConf()
                            me.curPageData = me.getDataByPage(me.pageConf.current)
                            // me.searchMetric(true)
                            me.$bkLoading.hide()
                        } catch (e) {
                            console.error(e)
                            me.$bkLoading.hide()
                        }
                    }
                })
            },

            /**
             * 暂停/恢复 metric
             *
             * @param {Object} metric 当前 metric 对象
             * @param {string} idx 暂停/恢复 标识
             * @param {Array} namespaceIdList 命名空间 id 集合
             */
            async pauseAndResume (metric, idx, namespaceIdList = []) {
                const idxStr = idx === 'pause' ? this.$t('暂停') : this.$t('恢复')
                const opType = idx === 'pause' ? 'pause' : 'resume'

                const me = this
                me.$bkInfo({
                    // title: `确认${idxStr}【${metric.name}】？`,
                    title: this.$t(`确认{action}【{metricName}】？`, { action: idxStr, metricName: metric.name }),
                    async confirmFn () {
                        try {
                            me.$bkLoading({
                                // title: me.$createElement('span', `${idxStr}Metric中，请稍候...`)
                                title: me.$createElement('span', me.$t(`{action}Metric中，请稍候`, { action: idxStr }))
                            })

                            await me.$store.dispatch('metric/pauseAndResumeMetric', {
                                projectId: me.projectId,
                                metricId: metric.id,
                                op_type: opType,
                                ns_id_list: namespaceIdList
                            })

                            const res = await me.$store.dispatch('metric/getMetricList', {
                                projectId: me.projectId
                            })

                            me.dataList.splice(0, me.dataList.length, ...(res.data || []))
                            me.dataListTmp.splice(0, me.dataListTmp.length, ...(res.data || []))
                            me.pageConf.current = 1
                            me.searchKeyWord = ''
                            me.initPageConf()
                            me.curPageData = me.getDataByPage(me.pageConf.current)
                            // me.searchMetric(true)
                            me.$bkLoading.hide()
                        } catch (e) {
                            console.error(e)
                            me.$bkLoading.hide()
                        }
                    }
                })
            }
        }
    }
</script>

<style scoped>
    @import './metric-main.css';
</style>
