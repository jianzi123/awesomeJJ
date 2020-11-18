package jwt

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"reflect"
	"time"
)

//func Parse()  {
//
//	// sample token string taken from the New example
//	tokenString := "eyJhbGciOiJSUzI1NiIsImtpZCI6Ijc5NDM5ODQxNTQxMjA0ODk4MTYifQ.eyJlbWFpbCI6InlmZEAxMjYuY29tIiwibmFtZSI6InlmZF9hZG1pbiIsIm1vYmlsZSI6IjE1NTAxMTU2ODU3IiwicGhvbmVOdW1iZXIiOiIxNTUwMTE1Njg1NyIsImV4dGVybmFsSWQiOiI4NDM4MTQyNjI5NTY4NTk1MTk5IiwidWRBY2NvdW50VXVpZCI6ImE2ODA2MDU1Mzk4ZTljM2UwMzZiZTI5YzQwNWFlNTRhRTNXU0o2TE1jWHMiLCJvdUlkIjoiNDYzMzcwODk0MzEwNDQ5NzMzNSIsIm91TmFtZSI6IueMv-i-heWvvOaVmeiCsiIsInB1cmNoYXNlSWQiOiJ5ZmRqd3Q5Iiwib3BlbklkIjpudWxsLCJpZHBVc2VybmFtZSI6InlmZF9hZG1pbiIsInVzZXJuYW1lIjoid2FuZ3NodWFpamlhbiIsImFwcGxpY2F0aW9uTmFtZSI6Imt1YmVjdGxfMSIsImVudGVycHJpc2VJZCI6InlmZCIsImluc3RhbmNlSWQiOiJ5ZmQiLCJhbGl5dW5Eb21haW4iOiIiLCJwc1N5c3RlbVByaXZpbGVnZXMiOlt7InBlcm1pc3Npb25zIjpbeyJuYW1lIjoiUkFESVVTIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0FVVEhfUkFESVVTIn0seyJuYW1lIjoi6K-B5Lmm566h55CGIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX1BLSV9NR1QifSx7Im5hbWUiOiLlupTnlKjmjojmnYMiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfQVVUSF9BUFAifSx7Im5hbWUiOiLlhbbku5bnrqHnkIYiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfT1RIRVJTX01HVCJ9LHsibmFtZSI6IuamguiniCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9PVkVSVklFVyJ9LHsibmFtZSI6IuaTjeS9nOaXpeW_lyIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9PUEVSQVRFX0xPRyJ9LHsibmFtZSI6Iuadg-mZkOezu-e7nyIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9QU19NR1QifSx7Im5hbWUiOiLorr7nva4iLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfU0VUVElORyJ9LHsibmFtZSI6IuiupOivgea6kCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9BVVRIX1NPVVJDRV9SRUFMIn0seyJuYW1lIjoi5Lmm562-IiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0JPT0tNQVJLIn0seyJuYW1lIjoi6L-bL-WHuuaXpeW_lyIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9MT0dJTl9PVVRfTE9HIn0seyJuYW1lIjoi5bqU55So5YiX6KGoIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0FQUF9MSVNUIn0seyJuYW1lIjoi6K6k6K-BIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0FVVEhfU09VUkNFIn0seyJuYW1lIjoi5Liq5oCn5YyW6K6-572uIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX1NQRUNJQUxfU0VUVElORyJ9LHsibmFtZSI6IuWIhuexu-euoeeQhiIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9VRF9DTEFTU0lGWV9NR1QifSx7Im5hbWUiOiLlronlhajorr7nva4iLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfU0VDVVJJVFlfU0VUVElORyJ9LHsibmFtZSI6IuS8muivneeuoeeQhiIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9TRVNTSU9OX01HVCJ9LHsibmFtZSI6IuW6lOeUqCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9BUFAifSx7Im5hbWUiOiLmjojmnYMiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfQVVUSE9SSVpBVElPTiJ9LHsibmFtZSI6Iui0puaIt-euoeeQhiIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9VU0VSX01HVCJ9LHsibmFtZSI6IuWQjOatpeS4reW_gyIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9TWU5DX01HVCJ9LHsibmFtZSI6IuWIhue6p-euoeeQhiIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9ISUVSQVJDSFlfTUdUIn0seyJuYW1lIjoi5raI5oGv566h55CGIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX01TR19NR1QifSx7Im5hbWUiOiLnlKjmiLfnm67lvZUiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfVUQifSx7Im5hbWUiOiLotYTmupDnrqHnkIYiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfUFNfTUdUX1NPVVJDRSJ9LHsibmFtZSI6IuacuuaehOWPiue7hCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9VRF9NR1QifSx7Im5hbWUiOiLmiJHnmoTmtojmga8iLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfTVlfTVNHIn0seyJuYW1lIjoi6KeS6Imy566h55CGIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX1BTX01HVF9ST0xFIn0seyJuYW1lIjoi5a6h5om55Lit5b-DIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0FQUFJPVkFMX01HVCJ9LHsibmFtZSI6Iua3u-WKoOW6lOeUqCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9BUFBfUExVUyJ9LHsibmFtZSI6IuWuoeiuoSIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9BVURJVCJ9XSwicm9sZXMiOltdLCJzeXN0ZW1JZCI6ImlkcF9wcyIsInN5c3RlbU5hbWUiOiJJRGFhU-adg-mZkOezu-e7nyJ9XSwiZXh0ZW5kRmllbGRzIjp7ImFwcE5hbWUiOiJrdWJlY3RsXzEifSwiZXhwIjoxNjA1MzQzODQ3LCJqdGkiOiJuNVZOMHBiN3o2TnF6N2dtODE3Zi1nIiwiaWF0IjoxNjA1MzQzMjQ3LCJuYmYiOjE2MDUzNDMxODcsInN1YiI6IndhbmdzaHVhaWppYW4ifQ.AnP0s0bnQqVfuFBukr0bQf4Uro6sN-k5uns-ga3935P0kwCdSIts5zAsLISAI0GXBqL8_AozvSP384jx6DCRWQF04BoLvVckeyFVs7NhXvsGs586ce3sAfcZWqRXjP-ASbAds2yG5x8wh4BOSB-qM-t0EhFB3gDoPW6gHAtv5FzAj17DRrPeoRQB4QFQlJdibirFUXGEWWkBxl9JmeFAMXYtlRg-6UXL1W55olR2wao_1rGKLlXoFfwnsRZ8AEFIr6xRhx7t1G3jeA_6SFf8_0iNa4cCCCes4pkax99_Kqdw5CYwBM7Pyb5wHjersrpU5JeqxxTCZtIyEkGF36-PuQ"
//	// Parse takes the token string and a function for looking up the key. The latter is especially
//	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
//	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
//	// to the callback, providing flexibility.
//	token, err := jwtgo.Parse(tokenString, func(token *jwtgo.Token) (interface{}, error) {
//		// Don't forget to validate the alg is what you expect:
//		if _, ok := token.Method.(*jwtgo.SigningMethodHMAC); !ok {
//			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
//		}
//
//		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
//		return hmacSampleSecret, nil
//		if tokenSign, ok := token.Method.(*jwtgo.SigningMethodRSA); ok {
//			tokenSign.Verify()
//		}
//	})
//
//	if claims, ok := token.Claims.(jwtgo.MapClaims); ok && token.Valid {
//		fmt.Println(claims["foo"], claims["nbf"])
//	} else {
//		fmt.Println(err)
//	}
//}
type TokenInfo struct{
	Expired int64 `json:"expired"`
	UserName string `json:"username"`
	Uid string `json:"uid"`
	GroupName string `json:"groupName"`
	Token string `json:"token"`
	IsValid bool `json:"isValid"`
}
func Parse(tokenStr, pkPath string) (TokenInfo, error){
	logrus.Infof("++++++ parse token: %s", tokenStr)
	info := TokenInfo{Token: tokenStr}
	// tokenStr := "eyJhbGciOiJSUzI1NiIsImtpZCI6Ijc5NDM5ODQxNTQxMjA0ODk4MTYifQ.eyJlbWFpbCI6InlmZEAxMjYuY29tIiwibmFtZSI6InlmZF9hZG1pbiIsIm1vYmlsZSI6IjE1NTAxMTU2ODU3IiwicGhvbmVOdW1iZXIiOiIxNTUwMTE1Njg1NyIsImV4dGVybmFsSWQiOiI4NDM4MTQyNjI5NTY4NTk1MTk5IiwidWRBY2NvdW50VXVpZCI6ImE2ODA2MDU1Mzk4ZTljM2UwMzZiZTI5YzQwNWFlNTRhRTNXU0o2TE1jWHMiLCJvdUlkIjoiNDYzMzcwODk0MzEwNDQ5NzMzNSIsIm91TmFtZSI6IueMv-i-heWvvOaVmeiCsiIsInB1cmNoYXNlSWQiOiJ5ZmRqd3Q5Iiwib3BlbklkIjpudWxsLCJpZHBVc2VybmFtZSI6InlmZF9hZG1pbiIsInVzZXJuYW1lIjoid2FuZ3NodWFpamlhbiIsImFwcGxpY2F0aW9uTmFtZSI6Imt1YmVjdGxfMSIsImVudGVycHJpc2VJZCI6InlmZCIsImluc3RhbmNlSWQiOiJ5ZmQiLCJhbGl5dW5Eb21haW4iOiIiLCJwc1N5c3RlbVByaXZpbGVnZXMiOlt7InBlcm1pc3Npb25zIjpbeyJuYW1lIjoiUkFESVVTIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0FVVEhfUkFESVVTIn0seyJuYW1lIjoi6K-B5Lmm566h55CGIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX1BLSV9NR1QifSx7Im5hbWUiOiLlupTnlKjmjojmnYMiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfQVVUSF9BUFAifSx7Im5hbWUiOiLlhbbku5bnrqHnkIYiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfT1RIRVJTX01HVCJ9LHsibmFtZSI6IuamguiniCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9PVkVSVklFVyJ9LHsibmFtZSI6IuaTjeS9nOaXpeW_lyIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9PUEVSQVRFX0xPRyJ9LHsibmFtZSI6Iuadg-mZkOezu-e7nyIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9QU19NR1QifSx7Im5hbWUiOiLorr7nva4iLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfU0VUVElORyJ9LHsibmFtZSI6IuiupOivgea6kCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9BVVRIX1NPVVJDRV9SRUFMIn0seyJuYW1lIjoi5Lmm562-IiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0JPT0tNQVJLIn0seyJuYW1lIjoi6L-bL-WHuuaXpeW_lyIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9MT0dJTl9PVVRfTE9HIn0seyJuYW1lIjoi5bqU55So5YiX6KGoIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0FQUF9MSVNUIn0seyJuYW1lIjoi6K6k6K-BIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0FVVEhfU09VUkNFIn0seyJuYW1lIjoi5Liq5oCn5YyW6K6-572uIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX1NQRUNJQUxfU0VUVElORyJ9LHsibmFtZSI6IuWIhuexu-euoeeQhiIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9VRF9DTEFTU0lGWV9NR1QifSx7Im5hbWUiOiLlronlhajorr7nva4iLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfU0VDVVJJVFlfU0VUVElORyJ9LHsibmFtZSI6IuS8muivneeuoeeQhiIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9TRVNTSU9OX01HVCJ9LHsibmFtZSI6IuW6lOeUqCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9BUFAifSx7Im5hbWUiOiLmjojmnYMiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfQVVUSE9SSVpBVElPTiJ9LHsibmFtZSI6Iui0puaIt-euoeeQhiIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9VU0VSX01HVCJ9LHsibmFtZSI6IuWQjOatpeS4reW_gyIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9TWU5DX01HVCJ9LHsibmFtZSI6IuWIhue6p-euoeeQhiIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9ISUVSQVJDSFlfTUdUIn0seyJuYW1lIjoi5raI5oGv566h55CGIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX01TR19NR1QifSx7Im5hbWUiOiLnlKjmiLfnm67lvZUiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfVUQifSx7Im5hbWUiOiLotYTmupDnrqHnkIYiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfUFNfTUdUX1NPVVJDRSJ9LHsibmFtZSI6IuacuuaehOWPiue7hCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9VRF9NR1QifSx7Im5hbWUiOiLmiJHnmoTmtojmga8iLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfTVlfTVNHIn0seyJuYW1lIjoi6KeS6Imy566h55CGIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX1BTX01HVF9ST0xFIn0seyJuYW1lIjoi5a6h5om55Lit5b-DIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0FQUFJPVkFMX01HVCJ9LHsibmFtZSI6Iua3u-WKoOW6lOeUqCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9BUFBfUExVUyJ9LHsibmFtZSI6IuWuoeiuoSIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9BVURJVCJ9XSwicm9sZXMiOltdLCJzeXN0ZW1JZCI6ImlkcF9wcyIsInN5c3RlbU5hbWUiOiJJRGFhU-adg-mZkOezu-e7nyJ9XSwiZXh0ZW5kRmllbGRzIjp7ImFwcE5hbWUiOiJrdWJlY3RsXzEifSwiZXhwIjoxNjA1MzQzODQ3LCJqdGkiOiJuNVZOMHBiN3o2TnF6N2dtODE3Zi1nIiwiaWF0IjoxNjA1MzQzMjQ3LCJuYmYiOjE2MDUzNDMxODcsInN1YiI6IndhbmdzaHVhaWppYW4ifQ.AnP0s0bnQqVfuFBukr0bQf4Uro6sN-k5uns-ga3935P0kwCdSIts5zAsLISAI0GXBqL8_AozvSP384jx6DCRWQF04BoLvVckeyFVs7NhXvsGs586ce3sAfcZWqRXjP-ASbAds2yG5x8wh4BOSB-qM-t0EhFB3gDoPW6gHAtv5FzAj17DRrPeoRQB4QFQlJdibirFUXGEWWkBxl9JmeFAMXYtlRg-6UXL1W55olR2wao_1rGKLlXoFfwnsRZ8AEFIr6xRhx7t1G3jeA_6SFf8_0iNa4cCCCes4pkax99_Kqdw5CYwBM7Pyb5wHjersrpU5JeqxxTCZtIyEkGF36-PuQ"
	//tokenStr := "eyJhbGciOiJSUzI1NiIsImtpZCI6Ijc5NDM5ODQxNTQxMjA0ODk4MTYifQ.eyJlbWFpbCI6InlmZEAxMjYuY29tIiwibmFtZSI6InlmZF9hZG1pbiIsIm1vYmlsZSI6IjE1NTAxMTU2ODU3IiwicGhvbmVOdW1iZXIiOiIxNTUwMTE1Njg1NyIsImV4dGVybmFsSWQiOiI4NDM4MTQyNjI5NTY4NTk1MTk5IiwidWRBY2NvdW50VXVpZCI6ImE2ODA2MDU1Mzk4ZTljM2UwMzZiZTI5YzQwNWFlNTRhRTNXU0o2TE1jWHMiLCJvdUlkIjoiNDYzMzcwODk0MzEwNDQ5NzMzNSIsIm91TmFtZSI6IueMv-i-heWvvOaVmeiCsiIsInB1cmNoYXNlSWQiOiJ5ZmRqd3Q5Iiwib3BlbklkIjpudWxsLCJpZHBVc2VybmFtZSI6InlmZF9hZG1pbiIsInVzZXJuYW1lIjoid2FuZ3NodWFpamlhbiIsImFwcGxpY2F0aW9uTmFtZSI6Imt1YmVjdGxfMSIsImVudGVycHJpc2VJZCI6InlmZCIsImluc3RhbmNlSWQiOiJ5ZmQiLCJhbGl5dW5Eb21haW4iOiIiLCJwc1N5c3RlbVByaXZpbGVnZXMiOlt7InBlcm1pc3Npb25zIjpbeyJuYW1lIjoiUkFESVVTIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0FVVEhfUkFESVVTIn0seyJuYW1lIjoi6K-B5Lmm566h55CGIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX1BLSV9NR1QifSx7Im5hbWUiOiLlupTnlKjmjojmnYMiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfQVVUSF9BUFAifSx7Im5hbWUiOiLlhbbku5bnrqHnkIYiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfT1RIRVJTX01HVCJ9LHsibmFtZSI6IuamguiniCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9PVkVSVklFVyJ9LHsibmFtZSI6IuaTjeS9nOaXpeW_lyIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9PUEVSQVRFX0xPRyJ9LHsibmFtZSI6Iuadg-mZkOezu-e7nyIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9QU19NR1QifSx7Im5hbWUiOiLorr7nva4iLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfU0VUVElORyJ9LHsibmFtZSI6IuiupOivgea6kCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9BVVRIX1NPVVJDRV9SRUFMIn0seyJuYW1lIjoi5Lmm562-IiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0JPT0tNQVJLIn0seyJuYW1lIjoi6L-bL-WHuuaXpeW_lyIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9MT0dJTl9PVVRfTE9HIn0seyJuYW1lIjoi5bqU55So5YiX6KGoIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0FQUF9MSVNUIn0seyJuYW1lIjoi6K6k6K-BIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0FVVEhfU09VUkNFIn0seyJuYW1lIjoi5Liq5oCn5YyW6K6-572uIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX1NQRUNJQUxfU0VUVElORyJ9LHsibmFtZSI6IuWIhuexu-euoeeQhiIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9VRF9DTEFTU0lGWV9NR1QifSx7Im5hbWUiOiLlronlhajorr7nva4iLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfU0VDVVJJVFlfU0VUVElORyJ9LHsibmFtZSI6IuS8muivneeuoeeQhiIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9TRVNTSU9OX01HVCJ9LHsibmFtZSI6IuW6lOeUqCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9BUFAifSx7Im5hbWUiOiLmjojmnYMiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfQVVUSE9SSVpBVElPTiJ9LHsibmFtZSI6Iui0puaIt-euoeeQhiIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9VU0VSX01HVCJ9LHsibmFtZSI6IuWQjOatpeS4reW_gyIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9TWU5DX01HVCJ9LHsibmFtZSI6IuWIhue6p-euoeeQhiIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9ISUVSQVJDSFlfTUdUIn0seyJuYW1lIjoi5raI5oGv566h55CGIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX01TR19NR1QifSx7Im5hbWUiOiLnlKjmiLfnm67lvZUiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfVUQifSx7Im5hbWUiOiLotYTmupDnrqHnkIYiLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfUFNfTUdUX1NPVVJDRSJ9LHsibmFtZSI6IuacuuaehOWPiue7hCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9VRF9NR1QifSx7Im5hbWUiOiLmiJHnmoTmtojmga8iLCJ1cmkiOiIiLCJ2YWx1ZSI6IkVOVEVSUFJJU0VfTVlfTVNHIn0seyJuYW1lIjoi6KeS6Imy566h55CGIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX1BTX01HVF9ST0xFIn0seyJuYW1lIjoi5a6h5om55Lit5b-DIiwidXJpIjoiIiwidmFsdWUiOiJFTlRFUlBSSVNFX0FQUFJPVkFMX01HVCJ9LHsibmFtZSI6Iua3u-WKoOW6lOeUqCIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9BUFBfUExVUyJ9LHsibmFtZSI6IuWuoeiuoSIsInVyaSI6IiIsInZhbHVlIjoiRU5URVJQUklTRV9BVURJVCJ9XSwicm9sZXMiOltdLCJzeXN0ZW1JZCI6ImlkcF9wcyIsInN5c3RlbU5hbWUiOiJJRGFhU-adg-mZkOezu-e7nyJ9XSwiZXh0ZW5kRmllbGRzIjp7ImFwcE5hbWUiOiJrdWJlY3RsXzEifSwiZXhwIjoxNjA1MzcxMTI3LCJqdGkiOiJrQjlmOHBHejdUQVd2V1NwTHJfUGdRIiwiaWF0IjoxNjA1MzcwNTI3LCJuYmYiOjE2MDUzNzA0NjcsInN1YiI6IndhbmdzaHVhaWppYW4ifQ.eauMiLdlPFYig3gMuhgZHNSiCMIM80A41BmDSQvbCtp7Hro6j23E6L17r7cSrbAbs2dobLMDB4QFxYk5lNf3Us8U6XqDUh7XjyOLl0Dswgw-Mip_ZusXLoYmpvdL3U-CDb3jK-0rAdeb9C5rlRACNyWa8Q5lWnetkSxFjPmXvHGzOz9advy-tciLhEDGBg4J3MpZY_k5YanNSINEEVS_jPbEQJ7NrL4hd8eJLhDcLu3fJHuvfaE1mWZKuDg_cT-wRHShQq3gBNeNroDWtDgBZyttvHwfe3u9n8z4UvxKAXtZuFDyZMxToG323zYRyITebDVKkvOpa1dUXkIixFf1vQ"
	parseAuth, err := jwtgo.Parse(tokenStr, func(token *jwtgo.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtgo.SigningMethodHMAC); ok {
			mySingKeyBytes := []byte("7943984154120489816")
			return mySingKeyBytes, nil
		}

		if _, ok := token.Method.(*jwtgo.SigningMethodRSA); ok {
			logrus.Infof("++++SigningMethodRSA start")
			//rawPem, err := ioutil.ReadFile("/Users/wangshuaijian/go_workspace/src/github.com/jianzi123/awesomeJJ/script/public_key")
			rawPem, err := ioutil.ReadFile(pkPath)
			if err != nil {
				logrus.Errorf("Parse token failed: %v", err)
				return nil, err
			}
			//rawPem, err := ioutil.ReadFile("/root/wsj/public_key")
			pemBlock, _ := pem.Decode(rawPem)
			publickey, err := x509.ParsePKCS1PublicKey(pemBlock.Bytes)
			//// read file
			//buff, err := ioutil.ReadFile("/Users/wangshuaijian/go_workspace/src/github.com/jianzi123/awesomeJJ/script/public_key")
			//if err != nil{
			//	logrus.Errorf("read public.pem failed: %v", err)
			//	return nil, err
			//}
			//// generate rsa.PublicKey and return
			//logrus.Infof("++++SigningMethodRSA end")
			//publickey, err := jwtgo.ParseRSAPublicKeyFromPEM(buff)
			if err != nil {
				logrus.Errorf("ParseRSAPublicKeyFromPEM failed: %v", err)
				return nil, err
			}
			return publickey, err
		}
		return nil, nil
	})
	if err != nil {
		logrus.Errorf("parse token failed 1: %v", err)
		return info, err
	}

	//将token中的内容存入parmMap
	if claim, ok := parseAuth.Claims.(jwtgo.MapClaims); ok != true {
		logrus.Errorf("parseAuth.Claims.(jwtgo.MapClaims) not fit")
		return info, fmt.Errorf("parseAuth.Claims.(jwtgo.MapClaims) not fit")
	} else {
		var parmMap map[string]interface{}
		parmMap = make(map[string]interface{})
		for key, val := range claim {
			parmMap[key] = val
		}
		if appName, ok := parmMap["applicationName"]; ok {
			logrus.Infof("appName: %v", appName)
		}
		if exp, ok := parmMap["exp"]; ok {
			logrus.Infof("+++++++++ exp    outdate: %v", exp)
			logrus.Infof("+++++++++ exp    outdate datatype: %s",
			reflect.TypeOf(exp).String())

			if expTime, ok := exp.(float64); ok{
				logrus.Infof("+++++++++ expTime %f", expTime)
				info.Expired = int64(expTime)
			}else{
				info.Expired = time.Now().Unix() + 20
			}
			logrus.Infof("+++++++++ exp    outdate1: %d", info.Expired)
		}
		if sub, ok := parmMap["sub"]; ok {
			logrus.Infof("subname: %v", sub)
		}
		if username, ok := parmMap["username"]; ok {
			if user, ok := username.(string); ok{
				info.UserName = user
			}
			logrus.Infof("username: %v", username)
		}
		if idpUsername, ok := parmMap["idpUsername"]; ok {
			logrus.Infof("idpUsername: %v", idpUsername)
		}
		if purchaseId, ok := parmMap["purchaseId"]; ok {
			logrus.Infof("purchaseId: %v", purchaseId)
		}
		info.IsValid = parseAuth.Valid
		logrus.Infof("parseAuth.Claims: %v", parmMap)
		return info, nil
	}

}
