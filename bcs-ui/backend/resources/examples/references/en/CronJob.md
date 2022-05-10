# CronJob

> CronJob (CJ) creates Jobs that are scheduled based on time interval recurrence.

## What is a CronJob?

CronJob is similar to the combination of crontab in Linux and Job in kubernetes.

A CronJob object is like a line in crontab file. It is written in Cron format and periodically creates and executes a Job at a given scheduled time.

## Using CronJob

> Note:
> All CronJob's schedule time is based on the time zone of `kube-controller-manager`, that is, CronJob does not support setting the time zone separately.
> If your control plane runs kube-controller-manager in a Pod or a bare container, the time zone set for that container will determine the time zone used by the CronJob's controller.

CronJob is generally used to perform scheduled tasks, such as data synchronization, backup, report generation, and periodic data reporting. Each of these tasks should be configured to recur periodically (eg: daily/weekly/monthly); the user can define the time interval at which the task starts executing.

### Cron Timetable Syntax

`* * * * *` corresponds to:

- minutes (0 - 59)
- Hours (0 - 23)
- day of the month (1 - 31)
- Month (1 - 12)
- Day of the week (0 - 6) (Sunday to Monday; on some systems, 7 is also Sunday)

| Input                   | Cron Expression   | Description                         |
| ---------------------- | ------------- | ---------------------------- |
| @yearly (or @annually) | 0 0 1 1 \*    | run every year at midnight on January 1st |
| @monthly               | 0 0 1 \* \*   | Runs at midnight on the first day of every month     |
| @weekly                | 0 0 \* \* 0   | run weekly on Sunday at midnight       |
| @daily (or @midnight)  | 0 0 \* \* \*  | run once a day at midnight             |
| @hourly                | 0 \* \* \* \* | run at the start of every hour             |

For example, the following line states that the task must start at midnight every Friday and at midnight on the 13th of every month:

`0 0 13 \* 5`

To generate CronJob timetable expressions, you can also use web tools like [crontab.guru](https://crontab.guru/).

## Reference

1. [Kubernetes / Workload / CronJobs](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
2. [kubernetes CronJob field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#cronjob-v1beta1-batch)
