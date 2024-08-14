package main

import (
	"ginchat/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {
	dsn := "root:fuxm0228.@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	//   db, err := gorm.Open(mysql.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 迁移 schema
	//db.AutoMigrate(&models.UserBasic{})

	//db.AutoMigrate(&models.Message{})
	db.AutoMigrate(&models.Community{})
	// db.AutoMigrate(&models.Contact{})
	// Create
	// user := &models.UserBasic{}
	// user.Name = "SinmFull"
	// db.Create(user)

	// Read
	//fmt.Println(db.First(user, 1)) // 根据整型主键查找
	//   db.First(user, "code = ?", "D42") // 查找 code 字段值为 D42 的记录

	// Update - 将 product 的 price 更新为 200
	// db.Model(user).Update("PassWord", "1234")
	// Update - 更新多个字段
	//   db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // 仅更新非零值字段
	//   db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - 删除 product
	//   db.Delete(&product, 1)
}
