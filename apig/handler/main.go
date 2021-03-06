package main

import (
	"context"
	"log"
	"time"

	"futuagro.com/pkg/config"
	"futuagro.com/pkg/domain/services"
	"futuagro.com/pkg/http/rest"
	"futuagro.com/pkg/store"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

var chiLambda *chiadapter.ChiLambda

func init() {
	conf := config.NewDefaultConfig()
	mongoClient, err := store.NewDB(conf)
	if err != nil {
		log.Fatalf("FATAL: %v\n", err)
	}
	supplierRepository := store.NewMongoSupplierRepository(conf, mongoClient)
	countryRepository := store.NewMongoCountryRepository(conf, mongoClient)
	cityRepository := store.NewMongoCityRepository(conf, mongoClient)
	itemRepository := store.NewMongoItemRepository(conf, mongoClient)
	variantRepository := store.NewMongoVariantRepository(conf, mongoClient)
	cropRepository := store.NewMongoCropRepository(conf, mongoClient)
	userRepository := store.NewMongoUserRepository(conf, mongoClient)

	supplierService := services.NewSupplierService(supplierRepository)
	countryService := services.NewCountryService(countryRepository)
	cityService := services.NewCityService(cityRepository)
	itemService := services.NewItemService(itemRepository)
	variantService := services.NewVariantService(variantRepository)
	cropService := services.NewCropService(cropRepository)
	userService := services.NewUserService(userRepository)
	authService := services.NewAuthService(userRepository)

	// Setup chi router
	r := chi.NewRouter()
	// Setup CORS
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"ETag", "Link", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "X-OAuth-Scopes", "X-Accepted-OAuth-Scopes"},
		AllowCredentials: true,
		MaxAge:           3600, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(5 * time.Minute))

	rSupplier := rest.SupplierHandler{Service: supplierService}
	rCountry := rest.CountryHandler{Service: countryService}
	rCity := rest.CityHandler{Service: cityService}
	rItem := rest.ItemHandler{Service: itemService}
	rVariant := rest.VariantHandler{Service: variantService}
	rCrop := rest.CropHandler{Service: cropService}
	rUser := rest.UserHandler{Service: userService}
	rAuth := rest.AuthHandler{Service: authService}

	r.Mount("/suppliers", rSupplier.NewRouter())
	r.Mount("/countries", rCountry.NewRouter())
	r.Mount("/country-states", rCity.NewRouter())
	r.Mount("/items", rItem.NewRouter())
	r.Mount("/items/{itemID}/variants", rVariant.NewRouter())
	r.Mount("/crops", rCrop.NewRouter())
	r.Mount("/users", rUser.NewRouter())
	r.Mount("/auth", rAuth.NewRouter())

	chiLambda = chiadapter.New(r)
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return chiLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
