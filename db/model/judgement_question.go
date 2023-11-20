package model

import "gorm.io/gorm"

type JudgementQuestion struct {
	gorm.Model
	Number  uint `gorm:"unique"`
	Content string
	Answer  bool
}
