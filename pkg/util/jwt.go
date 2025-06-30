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
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var JwtSecret = []byte("61647649@qq.com") // 声明签名信息

// Claims 自定义有效载荷
type Claims struct {
	Uid                uint   `json:"uid"`
	Username           string `json:"username"`
	jwt.StandardClaims        // StandardClaims结构体实现了Claims接口(Valid()函数)
}

// GenerateToken 签发token（调用jwt-go库生成token）, 传入用户名和ID 返回一个token字符串. 用户登录成功签发token
func GenerateToken(uid uint, username string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Hour * 24)
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

// ParseToken token解码, 传入token字符串， 解析出Claims结构体. 用户请求携带token， 解析出Claims结构体
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

// ------------------  例子 ------------------
/*
登录成功的时候需要签发JWT, 分发token（其他功能需要身份验证，给前端存储的）,
token返回给前端以后， 前端需要存储下来，每次请求要带过来, 服务端需要将token存储到Redis, 方便后面校验

// -----> 登录成功，存token <-----
token, err := util.GenerateToken(user.ID, user.UserName) // 生产token
if err != nil {
	return httpx.Response{
		Status: http.StatusInternalServerError,
		Msg:    "token签发失败！",
		Error:  err.Error(),
	}
}
// 签发token后，存储到redis中（为了保证token唯一有效）
ua := r.Header.Get("User-Agent") // key用户user-agent可以保证换了浏览器token失效
m := map[string]string{ua: token}
redisx.Redis.HSet(r.Context(), strconv.FormatUint(uint64(user.ID), 10), ua, m)
return response.Response{
	Status: http.StatusOK,
	Msg:    "登录成功！",
	Data:   map[string]string{"token": token},
}

// -----> 中间件校验token <------
// 判断该token是不是最新token（从redis里查）
ua := r.Header.Get("User-Agent")
// 存的时候  key = userid  value = map["User-Agent"]token, 取的时候  取 UID 对于的 map里面的key="User-Agent"对应的值
val, err := redisx.Redis.HGet(r.Context(), strconv.Itoa(int(claims.Uid)), ua).Result()
*/
