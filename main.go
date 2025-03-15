package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	// Google OAuth2 config
	googleOauthConfig = oauth2.Config{
		ClientID:     "", // Replace with your Google Client ID
		ClientSecret: "", // Replace with your Client Secret
		RedirectURL:  "http://localhost:8080/auth/callback", // Set to /auth/callback
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}

	oauthStateString = "random" // Protect against CSRF attacks
)

func main() {
	r := gin.Default()

	// Step 1: Route to start OAuth2 flow (this redirects user to Google OAuth)
	r.GET("/auth", func(c *gin.Context) {
		// Generate OAuth2 authorization URL
		url := googleOauthConfig.AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline)
		c.Redirect(http.StatusFound, url)
	})

	// Step 2: Callback route where Google redirects after successful authentication
	r.GET("/auth/callback", func(c *gin.Context) {
		// Get the authorization code from the URL
		code := c.DefaultQuery("code", "")
		state := c.DefaultQuery("state", "")

		if code == "" {
			log.Println("Authorization code not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not found"})
			return
		}

		if state != oauthStateString {
			log.Println("Invalid state parameter")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
			return
		}

		// Step 3: Exchange the code for an access token
		token, err := googleOauthConfig.Exchange(c, code)
		if err != nil {
			log.Printf("Token exchange error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token exchange failed"})
			return
		}

		// Step 4: Use the token to get user information from Google
		client := googleOauthConfig.Client(c, token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			log.Printf("Failed to get user info: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
			return
		}
		defer resp.Body.Close()

		// Step 5: Decode the user information
		userInfo := struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}{}

		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			log.Printf("Failed to decode user info: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
			return
		}

		// Step 6: Return the user info as a response
		c.JSON(http.StatusOK, gin.H{
			"ID":    userInfo.ID,
			"Email": userInfo.Email,
			"Name":  userInfo.Name,
		})
	})

	// Start the server
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
