package model

import "gorm.io/gorm"

type ChoiceQuestion struct {
	gorm.Model
	Number           uint `gorm:"unique"`
	Content          string
	Answer           string
	IsMultipleChoice bool
	Options          []QuestionOption `gorm:"foreignKey:QuestionID"`
}

type QuestionOption struct {
	gorm.Model
	QuestionID uint
	Key        string
	Value      string
}
