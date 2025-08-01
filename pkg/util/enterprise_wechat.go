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
	"encoding/json"
	"encoding/xml"
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/stones-hub/taurus-pro-common/pkg/util/secure"
	"github.com/stones-hub/taurus-pro-common/pkg/util/upload"
)

/*
企业微信应用对接
https://developer.work.weixin.qq.com/document/path/90254
*/

type EnterWechat struct {
	Corpid  string   // 企业ID
	AgentId int      // 应用ID
	Name    string   // 应用名
	Secret  string   // 应用密钥
	SendUid []string // 可接受推送消息的用户账户
}

const (
	// ACCESS_TOKEN_KEY 存在缓存的key 名
	ACCESS_TOKEN_KEY = "access_token_key"

	// ACCESS_TOKEN_API 获取token的api
	ACCESS_TOKEN_API = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"

	// SEND_MESSAGE_API 发送消息给用户 api
	SEND_MESSAGE_API = "https://qyapi.weixin.qq.com/cgi-bin/message/send"

	// UPLOAD_MEDIA_API 上传临时素材到企业微信的api
	UPLOAD_MEDIA_API = "https://qyapi.weixin.qq.com/cgi-bin/media/upload"
)

// NewEnterWechat 创建一个企业微信应用实例,
// Corpid 是企业ID, AgentId 是应用ID, Name 是应用名称，Secret 是应用密钥, SendUid 是推送消息的用户ID
func NewEnterWechat(Corpid string, AgentId int, Name string, Secret string, SendUid []string) *EnterWechat {
	return &EnterWechat{
		Corpid:  Corpid,
		AgentId: AgentId,
		Name:    Name,
		Secret:  Secret,
		SendUid: SendUid,
	}
}

// EnterWechatResp 统一所有的企业微信返回值
type EnterWechatResp struct {
	ErrCode     int    `json:"errcode"`      // 返回码
	ErrMsg      string `json:"errmsg"`       // 返回信息
	AccessToken string `json:"access_token"` // 获取到的凭证
	ExpiresIn   int    `json:"expires_in"`   // 凭证有效期

	InValidUser    string `json:"invaliduser"`    // 不合法的userid，不区分大小写，统一转为小写
	InValidParty   string `json:"invalidparty"`   // 不合法的partyid，不区分大小写，统一转为小写
	InValidTag     string `json:"invalidtag"`     // 不合法的tagid，不区分大小写，统一转为小写
	UnlicensedUser string `json:"unlicenseduser"` // 没有基础接口许可(包含已过期)的userid
	MsgId          string `json:"msgid"`          // 消息id, 可用于撤回
	ResponseCode   string `json:"response_code"`  // 仅消息类型为"按钮交互型"，"投票选择型"和"多项选择型"的模板卡片消息返回 , 没实现类似消息的发送，没啥用

	Type      string `json:"type"`
	MediaId   string `json:"media_id"`
	CreatedAt string `json:"created_at"`
}

func (r *EnterWechatResp) String() string {
	if b, err := json.Marshal(r); err != nil {
		return err.Error()
	} else {
		return string(b)
	}
}

// GenEnWechatAccessTokenKey 每个企业的每个应用，都应该有独立的accesstoken存储KEY
func GenEnWechatAccessTokenKey(e *EnterWechat) string {
	return secure.MD5([]byte(ACCESS_TOKEN_KEY + e.Corpid + e.Name + e.Secret))
}

func (e *EnterWechat) AccessToken() (*EnterWechatResp, error) {
	var (
		params          = make(map[string]string)
		err             error
		resp            *http.Response
		b               []byte
		enterWechatResp EnterWechatResp
	)

	params["corpid"] = e.Corpid
	params["corpsecret"] = e.Secret

	if resp, err = HttpRequest(ACCESS_TOKEN_API, "GET", nil, params, nil, DefaultTimeout); err != nil {

		return nil, err
	}

	if b, err = ReadResponse(resp); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(b, &enterWechatResp); err != nil {
		return nil, err
	}

	return &enterWechatResp, nil
}

// EnterWeChatReadMessage  text, image, voice, video 只用于支持这几种类型的消息, 读取消息
type EnterWeChatReadMessage struct {
	ToUserName   string `json:"to_user_name,omitempty" xml:"ToUserName"`     // 企业微信CorpID 企业ID
	FromUserName string `json:"from_user_name,omitempty" xml:"FromUserName"` // 成员UserID 用户账户
	CreateTime   string `json:"create_time,omitempty" xml:"CreateTime"`      // 消息创建时间
	MsgType      string `json:"msg_type,omitempty" xml:"MsgType"`            // 消息类型 text, image, voice, video
	Content      string `json:"content,omitempty" xml:"Content"`             // 消息内容 <文本消息特有>
	PicUrl       string `json:"pic_url,omitempty" xml:"PicUrl"`              // 图片消息链接地址 <图片消息特有>
	MediaId      string `json:"media_id,omitempty" xml:"MediaId"`            // 语音媒体文件id，可以调用获取媒体文件接口拉取数据，仅三天内有效 <图片消息特有> <语音消息特有> <视频消息特有>
	Format       string `json:"format,omitempty" xml:"Format"`               // 语音格式，如amr，speex等 <语音消息特有>
	ThumbMediaId string `json:"thumb_media_id,omitempty" xml:"ThumbMediaId"` // 视频消息缩略图的媒体id，可以调用获取媒体文件接口拉取数据，仅三天内有效 <视频消息特有>
	MsgId        string `json:"msg_id,omitempty" xml:"MsgId"`                // 消息ID
	AgentID      string `json:"agent_id,omitempty" xml:"AgentID"`            // 企业应用的id
}

func (e *EnterWeChatReadMessage) String() string {
	if b, err := json.Marshal(e); err != nil {
		return err.Error()
	} else {
		return string(b)
	}
}

// ReadMessage 接收消息为xml， 生成结构体, 由于接收消息需要配置线上地址，不同的应用配置不一样，所以不做实现，只做转换，方便后续逻辑处理
func (e *EnterWechat) ReadMessage(buf []byte) (*EnterWeChatReadMessage, error) {
	var (
		messsage EnterWeChatReadMessage
		err      error
	)
	if err = xml.Unmarshal(buf, &messsage); err != nil {
		return nil, err
	}
	return &messsage, nil
}

// EnterWeChatSendMessage  企业微信应用发送消息包, touser、toparty、totag不能同时为空，后面不再强调。, 其他消息类型暂不介入
type EnterWeChatSendMessage struct {
	ToUser  string `json:"touser"`  //指定接收消息的成员，成员ID列表（多个接收者用'分隔，最多支持1000个）。 特殊情况：指定为"@all"，则向该企业应用的全部成员发送
	ToParty string `json:"toparty"` // 指定接收消息的部门，部门ID列表，多个接收者用'分隔，最多支持100个。
	ToTag   string `json:"totag"`   // 指定接收消息的标签，标签ID列表，多个接收者用'分隔，最多支持100个。
	ToAll   int    `json:"toall"`
	MsgType string `json:"msgtype"` // 消息类型 text, image, voice, video
	AgentId int    `json:"agentid"` // 企业应用ID

	Text     Text     `json:"text"`     // 文本消息
	Markdown Markdown `json:"markdown"` // markdown 消息
	Image    Image    `json:"image"`    // 图片消息
	Voice    Voice    `json:"voice"`    // 语言消息
	Video    Video    `json:"video"`    //  视频消息
	File     File     `json:"file"`     // 文件消息
	TextCard TextCard `json:"textcard"` // 文本卡片消息
	News     News     `json:"news"`     // 图文消息

	Safe                   int `json:"safe"`                     // 表示是否是保密消息，0表示可对外分享，1表示不能分享且内容显示水印，默认为0
	EnableIdTrans          int `json:"enable_id_trans"`          // 表示是否开启id转译，0表示否，1表示是，默认0
	EnableDuplicateCheck   int `json:"enable_duplicate_check"`   // 表示是否开启重复消息检查，0表示否，1表示是，默认0
	DuplicateCheckInterval int `json:"duplicate_check_interval"` // 表示是否重复消息检查的时间间隔，默认1800s，最大不超过4小时
}

type Text struct {
	Content string `json:"content,omitempty"`
}

type Markdown struct {
	Content string `json:"content"`
}

type Image struct {
	MediaId string `json:"media_id,omitempty"`
}

type Voice struct {
	MediaId string `json:"media_id,omitempty"`
}

type Video struct {
	MediaId     string `json:"media_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type File struct {
	MediaId string `json:"media_id"`
}

type TextCard struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
	BtnTxt      string `json:"btntxt,omitempty"`
}

type News struct {
	Articles []Article `json:"articles,omitempty"`
}

type Article struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
	PicUrl      string `json:"picUrl,omitempty"`
	AppId       string `json:"appid,omitempty"`    // "wx123123123123123" 微信小程序小程序appid
	PagePath    string `json:"pagepath,omitempty"` // "pages/index?userid=zhangsan&orderid=123123123" 点击消息卡片后的小程序页面，最长128字节，仅限本小程序内的页面。appid和pagepath必须同时填写
}

func (e *EnterWeChatSendMessage) String() string {
	if b, err := json.Marshal(e); err != nil {
		return err.Error()
	} else {
		return string(b)
	}
}

func (e *EnterWechat) SendMessageToUsers(message *EnterWeChatSendMessage, users []string, token string) (*EnterWechatResp, error) {

	var (
		err             error
		params          = make(map[string]string)
		resp            *http.Response
		b               []byte
		enterWechatResp EnterWechatResp
	)

	params["access_token"] = token

	message.ToUser = strings.Join(users, "|")
	message.AgentId = e.AgentId
	message.EnableDuplicateCheck = 1
	message.DuplicateCheckInterval = 60
	if message.MsgType == "news" {
		message.Safe = 0
	}

	if resp, err = HttpRequest(SEND_MESSAGE_API, "POST", nil, params, message, DefaultTimeout); err != nil {
		return nil, err
	}

	if b, err = ReadResponse(resp); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(b, &enterWechatResp); err != nil {
		return nil, err
	}

	return &enterWechatResp, nil
}

// SendTextMessage 发送文本消息
func (e *EnterWechat) SendTextMessage(text *Text, token string) (*EnterWechatResp, error) {
	return e.sendMessage(&EnterWeChatSendMessage{
		ToUser:                 strings.Join(e.SendUid, "|"),
		ToParty:                "",
		ToTag:                  "",
		ToAll:                  0,
		MsgType:                "text",
		AgentId:                e.AgentId,
		Text:                   *text,
		Image:                  Image{},
		Voice:                  Voice{},
		Video:                  Video{},
		File:                   File{},
		TextCard:               TextCard{},
		News:                   News{},
		Safe:                   1,
		EnableIdTrans:          0,
		EnableDuplicateCheck:   1,  //  开启重复消息检查
		DuplicateCheckInterval: 60, // 检查时间60s
	}, token)
}

func (e *EnterWechat) SendMarkdownMessage(markdown *Markdown, token string) (*EnterWechatResp, error) {

	return e.sendMessage(&EnterWeChatSendMessage{
		ToUser:                 strings.Join(e.SendUid, "|"),
		ToParty:                "",
		ToTag:                  "",
		ToAll:                  0,
		MsgType:                "markdown",
		AgentId:                e.AgentId,
		Markdown:               *markdown,
		Text:                   Text{},
		Image:                  Image{},
		Voice:                  Voice{},
		Video:                  Video{},
		File:                   File{},
		TextCard:               TextCard{},
		News:                   News{},
		Safe:                   0,
		EnableIdTrans:          0,
		EnableDuplicateCheck:   1,  //  开启重复消息检查
		DuplicateCheckInterval: 60, // 检查时间60s
	}, token)

}

// SendImageMessage 发送图片消息
func (e *EnterWechat) SendImageMessage(image *Image, token string) (*EnterWechatResp, error) {
	return e.sendMessage(&EnterWeChatSendMessage{
		ToUser:                 strings.Join(e.SendUid, "|"),
		ToParty:                "",
		ToTag:                  "",
		ToAll:                  0,
		MsgType:                "image",
		AgentId:                e.AgentId,
		Text:                   Text{},
		Image:                  *image,
		Voice:                  Voice{},
		Video:                  Video{},
		File:                   File{},
		TextCard:               TextCard{},
		News:                   News{},
		Safe:                   1,
		EnableIdTrans:          0,
		EnableDuplicateCheck:   1,
		DuplicateCheckInterval: 60,
	}, token)

}

// SendNewsMessage 发送图文消息
func (e *EnterWechat) SendNewsMessage(news *News, token string) (*EnterWechatResp, error) {
	return e.sendMessage(&EnterWeChatSendMessage{
		ToUser:                 strings.Join(e.SendUid, "|"),
		ToParty:                "",
		ToTag:                  "",
		ToAll:                  0,
		MsgType:                "news",
		AgentId:                e.AgentId,
		Text:                   Text{},
		Image:                  Image{},
		Voice:                  Voice{},
		Video:                  Video{},
		File:                   File{},
		TextCard:               TextCard{},
		News:                   *news,
		Safe:                   0, // 图文消息不可以发送safe=1的消息
		EnableIdTrans:          0,
		EnableDuplicateCheck:   1,
		DuplicateCheckInterval: 60,
	}, token)
}

// SendVoiceMessage 发送语音消息
func (e *EnterWechat) SendVoiceMessage(voice *Voice, token string) (*EnterWechatResp, error) {
	return e.sendMessage(&EnterWeChatSendMessage{
		ToUser:                 strings.Join(e.SendUid, "|"),
		ToParty:                "",
		ToTag:                  "",
		ToAll:                  0,
		MsgType:                "voice",
		AgentId:                e.AgentId,
		Text:                   Text{},
		Image:                  Image{},
		Voice:                  *voice,
		Video:                  Video{},
		File:                   File{},
		TextCard:               TextCard{},
		News:                   News{},
		Safe:                   1,
		EnableIdTrans:          0,
		EnableDuplicateCheck:   1,
		DuplicateCheckInterval: 60,
	}, token)

}

// SendVideoMessage 发送视频消息
func (e *EnterWechat) SendVideoMessage(video *Video, token string) (*EnterWechatResp, error) {

	return e.sendMessage(&EnterWeChatSendMessage{
		ToUser:                 strings.Join(e.SendUid, "|"),
		ToParty:                "",
		ToTag:                  "",
		ToAll:                  0,
		MsgType:                "video",
		AgentId:                e.AgentId,
		Text:                   Text{},
		Image:                  Image{},
		Voice:                  Voice{},
		Video:                  *video,
		File:                   File{},
		TextCard:               TextCard{},
		News:                   News{},
		Safe:                   1,
		EnableIdTrans:          0,
		EnableDuplicateCheck:   1,
		DuplicateCheckInterval: 60,
	}, token)
}

// SendTextCardMessage 发送卡片消息
func (e *EnterWechat) SendTextCardMessage(card *TextCard, token string) (*EnterWechatResp, error) {
	return e.sendMessage(&EnterWeChatSendMessage{
		ToUser:                 strings.Join(e.SendUid, "|"),
		ToParty:                "",
		ToTag:                  "",
		ToAll:                  0,
		MsgType:                "textcard",
		AgentId:                e.AgentId,
		Text:                   Text{},
		Image:                  Image{},
		Voice:                  Voice{},
		Video:                  Video{},
		File:                   File{},
		TextCard:               *card,
		News:                   News{},
		Safe:                   1,
		EnableIdTrans:          0,
		EnableDuplicateCheck:   1,
		DuplicateCheckInterval: 60,
	}, token)

}

// SendFileMessage 发送文件消息
func (e *EnterWechat) SendFileMessage(file *File, token string) (*EnterWechatResp, error) {
	return e.sendMessage(&EnterWeChatSendMessage{
		ToUser:                 strings.Join(e.SendUid, "|"),
		ToParty:                "",
		ToTag:                  "",
		ToAll:                  0,
		MsgType:                "file",
		AgentId:                e.AgentId,
		Text:                   Text{},
		Image:                  Image{},
		Voice:                  Voice{},
		Video:                  Video{},
		File:                   *file,
		TextCard:               TextCard{},
		News:                   News{},
		Safe:                   1,
		EnableIdTrans:          0,
		EnableDuplicateCheck:   1,
		DuplicateCheckInterval: 60,
	}, token)
}

// SendMessage 发送消息
func (e *EnterWechat) sendMessage(message *EnterWeChatSendMessage, token string) (*EnterWechatResp, error) {

	var (
		err             error
		params          = make(map[string]string)
		resp            *http.Response
		b               []byte
		enterWechatResp EnterWechatResp
	)

	params["access_token"] = token

	if resp, err = HttpRequest(SEND_MESSAGE_API, "POST", nil, params, message, DefaultTimeout); err != nil {
		return nil, err
	}

	if b, err = ReadResponse(resp); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(b, &enterWechatResp); err != nil {
		return nil, err
	}

	return &enterWechatResp, nil
}

// UploadMaterialByPath 上传临时素材,
// filePath 本地文件地址
// materialType: image, voice, video, file   素材类型
func (e *EnterWechat) UploadMaterialByPath(fp string, materialType string, token string) (*EnterWechatResp, error) {
	var (
		resp            []byte
		err             error
		enterWechatResp EnterWechatResp
	)

	if materialType == "voice" {
		if filepath.Ext(fp) != ".amr" {
			return nil, errors.New("only support amr file")
		}
	}

	resp, err = upload.UploadFile2Remote(UPLOAD_MEDIA_API, map[string]string{
		"access_token": token,
		"type":         materialType,
	}, fp, "media")

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp, &enterWechatResp)
	if err != nil {
		return nil, err
	}
	return &enterWechatResp, nil
}

// UploadMaterialByFile 上传临时素材,
// filePath 本地文件地址
// materialType: image, voice, video, file   素材类型
func (e *EnterWechat) UploadMaterialByFile(file *multipart.FileHeader, materialType string, token string) (*EnterWechatResp, error) {

	var (
		resp            []byte
		err             error
		enterWechatResp EnterWechatResp
	)

	if materialType == "voice" {
		if filepath.Ext(file.Filename) != ".amr" {
			return nil, errors.New("only support amr file")
		}
	}

	resp, err = upload.Upload2Remote(UPLOAD_MEDIA_API, map[string]string{
		"access_token": token,
		"type":         materialType,
	}, file, "media")

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp, &enterWechatResp)
	if err != nil {
		return nil, err
	}
	return &enterWechatResp, nil
}
