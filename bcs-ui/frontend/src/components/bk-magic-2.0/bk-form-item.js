export default {
    name: 'bk-form-item',
    props: {
        
    },
    mounted () {
        // to do
        const bcsFormItem = this.$refs.bcsFormItem
        bcsFormItem.dispatch('bcs-form', 'form-item-add', bcsFormItem)

        this.$off('form-blur', this.handlerBlur)
        this.$off('form-change', this.handlerChange)

        this.$on('form-blur', this.handlerBlur)
        this.$on('form-focus', this.handlerFocus)
        this.$on('form-change', this.handlerChange)

        const attrs = Object.assign({}, this.$props, this.$attrs)
        if (attrs.autoCheck) {
            this.validate()
        }
    },
    beforeDestroy () {
        this.$refs.bcsFormItem.dispatch('bcs-form', 'form-item-delete', this.$refs.bcsFormItem)
    },
    methods: {
        handlerBlur () {
            this.$refs.bcsFormItem.handlerBlur()
        },
        handlerChange () {
            this.$refs.bcsFormItem.handlerChange()
        },
        handlerFocus () {
            this.$refs.bcsFormItem.clearValidator()
        },
        validate () {
            this.$refs.bcsFormItem.validate()
        },
        clearError () {
            this.$refs.bcsFormItem.clearError()
        }
    },
    render (h) {
        const slots = Object.keys(this.$slots)
            .reduce((arr, key) => arr.concat(this.$slots[key]), [])
            .map(vnode => {
                vnode.context = this._self
                return vnode
            })

        return h('bcs-form-item', {
            ref: 'bcsFormItem',
            on: this.$listeners, // 透传事件
            props: Object.assign({}, this.$props, this.$attrs), // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    }
}
