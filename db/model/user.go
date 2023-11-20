package model

type User struct {
	OpenID      string `gorm:"primarykey"`
	GroupNumber uint
}
