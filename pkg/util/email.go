// Copyright (c) 2025 Taurus Team. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: yelei
// Email: 61647649@qq.com
// Date: 2025-06-13

package util

import (
	"log"
	"strings"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type EmailInfo struct {
	Username       string          // 61647649@qq.com
	Password       string          // 邮箱授权码
	ConnectTimeout time.Duration   // 链接超时时间
	SendTimeout    time.Duration   // 发送邮件超时时间
	Host           string          // smtp.example.com ， 邮件服务器地址
	Port           int             // 端口
	KeepAlive      bool            // 是否长链
	Encryption     mail.Encryption // 加密方式 encryption: mail.EncryptionNone, mail.EncryptionSSLTLS, mail.EncryptionSTARTTLS, 其他的已经废弃
	Auth           mail.AuthType   // 认证方式 mail.AuthCRAMMD5, mail.AuthLogin, mail.AuthPlain
}

// Content 邮件内容
type EmailContent struct {
	From    string   // 来源 61647649 <61647649@qq.com>  需要跟发送邮件的用户名一致
	Subject string   // 标题
	Body    string   // 内容，目前只支持html解析
	File    []string // 内容带附件
}

func send(content *EmailContent, to string, smtpClient *mail.SMTPClient) error {
	//Create the email message
	email := mail.NewMSG()

	email.SetFrom(content.From).AddTo(to).SetSubject(content.Subject)

	//Get from each mail
	email.GetFrom()
	email.SetBody(mail.TextHTML, content.Body)

	//Send with high priority
	email.SetPriority(mail.PriorityHigh)

	// 判断是否有File来发送附件
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

	// always check error after send
	if email.Error != nil {
		return email.Error
	}

	//Pass the client to the email message to send it
	return email.Send(smtpClient)
}

// configureSMTPClient configure the SMTP client
// encryption: mail.EncryptionNone, mail.EncryptionSSLTLS, mail.EncryptionSTARTTLS, 其他的已经废弃
// 高加密方式,使用mail.EncryptionSSLTLS，低加密方式，使用mail.EncryptionSTARTTLS
// 安全性：AuthCRAMMD5 > AuthLogin > AuthPlain
// 兼容性：AuthPlain和AuthLogin通常比AuthCRAMMD5更广泛支持。
// 使用建议：在加密连接（如TLS或SSL）上使用AuthPlain或AuthLogin，以确保凭据的安全性。如果服务器支持并且需要更高的安全性，可以选择AuthCRAMMD5。
func configureSMTPClient(info EmailInfo) (*mail.SMTPClient, error) {
	smtpServer := mail.NewSMTPClient()
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

// SendMail with one email, keepAlive is false
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

// SendMultipleEmails with multiple emails, keepAlive is true and Noop every 30 seconds
func SendMultipleEmails(content *EmailContent, toList []string, info EmailInfo) error {
	smtpClient, err := configureSMTPClient(info)
	if err != nil {
		return err
	}
	defer smtpClient.Close()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	done := make(chan bool)

	// start a goroutine to send NOOP command every 10 seconds
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := smtpClient.Noop()
				if err != nil {
					log.Printf("Failed to send NOOP: %v \n", err)
					return
				}
			}
		}
	}()

	// 发送邮件
	for _, to := range toList {
		err = send(content, to, smtpClient)
		if err != nil {
			log.Printf("Failed to send email to %s: %v \n", to, err)
			done <- true
			return err
		}
	}

	// 发送完成后停止NOOP goroutine
	done <- true
	return nil
}

// SendWithTLS 使用TLS加密发送邮件 安全性：AuthCRAMMD5 > AuthLogin > AuthPlain
func SendWithTLS(content *EmailContent, to string, info EmailInfo) error {
	info.Encryption = mail.EncryptionSTARTTLS
	info.Auth = mail.AuthLogin
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

// SendWithSSL 使用SSL加密发送邮件 安全性：AuthCRAMMD5 > AuthLogin > AuthPlain
func SendWithSSL(content *EmailContent, to string, info EmailInfo) error {
	info.Encryption = mail.EncryptionSSLTLS
	info.Auth = mail.AuthLogin
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
