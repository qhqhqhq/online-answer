package dao

import "online-answer/db/model"

var GroupImp GroupInterface = &GroupInterfaceImp{}

type GroupInterfaceImp struct{}

func (imp *GroupInterfaceImp) GetGroup(number uint) (*model.Group, error) {
	var group = new(model.Group)
	err := cli.Where(&model.Group{Number: number}).First(&group).Error
	return group, err
}

func (imp *GroupInterfaceImp) GetAllGroups() ([]*model.Group, error) {
	var groups []*model.Group
	err := cli.Find(&groups).Error
	return groups, err
}

func (imp *GroupInterfaceImp) UpsertGroup(group *model.Group) error {
	return cli.Save(group).Error
}

func (imp *GroupInterfaceImp) DeleteGroup(number uint) error {
	return cli.Where(&model.Group{Number: number}).Delete(&model.Group{}).Error
}
