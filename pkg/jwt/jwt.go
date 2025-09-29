package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YaoAzure/wsgateway/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/do/v2"
)

var (
	ErrDecodeJWTTokenFailed   = errors.New("JWT令牌解析失败")
	ErrInvalidJWTToken        = errors.New("无效的令牌")
	ErrSupportedSignAlgorithm = errors.New("不支持的签名算法")
)

type MapClaims jwt.MapClaims

// Token JWT令牌处理器，封装了JWT的编码和解码功能
type Token struct {
	key    string // JWT 密钥，生成和验证 JWT Token 签名时使用
	issuer string // JWT 令牌的签发者，通常是应用服务名
}

func NewToken(i do.Injector) (*Token, error) {
	jwtConfig := do.MustInvoke[config.JWTConfig](i)
	return &Token{
		key:    jwtConfig.Key,
		issuer: jwtConfig.Issuer,
	}, nil
}

// Encode 生成 JWT Token，支持自定义声明和自动添加标准声明
// customClaims: 用户自定义的声明信息
func (t *Token) Encode(customClaims MapClaims) (string, error) {
	// 合并自定义声明和默认声明
	claims := jwt.MapClaims{
		"iat": time.Now().Unix(), // Token 签发时间戳
		"iss": t.issuer,          // Token 签发者
	}
	// 合并用户自定义声明（覆盖默认声明）
	for k, v := range customClaims {
		claims[k] = v
	}
	// 自动处理过期时间，默认 24h
	const exp = 24 * time.Hour
	if _, ok := claims["exp"]; !ok {
		claims["exp"] = time.Now().Add(exp).Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.key))
}

// Decode 解码JWT令牌并返回声明信息
// tokenString: 待解码的JWT令牌字符串，支持Bearer前缀
func (t *Token) Decode(tokenString string) (MapClaims, error) {
	// 移除可能的 Bearer 前缀
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// 解析 JWT 令牌
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// 验证签名算法是否为 HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrSupportedSignAlgorithm, token.Header["alg"])
		}
		return []byte(t.key), nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeJWTTokenFailed, err)
	}
	// 验证令牌是否有效
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return MapClaims(claims), nil
	}
	return nil, fmt.Errorf("%w", ErrInvalidJWTToken)
}
