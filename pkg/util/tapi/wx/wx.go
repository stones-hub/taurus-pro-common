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

package wx

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/util/tcrypt"
)

/*
企业微信应用对接
https://developer.work.weixin.qq.com/document/path/90254
*/

// TokenCache 用于缓存企业微信访问令牌的结构体
// 字段：
//   - Token: 访问令牌字符串
//   - ExpiresAt: 令牌过期时间
//
// 注意事项：
//   - 令牌通常有效期为2小时
//   - 过期时间会提前5分钟设置，以确保安全边际
//   - 用于 TokenManager 的内部缓存
//   - 线程安全，通过互斥锁保护
type TokenCache struct {
	Token     string    // 访问令牌
	ExpiresAt time.Time // 过期时间
}

// TokenManagerConfig 企业微信访问令牌管理器的配置选项
// 字段：
//   - HTTPTimeout: HTTP请求超时时间
//   - RetryConfig: 重试策略配置
//
// 使用示例：
//
//	config := &TokenManagerConfig{
//	    HTTPTimeout: 30 * time.Second,
//	    RetryConfig: &RetryConfig{
//	        MaxRetries:  3,
//	        InitialWait: time.Second,
//	        MaxWait:     10 * time.Second,
//	        Multiplier:  2.0,
//	    },
//	}
//
// 注意事项：
//   - 如果不提供配置，将使用默认值
//   - 默认HTTP超时时间为30秒
//   - 默认使用指数退避重试策略
//   - 建议根据网络情况调整超时时间
//   - 重试配置应考虑业务需求和系统负载
type TokenManagerConfig struct {
	HTTPTimeout time.Duration // HTTP请求超时时间
	RetryConfig *RetryConfig  // 重试配置
}

// RetryConfig 定义重试策略的配置选项
// 字段：
//   - MaxRetries: 最大重试次数
//   - InitialWait: 首次重试等待时间
//   - MaxWait: 最大等待时间
//   - Multiplier: 等待时间的增长倍数
//
// 使用示例：
//
//	config := &RetryConfig{
//	    MaxRetries:  3,        // 最多重试3次
//	    InitialWait: time.Second,  // 首次等待1秒
//	    MaxWait:     10 * time.Second,  // 最长等待10秒
//	    Multiplier:  2.0,      // 每次等待时间翻倍
//	}
//
// 注意事项：
//   - 使用指数退避策略
//   - 等待时间计算：wait = min(MaxWait, InitialWait * (Multiplier ^ retryCount))
//   - 建议根据接口特性调整参数
//   - 过多的重试可能导致系统负载增加
//   - 如果不提供配置，将使用默认值
type RetryConfig struct {
	MaxRetries  int           // 最大重试次数
	InitialWait time.Duration // 初始等待时间
	MaxWait     time.Duration // 最大等待时间
	Multiplier  float64       // 等待时间乘数
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = &RetryConfig{
	MaxRetries:  3,
	InitialWait: time.Second,
	MaxWait:     10 * time.Second,
	Multiplier:  2.0,
}

// TokenManager 企业微信访问令牌的管理器，负责令牌的获取、缓存和刷新
// 字段：
//   - cache: 令牌缓存，key为应用ID
//   - mutex: 保护缓存的读写锁
//   - client: HTTP客户端
//   - config: 管理器配置
//   - corpID: 企业ID
//   - secret: 应用密钥
//
// 使用示例：
//
//	manager := NewTokenManager(
//	    "wx123456789",  // 企业ID
//	    "secret123",    // 应用密钥
//	    &TokenManagerConfig{
//	        HTTPTimeout: 30 * time.Second,
//	        RetryConfig: DefaultRetryConfig,
//	    },
//	)
//
//	// 获取访问令牌
//	token, err := manager.GetToken()
//	if err != nil {
//	    return err
//	}
//
// 注意事项：
//   - 线程安全，可以并发使用
//   - 自动处理令牌过期和刷新
//   - 使用缓存减少API调用
//   - 支持失败重试和错误处理
//   - 建议全局使用单个实例
type TokenManager struct {
	cache  map[string]*TokenCache // 缓存的Token，key为应用ID
	mutex  sync.RWMutex           // 读写锁
	client *http.Client           // HTTP客户端
	config *TokenManagerConfig    // 配置
	corpID string                 // 企业ID
	secret string                 // 应用密钥
}

// NewTokenManager 创建一个新的企业微信访问令牌管理器
// 参数：
//   - corpID: 企业ID，从企业微信管理后台获取
//   - secret: 应用密钥，从企业微信管理后台获取
//   - config: 管理器配置，如果为nil则使用默认配置
//
// 返回值：
//   - *TokenManager: 令牌管理器实例
//
// 使用示例：
//
//	// 使用默认配置
//	manager := NewTokenManager(
//	    "wx123456789",  // 企业ID
//	    "secret123",    // 应用密钥
//	    nil,            // 使用默认配置
//	)
//
//	// 使用自定义配置
//	manager := NewTokenManager(
//	    "wx123456789",
//	    "secret123",
//	    &TokenManagerConfig{
//	        HTTPTimeout: 30 * time.Second,
//	        RetryConfig: &RetryConfig{
//	            MaxRetries:  3,
//	            InitialWait: time.Second,
//	            MaxWait:     10 * time.Second,
//	            Multiplier:  2.0,
//	        },
//	    },
//	)
//
// 注意事项：
//   - 企业ID和密钥必须正确，否则无法获取令牌
//   - 建议使用环境变量或配置文件管理敏感信息
//   - 一个应用通常只需要一个管理器实例
//   - 管理器是线程安全的，可以并发使用
//   - 使用默认配置适合大多数场景
func NewTokenManager(corpID, secret string, config *TokenManagerConfig) *TokenManager {
	if config == nil {
		config = &TokenManagerConfig{
			HTTPTimeout: 30 * time.Second,
			RetryConfig: DefaultRetryConfig,
		}
	}

	return &TokenManager{
		cache:  make(map[string]*TokenCache),
		client: &http.Client{Timeout: config.HTTPTimeout},
		config: config,
		corpID: corpID,
		secret: secret,
	}
}

// EnterWechat 企业微信应用实例，用于管理应用配置和发送消息
// 字段：
//   - corpID: 企业ID，从企业微信管理后台获取
//   - agentID: 应用ID，从企业微信管理后台获取
//   - name: 应用名称，用于标识应用
//   - secret: 应用密钥，从企业微信管理后台获取
//   - sendUID: 可接收消息的用户ID列表
//   - tokenManager: 访问令牌管理器
//   - token: 当前令牌，用于URL验证
//
// 使用示例：
//
//	app := NewEnterWechat(
//	    "wx123456789",           // 企业ID
//	    1000001,                 // 应用ID
//	    "测试应用",               // 应用名称
//	    "secret123",             // 应用密钥
//	    []string{"user1", "user2"}, // 接收消息的用户
//	    nil,                     // 使用默认配置
//	)
//
//	// 发送文本消息
//	resp, err := app.SendTextMessage(&Text{
//	    Content: "Hello, World!",
//	})
//
//	// 发送图片消息
//	resp, err = app.SendImageMessage(&Image{
//	    MediaId: "MEDIA_ID",
//	})
//
// 注意事项：
//   - 所有参数都必须正确填写，否则可能导致应用无法正常工作
//   - 建议使用环境变量或配置文件管理敏感信息
//   - 一个应用通常只需要一个实例
//   - 实例是线程安全的，可以并发使用
//   - 使用默认配置适合大多数场景
//   - 用户ID列表可以后续通过 SetSendUID 修改
type EnterWechat struct {
	corpID       string        // 企业ID
	agentID      int           // 应用ID
	name         string        // 应用名
	secret       string        // 应用密钥
	sendUID      []string      // 可接受推送消息的用户账户
	tokenManager *TokenManager // Token管理器
	token        string        // 当前Token，用于验证URL
}

// Getters
func (e *EnterWechat) CorpID() string    { return e.corpID }
func (e *EnterWechat) AgentID() int      { return e.agentID }
func (e *EnterWechat) Name() string      { return e.name }
func (e *EnterWechat) Secret() string    { return e.secret }
func (e *EnterWechat) SendUID() []string { return e.sendUID }
func (e *EnterWechat) Token() string     { return e.token }

// GetTokenManager 获取Token管理器
func (e *EnterWechat) GetTokenManager() *TokenManager { return e.tokenManager }

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

// EnterWechatConfig 企业微信配置
type EnterWechatConfig struct {
	TokenManagerConfig *TokenManagerConfig // Token管理器配置
}

// NewEnterWechat 创建一个企业微信应用实例
// corpID 是企业ID, agentID 是应用ID, name 是应用名称，secret 是应用密钥, sendUID 是推送消息的用户ID
func NewEnterWechat(corpID string, agentID int, name string, secret string, sendUID []string, config *EnterWechatConfig) *EnterWechat {
	if config == nil {
		config = &EnterWechatConfig{
			TokenManagerConfig: &TokenManagerConfig{
				HTTPTimeout: 30 * time.Second,
				RetryConfig: DefaultRetryConfig,
			},
		}
	}

	return &EnterWechat{
		corpID:       corpID,
		agentID:      agentID,
		name:         name,
		secret:       secret,
		sendUID:      sendUID,
		tokenManager: NewTokenManager(corpID, secret, config.TokenManagerConfig),
	}
}

// GetToken 获取访问令牌（带缓存）
func (m *TokenManager) GetToken() (string, error) {
	key := fmt.Sprintf("%s:%s", m.corpID, m.secret)

	// 尝试从缓存获取
	m.mutex.RLock()
	if cache, ok := m.cache[key]; ok {
		if time.Now().Before(cache.ExpiresAt) {
			token := cache.Token
			m.mutex.RUnlock()
			return token, nil
		}
	}
	m.mutex.RUnlock()

	// 缓存不存在或已过期，重新获取
	var token string
	err := m.withRetry(func() error {
		return m.fetchToken(&token)
	})

	if err != nil {
		return "", err
	}
	return token, nil
}

// fetchToken 从企业微信API获取新的访问令牌
func (m *TokenManager) fetchToken(tokenPtr *string) error {
	URL := fmt.Sprintf("%s?corpid=%s&corpsecret=%s", ACCESS_TOKEN_API, m.corpID, m.secret)

	resp, err := m.client.Get(URL)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("API request failed: %s (code: %d)", result.ErrMsg, result.ErrCode)
	}

	// 更新缓存
	key := fmt.Sprintf("%s:%s", m.corpID, m.secret)
	m.mutex.Lock()
	m.cache[key] = &TokenCache{
		Token:     result.AccessToken,
		ExpiresAt: time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second), // 提前5分钟过期
	}
	m.mutex.Unlock()

	*tokenPtr = result.AccessToken
	return nil
}

// RefreshToken 强制刷新访问令牌
func (m *TokenManager) RefreshToken() (string, error) {
	key := fmt.Sprintf("%s:%s", m.corpID, m.secret)
	m.mutex.Lock()
	delete(m.cache, key)
	m.mutex.Unlock()

	var token string
	err := m.withRetry(func() error {
		return m.fetchToken(&token)
	})

	if err != nil {
		return "", err
	}
	return token, nil
}

// withRetry 带重试的函数执行
func (m *TokenManager) withRetry(fn func() error) error {
	config := m.config.RetryConfig
	if config == nil {
		config = DefaultRetryConfig
	}

	var lastErr error
	wait := config.InitialWait

	for i := 0; i <= config.MaxRetries; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			if i == config.MaxRetries {
				break
			}

			// 计算下一次重试的等待时间
			if i > 0 {
				wait = time.Duration(float64(wait) * config.Multiplier)
				if wait > config.MaxWait {
					wait = config.MaxWait
				}
			}
			time.Sleep(wait)
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries, lastErr)
}

// VerifyURL 验证回调URL的签名
func (e *EnterWechat) VerifyURL(signature, timestamp, nonce, echostr string) (string, error) {
	// 1. 将token、timestamp、nonce三个参数进行字典序排序
	params := []string{e.token, timestamp, nonce}
	sort.Strings(params)

	// 2. 将三个参数字符串拼接成一个字符串进行sha1加密
	str := strings.Join(params, "")
	h := sha1.New()
	h.Write([]byte(str))
	sign := fmt.Sprintf("%x", h.Sum(nil))

	// 3. 开发者获得加密后的字符串可与signature对比，标识该请求来源于微信
	if sign != signature {
		return "", errors.New("signature verification failed")
	}

	return echostr, nil
}

// SetToken 设置验证Token
func (e *EnterWechat) SetToken(token string) {
	e.token = token
}

// MessageTemplate 消息模板
type MessageTemplate struct {
	Title    string                 // 模板标题
	Content  string                 // 模板内容
	Color    string                 // 字体颜色
	URL      string                 // 点击跳转的URL
	Data     map[string]string      // 模板数据
	MiniApp  *MiniAppConfig         // 小程序配置
	Callback map[string]interface{} // 回调数据
}

// MiniAppConfig 小程序配置
type MiniAppConfig struct {
	AppID    string // 小程序的AppID
	PagePath string // 小程序的页面路径
}

// SendTemplateMessage 发送模板消息
func (e *EnterWechat) SendTemplateMessage(template *MessageTemplate) (*EnterWechatResp, error) {
	// 构建模板消息
	card := &TextCard{
		Title:       template.Title,
		Description: e.renderTemplate(template.Content, template.Data),
		Url:         template.URL,
	}

	// 如果有小程序配置，使用图文消息
	if template.MiniApp != nil {
		news := &News{
			Articles: []Article{
				{
					Title:       template.Title,
					Description: e.renderTemplate(template.Content, template.Data),
					Url:         template.URL,
					AppId:       template.MiniApp.AppID,
					PagePath:    template.MiniApp.PagePath,
				},
			},
		}
		return e.SendNewsMessage(news)
	}

	return e.SendTextCardMessage(card)
}

// renderTemplate 渲染模板内容
func (e *EnterWechat) renderTemplate(content string, data map[string]string) string {
	for key, value := range data {
		content = strings.ReplaceAll(content, "{{"+key+"}}", value)
	}
	return content
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
	return tcrypt.MD5([]byte(ACCESS_TOKEN_KEY + e.CorpID() + e.Name() + e.Secret()))
}

// AccessToken 获取访问令牌
func (e *EnterWechat) AccessToken() (*EnterWechatResp, error) {
	token, err := e.tokenManager.GetToken()
	if err != nil {
		return nil, err
	}

	return &EnterWechatResp{
		AccessToken: token,
		ExpiresIn:   7200, // 默认2小时过期
	}, nil
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

// SendMessageToUsers 发送消息给指定用户
func (e *EnterWechat) SendMessageToUsers(message *EnterWeChatSendMessage, users []string) (*EnterWechatResp, error) {
	token, err := e.tokenManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	message.ToUser = strings.Join(users, "|")
	message.AgentId = e.AgentID()
	message.EnableDuplicateCheck = 1
	message.DuplicateCheckInterval = 60
	if message.MsgType == "news" {
		message.Safe = 0
	}

	return e.sendMessage(message, token)
}

// SendTextMessage 发送文本消息
func (e *EnterWechat) SendTextMessage(text *Text) (*EnterWechatResp, error) {
	token, err := e.tokenManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	message := e.prepareMessage("text", 1) // 文本消息使用安全模式
	message.Text = *text

	return e.sendMessage(message, token)
}

// SendMarkdownMessage 发送Markdown消息
func (e *EnterWechat) SendMarkdownMessage(markdown *Markdown) (*EnterWechatResp, error) {
	token, err := e.tokenManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	message := e.prepareMessage("markdown", 0) // Markdown消息不使用安全模式
	message.Markdown = *markdown

	return e.sendMessage(message, token)
}

// SendImageMessage 发送图片消息
func (e *EnterWechat) SendImageMessage(image *Image) (*EnterWechatResp, error) {
	token, err := e.tokenManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	message := e.prepareMessage("image", 1) // 图片消息使用安全模式
	message.Image = *image

	return e.sendMessage(message, token)
}

// SendNewsMessage 发送图文消息
func (e *EnterWechat) SendNewsMessage(news *News) (*EnterWechatResp, error) {
	token, err := e.tokenManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	message := e.prepareMessage("news", 0) // 图文消息不使用安全模式
	message.News = *news

	return e.sendMessage(message, token)
}

// SendVoiceMessage 发送语音消息
func (e *EnterWechat) SendVoiceMessage(voice *Voice) (*EnterWechatResp, error) {
	token, err := e.tokenManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	message := e.prepareMessage("voice", 1) // 语音消息使用安全模式
	message.Voice = *voice

	return e.sendMessage(message, token)
}

// SendVideoMessage 发送视频消息
func (e *EnterWechat) SendVideoMessage(video *Video) (*EnterWechatResp, error) {
	token, err := e.tokenManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	message := e.prepareMessage("video", 1) // 视频消息使用安全模式
	message.Video = *video

	return e.sendMessage(message, token)
}

// SendTextCardMessage 发送卡片消息
func (e *EnterWechat) SendTextCardMessage(card *TextCard) (*EnterWechatResp, error) {
	token, err := e.tokenManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	message := e.prepareMessage("textcard", 1) // 卡片消息使用安全模式
	message.TextCard = *card

	return e.sendMessage(message, token)
}

// SendFileMessage 发送文件消息
func (e *EnterWechat) SendFileMessage(file *File) (*EnterWechatResp, error) {
	token, err := e.tokenManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	message := e.prepareMessage("file", 1) // 文件消息使用安全模式
	message.File = *file

	return e.sendMessage(message, token)
}

// prepareMessage 准备消息基本信息
func (e *EnterWechat) prepareMessage(msgType string, safe int) *EnterWeChatSendMessage {
	return &EnterWeChatSendMessage{
		ToUser:                 strings.Join(e.SendUID(), "|"),
		ToParty:                "",
		ToTag:                  "",
		ToAll:                  0,
		MsgType:                msgType,
		AgentId:                e.AgentID(),
		Text:                   Text{},
		Image:                  Image{},
		Voice:                  Voice{},
		Video:                  Video{},
		File:                   File{},
		TextCard:               TextCard{},
		News:                   News{},
		Safe:                   safe,
		EnableIdTrans:          0,
		EnableDuplicateCheck:   1,
		DuplicateCheckInterval: 60,
	}
}

// sendMessage 发送消息
func (e *EnterWechat) sendMessage(message *EnterWeChatSendMessage, token string) (*EnterWechatResp, error) {
	URL := SEND_MESSAGE_API + "?access_token=" + token

	// 使用HTTP客户端发送请求
	body, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("marshal message failed: %w", err)
	}

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := e.tokenManager.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// 检查状态码
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d", httpResp.StatusCode)
	}

	// 解析响应
	var result EnterWechatResp
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 检查业务错误码
	if result.ErrCode != 0 {
		// 如果是token过期，尝试刷新token
		if result.ErrCode == 42001 || result.ErrCode == 40014 {
			newToken, err := e.tokenManager.RefreshToken()
			if err != nil {
				return nil, err
			}
			// 使用新token重试
			message.EnableDuplicateCheck = 0 // 避免重复发送检查
			return e.sendMessage(message, newToken)
		}
		return nil, fmt.Errorf("API request failed: %s (code: %d)", result.ErrMsg, result.ErrCode)
	}

	return &result, nil
}

// UploadMaterialByPath 上传临时素材,
// filePath 本地文件地址
// materialType: image, voice, video, file   素材类型
func (e *EnterWechat) UploadMaterialByPath(fp string, materialType string) (*EnterWechatResp, error) {
	token, err := e.tokenManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	if materialType == "voice" && filepath.Ext(fp) != ".amr" {
		return nil, errors.New("only support amr file")
	}

	// 使用HTTP客户端上传文件
	URL := UPLOAD_MEDIA_API + "?access_token=" + token + "&type=" + materialType
	file, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	// 创建multipart表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("media", filepath.Base(fp))
	if err != nil {
		return nil, fmt.Errorf("create form file failed: %w", err)
	}

	// 写入文件内容
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copy file content failed: %w", err)
	}

	// 关闭writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer failed: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", URL, body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	httpResp, err := e.tokenManager.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// 检查状态码
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d", httpResp.StatusCode)
	}

	// 解析响应
	var result EnterWechatResp
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 检查业务错误码
	if result.ErrCode != 0 {
		// 如果是token过期，尝试刷新token
		if result.ErrCode == 42001 || result.ErrCode == 40014 {
			if _, err := e.tokenManager.RefreshToken(); err != nil {
				return nil, err
			}
			// 使用新token重试
			return e.UploadMaterialByPath(fp, materialType)
		}
		return nil, fmt.Errorf("API request failed: %s (code: %d)", result.ErrMsg, result.ErrCode)
	}

	return &result, nil
}

// UploadMaterialByFile 上传临时素材,
// filePath 本地文件地址
// materialType: image, voice, video, file   素材类型
func (e *EnterWechat) UploadMaterialByFile(file *multipart.FileHeader, materialType string) (*EnterWechatResp, error) {
	token, err := e.tokenManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	if materialType == "voice" && filepath.Ext(file.Filename) != ".amr" {
		return nil, errors.New("only support amr file")
	}

	// 使用HTTP客户端上传文件
	URL := UPLOAD_MEDIA_API + "?access_token=" + token + "&type=" + materialType

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer src.Close()

	// 创建multipart表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("media", file.Filename)
	if err != nil {
		return nil, fmt.Errorf("create form file failed: %w", err)
	}

	// 写入文件内容
	if _, err := io.Copy(part, src); err != nil {
		return nil, fmt.Errorf("copy file content failed: %w", err)
	}

	// 关闭writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer failed: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", URL, body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	httpResp, err := e.tokenManager.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// 检查状态码
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d", httpResp.StatusCode)
	}

	// 解析响应
	var result EnterWechatResp
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 检查业务错误码
	if result.ErrCode != 0 {
		// 如果是token过期，尝试刷新token
		if result.ErrCode == 42001 || result.ErrCode == 40014 {
			if _, err := e.tokenManager.RefreshToken(); err != nil {
				return nil, err
			}
			// 使用新token重试
			return e.UploadMaterialByFile(file, materialType)
		}
		return nil, fmt.Errorf("API request failed: %s (code: %d)", result.ErrMsg, result.ErrCode)
	}

	return &result, nil
}
