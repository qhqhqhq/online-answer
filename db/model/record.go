package model

import (
	"gorm.io/gorm"
)

type Record struct {
	gorm.Model
	Result     string
	Candidates string
}
