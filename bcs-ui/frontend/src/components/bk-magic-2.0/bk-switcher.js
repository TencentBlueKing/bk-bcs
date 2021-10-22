export default {
    name: 'bk-switcher',
    props: {
        selected: {
            type: Boolean,
            default: true
        },
        value: {
            type: Boolean,
            default: false
        }
    },
    
    render (h) {
        const slots = Object.keys(this.$slots)
            .reduce((arr, key) => arr.concat(this.$slots[key]), [])
            .map(vnode => {
                vnode.context = this._self
                return vnode
            })

        if (this.$props.selected) {
            this.$props.value = this.$props.selected
        }

        return h('bcs-switcher', {
            on: this.$listeners,
            props: this.$props,
            scopedSlots: this.$scopedSlots,
            attrs: this.$attrs
        }, slots)
    }
}
