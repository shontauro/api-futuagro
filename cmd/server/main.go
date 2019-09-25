package main

import (
	"fmt"
	"log"

	"futuagro.com/pkg/config"
	"futuagro.com/pkg/domain/services"
	"futuagro.com/pkg/http"
	"futuagro.com/pkg/store"
	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
		log.Fatal(err)
	}
}

func main() {
	conf := config.NewDefaultConfig()
	// Print out environment variables
	fmt.Println(conf.Database.URI)
	fmt.Println(conf.Database.PoolSize)
	fmt.Println(conf.Database.Name)
	fmt.Println(conf.Port)

	mongoClient, err := store.NewDB(conf)
	if err != nil {
		log.Fatalf("FATAL: %v\n", err)
	}

	supplierRepository := store.NewMongoSupplierRepository(conf, mongoClient)
	countryRepository := store.NewMongoCountryRepository(conf, mongoClient)
	cityRepository := store.NewMongoCityRepository(conf, mongoClient)

	supplierService := services.NewSupplierService(supplierRepository)
	countryService := services.NewCountryService(countryRepository)
	cityService := services.NewCityService(cityRepository)

	server := http.NewServer(conf, supplierService, countryService, cityService)

	server.Run()
}
