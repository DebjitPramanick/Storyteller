package controllers

import (
	"context"
	"fmt"
	"server/database"
	"server/helpers"
	"server/middleware"
	"server/models"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(c *fiber.Ctx) error {
	
	var data map[string]string

	err := c.BodyParser(&data)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Something went wrong. Please try again later.",
		})
	}

	pass, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := models.User{
		ID: primitive.NewObjectID(),
		Name: data["name"],
		Bio: data["bio"],
		Email: data["email"],
		Password: pass,
		Username: data["username"],
		Avatar: data["avatar"],
		CreatedAt: time.Now(),
	}

	result, err := database.Users.InsertOne(context.TODO(), user)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error occurred when creating user. Please try again later.",
		})
	}

	fmt.Println("Created user with email & ID: ", user.Email, result.InsertedID)
	return c.JSON(user)
}

func LoginUser(c *fiber.Ctx) error {
	
	var data map[string]string
	err := c.BodyParser(&data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Something went wrong. Please try again later.",
		})
	}

	var user models.User
	var query bson.M

	if helpers.IsKeyPresent(data, "email")  {
		query = bson.M{"email": data["email"]}
	} else {
		query = bson.M{"username": data["username"]}
	}
	queryError := database.Users.FindOne(context.TODO(), query).Decode(&user)

	if queryError == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"])); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Password is not correct.",
		})
	}

	claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer: user.ID.Hex(),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	}) 

	token, tokenError := claim.SignedString([]byte(middleware.JWTSECRET))

	if tokenError != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not login.",
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
		"token": token,
	})
}

func GetUserByID(c *fiber.Ctx) error {
	userID, _ := primitive.ObjectIDFromHex(c.Params("userId"));
	var user models.User

	queryError := database.Users.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&user)

	if queryError == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	data := fiber.Map{
		"_id": user.ID,
		"name": user.Name,
		"bio": user.Bio,
		"avatar": user.Avatar,
	}

	return c.JSON(data)
}

func RemoveUserByID(c *fiber.Ctx) error {
	userID, _ := primitive.ObjectIDFromHex(c.Params("userId"));

	_, queryError := database.Users.DeleteOne(context.Background(), bson.M{"_id": userID})

	if queryError == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	res, _ := database.Stories.DeleteMany(context.Background(), bson.M{"author._id": userID})
	fmt.Println(res)

	return c.JSON(fiber.Map{
		"message": "Removed user successfully.",
	})
}

func UpdateUserByID(c *fiber.Ctx) error {
	userID, _ := primitive.ObjectIDFromHex(c.Params("userId"));

	var data map[string]string

	err := c.BodyParser(&data)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Something went wrong. Please try again later.",
		})
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{
		"name": data["name"],
		"bio": data["bio"],
		"email": data["email"],
		"avatar": data["avatar"],
	}}

	_, queryError := database.Users.UpdateOne(context.Background(), filter, update)

	if queryError == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Updated user successfully.",
	})
}

func GetUsersByUserName(c *fiber.Ctx) error {

	var users []models.User

	username := c.Params("username");

	filter := bson.D{{Key: "username", Value: primitive.Regex{Pattern: username, Options: ""}}}

	cur, _ := database.Users.Find(context.TODO(), filter)

	for cur.Next(context.TODO()) {
		var user models.User
		err := cur.Decode(&user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Something went wrong. Please try again later.",
			})
		}
		users = append(users, user)
	}

	if err := cur.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Something went wrong. Please try again later.",
		})
	}

	defer cur.Close(context.TODO())

	return c.JSON(users)
}