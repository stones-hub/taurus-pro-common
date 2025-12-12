# 统一错误处理模块

提供统一的错误处理基础模块，支持错误码体系、错误包装/解包，遵循 Go 的 error 返回习惯。

## 特性

- ✅ 遵循 Go 的 error 返回习惯
- ✅ 支持外部协议（CS/SS）所需的显式错误码
- ✅ 允许错误包装 / 解包
- ✅ 各模块错误码独立管理
- ✅ 支持错误链（error chain）
- ✅ 兼容标准库 `errors` 包

## 核心概念

### 1. 模块专属错误码区间

每个服务分配独立范围，例如：
- loginsvr: 10000–19999
- usersvr: 20000–29999
- ordersvr: 30000–39999

避免冲突并便于追踪。

### 2. 自定义错误类型 `Error`

```go
type Error struct {
    code int32
    msg  string
}
```

## 使用方法

### 创建错误

```go
import "github.com/stones-hub/taurus-pro-common/pkg/errs"

// 创建简单错误
err := errs.New(10001, "invalid params")

// 使用格式化字符串创建错误
err := errs.Errorf(10002, "invalid params: %s", "username")
```

### 提取错误码和消息

```go
// 从任意 error 中提取错误码
code := errs.Code(err)

// 从任意 error 中提取错误消息
msg := errs.Message(err)
```

### 错误包装

```go
// 包装错误，保留原始错误信息
wrapped := errs.Wrap(originalErr, 20001, "work failed")

// 使用格式化字符串包装错误
wrapped := errs.Wrapf(originalErr, 20001, "work failed: %s", "details")
```

### 实际使用示例

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/stones-hub/taurus-pro-common/pkg/errs"
)

// 业务函数，返回 error
func doSomething(a, b int) (int, error) {
    if a < 0 {
        return 0, errs.New(10001, "invalid params")
    }
    return a + b, nil
}

// 包装错误后继续向上返回
func work() error {
    v, err := doSomething(-1, 2)
    if err != nil {
        errcode := errs.Code(err)
        log.Printf("doSomething failed: result = %d", errcode)
        // 包装后继续向上返回
        return errs.Wrap(err, 20001, "work failed")
    }
    return nil
}

func main() {
    if err := work(); err != nil {
        log.Printf("work failed: %v", err)
        log.Printf("work failed: errcode = %d", errs.Code(err))
        log.Printf("work failed: errmsg = %s", errs.Message(err))
    }
}
```

### 错误比较

```go
err1 := errs.New(10001, "error 1")
err2 := errs.New(10001, "error 2") // 相同错误码

// 使用标准库 errors.Is 进行比较
if errors.Is(err1, err2) {
    // 相同错误码，视为相等
}
```

### 错误类型断言

```go
var e *errs.Error
if errors.As(err, &e) {
    // err 是 *errs.Error 类型
    code := e.Code()
    msg := e.Msg()
}
```

### 映射到外部协议

在 API 处理器中，将错误码传给上游：

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    result, err := doBusinessLogic()
    if err != nil {
        // 提取错误码
        rsp.Result = errs.Code(err)
        // 或者使用默认值
        if rsp.Result == 0 {
            rsp.Result = 999999 // UnclassifiedError
        }
        return
    }
    // ...
}
```

## API 参考

### 类型

#### `Error`

自定义错误类型，包含错误码和错误消息。

```go
type Error struct {
    code int32
    msg  string
}
```

**方法：**
- `Code() int32` - 返回错误码
- `Msg() string` - 返回错误消息
- `Error() string` - 实现 error 接口
- `String() string` - 返回字符串表示
- `Unwrap() error` - 支持错误解包
- `Is(target error) bool` - 支持错误比较

### 函数

#### `New(code int32, msg string) *Error`

创建一个新的错误。

#### `Errorf(code int32, format string, args ...any) *Error`

使用格式化字符串创建错误。

#### `Code(err error) int32`

从 error 中提取错误码。如果 error 是 `*Error` 类型，返回其错误码；否则返回未知错误码 999999。

#### `Message(err error) string`

从 error 中提取错误消息。如果 error 是 `*Error` 类型，返回其错误消息；否则返回 `err.Error()`。

#### `Wrap(err error, code int32, msg string) error`

包装错误，保留原始错误信息。返回一个新的错误，但保留原始错误链。

#### `Wrapf(err error, code int32, format string, args ...any) error`

使用格式化字符串包装错误。

## 错误码规范

### 错误码区间分配

建议按服务模块分配错误码区间：

- **10000-19999**: 登录服务 (loginsvr)
- **20000-29999**: 用户服务 (usersvr)
- **30000-39999**: 订单服务 (ordersvr)
- **40000-49999**: 支付服务 (paymentsvr)
- **50000-59999**: 商品服务 (productsvr)
- **60000-69999**: 库存服务 (inventorysvr)
- **70000-79999**: 消息服务 (messagesvr)
- **80000-89999**: 通知服务 (notificationsvr)
- **90000-99998**: 其他服务
- **999999**: 未知错误 (系统保留)

### 错误码定义建议

建议使用 Protobuf 枚举定义错误码，客户端、服务器共享该错误定义：

```protobuf
enum ErrCode {
    // 登录服务错误码 (10000-19999)
    LoginsvrInvalidReqParams = 10001; // invalid params
    LoginsvrUserNotFound = 10002;     // user not found
    LoginsvrPasswordError = 10003;    // password error
    // ...
}
```

然后通过工具或手动生成对应的 Go 错误变量：

```go
var ErrLoginsvrInvalidReqParams = errs.New(xmsg.ErrCode_LoginsvrInvalidReqParams, "invalid params")
var ErrLoginsvrUserNotFound = errs.New(xmsg.ErrCode_LoginsvrUserNotFound, "user not found")
var ErrLoginsvrPasswordError = errs.New(xmsg.ErrCode_LoginsvrPasswordError, "password error")
```

## 设计优势

1. **可直接返回 `*errs.Error` 获取丰富上下文**
2. **也能返回普通 error，`errs.Code`、`errs.Message` 通用处理**
3. **内部日志详细，外部响应简洁**
4. **支持错误包装，保留错误链**
5. **兼容标准库 `errors` 包的所有功能**

## 注意事项

1. 错误码建议使用 Protobuf 枚举定义，便于客户端和服务器共享
2. 每个服务模块使用独立的错误码区间，避免冲突
3. 使用 `errs.Wrap` 包装错误时，会保留原始错误链，便于调试
4. 使用 `errs.Code` 和 `errs.Message` 提取错误信息时，会自动处理各种错误类型

