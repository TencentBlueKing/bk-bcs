E2E测试
======

# 编译

执行`make`编译e2e测试工具,

```json
make

Building bk-bscp-e2e-testing...
go test -c ./... -o bk-bscp-e2e-testing
Build bk-bscp-e2e-testing success!
```

# 使用

`--help`查看可用的测试选项,

```json
./bk-bscp-e2e-testing --help

Usage of ./bk-bscp-e2e-testing:
    -test.bench regexp
        run only benchmarks matching regexp
    -test.benchmem
        print memory allocations for benchmarks
    -test.benchtime d
        run each benchmark for duration d (default 1s)
    -test.blockprofile file
        write a goroutine blocking profile to file
    -test.blockprofilerate rate
        set blocking profile rate (see runtime.SetBlockProfileRate) (default 1)
    -test.count n
        run tests and benchmarks n times (default 1)
    -test.coverprofile file
        write a coverage profile to file
    -test.cpu list
        comma-separated list of cpu counts to run each test with
    -test.cpuprofile file
        write a cpu profile to file
    -test.failfast
        do not start new tests after the first test failure
    -test.list regexp
        list tests, examples, and benchmarks matching regexp then exit
    -test.memprofile file
        write an allocation profile to file
    -test.memprofilerate rate
        set memory allocation profiling rate (see runtime.MemProfileRate)
    -test.mutexprofile string
        write a mutex contention profile to the named file after execution
    -test.mutexprofilefraction int
        if >= 0, calls runtime.SetMutexProfileFraction() (default 1)
    -test.outputdir dir
        write profiles to dir
    -test.parallel n
        run at most n tests in parallel (default 8)
    -test.run regexp
        run only tests and examples matching regexp
    -test.short
        run smaller test suite to save time
    -test.testlogfile file
        write test action log to file (for use only by cmd/go)
    -test.timeout d
        panic test binary after duration d (default 0, timeout disabled)
    -test.trace file
        write an execution trace to file
    -test.v
        verbose: print additional output
```

示例:

```json
./bk-bscp-e2e-testing -test.v
...
```

# 开始测试

- 1. 设置环境变量`BSCP_E2E_TESTING_TESTHOST`配置目标API服务地址, 例如`export BSCP_E2E_TESTING_TESTHOST=http://x.x.x.x:8080`;
- 2. 执行`bk-bscp-e2e-testing -test.v`启动测试;
