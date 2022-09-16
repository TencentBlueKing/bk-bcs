<template>
    <div :class="['bk-number', { 'focus': isFocus, 'disabled': disabled, 'is-error': isError }]" :style="exStyle">
        <div class="bk-number-content " :class="[{ 'bk-number-larger': size === 'large','bk-number-small': size === 'small' }]">
            <input
                type="text"
                class="bk-number-input"
                autocomplete="off"
                id="bk-number-input"
                :disabled="disabled"
                :placeholder="placeholder"
                @keydown.up.prevent="add"
                @keydown.down.prevent="minus"
                @focus="focus"
                @blur="blur"
                @input="debounceHandleInput"
                :value="currentValue">
            <div class="bk-number-icon-content" v-if="!hideOperation">
                <div :class="['bk-number-icon-top', { 'btn-disabled': isMax }]" @click="add">
                    <i class="bcs-icon bcs-icon-angle-up"></i>
                </div>
                <div :class="['bk-number-icon-lower', { 'btn-disabled': isMin }]" @click="minus">
                    <i class="bcs-icon bcs-icon-angle-down"></i>
                </div>
            </div>
        </div>
    </div>
</template>
<script>
    import { debounce } from 'throttle-debounce'

    export default {
        name: 'bk-number-input',
        props: {
            value: {
                type: [Number, String],
                default: 0
            },
            hideOperation: {
                type: Boolean,
                default: false
            },
            type: {
                type: String,
                default: 'int'
            },
            exStyle: {
                type: Object,
                default () {
                    return {}
                }
            },
            placeholder: {
                type: String,
                default: ''
            },
            disabled: {
                type: [String, Boolean],
                default: false
            },
            min: {
                type: Number,
                default: Number.NEGATIVE_INFINITY
            },
            max: {
                type: Number,
                default: Number.POSITIVE_INFINITY
            },
            steps: {
                type: Number,
                default: 1
            },
            size: {
                type: String,
                default: 'large',
                validator (value) {
                    return [
                        'large',
                        'small'
                    ].indexOf(value) > -1
                }
            },
            debounceTimer: {
                type: Number,
                default: 100
            }
        },
        data () {
            return {
                isMax: false,
                isMin: false,
                currentValue: '',
                isFocus: false,
                maxNumber: this.max,
                minNumber: this.min,
                isError: false
            }
        },
        watch: {
            min () {
                this.minNumber = this.min
            },
            max () {
                this.maxNumber = this.max
            },
            value: {
                immediate: true,
                handler (value) {
                    value = value + ''
                    if (value === '') {
                        this.currentValue = value
                        return
                    }

                    // let newVal = parseInt(value)

                    // if (this.type === 'decimals') {
                    //     newVal = Number(value)
                    // }

                    this.currentValue = value
                }
            }
        },
        created () {
            this.debounceHandleInput = debounce(this.debounceTimer, event => {
                const value = event.target.value
                this.inputHandler(value, event.target)
            })
        },
        methods: {
            focus (event) {
                this.isFocus = true
                this.$emit('focus', event)
            },
            blur () {
                this.isFocus = false
                this.$emit('blur', event)
            },
            getPower (val) {
                const valueString = val.toString()
                const dotPosition = valueString.indexOf('.')

                let power = 0
                if (dotPosition > -1) {
                    power = valueString.length - dotPosition - 1
                }
                return Math.pow(10, power)
            },
            checkMinMax (val) {
                if (val <= this.minNumber) {
                    val = this.minNumber
                    this.isMin = true
                } else {
                    this.isMin = false
                }
                if (val >= this.maxNumber) {
                    val = this.maxNumber
                    this.isMax = true
                } else {
                    this.isMax = false
                }
                return val
            },
            inputHandler (value, target) {
                if (value === '') {
                    this.$emit('update:value', value)
                    this.$emit('change', value)
                    this.currentValue = value
                    target && (target.value = value)
                    return
                }
                if (value !== '' && value.indexOf('.') === (value.length - 1)) {
                    return
                }

                if (value !== '' && value.indexOf('.') > -1 && Number(value) === 0) {
                    return
                }
                // if (value !== '' && value.indexOf('-') === (value.length - 1)) {
                //     return
                // }

                let newVal = parseInt(value)

                if (this.type === 'decimals') {
                    newVal = Number(value)
                }

                if (!isNaN(newVal)) {
                    this.setCurrentValue(newVal, target)
                } else {
                    target.value = this.currentValue
                }
            },
            setCurrentValue (val, target) {
                // const oldVal = this.currentValue.toFixed(2)
                val = this.checkMinMax(val)
                this.$emit('update:value', val)
                this.$emit('change', val)
                this.currentValue = val
                target && (target.value = val)
            },
            add () {
                if (this.disabled) return
                const value = this.value || 0
                if (typeof value !== 'number') return this.currentValue
                const power = this.getPower(value)
                const newVal = (power * value + power * this.steps) / power
                if (newVal > this.max) return
                this.setCurrentValue(newVal)
            },
            minus () {
                if (this.disabled) return
                const value = this.value || 0
                if (typeof value !== 'number') return this.currentValue
                const power = this.getPower(value)
                const newVal = parseInt(power * value - power * this.steps) / power
                if (newVal < this.min) return
                this.setCurrentValue(newVal)
            }
        }
    }
</script>

<style scoped lang="postcss">
    @import './index.css';
</style>
