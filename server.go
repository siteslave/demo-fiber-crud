package main

import (
	"fmt"
	"log"

	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/jinzhu/gorm"
	"github.com/siteslave/demo-fiber/database"
	"github.com/siteslave/demo-fiber/user"
)

func initDatabase() {
	var err error
	database.DBConn, err = gorm.Open("mysql", "root:789124@tcp(127.0.0.1:3306)/test?charset=utf8&parseTime=True");
	if err != nil {
		panic("failed to connect database")
	}
	fmt.Println("Connection Opened to Database")
	// database.DBConn.AutoMigrate(&user.User{})
	// fmt.Println("Database Migrated")
}

func main() {
  app := fiber.New()
	app.Use(cors.New())

	initDatabase()
	defer database.DBConn.Close()

  app.Get("/", func(c *fiber.Ctx) error {
    return c.SendString("Hello, World!")
  })

	api := app.Group("api")

	api.Post("/v1/login", user.Login)
	
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte("secret"),
	}))
	
	app.Get("/jwt", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		name := claims["name"].(string)
		return c.SendString("Welcome " + name)
	})

	api.Get("/v1/users", user.GetUsers)
	api.Get("/v1/users/:id", user.GetUser)
	api.Put("/v1/users/:id", user.UpdateUser)
	api.Post("/v1/users", user.NewUser)
	api.Delete("/v1/users/:id", user.DeleteUser)

  log.Fatal(app.Listen(":3000"))
}