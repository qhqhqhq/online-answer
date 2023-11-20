package dao

import (
	"online-answer/db/model"
)

var UserImp UserInterface = &UserInterfaceImp{}

type UserInterfaceImp struct{}

func (imp *UserInterfaceImp) GetUser(openId string) (*model.User, error) {
	user := new(model.User)
	err := cli.Where(&model.User{OpenID: openId}).First(user).Error

	return user, err
}

func (imp *UserInterfaceImp) GetUsersByGroup(groupNumber uint) ([]*model.User, error) {
	var users []*model.User
	err := cli.Where(&model.User{GroupNumber: groupNumber}).Find(&users).Error

	return users, err
}

func (imp *UserInterfaceImp) CountUsersByGroup(groupNumber uint) (int, error) {
	var count int64

	err := cli.Where(&model.User{GroupNumber: groupNumber}).Count(&count).Error

	return int(count), err
}

func (imp *UserInterfaceImp) UpsertUser(user *model.User) error {
	return cli.Save(user).Error
}

func (imp *UserInterfaceImp) DisassociateUserForGroup(openId string) error {
	return cli.Model(&model.User{}).Where(&model.User{OpenID: openId}).Update("group_number", nil).Error
}

func (imp *UserInterfaceImp) DeleteUser(openId string) error {
	return cli.Where(&model.User{OpenID: openId}).Delete(&model.User{}).Error
}
