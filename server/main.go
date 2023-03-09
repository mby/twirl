package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"twirl-server/shared"

	_ "twirl-server/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

//	@title			twirl
//	@version		3.0
//	@description	twirl
//	@host			localhost:1873
//	@BasePath		/api/v1

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							Header
//	@name						Authorization
//	@description				Use the token from /auth/login

const PORT = ":1873"

type (
	App struct {
		secretKey []byte
		users     *mongo.Collection
	}

	User struct {
		Username shared.Username `json:"username" bson:"username"`
		Password []byte          `json:"password" bson:"password"`
	}

	Claims struct {
		Username shared.Username `json:"username"`

		jwt.RegisteredClaims
	}

	upwdRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
)

func main() {
	// load .env
	godotenv.Load()

	// env vars
	mongoURI := getenv("MONGO_URI")
	secretKey := getenv("SECRET_KEY")
	founderUsername := getenv("FOUNDER_USERNAME")
	founderPassword := getenv("FOUNDER_PASSWORD")

	// connect to mongodb
	client, e := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if e != nil {
		panic(e)
	}
	defer func() {
		if e = client.Disconnect(context.TODO()); e != nil {
			panic(e)
		}
	}()

	if e := client.Ping(context.TODO(), readpref.Primary()); e != nil {
		panic(e)
	}
	log.Printf("Connected to MongoDB")

	// create app
	app := App{
		secretKey: []byte(secretKey),
		users:     client.Database("twirl").Collection("users"),
	}

	// deal with founder issue (solving the chicken-egg problem)
	go func() {
		for {
			// we delete the found first
			_, e := app.users.DeleteOne(context.TODO(), bson.M{"username": founderUsername})
			if e != nil {
				panic(e)
			}

			// if there are users, we don't need a founder
			count, e := app.users.CountDocuments(context.TODO(), bson.M{})
			if e != nil {
				panic(e) // wtf?
			}
			if count > 0 {
				return
			}

			// if there are no users after deleting him, we need a founder
			hashedPassword, e := bcrypt.GenerateFromPassword([]byte(founderPassword), 10)
			if e != nil {
				panic(e)
			}
			_, e = app.users.InsertOne(context.TODO(), User{
				Username: shared.Username(founderUsername),
				Password: hashedPassword,
			})
			if e != nil {
				panic(e)
			}

			// check back later
			time.Sleep(5 * time.Minute)
		}
	}()

	// run server
	log.Printf("Server started")
	log.Fatal(http.ListenAndServe(PORT, app.routes()))
}

func (app *App) routes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost%s/swagger/swagger.json", PORT)),
	))
	r.Get("/swagger/swagger.json", func(w http.ResponseWriter, _ *http.Request) {
		contents, e := os.ReadFile("docs/swagger.json")
		if e != nil {
			panic(e)
		}
		w.Write(contents)
	})

	r.Post("/api/v1/auth/check", app.check)
	r.Post("/api/v1/auth/login", app.login)
	r.Post("/api/v1/auth/register", app.register)

	return r
}

// check
//
//	@Sumamry		check if the token is valid
//	@Description	check if the token is valid
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"Bearer <token>"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		400	{object}	interface{}
//	@Security		ApiKeyAuth
//	@Router			/auth/check	[post]
func (app *App) check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// validate input
	claims, err := app.validateJWTToken(ctx, r.Header.Get("Authorization"))
	if err != nil {
		err.JSON(w, r)
		return
	}

	// return user
	var buf bytes.Buffer
	e := json.NewEncoder(&buf).Encode(claims)
	if e != nil {
		panic(e)
	}
	w.Write(buf.Bytes())
}

// login
//
//	@Summary		login with this
//	@Description	login with this
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		upwdRequest	true	"username and password for login"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	interface{}
//	@Security		ApiKeyAuth
//	@Router			/auth/login	[post]
func (app *App) login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// validate input
	req := upwdRequest{}
	e := json.NewDecoder(r.Body).Decode(&req)
	if e != nil {
		shared.NewError(http.StatusBadRequest, 0, "invalid body").JSON(w, r)
		return
	}

	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)

	// find user in db
	var user User
	e = app.users.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if e == mongo.ErrNoDocuments {
		shared.NewError(http.StatusBadRequest, 0, "user not found").JSON(w, r)
		return
	} else if e != nil {
		panic(e)
	}

	// check password
	e = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if e != nil {
		shared.NewError(http.StatusBadRequest, 0, "password is invalid").JSON(w, r)
		return
	}

	// create jwt token
	claims := Claims{
		Username: shared.Username(username),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, e := token.SignedString(app.secretKey)
	if e != nil {
		shared.NewError(http.StatusInternalServerError, 0, fmt.Sprintf("failed signing jwt: %v", e)).JSON(w, r)
		return
	}

	// success
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(map[string]string{"token": tokenString})
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}

// register
//
//	@Summary		register with this
//	@Description	register with this
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"a sponsor is needed for registering"
//	@Param			body			body		upwdRequest	true	"username and password for register"
//	@Success		200				{object}	map[string]interface{}
//	@Failure		400				{object}	interface{}
//	@Security		ApiKeyAuth
//	@Router			/auth/register	[post]
func (app *App) register(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// authorize
	_, err := app.validateJWTToken(ctx, r.Header.Get("Authorization"))
	if err != nil {
		err.JSON(w, r)
		return
	}

	// validate input
	req := upwdRequest{}
	e := json.NewDecoder(r.Body).Decode(&req)
	if e != nil {
		shared.NewError(http.StatusBadRequest, 0, "invalid body").JSON(w, r)
		return
	}

	username, err := shared.ValidateUsername(req.Username)
	if err != nil {
		err.JSON(w, r)
		return
	}

	password, err := shared.ValidatePassword(req.Password)
	if err != nil {
		err.JSON(w, r)
		return
	}

	// check if username is taken
	var user User
	e = app.users.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if e != mongo.ErrNoDocuments {
		shared.NewError(http.StatusBadRequest, 0, "username is taken").JSON(w, r)
		return
	}

	// hash password
	hashedPassword, e := bcrypt.GenerateFromPassword([]byte(password), 10)
	if e != nil {
		msg := fmt.Sprintf("failed hashing password: %v", e)
		shared.NewError(http.StatusInternalServerError, 0, msg).JSON(w, r)
		return
	}

	// insert new user
	app.users.InsertOne(ctx, User{
		Username: username,
		Password: hashedPassword,
	})
}

func (app *App) validateJWTToken(ctx context.Context, input string) (*Claims, *shared.Error) {
	// check if token is valid
	token := strings.TrimSpace(input)
	if token == "" {
		return nil, shared.NewError(http.StatusBadRequest, 0, "no token provided")
	}

	if !strings.HasPrefix(token, "Bearer ") {
		return nil, shared.NewError(http.StatusBadRequest, 0, "invalid token")
	}
	token = strings.TrimPrefix(token, "Bearer ")

	// parse token
	claims := Claims{}
	_, e := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return app.secretKey, nil
	})
	if e != nil {
		return nil, shared.NewError(http.StatusBadRequest, 0, "invalid token")
	}

	// check if token is expired
	if claims.ExpiresAt.Before(time.Now()) {
		return nil, shared.NewError(http.StatusBadRequest, 0, "token is expired")
	}

	// check if user exists
	var user User
	e = app.users.FindOne(ctx, bson.M{"username": claims.Username}).Decode(&user)
	if e == mongo.ErrNoDocuments {
		return nil, shared.NewError(http.StatusBadRequest, 0, "user not found")
	} else if e != nil {
		panic(e)
	}

	return &claims, nil
}

func getenv(name string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if len(value) == 0 {
		panic(fmt.Errorf("%s is not set", name))
	}
	return value
}
