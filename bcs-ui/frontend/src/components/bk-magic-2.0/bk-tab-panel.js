export default {
    name: 'bk-tab-panel',
    props: {
        name: {
            type: [String, Number],
            required: true
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
        this.$props.label = this.$props.title
        
        return h('bcs-tab-panel', {
            on: this.$listeners, // 透传事件
            props: this.$props, // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    }
}
