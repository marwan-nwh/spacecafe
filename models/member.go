package models

// import (
// 	"spacecafe/db"
// )

// type Member struct {
// 	Table  string
// 	Role   string
// 	UserId int
// }

// func CanPost(userid, table string) bool {
// 	var role string
// 	err = data.Query(`select role from memberships where table_name = $1 and user_id = $2`, r.FormValue("table"), userid).Rows(&role)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if role != "admin" && role != "writer" && role != "manager" {
// 		return false
// 	}
// 	return true
// }

// func (m *Member) fill(missing string) *Member {
// 	if m.Table == "" || m.UserId == 0 {
// 		//
// 	}
// 	err := db.Query(`select `+missing+` from memberships where user_id = $1 and table_name = $2`, m.UserId, m.Table).Rows(m)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return m
// }
