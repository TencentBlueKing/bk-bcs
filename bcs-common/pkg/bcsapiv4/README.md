# bcsapiv4

生成mockgen文件

### Installation

```bash
go install github.com/golang/mock/mockgen@latest
```

### Running mockgen

```bash
mockgen -package=mock -source=storage.go  -destination="mock/storage_mock.go"
```

命令说明
- destination：mockgen生成的文件存放的位置以及文件的名字。
- package：生成的mock文件的包名。
- source：源文件。

  上面的命令的意思是：利用mockgen，生成storage接口的mock文件，mock文件存放在当前项目下面mock文件夹下，文件名为storage.go。生成的mock文件的包名为mock