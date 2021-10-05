package handlers

import (
	"context"
	"example.com/m/v2/dbiface"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

var (
	v = validator.New()
)

type Response struct {
	input interface{}
	error string `json:"error"`
}

type Product struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"id,omitempty"`
	Name        string             `json:"product_name" bson:"product_name" validate:"required,max=10"`
	Price       int                `json:"price" bson:"price" validate:"required,max=2000"`
	Currency    string             `json:"currency" bson:"currency"  validate:"required,max=3"`
	Discount    int                `json:"discount" bson:"discount"`
	Vendor      string             `json:"vendor" bson:"vendor" validate:"required"`
	Accessories []string           `json:"accessories,omitempty" bson:"accessories,omitempty"`
	IsEssential bool               `json:"is_essential" bson:"is_essential"`
}

type ProductHandler struct {
	Col *mongo.Collection
}

// ProductValidator a product custom validator
type ProductValidator struct {
	validator *validator.Validate
}

// Validate validate product
func (v *ProductValidator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

func (p *ProductHandler) CreateProducts(c echo.Context) error {
	var products []Product
	c.Echo().Validator = &ProductValidator{validator: v}
	if err := c.Bind(&products); err != nil {
		log.Errorf("Unable to bind : %v", err)
		return err
	}
	for _, product := range products {
		err := c.Validate(product)
		if err != nil {
			log.Errorf("Unable to validate product %+v error -> %v", product, err)
			return err
		}
	}
	IDs, err := insertProducts(context.Background(), products, p.Col)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, IDs)

}

func (p *ProductHandler) DeleteProducts(c echo.Context) error {
	products, err := findProducts(context.Background(), p.Col)
	if err != nil {
		log.Errorf("Unable to delete product %v", err)
		return err
	}
	return c.JSON(http.StatusOK, products)
}

func (p *ProductHandler) DeleteProduct(c echo.Context) error {
	_, err := deleteProduct(context.Background(), c.Param("id"), p.Col)
	if err != nil {
		log.Errorf("Unable to delete product %v", err)
		return err
	}
	return c.JSON(http.StatusOK, nil)
}

func (p *ProductHandler) GetProduct(c echo.Context) error {
	product, err := findProduct(context.Background(), c.Param("id"), p.Col)
	if err != nil {
		log.Errorf("Unable to find products %v", err)
		return err
	}
	return c.JSON(http.StatusOK, product)
}

func (p *ProductHandler) GetProducts(c echo.Context) error {
	product, err := findProducts(context.Background(), p.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, product)
}

func (p *ProductHandler) PutProduct(c echo.Context) error {
	var product Product
	c.Echo().Validator = &ProductValidator{validator: v}
	if err := c.Bind(&product); err != nil {
		log.Errorf("Unable to bind : %v", err)
		return err
	}
	err := c.Validate(product)
	if err != nil {
		log.Errorf("Unable to validate product %+v error -> %v", product, err)
		return err
	}

	updatedProduct, err := updateProduct(context.Background(), c.Param("id"), product, p.Col)
	if err != nil {
		log.Errorf("Unable to update : %v", err)
		return err
	}
	return c.JSON(http.StatusOK, updatedProduct)
}

func deleteProduct(ctx context.Context, param string, col *mongo.Collection) (bool, error) {
	docID, err := primitive.ObjectIDFromHex(param)
	if err != nil {
		return false, err
	}
	_, err = col.DeleteOne(ctx, bson.M{"_id": docID})
	if err != nil {
		return false, err
	}
	return true, err
}

func findProducts(ctx context.Context, col dbiface.CollectionAPI) ([]Product, error) {
	var products []Product
	findCursor, err := col.Find(ctx, bson.M{})
	err = findCursor.All(ctx, &products)
	if err != nil {
		fmt.Println(err)
	}
	return products, err

}

func updateProduct(ctx context.Context, param string, p Product, collection dbiface.CollectionAPI) (Product, error) {
	var product Product
	docID, err := primitive.ObjectIDFromHex(param)
	if err != nil {
		return product, err
	}
	res := collection.FindOne(ctx, bson.M{"_id": docID})
	err = res.Decode(&product)
	if err != nil {
		return product, err
	}
	product.Price = p.Price
	product.Discount = p.Discount
	product.Name = p.Name
	product.Accessories = p.Accessories
	product.Vendor = p.Vendor
	product.Currency = p.Currency
	product.IsEssential = p.IsEssential
	update := bson.M{
		"$set": product,
	}
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": docID}, update)
	if err != nil {
		return product, err
	}

	return product, nil
}

func findProduct(ctx context.Context, param string, collection dbiface.CollectionAPI) (Product, error) {
	var product Product
	docID, err := primitive.ObjectIDFromHex(param)
	if err != nil {
		return product, err
	}
	res := collection.FindOne(ctx, bson.M{"_id": docID})
	err = res.Decode(&product)
	if err != nil {
		return product, err
	}
	return product, nil
}

func insertProducts(ctx context.Context, products []Product, collection dbiface.CollectionAPI) ([]interface{}, error) {
	var insertedIds []interface{}
	for _, product := range products {
		product.ID = primitive.NewObjectID()
		insertID, err := collection.InsertOne(ctx, product)
		if err != nil {
			log.Errorf("Unable to insert %v", err)
			return nil, err
		}
		insertedIds = append(insertedIds, insertID.InsertedID)
	}

	return insertedIds, nil
}
