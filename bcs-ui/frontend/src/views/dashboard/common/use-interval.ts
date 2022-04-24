import { Ref, ref, onUnmounted, getCurrentInstance, onDeactivated } from '@vue/composition-api'

export type Fn = () => void

export interface ITimeoutFnResult {
    start: Fn;
    stop: Fn;
    timer: Ref<number | null>;
    isPending: Ref<boolean>;
}

/**
 * 轮询
 * @param cb 回调
 * @param interval 轮询周期
 * @param immediate 立即执行
 */
export default function useIntervalFn (
    cb: (...args: unknown[]) => Promise<any>,
    interval: number = 5000,
    immediate = false
): ITimeoutFnResult {
    const isPending = ref(false)
    const flag = ref(false)

    const timer = ref<number | null>(null)

    function clear () {
        if (timer.value) {
            clearTimeout(timer.value)
            timer.value = null
        }
    }

    function stop () {
        isPending.value = false
        flag.value = false
        clear()
    }

    function start (...args: unknown[]) {
        clear()
        if (!interval) return

        flag.value = true
        async function timerFn () {
            // 上一个接口未执行完，不执行本次轮询
            if (isPending.value) return

            isPending.value = true
            // eslint-disable-next-line standard/no-callback-literal
            await cb(...args)
            isPending.value = false
            if (flag.value) {
                timer.value = setTimeout(timerFn, interval) as unknown as number
            }
        }
        setTimeout(() => timerFn(), immediate ? 0 : interval)
    }

    if (getCurrentInstance()) {
        onUnmounted(stop)
        onDeactivated(stop)
    }

    return {
        isPending,
        timer,
        start,
        stop
    }
}
