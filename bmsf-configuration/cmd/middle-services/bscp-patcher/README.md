Patcher模块开发指南
===================

## Patch补丁说明

### 增加新Patch补丁流程

- 1.在patchs目录下创建新Patch补丁目录, 目录命名规范格式为{tag}-{datetime}, 如v2.0.1-202011201517;
- 2.在新创建的Patch补丁目录下实现Patch，需实现GetName、NeedToSkip、PatchFunc方法, 即实现PatchInterface, 详细参见modules下hpm内定义;
- 3.在patchs的init中基于register函数注册新的Patch补丁;

`说明`:

    1.需要注意产品类型，在NeedToSkip内进行执行与否的逻辑判断！

## Crontab说明

### 增加新Crontab任务流程

- 1.在crons目录下创建新Crontab任务目录，目录命名规范格式为{group-name}.{job-name}, 如group-default.job-echo；
- 2.在新创建的Crontab任务目录下实现Job，需要实现GetName、NeedRun、BeforeRun、Run、AfterRun方法， 即实现Job interface, 详细参见modules下ctm内定义;
- 3.在crons的init中基于register函数注册新的Crontab任务;

`说明`:

    1.需要注意任务间隔和执行耗时，在当前任务还没执行完之前下一轮执行可能已经触达，
    需要进行重入逻辑的判定(可基于ctm内Controller提供的方法实现)
    2.任何Job任务不能阻塞等待上一次执行结束，若重入判断出上一次仍未执行结束，应立即返回当前上下文协程！

## 其他

    Patch补丁和Crontab任务框架整体上一致，比较简洁实现。但两者本质上也有差异，如Patch补丁的Interface函数带有DB等相关参数，
    实现新的Patch补丁只需要填写响应的业务逻辑即可。但Crontab任务中的Interface都是无参数的，之所以没有固定参数是因为Patch
    的场景都是针对DB数据层面操作，但Crontab任务却不仅仅是DB，还可能是调用其他模块接口、定时发送通知等外部调用操作，故此在
    Crontab的Interface函数中没有固定的参数，需要什么都可以在CTM内直接GetXXXX拿到，CTM作为Crontab的管理者，提供丰富的工具支持。
