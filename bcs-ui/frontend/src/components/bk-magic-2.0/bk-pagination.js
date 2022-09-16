export default {
    name: 'bk-pagination',
    props: {
        size: {
            type: String,
            default: 'small'
        },
        location: {
            type: String,
            default: 'left'
        },
        align: {
            type: String,
            default: 'right'
        }
    },
    render (h) {
        const slots = Object.keys(this.$slots)
            .reduce((arr, key) => arr.concat(this.$slots[key]), [])
            .map(vnode => {
                vnode.context = this._self
                return vnode
            })

        return h('bcs-pagination', {
            on: this.$listeners, // 透传事件
            props: this.$props, // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    }
}
