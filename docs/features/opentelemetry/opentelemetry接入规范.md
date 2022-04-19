# OpenTelemetry接入规范

## Overview

### 什么是Opentelemetry？

OpenTelemetry 由一系列API，SDK，工具组成，旨在生成和管理观测类数据，如 trace、metrics、logs。OpenTelemetry 提供与 vendor 无关的实现，根据你自身的需要将观测类数据导出到不同的后端，它支持多种流行的开源项目，如Jaeger和Prometheus。
Opentelemetry不提供像Jaeger和Prometheus一样的可观测性后端服务，它只是支持生成和导出可观测数据到各种开源和商业性的后端服务。
## **OpenTelemetry整体架构**


**Application**： 一般的应用程序，使用了OpenTelemetry的Library (实现了API的SDK)。

**OTel Library**：也称为 SDK，负责在客户端程序里采集观测数据，包括 metrics，traces，logs，对观测数据进行处理，之后根据 Exporter 的类型，将观测数据发送到 Collector 或者直接发送到 Backend 中。

**OTel Collector**：负责根据 OTLP收集数据，以及将观测数据导出到外部系统的组件。不同的提供商要想能让观测数据持久化到自己的产品里，需要按照 OpenTelemetry 的标准开发Exporter以接收和导出数据。如 Prometheus，Jaeger，Kafka，zipkin 等。

**Backend**： 负责持久化观测数据，Collector 本身不会去负责持久化观测数据，需要外部系统提供。

想要了解更多OpenTelemetry的规范，可参考[官方文档](https://opentelemetry.io/docs/reference/specification/)
## otel包
```
otel
├── examples
│   ├── jaeger          # 使用jaeger作为otel后端的demo
│   ├── prometheus      # 使用prometheus作为otel后端的demo
│   └── unified         # 使用jaeger和prometheus作为otel后端的demo
├── metric              
│   ├── metric.go       # otel的metric的实现入口
│   ├── options.go      # 初始化options参数
│   └── prometheus      # 使用prometheus作为otel后端的实现
└── trace               
    ├── jaeger          # 使用jaeger作为otel后端的实现
    ├── options.go      # 初始化options参数
    ├── resource        # trace的资源对象
    ├── tracer.go       # otel的trace的实现入口
    ├── utils           # 封装的tracer API
    └── zipkin          # 使用zipkin作为otel后端的实现
```
详情请查看包 `github.com/Tencent/bk-bcs/bcs-common/pkg/otel`
## tracing接入
### tracing初始化
tracing定义了工厂`InitTracerProvider`接口，可以将可观测数据导出到不同的后端（Jaeger/Zipkin）。
#### 参数介绍
```
type Options struct {
	// factory parameter
	TracingSwitch string `json:"tracingSwitch" value:"off" usage:"tracing switch"`
	TracingType   string `json:"tracingType" value:"jaeger" usage:"tracing type(default jaeger)"`

	ServiceName string `json:"serviceName" value:"bcs-common/pkg/otel" usage:"tracing serviceName"`

	ExporterURL string `json:"exporterURL" value:"" usage:"url of exporter"`

	ResourceAttrs []attribute.KeyValue `json:"resourceAttrs" value:"" usage:"attributes of traced service"`
}
```
* TracingSwitch 开关类型，默认值是`off`，`off`时初始化为 `TracerProvider{}`对象；`on`时会根据参数`TracingType`选择不同的后端链路追踪系统实现
* TracingType 后端具体的链路系统，默认值是`jaeger`，还对`Zipkin`提供了简单支持
* ServiceName trace的服务名称
* ExporterURL 后端链路接收数据的URL
* ResourceAttrs 被追踪对象/服务的属性信息，本质上是一系列键值对

#### 初始化
以`bcs-common/pkg/otel/examples/jaeger/single-process/main.go`中的demo为例：
```
opts := trace.Options{
    	TracingSwitch: "on",
    	TracingType: "jaeger",
        ServiceName: service,
        ExporterURL: "http://localhost:14268/api/traces",
        ResourceAttrs: []attribute.KeyValue{
    		attribute.String("environment", environment),
			attribute.Int64("ID", id),
    	},
    }

    op := []trace.Option{}
    if opts.TracingSwitch != "" {
    	op = append(op, trace.TracerSwitch(opts.TracingSwitch))
	}
	if opts.TracingType != "" {
		op = append(op, trace.TracerType(opts.TracingType))
	}
	if opts.ServiceName != "" {
		op = append(op, trace.ServiceName(opts.ServiceName))
	}
	if opts.ExporterURL != "" {
		op = append(op, trace.ExporterURL(opts.ExporterURL))
	}
	if opts.ResourceAttrs != nil {
		op = append(op, trace.ResourceAttrs(opts.ResourceAttrs))
	}

    tp, err := trace.InitTracerProvider(opts.ServiceName, op...)
```
    
可以初始化生成global tracer provider，然后利用OpenTelemetry的接口进行操作。
#### 进程方法设置追踪
```
1. init tracer provider，生成otel global tracer provider
2. 利用tracer provider生成tracer
3. _, span := tracer.Start(ctx, spanName)
    defer span.End()
4. span.SetAttributes(kv)
5. span.SetStatus(code, description)
```

#### 进程内跨方法设置追踪
```
1. init tracer provider，生成otel global tracer provider
2. 利用tracer provider生成tracer
主方法A
3. ctx, span := tracer.Start(ctx, spanName)
    defer span.End()
4. span.SetAttributes(kv)
5. span.SetStatus(code, description)
6. 调用方法 B，通过ctx进程传递
方法B(ctx, ....)
7. 利用tracer provider生成tracer
8. _, span := tracer.Start(ctx, spanName)
    defer span.End()
9. span.SetAttributes(kv)
10. span.SetStatus(code, description)
```
#### 跨进程设置追踪
使用	SetTextMapPropagator在进程之间传递上下文

````
client端
1. init tracer provider，生成otel global tracer provider
2. 利用SetTextMapPropagator(propagation.TraceContext{})，将SpanContext进行注入到 http headers，并建立联系
3. 利用tracer provider生成tracer
4. 主方法
    ctx, span := tracer.Start(ctx, spanName)
    defer span.End()
5. 被调用方法一：(rpc、HTTP调用) 客户端，第一个参数是ctx
    // 从ctx获取parentSpan 并设置span child关系；开启子span并将子span通过ctx传递
    ctx, span := tracer.Start(ctx, spanName)
    defer span.End()
    // 设置tag和span status
    span.SetAttributes(kv)
    span.SetStatus(code, description)
    
服务端
1. init tracer provider，生成otel global tracer provider
2. 利用tracer provider生成tracer
2. 接口实现中
    // 开启服务端span，通过request的context传递根span
    _, span := utils.Tracer(tracerName).Start(ctx, spanName)    
    defer span.End()
````
## metric接入
当前OpenTelemetry metric SDK[尚不稳定](https://opentelemetry.io/status/#metrics) ，推荐在稳定后使用。
### metric初始化
tracing定义了工厂`InitMeterProvider`接口，目前支持将metrics数据导出到prometheus后端。
#### 参数介绍
```
type Options struct {
	// factory parameter
	MetricSwitchStatus string `json:"meterSwitchStatus" value:"off" usage:"meter switch status"`
	MetricType         string `json:"meterType" value:"prometheus" usage:"meter type(default prometheus)"`

	// ProcessorWithMemory controls whether the processor remembers metric
	// instruments and label sets that were previously reported.
	// When Memory is true, Reader.ForEach() will visit
	// metrics that were not updated in the most recent interval.
	ProcessorWithMemory prometheus.MemoryOption `json:"processorWithMemory" value:"false" usage:"processor memory policy"`

	// ControllerOption is the interface type slice that applies the value to a configuration option.
	ControllerOption []controller.Option `json:"controllerOption" value:"" usage:"applies the value to a configuration option"`
}
```
* MetricSwitchStatus 开关类型，默认值是`off`，`off`时初始化为 `Controller{}`对象；`on`时会根据参数`TracingType`选择不同的metric后端
* MetricType 后端具体的metric系统，默认值是`prometheus`，目前只支持prometheus
* ProcessorWithMemory 设置processor是否记录之前上报到metric instruments和labels

#### 初始化
```
1. init meter provider，生成otel global meter provider
2. 利用meter provider生成meter
3. 利用meter根据需要生成各种类型的instruments
    requestLatency := otelmetric.Must(meter).NewFloat64Histogram(name,
		otelmetric.WithDescription(description))
	requestCount := otelmetric.Must(meter).NewFloat64Counter(name",
		otelmetric.WithDescription(description))
4. 启动http metric server
5. 设置metric labels
6. 使用生成的instruments进行数据采集
    requestCount.Add(ctx, value, labels...)
    requestLatency.Record(ctx, value, labels...)	
```