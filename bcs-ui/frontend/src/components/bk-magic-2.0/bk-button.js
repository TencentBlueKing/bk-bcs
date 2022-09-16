export default {
    name: 'bk-button',
    props: {
        type: {
            type: String,
            default: 'default',
            validator (value) {
                const types = value.split(' ') // ['submit', 'success', 'reset']
                const buttons = [
                    'button',
                    'submit',
                    'reset'
                ]
                const thenme = [
                    'default',
                    'info',
                    'primary',
                    'warning',
                    'success',
                    'danger'
                ]
                let valid = true
                types.forEach(type => {
                    if (buttons.indexOf(type) === -1 && thenme.indexOf(type) === -1) {
                        valid = false
                    }
                })
                return valid
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

        if (this.$props.type) {
            this.$props.theme = this.$props.type
        }

        return h('bcs-button', {
            on: this.$listeners, // 透传事件
            props: Object.assign({}, this.$props, this.$attrs), // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    }
}
