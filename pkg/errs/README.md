# errs 模块

提供统一的错误处理基础模块，支持错误码体系、错误包装/解包，遵循 Go 的 error 返回习惯。

## 目录

- [特性](#特性)
- [核心概念](#核心概念)
- [快速上手](#快速上手)
- [API 参考](#api-参考)
- [使用场景](#使用场景)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

## 特性

- ✅ **遵循 Go 的 error 返回习惯**：完全兼容标准库 `errors` 包
- ✅ **支持外部协议所需的显式错误码**：`uint64` 类型，范围广泛
- ✅ **允许错误包装/解包**：支持错误链（error chain）
- ✅ **各模块错误码独立管理**：可按模块分配错误码区间
- ✅ **兼容标准库 `errors` 包**：支持 `errors.Is`、`errors.As`、`errors.Unwrap`
- ✅ **支持普通错误码和位掩码错误两种模式**：灵活应对不同场景

## 核心概念

### 1. 模块专属错误码区间

每个服务分配独立范围，例如：
- loginsvr: 10000–19999
- usersvr: 20000–29999
- ordersvr: 30000–39999

避免冲突并便于追踪。

### 2. 两种错误模式

#### 普通错误码模式（推荐用于大多数场景）

**优势：**
- ✅ **几乎无限扩展**：`uint64` 可表示 0 到 18,446,744,073,709,551,615（约 1844 亿亿）
- ✅ **语义清晰**：每个错误码对应一个具体的错误场景
- ✅ **易于管理**：可以按模块分配错误码区间（如 10000-19999）
- ✅ **精确匹配**：使用 `errors.Is()` 进行精确的错误码比较
- ✅ **适合业务错误**：适合表示业务逻辑中的各种错误情况

**使用场景：**
- 业务错误码（如：10001=无效参数，10002=用户不存在）
- API 错误响应
- 需要大量不同错误码的场景

**示例：**
```go
err := errs.New(10001, "invalid params")
if errors.Is(err, errs.New(10001, "")) {
    // 精确匹配
}
```

#### 位掩码错误模式（适合组合错误场景）

**优势：**
- ✅ **组合错误**：可以同时表示多个错误（使用 `|` 操作符）
- ✅ **位检查高效**：使用位操作快速检查是否包含某个错误
- ✅ **节省空间**：一个错误对象可以表示多个错误标志
- ✅ **适合验证场景**：适合表单验证、批量操作等需要同时报告多个错误的场景

**限制：**
- ⚠️ **最多 64 个错误标志**：`uint64` 有 64 位，所以最多支持 64 个不同的错误标志
- ⚠️ **值范围限制**：错误码值通常较小（1, 2, 4, 8...）

**使用场景：**
- 表单验证（同时报告多个字段错误）
- 批量操作（同时报告多个失败项）
- 配置检查（同时报告多个配置问题）
- 需要组合错误的场景

**示例：**
```go
const (
    ErrInvalidParam uint64 = 1 << 0  // 位0: 无效参数
    ErrNotFound     uint64 = 1 << 1  // 位1: 未找到
    ErrTimeout      uint64 = 1 << 2  // 位2: 超时
    // ... 最多64个错误标志（1 << 0 到 1 << 63）
)

// 创建单个位错误
err1 := errs.NewBitmask(ErrInvalidParam, "invalid parameter")

// 创建组合错误（多个错误同时发生）
combined := ErrInvalidParam | ErrNotFound | ErrTimeout
err2 := errs.NewBitmask(combined, "multiple errors occurred")

// 使用 errors.Is 检查是否包含某个错误（位掩码模式）
if errors.Is(err2, errs.NewBitmask(ErrInvalidParam, "")) {
    // 包含 ErrInvalidParam
}

// 使用位操作检查是否包含某个位
if err2.Code() & ErrNotFound != 0 {
    // 包含 ErrNotFound
}

// 检查是否包含所有指定的位
if err2.Code() & (ErrInvalidParam | ErrNotFound) == (ErrInvalidParam | ErrNotFound) {
    // 同时包含 ErrInvalidParam 和 ErrNotFound
}
```

### 3. 错误链（Error Chain）

支持错误包装，保留原始错误信息，便于错误追踪和调试。

```go
originalErr := errs.New(10001, "invalid params")
wrappedErr := errs.Wrap(originalErr, 20001, "work failed")

// 使用 errors.Unwrap 获取原始错误
if original := errors.Unwrap(wrappedErr); original != nil {
    // original 是 originalErr
}

// errors.Is 会检查整个错误链
if errors.Is(wrappedErr, errs.New(10001, "")) {
    // 匹配成功，因为错误链中包含 10001
}
```

## 快速上手

### 安装

```bash
go get github.com/stones-hub/taurus-pro-common/pkg/errs
```

### 基本使用

```go
package main

import (
    "errors"
    "fmt"
    "github.com/stones-hub/taurus-pro-common/pkg/errs"
)

func main() {
    // 创建普通错误
    err := errs.New(10001, "invalid params")
    fmt.Printf("Code: %d, Message: %s\n", errs.Code(err), errs.Message(err))
    
    // 创建格式化错误
    err = errs.Errorf(10002, "invalid params: %s", "name")
    
    // 错误比较
    err1 := errs.New(10001, "error 1")
    err2 := errs.New(10001, "error 2") // 相同错误码
    if errors.Is(err1, err2) {
        fmt.Println("Same error code")
    }
}
```

### 位掩码错误使用

```go
package main

import (
    "errors"
    "fmt"
    "github.com/stones-hub/taurus-pro-common/pkg/errs"
)

func main() {
    // 定义位掩码错误标志
    const (
        ErrInvalidParam uint64 = 1 << 0
        ErrNotFound     uint64 = 1 << 1
        ErrTimeout      uint64 = 1 << 2
    )
    
    // 创建组合错误
    combined := ErrInvalidParam | ErrNotFound
    err := errs.NewBitmask(combined, "multiple errors")
    
    // 检查是否包含某个错误
    if errors.Is(err, errs.NewBitmask(ErrInvalidParam, "")) {
        fmt.Println("Contains ErrInvalidParam")
    }
    
    // 使用位操作检查
    if err.Code() & ErrNotFound != 0 {
        fmt.Println("Contains ErrNotFound")
    }
}
```

## API 参考

### 创建错误

#### `New(code uint64, msg string) *Error`

创建一个新的普通错误。

**参数：**
- `code`: 错误码（`uint64`）
- `msg`: 错误消息

**返回：**
- `*Error`: 错误对象

**示例：**
```go
err := errs.New(10001, "invalid params")
```

#### `Errorf(code uint64, format string, args ...any) *Error`

使用格式化字符串创建普通错误。

**参数：**
- `code`: 错误码（`uint64`）
- `format`: 格式化字符串
- `args`: 格式化参数

**返回：**
- `*Error`: 错误对象

**示例：**
```go
err := errs.Errorf(10002, "invalid params: %s", "name")
```

#### `NewBitmask(code uint64, msg string) *Error`

创建位掩码错误。

**参数：**
- `code`: 错误码（`uint64`，通常是多个位的组合，如 `1<<0 | 1<<1`）
- `msg`: 错误消息

**返回：**
- `*Error`: 错误对象

**示例：**
```go
const ErrInvalidParam uint64 = 1 << 0
const ErrNotFound uint64 = 1 << 1
err := errs.NewBitmask(ErrInvalidParam | ErrNotFound, "multiple errors")
```

#### `Bitmaskf(code uint64, format string, args ...any) *Error`

使用格式化字符串创建位掩码错误。

**参数：**
- `code`: 错误码（`uint64`）
- `format`: 格式化字符串
- `args`: 格式化参数

**返回：**
- `*Error`: 错误对象

**示例：**
```go
err := errs.Bitmaskf(ErrInvalidParam | ErrNotFound, "errors: %s", "details")
```

### 提取错误信息

#### `Code(err error) uint64`

从 error 中提取错误码。

**参数：**
- `err`: 错误对象

**返回：**
- `uint64`: 错误码。如果 error 是 `*Error` 类型，返回其错误码；否则返回 999999（未知错误码）

**示例：**
```go
code := errs.Code(err)
```

#### `Message(err error) string`

从 error 中提取错误消息。

**参数：**
- `err`: 错误对象

**返回：**
- `string`: 错误消息。如果 error 是 `*Error` 类型，返回其错误消息；否则返回 `err.Error()`

**示例：**
```go
msg := errs.Message(err)
```

### 包装错误

#### `Wrap(err error, code uint64, msg string) error`

包装错误，保留原始错误信息。

**参数：**
- `err`: 原始错误
- `code`: 新的错误码
- `msg`: 新的错误消息

**返回：**
- `error`: 包装后的错误（`*wrappedError` 类型）

**示例：**
```go
originalErr := errs.New(10001, "invalid params")
wrappedErr := errs.Wrap(originalErr, 20001, "work failed")
```

#### `Wrapf(err error, code uint64, format string, args ...any) error`

使用格式化字符串包装错误。

**参数：**
- `err`: 原始错误
- `code`: 新的错误码
- `format`: 格式化字符串
- `args`: 格式化参数

**返回：**
- `error`: 包装后的错误

**示例：**
```go
wrappedErr := errs.Wrapf(originalErr, 20001, "work failed: %s", "details")
```

#### `WrapBitmask(err error, code uint64, msg string) error`

包装错误为位掩码错误。

**参数：**
- `err`: 原始错误
- `code`: 位掩码错误码
- `msg`: 错误消息

**返回：**
- `error`: 包装后的错误

**示例：**
```go
wrappedErr := errs.WrapBitmask(originalErr, ErrInvalidParam | ErrNotFound, "validation failed")
```

#### `WrapBitmaskf(err error, code uint64, format string, args ...any) error`

使用格式化字符串包装错误为位掩码错误。

**参数：**
- `err`: 原始错误
- `code`: 位掩码错误码
- `format`: 格式化字符串
- `args`: 格式化参数

**返回：**
- `error`: 包装后的错误

**示例：**
```go
wrappedErr := errs.WrapBitmaskf(originalErr, ErrInvalidParam | ErrNotFound, "validation failed: %s", "details")
```

### Error 类型方法

#### `Error() string`

实现标准库 `error` 接口，返回错误的字符串表示。

#### `Code() uint64`

返回错误码。

#### `Msg() string`

返回错误消息。

#### `String() string`

返回错误的字符串表示（实现 `fmt.Stringer` 接口）。

#### `Unwrap() error`

实现标准库错误解包接口，用于支持错误链。对于 `Error` 类型，返回 `nil`。

#### `Is(target error) bool`

实现标准库错误比较接口。当调用 `errors.Is(err, target)` 时，会自动调用此方法。

**比较逻辑：**
- 如果源错误和目标错误的模式不同（一个是普通错误，一个是位掩码错误），返回 `false`
- 如果源错误是位掩码模式，且目标错误码是单个位，使用位检查（`e.code & t.code != 0`）
- 否则使用精确匹配（`e.code == t.code`）

### wrappedError 类型方法

`wrappedError` 是内部类型，用于表示包装后的错误。它实现了以下方法：

- `Error() string`: 返回错误信息，包含被包装的错误
- `Unwrap() error`: 返回被包装的原始错误
- `Code() uint64`: 返回错误码
- `Msg() string`: 返回错误消息
- `Is(target error) bool`: 先检查包装的错误信息，再检查原始错误链
- `As(target any) bool`: 先检查包装的错误信息，再检查原始错误链

## 使用场景

### 场景 1：业务错误处理

```go
package main

import (
    "errors"
    "github.com/stones-hub/taurus-pro-common/pkg/errs"
)

// 定义业务错误码
const (
    ErrInvalidParams = 10001
    ErrUserNotFound  = 10002
    ErrPermissionDenied = 10003
)

func GetUser(userID int) (*User, error) {
    if userID <= 0 {
        return nil, errs.New(ErrInvalidParams, "userID must be positive")
    }
    
    user, err := db.FindUser(userID)
    if err != nil {
        return nil, errs.Wrap(err, ErrUserNotFound, "user not found")
    }
    
    return user, nil
}

func main() {
    user, err := GetUser(-1)
    if err != nil {
        if errors.Is(err, errs.New(ErrInvalidParams, "")) {
            // 处理参数错误
        } else if errors.Is(err, errs.New(ErrUserNotFound, "")) {
            // 处理用户不存在错误
        }
    }
}
```

### 场景 2：表单验证（位掩码错误）

```go
package main

import (
    "errors"
    "github.com/stones-hub/taurus-pro-common/pkg/errs"
)

// 定义验证错误标志
const (
    ErrUsernameInvalid uint64 = 1 << 0
    ErrPasswordTooShort uint64 = 1 << 1
    ErrEmailInvalid     uint64 = 1 << 2
)

func ValidateForm(username, password, email string) error {
    var code uint64
    
    if username == "" || len(username) < 3 {
        code |= ErrUsernameInvalid
    }
    if len(password) < 6 {
        code |= ErrPasswordTooShort
    }
    if email == "" || !isValidEmail(email) {
        code |= ErrEmailInvalid
    }
    
    if code != 0 {
        return errs.NewBitmask(code, "validation failed")
    }
    return nil
}

func main() {
    err := ValidateForm("", "123", "bad@email")
    if err != nil {
        // 检查具体的错误
        if errors.Is(err, errs.NewBitmask(ErrUsernameInvalid, "")) {
            fmt.Println("Username is invalid")
        }
        if errors.Is(err, errs.NewBitmask(ErrPasswordTooShort, "")) {
            fmt.Println("Password is too short")
        }
        if errors.Is(err, errs.NewBitmask(ErrEmailInvalid, "")) {
            fmt.Println("Email is invalid")
        }
        
        // 检查是否包含所有错误
        allErrors := ErrUsernameInvalid | ErrPasswordTooShort | ErrEmailInvalid
        if errs.Code(err) & allErrors == allErrors {
            fmt.Println("All fields are invalid")
        }
    }
}
```

### 场景 3：批量操作（位掩码错误）

```go
package main

import (
    "github.com/stones-hub/taurus-pro-common/pkg/errs"
)

const (
    ErrItem1Failed uint64 = 1 << 0
    ErrItem2Failed uint64 = 1 << 1
    ErrItem3Failed uint64 = 1 << 2
)

func BatchProcess(items []Item) error {
    var code uint64
    
    for i, item := range items {
        if err := processItem(item); err != nil {
            code |= 1 << i
        }
    }
    
    if code != 0 {
        return errs.NewBitmask(code, "batch operation failed")
    }
    return nil
}
```

### 场景 4：错误链追踪

```go
package main

import (
    "errors"
    "github.com/stones-hub/taurus-pro-common/pkg/errs"
)

func doSomething() error {
    return errs.New(10001, "invalid params")
}

func work() error {
    err := doSomething()
    if err != nil {
        return errs.Wrap(err, 20001, "work failed")
    }
    return nil
}

func main() {
    err := work()
    if err != nil {
        // 检查包装后的错误
        if errors.Is(err, errs.New(20001, "")) {
            fmt.Println("Work failed")
        }
        
        // 检查原始错误
        if errors.Is(err, errs.New(10001, "")) {
            fmt.Println("Original error: invalid params")
        }
        
        // 获取原始错误
        if original := errors.Unwrap(err); original != nil {
            fmt.Printf("Original error: %v\n", original)
        }
    }
}
```

## 最佳实践

### 1. 错误码设计

- **按模块分配错误码区间**：每个服务分配独立范围，避免冲突
- **使用有意义的错误码**：不要使用随机数字，使用有规律的区间
- **文档化错误码**：在代码或文档中记录每个错误码的含义

```go
// 推荐：按模块分配错误码区间
const (
    // 用户模块：10000-19999
    ErrUserInvalidParams = 10001
    ErrUserNotFound      = 10002
    ErrUserExists        = 10003
    
    // 订单模块：20000-29999
    ErrOrderInvalidParams = 20001
    ErrOrderNotFound     = 20002
)
```

### 2. 错误消息设计

- **提供有意义的错误消息**：帮助开发者快速定位问题
- **包含上下文信息**：使用格式化字符串包含相关参数

```go
// 推荐
err := errs.Errorf(10001, "invalid userID: %d", userID)

// 不推荐
err := errs.New(10001, "error")
```

### 3. 错误处理

- **使用 `errors.Is` 进行错误比较**：不要直接比较错误对象
- **检查错误链**：`errors.Is` 会自动检查整个错误链
- **使用 `errors.As` 进行类型断言**：提取错误信息

```go
// 推荐
if errors.Is(err, errs.New(10001, "")) {
    // 处理错误
}

// 不推荐
if errs.Code(err) == 10001 {
    // 直接比较错误码，可能遗漏错误链中的错误
}
```

### 4. 位掩码错误使用

- **定义清晰的错误标志**：使用常量定义每个位
- **使用位操作进行组合**：使用 `|` 操作符组合多个错误
- **使用 `errors.Is` 检查单个位**：自动进行位检查
- **使用位操作检查多个位**：使用 `&` 操作符检查多个位

```go
// 推荐
const (
    ErrInvalidParam uint64 = 1 << 0
    ErrNotFound     uint64 = 1 << 1
)

combined := ErrInvalidParam | ErrNotFound
err := errs.NewBitmask(combined, "multiple errors")

// 检查单个位
if errors.Is(err, errs.NewBitmask(ErrInvalidParam, "")) {
    // 包含 ErrInvalidParam
}

// 检查多个位
if err.Code() & (ErrInvalidParam | ErrNotFound) == (ErrInvalidParam | ErrNotFound) {
    // 同时包含两个错误
}
```

### 5. 错误包装

- **在适当的层级包装错误**：在函数边界包装错误，添加上下文信息
- **保留原始错误**：使用 `Wrap` 系列函数保留原始错误信息
- **不要过度包装**：避免创建过深的错误链

```go
// 推荐
func work() error {
    err := doSomething()
    if err != nil {
        return errs.Wrap(err, 20001, "work failed")
    }
    return nil
}
```

## 常见问题

### Q1: 什么时候使用普通错误码，什么时候使用位掩码错误？

**A:** 
- **普通错误码**：用于绝大多数业务错误，需要精确匹配，错误码数量可能很多
- **位掩码错误**：用于需要同时表示多个错误的场景，如表单验证、批量操作，最多 64 个错误标志

### Q2: 普通错误码和位掩码错误可以混用吗？

**A:** 可以混用，但需要注意：
- 普通错误码和位掩码错误不会互相匹配（即使错误码值相同）
- `errors.Is` 会自动区分两种模式

```go
normalErr := errs.New(10001, "normal error")
bitmaskErr := errs.NewBitmask(1, "bitmask error") // 1 也是 1 << 0

// 不会匹配，因为模式不同
errors.Is(normalErr, bitmaskErr) // false
errors.Is(bitmaskErr, normalErr) // false
```

### Q3: 如何检查位掩码错误是否包含多个位？

**A:** 使用位操作：

```go
const (
    ErrA uint64 = 1 << 0
    ErrB uint64 = 1 << 1
    ErrC uint64 = 1 << 2
)

err := errs.NewBitmask(ErrA | ErrB, "errors")

// 检查是否包含 ErrA 和 ErrB
if err.Code() & (ErrA | ErrB) == (ErrA | ErrB) {
    // 同时包含 ErrA 和 ErrB
}

// 检查是否包含任意一个
if err.Code() & (ErrA | ErrC) != 0 {
    // 包含 ErrA 或 ErrC 中的至少一个
}
```

### Q4: `errors.Is` 和直接比较错误码有什么区别？

**A:** 
- `errors.Is` 会检查整个错误链，包括包装的错误
- 直接比较错误码只能检查当前错误，可能遗漏错误链中的错误

```go
originalErr := errs.New(10001, "original")
wrappedErr := errs.Wrap(originalErr, 20001, "wrapped")

// 推荐：使用 errors.Is
if errors.Is(wrappedErr, errs.New(10001, "")) {
    // 匹配成功，因为错误链中包含 10001
}

// 不推荐：直接比较错误码
if errs.Code(wrappedErr) == 10001 {
    // 不匹配，因为包装后的错误码是 20001
}
```

### Q5: 为什么 `Error` 类型不实现 `As()` 方法？

**A:** 
- `errors.As()` 会自动使用类型断言来处理，不需要显式实现 `As()` 方法
- `wrappedError` 需要实现 `As()` 方法，因为它需要处理错误链

### Q6: 错误码的最大值是多少？

**A:** 
- 普通错误码：`uint64` 的最大值（18,446,744,073,709,551,615）
- 位掩码错误：最多 64 个错误标志（1 << 0 到 1 << 63）

### Q7: 如何从标准库的 `error` 转换为 `*Error`？

**A:** 使用 `errors.As`：

```go
var e *errs.Error
if errors.As(err, &e) {
    // err 是 *Error 类型，可以使用 e.Code() 和 e.Msg()
    code := e.Code()
    msg := e.Msg()
}
```

## 许可证

Apache License 2.0
