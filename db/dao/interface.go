package dao

import (
	"online-answer/db"
	"online-answer/db/model"
)

var cli = db.Get()

type UserInterface interface {
	GetUser(openId string) (*model.User, error)
	GetUsersByGroup(groupNumber uint) ([]*model.User, error)
	CountUsersByGroup(groupNumber uint) (int, error)
	UpsertUser(user *model.User) error
	DisassociateUserForGroup(openId string) error
	DeleteUser(openId string) error
}

type GroupInterface interface {
	GetGroup(number uint) (*model.Group, error)
	GetAllGroups() ([]*model.Group, error)
	UpsertGroup(group *model.Group) error
	DeleteGroup(number uint) error
}

type JudgementQuestionInterface interface {
	GetQuestion(number uint) (*model.JudgementQuestion, error)
	GetRandomQuestion() (*model.JudgementQuestion, error)
	UpsertQuestion(question *model.JudgementQuestion) error
}

type ChoiceQuestionInterface interface {
	GetQuestion(number uint) (*model.ChoiceQuestion, error)
	GetRandomQuestion() (*model.ChoiceQuestion, error)
	UpsertQuestion(question *model.ChoiceQuestion) error
}

type RecordInterface interface {
	GetRecord(id uint) (*model.Record, error)
	UpsertRecord(record *model.Record) error
}
