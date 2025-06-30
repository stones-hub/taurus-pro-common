package util

import (
	"testing"
	"time"
)

var (
	/*
		content = &EmailContent{
			Body:    "hello world",
			From:    "61647649@qq.com",
			Subject: "hello world",
		}

		info = EmailInfo{
			ConnectTimeout: 10 * time.Second,
			Host:           "smtp.qq.com",
			KeepAlive:      true,
			Password:       "xgrwilsumdpvcabb",
			Port:           587,
			SendTimeout:    10 * time.Second,
			Username:       "61647649@qq.com",
		}
	*/

	SSLInfo = EmailInfo{
		ConnectTimeout: 10 * time.Second,
		Host:           "smtp.qq.com",
		KeepAlive:      true,
		Password:       "xgrwilsumdpvcabb",
		Port:           465,
		SendTimeout:    10 * time.Second,
		Username:       "61647649@qq.com",
	}

	TSLInfo = EmailInfo{
		ConnectTimeout: 10 * time.Second,
		Host:           "smtp.qq.com",
		KeepAlive:      true,
		Password:       "xgrwilsumdpvcabb",
		Port:           587,
		SendTimeout:    10 * time.Second,
		Username:       "61647649@qq.com",
	}
)

func TestSendMail(t *testing.T) {

	/*
		err := SendMail(content, "yelei@3k.com", info)

		if err != nil {
			t.Errorf("send mail failed %v", err)
		}

	*/
}

func TestSendMultipleEmails(t *testing.T) {
	/*
		if err := SendMultipleEmails(content, []string{"yelei@3k.com", "yelei@qq.com"}, info); err != nil {
			t.Errorf("send mail failed %v", err)
		}

	*/
}

func TestSendMailWithTSL(t *testing.T) {
	/*
		if err := SendMailWithTSL(content, "yelei@3k.com", info); err != nil {
			t.Errorf("send mail failed %v", err)
		}

	*/
}

func TestSendMailWithSSL(t *testing.T) {
	/*
		if err := SendMailWithSSL(content, "yelei@3k.com", SSLInfo); err != nil {
			t.Errorf("send mail failed %v", err)
		}

	*/

}
