<!-- eslint-disable max-len -->
<template>
  <div class="ring-wrapper" :class="extCls" :style="{ width: size + 'px', height: size + 'px' }">
    <svg :height="size" :width="size" :class="[text === 'hover' ? 'show-text' : '']">
      <circle :cx="size / 2" :cy="size / 2" :r="size / 2 - strokeWidth" fill="none" :stroke="strokeColor" :stroke-width="strokeWidth" stroke-linecap="round" />
      <circle :cx="size / 2" :cy="size / 2" :r="size / 2 - fillWidth" fill="none" :stroke="fillColor" stroke-linecap="round" :stroke-width="fillWidth" stroke-dasharray="0,10000" />
      <text :style="textStyle" :fill="(textStyle || {}).color" :class="text !== 'always' ? 'hide' : ''" x="50%" y="50%" dy=".3em" text-anchor="middle">
        {{roundDecimalNum(realPercent, 2)}}%
      </text>
    </svg>
    <slot name="text"></slot>
  </div>
</template>

<script>
/**
     * Ring 组件
     * Usage: <Ring :percent="percent" :size="40" :text="'hover'" :extCls="'biz-ring'"/>
     *
     * setTimeout(() => {
     *     this.percent = 50
     *     setTimeout(() => {
     *         this.percent = 75
     *         setTimeout(() => {
     *             this.percent = 30
     *         }, 1500)
     *     }, 1500)
     * }, 1500)
     */

export default {
  name: 'RingLoading',
  props: {
    // 圆环百分比数字
    percent: {
      type: [Number, String],
      default: 0,
    },
    // 圆环大小
    size: {
      type: Number,
      default: 100,
    },
    // 外圈宽度，外圈指 100% 的圈
    strokeWidth: {
      type: Number,
      default: 5,
    },
    strokeColor: {
      type: String,
      default: '#ebf0f5',
    },
    // 内圈宽度，内圈指根据 percent 计算的那个圈
    fillWidth: {
      type: Number,
      default: 5,
    },
    fillColor: {
      type: String,
      default: '#3a84ff',
    },
    // 显示 ring-text 的方式
    text: {
      validator(value) {
        return [
          'none', // 不显示 text
          'always', // 总是显示 text
          'hover', // hover 时显示 text
        ].indexOf(value) > -1;
      },
      default: 'always',
    },
    textStyle: {
      type: Object,
      default: () => ({
        fontSize: '12px',
        color: '#3a84ff',
      }),
    },
    extCls: {
      type: String,
    },
    percentChangeHandler: {
      type: Function,
      default: () => {},
    },
  },
  data() {
    return {
      queue: [],
      timer: null,
      realPercent: this.percent,
      node: null,
      circleLen: 0,
    };
  },
  watch: {
    percent(val) {
      this.updateProcess(val);
      this.queue.push(val);
      if (this.timer) {
        return;
      }
      while (1) {
        const curTarget = this.queue.pop();
        if (!curTarget) {
          return;
        }
        let curNum = this.realPercent;
        let change;
        this.timer = setInterval(() => {
          if (parseInt(parseFloat(curNum).toFixed(0), 10) !== parseInt(parseFloat(curTarget).toFixed(0), 10)) {
            if (curNum < curTarget) {
              change = ++curNum;
            } else if (curNum > curTarget) {
              change = --curNum;
            }
            this.percentChangeHandler(change);
            this.updateProcess(change);
          } else {
            clearInterval(this.timer);
            this.timer = null;
          }
        }, 30);
      }
    },
  },
  mounted() {
    const r = this.size / 2 - this.fillWidth;
    this.circleLen = Math.floor(2 * Math.PI * r);
    this.node = this.$el.querySelectorAll('.ring-wrapper circle')[1];
    this.updateProcess(this.percent);
  },
  methods: {
    /**
             * 更新圆环百分比
             *
             * @param {number} percent 百分比数字
             */
    updateProcess(percent) {
      this.node.setAttribute('stroke-dasharray', `${this.circleLen * percent / 100},10000`);
      this.realPercent = percent;
    },

    roundDecimalNum(value, n) {
      return Math.round(value * Math.pow(10, n)) / Math.pow(10, n);
    },
  },
};
</script>

<style scoped lang="postcss">
    .ring-wrapper {
        position: relative;
        display: inline-block;
        margin: 0;
        padding: 0;
        font-size: 0;

        .show-text {
            &:hover {
                text {
                    display: block;
                }
            }
        }

        circle {
            -webkit-transform-origin: center;
            transform-origin: center;
            -webkit-transform: rotate(-90deg);
            transform: rotate(-90deg);
        }

        text {
            cursor: default;
            letter-spacing: -0.3px;

            &.hide {
                display: none;
            }
        }
    }

</style>
