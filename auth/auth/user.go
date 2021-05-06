package auth

import "sort"

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
