<template>
    <bk-sideslider
        :is-show.sync="isVisible"
        :title="editMetricConf.title"
        :width="editMetricConf.width"
        :quick-close="false"
        class="biz-metric-manage-create-sideslider"
        @hidden="hideEditMetric">
        <div slot="content">
            <div class="wrapper" style="position: relative;">
                <form class="bk-form bk-form-vertical create-form">
                    <div class="bk-form-item">
                        <label class="bk-label label">{{$t('名称')}}：</label>
                        <div class="bk-form-content">
                            <bk-input v-model="editParams.name" :disabled="true" style="width: 250px;" />
                        </div>
                    </div>
                    <div class="bk-form-item flex-item">
                        <div class="left">
                            <label class="bk-label label">
                                {{$t('选择Service')}}：<span class="red">*</span>
                                <span class="tip">{{$t('选择Service以获取Label')}}</span>
                            </label>
                            <div class="bk-form-content">
                                <bk-input :disabled="true" v-model="clusterName" />
                            </div>
                        </div>
                        <div class="right">
                            <label class="bk-label label">&nbsp;</label>
                            <div class="bk-form-content">
                                <bk-input :disabled="true" :value="`${editParams.namespace}/${editParams.service_name}`" />
                            </div>
                        </div>
                    </div>
                    <div class="bk-form-item" v-if="keyValueData">
                        <label class="bk-label label">
                            {{$t('选择关联Label')}}：<span class="red">*</span>
                        </label>
                        <div class="bk-form-content">
                            <div class="bk-keyer http-header">
                                <div class="biz-keys-list mb10">
                                    <div class="biz-key-item" v-for="(key, keyIndex) in Object.keys(keyValueData)" :key="keyIndex">
                                        <bk-input style="width: 235px;" :disabled="true" :value="key" />
                                        <span class="operator">=</span>
                                        <bk-input style="width: 235px;" :disabled="true" :value="keyValueData[key]" />
                                        <bk-checkbox class="ml10" :value="!!editParams.selector[key]" @change="valueChange(arguments[0], key)"></bk-checkbox>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="bk-form-item" v-if="portList.length">
                        <label class="bk-label label">{{$t('选择PortName')}}：<span class="red">*</span></label>
                        <div class="bk-form-content">
                            <bk-selector
                                searchable
                                :selected.sync="editParams.port"
                                :list="portList"
                                :setting-key="'name'"
                                :display-key="'name'"
                                @item-selected="changeSelectedPort">
                            </bk-selector>
                        </div>
                    </div>
                    <div class="bk-form-item">
                        <label class="bk-label label">
                            {{$t('Metric路径')}}：<span class="red">*</span>
                        </label>
                        <div class="bk-form-content">
                            <bk-input v-model="editParams.path" :placeholder="$t('请输入')" />
                        </div>
                    </div>
                    <div class="bk-form-item">
                        <label class="bk-label label">
                            {{$t('Metric参数')}}：
                        </label>
                        <div class="bk-form-content">
                            <metric-keyer :key-list="metricParamsList"></metric-keyer>
                        </div>
                    </div>
                    <div class="bk-form-item flex-item">
                        <div class="left">
                            <label class="bk-label label">{{$t('采集周期（秒）')}}：<span class="red">*</span></label>
                            <div class="bk-form-content">
                                <bk-number-input
                                    class="text-input-half"
                                    :value.sync="editParams.interval"
                                    :min="0"
                                    :max="999999999"
                                    :debounce-timer="0"
                                    :placeholder="$t('请输入')">
                                </bk-number-input>
                            </div>
                        </div>
                        <div class="right">
                            <label class="bk-label label">{{$t('允许最大Sample数')}}：<span class="red">*</span></label>
                            <div class="bk-form-content">
                                <bk-number-input
                                    class="text-input-half"
                                    :value.sync="editParams.sample_limit"
                                    :min="0"
                                    :debounce-timer="0">
                                </bk-number-input>
                            </div>
                        </div>
                    </div>
                    <div class="action-inner">
                        <bk-button type="primary" :loading="isLoading" @click="confirmEditMetric">
                            {{$t('提交')}}
                        </bk-button>
                        <bk-button type="button" :disabled="isLoading" @click="hideEditMetric">
                            {{$t('取消')}}
                        </bk-button>
                    </div>
                </form>
            </div>
        </div>
    </bk-sideslider>
</template>

<script>
    import MetricKeyer from './keyer'

    export default {
        name: 'EditMetric',
        components: {
            MetricKeyer
        },
        props: {
            isShow: {
                type: Boolean,
                default: false
            },
            clusterId: {
                type: String
            },
            clusterName: {
                type: String
            },
            data: {
                type: Object
            }
        },
        data () {
            return {
                bkMessageInstance: null,
                editMetricConf: {
                    isShow: false,
                    title: this.$t('编辑Metric'),
                    timer: null,
                    width: 644,
                    loading: false
                },
                // 创建的参数
                editParams: {
                    name: '',
                    cluster_id: '',
                    namespace: '',
                    service_name: '',
                    path: '/metrics',
                    selector: {},
                    interval: 30,
                    port: '',
                    sample_limit: 10000
                },

                isVisible: false,
                isLoading: false,
                keyValueData: null,
                portList: [],
                serviceList: [],
                selectedLabels: {},
                curSelectedService: null,
                metricParamsList: []
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
        watch: {
            isShow: {
                async handler (newVal, oldVal) {
                    this.isVisible = newVal
                    if (newVal) {
                        // this.isLoading = true
                        await this.fetchService()
                        this.editParams.name = this.data.name
                        this.editParams.cluster_id = this.data.cluster_id
                        this.editParams.projectId = this.projectId
                        this.editParams.namespace = this.data.namespace
                        this.editParams.service_name = this.data.metadata.service_name
                        this.editParams.selector = Object.assign({}, this.selectedLabels)
                        this.editParams.path = this.data.spec.endpoints[0].path
                        this.editParams.interval = this.data.spec.endpoints[0].interval
                        this.editParams.port = this.data.spec.endpoints[0].port
                        this.editParams.sample_limit = this.data.spec.sampleLimit

                        // 集群ID + resourceName + 命名空间确定唯一
                        const curSelectedService = this.serviceList.find(
                            item => (item.resourceName === this.editParams.service_name)
                                && (item.namespace === this.editParams.namespace)
                        )
                        this.curSelectedService = Object.assign({}, curSelectedService)

                        this.keyValueData = Object.assign({}, curSelectedService.data.metadata.labels)
                        this.portList.splice(0, this.portList.length, ...(curSelectedService.data.spec.ports || []))
                    } else {
                        this.metricParamsList = []
                    }
                },
                immediate: true
            }
        },
        mounted () {
        },
        destroyed () {
            this.bkMessageInstance && this.bkMessageInstance.close()
        },
        methods: {
            /**
             * 获取 service 数据
             */
            async fetchService () {
                try {
                    const res = await this.$store.dispatch('metric/listServices', {
                        projectId: this.projectId,
                        clusterId: this.clusterId
                    })
                    const list = res.data || []
                    list.forEach(item => {
                        item.displayName = `${item.namespace}/${item.resourceName}`
                    })
                    this.serviceList.splice(0, this.serviceList.length, ...list)
                    await this.fetchServiceMonitor()
                } catch (e) {
                    console.error(e)
                } finally {
                    this.isLoading = false
                }
            },

            /**
             * 获取当前service_monitor
             */
            async fetchServiceMonitor () {
                try {
                    const res = await this.$store.dispatch('metric/getServiceMonitor', {
                        projectId: this.projectId,
                        clusterId: this.clusterId,
                        namespace: this.data.namespace,
                        name: this.data.name
                    })
                    this.selectedLabels = Object.assign({}, res.data.spec.selector.matchLabels || {})

                    // const metricParamsList = []
                    const params = Object.assign({}, res.data.spec.endpoints[0].params || {})
                    Object.keys(params).forEach(k => {
                        this.metricParamsList.push({
                            key: k,
                            value: params[k][0] || ''
                        })
                    })
                    if (!this.metricParamsList.length) {
                        this.metricParamsList = [{ key: '', value: '' }]
                    }
                } catch (e) {
                    console.error(e)
                } finally {
                    this.isLoading = false
                }
            },

            /**
             * 切换 port 下拉框改变事件
             *
             * @param {string} port 选择的 port name
             */
            changeSelectedPort (port) {
                this.editParams.port = port
            },

            /**
             * 选择关联 Label 多选框勾选
             *
             * @param {Object} e 事件对象
             * @param {string} key 勾选的 key
             */
            valueChange (checked, key) {
                if (checked) {
                    this.editParams.selector[key] = this.keyValueData[key]
                } else {
                    delete this.editParams.selector[key]
                }
            },

            /**
             * 重置添加 metric 的参数
             */
            resetParams () {
                this.keyValueData = null
                this.portList.splice(0, this.portList.length, ...[])
                this.editParams = Object.assign({}, {
                    name: '',
                    cluster_id: '',
                    namespace: '',
                    service_name: '',
                    path: '/metrics',
                    selector: {},
                    interval: 30,
                    port: '',
                    sample_limit: 10000
                })
            },

            /**
             * 隐藏编辑 metric sideslider
             */
            hideEditMetric () {
                this.editMetricConf.isShow = false

                this.resetParams()
                this.$emit('hide-edit-metric', false)
            },

            /**
             * 编辑 Metric 确定按钮
             */
            async confirmEditMetric () {
                const path = this.editParams.path.trim()
                const port = this.editParams.port
                const selector = this.editParams.selector
                const interval = this.editParams.interval
                const sampleLimit = this.editParams.sample_limit

                if (!Object.keys(selector).length) {
                    this.$bkMessage({ theme: 'error', message: this.$t('至少要关联一个Label') })
                    return
                }

                if (!port) {
                    this.$bkMessage({ theme: 'error', message: this.$t('请选择PortName') })
                    return
                }

                if (!path) {
                    this.$bkMessage({ theme: 'error', message: this.$t('请输入Metric路径') })
                    return
                }

                if (path.length < 2) {
                    this.$bkMessage({ theme: 'error', message: this.$t('Metric路径不得小于两个字符') })
                    return
                }

                if (interval === null || interval === undefined || interval === '') {
                    this.$bkMessage({ theme: 'error', message: this.$t('请输入采集周期') })
                    return
                }

                if (sampleLimit === '' || sampleLimit === null || sampleLimit === undefined) {
                    this.$bkMessage({ theme: 'error', message: this.$t('请输入允许最大Sample数') })
                    return
                }

                const metricParamsList = []
                const len = this.metricParamsList.length

                for (let i = 0; i < len; i++) {
                    const metricParams = this.metricParamsList[i]
                    const key = metricParams.key.trim()
                    const value = metricParams.value.trim()
                    if (!key && value) {
                        this.bkMessageInstance && this.bkMessageInstance.close()
                        this.bkMessageInstance = this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请填写参数名称')
                        })
                        return
                    } else if (key && !value) {
                        this.bkMessageInstance && this.bkMessageInstance.close()
                        this.bkMessageInstance = this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请填写参数值')
                        })
                        return
                    }
                    if (key && value) {
                        metricParamsList.push({
                            key,
                            value
                        })
                    }
                }

                if (metricParamsList.length) {
                    this.editParams.params = {}
                    metricParamsList.forEach(m => {
                        this.editParams.params[m.key] = [m.value]
                    })
                }

                try {
                    this.isLoading = true
                    await this.$store.dispatch('metric/updateServiceMonitor', Object.assign({}, this.editParams))
                    this.$emit('edit-success')
                } catch (e) {
                    console.error(e)
                } finally {
                    this.isLoading = false
                }
            }
        }
    }
</script>

<style>
    @import './create.css';
</style>
