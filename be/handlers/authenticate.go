package handlers

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ubaidillahfaris/whatsapp.git/db"
	"github.com/ubaidillahfaris/whatsapp.git/utils"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type AuthenticateHandler struct {
	mongo *db.MongoService
}

func NewAuthenticateHandler() *AuthenticateHandler {
	return &AuthenticateHandler{
		mongo: nil,
	}
}

func (h *AuthenticateHandler) Register(mongo *db.MongoService, c *gin.Context) error {
	h.mongo = mongo

	username := c.PostForm("username")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	if password != confirmPassword {
		c.JSON(400, gin.H{"error": "Passwords do not match"})
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return err
	}
	_, err = h.mongo.InsertOne(context.Background(), "users", map[string]string{
		"username": username,
		"password": string(hashedPassword),
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to register user"})
		return err
	}
	c.JSON(200, gin.H{"message": "User registered successfully"})
	return nil

}

func (h *AuthenticateHandler) Authenticate(c *gin.Context) error {
	ctx := context.Background()
	username := c.PostForm("username")
	password := c.PostForm("password")

	users, _ := db.Mongo.FindAll(ctx, "users", bson.M{"username": username}, nil, nil)
	if len(users) == 0 {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return nil
	}

	storedHash := users[0]["password"].(string)
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))

	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return err
	}

	// Generate JWT token
	token, err := utils.GenerateToken(username)

	c.SetCookie(
		"jwt",
		token,
		3600*24,
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
	return nil
}

func (h *AuthenticateHandler) CheckAuth(c *gin.Context) {
	tokenStr, err := c.Cookie("jwt")
	if err != nil {
		c.JSON(401, gin.H{"loggedIn": false})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(401, gin.H{"loggedIn": false})
		return
	}

	c.JSON(200, gin.H{"loggedIn": true, "username": token.Claims.(jwt.MapClaims)["username"]})
}
