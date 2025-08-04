package main

import (
	"errors"

	"github.com/stones-hub/taurus-pro-common/pkg/util"
)

var (
	K3Wechat = util.NewEnterWechat(
		"ww245fc7dc10e8dc26",
		1000106,
		"摩羯",
		"-pQiscIUbJ6pXZC-O_q6us_4MCaXQVEphVM6S5w5IIQ",
		[]string{"yelei"})
)

// SendTextEnterWechatMsg 发送信息给企业微信
func SendTextEnterWechatMsg(title string, content string) error {
	var (
		err          error
		enWechatResp *util.EnterWechatResp
	)

	// 获取accessToken
	accessTokenResp, err := K3Wechat.AccessToken()
	if err != nil {
		return err
	}

	// 发送消息
	enWechatResp, err = K3Wechat.SendMarkdownMessage(&util.Markdown{Content: content}, accessTokenResp.AccessToken)

	if err == nil && enWechatResp.ErrCode == 0 {
		return nil
	}

	return errors.New(enWechatResp.ErrMsg)
}

func main() {
	SendTextEnterWechatMsg("test", "test")
}
