package tcrypt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JwtSecret 用于签名和验证JWT的密钥
// 注意：在生产环境中应该使用更安全的方式管理此密钥
// 建议：
//   - 使用环境变量或配置文件
//   - 使用足够长的随机字符串
//   - 定期轮换密钥
//   - 不同环境使用不同的密钥
var JwtSecret = []byte("61647649@qq.com")

// Claims 定义JWT的自定义声明（Claims）结构
// 包含用户ID、用户名和标准JWT声明
//
// 字段：
//   - Uid: 用户ID
//   - Username: 用户名
//   - StandardClaims: JWT标准声明（包含过期时间、签发时间等）
//
// 使用示例：
//
//	claims := &tcrypt.Claims{
//	    Uid: 12345,
//	    Username: "john_doe",
//	    StandardClaims: jwt.StandardClaims{
//	        ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
//	        IssuedAt: time.Now().Unix(),
//	        Issuer: "my-app",
//	    },
//	}
//
// 注意事项：
//   - 继承了jwt.StandardClaims以获取标准JWT字段
//   - 可以通过json标签自定义字段名
//   - 实现了jwt.Claims接口
//   - 用于生成和解析JWT令牌
type Claims struct {
	Uid                uint   `json:"uid"`
	Username           string `json:"username"`
	jwt.StandardClaims        // StandardClaims结构体实现了Claims接口(Valid()函数)
}

// GenerateToken 生成JWT令牌
// 参数：
//   - uid: 用户ID
//   - username: 用户名
//
// 返回值：
//   - string: 生成的JWT令牌字符串
//   - error: 生成过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	token, err := tcrypt.GenerateToken(12345, "john_doe")
//	if err != nil {
//	    log.Printf("生成令牌失败：%v", err)
//	    return
//	}
//	fmt.Printf("JWT令牌：%s\n", token)
//
// 注意事项：
//   - 使用HS256算法签名
//   - 令牌有效期为24小时
//   - 包含用户ID和用户名信息
//   - 使用JwtSecret进行签名
//   - 签发者设置为"cap-gin"
//   - 令牌包含生效时间和过期时间
func GenerateToken(uid uint, username string) (string, error) {
	return GenerateTokenWithExpiration(uid, username, time.Hour*24)
}

// GenerateTokenWithExpiration 生成具有指定过期时间的JWT令牌
func GenerateTokenWithExpiration(uid uint, username string, expiration time.Duration) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(expiration)
	claims := Claims{
		Uid:      uid,
		Username: username,
		StandardClaims: jwt.StandardClaims{
			NotBefore: nowTime.Unix(),    // 签名生效时间
			ExpiresAt: expireTime.Unix(), // 签名过期时间
			Issuer:    "cap-gin",         // 签名颁发者
		},
	}
	// 指定编码算法为jwt.SigningMethodHS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // 返回一个token结构体指针(*Token)
	//tokenString, err := token.SigningString(JwtSecret)
	//return tokenString, err
	return token.SignedString(JwtSecret)
}

// ParseToken 解析JWT令牌并验证其有效性
// 参数：
//   - tokenString: JWT令牌字符串
//
// 返回值：
//   - *Claims: 解析出的声明信息，包含用户ID和用户名
//   - error: 解析过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	token := "eyJhbGciOiJIUzI1NiIs..." // 从请求头或其他地方获取的JWT令牌
//	claims, err := tcrypt.ParseToken(token)
//	if err != nil {
//	    log.Printf("解析令牌失败：%v", err)
//	    return
//	}
//	fmt.Printf("用户ID：%d，用户名：%s\n", claims.Uid, claims.Username)
//
// 注意事项：
//   - 会验证令牌的签名
//   - 会检查令牌是否过期
//   - 使用JwtSecret验证签名
//   - 如果令牌无效会返回错误
//   - 支持HS256算法签名的令牌
//   - 返回的Claims可以直接访问用户信息
func ParseToken(tokenString string) (*Claims, error) {
	// 输入用户token字符串,自定义的Claims结构体对象,以及自定义函数来解析token字符串为jwt的Token结构体指针
	//Keyfunc是匿名函数类型: type Keyfunc func(*Token) (interface{}, error)
	//func ParseWithClaims(tokenString string, claims Claims, keyFunc Keyfunc) (*Token, error) {}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return JwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	// 将token中的claims信息解析出来,并断言成用户自定义的有效载荷结构
	claims, ok := token.Claims.(*Claims)
	if ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("token不可用")
}
