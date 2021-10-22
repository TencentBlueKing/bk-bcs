export default {
    name: 'bk-sideslider',
    props: {
        transfer: {
            type: Boolean,
            default: true
        },
        title: {
            type: String,
            default: ''
        }
    },
    
    render (h) {
        const slots = Object.keys(this.$slots)
            .reduce((arr, key) => arr.concat(this.$slots[key]), [])
            .map(vnode => {
                vnode.context = this._self
                return vnode
            })
        this.$props.transfer = true
        return h('bcs-sideslider', {
            on: this.$listeners, // 透传事件
            props: this.$props, // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    }
}
