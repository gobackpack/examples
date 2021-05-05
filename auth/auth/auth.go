package auth

import (
	"errors"
	"github.com/gobackpack/crypto"
	"github.com/gobackpack/jwt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
	TokenStore
}

type TokenDetails struct {
	ClientId     string
	AccessToken  string
	RefreshToken string
}

type UserTokenStore struct {
	Device string
}

type TokenStore interface {
	NewOrAppend(*Item) error
	Get(string) []byte
	Delete(string) error
}

type Item struct {
	Key        string
	Value      interface{}
	Expiration time.Duration
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
		logrus.Fatal("argon failed to hash Plain: ", err)
	}

	user := &User{
		Email:    email,
		Password: argon.Hashed,
	}

	saveUser(user)

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

		//if err := authSvc.TokenStore.NewOrAppend(&Item{
		//	Key:        fmt.Sprint(user.Id),
		//	Value:      tokenDetails.ClientId,
		//	Expiration: AccessTokenExpiry,
		//}); err != nil {
		//	return nil, err
		//}

		tokens := map[string]string{
			"access_token":  tokenDetails.AccessToken,
			"refresh_token": tokenDetails.RefreshToken,
		}

		return tokens, nil
	}

	return nil, errors.New("invalid credentials")
}

func createTokens(user *User) (*TokenDetails, error) {
	clientId := uuid.New().String()

	// access_token
	accessToken := &jwt.Token{
		Secret: AccessTokenSecret,
	}
	accessTokenStr, err := accessToken.Generate(map[string]interface{}{
		"sub":       user.Id,
		"email":     user.Email,
		"client_id": clientId,
		"exp":       jwt.TokenExpiry(AccessTokenExpiry),
	})
	if err != nil {
		return nil, err
	}

	// refresh_token
	refreshToken := &jwt.Token{
		Secret: RefreshTokenSecret,
	}
	refreshTokenTokenStr, err := refreshToken.Generate(map[string]interface{}{
		"sub":       user.Id,
		"client_id": clientId,
		"exp":       jwt.TokenExpiry(RefreshTokenExpiry),
	})
	if err != nil {
		return nil, err
	}

	return &TokenDetails{
		ClientId:     clientId,
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenTokenStr,
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
func saveUser(user *User) {
	sort.Slice(Users, func(i, j int) bool {
		return Users[i].Id > Users[j].Id
	})

	user.Id = Users[0].Id + 1

	Users = append(Users, user)
}
