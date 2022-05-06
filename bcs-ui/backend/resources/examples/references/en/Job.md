# Job

> A Job will create one or more Pods and execute a one-time task until the target number of successes ends.

## What is a Job?

The Job creates one or more Pods and will continue to retry the execution of the Pods until the specified number of Pods terminate successfully.

As Pods complete successfully, the Job keeps track of the number of Pods that completed successfully. When the number reaches the specified success threshold, the task (ie Job) ends.

Deleting a Job will clear all Pods created. Suspending a Job will delete all active Pods of the Job until the Job is resumed again.

Job supports parallelism, and you can run multiple Pods at the same time by specifying the degree of parallelism.

## Using Job

Common job usage scenarios are:

- Execute one-time tasks, such as DB Migrate, data sampling, single operation tasks, etc.
- Execute parallel tasks, such as distributed parallel computing, etc.

## References

1. [Kubernetes / Workloads / Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
2. [kubernetes Job field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#job-v1-batch)
