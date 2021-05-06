package auth

import (
	"errors"
	"github.com/gobackpack/crypto"
	"github.com/gobackpack/examples/auth/auth/cache"
	"github.com/gobackpack/jwt"
	"github.com/google/uuid"
	"sort"
	"time"
)

var (
	AccessTokenSecret  = []byte("secret-access-123")
	RefreshTokenSecret = []byte("secret-refresh-123")
	AccessTokenExpiry  = time.Minute * 15
	RefreshTokenExpiry = time.Hour * 24 * 7
)

var Users = []*User{
	{
		Id:    1,
		Email: "semir@mail.com",
		// test123
		Password: "19$65536$3$2$459702e19e548205e3803414fd4af86cc3db3a2eefa8332d1ccda7f6acd92aeb$2e55b641dd9b1b0c8af506a5ea8c8201513f1f316cef3fb3c14371e9e6cc1890",
	},
	{
		Id:    2,
		Email: "semir_2@mail.com",
		// test123
		Password: "19$65536$3$2$459702e19e548205e3803414fd4af86cc3db3a2eefa8332d1ccda7f6acd92aeb$2e55b641dd9b1b0c8af506a5ea8c8201513f1f316cef3fb3c14371e9e6cc1890",
	},
}

type Service struct {
	Cache
}

type TokenDetails struct {
	AccessToken        string
	AccessTokenUuid    string
	AccessTokenExpiry  time.Duration
	RefreshToken       string
	RefreshTokenUuid   string
	RefreshTokenExpiry time.Duration
}

type Cache interface {
	Store(items ...*cache.Item) error
	Get(keys ...string) ([]byte, error)
	Delete(keys ...string) error
}

type User struct {
	Id       uint
	Password string `json:"-"`
	Email    string
}

func (authSvc *Service) RegisterUser(email, password string) (*User, error) {
	existing := getUser(email)
	if existing != nil {
		return nil, errors.New("user email is already registered: " + email)
	}

	argon := crypto.NewArgon2()
	argon.Plain = password

	if err := argon.Hash(); err != nil {
		return nil, err
	}

	user := &User{
		Email:    email,
		Password: argon.Hashed,
	}

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

// TODO: Provide implementation
func getUser(email string) *User {
	for _, u := range Users {
		if u.Email == email {
			return u
		}
	}

	return nil
}

// TODO: Provide implementation
func saveUser(user *User) error {
	sort.Slice(Users, func(i, j int) bool {
		return Users[i].Id > Users[j].Id
	})

	user.Id = Users[0].Id + 1

	Users = append(Users, user)

	return nil
}
