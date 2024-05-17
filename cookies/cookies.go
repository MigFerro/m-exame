package cookies

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/MigFerro/exame/data"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/auth0"
	// "github.com/markbates/goth/providers/google"
)

func InitCookieStore() *sessions.CookieStore {
	fmt.Println("Initializing cookie store")

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	//COOKIE STORE
	key := os.Getenv("COOKIE_SESSION_KEY")

	gob.Register(&data.LoggedUser{})

	maxAge := 86400 * 30 // 30 days
	isProd := false      // Set to true when serving over https

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = isProd

	gothic.Store = store

	//PROVIDERS
	goth.UseProviders(
		// google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_SECRET_KEY"), "http://localhost:3000/auth/google/callback", "email", "profile"),
		auth0.New(os.Getenv("AUTH0_CLIENT_ID"), os.Getenv("AUTH0_SECRET"), "http://localhost:3000/auth/auth0/callback", os.Getenv("AUTH0_DOMAIN")),
	)

	return store
}
