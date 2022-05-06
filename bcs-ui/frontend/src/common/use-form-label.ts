import { ref } from "@vue/composition-api"

export default function useFormLabel () {
    const minFormLabelWidth = ref(100)
    const labelWidth = ref(100)
    const initFormLabelWidth = (formRef) => {
        const el = formRef ? formRef.$el : null
        if (!el) return
        let max = 0
        const leftPadding = 28
        const safePadding = 8
        const $labelEleList = el.querySelectorAll('.bk-label')
        $labelEleList.forEach((item) => {
            const spanEle = item.querySelector('span')
            spanEle.style = `${spanEle.style};white-space: nowrap`
            if (spanEle) {
                const { width } = spanEle.getBoundingClientRect()
                max = Math.max(minFormLabelWidth.value, max, width)
            }
        })
        const width = Math.ceil(max + leftPadding + safePadding)
        $labelEleList.forEach((item) => {
            item.style.width = `${width}px`
        })
        el.querySelectorAll('.bk-form-content').forEach((item) => {
            item.style.marginLeft = `${width}px`
        })
        labelWidth.value = width || minFormLabelWidth.value
        return labelWidth.value
    }
    return {
        labelWidth,
        initFormLabelWidth
    }
}
