package auth

import (
	"errors"
	"github.com/gobackpack/jwt"
	"sort"
	"time"
)

var Secret = []byte("secret-key-123")

var Users = []*User{
	{
		Id:       1,
		Email:    "semir@mail.com",
		Password: "test123",
	},
	{
		Id:       2,
		Email:    "semir_2@mail.com",
		Password: "test123",
	},
}

type User struct {
	Id       uint
	Password string
	Email    string
}

func RegisterUser(user *User) error {
	existing := getUser(user.Email)
	if existing != nil {
		return errors.New("user email is already registered: " + user.Email)
	}

	sort.Slice(Users, func(i, j int) bool {
		return Users[i].Id > Users[j].Id
	})

	user.Id = Users[0].Id + 1

	Users = append(Users, user)

	return nil
}

func Authenticate(email, password string) (string, error) {
	token := &jwt.Token{
		Secret: Secret,
	}

	// get user from database
	user := getUser(email)
	if user == nil {
		return "", errors.New("there is no such user")
	}

	if valid := validateCredentials(user, password); valid {
		if err := token.Generate(&jwt.Claims{
			Expiration: time.Second * 30,
			Fields: map[string]interface{}{
				"id":    user.Id,
				"email": user.Email,
			},
		}); err != nil {
			return "", err
		}

		return token.Content, nil
	}

	return "", errors.New("invalid credentials")
}

func IsAuthenticated(tokenStr string) bool {
	token := &jwt.Token{
		Secret: Secret,
	}

	_, valid := token.ValidateAndExtract(tokenStr)

	return valid
}

func validateCredentials(user *User, password string) bool {
	return user.Password == password
}

func getUser(email string) *User {
	for _, u := range Users {
		if u.Email == email {
			return u
		}
	}

	return nil
}
