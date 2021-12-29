## 编译

### Linux环境插件编译

```shell
make
```

### Windows环境插件编译

```shell
make windows
```

## Windows环境依赖
> Windows环境下进行编译需要额外的依赖

- Golang: 配置Windows下的Golang开发环境
- TDM-gcc-x64: 编译过程中需要将go build生成的中间文件根据导出函数定义生成dll, 下载地址(https://sourceforge.net/projects/tdm-gcc)
