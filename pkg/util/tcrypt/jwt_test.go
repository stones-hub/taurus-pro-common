package tcrypt

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func TestGenerateAndParseToken(t *testing.T) {
	tests := []struct {
		name     string
		uid      string
		username string
		wantErr  bool
	}{
		{
			name:     "valid user",
			uid:      "12345",
			username: "test_user",
			wantErr:  false,
		},
		{
			name:     "zero uid",
			uid:      "0",
			username: "test_user",
			wantErr:  false,
		},
		{
			name:     "empty username",
			uid:      "12345",
			username: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 生成token
			token, err := GenerateToken(tt.uid, tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// 解析token
			claims, err := ParseToken(token)
			if err != nil {
				t.Errorf("ParseToken() error = %v", err)
				return
			}

			// 验证解析出的信息
			if claims.Uid != tt.uid {
				t.Errorf("ParseToken() uid = %v, want %v", claims.Uid, tt.uid)
			}
			if claims.Username != tt.username {
				t.Errorf("ParseToken() username = %v, want %v", claims.Username, tt.username)
			}
			if claims.Issuer != "taurus-pro" {
				t.Errorf("ParseToken() issuer = %v, want taurus-pro", claims.Issuer)
			}

			// 验证时间相关字段
			now := time.Now().Unix()
			if claims.NotBefore > now {
				t.Errorf("ParseToken() notBefore = %v should be less than current time %v", claims.NotBefore, now)
			}
			if claims.ExpiresAt <= now {
				t.Errorf("ParseToken() expiresAt = %v should be greater than current time %v", claims.ExpiresAt, now)
			}
		})
	}
}

func TestParseToken_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			token:   "invalid.token.format",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpiredToken(t *testing.T) {
	// 使用测试密钥
	testSecret := "61647649@qq.com"

	// 创建一个已经过期的token（过期时间设置为当前时间前1秒）
	nowTime := time.Now()
	expiredTime := nowTime.Add(-1 * time.Second)

	claims := Claims{
		Uid:      "12345",
		Username: "test_user",
		StandardClaims: jwt.StandardClaims{
			NotBefore: nowTime.Add(-2 * time.Second).Unix(), // 2秒前生效
			ExpiresAt: expiredTime.Unix(),                   // 1秒前过期
			Issuer:    "taurus-pro",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// 尝试解析过期的token
	_, err = ParseToken(tokenString)
	if err == nil {
		t.Error("ParseToken() should return error for expired token")
	} else {
		t.Logf("Got expected error for expired token: %v", err)
	}
}

func TestGenerateTokenWithExpiration(t *testing.T) {
	tests := []struct {
		name       string
		uid        string
		username   string
		issuer     string
		secret     string
		expiration time.Duration
		wantErr    bool
	}{
		{
			name:       "custom expiration",
			uid:        "12345",
			username:   "test_user",
			issuer:     "custom-issuer",
			secret:     "custom-secret",
			expiration: time.Hour * 2,
			wantErr:    false,
		},
		{
			name:       "short expiration",
			uid:        "12345",
			username:   "test_user",
			issuer:     "short-issuer",
			secret:     "short-secret",
			expiration: time.Minute * 5,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 生成自定义token
			token, err := GenerateTokenWithExpiration(tt.uid, tt.username, tt.issuer, tt.secret, tt.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTokenWithExpiration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// 使用相同的密钥解析token
			claims, err := ParseTokenWithSecret(token, tt.secret)
			if err != nil {
				t.Errorf("ParseTokenWithSecret() error = %v", err)
				return
			}

			// 验证解析出的信息
			if claims.Uid != tt.uid {
				t.Errorf("ParseToken() uid = %v, want %v", claims.Uid, tt.uid)
			}
			if claims.Username != tt.username {
				t.Errorf("ParseToken() username = %v, want %v", claims.Username, tt.username)
			}
			if claims.Issuer != tt.issuer {
				t.Errorf("ParseToken() issuer = %v, want %v", claims.Issuer, tt.issuer)
			}

			// 验证过期时间
			expectedExpiry := time.Now().Add(tt.expiration).Unix()
			// 允许1秒的误差
			if abs(claims.ExpiresAt-expectedExpiry) > 1 {
				t.Errorf("ParseToken() expiresAt = %v, want close to %v", claims.ExpiresAt, expectedExpiry)
			}
		})
	}
}

// abs 返回整数的绝对值
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
