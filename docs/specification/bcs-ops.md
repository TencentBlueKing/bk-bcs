# bcs-ops shell 开发规范

## 1 脚本规范

---

### 1.1 shell规范
* 【强制】必须使用bash作为脚本语言
```shell
#!/bin/bash
```


### 1.2 命名风格

---

* 【强制】入口脚本和功能脚本没有扩展名。functions脚本必须使用.sh作为扩展名，且不可执行的。

* 【强制】脚本文件命名统一采用**英文**，不能采用拼音等其他方式；
  正例：

```shell
${action}_${target}
#如果target还能继续细分则加在中间
${action}_${target} create_private_dir
```

* 【强制】函数使用小写字母，并用下划线分隔单词。函数名后必须有圆括号。使用::分开库，不使用function关键字

* 【强制】变量使用小写字母，并用下划线分隔单词。

* 【强制】常量和环境变量名使用大写字母，并用下划线分隔单词，声明在脚本文件的顶部。

---

### 1.3 变量声明

* 【强制】只读变量使用 readonly 或者 declare -r 来确保变量只读。
```shell
zip_version="$(dpkg --status zip | grep Version: | cut -d ' ' -f 2)"
if [[ -z "${zip_version}" ]]; then
error_message
else
readonly zip_version
fi
```

* 【强制】使用 local 声明特定函数的局部变量。声明和赋值应该在不同行。
```shell
my_func2() {
  local name="$1"

  # Separate lines for declaration and assignment:
  local my_var
  my_var="$(my_func)"
  (( $? == 0 )) || return
}
```

---


### 1.4 文件内容

#### 1.4.1 文件头

* 【强制】每个文件的开头都必须包含其内容的描述。
```shell
#!/bin/bash
#
# Perform hot backups of Oracle databases.
```
* 【可选】版权声明和作者信息

#### 1.4.2 常量定义

* 【强制】函数必须定义在常量下面

#### 1.4.3 函数定义

* 【强制】函数与函数之间禁止写代码
* 【强制】main函数放在文件末尾
* 【强制】有main函数的文件，最后一个非注释行应该是main "$@"

---

### 1.5 函数注释
* 【强制】每个函数都必须注释，无论长度或复杂度

功能描述。

全局变量：使用和修改的全局变量列表。

论点：采取的论点。

输出：输出到 STDOUT 或 STDERR。

返回：除上次命令运行的默认退出状态之外的返回值。
```shell
#######################################
# Cleanup files from the backup directory.
# Globals:
#   BACKUP_DIR
#   ORACLE_SID
# Arguments:
#   None
#######################################
function cleanup() {
}
```
* 【可选】函数中的实现可以添加说明帮助他人理解

---

### 1.6 格式
* 【强制】缩进2个空格。禁止使用tab
* 【强制】无特殊情况最大行长度为 80 个字符。

#### 1.6.1 管道
* 【强制】每行只能包含一个管道，如果有多个管道，则应该分成多行
```shell
# All fits on one line
command1 | command2

# Long commands
command1 \
| command2 \
| command3 \
| command4
```
#### 1.6.2 循环
* 【强制】;do或;then应该与if while for关键字在同一行
```shell
# If inside a function, consider declaring the loop variable as
# a local to avoid it leaking into the global environment:
# local dir
for dir in "${dirs_to_cleanup[@]}"; do
  if [[ -d "${dir}/${ORACLE_SID}" ]]; then
    log_date "Cleaning up old files in ${dir}/${ORACLE_SID}"
    rm "${dir}/${ORACLE_SID}/"*
    if (( $? != 0 )); then
      error_message
    fi
  else
    mkdir -p "${dir}/${ORACLE_SID}"
    if (( $? != 0 )); then
      error_message
    fi
  fi
done
```

#### 1.6.3 case
* 【强制】case项缩进两个空格
* 【强制】单行需要以空格;;结尾
* 【强制】多行;;单独一行
```shell
case "${expression}" in
  a)
    variable="…"
    some_command "${variable}" "${other_expr}"
    ;;
  absolute)
    actions="relative"
    another_command "${actions}" "${other_expr}"
    ;;
  *) error "Unexpected option ${flag}" ;;
esac
```

#### 1.6.4 使用变量
* 【强制】使用${var}而不是$var
* 【强制】当变量包含变量，命令替换，空格或Shell 元字符，应始终使用引号括起来

---

### 1.7 命令替换
* 【强制】使用${}而不是``

---

### 1.8 测试
* 【强制】使用 [[ ]] 而不是[ ]和test

```shell
if [[ "filename" =~ ^[[:alnum:]]+name ]]; then
  echo "Match"
fi

# This matches the exact pattern "f*" (Does not match in this case)
if [[ "filename" == "f*" ]]; then
  echo "Match"
fi
```

---

### 1.9 算数
* 【强制】使用 (( … ))和$(( … ))而不是 let，$[  ]或expr。

---

## 2 脚本分类
1.入口脚本，接收和解析参数，封装各个主体功能的调用，可独立执行，包含main函数
2.流程脚本，封装入口脚本所调用的主体功能，可独立执行，包含main函数
3.功能脚本，封装各个具体的功能点，可独立执行，包含main函数
4.函数脚本，封装各类公共函数，无需独立执行

## 附: 参考文档

* [google Shell Style Guide](https://google.github.io/styleguide/shellguide.html)

## 推荐 vscode 开发环境
1. 安装 [ShellCheck](https://marketplace.visualstudio.com/items?itemName=timonwong.shellcheck) 插件。忽略错误：在对应行上方添加 `#shellcheck`。
	> 应尽可能满足 `shellsheck` 的要求，而不是大段忽略
	```bash
	# 未找到 source 文件路径而报错
	#shellcheck source=/dev/null

	# 忽略某个错误
	#shellcheck disable=2086
	```

2. 安装 [shfmt](https://github.com/mvdan/sh) 工具，使用 `shfmt` 格式化脚本。可以搭配 [shell-format](https://marketplace.visualstudio.com/items?itemName=foxundermoon.shell-format) 插件使用。
	> shfmt 无法对 `maxline` 做限制
	```bash
	shfmt -w -i 2 -ci -bn "<filename>"
	```
