/**
 * @fileOverview  service mixin
 */
import bkKeyer from '@/components/keyer'
import mixinBase from '@/mixins/network/mixin-base'
import { catchErrorHandler } from '@/common/util'

export default {
    components: {
        bkKeyer
    },
    mixins: [mixinBase],
    data () {
        return {}
    },
    computed: {
        projectId () {
            return this.$route.params.projectId
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
        serviceList () {
            const list = this.$store.state.network.serviceList
            list.forEach(item => {
                item.isChecked = false
            })
            return JSON.parse(JSON.stringify(list))
        },
        selector () {
            if (this.curService && this.curService.data.spec.selector) {
                let results = ''
                const selector = Object.entries(this.curService.data.spec.selector)
                selector.forEach(item => {
                    const key = item[0]
                    const value = item[1]
                    results += key + '=' + value + '\n'
                })
                return results
            } else {
                return ''
            }
        },
        labelList () {
            if (this.curService && this.curService.data.metadata.labels) {
                const labels = Object.entries(this.curService.data.metadata.labels)
                return labels
            } else {
                return []
            }
        },
        endpoints () {
            return this.$store.state.network.endpoints
        },
        curLabelList () {
            const list = []
            const labels = this.curServiceDetail?.config.metadata.labels || {}
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
        }
    },
    watch: {
        curServiceDetail () {
            const metadata = this.curServiceDetail.config.metadata
            if (metadata.lb_labels && metadata.lb_labels.BCSBALANCE) {
                this.algorithmIndex = metadata.lb_labels.BCSBALANCE
            } else {
                this.algorithmIndex = -1
            }
        },
        curPageData () {
            this.curPageData.forEach(item => {
                if (item.status === 'updating') {
                    this.getServiceStatus(item)
                }
            })
        }
    },
    methods: {
        /**
         * 刷新列表
         */
        refresh () {
            this.pageConf.current = 1
            this.isPageLoading = true
            this.getServiceList()
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
         * 切换页面时，清空轮询请求
         */
        leaveCallback () {
            for (const key of Object.keys(this.statusTimer)) {
                clearInterval(this.statusTimer[key])
            }
        },

        /**
         * 获取service的状态
         * @param  {object} service service
         * @param  {number} index   service索引
         */
        getServiceStatus (service, index) {
            const projectId = this.projectId
            const name = service.resourceName
            const namespace = service.namespace

            if (this.statusTimer[service._id]) {
                clearInterval(this.statusTimer[service._id])
            } else {
                this.statusTimer[service._id] = 0
            }

            // 对单个service的状态进行不断2秒间隔的查询
            this.statusTimer[service._id] = setInterval(async () => {
                try {
                    const res = await this.$store.dispatch('network/getServiceStatus', {
                        projectId,
                        namespace,
                        name
                    })
                    const data = res.data.data[0]

                    if (data.status !== 'updating') {
                        service.status = data.status
                        service.can_update = data.can_update
                        service.can_delete = data.can_delete
                        service.can_update_msg = data.can_update_msg
                        service.can_delete_msg = data.can_delete_msg
                        clearInterval(this.statusTimer[service._id])
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            }, 2000)
        },

        /**
         * 确认删除service
         * @param  {object} service service
         */
        async removeService (service) {
            const self = this

            this.$bkInfo({
                title: this.$t('确认删除'),
                clsName: 'biz-remove-dialog max-size',
                content: this.$createElement('p', {
                    class: 'biz-confirm-desc'
                }, [
                    `${this.$t('确定要删除Service【')}`,
                    this.$createElement('strong', service.cluster_id),
                    ' / ',
                    this.$createElement('strong', service.namespace),
                    ' / ',
                    this.$createElement('strong', service.resourceName),
                    '】？'
                ]),
                async confirmFn () {
                    self.deleteService(service)
                }
            })
        },

        /**
         * 删除service
         * @param  {object} service service
         */
        async deleteService (service) {
            const projectId = this.projectId
            const namespace = service.namespace
            const clusterId = service.clusterId
            const serviceId = service.resourceName
            this.isPageLoading = true
            try {
                await this.$store.dispatch('network/deleteService', {
                    projectId,
                    clusterId,
                    namespace,
                    serviceId
                })

                this.$bkMessage({
                    theme: 'success',
                    message: this.$t('删除成功')
                })
                this.initPageConf()
                this.getServiceList()
            } catch (e) {
                catchErrorHandler(e, this)
                this.isPageLoading = false
            }
        }
    }
}
