package user

import (
	"net/http"
	"time"

	jwt "github.com/form3tech-oss/jwt-go"

	"github.com/gofiber/fiber/v2"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/siteslave/demo-fiber/database"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	// gorm.Model
	UserId uint `gorm:"primary_key" json: "user_id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	// gorm.Model
	UserId uint `gorm:"primary_key" json: "user_id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
}

type jwtCustomClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.StandardClaims
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func GetUsers(c *fiber.Ctx) error {
	db := database.DBConn
	users := []UserResponse{}
	db.Table("users").Select([]string{"user_id", "first_name", "last_name", "email"}).Find(&users)

	return c.JSON(users)
}

func GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	db := database.DBConn
	var user User
	if err := db.First(&user, id).Error; err != nil {
		return c.SendStatus(http.StatusNotFound)
	}
	
	return c.JSON(user)
}

func NewUser(c *fiber.Ctx) error {
	db := database.DBConn
	// fmt.Print(c.FormValue("username"))
	
	firstName := c.FormValue("firstName")
	lastName := c.FormValue("lastName")
	email := c.FormValue("email")
	username := c.FormValue("username")

	_password := c.FormValue("password")

	hash, _ := HashPassword(_password)

	password := hash

	user := User{
		FirstName: firstName,
		LastName: lastName,
		Email: email,
		Username: username,
		Password: password,
	}
	if err := db.Create(&user).Error; err != nil {
		return c.Status(503).SendString(err.Error())
	}
	return c.JSON(user)
}

func UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")

	db := database.DBConn
	user := User{}

	if  err := db.First(&user, id).Error; err != nil {
		return c.SendStatus(http.StatusNotFound)
	}

	firstName := c.FormValue("firstName")
	lastName := c.FormValue("lastName")
	email := c.FormValue("email")
	
	if err := db.Model(&User{}).Updates(User{
		FirstName: firstName,
		LastName: lastName,
		Email: email,
	}).Error; err != nil {
		return c.SendStatus(http.StatusInternalServerError)
	}

	return c.SendStatus(http.StatusOK)
}

func DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	db := database.DBConn

	var user User
	db.First(&user, id)
	if user.FirstName == "" {
		return c.Status(500).SendString("No User Found with ID")
	}
	db.Delete(&user)
	return c.SendString("User Successfully deleted")
}

func Login(c *fiber.Ctx) error {
	db := database.DBConn
	user := User{}

	username := c.FormValue("username")
	password := c.FormValue("password")

	result := db.Where(&User{Username: username}).Find(&user)

	if result.Error != nil {
		return c.SendStatus(http.StatusInternalServerError)
	}

	if result.RowsAffected == 0 {
		return c.SendStatus(http.StatusNotFound)
	}

	hash := user.Password
	match := CheckPasswordHash(password, hash)

	if !match {
		return c.SendStatus(http.StatusUnauthorized)
	}
	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = "Satit Rianpit"
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"token": t,
	})
}
