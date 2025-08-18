# 企业微信应用使用示例

这个示例展示了如何使用 `pkg/util/tapi/wx` 包来集成企业微信应用，实现消息发送、素材上传等功能。

## 功能特性

- ✅ 企业微信应用实例管理
- ✅ 访问令牌自动管理（带缓存和重试）
- ✅ 多种消息类型支持（文本、Markdown、图片、语音、视频、文件、卡片、图文）
- ✅ 模板消息支持
- ✅ 小程序消息支持
- ✅ 素材上传（图片、语音、视频、文件）
- ✅ URL验证（用于回调配置）
- ✅ 自定义重试策略
- ✅ 线程安全

## 快速开始

### 1. 配置企业微信

在使用前，你需要：

1. 注册企业微信账号
2. 创建企业应用
3. 获取以下信息：
   - 企业ID (corpID)
   - 应用ID (agentID)
   - 应用密钥 (secret)
   - 验证Token（用于URL验证）

### 2. 修改配置

编辑 `wx_example.go` 文件，替换以下配置：

```go
corpID := "your_corp_id"           // 替换为你的企业ID
agentID := 1000001                 // 替换为你的应用ID
appName := "测试应用"                // 应用名称
secret := "your_app_secret"        // 替换为你的应用密钥
sendUID := []string{"user1", "user2"} // 替换为接收消息的用户ID
```

### 3. 运行示例

```bash
cd example/weixin
go run wx_example.go
```

## 主要功能演示

### 基础配置

```go
// 创建应用实例
app := wx.NewEnterWechat(corpID, agentID, appName, secret, sendUID, nil)

// 设置验证Token
app.SetToken("your_verify_token")
```

### 发送消息

#### 文本消息
```go
textMsg := &wx.Text{
    Content: "这是一条测试消息",
}
resp, err := app.SendTextMessage(textMsg)
```

#### Markdown消息
```go
markdownMsg := &wx.Markdown{
    Content: `# 标题
## 子标题
- 列表项1
- 列表项2`,
}
resp, err := app.SendMarkdownMessage(markdownMsg)
```

#### 图文消息
```go
newsMsg := &wx.News{
    Articles: []wx.Article{
        {
            Title:       "标题",
            Description: "描述",
            Url:         "https://example.com",
            PicUrl:      "https://example.com/image.jpg",
        },
    },
}
resp, err := app.SendNewsMessage(newsMsg)
```

#### 模板消息
```go
template := &wx.MessageTemplate{
    Title:   "系统通知",
    Content: "用户 {{username}} 完成了 {{action}}",
    Data: map[string]string{
        "username": "张三",
        "action":   "订单支付",
    },
}
resp, err := app.SendTemplateMessage(template)
```

### 素材上传

```go
// 上传文件
resp, err := app.UploadMaterialByPath("/path/to/file.jpg", "image")

// 上传multipart文件
file, _ := c.FormFile("file")
resp, err := app.UploadMaterialByFile(file, "image")
```

### 高级配置

#### 自定义重试策略
```go
config := &wx.EnterWechatConfig{
    TokenManagerConfig: &wx.TokenManagerConfig{
        HTTPTimeout: 30 * time.Second,
        RetryConfig: &wx.RetryConfig{
            MaxRetries:  5,
            InitialWait: 2 * time.Second,
            MaxWait:     20 * time.Second,
            Multiplier:  1.5,
        },
    },
}

app := wx.NewEnterWechat(corpID, agentID, appName, secret, sendUID, config)
```

#### Token管理
```go
// 获取Token管理器
tokenManager := app.GetTokenManager()

// 强制刷新Token
newToken, err := tokenManager.RefreshToken()

// 获取当前Token
token, err := tokenManager.GetToken()
```

## 消息类型说明

| 消息类型 | 说明 | 安全模式 |
|---------|------|----------|
| text | 文本消息 | ✅ |
| markdown | Markdown消息 | ❌ |
| image | 图片消息 | ✅ |
| voice | 语音消息 | ✅ |
| video | 视频消息 | ✅ |
| file | 文件消息 | ✅ |
| textcard | 文本卡片消息 | ✅ |
| news | 图文消息 | ❌ |

## 素材类型说明

| 素材类型 | 说明 | 支持格式 |
|---------|------|----------|
| image | 图片 | JPG, PNG, GIF等 |
| voice | 语音 | AMR |
| video | 视频 | MP4等 |
| file | 文件 | 任意格式 |

## 注意事项

1. **配置安全**：不要将真实的企业微信配置提交到代码仓库
2. **Token管理**：访问令牌会自动缓存和刷新，无需手动管理
3. **错误处理**：示例中的错误处理仅用于演示，生产环境需要更完善的错误处理
4. **并发安全**：所有方法都是线程安全的，可以并发使用
5. **网络超时**：建议根据网络情况调整HTTP超时时间
6. **重试策略**：默认使用指数退避重试策略，可根据需要调整

## 常见问题

### Q: 如何获取企业微信配置信息？
A: 登录企业微信管理后台 → 应用管理 → 自建应用 → 查看应用详情

### Q: 支持哪些消息格式？
A: 支持文本、Markdown、图片、语音、视频、文件、卡片、图文等多种格式

### Q: 如何处理Token过期？
A: 系统会自动检测Token过期并刷新，无需手动处理

### Q: 是否支持群发消息？
A: 支持通过设置 `ToUser` 为 `"@all"` 来群发消息

### Q: 如何配置回调URL？
A: 使用 `VerifyURL` 方法验证回调签名，确保请求来源的合法性

## 更多信息

- [企业微信开发文档](https://developer.work.weixin.qq.com/)
- [消息推送接口文档](https://developer.work.weixin.qq.com/document/path/90236)
- [素材管理接口文档](https://developer.work.weixin.qq.com/document/path/90253)
