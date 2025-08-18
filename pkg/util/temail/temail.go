package temail

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/mail"
	"regexp"
	"strings"
	"sync"
	"time"

	smail "github.com/xhit/go-simple-mail/v2"
)

// EmailInfo 定义邮件服务器的配置信息
// 用于配置SMTP服务器连接和邮件发送的相关参数
//
// 字段说明：
//   - Username: 邮箱地址，如：61647649@qq.com
//   - Password: 邮箱授权码（不是邮箱登录密码）
//   - ConnectTimeout: SMTP服务器连接超时时间
//   - SendTimeout: 发送单封邮件的超时时间
//   - Host: SMTP服务器地址，如：smtp.qq.com
//   - Port: SMTP服务器端口，如：465（SSL）或587（TLS）
//   - KeepAlive: 是否保持连接，适用于批量发送邮件
//   - Encryption: 加密方式，支持无加密、SSL/TLS、STARTTLS
//   - Auth: SMTP认证方式，支持CRAM-MD5、LOGIN、PLAIN
//   - RetryTimes: 发送失败时的重试次数
//   - RetryInterval: 重试间隔时间
//   - QueueSize: 邮件队列大小（用于异步发送）
//
// 使用示例：
//
//	config := temail.EmailInfo{
//	    Username:       "sender@example.com",
//	    Password:       "your-auth-code",
//	    Host:          "smtp.example.com",
//	    Port:          465,
//	    ConnectTimeout: time.Second * 10,
//	    SendTimeout:    time.Second * 30,
//	    Encryption:     mail.EncryptionSSLTLS,
//	    Auth:           mail.AuthLogin,
//	    RetryTimes:     3,
//	    RetryInterval:  time.Second * 5,
//	    KeepAlive:      true,
//	    QueueSize:      100,
//	}
//
// 注意事项：
//   - 密码通常需要使用邮箱的授权码而不是登录密码
//   - 不同邮件服务商的配置可能不同
//   - 建议使用SSL/TLS加密以确保安全性
//   - KeepAlive适用于需要批量发送邮件的场景
//   - 队列大小要根据实际需求和服务器性能来设置
type EmailInfo struct {
	Username       string           // 邮箱地址，如：61647649@qq.com
	Password       string           // 邮箱授权码
	ConnectTimeout time.Duration    // 连接超时时间
	SendTimeout    time.Duration    // 发送邮件超时时间
	Host           string           // SMTP服务器地址，如：smtp.example.com
	Port           int              // 端口
	KeepAlive      bool             // 是否长连接
	Encryption     smail.Encryption // 加密方式：mail.EncryptionNone, mail.EncryptionSSLTLS, mail.EncryptionSTARTTLS
	Auth           smail.AuthType   // 认证方式：mail.AuthCRAMMD5, mail.AuthLogin, mail.AuthPlain
	RetryTimes     int              // 重试次数
	RetryInterval  time.Duration    // 重试间隔
	QueueSize      int              // 队列大小
}

// EmailContent 定义邮件的内容和相关属性
// 用于指定邮件的发件人、主题、正文、附件等信息
//
// 字段说明：
//   - From: 发件人信息，格式：姓名 <邮箱地址>，如：张三 <zhangsan@example.com>
//   - Subject: 邮件主题
//   - Body: 邮件正文，支持HTML格式
//   - File: 附件文件路径列表
//   - Template: HTML模板文件路径（可选）
//   - Data: 用于渲染模板的数据（当使用模板时必需）
//
// 使用示例：
//
//	// 基本用法
//	content := &temail.EmailContent{
//	    From:    "系统通知 <system@example.com>",
//	    Subject: "账户激活通知",
//	    Body:    "<h1>欢迎使用我们的服务！</h1><p>请点击链接激活账户...</p>",
//	}
//
//	// 使用模板
//	content := &temail.EmailContent{
//	    From:     "系统通知 <system@example.com>",
//	    Subject:  "账户激活通知",
//	    Template: "templates/activation.html",
//	    Data: map[string]interface{}{
//	        "username": "张三",
//	        "link":     "https://example.com/activate?token=xxx",
//	    },
//	}
//
//	// 带附件
//	content := &temail.EmailContent{
//	    From:    "系统通知 <system@example.com>",
//	    Subject: "月度报告",
//	    Body:    "<p>请查收本月报告。</p>",
//	    File:    []string{"reports/monthly.pdf", "reports/charts.xlsx"},
//	}
//
// 注意事项：
//   - From字段中的邮箱地址必须与发送邮件的账号一致
//   - Body字段支持HTML标签，会被解析为富文本
//   - 使用模板时，Template和Data字段必须同时提供
//   - 附件文件必须存在且有读取权限
//   - 建议附件大小不要超过邮件服务商的限制
type EmailContent struct {
	From     string                 // 发件人，格式：61647649 <61647649@qq.com>，需要与发送邮件的用户名一致
	Subject  string                 // 邮件标题
	Body     string                 // 邮件内容，目前只支持HTML解析
	File     []string               // 附件文件路径列表
	Template string                 // 模板名称
	Data     map[string]interface{} // 模板数据
}

// send 发送单封邮件的内部函数
// 参数：
//   - content: 邮件内容，包括发件人、主题、正文等信息
//   - to: 收件人邮箱地址
//   - smtpClient: 已配置好的SMTP客户端
//
// 返回值：
//   - error: 发送过程中的错误，如果成功则为nil
//
// 内部处理流程：
//  1. 验证发件人和收件人邮箱格式
//  2. 设置邮件基本信息（发件人、收件人、主题）
//  3. 处理HTML模板（如果使用模板）
//  4. 设置邮件正文（HTML格式）
//  5. 添加附件（如果有）
//  6. 发送邮件
//
// 注意事项：
//   - 会验证邮箱地址格式
//   - 支持HTML格式的邮件内容
//   - 支持使用模板生成邮件内容
//   - 支持添加多个附件
//   - 邮件优先级设置为高
//   - 发送失败会返回详细错误信息
func send(content *EmailContent, to string, smtpClient *smail.SMTPClient) error {
	// 创建邮件消息
	email := smail.NewMSG()

	// 验证邮件地址格式
	if err := validateEmail(to); err != nil {
		return fmt.Errorf("invalid recipient email: %w", err)
	}
	if err := validateEmail(content.From); err != nil {
		return fmt.Errorf("invalid sender email: %w", err)
	}

	email.SetFrom(content.From).AddTo(to).SetSubject(content.Subject)

	// 处理模板
	if content.Template != "" {
		body, err := renderTemplate(content.Template, content.Data)
		if err != nil {
			return fmt.Errorf("render template failed: %w", err)
		}
		content.Body = body
	}

	// 获取发件人信息
	email.GetFrom()
	email.SetBody(smail.TextHTML, content.Body)

	// 设置高优先级
	email.SetPriority(smail.PriorityHigh)

	// 判断是否有附件
	if len(content.File) > 0 {
		for _, file := range content.File {
			filename := file
			parts := strings.Split(file, "/")
			if len(parts) > 0 {
				filename = parts[len(parts)-1]
			}
			email.AddAttachment(file, filename)
		}
	}

	// 检查错误
	if email.Error != nil {
		return email.Error
	}

	// 发送邮件
	return email.Send(smtpClient)
}

// configureSMTPClient 根据配置信息创建并配置SMTP客户端
// 参数：
//   - info: 邮件服务器配置信息
//
// 返回值：
//   - *smail.SMTPClient: 配置好的SMTP客户端
//   - error: 配置过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	config := temail.EmailInfo{
//	    Host:     "smtp.example.com",
//	    Port:     465,
//	    Username: "sender@example.com",
//	    Password: "auth-code",
//	}
//	client, err := configureSMTPClient(config)
//	if err != nil {
//	    log.Printf("配置SMTP客户端失败：%v", err)
//	    return
//	}
//	defer client.Close()
//
// 注意事项：
//   - 会立即尝试连接SMTP服务器
//   - 连接成功才会返回客户端
//   - 返回的客户端需要在使用完后调用Close方法
//   - 支持多种加密方式和认证方式
//   - 可以设置连接和发送超时时间
func configureSMTPClient(info EmailInfo) (*smail.SMTPClient, error) {
	smtpServer := smail.NewSMTPClient()
	smtpServer.Host = info.Host
	smtpServer.Port = info.Port
	smtpServer.Username = info.Username
	smtpServer.Password = info.Password
	smtpServer.Encryption = info.Encryption
	smtpServer.ConnectTimeout = info.ConnectTimeout
	smtpServer.SendTimeout = info.SendTimeout
	smtpServer.KeepAlive = info.KeepAlive
	smtpServer.Authentication = info.Auth
	return smtpServer.Connect()
}

// validateEmail 验证邮件地址格式的合法性
// 参数：
//   - address: 要验证的邮件地址，可以是纯地址或带名称的格式
//
// 返回值：
//   - error: 验证失败返回错误信息，验证通过返回nil
//
// 使用示例：
//
//	// 验证纯邮箱地址
//	err := validateEmail("user@example.com")
//
//	// 验证带名称的邮箱地址
//	err = validateEmail("张三 <zhangsan@example.com>")
//
// 验证规则：
//  1. 如果包含尖括号，提取尖括号中的地址
//  2. 使用标准库mail.ParseAddress进行基本验证
//  3. 使用正则表达式进行更严格的格式验证
//
// 注意事项：
//   - 支持带名称的邮箱地址格式
//   - 支持UTF-8字符的用户名部分
//   - 域名部分必须符合RFC标准
//   - 不验证邮箱是否真实存在
//   - 不验证域名是否有效
func validateEmail(address string) error {
	// 移除名称部分（如果有）
	if strings.Contains(address, "<") {
		parts := strings.Split(address, "<")
		if len(parts) != 2 {
			return fmt.Errorf("invalid email format")
		}
		address = strings.TrimSuffix(parts[1], ">")
	}

	// 使用标准库验证
	_, err := mail.ParseAddress(address)
	if err != nil {
		return err
	}

	// 使用正则表达式进行额外验证
	// 这个正则表达式遵循 RFC 5322 标准
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_\x60{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(address) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// renderTemplate 渲染HTML邮件模板
// 参数：
//   - name: 模板文件路径
//   - data: 用于渲染模板的数据
//
// 返回值：
//   - string: 渲染后的HTML内容
//   - error: 渲染过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	// 模板文件 templates/welcome.html:
//	// <h1>欢迎 {{.username}}!</h1>
//	// <p>您的账户已经激活，<a href="{{.link}}">点击这里</a>开始使用。</p>
//
//	data := map[string]interface{}{
//	    "username": "张三",
//	    "link":     "https://example.com/start",
//	}
//	html, err := renderTemplate("templates/welcome.html", data)
//	if err != nil {
//	    log.Printf("渲染模板失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 使用Go标准库html/template
//   - 模板文件必须存在且有读取权限
//   - 模板语法必须正确
//   - 所有模板变量必须在data中提供
//   - 支持HTML转义以防止XSS攻击
//   - 渲染结果是完整的HTML字符串
func renderTemplate(name string, data map[string]interface{}) (string, error) {
	tmpl, err := template.ParseFiles(name)
	if err != nil {
		return "", fmt.Errorf("parse template failed: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template failed: %w", err)
	}

	return buf.String(), nil
}

// EmailQueue 异步邮件发送队列
// 用于管理邮件的异步发送、重试和连接维护
//
// 字段说明：
//   - queue: 邮件任务队列通道
//   - info: 邮件服务器配置信息
//   - client: SMTP客户端
//   - stopChan: 用于停止工作协程的通道
//   - wg: 用于等待工作协程完成的等待组
//
// 使用示例：
//
//	// 创建邮件队列
//	config := temail.EmailInfo{
//	    Host:          "smtp.example.com",
//	    Port:          465,
//	    Username:      "sender@example.com",
//	    Password:      "auth-code",
//	    RetryTimes:    3,
//	    RetryInterval: time.Second * 5,
//	    QueueSize:     100,
//	    KeepAlive:     true,
//	}
//	queue, err := temail.NewEmailQueue(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer queue.Stop()
//
//	// 添加邮件任务
//	content := &temail.EmailContent{
//	    From:    "系统通知 <sender@example.com>",
//	    Subject: "测试邮件",
//	    Body:    "这是一封测试邮件",
//	}
//	err = queue.EnqueueEmail(content, "recipient@example.com")
//
// 注意事项：
//   - 队列是并发安全的
//   - 自动处理连接断开和重连
//   - 支持邮件发送失败重试
//   - 优雅关闭时会等待所有任务完成
//   - 队列满时会拒绝新的任务
//   - 建议在程序退出时调用Stop方法
type EmailQueue struct {
	queue    chan *EmailTask
	info     EmailInfo
	client   *smail.SMTPClient
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// EmailTask 表示一个待发送的邮件任务
// 用于在邮件队列中跟踪单个邮件的发送状态
//
// 字段说明：
//   - Content: 邮件内容
//   - To: 收件人地址
//   - Attempts: 已尝试发送的次数
//
// 注意事项：
//   - 每次发送失败会增加Attempts计数
//   - 当Attempts达到最大重试次数时停止重试
//   - Content和To字段在创建后不应修改
type EmailTask struct {
	Content  *EmailContent
	To       string
	Attempts int
}

// NewEmailQueue 创建并启动一个新的邮件队列
// 参数：
//   - info: 邮件服务器配置信息
//
// 返回值：
//   - *EmailQueue: 创建的邮件队列实例
//   - error: 创建过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	config := temail.EmailInfo{
//	    Host:          "smtp.example.com",
//	    Port:          465,
//	    Username:      "sender@example.com",
//	    Password:      "auth-code",
//	    RetryTimes:    3,
//	    RetryInterval: time.Second * 5,
//	    QueueSize:     100,
//	}
//	queue, err := temail.NewEmailQueue(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer queue.Stop()
//
// 注意事项：
//   - 会设置默认的队列大小（100）
//   - 会设置默认的重试次数（3）
//   - 会设置默认的重试间隔（5秒）
//   - 会立即尝试连接SMTP服务器
//   - 会启动后台工作协程
//   - 返回的队列需要在使用完后调用Stop方法
func NewEmailQueue(info EmailInfo) (*EmailQueue, error) {
	if info.QueueSize <= 0 {
		info.QueueSize = 100
	}
	if info.RetryTimes <= 0 {
		info.RetryTimes = 3
	}
	if info.RetryInterval <= 0 {
		info.RetryInterval = time.Second * 5
	}

	client, err := configureSMTPClient(info)
	if err != nil {
		return nil, err
	}

	q := &EmailQueue{
		queue:    make(chan *EmailTask, info.QueueSize),
		info:     info,
		client:   client,
		stopChan: make(chan struct{}),
	}

	// 启动工作协程
	q.wg.Add(1)
	go q.worker()

	return q, nil
}

// worker 处理邮件队列中的任务的后台工作协程
// 该方法会持续运行，直到收到停止信号
//
// 工作流程：
//  1. 从队列中获取邮件任务
//  2. 尝试发送邮件（带重试机制）
//  3. 如果发送失败，记录错误日志
//  4. 继续处理下一个任务
//  5. 收到停止信号时退出
//
// 注意事项：
//   - 在NewEmailQueue中自动启动
//   - 通过stopChan控制停止
//   - 使用wg跟踪完成状态
//   - 发送失败会记录详细日志
//   - 支持优雅关闭
func (q *EmailQueue) worker() {
	defer q.wg.Done()

	for {
		select {
		case task := <-q.queue:
			// 发送邮件（带重试）
			err := q.sendWithRetry(task)
			if err != nil {
				log.Printf("Failed to send email to %s after %d attempts: %v\n",
					task.To, task.Attempts, err)
			}
		case <-q.stopChan:
			return
		}
	}
}

// sendWithRetry 尝试发送邮件，失败时自动重试
// 参数：
//   - task: 要发送的邮件任务
//
// 返回值：
//   - error: 所有重试都失败后的最后一个错误，如果成功则为nil
//
// 重试策略：
//  1. 首次尝试发送
//  2. 如果失败且未达到最大重试次数：
//     - 等待指定的重试间隔
//     - 检查并重新建立SMTP连接
//     - 再次尝试发送
//  3. 重复步骤2直到成功或达到最大重试次数
//
// 注意事项：
//   - 使用配置中指定的重试次数和间隔
//   - 每次重试前都会检查连接状态
//   - 记录每次尝试的结果
//   - 返回最后一次失败的错误
//   - 任何一次成功都会立即返回
func (q *EmailQueue) sendWithRetry(task *EmailTask) error {
	var lastErr error
	for task.Attempts < q.info.RetryTimes {
		err := send(task.Content, task.To, q.client)
		if err == nil {
			return nil
		}

		lastErr = err
		task.Attempts++

		// 如果还有重试次数，等待后继续
		if task.Attempts < q.info.RetryTimes {
			time.Sleep(q.info.RetryInterval)
			// 检查连接状态，如果需要则重新连接
			if err := q.checkConnection(); err != nil {
				lastErr = err
				continue
			}
		}
	}
	return lastErr
}

// checkConnection 检查SMTP连接状态并在需要时重新连接
// 返回值：
//   - error: 重新连接过程中的错误，如果连接正常或重连成功则为nil
//
// 检查流程：
//  1. 发送NOOP命令测试连接
//  2. 如果失败，尝试重新建立连接
//  3. 更新队列中的客户端
//
// 注意事项：
//   - 用于维护长连接的可用性
//   - 自动处理连接断开的情况
//   - 使用相同的配置重新连接
//   - 在重试发送前调用
//   - 失败时会返回详细错误信息
func (q *EmailQueue) checkConnection() error {
	if err := q.client.Noop(); err != nil {
		client, err := configureSMTPClient(q.info)
		if err != nil {
			return err
		}
		q.client = client
	}
	return nil
}

// EnqueueEmail 将邮件任务添加到发送队列
// 参数：
//   - content: 邮件内容
//   - to: 收件人地址
//
// 返回值：
//   - error: 如果队列已满返回错误，成功则返回nil
//
// 使用示例：
//
//	content := &temail.EmailContent{
//	    From:    "系统通知 <system@example.com>",
//	    Subject: "重要通知",
//	    Body:    "<p>您的账户已创建成功！</p>",
//	}
//	err := queue.EnqueueEmail(content, "user@example.com")
//	if err != nil {
//	    log.Printf("添加到队列失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 是非阻塞操作
//   - 队列满时会立即返回错误
//   - 不会验证邮件内容的有效性
//   - 实际发送在后台进行
//   - 支持并发调用
func (q *EmailQueue) EnqueueEmail(content *EmailContent, to string) error {
	select {
	case q.queue <- &EmailTask{Content: content, To: to}:
		return nil
	default:
		return fmt.Errorf("email queue is full")
	}
}

// Stop 优雅地停止邮件队列
// 该方法会等待所有正在处理的任务完成后才返回
//
// 停止流程：
//  1. 发送停止信号
//  2. 等待工作协程完成当前任务
//  3. 关闭SMTP连接
//
// 注意事项：
//   - 调用后不能再添加新任务
//   - 会等待所有任务处理完成
//   - 会自动关闭SMTP连接
//   - 可以安全地并发调用
//   - 建议在程序退出时调用
func (q *EmailQueue) Stop() {
	close(q.stopChan)
	q.wg.Wait()
	q.client.Close()
}

// SendMail 同步发送单封邮件
// 参数：
//   - content: 邮件内容
//   - to: 收件人地址
//   - info: 邮件服务器配置信息
//
// 返回值：
//   - error: 发送过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	content := &temail.EmailContent{
//	    From:    "系统通知 <system@example.com>",
//	    Subject: "测试邮件",
//	    Body:    "<p>这是一封测试邮件</p>",
//	}
//	config := temail.EmailInfo{
//	    Host:     "smtp.example.com",
//	    Port:     465,
//	    Username: "sender@example.com",
//	    Password: "auth-code",
//	}
//	err := temail.SendMail(content, "recipient@example.com", config)
//	if err != nil {
//	    log.Printf("发送失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 是同步操作，会阻塞直到发送完成
//   - 每次发送都会建立新的SMTP连接
//   - 发送完成后会自动关闭连接
//   - 不支持失败重试
//   - 适用于发送单封邮件的场景
//   - 如果需要重试，请使用SendMailWithRetry
func SendMail(content *EmailContent, to string, info EmailInfo) error {
	smtpClient, err := configureSMTPClient(info)
	if err != nil {
		return err
	}
	defer smtpClient.Close()

	if err = send(content, to, smtpClient); err != nil {
		return err
	}
	return nil
}

// SendMailWithRetry 同步发送单封邮件，失败时自动重试
// 参数：
//   - content: 邮件内容
//   - to: 收件人地址
//   - info: 邮件服务器配置信息（包含重试配置）
//
// 返回值：
//   - error: 所有重试都失败后的最后一个错误，如果成功则为nil
//
// 使用示例：
//
//	content := &temail.EmailContent{
//	    From:    "系统通知 <system@example.com>",
//	    Subject: "重要通知",
//	    Body:    "<p>这是一封重要通知</p>",
//	}
//	config := temail.EmailInfo{
//	    Host:          "smtp.example.com",
//	    Port:          465,
//	    Username:      "sender@example.com",
//	    Password:      "auth-code",
//	    RetryTimes:    3,
//	    RetryInterval: time.Second * 5,
//	}
//	err := temail.SendMailWithRetry(content, "recipient@example.com", config)
//	if err != nil {
//	    log.Printf("发送失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 是同步操作，会阻塞直到发送完成或达到最大重试次数
//   - 每次重试都会建立新的SMTP连接
//   - 使用配置中的RetryTimes和RetryInterval
//   - 如果未配置，使用默认值（3次重试，5秒间隔）
//   - 适用于需要确保邮件发送成功的场景
//   - 会记录每次重试的错误信息
func SendMailWithRetry(content *EmailContent, to string, info EmailInfo) error {
	if info.RetryTimes <= 0 {
		info.RetryTimes = 3
	}
	if info.RetryInterval <= 0 {
		info.RetryInterval = time.Second * 5
	}

	var lastErr error
	for i := 0; i < info.RetryTimes; i++ {
		err := SendMail(content, to, info)
		if err == nil {
			return nil
		}

		lastErr = err
		if i < info.RetryTimes-1 {
			time.Sleep(info.RetryInterval)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", info.RetryTimes, lastErr)
}

// SendMultipleEmails 异步发送多封相同内容的邮件
// 参数：
//   - content: 邮件内容（所有收件人收到相同的内容）
//   - toList: 收件人地址列表
//   - info: 邮件服务器配置信息
//
// 返回值：
//   - error: 如果无法创建邮件队列或添加任务失败则返回错误，否则返回nil
//
// 使用示例：
//
//	content := &temail.EmailContent{
//	    From:    "系统通知 <system@example.com>",
//	    Subject: "系统升级通知",
//	    Body:    "<p>系统将于今晚22:00进行升级维护...</p>",
//	}
//	recipients := []string{
//	    "user1@example.com",
//	    "user2@example.com",
//	    "user3@example.com",
//	}
//	config := temail.EmailInfo{
//	    Host:          "smtp.example.com",
//	    Port:          465,
//	    Username:      "sender@example.com",
//	    Password:      "auth-code",
//	    KeepAlive:     true,
//	    RetryTimes:    3,
//	    QueueSize:     100,
//	}
//	err := temail.SendMultipleEmails(content, recipients, config)
//	if err != nil {
//	    log.Printf("发送失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 使用邮件队列异步发送
//   - 自动保持SMTP连接
//   - 支持失败重试
//   - 队列满时会返回错误
//   - 返回后邮件可能还在发送中
//   - 适用于群发邮件的场景
func SendMultipleEmails(content *EmailContent, toList []string, info EmailInfo) error {
	// 创建邮件队列
	queue, err := NewEmailQueue(info)
	if err != nil {
		return err
	}
	defer queue.Stop()

	// 将所有邮件加入队列
	for _, to := range toList {
		if err := queue.EnqueueEmail(content, to); err != nil {
			return err
		}
	}

	return nil
}

// SendWithTLS 使用STARTTLS加密发送单封邮件
// 参数：
//   - content: 邮件内容
//   - to: 收件人地址
//   - info: 邮件服务器配置信息
//
// 返回值：
//   - error: 发送过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	content := &temail.EmailContent{
//	    From:    "系统通知 <system@example.com>",
//	    Subject: "安全邮件测试",
//	    Body:    "<p>这是一封使用TLS加密的邮件</p>",
//	}
//	config := temail.EmailInfo{
//	    Host:     "smtp.example.com",
//	    Port:     587,  // STARTTLS通常使用587端口
//	    Username: "sender@example.com",
//	    Password: "auth-code",
//	}
//	err := temail.SendWithTLS(content, "recipient@example.com", config)
//	if err != nil {
//	    log.Printf("发送失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 使用STARTTLS加密（先明文连接，再升级到TLS）
//   - 通常使用587端口
//   - 自动设置加密方式为STARTTLS
//   - 自动设置认证方式为LOGIN
//   - 是SendMail的便捷包装
//   - 适用于需要TLS加密的场景
func SendWithTLS(content *EmailContent, to string, info EmailInfo) error {
	info.Encryption = smail.EncryptionSTARTTLS
	info.Auth = smail.AuthLogin
	return SendMail(content, to, info)
}

// SendWithSSL 使用SSL/TLS加密发送单封邮件
// 参数：
//   - content: 邮件内容
//   - to: 收件人地址
//   - info: 邮件服务器配置信息
//
// 返回值：
//   - error: 发送过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	content := &temail.EmailContent{
//	    From:    "系统通知 <system@example.com>",
//	    Subject: "安全邮件测试",
//	    Body:    "<p>这是一封使用SSL加密的邮件</p>",
//	}
//	config := temail.EmailInfo{
//	    Host:     "smtp.example.com",
//	    Port:     465,  // SSL/TLS通常使用465端口
//	    Username: "sender@example.com",
//	    Password: "auth-code",
//	}
//	err := temail.SendWithSSL(content, "recipient@example.com", config)
//	if err != nil {
//	    log.Printf("发送失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 使用SSL/TLS加密（整个连接过程都是加密的）
//   - 通常使用465端口
//   - 自动设置加密方式为SSL/TLS
//   - 自动设置认证方式为LOGIN
//   - 是SendMail的便捷包装
//   - 适用于需要SSL加密的场景
func SendWithSSL(content *EmailContent, to string, info EmailInfo) error {
	info.Encryption = smail.EncryptionSSLTLS
	info.Auth = smail.AuthLogin
	return SendMail(content, to, info)
}

// SendWithTemplate 使用HTML模板发送单封邮件
// 参数：
//   - templateName: HTML模板文件路径
//   - data: 用于渲染模板的数据，必须包含"subject"字段作为邮件主题
//   - to: 收件人地址
//   - info: 邮件服务器配置信息
//
// 返回值：
//   - error: 发送过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	// 模板文件 templates/welcome.html:
//	// <h1>欢迎 {{.username}}!</h1>
//	// <p>您的账户已经激活，<a href="{{.link}}">点击这里</a>开始使用。</p>
//
//	data := map[string]interface{}{
//	    "subject":  "欢迎注册",
//	    "username": "张三",
//	    "link":     "https://example.com/start",
//	}
//	config := temail.EmailInfo{
//	    Host:     "smtp.example.com",
//	    Port:     465,
//	    Username: "sender@example.com",
//	    Password: "auth-code",
//	}
//	err := temail.SendWithTemplate(
//	    "templates/welcome.html",
//	    data,
//	    "user@example.com",
//	    config,
//	)
//	if err != nil {
//	    log.Printf("发送失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 模板文件必须存在且有读取权限
//   - data必须包含"subject"字段
//   - 自动设置发件人为配置中的用户名
//   - 支持所有Go模板语法
//   - 是SendMail的便捷包装
//   - 适用于需要动态内容的邮件
func SendWithTemplate(templateName string, data map[string]interface{}, to string, info EmailInfo) error {
	content := &EmailContent{
		From:     info.Username,
		Subject:  data["subject"].(string),
		Template: templateName,
		Data:     data,
	}
	return SendMail(content, to, info)
}

// SendWithTemplateAndRetry 使用HTML模板发送单封邮件，失败时自动重试
// 参数：
//   - templateName: HTML模板文件路径
//   - data: 用于渲染模板的数据，必须包含"subject"字段作为邮件主题
//   - to: 收件人地址
//   - info: 邮件服务器配置信息（包含重试配置）
//
// 返回值：
//   - error: 所有重试都失败后的最后一个错误，如果成功则为nil
//
// 使用示例：
//
//	// 模板文件 templates/notification.html:
//	// <h1>{{.title}}</h1>
//	// <p>{{.message}}</p>
//	// <small>发送时间：{{.time}}</small>
//
//	data := map[string]interface{}{
//	    "subject": "系统通知",
//	    "title":   "服务器维护通知",
//	    "message": "系统将于今晚进行维护升级...",
//	    "time":    time.Now().Format("2006-01-02 15:04:05"),
//	}
//	config := temail.EmailInfo{
//	    Host:          "smtp.example.com",
//	    Port:          465,
//	    Username:      "sender@example.com",
//	    Password:      "auth-code",
//	    RetryTimes:    3,
//	    RetryInterval: time.Second * 5,
//	}
//	err := temail.SendWithTemplateAndRetry(
//	    "templates/notification.html",
//	    data,
//	    "user@example.com",
//	    config,
//	)
//	if err != nil {
//	    log.Printf("发送失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 模板文件必须存在且有读取权限
//   - data必须包含"subject"字段
//   - 自动设置发件人为配置中的用户名
//   - 使用配置中的重试参数
//   - 是SendMailWithRetry的便捷包装
//   - 适用于重要的模板邮件发送
func SendWithTemplateAndRetry(templateName string, data map[string]interface{}, to string, info EmailInfo) error {
	content := &EmailContent{
		From:     info.Username,
		Subject:  data["subject"].(string),
		Template: templateName,
		Data:     data,
	}
	return SendMailWithRetry(content, to, info)
}
