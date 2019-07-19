# CNI插件实践

CNI是Cloud Native Computing Foundation项目，为Linux容器提供配置网络接口的标准和以该标
准扩展插件提供基础函数库，当前在0.8+的版本上也已经支持windows插件。CNI仅关注两件事情：容器
网络资源分配、容器网络资源释放。网络资源主要包含两个层面：网络设备资源、IP资源。

CNI社区主要包含两个项目：

* [cni](https://github.com/containernetworking/cni)：提供标准定义，并针对容器标准实现
  了golang的基础库，方便用户使用
* [plugins](https://github.com/containernetworking/plugins)：社区提供的一些通用的插件实现

CNI标准规范插件是一个可执行主体，执行过程需要符合以下四个方面规范，但是并未限定具体实现。
* 插件输入
* 插件输出
* IP地址管理输入、输出
* 插件链式调用

输入输出的细节可以查看CNI社区定义：[https://github.com/containernetworking/cni/blob/master/SPEC.md](https://github.com/containernetworking/cni/blob/master/SPEC.md)，不同版本会有一些差别。

## plugins

社区已经提供了一些通用的插件，收录在plugins项目中。主要目录结构如下:

```
plugins
├── ipam
│   ├── dhcp
│   ├── host-local
│   └── static
├── main
│   ├── bridge
│   ├── host-device
│   ├── ipvlan
│   ├── loopback
│   ├── macvlan
│   ├── ptp 
│   ├── vlan
│   └── windows
└── meta
    ├── bandwidth
    ├── firewall
    ├── flannel
    ├── portmap
    ├── sbr
    └── tuning
```

* ipam：社区提供IP资源管理实现。
  * dhcp：利用dhcp实现容器IP地址租赁，需要有dhcp服务
  * host-local：为本地独立分配一个网段，可以在该网段内进行IP资源申请和释放
  * static：静态分配IP地址，一般用于测试
* main: 常用的CNI插件
* meta：辅助性工具，本身不能创建网络设备和IP地址管理，一般用于进行辅助性配置
  * bandwidth：限制容器网络带宽
  * Firewall：容器防火墙
  * flannel：flannel网络方案插件
  * portmap：容器与主机之间端口映射
  * sbr：容器路由控制
  * tuning：系统参数配置

## golang接口与库

为方便社区协作，CNI针对容器网络定义提供了golang的接口（0.6.0版本）：[https://github.com/containernetworking/cni/blob/v0.6.0/libcni/api.go#L51](https://github.com/containernetworking/cni/blob/v0.6.0/libcni/api.go#L51)

```golang
type CNI interface {
	AddNetworkList(net *NetworkConfigList, rt *RuntimeConf) (types.Result, error)
	DelNetworkList(net *NetworkConfigList, rt *RuntimeConf) error

	AddNetwork(net *NetworkConfig, rt *RuntimeConf) (types.Result, error)
	DelNetwork(net *NetworkConfig, rt *RuntimeConf) error
}
```

接口只定义了四个方法：增加网络，删除网络，增加网络列表，删除网络列表。其中NetworkConfig/NetworkConfigList
是CNI标准定义的JSON输入，RuntimeConf是CNI定义的每次运行时需要输入的环境变量。

如果我们只需要扩展单个CNI插件，一般已经不再需要直接实现该接口。针对JSON配置解析，环境变量解析，
CNI动作判定，社区都已经完成封装：https://github.com/containernetworking/cni/blob/v0.6.0/pkg/skel/skel.go#L221

```golang
// PluginMain is the core "main" for a plugin which includes automatic error handling.
//
// The caller must also specify what CNI spec versions the plugin supports.
//
// When an error occurs in either cmdAdd or cmdDel, PluginMain will print the error
// as JSON to stdout and call os.Exit(1).
//
// To have more control over error handling, use PluginMainWithError() instead.
func PluginMain(cmdAdd, cmdDel func(_ *CmdArgs) error, versionInfo version.PluginInfo) {
	if e := PluginMainWithError(cmdAdd, cmdDel, versionInfo); e != nil {
		if err := e.Print(); err != nil {
			log.Print("Error writing error JSON to stdout: ", err)
		}
		os.Exit(1)
    }
}
```

我们只需要对PluginMain注入我们自己实现增加网络(cmdAdd)，删除网络(cmdDel)的函数实现即可。
cmdAdd与cmdDel都需要接受参数CmdArgs，这个是封装了CNI的JSON配置和环境变量两部分输入。

```golang
type CmdArgs struct {
	ContainerID string  //需要设置网络的容器ID，来自环境变量CNI_CONTAINERID
	Netns       string  //容器的网络命名空间，来自环境变量CNI_NETNS
	IfName      string  //容器内网卡的名称，来自环境变量CNI_IFNAME
	Args        string  //CNI参数，来自环境变量CNI_ARGS
	Path        string  //CNI二进制工具目录，来自环境变量CNI_PATH
	StdinData   []byte  //CNI插件Json配置，需要自行解析内容
}
```

stdinData来自CNI插件Json配置，不同的插件该部分都有所差异，所以在CmdArgs中使用字符数组存储，
各CNI插件根据自行定义的具体结构体再进行反序列化。

## PTP插件分析

我们通过分析社区的ptp插件看下如何实现一个CNI插件。[https://github.com/containernetworking/plugins/tree/v0.6.0/plugins/main/ptp](https://github.com/containernetworking/plugins/tree/v0.6.0/plugins/main/ptp)

PTP主要是通过veth pair + 路由控制来实现容器网络：
* 创建veth pair，一端连接容器，一端连接主机，实现容器到主机节点的联通
* 针对容器IP地址，在主机上添加路由控制，实现本机上多个容器之间互联。

**PTP的cmdAdd相关步骤**
* step1：通过CmdArgs.StdinData解析自定义的CNI json配置
* step2：借助CNI库封装，调用JSON文件中指定的ipam工具，申请IP地址
* step3：格式化ipam工具的返回结果，避免不同版本之间的结果不兼容的问题
* step4：创建veth pair，一端设置到容器中，并配置申请到的IP地址
* step5，在主机节点设置容器IP的路由规则
* step6，如果开启了ipMasq特性，设置SNAT规则
* step7，将结果格式化，并打印到输出流

```golang
func cmdAdd(args *skel.CmdArgs) error {
    //step1：通过CmdArgs.StdinData解析自定义的CNI json配置
	conf := NetConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %v", err)
	}
    //step2：借助CNI库封装，调用JSON文件中指定的ipam工具，申请IP地址
	r, err := ipam.ExecAdd(conf.IPAM.Type, args.StdinData)
	if err != nil {
		return err
	}
	//step3：格式化ipam工具的返回结果，避免不同版本之间的结果不兼容的问题
	result, err := current.NewResultFromResult(r)
	if err != nil {
		return err
	}
	if len(result.IPs) == 0 {
		return errors.New("IPAM plugin returned missing IP config")
	}

	if err := ip.EnableForward(result.IPs); err != nil {
		return fmt.Errorf("Could not enable IP forwarding: %v", err)
	}
	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()
    //step4：创建veth pair，一端设置到容器中，并配置申请到的IP地址
	hostInterface, containerInterface, err := setupContainerVeth(netns, args.IfName, conf.MTU, result)
	if err != nil {
		return err
	}
    //step5，在主机节点设置容器IP的路由规则
	if err = setupHostVeth(hostInterface.Name, result); err != nil {
		return err
	}
    //step6，如果开启了ipMasq特性，设置SNAT规则
	if conf.IPMasq {
		chain := utils.FormatChainName(conf.Name, args.ContainerID)
		comment := utils.FormatComment(conf.Name, args.ContainerID)
		for _, ipc := range result.IPs {
			if err = ip.SetupIPMasq(&ipc.Address, chain, comment); err != nil {
				return err
			}
		}
	}

	result.DNS = conf.DNS
	result.Interfaces = []*current.Interface{hostInterface, containerInterface}
    //step7，将结果格式化，并打印到输出流
	return types.PrintResult(result, conf.CNIVersion)
}
```

**PTP的cmdDel**
* step1：通过CmdArgs.StdinData解析自定义的CNI json配置
* step2：调用json配置中指定的ipam工具，释放IP地址资源
* step3：切换到容器网络命名空间中，清理网卡
* step4：如果开启了IPMasq特性，则删掉在add阶段增加SNAT规则

```golang
func cmdDel(args *skel.CmdArgs) error {
    //step1：通过CmdArgs.StdinData解析自定义的CNI json配置
	conf := NetConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %v", err)
	}
    //step2：调用json配置中指定的ipam工具，释放IP地址资源
	if err := ipam.ExecDel(conf.IPAM.Type, args.StdinData); err != nil {
		return err
	}

	if args.Netns == "" {
		return nil
	}

	//step3：切换到容器网络命名空间中，清理网卡
	var ipn *net.IPNet
	err := ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
		var err error
		ipn, err = ip.DelLinkByNameAddr(args.IfName, netlink.FAMILY_V4)
		if err != nil && err == ip.ErrLinkNotFound {
			return nil
		}
		return err
	})

	if err != nil {
		return err
	}
    //如果开启了IPMasq特性，则删掉在add阶段增加SNAT规则
	if ipn != nil && conf.IPMasq {
		chain := utils.FormatChainName(conf.Name, args.ContainerID)
		comment := utils.FormatComment(conf.Name, args.ContainerID)
		err = ip.TeardownIPMasq(ipn, chain, comment)
	}
	return err
}
```



