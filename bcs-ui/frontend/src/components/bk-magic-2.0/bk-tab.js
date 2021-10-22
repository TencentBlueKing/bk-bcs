export default {
    name: 'bk-tab',
    props: {
        activeName: {
            type: [String, Number],
            default: ''
        },
        type: {
            type: String,
            default: 'border-card',
            validator (val) {
                return ['card', 'border-card', 'unborder-card', 'vertical-card'].includes(val)
            }
        }
    },
    
    render (h) {
        const slots = Object.keys(this.$slots)
            .reduce((arr, key) => arr.concat(this.$slots[key]), [])
            .map(vnode => {
                vnode.context = this._self
                return vnode
            })
        if (this.$props.type === 'fill') {
            this.$props.type = 'border-card'
        }
        this.$props.active = this.$props.activeName
        if (this.$listeners && this.$listeners['tab-changed']) {
            this.$listeners['tab-change'] = this.$listeners['tab-changed']
        }

        return h('bcs-tab', {
            on: this.$listeners, // 透传事件
            props: Object.assign({}, this.$props, this.$attrs), // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    }
}
