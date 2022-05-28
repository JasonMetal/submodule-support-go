package jwt

import (
	"creator-platform-api/app/constant"
	"creator-platform-api/app/entity"
	"creator-platform-api/helper/strings"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

func GenerateToken(userId uint32) string {
	claims := entity.ClaimData{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(constant.JwtTTL * time.Hour * time.Duration(1))), // 过期时间
			NotBefore: jwt.NewNumericDate(time.Now()),                                                     // 签发时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                                     // 生效时间
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, _ := t.SignedString(strings.StrToBytes(constant.JwtSecret))

	return token
}

func ParseToken(token string) (*entity.ClaimData, error) {
	t, err := jwt.ParseWithClaims(token, &entity.ClaimData{}, secret())
	if err != nil {
		if eV, ok := err.(*jwt.ValidationError); ok {
			if eV.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("token is no a jwt token")
			} else if eV.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New("token had expired")
			} else {
				return nil, errors.New("parse token error")
			}
		}

		return nil, errors.New("parse token error")
	}
	if claims, ok := t.Claims.(*entity.ClaimData); ok && t.Valid {
		return claims, nil
	}

	return nil, errors.New("parse token error")
}

func secret() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		return strings.StrToBytes(constant.JwtSecret), nil
	}
}
