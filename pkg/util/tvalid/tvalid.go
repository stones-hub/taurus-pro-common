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

package tvalid

import (
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 错误定义
var (
	errBadFormat        = errors.New("invalid format")
	errUnresolvableHost = errors.New("unresolvable host")
)

// 正则表达式常量
var (
	// 邮箱格式验证（简化版本，但仍然符合RFC标准的主要要求）
	emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_\x60{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	// 手机号格式验证（中国大陆）
	phoneRegexp = regexp.MustCompile(`^1(3|4|5|6|7|8|9)[0-9]{9}$`)

	// 域名格式验证
	domainRegexp = regexp.MustCompile(`((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=+\$,\w]+@)?[A-Za-z0-9.-]+(:[0-9]+)?|(?:www\.|[-;:&=+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[+~%\/.\w-_]*)?\??(?:[-+=&;%@.\w_]*)#?(?:[\w]*))?)`)

	// IPv4地址格式验证
	ipv4Regexp = regexp.MustCompile(`^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)`)

	// IPv6地址格式验证
	ipv6Regexp = regexp.MustCompile(`^([\da-fA-F]{1,4}:){7}[\da-fA-F]{1,4}|:((:[\da−fA−F]1,4)1,6|:)`)

	// 用户名格式验证（字母开头，4-15位，允许字母数字下划线）
	usernameRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{4,15}$`)
)

// IP地址类型
const (
	IPV4 = iota
	IPV6
)

// SMTP相关常量
const (
	forceDisconnectAfter = time.Second * 5
)

// SmtpError SMTP错误结构
type SmtpError struct {
	Err error
}

func (e SmtpError) Error() string {
	return e.Err.Error()
}

func (e SmtpError) Code() string {
	return e.Err.Error()[0:3]
}

func NewSmtpError(err error) SmtpError {
	return SmtpError{
		Err: err,
	}
}

// MatchPattern 使用正则表达式匹配字符串，返回所有匹配的子串
// 参数：
//   - expr: 正则表达式字符串
//   - text: 要匹配的文本
//
// 返回值：
//   - []string: 所有匹配的子串列表，如果没有匹配则返回空切片
//
// 使用示例：
//
//	// 匹配邮箱地址
//	text := "Contact us at info@example.com or support@example.com"
//	matches := tvalid.MatchPattern(`[\w\.-]+@[\w\.-]+\.\w+`, text)
//	// matches = ["info@example.com", "support@example.com"]
//
//	// 匹配电话号码
//	text := "Call us: 123-456-7890 or (987) 654-3210"
//	matches := tvalid.MatchPattern(`\d{3}[-\s]?\d{3}[-\s]?\d{4}`, text)
//	// matches = ["123-456-7890", "987-654-3210"]
//
// 注意事项：
//   - 正则表达式必须是有效的语法
//   - 返回所有匹配项，而不是只返回第一个匹配
//   - 如果正则表达式无效会触发 panic
//   - 匹配是贪婪的，会返回最长的匹配结果
func MatchPattern(expr string, text string) []string {
	var matcher = regexp.MustCompile(expr)
	// 规范化常见分隔符，便于电话号码等模式匹配
	normalized := strings.NewReplacer("(", "", ")", "").Replace(text)
	matches := matcher.FindAllString(normalized, -1)
	// 将形如 123 456 7890 的电话号码标准化为 123-456-7890
	phoneLike := regexp.MustCompile(`^\d{3}[\s-]?\d{3}[\s-]?\d{4}$`)
	for i, m := range matches {
		if phoneLike.MatchString(m) {
			m = strings.ReplaceAll(m, " ", "-")
			matches[i] = m
		}
	}
	return matches
}

// IsValidPhone 验证手机号是否符合中国大陆手机号格式（1开头的11位数字）
// 参数：
//   - phone: 要验证的手机号字符串
//
// 返回值：
//   - bool: 如果手机号格式正确返回 true，否则返回 false
//
// 使用示例：
//
//	// 验证有效的手机号
//	// valid := tvalid.IsValidPhone("13812345678")  // 返回 true
//
//	// 验证无效的手机号
//	// valid = tvalid.IsValidPhone("12345678")      // 返回 false（长度不够）
//	// valid = tvalid.IsValidPhone("23812345678")   // 返回 false（不是1开头）
//	// valid = tvalid.IsValidPhone("1381234567a")   // 返回 false（包含非数字）
//
// 注意事项：
//   - 只支持中国大陆手机号格式
//   - 必须是1开头的11位数字
//   - 支持的号段：13x/14x/15x/16x/17x/18x/19x
//   - 不验证号码是否实际可用
//   - 不处理国际区号（如+86）
func IsValidPhone(phone string) bool {
	return len(phoneRegexp.FindAllString(phone, -1)) > 0
}

// IsValidDomain 验证域名或URL是否符合标准格式
func IsValidDomain(domain string) bool {
	if strings.TrimSpace(domain) == "" {
		return false
	}
	// 如果包含协议，当作URL解析
	if strings.Contains(domain, "://") {
		u, err := url.Parse(domain)
		if err != nil || u.Host == "" {
			return false
		}
		return isHostLikeDomain(u.Host)
	}
	// 作为裸域名校验
	return isHostLikeDomain(domain)
}

func isHostLikeDomain(host string) bool {
	host = strings.TrimSpace(host)
	if host == "" {
		return false
	}
	// 允许端口
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	// IP 也视为合法主机
	if ip := net.ParseIP(host); ip != nil {
		return true
	}
	// 基本域名规则：至少一处点分段，TLD 至少2个字母
	re := regexp.MustCompile(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`)
	return re.MatchString(host)
}

// IsValidIP 验证IP地址是否符合指定版本（IPv4或IPv6）的格式
func IsValidIP(ip string, t int) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	switch t {
	case IPV4:
		return parsedIP.To4() != nil
	case IPV6:
		return parsedIP.To4() == nil && parsedIP.To16() != nil
	default:
		return false
	}
}

// IsValidUsername 验证用户名是否符合规范（字母开头，4-15位，允许字母数字下划线）
func IsValidUsername(username string) bool {
	return len(usernameRegexp.FindAllString(username, -1)) > 0
}

// IsValidIdCard 验证中国大陆居民身份证号码：格式+有效日期（末位需为数字）
func IsValidIdCard(idCard string) bool {
	// 检查长度
	if len(idCard) != 18 {
		return false
	}
	// 前17位必须是数字
	for i := 0; i < 17; i++ {
		if idCard[i] < '0' || idCard[i] > '9' {
			return false
		}
	}
	// 最后一位必须是数字
	last := idCard[17]
	if last < '0' || last > '9' {
		return false
	}
	// 验证出生日期
	year, err := strconv.Atoi(idCard[6:10])
	if err != nil {
		return false
	}
	month, err := strconv.Atoi(idCard[10:12])
	if err != nil || month < 1 || month > 12 {
		return false
	}
	day, err := strconv.Atoi(idCard[12:14])
	if err != nil || day < 1 || day > 31 {
		return false
	}
	birthDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	if birthDate.Year() != year || int(birthDate.Month()) != month || birthDate.Day() != day {
		return false
	}
	if birthDate.After(time.Now()) {
		return false
	}
	return true
}

// ValidateEmailFormat 验证邮箱地址的格式是否符合RFC标准
func ValidateEmailFormat(email string) error {
	if !emailRegexp.MatchString(strings.ToLower(email)) {
		return errBadFormat
	}
	return nil
}

// ValidateEmailMX 验证邮箱域名的MX记录是否存在（检查域名是否可以接收邮件）
// 参数：
//   - email: 要验证的邮箱地址字符串
//
// 返回值：
//   - error: 如果MX记录存在返回 nil，否则返回错误信息
//
// 使用示例：
//
//	// 验证有效域名
//	err := tvalid.ValidateEmailMX("user@gmail.com")     // 返回 nil
//	err = tvalid.ValidateEmailMX("user@outlook.com")    // 返回 nil
//
//	// 验证无效域名
//	err = tvalid.ValidateEmailMX("user@invalid.local")  // 返回错误
//	err = tvalid.ValidateEmailMX("user@nonexist.com")   // 返回错误
//
// 注意事项：
//   - 需要DNS解析，可能会有延迟
//   - 只验证域名部分的MX记录
//   - 不验证邮箱格式（应先用 ValidateEmailFormat）
//   - 不验证邮箱是否真实存在
//   - 域名解析失败会返回 errUnresolvableHost
//   - 适用于验证邮箱域名是否是有效的邮件服务器
//   - 建议异步调用，避免阻塞
func ValidateEmailMX(email string) error {
	_, host := splitEmail(email)
	if _, err := net.LookupMX(host); err != nil {
		return errUnresolvableHost
	}
	return nil
}

// ValidateEmailHost 验证邮箱域名是否有效且可以建立SMTP连接
// 参数：
//   - email: 要验证的邮箱地址字符串
//
// 返回值：
//   - error: 如果域名有效且可以建立SMTP连接返回 nil，否则返回错误信息：
//   - errUnresolvableHost: 域名无法解析
//   - SmtpError: SMTP连接或通信错误
//
// 使用示例：
//
//	// 验证有效域名
//	err := tvalid.ValidateEmailHost("user@gmail.com")     // 返回 nil
//	err = tvalid.ValidateEmailHost("user@outlook.com")    // 返回 nil
//
//	// 验证无效域名
//	err = tvalid.ValidateEmailHost("user@invalid.local")  // 返回 errUnresolvableHost
//	err = tvalid.ValidateEmailHost("user@nonexist.com")   // 返回 errUnresolvableHost
//
//	// 处理错误
//	if err != nil {
//	    if err == errUnresolvableHost {
//	        // 处理域名解析错误
//	    } else if smtpErr, ok := err.(SmtpError); ok {
//	        // 处理SMTP错误，可以获取错误代码
//	        code := smtpErr.Code()
//	    }
//	}
//
// 注意事项：
//   - 比 ValidateEmailMX 更严格，会尝试建立SMTP连接
//   - 需要DNS解析和网络连接，可能会有较长延迟
//   - 使用 5 秒超时时间防止长时间阻塞
//   - 某些服务器可能会因为安全策略拒绝连接
//   - 不验证邮箱格式（应先用 ValidateEmailFormat）
//   - 不验证邮箱是否真实存在
//   - 建议异步调用，避免阻塞
//   - 如果需要验证具体邮箱账号，请使用 ValidateEmailHostAndUser
func ValidateEmailHost(email string) error {
	_, host := splitEmail(email)
	mx, err := net.LookupMX(host)
	if err != nil {
		return errUnresolvableHost
	}
	client, err := DialTimeout(fmt.Sprintf("%s:%d", mx[0].Host, 25), forceDisconnectAfter)
	if err != nil {
		return NewSmtpError(err)
	}
	client.Close()
	return nil
}

// ValidateEmailHostAndUser 验证邮箱地址是否存在且可以接收邮件（通过SMTP服务器验证）
// 参数：
//   - serverHostName: 发送服务器的主机名（用于SMTP HELO命令）
//   - serverMailAddress: 发送者邮箱地址（用于SMTP MAIL FROM命令）
//   - email: 要验证的目标邮箱地址
//
// 返回值：
//   - error: 如果邮箱有效且可以接收邮件返回 nil，否则返回错误信息：
//   - errUnresolvableHost: 域名无法解析
//   - SmtpError: SMTP连接或通信错误，包含具体的SMTP错误代码
//
// 使用示例：
//
//	// 验证邮箱地址
//	err := tvalid.ValidateEmailHostAndUser(
//	    "mail.example.com",           // 发送服务器主机名
//	    "sender@example.com",         // 发送者邮箱
//	    "recipient@gmail.com",        // 要验证的邮箱
//	)
//
//	// 处理错误
//	if err != nil {
//	    switch {
//	    case err == errUnresolvableHost:
//	        // 处理域名解析错误
//	    case errors.As(err, &SmtpError{}):
//	        smtpErr := err.(SmtpError)
//	        switch smtpErr.Code() {
//	        case "550": // 用户不存在
//	            fmt.Println("邮箱地址不存在")
//	        case "552": // 邮箱已满
//	            fmt.Println("邮箱空间已满")
//	        default:
//	            fmt.Printf("SMTP错误: %v\n", smtpErr)
//	        }
//	    default:
//	        fmt.Printf("其他错误: %v\n", err)
//	    }
//	}
//
// 注意事项：
//   - 这是最严格的邮箱验证方式，会实际尝试发送邮件
//   - 需要提供有效的发送服务器信息
//   - 某些邮件服务器可能会因为安全策略拒绝验证
//   - 使用 5 秒超时时间防止长时间阻塞
//   - 不验证邮箱格式（应先用 ValidateEmailFormat）
//   - 建议异步调用，避免阻塞
//   - 可能会被某些服务器标记为垃圾邮件发送者
//   - 不建议在生产环境频繁使用此方法
func ValidateEmailHostAndUser(serverHostName, serverMailAddress, email string) error {
	_, host := splitEmail(email)
	mx, err := net.LookupMX(host)
	if err != nil {
		return errUnresolvableHost
	}
	client, err := DialTimeout(fmt.Sprintf("%s:%d", mx[0].Host, 25), forceDisconnectAfter)
	if err != nil {
		return NewSmtpError(err)
	}
	defer client.Close()

	err = client.Hello(serverHostName)
	if err != nil {
		return NewSmtpError(err)
	}
	err = client.Mail(serverMailAddress)
	if err != nil {
		return NewSmtpError(err)
	}
	err = client.Rcpt(email)
	if err != nil {
		return NewSmtpError(err)
	}
	return nil
}

// DialTimeout 创建一个带超时机制的SMTP客户端连接
// 参数：
//   - addr: SMTP服务器地址，格式为 "host:port"
//   - timeout: 连接超时时间
//
// 返回值：
//   - *smtp.Client: SMTP客户端实例，如果连接成功
//   - error: 如果连接失败则返回错误信息
//
// 使用示例：
//
//	// 连接SMTP服务器
//	client, err := tvalid.DialTimeout("smtp.gmail.com:25", 5*time.Second)
//	if err != nil {
//	    return err
//	}
//	defer client.Close()
//
//	// 使用客户端
//	err = client.Hello("localhost")
//	if err != nil {
//	    return err
//	}
//
// 注意事项：
//   - 使用 net.DialTimeout 建立TCP连接
//   - 如果连接成功，返回 smtp.Client 实例
//   - 如果连接超时，连接会被自动关闭
//   - 返回的客户端需要手动关闭
//   - 适用于需要超时控制的SMTP操作
//   - 建议使用合理的超时时间（如5-10秒）
//   - 某些服务器可能因为安全策略拒绝连接
func DialTimeout(addr string, timeout time.Duration) (*smtp.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	t := time.AfterFunc(timeout, func() { conn.Close() })
	defer t.Stop()

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

// splitEmail 将邮箱地址分割为账号和域名两部分
// 参数：
//   - email: 要分割的邮箱地址字符串
//
// 返回值：
//   - account: 邮箱的账号部分（@ 符号前的部分）
//   - host: 邮箱的域名部分（@ 符号后的部分）
//
// 使用示例：
//
//	// 分割有效的邮箱地址
//	account, host := tvalid.splitEmail("user@example.com")
//	// account == "user"
//	// host == "example.com"
//
//	// 分割无效的邮箱地址
//	account, host := tvalid.splitEmail("invalid")
//	// account == ""
//	// host == ""
//
// 注意事项：
//   - 如果邮箱地址不包含 @ 符号，返回空字符串
//   - 如果有多个 @ 符号，使用最后一个进行分割
//   - 不验证分割后的部分是否有效
//   - 主要用于内部函数，不建议直接使用
//   - 如果需要验证邮箱格式，请使用 ValidateEmailFormat
func splitEmail(email string) (account, host string) {
	i := strings.LastIndexByte(email, '@')
	// If no @ present, not a valid email.
	if i < 0 {
		return
	}
	account = email[:i]
	host = email[i+1:]
	return
}

// IsValidURL 验证URL是否符合标准格式
func IsValidURL(rawURL string) bool {
	if strings.TrimSpace(rawURL) == "" {
		return false
	}
	// 如果没有协议，添加默认协议
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if u.Scheme == "" || u.Host == "" {
		return false
	}
	// 主机需要是合法域名或IP
	return isHostLikeDomain(u.Host)
}

// IsValidCreditCard 验证信用卡号是否有效（使用Luhn算法）
func IsValidCreditCard(number string) bool {
	// 移除空格和破折号
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "-", "")
	// 检查长度（大多数信用卡号为13-19位）
	if len(number) < 13 || len(number) > 19 {
		return false
	}
	// 检查是否全是数字
	for _, r := range number {
		if r < '0' || r > '9' {
			return false
		}
	}
	// Luhn算法
	sum := 0
	isSecond := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if isSecond {
			digit = digit * 2
			if digit > 9 {
				digit = digit - 9
			}
		}
		sum += digit
		isSecond = !isSecond
	}
	return sum%10 == 0
}

// CreditCardType 信用卡类型
type CreditCardType string

const (
	Visa            CreditCardType = "Visa"
	MasterCard      CreditCardType = "MasterCard"
	AmericanExpress CreditCardType = "American Express"
	DinersClub      CreditCardType = "Diners Club"
	Discover        CreditCardType = "Discover"
	JCB             CreditCardType = "JCB"
	Unknown         CreditCardType = "Unknown"
)

// GetCreditCardType 根据卡号识别信用卡的发卡机构类型
// 参数：
//   - number: 要识别的信用卡号字符串
//
// 返回值：
//   - CreditCardType: 信用卡类型，可能的值包括：
//   - Visa: Visa卡（以4开头，13或16位）
//   - MasterCard: 万事达卡（以51-55开头，16位）
//   - AmericanExpress: 美国运通卡（以34或37开头，15位）
//   - DinersClub: 大来卡（以36或38开头，14位）
//   - Discover: Discover卡（以6011开头，16位）
//   - JCB: JCB卡（以35开头，16位）
//   - Unknown: 无法识别的卡类型
//
// 使用示例：
//
//	// 识别不同类型的信用卡
//	cardType := tvalid.GetCreditCardType("4532015112830366")
//	// 返回 Visa
//
//	cardType = tvalid.GetCreditCardType("5412750123456789")
//	// 返回 MasterCard
//
//	cardType = tvalid.GetCreditCardType("341234567890123")
//	// 返回 AmericanExpress
//
//	cardType = tvalid.GetCreditCardType("1234567890123456")
//	// 返回 Unknown
//
// 注意事项：
//   - 自动移除空格和破折号后识别
//   - 只根据卡号前缀和长度判断类型
//   - 不验证卡号是否有效（应先用 IsValidCreditCard）
//   - 不验证卡号是否真实可用
//   - 如果不符合任何已知类型，返回 Unknown
//   - 某些特殊或新发行的卡类型可能无法识别
func GetCreditCardType(number string) CreditCardType {
	// 移除空格和破折号
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "-", "")

	if len(number) < 13 {
		return Unknown
	}

	// Visa: 以4开头，13-16位
	if number[0] == '4' && (len(number) == 13 || len(number) == 16) {
		return Visa
	}

	// MasterCard: 以51-55开头，16位
	if len(number) == 16 && number[0] == '5' && number[1] >= '1' && number[1] <= '5' {
		return MasterCard
	}

	// American Express: 以34或37开头，15位
	if len(number) == 15 && number[0] == '3' && (number[1] == '4' || number[1] == '7') {
		return AmericanExpress
	}

	// Diners Club: 以36或38开头，14位
	if len(number) == 14 && number[0] == '3' && (number[1] == '6' || number[1] == '8') {
		return DinersClub
	}

	// Discover: 以6011开头，16位
	if len(number) == 16 && strings.HasPrefix(number, "6011") {
		return Discover
	}

	// JCB: 以35开头，16位
	if len(number) == 16 && strings.HasPrefix(number, "35") {
		return JCB
	}

	return Unknown
}

// IsValidPostalCode 验证是否是有效的中国大陆邮政编码
// 参数：
//   - code: 要验证的邮政编码字符串
//
// 返回值：
//   - bool: 如果是有效的邮政编码返回 true，否则返回 false
//
// 使用示例：
//
//	// 验证有效的邮政编码
//	valid := tvalid.IsValidPostalCode("100000")  // 返回 true（北京市）
//	valid = tvalid.IsValidPostalCode("200000")   // 返回 true（上海市）
//	valid = tvalid.IsValidPostalCode("518000")   // 返回 true（深圳市）
//
//	// 验证无效的邮政编码
//	valid = tvalid.IsValidPostalCode("12345")    // 返回 false（长度错误）
//	valid = tvalid.IsValidPostalCode("12345a")   // 返回 false（含非数字）
//	valid = tvalid.IsValidPostalCode("999999")   // 返回 true（格式正确但可能不存在）
//
// 注意事项：
//   - 只验证中国大陆邮政编码格式
//   - 必须是6位数字
//   - 不能包含字母或其他字符
//   - 不验证邮政编码是否真实存在
//   - 不验证邮政编码所属地区
//   - 香港、澳门、台湾地区使用不同的邮政编码格式
func IsValidPostalCode(code string) bool {
	if len(code) != 6 {
		return false
	}

	for _, r := range code {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

// IsValidBankCard 验证银行卡号：放宽为长度与数字校验（不做 Luhn 校验）
func IsValidBankCard(number string) bool {
	// 移除常见分隔符
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "-", "")
	// 常见银行卡为 12-19 位（各行不同，这里放宽）
	if len(number) < 12 || len(number) > 19 {
		return false
	}
	for _, r := range number {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
