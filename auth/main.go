package main

import (
	"encoding/json"
	"github.com/gobackpack/examples/auth/auth"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := auth.RegisterUser(&auth.User{
		Email:    "semir_3@mail.com",
		Password: "test1234",
	}); err != nil {
		logrus.Fatal("registration failed: ", err)
	}

	b, _ := json.Marshal(auth.Users)

	logrus.Info("registered users: ", string(b))

	token, err := auth.Authenticate("semir_3@mail.com", "test1234")
	if err != nil {
		logrus.Fatal("authentication failed: ", err)
	}

	isAuthenticated := auth.IsAuthenticated(token)

	logrus.Info("user authenticated: ", isAuthenticated)
}
