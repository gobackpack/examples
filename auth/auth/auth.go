package auth

import (
	"errors"
	"github.com/gobackpack/crypto"
	"github.com/gobackpack/jwt"
	"github.com/sirupsen/logrus"
	"log"
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

	sort.Slice(Users, func(i, j int) bool {
		return Users[i].Id > Users[j].Id
	})

	user := &User{
		Id:    Users[0].Id + 1,
		Email: email,
	}

	argon := crypto.NewArgon2()
	argon.Plain = password

	if err := argon.Hash(); err != nil {
		logrus.Fatal("argon failed to hash Plain: ", err)
	}

	user.Password = argon.Hashed

	Users = append(Users, user)

	return user, nil
}

func Authenticate(user *User, password string) (map[string]string, error) {
	if valid := validateCredentials(user, password); valid {
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
			log.Fatalln("failed to generate jwt: ", err)
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
			log.Fatalln("failed to generate jwt: ", err)
		}

		tokens := map[string]string{
			"access_token":  accessTokenStr,
			"refresh_token": refreshTokenTokenStr,
		}

		return tokens, nil
	}

	return nil, errors.New("invalid credentials")
}

func IsAuthenticated(tokenStr string) bool {
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

func validateCredentials(user *User, password string) bool {
	argon := crypto.NewArgon2()

	argon.Hashed = user.Password
	argon.Plain = password

	return argon.Validate()
}
