go示例项目


## mod相关操作

```shell
# 更新packages
go get -u
# 整理packages
go mod tidy
```

### 构建Docker镜像

```bash
# 构建docker镜像
docker build --progress=plain -t portal .

```
