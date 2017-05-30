package models

type User struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"-"`
	Password []byte `json:"-"`
	Nimg     string `json:"-"`
}

// func (u *User) Fill(missing string) *User {
// 	if u.Id == 0 {
// 		//
// 	}
// 	err := db.Query(`select `+missing+` from users where id = $1`, User.Id).Rows(u)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return u
// }

func (u *User) NameByIdentity(identity string) string {
	if identity != "real" {
		err = db.Query(`select nname as name, nimg from users where id = $1`, u.Id).Rows(u)
		if err != nil {
			panic(err)
		}
	} else {
		err = db.Query(`select name from users where id = $1`, u.Id).Rows(u)
		if err != nil {
			panic(err)
		}
	}
	return u.Name
}
