package tcrypt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

// Claims 定义JWT的自定义声明（Claims）结构
// 包含用户ID、用户名和标准JWT声明
//
// 字段：
//   - Uid: 用户ID
//   - Username: 用户名
//   - StandardClaims: JWT标准声明（包含过期时间、签发时间等）
type Claims struct {
	Uid                string `json:"uid"`
	Username           string `json:"username"`
	jwt.StandardClaims        // StandardClaims结构体实现了Claims接口(Valid()函数)
}

// TokenParams 定义生成 Token 时可选的标准声明参数
type TokenParams struct {
	Id        string        // jti 唯一ID，用于标识这个token, 比如：1234567890 √
	ExpiresIn time.Duration // 过期时长；<=0 表示不设置 exp，过期时间, 比如：1小时 √
	Issuer    string        // iss 签发者，哪个服务/系统签发的, 比如：taurus-pro √
	NotBefore time.Time     // nbf 生效时间；为零值不设置，生效时间, 比如：1小时前 √
	Subject   string        // sub 主体，哪个用户在使用这个token, 比如：user-1234567890
	Audience  string        // aud 受众，客户端是哪个, 比如：mobile-app
	IssuedAt  time.Time     // iat 签发时间；为零值使用当前时间，签发时间, 比如：当前时间 √
}

// GenerateToken 生成JWT令牌
// 参数：
//   - uid: 用户ID
//   - username: 用户名
//
// 返回值：
//   - string: 生成的JWT令牌字符串
//   - error: 生成过程中的错误，如果成功则为nil
func GenerateToken(uid string, username string) (string, error) {
	return GenerateTokenWithExpiration(uid, username, "taurus-pro", "61647649@qq.com", time.Hour*24)
}

// GenerateTokenWithExpiration 生成具有指定过期时间的JWT令牌
func GenerateTokenWithExpiration(uid string, username string, issuer string, secret string, expiration time.Duration) (string, error) {
	nowTime := time.Now()
	return GenerateTokenWithParams(uid, username, secret, TokenParams{
		Id:        uuid.New().String(), // 唯一ID
		ExpiresIn: expiration,          // 过期时长；<=0 表示不设置 exp
		Issuer:    issuer,              // 签名颁发者
		NotBefore: nowTime,             // 签名生效时间
		Subject:   "none",              // 用户是谁
		Audience:  "api",               // 客户端是谁
		IssuedAt:  nowTime,             // 签发时间
	})
}

// defaultString 返回非空字符串，如果为空则返回默认值
func defaultString(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

// defaultTime 返回非零时间，如果为零值则返回默认值
func defaultTime(value, defaultValue time.Time) time.Time {
	if !value.IsZero() {
		return value
	}
	return defaultValue
}

// GenerateTokenWithParams 使用完整标准声明参数生成 JWT
func GenerateTokenWithParams(uid string, username string, secret string, p TokenParams) (string, error) {
	now := time.Now()

	// 设置默认值
	id := defaultString(p.Id, uuid.New().String())
	issuer := defaultString(p.Issuer, "taurus-pro")
	subject := defaultString(p.Subject, username)
	audience := defaultString(p.Audience, "api")
	issuedAt := defaultTime(p.IssuedAt, now)

	// 计算过期时间
	var expiresAt int64
	if p.ExpiresIn > 0 {
		expiresAt = now.Add(p.ExpiresIn).Unix()
	}

	// 计算生效时间
	notBefore := now.Unix()
	if !p.NotBefore.IsZero() {
		notBefore = p.NotBefore.Unix()
	}

	claims := Claims{
		Uid:      uid,
		Username: username,
		StandardClaims: jwt.StandardClaims{
			Audience:  audience,
			ExpiresAt: expiresAt,
			Id:        id,
			IssuedAt:  issuedAt.Unix(),
			Issuer:    issuer,
			NotBefore: notBefore,
			Subject:   subject,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析JWT令牌并验证其有效性
// 参数：
//   - tokenString: JWT令牌字符串
//
// 返回值：
//   - *Claims: 解析出的声明信息，包含用户ID和用户名
//   - error: 解析过程中的错误，如果成功则为nil
func ParseToken(token string) (*Claims, error) {
	return ParseTokenWithSecret(token, "61647649@qq.com")
}

// ParseTokenWithSecret 使用指定密钥解析JWT令牌并验证其有效性
// 参数：
//   - token: JWT令牌字符串
//   - secret: 用于验证签名的密钥
//
// 返回值：
//   - *Claims: 解析出的声明信息，包含用户ID和用户名
//   - error: 解析过程中的错误，如果成功则为nil
//
// 注意：内部调用 ParseTokenWithSecretAndOptions，只进行基本的签名和时间校验
// jwtToken.Valid 会自动调用 StandardClaims.Valid() 方法，该方法仅校验时间点的合法性：
//   - ExpiresAt (exp): 检查当前时间是否超过过期时间 (now <= exp)
//   - IssuedAt (iat): 检查签发时间是否在未来 (now >= iat)，仅校验时间点合法性，不做业务校验
//   - NotBefore (nbf): 检查当前时间是否早于生效时间 (now >= nbf)
//     注意：这些校验只验证时间点是否合理，不会校验 Subject/Issuer/Audience 等业务字段
func ParseTokenWithSecret(token string, secret string) (*Claims, error) {
	return ParseTokenWithSecretAndOptions(token, secret, ParseOptions{})
}

// ParseOptions 定义额外的校验选项
// 注意：ExpiresAt、IssuedAt、NotBefore 由 jwt-go 包的 StandardClaims.Valid() 自动校验
// 但该校验仅检查时间点的合法性（不能过期、签发时间不能在未来、必须已生效），
// 不进行业务逻辑校验，业务字段如 Subject/Issuer/Audience 需要在此处手动校验
type ParseOptions struct {
	ExpectedAudience string                 // 期望的 aud；为空则不校验 客户端
	RequireAudience  bool                   // 是否强制要求 aud 存在 客户端
	ExpectedIssuer   string                 // 期望的 iss；为空则不校验 签发者
	RequireIssuer    bool                   // 是否强制要求 iss 存在 签发者
	ExpectedSubject  string                 // 期望的 sub；为空则不校验 主体（jwt-go 不会自动校验，需要手动校验）
	RequireSubject   bool                   // 是否强制要求 sub 存在 主体
	JTIValidator     func(jti string) error // jti 校验回调；返回非 nil 视为无效（如重放/黑名单）
}

// ParseTokenWithSecretAndOptions 解析并按选项做额外校验（aud/iss/sub/jti）
// 说明：
//   - ExpiresAt、IssuedAt、NotBefore 由 jwtToken.Valid 自动校验（通过 StandardClaims.Valid()）
//   - Subject 需要在此处手动校验（jwt-go 不自动校验）
func ParseTokenWithSecretAndOptions(token string, secret string, opts ParseOptions) (*Claims, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	// jwtToken.Valid 内部调用 claims.Valid()，仅校验时间点合法性：
	// - ExpiresAt: 当前时间 <= 过期时间（未过期）
	// - IssuedAt: 当前时间 >= 签发时间（签发时间不在未来，仅检查时间点合法性）
	// - NotBefore: 当前时间 >= 生效时间（已生效）
	claims, ok := jwtToken.Claims.(*Claims)
	if !ok || !jwtToken.Valid {
		return nil, errors.New("token不可用")
	}

	// 校验 Audience（如配置）
	if opts.ExpectedAudience != "" || opts.RequireAudience {
		if !claims.VerifyAudience(opts.ExpectedAudience, opts.RequireAudience) {
			return nil, errors.New("audience 不匹配或缺失")
		}
	}

	// 校验 Issuer（如配置）
	if opts.ExpectedIssuer != "" || opts.RequireIssuer {
		if !claims.VerifyIssuer(opts.ExpectedIssuer, opts.RequireIssuer) {
			return nil, errors.New("issuer 不匹配或缺失")
		}
	}

	// 校验 Subject（需要手动校验，jwt-go 不自动校验）
	if opts.ExpectedSubject != "" || opts.RequireSubject {
		if opts.RequireSubject && claims.Subject == "" {
			return nil, errors.New("subject 缺失")
		}
		if opts.ExpectedSubject != "" {
			if claims.Subject != opts.ExpectedSubject {
				return nil, errors.New("subject 不匹配")
			}
		}
	}

	// 校验 JTI（如配置）
	if opts.JTIValidator != nil {
		if err := opts.JTIValidator(claims.Id); err != nil {
			return nil, err
		}
	}

	return claims, nil
}

// GetIssuedAt 获取 token 的签发时间（Unix 时间戳）
// 参数：
//   - claims: 解析出的 Claims
//
// 返回值：
//   - int64: 签发时间（Unix 时间戳）
func GetIssuedAt(claims *Claims) int64 {
	return claims.IssuedAt
}

// GetIssuedAtTime 获取 token 的签发时间（time.Time）
// 参数：
//   - claims: 解析出的 Claims
//
// 返回值：
//   - time.Time: 签发时间
func GetIssuedAtTime(claims *Claims) time.Time {
	return time.Unix(claims.IssuedAt, 0)
}
