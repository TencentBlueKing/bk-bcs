// - 监控 Performance 指标
const typeMap = {
  navigation: '导航加载',
  paint: '首次绘制(FP)和首次内容绘制(FCP)',
  'first-input': '首次输入延迟(FID)',
  'largest-contentful-paint': '最大内容绘制(LCP)',
};

export function monitorPerformance() {
  if (!window.PerformanceObserver) return;

  const observer = new PerformanceObserver((list) => {
    for (const entry of list.getEntries()) {
      // 数据上报
      window.BkTrace?.startReported({
        module: 'performance',
        operation: entry.entryType,
        desc: typeMap[entry.entryType],
        performance: entry.toJSON(),
      }, 'performance');
    }
  });

  observer.observe({ entryTypes: Object.keys(typeMap) });
}

// - 监控UI的FPS
const FPS_THRESHOLD = 30;    // 设置FPS的阈值为30
let frameTimes: number[] = [];         // 保存帧间隔时间
let lastFrameTimestamp = performance.now();
let reportTimeout;           // 用于存储 setTimeout 句柄
export function monitorFrame() {
  const updateFrameStats = () => {
    const now = performance.now();
    const elapsed = now - lastFrameTimestamp;
    lastFrameTimestamp = now;
    frameTimes.push(elapsed);
    // 每隔10秒检查一次平均FPS
    if (!reportTimeout) {
      reportTimeout = setTimeout(() => {
        // 计算平均帧间隔时间
        const averageElapsed = frameTimes.reduce((a, b) => a + b, 0) / frameTimes.length;
        const fps = 1000 / averageElapsed;

        // 上报低FPS情况
        if (fps < FPS_THRESHOLD) {
          // 数据上报
          window.BkTrace?.startReported({
            module: 'performance',
            operation: 'FPS',
            desc: `FPS低于${FPS_THRESHOLD}`,
          }, 'performance');
        }

        // 清空帧时间列表，重置计时器
        frameTimes = [];
        clearTimeout(reportTimeout);
        reportTimeout = null;
      }, 10000);
    }
    requestAnimationFrame(updateFrameStats);
  };
  requestAnimationFrame(updateFrameStats);
}

monitorPerformance();
monitorFrame();
