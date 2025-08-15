package auth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mikeudacha/paybuy/config"
	"github.com/mikeudacha/paybuy/models"
	"github.com/mikeudacha/paybuy/utils"
)

type contextKey string

const UserKey contextKey = "userID"

type JWTClaims struct {
	UserID    string `json:"userID"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

func WithJWTAuth(handlerFunc http.HandlerFunc, store models.UserStore, blacklistStore *BlacklistStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := utils.GetTokenFromRequest(r)

		token, err := ValidateJWT(tokenString, "access")
		if err != nil {
			permissionDenied(w)
			return
		}
		if !token.Valid {
			permissionDenied(w)
			return
		}
		if blacklistStore != nil {
			isBlacklisted, err := blacklistStore.IsBlacklisted(tokenString)
			if err != nil {
				permissionDenied(w)
				return
			}
			if isBlacklisted {
				permissionDenied(w)
				return
			}
		}
		claims := token.Claims.(*JWTClaims)
		str := claims.UserID

		userID, err := strconv.Atoi(str)
		if err != nil {
			permissionDenied(w)
			return
		}

		u, err := store.GetUserByID(userID)
		if err != nil {
			permissionDenied(w)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserKey, u.ID)
		r = r.WithContext(ctx)

		handlerFunc(w, r)
	}
}

func CreateJWT(secret []byte, userID int, tokenType string) (string, error) {
	cfg := config.LoadConfig()

	var expiration time.Duration
	if tokenType == "access" {
		jwtTTL, _ := strconv.Atoi(cfg.JWTExpirationInSeconds)
		expiration = time.Second * time.Duration(jwtTTL)
	} else if tokenType == "refresh" {
		jwtTTL, _ := strconv.Atoi(cfg.JWTRefreshExpirationInSeconds)
		expiration = time.Second * time.Duration(jwtTTL)
	} else {
		return "", fmt.Errorf("invalid token type: %s", tokenType)
	}

	claims := &JWTClaims{
		UserID:    strconv.Itoa(userID),
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, err
}

func CreateTokenPair(userID int) (*models.TokenPair, error) {
	cfg := config.LoadConfig()

	accessToken, err := CreateJWT([]byte(cfg.JWTSecret), userID, "access")
	if err != nil {
		return nil, err
	}

	refreshToken, err := CreateJWT([]byte(cfg.JWTRefreshSecret), userID, "refresh")
	if err != nil {
		return nil, err
	}

	jwtTTL, _ := strconv.Atoi(cfg.JWTExpirationInSeconds)
	expiresIn := int64(jwtTTL)

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func ValidateJWT(tokenString string, expectedType string) (*jwt.Token, error) {
	cfg := config.LoadConfig()

	var secret []byte
	if expectedType == "access" {
		secret = []byte(cfg.JWTSecret)
	} else if expectedType == "refresh" {
		secret = []byte(cfg.JWTRefreshSecret)
	} else {
		return nil, fmt.Errorf("invalid expected token type: %s", expectedType)
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		if claims.TokenType != expectedType {
			return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
		}
	}

	return token, nil
}

func RefreshAccessToken(refreshToken string, blacklistStore *BlacklistStore) (*models.TokenPair, error) {
	token, err := ValidateJWT(refreshToken, "refresh")
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if blacklistStore != nil {
		isBlacklisted, err := blacklistStore.IsBlacklisted(refreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to check blacklist: %w", err)
		}
		if isBlacklisted {
			return nil, fmt.Errorf("refresh token is blacklisted")
		}
	}

	claims := token.Claims.(*JWTClaims)
	userID, err := strconv.Atoi(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return CreateTokenPair(userID)
}

func permissionDenied(w http.ResponseWriter) {
	utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied"))
}

func GetUserIDFromContext(ctx context.Context) int {
	userID, ok := ctx.Value(UserKey).(int)
	if !ok {
		return -1
	}

	return userID
}
