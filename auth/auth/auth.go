package auth

import (
	"errors"
	"fmt"
	"github.com/gobackpack/crypto"
	"github.com/gobackpack/examples/auth/auth/cache"
	"github.com/gobackpack/jwt"
	"github.com/google/uuid"
	"strconv"
	"time"
)

var (
	AccessTokenSecret  = []byte("secret-access-123")
	RefreshTokenSecret = []byte("secret-refresh-123")
	AccessTokenExpiry  = time.Minute * 15
	RefreshTokenExpiry = time.Hour * 24 * 7
)

type Service struct {
	Cache
}

type Cache interface {
	Store(items ...*cache.Item) error
	Get(keys ...string) ([]byte, error)
	Delete(keys ...string) error
}

type TokenDetails struct {
	AccessToken        string
	AccessTokenUuid    string
	AccessTokenExpiry  time.Duration
	RefreshToken       string
	RefreshTokenUuid   string
	RefreshTokenExpiry time.Duration
}

func (authSvc *Service) RegisterUser(user *User) (*User, error) {
	existing := getUser(user.Email)
	if existing != nil {
		return nil, errors.New("user email is already registered: " + user.Email)
	}

	argon := crypto.NewArgon2()
	argon.Plain = user.Password

	if err := argon.Hash(); err != nil {
		return nil, err
	}

	user.Password = argon.Hashed

	if err := saveUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (authSvc *Service) Authenticate(email, password string) (map[string]string, error) {
	user := getUser(email)
	if user == nil {
		return nil, errors.New("user email not registered: " + email)
	}

	if valid := validateCredentials(user, password); valid {
		tokens, err := authSvc.createAuth(user)
		if err != nil {
			return nil, err
		}

		return tokens, nil
	}

	return nil, errors.New("invalid credentials")
}

func (authSvc *Service) DestroyAuthenticationSession(accessToken string) error {
	claims, valid := extractAccessToken(accessToken)
	if !valid {
		return errors.New("invalid access_token")
	}

	accessTokenUuid := claims["uuid"]
	if accessTokenUuid == nil {
		return errors.New("invalid uuid claims from access_token")
	}

	// TODO: remove refresh token uuid from cache : low priority

	if err := authSvc.Cache.Delete(fmt.Sprint(accessTokenUuid)); err != nil {
		return err
	}

	return nil
}

func (authSvc *Service) RefreshToken(refreshToken string) (map[string]string, error) {
	// get old refresh token uuid so it can be deleted from cache
	claims, valid := extractRefreshToken(refreshToken)
	if !valid {
		return nil, errors.New("invalid refresh_token")
	}

	refreshTokenUuid := claims["uuid"]
	userId := claims["sub"]
	userEmail := claims["email"]
	if refreshTokenUuid == nil || userId == nil || userEmail == nil {
		return nil, errors.New("invalid claims from refresh_token")
	}

	uId, err := strconv.Atoi(fmt.Sprint(userId))
	if err != nil {
		return nil, errors.New("invalid user id from refresh_token claims")
	}

	// TODO: remove access token uuid from cache : low priority

	// delete refresh token uuid
	if err := authSvc.Cache.Delete(fmt.Sprint(refreshTokenUuid)); err != nil {
		return nil, err
	}

	// generate new access token and refresh token
	user := &User{
		Id:    uint(uId),
		Email: fmt.Sprint(userEmail),
	}

	tokens, err := authSvc.createAuth(user)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func createTokens(user *User) (*TokenDetails, error) {
	accessTokenUuid := uuid.New().String()
	// access_token
	accessToken := &jwt.Token{
		Secret: AccessTokenSecret,
	}
	accessTokenStr, err := accessToken.Generate(map[string]interface{}{
		"sub":   user.Id,
		"email": user.Email,
		"uuid":  accessTokenUuid,
		"exp":   jwt.TokenExpiry(AccessTokenExpiry),
	})
	if err != nil {
		return nil, err
	}

	// refresh_token
	refreshToken := &jwt.Token{
		Secret: RefreshTokenSecret,
	}
	refreshTokenUuid := uuid.New().String()
	refreshTokenTokenStr, err := refreshToken.Generate(map[string]interface{}{
		"sub":   user.Id,
		"uuid":  refreshTokenUuid,
		"email": user.Email,
		"exp":   jwt.TokenExpiry(RefreshTokenExpiry),
	})
	if err != nil {
		return nil, err
	}

	return &TokenDetails{
		AccessToken:        accessTokenStr,
		AccessTokenUuid:    accessTokenUuid,
		AccessTokenExpiry:  AccessTokenExpiry,
		RefreshToken:       refreshTokenTokenStr,
		RefreshTokenUuid:   refreshTokenUuid,
		RefreshTokenExpiry: RefreshTokenExpiry,
	}, nil
}

func extractAccessToken(tokenStr string) (map[string]interface{}, bool) {
	token := &jwt.Token{
		Secret: AccessTokenSecret,
	}

	return token.ValidateAndExtract(tokenStr)
}

func extractRefreshToken(tokenStr string) (map[string]interface{}, bool) {
	token := &jwt.Token{
		Secret: RefreshTokenSecret,
	}

	return token.ValidateAndExtract(tokenStr)
}

func validateCredentials(user *User, password string) bool {
	argon := crypto.NewArgon2()

	argon.Hashed = user.Password
	argon.Plain = password

	return argon.Validate()
}

func (authSvc *Service) createAuth(user *User) (map[string]string, error) {
	tokenDetails, err := createTokens(user)
	if err != nil {
		return nil, err
	}

	if err := authSvc.Cache.Store(
		&cache.Item{
			Key:        tokenDetails.AccessTokenUuid,
			Value:      user.Id,
			Expiration: tokenDetails.AccessTokenExpiry,
		}, &cache.Item{
			Key:        tokenDetails.RefreshTokenUuid,
			Value:      user.Id,
			Expiration: tokenDetails.RefreshTokenExpiry,
		}); err != nil {
		return nil, err
	}

	tokens := map[string]string{
		"access_token":  tokenDetails.AccessToken,
		"refresh_token": tokenDetails.RefreshToken,
	}

	return tokens, nil
}
