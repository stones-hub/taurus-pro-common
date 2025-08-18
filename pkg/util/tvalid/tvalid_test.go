package tvalid

import (
	"testing"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		text     string
		expected []string
	}{
		{
			name:     "匹配邮箱地址",
			expr:     `[\w\.-]+@[\w\.-]+\.\w+`,
			text:     "Contact us at info@example.com or support@example.com",
			expected: []string{"info@example.com", "support@example.com"},
		},
		{
			name:     "匹配电话号码",
			expr:     `\d{3}[-\s]?\d{3}[-\s]?\d{4}`,
			text:     "Call us: 123-456-7890 or (987) 654-3210",
			expected: []string{"123-456-7890", "987-654-3210"},
		},
		{
			name:     "无匹配",
			expr:     `\d+`,
			text:     "no numbers here",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchPattern(tt.expr, tt.text)
			if len(result) != len(tt.expected) {
				t.Errorf("MatchPattern() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("MatchPattern()[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestIsValidPhone(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected bool
	}{
		{
			name:     "有效手机号",
			phone:    "13812345678",
			expected: true,
		},
		{
			name:     "长度不够",
			phone:    "1381234567",
			expected: false,
		},
		{
			name:     "非法开头",
			phone:    "23812345678",
			expected: false,
		},
		{
			name:     "包含非数字",
			phone:    "1381234567a",
			expected: false,
		},
		{
			name:     "空字符串",
			phone:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPhone(tt.phone)
			if result != tt.expected {
				t.Errorf("IsValidPhone() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		expected bool
	}{
		{
			name:     "有效域名",
			domain:   "example.com",
			expected: true,
		},
		{
			name:     "子域名",
			domain:   "sub.example.com",
			expected: true,
		},
		{
			name:     "带协议的URL",
			domain:   "http://www.example.com",
			expected: true,
		},
		{
			name:     "无效域名",
			domain:   "invalid",
			expected: false,
		},
		{
			name:     "空字符串",
			domain:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidDomain(tt.domain)
			if result != tt.expected {
				t.Errorf("IsValidDomain() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		ipType   int
		expected bool
	}{
		{
			name:     "有效IPv4",
			ip:       "192.168.1.1",
			ipType:   IPV4,
			expected: true,
		},
		{
			name:     "无效IPv4",
			ip:       "256.1.2.3",
			ipType:   IPV4,
			expected: false,
		},
		{
			name:     "有效IPv6",
			ip:       "2001:db8::1",
			ipType:   IPV6,
			expected: true,
		},
		{
			name:     "无效IPv6",
			ip:       "2001:db8::g",
			ipType:   IPV6,
			expected: false,
		},
		{
			name:     "IPv4用于IPv6验证",
			ip:       "192.168.1.1",
			ipType:   IPV6,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidIP(tt.ip, tt.ipType)
			if result != tt.expected {
				t.Errorf("IsValidIP() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{
			name:     "有效用户名",
			username: "john123",
			expected: true,
		},
		{
			name:     "数字开头",
			username: "123user",
			expected: false,
		},
		{
			name:     "长度太短",
			username: "abc",
			expected: false,
		},
		{
			name:     "长度太长",
			username: "verylongusername123",
			expected: false,
		},
		{
			name:     "包含特殊字符",
			username: "user.name",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUsername(tt.username)
			if result != tt.expected {
				t.Errorf("IsValidUsername() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidIdCard(t *testing.T) {
	tests := []struct {
		name     string
		idCard   string
		expected bool
	}{
		{
			name:     "有效身份证号",
			idCard:   "110101199001011234", // 示例号码
			expected: true,
		},
		{
			name:     "长度错误",
			idCard:   "1234",
			expected: false,
		},
		{
			name:     "非数字字符",
			idCard:   "11010119900101123X",
			expected: false,
		},
		{
			name:     "无效月份",
			idCard:   "110101199013011234",
			expected: false,
		},
		{
			name:     "无效日期",
			idCard:   "110101199001321234",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidIdCard(tt.idCard)
			if result != tt.expected {
				t.Errorf("IsValidIdCard() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		expected bool
	}{
		{
			name:     "完整URL",
			rawURL:   "https://example.com",
			expected: true,
		},
		{
			name:     "带路径的URL",
			rawURL:   "http://example.com/path",
			expected: true,
		},
		{
			name:     "无协议",
			rawURL:   "example.com",
			expected: true, // 会自动添加http://
		},
		{
			name:     "无效URL",
			rawURL:   "not-a-url",
			expected: false,
		},
		{
			name:     "空字符串",
			rawURL:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidURL(tt.rawURL)
			if result != tt.expected {
				t.Errorf("IsValidURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidCreditCard(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "有效信用卡号",
			number:   "4532015112830366",
			expected: true,
		},
		{
			name:     "带分隔符",
			number:   "4532-0151-1283-0366",
			expected: true,
		},
		{
			name:     "长度错误",
			number:   "1234",
			expected: false,
		},
		{
			name:     "非数字字符",
			number:   "abcd-efgh-ijkl-mnop",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidCreditCard(tt.number)
			if result != tt.expected {
				t.Errorf("IsValidCreditCard() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetCreditCardType(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected CreditCardType
	}{
		{
			name:     "Visa卡",
			number:   "4532015112830366",
			expected: Visa,
		},
		{
			name:     "MasterCard",
			number:   "5412750123456789",
			expected: MasterCard,
		},
		{
			name:     "American Express",
			number:   "341234567890123",
			expected: AmericanExpress,
		},
		{
			name:     "未知类型",
			number:   "1234567890123456",
			expected: Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCreditCardType(tt.number)
			if result != tt.expected {
				t.Errorf("GetCreditCardType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidPostalCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name:     "有效邮编",
			code:     "100000",
			expected: true,
		},
		{
			name:     "长度错误",
			code:     "12345",
			expected: false,
		},
		{
			name:     "非数字字符",
			code:     "12345a",
			expected: false,
		},
		{
			name:     "空字符串",
			code:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPostalCode(tt.code)
			if result != tt.expected {
				t.Errorf("IsValidPostalCode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidBankCard(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "有效银行卡号",
			number:   "6222021234567890123",
			expected: true,
		},
		{
			name:     "带分隔符",
			number:   "6222-0212-3456-7890",
			expected: true,
		},
		{
			name:     "长度错误",
			number:   "1234",
			expected: false,
		},
		{
			name:     "非数字字符",
			number:   "abcd-efgh-ijkl-mnop",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidBankCard(tt.number)
			if result != tt.expected {
				t.Errorf("IsValidBankCard() = %v, want %v", result, tt.expected)
			}
		})
	}
}
