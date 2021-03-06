package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"
)

var db gorm.DB

// ユーザモデル
type User struct {
	ID       int `gorm:"primary_key"`
	UserName string
}
type Users []User

// 残高モデル
type Remainder struct {
	ID        int `gorm:"primary_key"`
	UserID    int
	Remainder int
	UpdatedAt time.Time
}
type Remainders []Remainder

// 商品モデル
type Commodity struct {
	ID        int
	Name      string
	JANCode   string `gorm:"unique"`
	Price     int
	Stock     int
	CreatedAt time.Time
	UpdatedAt time.Time
}
type Commodities []Commodity

// HTTPレスポンス用
type basicResponseJSON struct {
	Code    int
	Message string
}

func main() {
	db, err := gorm.Open("sqlite3", "database/database.sqlite3")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	// For Dev
	db.DropTableIfExists(&User{})
	db.CreateTable(&User{})
	db.AutoMigrate(&User{})
	db.DropTableIfExists(&Remainder{})
	db.CreateTable(&Remainder{})
	db.AutoMigrate(&Remainder{})
	db.DropTableIfExists(&Commodity{})
	db.CreateTable(&Commodity{})
	db.AutoMigrate(&Commodity{})

	user := new(User)
	user.UserName = "admin"
	db.NewRecord(user)
	db.Create(&user)

	e := echo.New()

	/*
	 * Static files
	 */
	e.Static("/static", "assets")
	e.File("/", "public/index.html")

	/*
	 * ユーザ
	 */
	users := e.Group("/user")
	// ユーザ一覧
	users.GET("/", func(c echo.Context) error {
		mulusers := new(Users)
		db.Find(&mulusers)
		return c.JSON(http.StatusOK, mulusers)
	})

	// useridのユーザ情報の取得
	users.GET("/:userid", func(c echo.Context) error {
		userid, _ := strconv.Atoi(c.Param("userid"))
		user := new(User)
		db.Where("id = ?", userid).First(&user)
		if user.ID == 0 {
			return c.JSON(http.StatusOK, basicResponseJSON{Code: 301, Message: "Not Found."})
		}
		return c.JSON(http.StatusOK, user)
	})

	// ユーザ追加
	users.POST("/add", func(c echo.Context) error {
		user := new(User)
		if err := c.Bind(user); err != nil {
			return c.JSON(http.StatusOK, basicResponseJSON{Code: 300, Message: "Faild Binding posted data"})
		}
		db.NewRecord(user)
		db.Create(&user)
		rem := new(Remainder)
		rem.UserID = user.ID
		db.NewRecord(rem)
		db.Create(&rem)
		return c.JSON(http.StatusOK, basicResponseJSON{Code: 200, Message: "OK"})
	})

	// ユーザ情報更新
	users.POST("/update", func(c echo.Context) error {
		newUserData := new(User)
		if err := c.Bind(newUserData); err != nil {
			return c.JSON(http.StatusOK, basicResponseJSON{Code: 300, Message: "Faild Binding posted data"})
		}
		db.Model(&user).Where("id = ?", newUserData.ID).Update(newUserData)
		return c.JSON(http.StatusOK, basicResponseJSON{Code: 200, Message: "OK"})
	})

	// ユーザ削除
	users.POST("/delete", func(c echo.Context) error {
		userid, _ := strconv.Atoi(c.FormValue("ID"))
		db.Where("id = ?", userid).Delete(&User{})
		db.Where("user_id = ?", userid).Delete(&Remainder{})
		return c.JSON(http.StatusOK, basicResponseJSON{Code: 200, Message: "OK"})
	})

	/*
	 * 残高
	 */
	remainder := e.Group("/remainder")
	// useridの残高参照
	remainder.GET("/:userid", func(c echo.Context) error {
		userid, _ := strconv.Atoi(c.Param("userid"))
		rem := new(Remainder)
		db.Where("user_id = ?", userid).Last(&rem)
		return c.JSON(http.StatusOK, rem)
	})

	// ユーザの残高の履歴参照
	remainder.GET("/:userid/all", func(c echo.Context) error {
		userid, _ := strconv.Atoi(c.Param("userid"))
		rems := new(Remainders)
		db.Where("user_id = ?", userid).Find(&rems)
		return c.JSON(http.StatusOK, rems)
	})

	// 残高チャージ
	remainder.POST("/charge", func(c echo.Context) error {
		rem := new(Remainder)
		if err = c.Bind(rem); err != nil {
			return c.JSON(http.StatusOK, basicResponseJSON{Code: 300, Message: "Can't bind data"})
		}
		lastRem := new(Remainder)
		db.Where("user_id = ?", rem.UserID).Last(&lastRem)
		newRem := new(Remainder)
		newRem.UserID = rem.UserID
		newRem.Remainder = lastRem.Remainder + rem.Remainder
		db.NewRecord(newRem)
		db.Create(&newRem)
		return c.JSON(http.StatusOK, basicResponseJSON{Code: 200, Message: "OK"})
	})

	// 残高引き出し
	remainder.POST("/:userid/withdraw", func(c echo.Context) error {
		userid, _ := strconv.Atoi(c.Param("userid"))
		price, _ := strconv.Atoi(c.FormValue("price"))
		lastRem := new(Remainder)
		db.Where("user_id = ?", userid).Last(&lastRem)
		rem := new(Remainder)
		rem.UserID = userid
		rem.Remainder = lastRem.Remainder - price
		if rem.Remainder < 0 {
			return c.JSON(http.StatusOK, basicResponseJSON{Code: 401, Message: "Invalid value"})
		}
		db.NewRecord(rem)
		db.Create(&rem)
		return c.JSON(http.StatusOK, basicResponseJSON{Code: 200, Message: "OK"})
	})

	/*
	 * 商品
	 */
	commodity := e.Group("/commodity")
	// 全商品取得
	commodity.GET("/", func(c echo.Context) error {
		coms := new(Commodities)
		db.Find(&coms)
		return c.JSON(http.StatusOK, coms)
	})

	// 新商品追加
	commodity.POST("/add", func(c echo.Context) error {
		com := new(Commodity)
		if err := c.Bind(com); err != nil {
			return c.JSON(http.StatusOK, basicResponseJSON{Code: 300, Message: "Faild Binding posted data"})
		}
		addedCom := new(Commodity)
		db.Where("jan_code = ?", com.JANCode).First(&addedCom)
		if addedCom.ID != 0 {
			db.Model(&addedCom).Where("jan_code = ?", com.JANCode).Update("stock", addedCom.Stock+com.Stock)
			return c.JSON(http.StatusOK, basicResponseJSON{Code: 200, Message: "OK"})
		}
		db.NewRecord(com)
		if err := db.Create(&com).Error; err != nil {
			return c.JSON(http.StatusOK, basicResponseJSON{Code: 500, Message: "Database Error"})
		}
		return c.JSON(http.StatusOK, basicResponseJSON{Code: 200, Message: "OK"})
	})

	// 商品削除
	commodity.POST("/delte", func(c echo.Context) error {
		comid, _ := strconv.Atoi(c.FormValue("ID"))
		db.Where("id = ?", comid).Delete(&Commodity{})
		return c.JSON(http.StatusOK, basicResponseJSON{Code: 200, Message: "OK"})
	})

	// 商品購入
	commodity.POST("/buy", func(c echo.Context) error {
		// Fetch commodity data
		com := new(Commodity)
		comid, _ := strconv.Atoi(c.FormValue("comid"))
		db.Where("id = ?", comid).First(&com)

		// Fetxch user and remainder data
		userid, _ := strconv.Atoi(c.FormValue("userid"))
		lastUserRem := new(Remainder)
		db.Where("user_id = ?", userid).Last(lastUserRem)

		// Update user remainder
		newUserRem := new(Remainder)
		newUserRem.UserID = userid
		newUserRem.Remainder = lastUserRem.Remainder - com.Price
		db.NewRecord(newUserRem)
		db.Create(&newUserRem)

		// Update commodity stock
		db.Model(&com).Update("Stock", com.Stock-1)

		return c.JSON(http.StatusOK, basicResponseJSON{Code: 200, Message: "OK"})
	})

	e.Logger.Fatal(e.Start(":1323"))
}
