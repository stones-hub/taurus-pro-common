package tcrypt

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func TestGenerateAndParseToken(t *testing.T) {
	tests := []struct {
		name     string
		uid      uint
		username string
		wantErr  bool
	}{
		{
			name:     "valid user",
			uid:      12345,
			username: "test_user",
			wantErr:  false,
		},
		{
			name:     "zero uid",
			uid:      0,
			username: "test_user",
			wantErr:  false,
		},
		{
			name:     "empty username",
			uid:      12345,
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
			if claims.Issuer != "cap-gin" {
				t.Errorf("ParseToken() issuer = %v, want cap-gin", claims.Issuer)
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
	// 修改 JwtSecret 的备份
	originalSecret := JwtSecret
	defer func() {
		JwtSecret = originalSecret
	}()

	// 使用测试密钥
	JwtSecret = []byte("test_secret")

	// 创建一个已经过期的token（过期时间设置为当前时间前1秒）
	nowTime := time.Now()
	expiredTime := nowTime.Add(-1 * time.Second)

	claims := Claims{
		Uid:      12345,
		Username: "test_user",
		StandardClaims: jwt.StandardClaims{
			NotBefore: nowTime.Add(-2 * time.Second).Unix(), // 2秒前生效
			ExpiresAt: expiredTime.Unix(),                   // 1秒前过期
			Issuer:    "cap-gin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtSecret)
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
