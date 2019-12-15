# broccoli

公共库

# 代码生成工具 tools/bin/gen-broccoli

* 生成二进制文件
```bash
go build -o tools/bin/ ./tools/gen-broccoli

```
* 将工具添加系统环境变量中
* linux : export PATH=$PATH:$GOPATH/src/broccoli/tools/bin
* windows : PATH=%PATH%;%GOPATH%/src/broccoli/tools/bin