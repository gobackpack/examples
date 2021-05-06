package auth

import (
	"errors"
	"fmt"
	"github.com/gobackpack/crypto"
	"github.com/gobackpack/examples/auth/auth/cache"
	"github.com/gobackpack/jwt"
	"github.com/google/uuid"
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

	return nil, errors.New("invalid credentials")
}

func (authSvc *Service) DestroyAuthenticationSession(accessToken string) error {
	claims, valid := extractToken(accessToken)
	if !valid {
		return errors.New("invalid access_token")
	}

	accessTokenUuid := claims["uuid"]
	if accessTokenUuid == nil {
		return errors.New("invalid access_token")
	}

	if err := authSvc.Cache.Delete(fmt.Sprint(accessTokenUuid)); err != nil {
		return err
	}

	return nil
}

func (authSvc *Service) RefreshToken(refreshToken string) {
	// TODO: Checklist
	// - check if refresh_token is valid
	// - get user_id client_id from refresh_token
	// - remove client_id from cache store for user_id
	// - create new access_token and refresh_token
	// - store client_id in cache store for user_id
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
		"sub":  user.Id,
		"uuid": refreshTokenUuid,
		"exp":  jwt.TokenExpiry(RefreshTokenExpiry),
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

func extractToken(tokenStr string) (map[string]interface{}, bool) {
	token := &jwt.Token{
		Secret: AccessTokenSecret,
	}

	return token.ValidateAndExtract(tokenStr)
}

func validateCredentials(user *User, password string) bool {
	argon := crypto.NewArgon2()

	argon.Hashed = user.Password
	argon.Plain = password

	return argon.Validate()
}
