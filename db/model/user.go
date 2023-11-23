package model

type User struct {
	OpenID      string `gorm:"primarykey"`
	Name        string
	GroupNumber uint
}
