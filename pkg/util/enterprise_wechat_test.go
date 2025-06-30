package util

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var (
	filePath  = "/test/files/8956df30-757d-49b7-8ec1-f14421d0ba12.png"
	videoPath = "/test/files/df5fc7a6d596dcab2ab4ed613bbe613c.mov"
	// voicePath = "/test/files/d86c35d42b48501fefb53a63d79c99d8.mp3"
)

func getFilePath(file string) string {
	basePath, _ := os.Getwd()
	file = filepath.Dir(filepath.Dir(basePath)) + file
	return file
}

func new() *EnterWechat {
	// 企业ID： ww245fc7dc10e8dc26
	// Secret : -pQiscIUbJ6pXZC-O_q6us_4MCaXQVEphVM6S5w5IIQ
	// 应用ID： 1000106
	return NewEnterWechat(
		"ww245fc7dc10e8dc26",
		1000106,
		"摩羯监控",
		"-pQiscIUbJ6pXZC-O_q6us_4MCaXQVEphVM6S5w5IIQ",
		[]string{"yelei"})
}

func accessToken() string {
	resp, _ := new().AccessToken()
	return resp.AccessToken
}

func uploadFile(filePath string, fileType string) string {
	e, err := new().UploadMaterialByPath(filePath, fileType, accessToken())
	if err != nil || e.ErrCode != 0 {
		fmt.Println(err, e)
		return ""
	}
	return e.MediaId
}

func TestEnterWechat_SendTextMessage(t *testing.T) {
	if res, e := new().SendTextMessage(&Text{Content: "hello world"}, accessToken()); e != nil || res.ErrCode != 0 {
		t.Errorf("send text message failed, %v, %v", e, res.ErrMsg)
	} else {
		t.Logf("send text message success, %v", res)
	}
}

func TestEnterWechat_SendMarkdownMessage(t *testing.T) {

	if res, e := new().SendMarkdownMessage(&Markdown{Content: "您的会议室已经预定，稍后会同步到`邮箱`" +
		">**事项详情**" +
		">事 项：<font color=\"warning\">开会</font>" +
		">时 间：<font color=\"comment\">2018年5月18日9:00-11:00</font>" +
		">地 点：<font color=\"comment\">3K游戏</font>" +
		">" +
		">" +
		">请准时参加会议。" +
		">" +
		">" +
		">如需修改会议信息，请点击：[修改会议信息](https://work.weixin.qq.com)"}, accessToken()); e != nil || res.ErrCode != 0 {
		t.Errorf("send text message failed, %v, %v", e, res.ErrMsg)
	} else {
		t.Logf("send text message success, %v", res)
	}

}

func TestEnterWechat_SendImageMessage(t *testing.T) {

	if res, e := new().SendImageMessage(&Image{MediaId: uploadFile(getFilePath(filePath), "image")}, accessToken()); e != nil || res.ErrCode != 0 {
		t.Errorf("send image message failed, %v, %v", e, res.ErrMsg)
	} else {
		t.Logf("send image message success, %v", res)
	}
}

func TestEnterWechat_SendNewsMessage(t *testing.T) {
	if res, e := new().SendNewsMessage(&News{
		Articles: []Article{
			{
				Title:       "图文消息测试News",
				Description: "关联小程序的图文消息测试",
				Url:         "http://www.baidu.com",
				PicUrl:      "http://res.mail.qq.com/node/ww/wwopenmng/images/independent/doc/test_pic_msg1.png",
				AppId:       "",
				PagePath:    "",
			},
		},
	}, accessToken()); e != nil || res.ErrCode != 0 {
		t.Errorf("send news message failed, %v, %v", e, res.ErrMsg)
	} else {
		t.Logf("send news message success, %v", res)
	}
}

// 等找到amr文件格式的语音再测试吧
func TestEnterWechat_SendVoiceMessage(t *testing.T) {
	/*
		if res, e := new().SendVoiceMessage(&Voice{MediaId: uploadFile(getFilePath(voicePath), "voice")}, accessToken()); e != nil || res.ErrCode != 0 {
			t.Errorf("send voice message failed, %v, %v", e, res.ErrMsg)
		} else {
			t.Logf("send voice message success, %v", res)
		}

	*/
}

func TestEnterWechat_SendVideoMessage(t *testing.T) {
	if res, e := new().SendVideoMessage(&Video{MediaId: uploadFile(getFilePath(videoPath), "video")}, accessToken()); e != nil || res.ErrCode != 0 {
		t.Errorf("send video message failed, %v, %v", e, res.ErrMsg)
	} else {
		t.Logf("send video message success, %v", res)
	}

}

func TestEnterWechat_SendTextCardMessage(t *testing.T) {

	if res, e := new().SendTextCardMessage(&TextCard{
		Title:       "文本卡片测试",
		Description: "这是一个测试",
		Url:         "http://res.mail.qq.com/node/ww/wwopenmng/images/independent/doc/test_pic_msg1.png",
		BtnTxt:      "更多",
	}, accessToken()); e != nil || res.ErrCode != 0 {
		t.Errorf("send text card message failed, %v, %v", e, res.ErrMsg)
	} else {
		t.Logf("send text card message success, %v", res)
	}
}

func TestEnterWechat_SendFileMessage(t *testing.T) {
	if res, e := new().SendFileMessage(&File{MediaId: uploadFile(getFilePath(filePath), "file")}, accessToken()); e != nil || res.ErrCode != 0 {
		t.Errorf("send file message failed, %v, %v", e, res.ErrMsg)
	} else {
		t.Logf("send file message success, %v", res)
	}
}
