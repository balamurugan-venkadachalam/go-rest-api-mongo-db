package main

import (
	"context"
	"example.com/m/v2/config"
	"example.com/m/v2/handlers"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/labstack/gommon/random"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CorrelationID = "X-Correlation-ID"
)

var (
	c       *mongo.Client
	db      *mongo.Database
	pCol    *mongo.Collection
	userCol *mongo.Collection
	cfg     config.AppConfig
)

func init() {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Configuration cannot be read : %v", err)
	}
	connectionURI := fmt.Sprintf("mongodb://%s:%s", cfg.DBHost, cfg.DBPort)
	c, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connectionURI))
	if err != nil {
		log.Fatalf("Unable to connect o database : %v", err)
	}
	db = c.Database(cfg.DBName)
	pCol = db.Collection(cfg.CollectionName)
	userCol = db.Collection(cfg.UserCollection)
	isUserIndexUnique := true
	indexModel := mongo.IndexModel{
		Keys: bson.D{{"username", 1}},
		Options: &options.IndexOptions{
			Unique: &isUserIndexUnique,
		},
	}
	_, err = userCol.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatalf("Unable to create an index : %+v", err)
	}
	//pCol.InsertOne()
}

func addCorrelationID(next echo.HandlerFunc) echo.HandlerFunc {

	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()
		rid := req.Header.Get(CorrelationID)
		if rid == "" {
			rid = random.String(24)
		}
		req.Header.Set(CorrelationID, rid)
		res.Header().Set(CorrelationID, rid)
		return next(c)
	}
}

func main() {
	fmt.Println("test")
	e := echo.New()
	e.Logger.SetLevel(log.ERROR)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(addCorrelationID)
	h := &handlers.ProductHandler{Col: pCol}
	userHandler := &handlers.UserHandler{Col: userCol}
	e.POST("/products", h.CreateProducts, middleware.BodyLimit("1M"))
	e.DELETE("/products/:id", h.DeleteProduct)
	e.GET("/products/:id", h.GetProduct)
	e.GET("/products", h.GetProducts)
	e.PUT("/products/:id", h.PutProduct)
	e.POST("/user", userHandler.CreateUser, middleware.BodyLimit("1M"))

	server := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	fmt.Println(server)
	e.Logger.Fatal(e.Start(server))

}
