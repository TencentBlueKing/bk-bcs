# CronJob

> CronJob（CJ）创建基于时隔重复调度的 Jobs。

## 什么是 CronJob ？

CronJob 类似于 Linux 中的 crontab 与 kubernetes 中 Job 的结合体。

CronJob 对象就像 crontab 文件中的一行。它用 Cron 格式进行编写，并周期性地在给定的调度时间创建并执行 Job。

## 使用 CronJob

> 注意：
> 所有 CronJob 的 schedule 时间都是基于 `kube-controller-manager` 的时区，即 CronJob 不支持单独设置时区。
> 如果你的控制平面在 Pod 或是裸容器中运行了 kube-controller-manager，那么为该容器所设置的时区将会决定 CronJob 的控制器所使用的时区。

CronJob 一般用于执行定时任务，如数据同步，备份，报告生成以及定期数据上报等。这些任务中的每一个都应该配置为周期性重复的（例如：每天/每周/每月一次）；用户可以定义任务开始执行的时间间隔。

### Cron 时间表语法

`* * * * *` 依次对应：
- 分钟 (0 - 59)
- 小时 (0 - 23)
- 月的某天 (1 - 31)
- 月份 (1 - 12)
- 周的某天 (0 - 6) （周日到周一；在某些系统上，7 也是星期日）

| 输入                   | Cron 表达式   | 描述                         |
| ---------------------- | ------------- | ---------------------------- |
| @yearly (or @annually) | 0 0 1 1 \*    | 每年 1 月 1 日的午夜运行一次 |
| @monthly               | 0 0 1 \* \*   | 每月第一天的午夜运行一次     |
| @weekly                | 0 0 \* \* 0   | 每周的周日午夜运行一次       |
| @daily (or @midnight)  | 0 0 \* \* \*  | 每天午夜运行一次             |
| @hourly                | 0 \* \* \* \* | 每小时的开始一次             |

例如，下面这行指出必须在每个星期五的午夜以及每个月 13 号的午夜开始任务：

`0 0 13 \* 5`

要生成 CronJob 时间表表达式，你还可以使用 [crontab.guru](https://crontab.guru/) 之类的 Web 工具。

## 参考资料

1. [Kubernetes / 工作负载 / CronJobs](https://kubernetes.io/zh/docs/concepts/workloads/controllers/cron-jobs/)
2. [kubernetes CronJob 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#cronjob-v1beta1-batch)
