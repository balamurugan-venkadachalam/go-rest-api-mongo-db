package handlers

import (
	"context"
	"example.com/m/v2/config"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	c   *mongo.Client
	db  *mongo.Database
	col *mongo.Collection
	cfg config.AppConfig
	h   ProductHandler
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

func TestProduct(t *testing.T) {
	body := `
			[
  				{
					"product_name":"Pencil",
					"price": 10,
					"currency":"nzd",
					"discount": 2,
					"vendor":"test",
					"accessories": ["test"],
					"is_essential":false
  				}
			]		
	`
	t.Run("test create product", func(t *testing.T) {
		req := httptest.NewRequest("post", "/products", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, res)
		h.Col = col
		assert.Nil(t, h.CreateProducts(c))
		assert.Equal(t, http.StatusCreated, res.Code)
	})

	t.Run("get product", func(t *testing.T) {
		req := httptest.NewRequest("get", "/products", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		res := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, res)
		c.SetParamNames("id")
		c.SetParamValues("615915766aab9e4f3bb138eb")
		h.Col = col
		err := h.GetProduct(c)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code)
	})
}
