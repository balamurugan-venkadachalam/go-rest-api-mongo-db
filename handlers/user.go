package handlers

import (
	"context"
	"example.com/m/v2/config"
	"github.com/golang-jwt/jwt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"time"
)

var (
	cfg config.AppConfig
)

type User struct {
	Email    string `json:"username" bson:"username" validate:"required,max=50"`
	Password string `json:"password,omitempty" bson:"password" validate:"required,min=8,max=300"`
	isAdmin  bool   `json:"is_admin,omitempty" bson:"is_admin"`
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

func (h UserHandler) AuthUser(c echo.Context) error {
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
	_, err = authenticateUser(context.Background(), user, h.Col)
	if err != nil {
		log.Errorf("Unable to authenticate user %+v error -> %v", user, err)
		return echo.NewHTTPError(http.StatusBadRequest, "Unable to authenticate user")
	}
	token, er := createToken(user.Email)
	if er != nil {
		log.Errorf("Unable to generate token error -> %v", user, err)
		return echo.NewHTTPError(http.StatusBadRequest, "Unable to generate token")
	}
	c.Response().Header().Set("x-auth-token", "Bearer "+token)
	return c.JSON(http.StatusOK, User{Email: user.Email})
}

func createToken(email string) (string, error) {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Configuration can not be read : %v", err)
	}
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = email
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := at.SignedString([]byte(cfg.JwtTokenSecret))
	if err != nil {
		log.Errorf("Unable to generate token error -> %v", err)
		return "", err
	}
	return token, nil
}

func authenticateUser(ctx context.Context, user User, col *mongo.Collection) (interface{}, error) {
	var existingUser User
	res := col.FindOne(ctx, bson.M{"username": user.Email})
	err := res.Decode(&existingUser)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Errorf("unable to decode user %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "unable to decode user")
	}
	if !compareUserPassword(err, existingUser, user) {
		log.Errorf("unable to decode user %v", err)
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid user credentials")
	}
	return existingUser, nil
}

func compareUserPassword(err error, existingUser User, user User) bool {
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password))

	if err != nil {
		log.Errorf("unable to decode user %v", err)
		return false
	}
	return true
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		log.Errorf("Unable to encrypt the password %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Unable to encrypt the password")
	}
	user.Password = string(hashedPassword)
	insertID, err := col.InsertOne(ctx, user)
	if err != nil {
		log.Errorf("Unable to insert %v", err)
		return nil, err
	}
	return insertID.InsertedID, nil
}
