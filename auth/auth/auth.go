package auth

import (
	"errors"
	"github.com/gobackpack/crypto"
	"github.com/gobackpack/jwt"
	"github.com/sirupsen/logrus"
	"sort"
	"time"
)

var (
	AccessTokenSecret  = []byte("secret-access-123")
	RefreshTokenSecret = []byte("secret-refresh-123")
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

type User struct {
	Id       uint
	Password string `json:"-"`
	Email    string
}

func RegisterUser(email, password string) (*User, error) {
	existing := GetUser(email)
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

	SaveUser(user)

	return user, nil
}

func Authenticate(email, password string) (map[string]string, error) {
	user := GetUser(email)
	if user == nil {
		return nil, errors.New("user email not registered: " + email)
	}

	if valid := validateCredentials(user, password); valid {
		tokens, err := createTokens(user)
		if err != nil {
			return nil, err
		}

		return tokens, nil
	}

	return nil, errors.New("invalid credentials")
}

func IsTokenValid(tokenStr string) bool {
	token := &jwt.Token{
		Secret: AccessTokenSecret,
	}

	_, valid := token.ValidateAndExtract(tokenStr)

	return valid
}

func GetUser(email string) *User {
	for _, u := range Users {
		if u.Email == email {
			return u
		}
	}

	return nil
}

func SaveUser(user *User) {
	sort.Slice(Users, func(i, j int) bool {
		return Users[i].Id > Users[j].Id
	})

	user.Id = Users[0].Id + 1

	Users = append(Users, user)
}

func createTokens(user *User) (map[string]string, error) {
	// access_token
	accessToken := &jwt.Token{
		Secret: AccessTokenSecret,
	}
	accessTokenStr, err := accessToken.Generate(map[string]interface{}{
		"sub":   user.Id,
		"email": user.Email,
		"exp":   jwt.TokenExpiry(time.Second * 15),
	})
	if err != nil {
		return nil, err
	}

	// refresh_token
	refreshToken := &jwt.Token{
		Secret: RefreshTokenSecret,
	}
	refreshTokenTokenStr, err := refreshToken.Generate(map[string]interface{}{
		"sub": user.Id,
		"exp": jwt.TokenExpiry(time.Minute * 1),
	})
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"access_token":  accessTokenStr,
		"refresh_token": refreshTokenTokenStr,
	}, nil
}

func validateCredentials(user *User, password string) bool {
	argon := crypto.NewArgon2()

	argon.Hashed = user.Password
	argon.Plain = password

	return argon.Validate()
}
