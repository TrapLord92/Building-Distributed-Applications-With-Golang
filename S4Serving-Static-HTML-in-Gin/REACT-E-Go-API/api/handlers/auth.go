package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"react-go-api/api/models"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AuthHandler holds the MongoDB collection and context for handling requests.
type AuthHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

// Claims struct to include JWT claims for username and standard claims.
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// JWTOutput represents the token and its expiration time.
type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

// NewAuthHandler creates a new AuthHandler with MongoDB collection and context.
func NewAuthHandler(ctx context.Context, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		collection: collection,
		ctx:        ctx,
	}
}

// SignInHandler handles the login process and issues JWT tokens.
func (handler *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User

	// Bind the incoming JSON payload to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the user's password using SHA-256
	h := sha256.New()
	h.Write([]byte(user.Password))
	hashedPassword := hex.EncodeToString(h.Sum(nil)) // Convert hash to hex string

	// Look for a matching user in the MongoDB collection
	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"username": user.Username,
		"password": hashedPassword, // Compare against the hashed password in hex
	})

	if cur.Err() != nil {
		// User not found or invalid credentials
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// If valid, create a JWT token for the user
	expirationTime := time.Now().Add(10 * time.Minute) // Token expires in 10 minutes
	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Create the token using HS256 signing method and the JWT secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		// Error generating the token
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the token and its expiration time as JSON response
	jwtOutput := JWTOutput{
		Token:   tokenString,
		Expires: expirationTime,
	}
	c.JSON(http.StatusOK, jwtOutput)
}

// RefreshHandler handles the token refresh process.
func (handler *AuthHandler) RefreshHandler(c *gin.Context) {
	tokenValue := c.GetHeader("Authorization")

	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !tkn.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Check if token is close to expiration
	timeLeft := time.Until(time.Unix(claims.ExpiresAt, 0))
	if timeLeft > 30*time.Second {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is not expired yet"})
		return
	}

	// Create a new expiration time and refresh the token
	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newTokenString, err := newToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the refreshed token
	jwtOutput := JWTOutput{
		Token:   newTokenString,
		Expires: expirationTime,
	}
	c.JSON(http.StatusOK, jwtOutput)
}

// AuthMiddleware is a middleware to protect routes by checking JWT tokens.
func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenValue := c.GetHeader("Authorization")
		claims := &Claims{}

		tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !tkn.Valid {
			// Token is invalid or parsing failed
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, invalid token"})
			c.Abort()
			return
		}

		// Allow the request to proceed
		c.Next()
	}
}
