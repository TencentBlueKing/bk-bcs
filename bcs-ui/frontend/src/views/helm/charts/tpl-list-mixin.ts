/* eslint-disable camelcase */

import { ref, computed, watch, reactive } from '@vue/composition-api'

export default {
    props: {
        tplList: {
            type: Array,
            default: () => []
        },
        selectedName: {
            type: String,
            default: ''
        }
    },
    setup (props, ctx) {
        const { $INTERNAL, $store, $route, $i18n, $bkMessage, $bkInfo } = ctx.root
        const isVersionLoading = ref(false)
        const isVersionDetailLoading = ref(false)
        const isReleaseLoading = ref(false)
        const isVersionDeleting = ref(false)
        const isTplSynLoading = ref(false)
        const deleteVersionTimer = ref()
        const downloadDialog = reactive({
            isShow: false,
            downloadVersion: '',
            chartName: '',
            versions: [],
            chartId: '',
            downloadVersionId: ''
        })

        const delTemplateDialogConf = reactive({
            isShow: false,
            width: 650,
            title: '',
            closeIcon: false,
            template: {},
            releases: [],
            canDeleted: false,
            name: ''
        })

        const delInstanceDialogConf = reactive({
            isShow: false,
            width: 550,
            title: $i18n.t('删除'),
            closeIcon: false,
            template: {},
            versions: [],
            releases: [],
            versionIds: []
        })

        const projectId = computed(() => {
            return $route.params.projectId
        })
        const projectCode = computed(() => {
            return $store.state.curProjectCode
        })
        const username = computed(() => {
            return $store.state.user.username
        })

        watch(() => delInstanceDialogConf.versionIds, (newVal) => {
            if (delInstanceDialogConf.isShow) {
                delInstanceDialogConf.releases = []
                if (deleteVersionTimer.value) {
                    clearTimeout(deleteVersionTimer.value)
                    deleteVersionTimer.value = null
                }
                deleteVersionTimer.value = setTimeout(getReleaseByVersion, 300)
            }
        })

        /**
         * 简单判断是否为图片
         * @param  {string} img 图片url
         * @return {Boolean} true/false
         */
        const isImage = (img) => {
            if (!img) {
                return false
            }
            if (img.startsWith('http://') || img.startsWith('https://') || img.startsWith('data:image/')) {
                return true
            }
            return false
        }
        
        /**
         * 下载版本-打开下载版本弹框
         * @param template {object} 当前模板集对象
         */
        const handleDownloadChart = async (template) => {
            downloadDialog.downloadVersion = ''
            downloadDialog.versions = []
            downloadDialog.chartName = template.name
            downloadDialog.chartId = template.id
            downloadDialog.isShow = true
            downloadDialog.downloadVersionId = ''
            await getTplVersionsList(template, 'down')
        }

        /**
         * 选择版本
         * @param version 选中版本数据
         */
        const handleSelectVersion = (version) => {
            const curVersionData = downloadDialog.versions.find(item => item['version'] === version)
            if (curVersionData) downloadDialog.downloadVersionId = curVersionData['id']
        }

        /**
         * 确认下载
         */
        const handleComfirmDownload = async () => {
            const url = await getChartVersionDetail(downloadDialog)
            const a = document.createElement('a')
            a.href = url as string
            a.click()
            handleCancelDownload()
        }

        /**
         * 取消下载
         */
        const handleCancelDownload = () => {
            downloadDialog.isShow = false
            downloadDialog.versions = []
            downloadDialog.chartName = ''
            downloadDialog.downloadVersion = ''
            downloadDialog.downloadVersionId = ''
        }

        const getChartVersionDetail = async (payload) => {
            const { chartId, chartName, downloadVersion, downloadVersionId } = payload
            const selectedName = props.selectedName.value
            isVersionDetailLoading.value = true

            let url = ''
            const fnPath = $INTERNAL ? 'helm/getChartVersionDetail' : 'helm/getChartByVersion'
            const isPublic = $INTERNAL ? selectedName === 'publicRepo' : undefined

            const res = await $store.dispatch(fnPath, {
                projectId: projectId.value,
                chartId: $INTERNAL ? chartName : chartId,
                version: $INTERNAL ? downloadVersion : downloadVersionId,
                isPublic
            }).catch(() => false)

            isVersionDetailLoading.value = false

            if (!res) return
            const data = res.data || {}
            const urls = (data.data || {}).urls || []
            url = urls[0] || ''
            return url
        }
        const handleRemoveChart = async (template) => {
            // 先检测当前Chart是否有release
            const res = await $store.dispatch('helm/getExistReleases', {
                projectId: projectId.value,
                chartName: template.name
            })

            // 如果没有release，可删除
            if (!res.data.length) {
                delTemplateDialogConf.canDeleted = true
                delTemplateDialogConf.isShow = true
                delTemplateDialogConf.title = $i18n.t('删除')
                delTemplateDialogConf.name = template.name
            } else {
                delTemplateDialogConf.canDeleted = false
                delTemplateDialogConf.isShow = true
                delTemplateDialogConf.title = template.name
                delTemplateDialogConf.template = Object.assign({}, template)
                delTemplateDialogConf.releases = res.data
            }
        }

        /**
         * 确认删除Chart
         * @param {Object} template 当前模板集对象
         */
        const handleDeleteTemplate = async () => {
            await $store.dispatch('helm/removeTemplate', {
                projectId: projectId.value,
                chartName: delTemplateDialogConf.name
            })

            $bkMessage({
                theme: 'success',
                message: $i18n.t('删除成功')
            })
            ctx.emit('fetchList')
        }

        /**
         * 取消删除template - 隐藏弹框
         */
        const delTemplateCancel = () => {
            delTemplateDialogConf.isShow = false
            delTemplateDialogConf.title = ''
            delTemplateDialogConf.template = Object.assign({}, {})
            delTemplateDialogConf.releases = []
        }

        const showChooseDialog = (template) => {
            // 之前没选择过，那么展开第一个
            delInstanceDialogConf.isShow = true
            delInstanceDialogConf.template = Object.assign({}, template)
            delInstanceDialogConf.versions = []
            delInstanceDialogConf.releases = []
            delInstanceDialogConf.versionIds = []
            getTplVersionsList(template, 'del')
        }

        const getTplVersionsList = async (template, type) => {
            isVersionLoading.value = true
            const name = template.name
            const res = await $store.dispatch('helm/getTplVersions', {
                projectId: projectCode.value,
                repository: projectCode.value,
                name,
                params: {
                    page: 1,
                    size: 1500,
                    operator: username.value
                }
            })
            isVersionLoading.value = false
            if (!res) return
            if (type === 'del') {
                delInstanceDialogConf.versions = res.data.data
                delInstanceDialogConf.releases = []
            } else {
                downloadDialog.versions = res.data.results || []
            }
        }

        /**
         * 删除命名空间弹层确认
         */
        const confirmDelVersion = async () => {
            if (!delInstanceDialogConf.versionIds.length) {
                $bkMessage({
                    theme: 'error',
                    message: $i18n.t('请选择Chart版本')
                })
                return
            }
            const versions = delInstanceDialogConf.versions.filter(item => delInstanceDialogConf.versionIds.includes(item['version']))
            $bkInfo({
                title: $i18n.t('确认删除以下版本'),
                content: versions.map(item => item['version']).join(', '),
                async confirmFn () {
                    deleteTemplateVersion()
                }
            })
        }

        /**
         * 确认删除版本
         * @param {Object} template 当前模板集对象
         */
        const deleteTemplateVersion = async () => {
            const versions = delInstanceDialogConf.versions.filter(item => delInstanceDialogConf.versionIds.includes(item['version']))
            isVersionDeleting.value = true
            const res = await $store.dispatch('helm/removeTemplate', {
                projectId: projectId.value,
                chartName: delInstanceDialogConf.template['name'],
                versions: versions.map(item => item['version'])
            }).catch(() => false)
            if (!res) return

            $bkMessage({
                theme: 'success',
                message: $i18n.t('删除成功')
            })

            ctx.emit('fetchList')
            cancelDelVersion()
        }

        const cancelDelVersion = () => {
            delInstanceDialogConf.isShow = false
            delInstanceDialogConf.template = {}
            delInstanceDialogConf.versions = []
            delInstanceDialogConf.releases = []
            delInstanceDialogConf.versionIds = []
        }

        const getReleaseByVersion = async () => {
            const versions = delInstanceDialogConf.versions.filter(item => delInstanceDialogConf.versionIds.includes(item['id']))
            isReleaseLoading.value = true
            const res = await $store.dispatch('helm/getExistReleases', {
                projectId: projectId.value,
                chartName: delInstanceDialogConf.template['name'],
                versions: versions.map(item => item['version'])
            }).catch(() => false)
            isReleaseLoading.value = false
            if (!res) return

            delInstanceDialogConf.releases = res.data
        }

        /**
         * 同步仓库
         */
        const syncHelmTpl = async () => {
            if (isTplSynLoading.value) {
                return false
            }

            isTplSynLoading.value = true
            await $store.dispatch('helm/syncHelmTpl', { projectId: projectId.value })

            $bkMessage({
                theme: 'success',
                message: $i18n.t('同步成功')
            })
            ctx.emit('fetchList')

            isTplSynLoading.value = false
        }

        return {
            downloadDialog,
            isVersionDetailLoading,
            delTemplateDialogConf,
            delInstanceDialogConf,
            isVersionLoading,
            isImage,
            handleDownloadChart,
            handleComfirmDownload,
            handleSelectVersion,
            handleCancelDownload,
            handleRemoveChart,
            delTemplateCancel,
            handleDeleteTemplate,
            showChooseDialog,
            cancelDelVersion,
            syncHelmTpl,
            confirmDelVersion
        }
    }
}
