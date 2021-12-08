// 公共pods日志逻辑
import $store from '@/store'
import { reactive, toRefs, watch, toRef } from '@vue/composition-api'

export default function useLog () {
    // 获取日志容器组件容器列表数据
    const logState = reactive<{
        logShow: boolean;
        logLoading: boolean;
        curPodId: string;
        curNamespace: string;
        defaultContainer: string;
        containerList: any[];
    }>({
        logShow: false,
        logLoading: false,
        curPodId: '',
        curNamespace: '',
        defaultContainer: '',
        containerList: []
    })
    watch(toRef(logState, 'logShow'), (show) => {
        if (!show) {
            logState.curPodId = ''
            logState.curNamespace = ''
            logState.containerList = []
            logState.defaultContainer = ''
        }
    })
    const handleGetContainer = async (podId: string, namespace: string) => {
        const data = await $store.dispatch('dashboard/listContainers', {
            $podId: podId,
            $namespaceId: namespace
        })
        return data
    }
    // 显示操作日志
    const handleShowLog = async (row) => {
        logState.logShow = true
        const { name, namespace } = row.metadata
        logState.curPodId = name
        logState.curNamespace = namespace
        logState.containerList = await handleGetContainer(name, namespace)
        logState.defaultContainer = logState.containerList[0]?.name
    }

    return {
        ...toRefs(logState),
        handleShowLog
    }
}
