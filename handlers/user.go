package handlers

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type User struct {
	Email    string `json:"username" bson:"username" validate:"required,max=50"`
	Password string `json:"password" bson:"password" validate:"required,min=8,max=300"`
}

// UserHandler user handler
type UserHandler struct {
	Col *mongo.Collection
}

// UserValidator a product custom validator
type UserValidator struct {
	validator *validator.Validate
}

// Validate validate product
func (v *UserValidator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

func (p *UserHandler) CreateUser(c echo.Context) error {
	var user User
	c.Echo().Validator = &UserValidator{validator: v}
	if err := c.Bind(&user); err != nil {
		log.Errorf("Unable to bind : %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Unable to parse the request payload")
	}
	err := c.Validate(user)
	if err != nil {
		log.Errorf("Unable to validate user %+v error -> %v", user, err)
		return echo.NewHTTPError(http.StatusBadRequest, "Unable to validate the request payload")
	}
	IDs, err := insertUser(context.Background(), user, p.Col)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, IDs)

}

func insertUser(ctx context.Context, user User, col *mongo.Collection) (interface{}, error) {
	var newUser User
	res := col.FindOne(ctx, bson.M{"username": user.Email})
	err := res.Decode(&newUser)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Errorf("unable to decode user %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "unable to decode user")
	}
	if newUser.Email != "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "user already exits")
	}
	insertID, err := col.InsertOne(ctx, user)
	if err != nil {
		log.Errorf("Unable to insert %v", err)
		return nil, err
	}
	return insertID.InsertedID, nil
}
