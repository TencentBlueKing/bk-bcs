import { computed, ref } from '@vue/composition-api'
/**
 * 获取项目文档配置信息
 * @returns
 */
export default function useConfig () {
    // 当前版本的文档链接信息
    const PROJECT_CONFIG = ref(window.BCS_CONFIG)
    // 是否是内部版
    const $INTERNAL = computed(() => {
        return !['ce', 'ee'].includes(window.REGION)
    })
    return {
        PROJECT_CONFIG,
        $INTERNAL
    }
}
