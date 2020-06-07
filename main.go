package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Credentials json configuration file rappresentation
type Credentials struct {
	Cid     string `json:"cid"`
	Csecret string `json:"csecret"`
}

// OAuthUser is the OAuth user rappresentation
type OAuthUser struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
}

// User in database rappresentation
type User struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
	Firstname string     `json:"firstname"`
	Lastname  string     `json:"lastname"`
	Email     string     `json:"email"`
}

var cred Credentials
var conf *oauth2.Config
var state string
var db *gorm.DB

func initDb() *gorm.DB {
	dbi, err := gorm.Open("postgres", "host=goauthdb port=5432 user=admin dbname=goauthdb password=123  sslmode=disable")

	if err != nil {
		panic(err.Error())
	}

	db = dbi
	migration(db)
	return db
}

func migration(db *gorm.DB) {
	db.AutoMigrate(&User{})
}

func init() {
	initDb()
	file, err := ioutil.ReadFile("./creds.json")
	if err != nil {
		log.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &cred)

	var ru string = os.Getenv("URL") + "v1/oauth"

	conf = &oauth2.Config{
		ClientID:     cred.Cid,
		ClientSecret: cred.Csecret,
		RedirectURL:  ru,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email", // You have to select your own scope from here -> https://developers.google.com/identity/protocols/googlescopes#google_sign-in
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{})
}

func getLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

func authHandler(c *gin.Context) {
	tok, err := conf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client := conf.Client(oauth2.NoContext, tok)
	userinfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer userinfo.Body.Close()
	data, _ := ioutil.ReadAll(userinfo.Body)
	log.Println("Email body: ", string(data))
	var ou OAuthUser
	if err := json.Unmarshal(data, &ou); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "parsing OauthUser", "e": err.Error()})
		return
	}

	var u User
	db.Where("email = ?", ou.Email).First(&u)

	if u.ID > 0 {
		c.JSON(http.StatusOK, gin.H{"data": u, "exists": "yes"})
		return
	}

	u.Email = ou.Email
	u.Firstname = ou.GivenName
	u.Lastname = ou.FamilyName
	db.Save(&u)

	c.JSON(http.StatusOK, gin.H{"data": u, "exists": "no"})
}

func loginHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"url": getLoginURL(state)})
}

func main() {
	router := gin.Default()

	router.GET("/todoAPI", indexHandler)
	router.GET("/todoAPI/login", loginHandler)
	router.GET("/todoAPI/v1/oauth", authHandler)

	router.Run(":8080")
}
