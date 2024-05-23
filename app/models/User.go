package models

import (
	"github.com/goal-web/database/table"
	"github.com/goal-web/supports/class"
)

var (
	UserClass = class.Make[User]()
)

func Users() *table.Table[User] {
	return table.Class(UserClass, "users").SetPrimaryKey("id")
}

type User struct {
	table.Model[User] `json:"-"`

	Id   int64  `json:"id"`
	Name string `json:"name"`
	Age  int64  `json:"age"`
}
