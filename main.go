package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"log"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/comprehend"

	"github.com/coreos/go-oidc"
)

const clientID = "ppGn4wahztmqeKLBQd7wVT8UlfUDwW3m"
const authDomain = "https://dev-vuys1721.us.auth0.com"

// ScoreResponse response struct for sentiment score
type ScoreResponse struct {
	Score float32 `json:"score"`
	Text  string  `json:"text"`
}

// TextInput input struct for sentiment score
type TextInput struct {
	Text string `json:"text" validate:"required"`
}

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

var comprehendClient *comprehend.Comprehend

var provider *oidc.Provider
var verifier *oidc.IDTokenVerifier

func authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		h := c.Request().Header.Get(echo.HeaderAuthorization)
		hSplit := strings.Split(h, "Bearer ")
		if len(hSplit) != 2 {
			return echo.NewHTTPError(http.StatusUnauthorized, "Malformed bearer token.")
		}
		_, vErr := verifier.Verify(ctx, hSplit[1])
		if vErr != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, vErr.Error())
		}
		return next(c)
	}
}

func main() {
	ctx := context.Background()
	sess := session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))
	comprehendClient = comprehend.New(sess)

	p, err := oidc.NewProvider(ctx, authDomain+"/")
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	provider = p
	verifier = p.Verifier(&oidc.Config{ClientID: clientID})
	// Echo instance
	e := echo.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(authMiddleware)

	// Routes
	e.POST("/analyze-sentiment", analyzeSentiment)

	// Validators
	e.Validator = &customValidator{validator: validator.New()}

	// Start server
	e.Debug = true
	e.Logger.Error(e.Start(":8080"))
}

// Handler
func analyzeSentiment(c echo.Context) error {
	ti := new(TextInput)
	if err := c.Bind(ti); err != nil {
		log.Print(err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(ti); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	lc := "en"
	ds, err := comprehendClient.DetectSentiment(&comprehend.DetectSentimentInput{
		Text:         &ti.Text,
		LanguageCode: &lc,
	})
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	r := &ScoreResponse{
		Score: float32(*ds.SentimentScore.Positive) - float32(*ds.SentimentScore.Negative),
		Text:  ti.Text,
	}
	return c.JSON(http.StatusOK, r)
}
