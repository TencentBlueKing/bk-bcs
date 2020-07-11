# 简介
　　bcs Metric SDK 包致力于为bcs体系内的各组件提供灵活、方便、可插拔式的运行时指标导出服务。指标导出服务以http服务的方式对外提供。Metric SDK支持基础的数值型（仅包括int8, int, int16, int32, int64, float32, float64）和字符串两种类型的key-value指标导出服务。对于数值型的指标，为了兼顾不同场景，SDK统一以float64对外展示。

各组件可根据自身需求定制自身需要导出的指标信息，Metric SDK 的主要包含以下特性：
 - 目前仅支持`golang`语言，其它语言暂不支持。
 - 默认提供该组件的runtime metric。如：使用的CPU数量、goroutine数量和内存等信息。
 - 非侵入式接口设计。核心指标数据通过回调函数进行采集，在满足接口间隔离、解耦的同时，方便后续的功能扩展，尽量避免SDK的更新、升级、迭代会影响到相关的使用组件。
 - `非阻塞`式接口设计。由于无法保证各组件在使用接口时的具体行为，metric在处理每一个metric时会默认设置一个`超时时间`为`5s`，保证所有的metric信息不会因为某一个metric的阻塞而造成拉取metric信息失败。所以使用SDK的组件在实现相关接口时尽量采用`非阻塞`式设计。
 - 提供metric`分组管理`机制，各组件可根据自身情况进行metric的分类管理和导出。
 - 对于同一组件，metric SDK 也提供各实例间的`个性化标识`功能，方便指标数据汇聚后的过滤、清洗等。该功能通过label(key, value)实现。
 　　考虑以下场景，bcs-health在深圳、上海、成都均部署有，那么当指标数据汇聚以后如何区分不同的bcs-health集群和实例呢，我们可以在启动bcs-health时配置不同的label，如("set", "shenzhen")，("set", "shanghai")等。

# Golang Metric

|指标                                 |     意义                                                                 |
|-------------------------------------|------------------------------------------------------------------------|
|go_goroutines                        |Number of goroutines that currently exist.                           |
|go_threads                           |Number of OS threads created                                         |
|go_cpu_used                          |The number of logical CPUs usable by the current process            |
|go_memstats_alloc_bytes              |Number of bytes allocated and still in use.                          |
|go_memstats_alloc_bytes_total        |Total number of bytes allocated, even if freed.                      |
|go_memstats_sys_bytes                |Number of bytes obtained from system.                                |
|go_memstats_mallocs_total            |Total number of mallocs.                                             |
|go_memstats_frees_total              |Total number of frees.                                               |
|go_memstats_lookups_total            |Total number of pointer lookups.                                     |
|go_memstats_heap_alloc_bytes         |Number of heap bytes allocated and still in use.                     |
|go_memstats_heap_sys_bytes           |Number of heap bytes obtained from system.                           |
|go_memstats_heap_idle_bytes          |Number of heap bytes waiting to be used.                             |
|go_memstats_heap_inuse_bytes         |Number of heap bytes that are in use.                                |
|go_memstats_heap_released_bytes      |Number of heap bytes released to OS.                                 |
|go_memstats_heap_objects             |Number of allocated objects.                                         |
|go_memstats_stack_inuse_bytes        |Number of bytes in use by the stack allocator.                       |
|go_memstats_stack_sys_bytes          |Number of bytes obtained from system for stack allocator.            |
|go_memstats_mspan_inuse_bytes        |Number of bytes in use by mspan structures.                          |
|go_memstats_mspan_sys_bytes          |Number of bytes used for mspan structures obtained from s            |
|go_memstats_mcache_inuse_bytes       |Number of bytes in use by mcache structures.                         |
|go_memstats_mcache_sys_bytes         |Number of bytes used for mcache structures obtained from             |
|go_memstats_buck_hash_sys_bytes      |Number of bytes used by the profiling bucket hash table.             |
|go_memstats_gc_sys_bytes             |Number of bytes used for garbage collection system metada            |
|go_memstats_other_sys_bytes          |Number of bytes used for other system allocations.                   |
|go_memstats_next_gc_bytes            |Number of heap bytes when next garbage collection will take place    |
|go_memstats_last_gc_time_seconds     |Number of seconds since 1970 of last garbage collection.             |
|go_memstats_gc_cpu_fraction          |The fraction of this program's available CPU time used by the GC since the program started.           |

# 设计原理
使用metric SDK的组件在启用metric导出服务时，仅调用函数[metric.NewMetricController()] (./api.go#3)。调用NewMetricController()会完成以下动作：

 - 注册该组件的`元数据`信息，如下。具体含义如下：
    ```golang
   type Config struct {
   	// name of your module
   	ModuleName string
   	// running mode of your module
   	// could be one of Master_Slave_Mode or Master_Master_Mode
   	RunMode RunModeType
   	// ip address of this module running on
   	IP string
   	// port number of the metric's http handler depends on.
   	MetricPort uint
   	// cluster id of your module belongs to.
   	ClusterID string
   	// self defined info labeled on your metrics.
   	// deprecated, unused now.
   	Labels map[string]string
   	// whether disable golang's metric, default is false.
   	DisableGolangMetric bool
   	// metric http server's ssl configuration
   	SvrCaFile   string
   	SvrCertFile string
   	SvrKeyFile  string
   }
    ```

   - `ModuleName`即为使用SDK模块的名称（如bcs-health），属必填字段；
   - `IP`为该组件所在主机的物理IP地址，同时也是metric提供http服务时所绑定的IP地址，属必填字段。
   - `ClusterID`为该组件的获取的bcs Cluster ID值，属选填字段，建议有该属性的组件必填，没有的不填。方便数据的后续过滤、清洗。
   - `MetricPort`为该组件监听的http端口。

 - Metric 基础元数据信息
 
   每一个metric包含两部分的基础信息：
   - 静态描述信息
      * `Name`: Metric的name， 在同一个组件中必须保持唯一
      * `Help`: Help是关于这个metric的描述信息
      * `ConstLabels`: 是关于这个metric所需要标注的静态配置信息，可以存放如clusterid等这些信息。这个sdk默认带了一些配置信息。
    ```golang
        type MetricMeta struct {
        	// metric's name
        	Name string
        	// metric's help info, which should be short and briefly.
        	Help string
        	// metric labels, which can describe the special info about this metric.
        	ConstLables map[string]string
        }
    ```
    - 动态描述信息
      * `Value`: 这个指标具体的值，可以是数值型，也可以是字符串型。对于字符型的指标，受限于prometheus规范，会将metric的value写入到
      metric label里，具体key为`bcs_metric_value`，metric真正的值统一默认为1。
      * `VariableLabels`: 用户可以用来配置具体的动态label信息。
      
    
```golang
    type MetricResult struct {
        Value *FloatOrString
        // variable labels means that this labels value can be changed with each call.
        VariableLabels map[string]string
    }
```
# Demo
　　下面给大家展示一下具体的实用方法，源码在[这里] (fake/main.go)。

　　该Demo为大家展示以下信息：
- 如何使用FloatOrString接口。
- 如何实例化Metric SDK。
- 如何使用Metric SDK提供的插件。

```golang
package main

import (
    "fmt"
    "time"

    "github.com/Tencent/bk-bcs/bcs-common/common/metric"
)

func main() {
	c := metric.Config{
		ModuleName:          "fake_module",
		IP:                  "127.0.0.11",
		MetricPort:          9089,
		DisableGolangMetric: true,
		ClusterID:           "breeze-demo-clusterid",
	}

	healthz := func() metric.HealthMeta {
		return metric.HealthMeta{
			CurrentRole: "Master",
			IsHealthy:   true,
		}
	}

	demo := new(DemoMetric)
	numeric := metric.MetricContructor{
		GetMeta:   demo.GetNumericMeta,
		GetResult: demo.GetNumericResult,
	}

	sm := metric.MetricContructor{
		GetMeta:   demo.GetStringMeta,
		GetResult: demo.GetStringResult,
	}

	if err := metric.NewMetricController(c, healthz, &numeric, &sm); err != nil {
		fmt.Printf("new metric collector failed. err: %v\n", err)
		return
	}
	fmt.Println("start running")
	select {}
}

type DemoMetric struct{}

func (DemoMetric) GetNumericMeta() *metric.MetricMeta {
	return &metric.MetricMeta{
		Name: "timenow_seconds",
		Help: "show current time in unix time.",
		ConstLables: map[string]string{
			"c_time": "c_demolable",
		},
	}
}

func (DemoMetric) GetNumericResult() (*metric.MetricResult, error) {
	v, err := metric.FormFloatOrString(time.Now().Unix())
	if err != nil {
		return nil, err
	}
	return &metric.MetricResult{
		Value: v,
		VariableLabels: map[string]string{
			"var_key": "var_value",
		},
	}, nil
}

func (DemoMetric) GetStringMeta() *metric.MetricMeta {
	return &metric.MetricMeta{
		Name: "birthday_string",
		Help: "show current time in string.",
		ConstLables: map[string]string{
			"c_time": "c_demolable",
		},
	}
}

func (DemoMetric) GetStringResult() (*metric.MetricResult, error) {
	v, err := metric.FormFloatOrString(time.Now().String())
	if err != nil {
		return nil, err
	}
	return &metric.MetricResult{
		Value: v,
		VariableLabels: map[string]string{
			"var_key": "var_value",
		},
	}, nil
}

```

　　返回数据数据示例：
```yaml
# HELP bcs_module_infos module infos about this module
# TYPE bcs_module_infos gauge
bcs_module_infos{module_cluster_id="breeze-demo-clusterid",module_ip="127.0.0.11",module_name="fake_module"} 1
# HELP bcs_runtime_infos module infos about this module
# TYPE bcs_runtime_infos gauge
bcs_runtime_infos{module_ip="127.0.0.11",module_name="fake_module",module_pid="896",os="windows"} 1
# HELP bcs_version_infos version infos about this module
# TYPE bcs_version_infos gauge
bcs_version_infos{build_time="2017-03-28 19:50:00",git_hash="unknown",module_ip="127.0.0.11",module_name="fake_module",tag="2017-03-28 Release",version="17.03.28"} 1
# HELP fake_module_birthday_string show current time in unix time.
# TYPE fake_module_birthday_string gauge
fake_module_birthday_string{bcs_metric_value="2018-05-02 11:15:10.4181409 +0800 CST m=+14.027791501",c_time="c_demolable",module_ip="127.0.0.11",var_key="var_value"} 1
# HELP fake_module_timenow_seconds show current time in unix time.
# TYPE fake_module_timenow_seconds gauge
fake_module_timenow_seconds{c_time="c_demolable",module_ip="127.0.0.11",var_key="var_value"} 1.52523091e+09
```


