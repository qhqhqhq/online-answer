package model

type Group struct {
	Number    uint `gorm:"primarykey"`
	Secret    string
	Members   []User `gorm:"foreignKey:GroupNumber;references:Number"`
	Eliminate bool
	Promotion bool
	College   string
}
