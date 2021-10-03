package main

import (
	"context"
	"example.com/m/v2/config"
	"example.com/m/v2/handlers"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/random"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

const (
	CorrelationID = "X-Correlation-ID"
)

var (
	c   *mongo.Client
	db  *mongo.Database
	col *mongo.Collection
	cfg config.AppConfig
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
	col = db.Collection(cfg.CollectionName)
	//col.InsertOne()
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
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(addCorrelationID)
	h := handlers.ProductHandler{Col: col}
	e.POST("/products", h.CreateProducts, middleware.BodyLimit("1M"))
	e.DELETE("/products/:id", h.DeleteProduct)
	e.GET("/products/:id", h.GetProduct)
	e.GET("/products", h.GetProducts)
	e.PUT("/products/:id", h.PutProduct)
	server := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	fmt.Println(server)
	e.Logger.Fatal(e.Start(server))

}
