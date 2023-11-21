package db

import (
	"fmt"
	"log"
	"online-answer/db/model"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var dbInstance *gorm.DB

func Init() error {
	source := "%s:%s@tcp(%s)/%s?readTimeout=1500ms&writeTimeout=1500ms&charset=utf8&loc=Local&&parseTime=true"
	user := os.Getenv("MYSQL_USERNAME")
	pwd := os.Getenv("MYSQL_PASSWORD")
	addr := os.Getenv("MYSQL_ADDRESS")
	dataBase := os.Getenv("MYSQL_DATABASE")
	if dataBase == "" {
		dataBase = "online_answer"
	}
	source = fmt.Sprintf(source, user, pwd, addr, dataBase)
	log.Println("start init mysql")

	db, err := gorm.Open(mysql.Open(source), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Println("DB open error, err=", err.Error())
		return err
	}
	db.AutoMigrate(&model.Group{}, &model.User{}, &model.JudgementQuestion{}, &model.ChoiceQuestion{}, &model.QuestionOption{}, &model.Record{})

	dbInstance = db
	log.Println("finish init mysql")
	return nil
}

func Get() *gorm.DB {
	return dbInstance
}
