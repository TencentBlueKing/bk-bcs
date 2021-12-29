/**
 * @file 应用 instantiation 详情页的 mixin
 */

import yamljs from 'js-yaml'

import { catchErrorHandler, escape } from '@/common/util'
import ace from '@/components/ace-editor'

const ARR = [
    'Application',
    'Deployment',
    'Service',
    'ConfigMap',
    'Secret',
    'DaemonSet',
    'Job',
    'StatefulSet',
    'Ingress'
]

const ABBR_ARR = [
    'app',
    'dep',
    'svc',
    'cm',
    'srt',
    'ds',
    'job',
    'sts',
    'Ing'
]

export default {
    components: {
        ace
    },
    data () {
        return {
            tplList: [],
            // 存放给的是模板的 id 集合
            tplIndex: [],
            tplsetVerList: [],
            // 存放给的是模板集的 show_version_id
            tplsetVerIndex: -1,
            // 存放给的是模板集的 id
            tplsetVerId: -1,
            dialogConf: {
                isShow: false,
                width: 912,
                // width: 895,
                title: this.$t('选择运行的集群及命名空间'),
                closeIcon: false,
                loading: false
            },
            existList: [],
            curTemplateTmp: {},
            // 弹层中的 namespace 集合
            candidateNamespaceList: [],
            // 在弹层中选择的 namespace 缓存
            namespaceListTmp: {},
            // 弹层点击确定后把 namespaceListTmp 赋值给 selectedNamespaceList，用于显示在页面上
            selectedNamespaceList: [],
            selectedNamespaceCluster: {},
            previewTitle: '',
            previewList: [],
            previewLoading: false,
            previewShow: false,
            previewNs: null,
            lbServiceList: [],
            lbServiceListInPage: [],
            // 映射 lb 的数据，下拉框里面的是 id
            lbServiceListInPageTmp: {},
            lbSelectData: {},
            // 没有 lb 信息的 ns 的集合
            invalidNsList: [],
            // 创建后启动实例多选框
            createInstanceLoading: false,
            variableMap: {},
            bkMessageInstance: null,
            exceptionCode: null,
            editorConfig: {
                width: '100%',
                height: '100%',
                lang: 'json',
                readOnly: true,
                fullScreen: false,
                values: [],
                editors: []
            },
            // 模板多选下拉框选择的值，提交给后端的
            instanceEntity: {},
            projectId: '',
            goNamespaceDialogConf: {
                isShow: false,
                width: 295,
                title: '',
                closeIcon: false,
                hasFooter: false,
                hasHeader: false
            },
            curProject: null,
            isSelectAllTpl: false
        }
    },
    computed: {
        projectCode () {
            return this.$route.params.projectCode
        },
        templateId () {
            return this.$route.params.templateId
        },
        category () {
            return this.$route.params.category
        },
        tmplAppId () {
            return this.$route.params.tmplAppId
        },
        tmplAppName () {
            return this.$route.params.tmplAppName
        },
        curTemplate: {
            get () {
                if (this.$route.params.curTemplate) {
                    this.curTemplateTmp = this.$route.params.curTemplate
                }
                return this.curTemplateTmp
            },
            set (val) {
                this.curTemplateTmp = val
                return this.curTemplateTmp
            }
        },
        onlineProjectList () {
            return this.$store.state.sideMenu.onlineProjectList
        },
        searchParamsList () {
            return this.$route.params.searchParamsList
        },
        isEn () {
            return this.$store.state.isEn
        },
        curClusterId () {
            return this.$store.state.curClusterId
        },
        curProjectId () {
            return this.$store.state.curProjectId
        }
    },
    created () {
        // router > localStorage > onlineProjectList[0]
        const len = this.onlineProjectList.length
        if (len) {
            this.projectId = this.$route.params.projectId
                || this.curProjectId
                || this.onlineProjectList[0].project_id

            for (let i = 0; i < len; i++) {
                const project = this.onlineProjectList[i]
                if (project.project_id === this.projectId) {
                    this.curProject = Object.assign({}, project)
                    break
                }
            }
            // k8s
            if (this.curProject.kind === PROJECT_K8S || this.curProject.kind === PROJECT_TKE) {
                this.editorConfig.lang = 'yaml'
            } else {
                this.editorConfig.lang = 'json'
            }
            if (Object.keys(this.curTemplate).length === 0) {
                this.fetchTemplate()
            }
            this.fetchTemplatesetVerList()
            this.fetchNamespaceList()
        }
    },
    methods: {
        /**
         * 模板集的排序，顺序依次为：
         * Application/app
         * Deployment/dep
         * Service/svc
         * ConfigMap/cm
         * Secret/srt
         * DaemonSet/ds
         * Job/job
         * StatefulSet/sts
         * Ingress/Ing
         *
         * @param {Array} list 要排序的数组
         * @param {string} key 要排序的 key
         * @param {boolean} isAbbr 是否是简写
         *
         * @return {Array} 结果
         */
        sortTplType (list, key, isAbbr) {
            const arr = isAbbr ? ABBR_ARR : ARR
            return list.sort((a, b) => arr.indexOf(a[key]) - arr.indexOf(b[key]) >= 0)
        },

        /**
         * ace 编辑器 annotation change 回调
         *
         * @param {number} index 当前是哪个 ace editor 的索引
         * @param {Array} annotations annotations 数据
         */
        changeAnnotation (index, annotations) {
            this.editorConfig.editors[index].gotoLine(annotations[0].row + 1, annotations[0].column, true)
        },

        /**
         *  编辑器初始化之后的回调函数
         *  @param editor - 编辑器对象
         */
        editorInitAfter (editor) {
            this.editorConfig.editors.push(editor)
            setTimeout(() => {
                editor.resize(true)
                editor.gotoLine(0, 0, true)
            }, 100)
        },

        /**
         * 获取当前 template 的信息
         */
        async fetchTemplate () {
            try {
                const res = await this.$store.dispatch('configuration/getTemplateById', {
                    projectId: this.projectId,
                    templateId: this.templateId
                })
                this.curTemplate = Object.assign({}, res.data)
            } catch (e) {
                catchErrorHandler(e, this)
            }
        },

        /**
         * 加载所有命名空间列表，不在点击选择命名空间按钮点击事件中加载，而是提前加载
         * 是为了防止选择命名空间弹层异步加载数据时高度变化
         * 获取到 candidateNamespaceList，弹层中所有的命名空间集合
         */
        async fetchNamespaceList () {
            try {
                const res = await this.$store.dispatch('configuration/getAllNamespaceList', {
                    projectId: this.projectId,
                    group_by: 'cluster_name',
                    perm_can_use: 1
                })

                const list = res.data
                list.forEach(item => {
                    this.candidateNamespaceList.push({ ...item, isOpen: false })
                })
            } catch (e) {
                catchErrorHandler(e, this)
            }
        },

        /**
         * 获取模板集版本。模板集版本下拉框数据
         */
        async fetchTemplatesetVerList () {
            try {
                const res = await this.$store.dispatch('configuration/getTemplatesetVerList', {
                    projectId: this.projectId,
                    templateId: this.templateId
                })
                const list = res.data.results || []
                list.forEach(item => {
                    this.tplsetVerList.push({
                        id: item.id,
                        name: item.version,
                        show_version_id: item.show_version_id,
                        show_version_name: item.show_version_name
                    })
                })
            } catch (e) {
                catchErrorHandler(e, this)
            }
        },

        /**
         * 切换模板集下拉框，获取模板下拉框的数据
         *
         * @param {number} index 索引
         * @param {Object} data 当前下拉框数据
         */
        async changeTplset (index, data) {
            this.tplIndex.splice(0, this.tplIndex.length, ...[])
            this.tplList.splice(0, this.tplList.length, ...[])

            this.tplsetVerId = data.id

            this.clearCandidateNamespaceStatus()
            this.clearNamespaceStatus()

            this.previewTitle = ''
            this.previewNs = Object.assign({}, {})
            this.previewList.splice(0, this.previewList.length, ...[])
            this.editorConfig.editors.splice(0, this.editorConfig.editors.length, ...[])
            this.editorConfig.values.splice(0, this.editorConfig.values.length, ...[])
            this.previewShow = false

            // 清空已经选择的 namespace
            this.selectedNamespaceList.splice(0, this.selectedNamespaceList.length, ...[])
            this.selectedNamespaceCluster = {}

            try {
                const res = await this.$store.dispatch('configuration/getTemplateListByIdCategoryTmplName', {
                    projectId: this.projectId,
                    tplVerId: data.id,
                    tmplAppName: this.tmplAppName,
                    category: this.category
                })

                const tplList = []
                const tplIndex = []
                const tplData = res.data.data || {}
                Object.keys(tplData).forEach(key => {
                    const obj = {
                        name: key,
                        children: []
                    }
                    tplData[key].forEach(item => {
                        // 默认全选中
                        // tplIndex.push(item.id)
                        tplIndex.push(key + '-' + item.id)
                        obj.children.push({
                            id: item.id,
                            settingKey: key + '-' + item.id,
                            name: item.name,
                            type: key
                        })
                    })
                    tplList.push(obj)
                })
                this.sortTplType(tplList, 'name')

                this.tplIndex.splice(0, this.tplIndex.length, ...tplIndex)
                this.tplList.splice(0, this.tplList.length, ...tplList)
                this.isSelectAllTpl = true

                this.instanceEntity = Object.assign({}, tplData)
            } catch (e) {
                console.error(e)
            }
        },

        /**
         * 模板 多选下拉框选择事件
         *
         * @param {Array} index 索引的数组
         * @param {Array} data 选择的数据对象的数组
         */
        multiSelect (index, data) {
            const ret = {}
            data.forEach(item => {
                if (!ret[item.type]) {
                    ret[item.type] = []
                }
                ret[item.type].push({
                    id: item.id,
                    name: item.name
                })
            })

            let count = 0
            this.tplList.forEach(item => {
                count += (item.children || []).length
            })
            this.isSelectAllTpl = data.length === count

            this.instanceEntity = Object.assign({}, ret)

            this.clearCandidateNamespaceStatus()
            this.clearNamespaceStatus()
            this.previewTitle = ''
            this.previewNs = Object.assign({}, {})
            this.previewList.splice(0, this.previewList.length, ...[])
            this.editorConfig.editors.splice(0, this.editorConfig.editors.length, ...[])
            this.editorConfig.values.splice(0, this.editorConfig.values.length, ...[])
            this.previewShow = false

            // 清空已经选择的 namespace
            this.selectedNamespaceList.splice(0, this.selectedNamespaceList.length, ...[])
            this.selectedNamespaceCluster = {}
        },

        /**
         * 全选模板
         */
        selectAllTpl () {
            if (this.isSelectAllTpl) {
                this.isSelectAllTpl = false
                this.tplIndex = [...[]]
                this.yamlTplResource = [...[]]
                this.instanceEntity = Object.assign({}, {})
                return
            }

            const ret = {}
            const tplIndex = []
            const yamlTplResource = []
            this.tplList.forEach(item => {
                if (!ret[item.name]) {
                    ret[item.name] = []
                }

                const yamlResource = {
                    resource_name: item.name,
                    files: []
                }
                item.children.forEach(child => {
                    ret[item.name].push({
                        id: child.id,
                        name: child.name
                    })
                    tplIndex.push(child.settingKey)
                    yamlResource.files.push({
                        id: child.id,
                        name: child.name
                    })
                })
                yamlTplResource.push(yamlResource)
            })

            this.tplIndex = [...tplIndex]
            this.yamlTplResource = [...yamlTplResource]
            this.instanceEntity = Object.assign({}, ret)

            this.clearCandidateNamespaceStatus()
            this.clearNamespaceStatus()
            this.previewTitle = ''
            this.previewNs = Object.assign({}, {})
            this.previewList.splice(0, this.previewList.length, ...[])
            this.editorConfig.editors.splice(0, this.editorConfig.editors.length, ...[])
            this.editorConfig.values.splice(0, this.editorConfig.values.length, ...[])
            this.previewShow = false

            // 清空已经选择的 namespace
            this.selectedNamespaceList.splice(0, this.selectedNamespaceList.length, ...[])
            this.selectedNamespaceCluster = {}
            this.isSelectAllTpl = true
        },

        /**
         * 根据模板集 id 获取已经被使用过的 namespace
         */
        async fetchExistNamespace () {
            try {
                const res = await this.$store.dispatch('configuration/getExistNamespace', {
                    projectId: this.projectId,
                    tplVerId: this.tplsetVerId,
                    instanceEntity: this.instanceEntity
                })
                const existNamespaceList = res.data.ns_resources || {}
                const list = []
                const existList = Object.keys(existNamespaceList)

                list.splice(0, 0, ...this.candidateNamespaceList)
                list.forEach(item => {
                    (item.results || []).forEach(ns => {
                        if (existList.indexOf(String(ns.id)) > -1) {
                            const message = existNamespaceList[ns.id].join(', ')
                            ns.isExist = true
                            ns.message = this.isEn
                                ? `An instance of the same name of the ${message.replace(/K8s/ig, '')} type under the`
                                    + ` namespace【${ns.name}】, which cannot be instantiated again.`
                                : `命名空间【${ns.name}】下存在${message.replace(/K8s/ig, '')}类型的同名实例, 无法再次实例化`
                        } else {
                            ns.isExist = false
                        }
                    })
                    item.results.sort((cur, next) => {
                        // boolean（强制转数整数）相减，isEist为true排后面
                        return cur.isExist - next.isExist
                    })
                })
                this.existList.splice(0, this.existList.length, ...existList)
                this.candidateNamespaceList.splice(0, this.candidateNamespaceList.length, ...list)
            } catch (e) {
                console.error(e)
            } finally {
                setTimeout(() => {
                    this.dialogConf.loading = false
                }, 300)
            }
        },

        /**
         * 显示选择命名空间弹层
         */
        async showChooseDialog () {
            if (this.tplsetVerId === -1) {
                this.bkMessageInstance && this.bkMessageInstance.close()
                this.bkMessageInstance = this.$bkMessage({
                    theme: 'error',
                    message: this.$t('请选择模板集版本')
                })
                return
            }

            if (!this.tplIndex.length) {
                this.bkMessageInstance && this.bkMessageInstance.close()
                this.bkMessageInstance = this.$bkMessage({
                    theme: 'error',
                    message: this.$t('请选择要实例化的模板')
                })
                return
            }

            const { candidateNamespaceList, selectedNamespaceList } = this

            if (!candidateNamespaceList.length) {
                this.goNamespaceDialogConf.isShow = true
                return
            }

            this.dialogConf.isShow = true
            this.dialogConf.loading = true

            await this.fetchExistNamespace()

            // 清除弹层中的选中状态，不需要清除已选择的 ns 的状态
            this.clearCandidateNamespaceStatus()

            // 之前选择过，那么把之前选择的回填到弹层中，同时展开有选择的
            if (selectedNamespaceList.length) {
                selectedNamespaceList.forEach(ns => {
                    candidateNamespaceList[ns.candidateIndex].isOpen = true
                    candidateNamespaceList[ns.candidateIndex].results[ns.index].isChoose = true
                    this.$set(candidateNamespaceList, ns.candidateIndex, candidateNamespaceList[ns.candidateIndex])

                    this.namespaceListTmp[`${ns.env_type}_${ns.id}`] = {
                        ...ns,
                        candidateIndex: ns.candidateIndex,
                        index: ns.index
                    }
                })
            } else {
                // 之前没选择过，那么展开第一个
                if (!this.curClusterId) {
                    candidateNamespaceList[0].isOpen = true
                    return
                }
                // 单集群状态默认展开
                const curCandidateNamespace = candidateNamespaceList.find(i => i.cluster_id === this.curClusterId)
                curCandidateNamespace.isOpen = true
            }
        },

        /**
         * 清除弹层中 namespace trigger 的展开以及 namespace 的选中
         */
        clearCandidateNamespaceStatus () {
            const list = this.candidateNamespaceList
            list.forEach(item => {
                item.isOpen = false
                item.results.forEach(ns => {
                    ns.isChoose = false
                })
            })

            this.candidateNamespaceList.splice(0, this.candidateNamespaceList.length, ...list)
            this.namespaceListTmp = {}
        },

        /**
         * 清除 selectedNamespaceList 中的选中状态
         */
        clearNamespaceStatus () {
            const selectedNamespaceList = this.selectedNamespaceList
            this.selectedNamespaceCluster = {}
            selectedNamespaceList.forEach(ns => {
                ns.isSelected = false
                if (!this.selectedNamespaceCluster[ns.cluster_id]) {
                    this.selectedNamespaceCluster[ns.cluster_id] = {
                        cluster_id: ns.cluster_id,
                        environment: ns.environment,
                        cluster_name: ns.cluster_name,
                        namespaceList: [ns]
                    }
                } else {
                    this.selectedNamespaceCluster[ns.cluster_id].namespaceList.push(ns)
                }
            })
            this.selectedNamespaceList.splice(0, this.selectedNamespaceList.length, ...selectedNamespaceList)
        },

        /**
         * 收起所有的 trigger
         */
        collapseTrigger () {
            const list = this.candidateNamespaceList
            list.forEach(item => {
                item.isOpen = false
            })
            this.candidateNamespaceList.splice(0, this.candidateNamespaceList.length, ...list)
        },

        /**
         * 选择命名空间弹层 trigger 点击事件
         *
         * @param {Object} item 当前 namespace 对象
         * @param {number} index 当前 namespace 对象的索引
         */
        triggerHandler (item, index) {
            // 展开时滚动条回到顶部
            // document.querySelectorAll('.namespace-wrapper .namespace-inner')[1].scrollTop = 0
            this.collapseTrigger()
            item.isOpen = !item.isOpen
            this.$set(this.candidateNamespaceList, index, item)
        },

        /**
         * 在弹层中选择命名空间
         *
         * @param {number} index candidateNamespaceList 的索引
         * @param {Object} namespace 当前点击的这个 namespace
         * @param {number} i 当前点击的这个 namespace 在 item.results 的索引
         */
        selectNamespaceInDialog (index, namespace, i) {
            namespace.isChoose = !namespace.isChoose
            this.$set(this.candidateNamespaceList[index].results, i, namespace)
            if (this.namespaceListTmp[`${namespace.env_type}_${namespace.id}`]) {
                delete this.namespaceListTmp[`${namespace.env_type}_${namespace.id}`]
            } else {
                this.namespaceListTmp[`${namespace.env_type}_${namespace.id}`] = {
                    ...namespace,
                    candidateIndex: index,
                    index: i
                }
            }
        },

        /**
         * 在弹层中全选命名空间
         *
         * @param {Object} item 当前的 candidateNamespace 对象
         * @param {number} index 当前的 candidateNamespace 对象在 candidateNamespaceList 中的索引
         */
        selectAll (item, index) {
            this.collapseTrigger()
            item.results.forEach((ns, i) => {
                if (this.existList.indexOf(ns.id) < 0) {
                    ns.isChoose = true
                    this.namespaceListTmp[`${ns.env_type}_${ns.id}`] = {
                        ...ns,
                        candidateIndex: index,
                        index: i
                    }
                }
            })
            item.isOpen = true
            this.$set(this.candidateNamespaceList, index, item)
        },

        /**
         * 在弹层中反选命名空间
         *
         * @param {Object} item 当前的 candidateNamespace 对象
         * @param {number} index 当前的 candidateNamespace 对象在 candidateNamespaceList 中的索引
         */
        selectInvert (item, index) {
            this.collapseTrigger()
            item.results.forEach((ns, i) => {
                if (this.existList.indexOf(ns.id) < 0) {
                    ns.isChoose = !ns.isChoose
                    if (this.namespaceListTmp[`${ns.env_type}_${ns.id}`]) {
                        delete this.namespaceListTmp[`${ns.env_type}_${ns.id}`]
                    } else {
                        this.namespaceListTmp[`${ns.env_type}_${ns.id}`] = {
                            ...ns,
                            candidateIndex: index,
                            index: i
                        }
                    }
                }
            })
            item.isOpen = true
            this.$set(this.candidateNamespaceList, index, item)
        },

        /**
         * 选择命名空间弹层确认
         */
        async confirmSelect () {
            const list = Object.keys(this.namespaceListTmp)
            if (list.length === 0) {
                this.bkMessageInstance && this.bkMessageInstance.close()
                this.bkMessageInstance = this.$bkMessage({
                    theme: 'error',
                    message: this.$t('请选择命名空间')
                })
                return
            }

            const namespaces = []
            list.forEach(item => {
                namespaces.push(item.split('_')[1])
            })

            this.dialogConf.loading = true
            try {
                const res = await this.$store.dispatch('configuration/getLbVariable', {
                    projectId: this.projectId,
                    tplVerId: this.tplsetVerId,
                    namespaces: namespaces.join(','),
                    instanceEntity: this.instanceEntity
                })

                this.lbServiceList.splice(0, this.lbServiceList.length, ...(res.data.lb_services || []))

                // 前一次的 lbSelectData
                const prevLbSelectData = Object.assign({}, this.lbSelectData)
                this.lbSelectData = Object.assign({}, {})

                const lbServiceListLen = this.lbServiceList.length

                const variableList = res.data.variable_dict || {}

                const tmp = []
                list.forEach(key => {
                    const obj = Object.assign({}, this.namespaceListTmp[key])
                    obj.variableList = variableList[key.split('_')[1]] || []

                    const lbSelectDataKey = obj.name
                    this.lbSelectData[lbSelectDataKey] = prevLbSelectData[lbSelectDataKey] || {}

                    for (let i = 0; i < lbServiceListLen; i++) {
                        const serviceName = this.lbServiceList[i].name
                        if (this.lbSelectData[lbSelectDataKey][serviceName] === null
                            || this.lbSelectData[lbSelectDataKey][serviceName] === undefined
                        ) {
                            this.lbSelectData[lbSelectDataKey][serviceName] = -1
                        }
                    }

                    tmp.push(obj)
                })

                this.selectedNamespaceList.splice(0, this.selectedNamespaceList.length, ...tmp)
                this.dialogConf.isShow = false

                this.invalidNsList.splice(0, this.invalidNsList.length, ...[])
                this.previewList.splice(0, this.previewList.length, ...[])
                this.lbServiceListInPage.splice(0, this.lbServiceListInPage.length, ...[])

                // 点击选择命名空间弹层确认时，如果之前在 selectedNamespaceList 有选中，那么保持这个选中，同时强制发送
                // 这个选中的 ns 的 previewNamespace 请求，如果没有选中，那么默认选中 selectedNamespaceList 第一个，
                // 同时强制发送这个 ns 的 previewNamespace 请求
                const alreadySelected = this.selectedNamespaceList.filter(item => item.isSelected)[0]
                this.previewTitle = ''
                this.previewNs = Object.assign({}, {})
                this.previewList.splice(0, this.previewList.length, ...[])
                this.editorConfig.editors.splice(0, this.editorConfig.editors.length, ...[])
                this.editorConfig.values.splice(0, this.editorConfig.values.length, ...[])
                this.previewShow = false
                this.previewNamespace(alreadySelected || this.selectedNamespaceList[0], 0, true)
            } catch (e) {
                console.error(e)
            } finally {
                setTimeout(() => {
                    this.dialogConf.loading = false
                    this.dialogConf.isShow = false
                }, 300)
            }
        },

        /**
         * 点击已选择的 namespace
         *
         * @param {Object} ns 当前点击的 namespace 对象
         * @param {number} index 当前点击的 namespace 对象的索引
         */
        async previewNamespace (ns, index, forceSelect) {
            if (this.tplsetVerId === -1) {
                this.bkMessageInstance && this.bkMessageInstance.close()
                this.bkMessageInstance = this.$bkMessage({
                    theme: 'error',
                    message: this.$t('请选择模板集版本')
                })
                return
            }

            if (!this.tplIndex.length) {
                this.bkMessageInstance && this.bkMessageInstance.close()
                this.bkMessageInstance = this.$bkMessage({
                    theme: 'error',
                    message: this.$t('请选择要实例化的模板')
                })
                return
            }

            this.previewList.splice(0, this.previewList.length, ...[])
            this.previewNs = Object.assign({}, {})
            this.editorConfig.editors.splice(0, this.editorConfig.editors.length, ...[])
            this.editorConfig.values.splice(0, this.editorConfig.values.length, ...[])

            this.previewLoading = true
            this.previewShow = true

            const variableInfo = {
                [`${ns.id}`]: {}
            }
            ns.variableList.forEach(variable => {
                variableInfo[`${ns.id}`][variable.key] = variable.value
            })

            this.clearNamespaceStatus()
            ns.isSelected = !ns.isSelected
            this.$set(this.selectedNamespaceList, index, ns)

            try {
                if (this.lbServiceList.length) {
                    const lbRes = await this.$store.dispatch('configuration/getLbInfo', {
                        projectId: this.projectId,
                        clusterId: ns.cluster_id
                    })

                    const lbData = lbRes.data || []

                    // 当前这个 ns 没有 lb 信息，不需要展示变量了，也不能 preview 和提交
                    if (!lbData.length) {
                        this.previewTitle = this.$t('{nsName}的详细配置', { nsName: ns.name })
                        this.previewNs = Object.assign({}, ns)
                        if (this.invalidNsList.indexOf(ns.name) < 0) {
                            this.invalidNsList.splice(0, this.invalidNsList.length, ...[].concat(ns.name))
                        }
                        return
                    }

                    const i = this.invalidNsList.indexOf(ns.name)
                    if (i !== -1) {
                        this.invalidNsList.splice(i, 1)
                    }

                    const lbServiceListInPage = []
                    const lbServiceListInPageTmp = {}

                    let segment = []
                    const lbServiceListLen = this.lbServiceList.length
                    for (let i = 0; i < lbServiceListLen; i++) {
                        const key = this.lbServiceList[i].name
                        if (!lbServiceListInPageTmp[key]) {
                            lbServiceListInPageTmp[key] = {}
                        }
                        const tmp = []
                        lbData.forEach((item, index) => {
                            tmp.push({
                                service: key,
                                ...item
                            })
                            lbServiceListInPageTmp[key][item.lb_id] = item.lb_name
                        })
                        segment.push({
                            key: key,
                            value: tmp
                        })

                        if (i % 2 !== 0) {
                            segment = []
                        } else {
                            lbServiceListInPage.push(segment)
                        }
                    }

                    this.lbServiceListInPage.splice(0, this.lbServiceListInPage.length, ...lbServiceListInPage)
                    this.lbServiceListInPageTmp = Object.assign({}, lbServiceListInPageTmp)

                    this.previewTitle = ns.cluster_name + ' / ' + this.$t('{nsName}的详细配置', { nsName: ns.name })
                    this.previewNs = Object.assign({}, ns)

                    if (!this.checkCurNamespacePreview(ns)) {
                        this.previewList.splice(0, this.previewList.length, ...[])
                        this.editorConfig.editors.splice(0, this.editorConfig.editors.length, ...[])
                        this.editorConfig.values.splice(0, this.editorConfig.values.length, ...[])
                        return
                    }

                    this.preview4Lb(this.previewNs)
                } else {
                    const curTplsetVer = this.tplsetVerList.filter(
                        tplsetVer => tplsetVer.id === this.tplsetVerId
                    )[0] || {}

                    const res = await this.$store.dispatch('configuration/previewNamespace', {
                        projectId: this.projectId,
                        namespace: ns.id,
                        version_id: curTplsetVer.id,
                        show_version_id: curTplsetVer.show_version_id,
                        show_version_name: curTplsetVer.show_version_name,
                        instance_entity: this.instanceEntity,
                        variable_info: variableInfo
                    })

                    const list = []
                    const data = res.data || {}
                    Object.keys(data).forEach(key => {
                        data[key].forEach(item => {
                            const content = JSON.stringify(item.config, null, 4)
                            list.push({
                                name: `${key}:${item.name}`, // 防止不同资源名称相同冲突
                                tag: key,
                                content: content,
                                originalContent: content
                            })
                        })
                    })
                    this.previewTitle = ns.cluster_name + ' / ' + this.$t('{nsName}的详细配置', { nsName: ns.name })
                    this.previewNs = Object.assign({}, ns)

                    this.sortTplType(list, 'tag', true)

                    this.previewList.splice(0, this.previewList.length, ...list)
                    this.previewList.forEach((preview, index) => {
                        this.editorConfig.values[index] = this.editorConfig.lang === 'yaml'
                            ? yamljs.dump(JSON.parse(preview.content, null, 4))
                            : preview.content
                    })
                    setTimeout(() => {
                        // 这里触发一次 change 为了让初始值也显示
                        this.variableValChange()
                        const editor = this.editorConfig.editors[0]
                        if (editor) {
                            editor.resize(true)
                            editor.gotoLine(0, 0, true)
                        }
                    }, 100)
                }
            } catch (e) {
                console.error(e)
            } finally {
                this.previewLoading = false
            }
        },

        /**
         * 变量修改的回调事件
         */
        variableValChange () {
            if (!this.previewNs.variableList || !this.previewNs.variableList.length) {
                return
            }
            const replacements = {}
            this.previewNs.variableList.forEach(variable => {
                replacements[`{{${variable.key}}}`] = variable.value
                replacements[`"{{${variable.key}}}}"`]
                    = variable.value === null || variable.value === undefined || variable.value === ''
                        ? '""'
                        : isNaN(variable.value) ? `"{{${variable.key}}}}"` : variable.value
            })
            const regex = new RegExp(Object.keys(replacements).map(escape).join('|'), 'g')

            const scrollTopList = []
            const previewList = []
            previewList.splice(0, 0, ...this.previewList)

            const values = []
            previewList.forEach((preview, previewIndex) => {
                scrollTopList.push(this.editorConfig.editors[previewIndex].session.getScrollTop())
                preview.content = preview.originalContent.replace(regex, $0 => replacements[$0])

                values.push(preview.content)
            })
            this.editorConfig.values.splice(0, this.editorConfig.values.length, ...values)
            this.previewList.splice(0, this.previewList.length, ...previewList)

            this.$nextTick(() => {
                this.previewList.forEach((preview, previewIndex) => {
                    this.editorConfig.editors[previewIndex].session.setScrollTop(scrollTopList[previewIndex])
                })
            })
        },

        /**
         * lbservice 下拉框改变事件
         *
         * @param {number} index 索引
         * @param {Object} data 索引对应的对象数据
         */
        async changeLbServiceSelect (index, data) {
            const lbSelectData = Object.assign({}, this.lbSelectData)
            lbSelectData[this.previewNs.name][data.service] = index
            this.lbSelectData = Object.assign({}, lbSelectData)

            if (!this.checkCurNamespacePreview(this.previewNs)) {
                return
            }

            this.preview4Lb(this.previewNs)
        },

        /**
         * 判断当前的 ns 是否可以发送 preview
         * 只有当 当前的 ns 的下拉框全部选择时才可以发送
         *
         * @param {Object} ns 当前 ns
         */
        checkCurNamespacePreview (ns) {
            let canPreview = true
            const keys = Object.keys(this.lbSelectData[ns.name])
            for (let i = 0, len = keys.length; i < len; i++) {
                if (this.lbSelectData[ns.name][keys[i]] === -1) {
                    canPreview = false
                    break
                }
            }
            return canPreview
        },

        /**
         * 带有 lb 的 ns 的 preview
         *
         * @param {Object} ns 当前 ns
         */
        async preview4Lb (ns) {
            const curTplsetVer = this.tplsetVerList.filter(
                tplsetVer => tplsetVer.id === this.tplsetVerId
            )[0] || {}

            const variableInfo = {
                [`${ns.id}`]: {}
            }
            ns.variableList.forEach(variable => {
                variableInfo[`${ns.id}`][variable.key] = variable.value
            })

            const params = {
                projectId: this.projectId,
                namespace: ns.id,
                version_id: curTplsetVer.id,
                show_version_id: curTplsetVer.show_version_id,
                show_version_name: curTplsetVer.show_version_name,
                instance_entity: this.instanceEntity,
                variable_info: variableInfo
            }

            const lbInfo = Object.assign({}, this.lbSelectData[ns.name])
            Object.keys(lbInfo).forEach(key => {
                lbInfo[key] = this.lbServiceListInPageTmp[key][lbInfo[key]]
            })
            params.lb_info = {
                [`${ns.id}`]: lbInfo
            }

            this.previewLoading = true

            try {
                const res = await this.$store.dispatch('configuration/previewNamespace', params)

                const list = []
                const data = res.data || {}
                Object.keys(data).forEach(key => {
                    data[key].forEach(item => {
                        const content = JSON.stringify(item.config, null, 4)
                        list.push({
                            tag: key,
                            name: item.name,
                            content: content,
                            originalContent: content
                        })
                    })
                })

                this.sortTplType(list, 'tag', true)

                this.previewList.splice(0, this.previewList.length, ...list)
                this.previewList.forEach((preview, index) => {
                    this.editorConfig.values[index] = this.editorConfig.lang === 'yaml'
                        ? yamljs.dump(JSON.parse(preview.content, null, 4))
                        : preview.content
                })

                setTimeout(() => {
                    // 这里触发一次 change 为了让初始值也显示
                    this.variableValChange()
                    this.editorConfig.editors[0].resize(true)
                    this.editorConfig.editors[0].gotoLine(0, 0, true)
                }, 100)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.previewLoading = false
            }
        },

        /**
         * 选项卡切换事件
         *
         * @param {string} name tab 名称
         * @param {index} index 索引，标识当前是第几个 tab 的 codeContent
         */
        tabChange (name, index) {
            // this.aceAnnotationErrorMsg = ''
            this.$nextTick(() => {
                const curEditor = this.editorConfig.editors[index]
                curEditor.resize(true)
                const annotations = curEditor.getSession().$annotations
                if (annotations && annotations.length) {
                    curEditor.gotoLine(annotations[0].row + 1, annotations[0].column, true)
                }
            })
        },

        /**
         * ace editor 全屏
         */
        setFullScreen () {
            this.editorConfig.fullScreen = true
        },

        /**
         * 取消全屏
         */
        cancelFullScreen () {
            this.editorConfig.fullScreen = false
        },

        /**
         * 创建模板实例化
         */
        async createInstance () {
            if (this.tplsetVerId === -1) {
                this.bkMessageInstance && this.bkMessageInstance.close()
                this.bkMessageInstance = this.$bkMessage({
                    theme: 'error',
                    message: this.$t('请选择模板集版本')
                })
                return
            }

            if (!this.tplIndex.length) {
                this.bkMessageInstance && this.bkMessageInstance.close()
                this.bkMessageInstance = this.$bkMessage({
                    theme: 'error',
                    message: this.$t('请选择要实例化的模板')
                })
                return
            }

            if (!this.selectedNamespaceList.length) {
                this.bkMessageInstance && this.bkMessageInstance.close()
                this.bkMessageInstance = this.$bkMessage({
                    theme: 'error',
                    message: this.$t('请选择命名空间')
                })
                return
            }

            if (this.invalidNsList.length) {
                this.$bkMessage({
                    theme: 'error',
                    message: this.isEn
                        ? `Namespace [${this.invalidNsList.join(',')}] has no associated LoadBalance, `
                            + `please go to Network -> LoadBalance page association first.`
                        : `命名空间[${this.invalidNsList.join(',')}]没有相关联的LoadBalance，`
                            + `请先到网络 -> LoadBalance页面关联`
                })
                return
            }

            const curTplsetVer = this.tplsetVerList.filter(
                tplsetVer => tplsetVer.id === this.tplsetVerId
            )[0] || {}

            const variableInfo = {}
            let canCreate = true

            let params = {}
            const namespaces = []
            if (this.lbServiceList.length) {
                if (this.checkLbServiceSelect()) {
                    params = {
                        projectId: this.projectId,
                        version_id: curTplsetVer.id,
                        show_version_id: curTplsetVer.show_version_id,
                        show_version_name: curTplsetVer.show_version_name,
                        instance_entity: this.instanceEntity,
                        lb_info: {},
                        is_start: true
                    }

                    this.selectedNamespaceList.forEach(item => {
                        namespaces.push(item.id)
                        variableInfo[`${item.id}`] = {}
                        item.variableList.forEach(variable => {
                            variableInfo[`${item.id}`][variable.key] = variable.value
                        })

                        const lbInfo = Object.assign({}, this.lbSelectData[item.name])
                        Object.keys(lbInfo).forEach(key => {
                            lbInfo[key] = this.lbServiceListInPageTmp[key][lbInfo[key]]
                        })
                        params.lb_info[item.id] = lbInfo
                    })

                    params.namespaces = namespaces.join(',')
                    params.variable_info = variableInfo
                } else {
                    canCreate = false
                }
            } else {
                this.selectedNamespaceList.forEach(item => {
                    namespaces.push(item.id)
                    variableInfo[`${item.id}`] = {}
                    item.variableList.forEach(variable => {
                        variableInfo[`${item.id}`][variable.key] = variable.value
                    })
                })
                params = {
                    projectId: this.projectId,
                    namespaces: namespaces.join(','),
                    version_id: curTplsetVer.id,
                    show_version_id: curTplsetVer.show_version_id,
                    show_version_name: curTplsetVer.show_version_name,
                    instance_entity: this.instanceEntity,
                    variable_info: variableInfo,
                    is_start: true
                }
            }

            if (!canCreate) {
                return
            }

            const me = this
            me.$bkInfo({
                title: me.$t('确认创建'),
                content: me.$createElement('p', this.$t('确定要进行创建操作？')),
                async confirmFn () {
                    me.createInstanceLoading = true
                    try {
                        await me.$store.dispatch('configuration/createInstance', params)
                        me.$router.back()
                    } catch (e) {
                        me.bkMessageInstance && me.bkMessageInstance.close()
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            delay: 20000,
                            hasCloseIcon: true,
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.createInstanceLoading = false
                    }
                }
            })
        },

        /**
         * 检测所有 ns 的 lbservice 下拉框是否都选择了，都选了以后才能 create
         */
        checkLbServiceSelect () {
            const keys = Object.keys(this.lbSelectData)
            const len = keys.length
            let outloop = true
            let ret = true
            for (let i = 0; i < len; i++) {
                if (!outloop) {
                    break
                }
                const key = keys[i]
                const serviceObj = this.lbSelectData[key]
                const serviceObjKeyLen = Object.keys(serviceObj).length
                for (let j = 0; j < serviceObjKeyLen; j++) {
                    const k = Object.keys(serviceObj)[j]
                    if (serviceObj[k] === -1) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请选择{key}命名空间的{k}', { key: key, k: k })
                        })
                        outloop = false
                        ret = false
                        break
                    }
                }
            }
            return ret
        },

        /**
         * 去命名空间页面
         */
        goNamespace () {
            this.$router.push({
                name: 'namespace',
                params: {
                    projectId: this.projectId,
                    projectCode: this.projectCode
                }
            })
        },

        hideNamesapceDialog () {
            this.goNamespaceDialogConf.isShow = false
        },

        /**
         * 触发 快速添加命名空间
         *
         * @param {string} paramName paramDesc
         *
         * @return {string} returnDesc
         */
        triggerAddNamespace (index) {
            this.namespaceName = ''
            this.$refs.addNamespaceNode.forEach(vnode => {
                vnode.instance.hide()
            })

            const vnode = this.$refs.addNamespaceNode[index]
            vnode.instance.show()
        },

        /**
         * 快速添加命名空间确认
         *
         * @param {Object} item 当前集群对象
         * @param {number} index 当前集群对象的索引
         */
        async addNamespace (item, index) {
            this.dialogConf.loading = true
            try {
                const clusterId = item.cluster_id
                if (!this.namespaceName.trim()) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请填写命名空间名称')
                    })
                    return
                }

                if (this.namespaceName.length < 2) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('命名空间名称不得小于2个字符')
                    })
                    return
                }

                if (!/^[a-z][a-z0-9-]+$/g.test(this.namespaceName)) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('命名空间名称只能包含小写字母、数字以及连字符(-)，且不能以数字开头')
                    })
                    return
                }

                if (!clusterId) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择所属集群')
                    })
                    return
                }

                const addedRes = await this.$store.dispatch('configuration/addNamespace', {
                    projectId: this.projectId,
                    name: this.namespaceName,
                    cluster_id: clusterId
                })

                const res = await this.$store.dispatch('configuration/getAllNamespaceList', {
                    projectId: this.projectId,
                    group_by: 'cluster_name',
                    perm_can_use: 1
                })

                const resList = res.data
                const resCluster = resList.find(cluster => cluster.name === item.name)
                if (resCluster) {
                    const resNamespaces = resCluster.results
                    const itemNamespaces = item.results
                    resNamespaces.forEach(ns => {
                        const inItemNamespaces = itemNamespaces.find(
                            itemNs => itemNs.id === ns.id && itemNs.name === ns.name
                        )
                        if (inItemNamespaces) {
                            ns.isChoose = inItemNamespaces.isChoose
                            ns.isExist = inItemNamespaces.isExist
                        }
                    })
                    const resClusterIndex = resList.findIndex(cluster => cluster.name === item.name)
                    this.$set(this.candidateNamespaceList, resClusterIndex, Object.assign(resCluster, {
                        isOpen: this.candidateNamespaceList[resClusterIndex].isOpen
                    }))

                    this.selectNamespaceInDialog(index, addedRes.data, 0)
                }
                this.cancelNamespace()
            } catch (e) {
                catchErrorHandler(e, this)
                this.$nextTick(() => {
                    this.$refs.addNamespaceInputNode[index] && this.$refs.addNamespaceInputNode[index].focus()
                })
            } finally {
                this.dialogConf.loading = false
            }
        },

        /**
         * 快速添加命名空间 tooltip 弹出回调函数
         */
        showAddNamespace (index) {
            this.$nextTick(() => {
                this.$refs.addNamespaceInputNode[index] && this.$refs.addNamespaceInputNode[index].focus()
            })
        },

        /**
         * 快速添加命名空间取消
         */
        cancelNamespace () {
            this.$nextTick(() => {
                this.namespaceName = ''
                this.$refs.addNamespaceNode.forEach(vnode => {
                    vnode.instance.hide()
                })
            })
        }
    }
}
