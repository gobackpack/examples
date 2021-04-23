package main

import (
	jwtLib "github.com/gobackpack/jwt"
	"github.com/sirupsen/logrus"
	"log"
	"time"
)

func main() {
	token := &jwtLib.Token{
		Secret: []byte("testkey"), // if we change this during validation, token will be invalid!
	}

	if err := token.Generate(&jwtLib.Claims{
		Expiration: time.Second * 30,
		Fields: map[string]interface{}{
			"username": "semir",
			"email":    "semir@mail.com",
			"id":       "semir-123",
		},
	}); err != nil {
		log.Fatalln("failed to generate jwt: ", err)
	}

	logrus.Info(token.Content)

	claims, valid := token.ValidateAndExtract(token.Content)
	if !valid {
		logrus.Error("invalid token: ", token.Content)
	}

	logrus.Info("claims: ", claims)
}
