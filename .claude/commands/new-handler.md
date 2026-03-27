新建一个 Portal API 处理器（Handler）。

用法：`/new-handler business/modulename`

请在 `business/` 目录下创建对应的 handler.go 文件，遵循以下规范：

1. **路由注册**：在 `server.go` 的 `setupRoutes()` 中注册新路由（前缀 `/portal`）
2. **认证**：公开接口使用 `business.FindAccountFromCookie()`，受保护接口检查 `accountModel.IsAnonymous()`
3. **数据库**：使用 `datastore.NamedQuery()` 或 `datastore.NamedExec()`，**必须使用命名参数防止 SQL 注入**
4. **响应格式**：使用 Gin 的 `gctx.JSON()` 返回，配合 `models` 包的响应模型

Handler 模板：
```go
package modulename

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "portal/business"
    "portal/models"
)

func ListHandler(gctx *gin.Context) {
    // 1. 可选：获取当前用户
    accountModel, err := business.FindAccountFromCookie(gctx)
    if err != nil || accountModel.IsAnonymous() {
        gctx.JSON(http.StatusUnauthorized, models.Fail("未登录"))
        return
    }

    // 2. 获取请求参数
    page := gctx.DefaultQuery("page", "1")

    // 3. 数据库查询（使用命名参数）
    sqlText := `select * from table where owner = :owner limit 10`
    sqlParams := map[string]interface{}{"owner": accountModel.Uid}
    rows, err := datastore.NamedQuery(sqlText, sqlParams)
    if err != nil {
        gctx.JSON(http.StatusInternalServerError, models.Fail("查询失败"))
        return
    }

    // 4. 返回结果
    gctx.JSON(http.StatusOK, models.Ok(results))
}
```

在 `server.go` 中注册：
```go
portalGroup.GET("/modulename", modulename.ListHandler)
portalGroup.POST("/modulename", modulename.CreateHandler)
```

**注意**：
- 所有 SQL 必须使用 `:param` 命名参数，禁止字符串拼接
- `SERVE_MODE=SELFHOST` 时跳过 Turnstile（本地开发）
- 表名格式：`community.articles`、`public.accounts`
