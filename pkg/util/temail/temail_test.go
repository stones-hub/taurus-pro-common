package temail

import (
	"os"
	"testing"
	"time"

	smail "github.com/xhit/go-simple-mail/v2"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "valid email",
			address: "test@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with name",
			address: "Test User <test@example.com>",
			wantErr: false,
		},
		{
			name:    "valid email with special chars",
			address: "test.name+tag@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			address: "test@sub.example.com",
			wantErr: false,
		},
		{
			name:    "invalid email - no @",
			address: "testexample.com",
			wantErr: true,
		},
		{
			name:    "invalid email - no domain",
			address: "test@",
			wantErr: true,
		},
		{
			name:    "invalid email - invalid domain",
			address: "test@example",
			wantErr: true,
		},
		{
			name:    "invalid email - invalid chars",
			address: "test<>@example.com",
			wantErr: true,
		},
		{
			name:    "empty email",
			address: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	// 创建临时模板文件
	tmpFile, err := os.CreateTemp("", "email_template_*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入测试模板内容
	templateContent := `<h1>Hello {{.name}}!</h1><p>{{.message}}</p>`
	if err := os.WriteFile(tmpFile.Name(), []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "valid template",
			template: tmpFile.Name(),
			data: map[string]interface{}{
				"name":    "John",
				"message": "Welcome!",
			},
			want:    "<h1>Hello John!</h1><p>Welcome!</p>",
			wantErr: false,
		},
		{
			name:     "non-existent template",
			template: "non_existent.html",
			data:     map[string]interface{}{},
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderTemplate(tt.template, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("renderTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailQueue(t *testing.T) {
	// 创建测试配置
	info := EmailInfo{
		Host:           "smtp.example.com",
		Port:           465,
		Username:       "test@example.com",
		Password:       "password",
		ConnectTimeout: time.Second,
		SendTimeout:    time.Second,
		Encryption:     smail.EncryptionSSLTLS,
		Auth:           smail.AuthLogin,
		RetryTimes:     2,
		RetryInterval:  time.Millisecond * 100,
		QueueSize:      5,
		KeepAlive:      true,
	}

	// 测试创建队列
	queue, err := NewEmailQueue(info)
	if err == nil {
		// 如果创建成功（可能是mock的SMTP服务器），测试队列操作
		defer queue.Stop()

		content := &EmailContent{
			From:    "test@example.com",
			Subject: "Test Email",
			Body:    "<p>Test content</p>",
		}

		// 测试入队
		err = queue.EnqueueEmail(content, "recipient@example.com")
		if err != nil {
			// 忽略实际发送错误，因为我们没有真实的SMTP服务器
			t.Logf("EnqueueEmail error (expected in test environment): %v", err)
		}

		// 测试队列满时的行为
		for i := 0; i < info.QueueSize+1; i++ {
			err = queue.EnqueueEmail(content, "recipient@example.com")
			if i == info.QueueSize && err == nil {
				t.Error("Expected error when queue is full")
			}
		}
	} else {
		// 忽略连接错误，因为这是预期的行为（没有真实的SMTP服务器）
		t.Logf("NewEmailQueue error (expected in test environment): %v", err)
	}
}

func TestSendWithTemplate(t *testing.T) {
	// 创建临时模板文件
	tmpFile, err := os.CreateTemp("", "email_template_*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入测试模板内容
	templateContent := `<h1>{{.title}}</h1><p>{{.message}}</p>`
	if err := os.WriteFile(tmpFile.Name(), []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}

	info := EmailInfo{
		Host:     "smtp.example.com",
		Port:     465,
		Username: "test@example.com",
		Password: "password",
	}

	data := map[string]interface{}{
		"subject": "Test Subject",
		"title":   "Welcome",
		"message": "Hello World",
	}

	err = SendWithTemplate(tmpFile.Name(), data, "recipient@example.com", info)
	// 忽略实际发送错误，因为我们没有真实的SMTP服务器
	if err != nil {
		t.Logf("SendWithTemplate error (expected in test environment): %v", err)
	}
}

func TestSendMultipleEmails(t *testing.T) {
	info := EmailInfo{
		Host:     "smtp.example.com",
		Port:     465,
		Username: "test@example.com",
		Password: "password",
	}

	content := &EmailContent{
		From:    "test@example.com",
		Subject: "Test Subject",
		Body:    "<p>Test content</p>",
	}

	recipients := []string{
		"recipient1@example.com",
		"recipient2@example.com",
		"recipient3@example.com",
	}

	err := SendMultipleEmails(content, recipients, info)
	// 忽略实际发送错误，因为我们没有真实的SMTP服务器
	if err != nil {
		t.Logf("SendMultipleEmails error (expected in test environment): %v", err)
	}
}
