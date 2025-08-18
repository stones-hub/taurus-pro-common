package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/util/tapi/wx"
)

func main() {
	// 示例配置 - 实际使用时请替换为真实的企业微信配置
	corpID := "ww245fc7dc10e8dc26"                          // 企业ID
	agentID := 1000048                                      // 应用ID
	appName := "推送小蜜"                                       // 应用名称
	secret := "AYs-o5nxJo2m_dR50kanqqcMSoDmWZ9p3KiCgY_ZVq8" // 应用密钥
	sendUID := []string{"yelei"}                            // 接收消息的用户ID列表

	fmt.Println("=== 企业微信应用使用示例 ===")

	// 1. 创建企业微信应用实例
	fmt.Println("\n1. 创建企业微信应用实例...")
	app := wx.NewEnterWechat(corpID, agentID, appName, secret, sendUID, nil)
	fmt.Printf("应用创建成功: %s (ID: %d)\n", app.Name(), app.AgentID())

	// 2. 设置验证Token（用于URL验证）
	fmt.Println("\n2. 设置验证Token...")
	fmt.Println("注意：这个Token用于企业微信回调URL验证，不是API调用的Access Token")
	fmt.Println("在企业微信管理后台 -> 应用管理 -> 自建应用 -> 接收消息 -> 设置接收消息服务器时配置")

	// 模拟真实的验证Token（实际使用时从配置或环境变量获取）
	verifyToken := "your_verify_token_for_callback"
	app.SetToken(verifyToken)
	fmt.Printf("验证Token设置完成: %s\n", verifyToken[:10]+"...")

	// 3. 获取访问令牌
	fmt.Println("\n3. 获取访问令牌...")
	fmt.Println("注意：这个Access Token由TokenManager自动管理，用于API调用")
	tokenResp, err := app.AccessToken()
	if err != nil {
		log.Printf("获取访问令牌失败: %v\n", err)
		// 注意：这里使用模拟数据继续演示
		fmt.Println("使用模拟数据继续演示...")
	} else {
		fmt.Printf("访问令牌获取成功，有效期: %d秒\n", tokenResp.ExpiresIn)
	}

	// 4. 发送文本消息
	fmt.Println("\n4. 发送文本消息...")
	textMsg := &wx.Text{
		Content: "这是一条测试文本消息，发送时间: " + time.Now().Format("2006-01-02 15:04:05"),
	}

	textResp, err := app.SendTextMessage(textMsg)
	if err != nil {
		log.Printf("发送文本消息失败: %v\n", err)
	} else {
		fmt.Printf("文本消息发送成功，消息ID: %s\n", textResp.MsgId)
	}

	// 5. 发送Markdown消息
	fmt.Println("\n5. 发送Markdown消息...")
	markdownMsg := &wx.Markdown{
		Content: `# 测试Markdown消息
## 支持多种格式
- **粗体文本**
- *斜体文本*
- ` + "`代码块`" + `
- [链接](https://example.com)

> 引用文本

发送时间: ` + time.Now().Format("2006-01-02 15:04:05"),
	}

	markdownResp, err := app.SendMarkdownMessage(markdownMsg)
	if err != nil {
		log.Printf("发送Markdown消息失败: %v\n", err)
	} else {
		fmt.Printf("Markdown消息发送成功，消息ID: %s\n", markdownResp.MsgId)
	}

	// 6. 发送文本卡片消息
	fmt.Println("\n6. 发送文本卡片消息...")
	cardMsg := &wx.TextCard{
		Title:       "测试卡片消息",
		Description: "这是一个文本卡片消息的示例，支持点击跳转。\n发送时间: " + time.Now().Format("2006-01-02 15:04:05"),
		Url:         "https://example.com",
		BtnTxt:      "查看详情",
	}

	cardResp, err := app.SendTextCardMessage(cardMsg)
	if err != nil {
		log.Printf("发送文本卡片消息失败: %v\n", err)
	} else {
		fmt.Printf("文本卡片消息发送成功，消息ID: %s\n", cardResp.MsgId)
	}

	// 7. 发送图文消息
	fmt.Println("\n7. 发送图文消息...")
	newsMsg := &wx.News{
		Articles: []wx.Article{
			{
				Title:       "测试图文消息",
				Description: "这是一个图文消息的示例，支持图片、标题和描述。",
				Url:         "https://example.com",
				PicUrl:      "https://via.placeholder.com/300x200",
			},
		},
	}

	newsResp, err := app.SendNewsMessage(newsMsg)
	if err != nil {
		log.Printf("发送图文消息失败: %v\n", err)
	} else {
		fmt.Printf("图文消息发送成功，消息ID: %s\n", newsResp.MsgId)
	}

	// 8. 发送模板消息
	fmt.Println("\n8. 发送模板消息...")
	template := &wx.MessageTemplate{
		Title:   "系统通知",
		Content: "用户 {{username}} 在 {{time}} 完成了 {{action}}",
		Color:   "#FF0000",
		URL:     "https://example.com/notification",
		Data: map[string]string{
			"username": "张三",
			"time":     time.Now().Format("2006-01-02 15:04:05"),
			"action":   "订单支付",
		},
	}

	templateResp, err := app.SendTemplateMessage(template)
	if err != nil {
		log.Printf("发送模板消息失败: %v\n", err)
	} else {
		fmt.Printf("模板消息发送成功，消息ID: %s\n", templateResp.MsgId)
	}

	// 9. 发送小程序消息
	fmt.Println("\n9. 发送小程序消息...")
	miniAppTemplate := &wx.MessageTemplate{
		Title:   "小程序通知",
		Content: "点击查看小程序页面",
		URL:     "https://example.com",
		MiniApp: &wx.MiniAppConfig{
			AppID:    "wx123456789",
			PagePath: "pages/index?userid=123",
		},
	}

	miniAppResp, err := app.SendTemplateMessage(miniAppTemplate)
	if err != nil {
		log.Printf("发送小程序消息失败: %v\n", err)
	} else {
		fmt.Printf("小程序消息发送成功，消息ID: %s\n", miniAppResp.MsgId)
	}

	// 10. 发送消息给指定用户
	fmt.Println("\n10. 发送消息给指定用户...")
	customUsers := []string{"user3", "user4"}
	customMessage := &wx.EnterWeChatSendMessage{
		ToUser:  "",
		MsgType: "text",
		Text: wx.Text{
			Content: "这是一条发送给指定用户的消息",
		},
	}

	customResp, err := app.SendMessageToUsers(customMessage, customUsers)
	if err != nil {
		log.Printf("发送自定义消息失败: %v\n", err)
	} else {
		fmt.Printf("自定义消息发送成功，消息ID: %s\n", customResp.MsgId)
	}

	// 11. 演示URL验证（企业微信回调验证）
	fmt.Println("\n11. 演示URL验证（企业微信回调验证）...")
	fmt.Println("这是企业微信回调URL验证的真实场景演示")
	fmt.Println("当企业微信向你的服务器发送回调请求时，需要验证请求的合法性")

	// 模拟企业微信回调请求的参数
	// 这些参数在企业微信向你的服务器发送回调时提供
	callbackParams := map[string]string{
		"msg_signature": "test_signature",   // 企业微信回调签名
		"timestamp":     "1234567890",       // 时间戳
		"nonce":         "test_nonce",       // 随机数
		"echostr":       "test_echo_string", // 随机字符串
	}

	fmt.Printf("回调参数: %+v\n", callbackParams)
	fmt.Printf("当前验证Token: %s\n", app.Token())

	// 验证回调请求的签名
	// 注意：这里使用模拟参数，实际使用时这些参数来自企业微信的HTTP请求
	verifyResult, err := app.VerifyURL(
		callbackParams["msg_signature"],
		callbackParams["timestamp"],
		callbackParams["nonce"],
		callbackParams["echostr"],
	)

	if err != nil {
		fmt.Printf("URL验证失败: %v\n", err)
		fmt.Println("验证失败原因：模拟的签名参数不正确")
		fmt.Println("在实际使用中，企业微信会提供正确的签名参数")
	} else {
		fmt.Printf("URL验证成功，返回: %s\n", verifyResult)
	}

	// 展示验证Token的用途
	fmt.Println("\n验证Token的用途说明:")
	fmt.Println("1. 在企业微信管理后台配置接收消息服务器时设置")
	fmt.Println("2. 用于验证回调请求是否真的来自企业微信官方")
	fmt.Println("3. 防止恶意请求冒充企业微信回调")
	fmt.Println("4. 与Access Token完全不同，Access Token用于API调用认证")

	// 12. 演示素材上传（需要实际文件）
	fmt.Println("\n12. 演示素材上传...")
	fmt.Println("注意：素材上传需要实际文件，这里仅演示代码结构")

	// 创建测试文件
	testFile := createTestFile()
	if testFile != "" {
		defer os.Remove(testFile) // 清理测试文件

		fmt.Printf("上传测试文件: %s\n", testFile)
		uploadResp, err := app.UploadMaterialByPath(testFile, "file")
		if err != nil {
			log.Printf("上传文件失败: %v\n", err)
		} else {
			fmt.Printf("文件上传成功，媒体ID: %s\n", uploadResp.MediaId)
		}
	}

	// 13. 演示Token管理器功能
	fmt.Println("\n13. 演示Token管理器功能...")
	fmt.Println("注意：这里演示的是Access Token的管理，与验证Token不同")

	tokenManager := app.GetTokenManager()

	// 强制刷新Token
	fmt.Println("强制刷新访问令牌...")
	newToken, err := tokenManager.RefreshToken()
	if err != nil {
		log.Printf("刷新Token失败: %v\n", err)
	} else {
		fmt.Printf("Token刷新成功: %s\n", newToken[:10]+"...")
	}

	// 展示两种Token的区别
	fmt.Println("\n两种Token的区别说明:")
	fmt.Println("1. 验证Token (EnterWechat.token):")
	fmt.Println("   - 用途：验证企业微信回调请求的合法性")
	fmt.Println("   - 来源：企业微信管理后台配置")
	fmt.Println("   - 特点：静态配置，长期有效，手动设置")
	fmt.Println("   - 使用：SetToken()设置，VerifyURL()验证")
	fmt.Println("")
	fmt.Println("2. Access Token (TokenManager管理):")
	fmt.Println("   - 用途：API调用的身份凭证")
	fmt.Println("   - 来源：企业微信服务器动态生成")
	fmt.Println("   - 特点：动态生成，短期有效，自动管理")
	fmt.Println("   - 使用：所有HTTP请求的URL参数")

	fmt.Println("\n=== 企业微信应用使用示例完成 ===")
}

// createTestFile 创建测试文件
func createTestFile() string {
	content := "这是一个测试文件，用于演示文件上传功能。\n创建时间: " + time.Now().Format("2006-01-02 15:04:05")

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "wx_test_*.txt")
	if err != nil {
		log.Printf("创建测试文件失败: %v\n", err)
		return ""
	}

	// 写入内容
	if _, err := tmpFile.WriteString(content); err != nil {
		log.Printf("写入测试文件失败: %v\n", err)
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return ""
	}

	tmpFile.Close()
	return tmpFile.Name()
}

// 演示配置结构
func showConfigExample() {
	fmt.Println("\n=== 配置示例 ===")

	// 基本配置
	basicConfig := &wx.EnterWechatConfig{
		TokenManagerConfig: &wx.TokenManagerConfig{
			HTTPTimeout: 30 * time.Second,
			RetryConfig: &wx.RetryConfig{
				MaxRetries:  3,
				InitialWait: time.Second,
				MaxWait:     10 * time.Second,
				Multiplier:  2.0,
			},
		},
	}

	fmt.Printf("基本配置: %+v\n", basicConfig)

	// 自定义重试配置
	customRetryConfig := &wx.RetryConfig{
		MaxRetries:  5,
		InitialWait: 2 * time.Second,
		MaxWait:     20 * time.Second,
		Multiplier:  1.5,
	}

	fmt.Printf("自定义重试配置: %+v\n", customRetryConfig)

	// 默认配置
	fmt.Printf("默认重试配置: %+v\n", wx.DefaultRetryConfig)
}
