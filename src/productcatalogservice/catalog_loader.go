//
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.

package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"time"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice/genproto"
	"github.com/golang/protobuf/jsonpb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func loadCatalog(catalog *pb.ListProductsResponse) error {
	catalogMutex.Lock()
	defer catalogMutex.Unlock()

	if os.Getenv("MONGO_CONNECTION_STRING") != "" {
		return loadCatalogFromMongo(catalog)
	}

	return loadCatalogFromLocalFile(catalog)
}

func loadCatalogFromLocalFile(catalog *pb.ListProductsResponse) error {
	log.Info("loading catalog from local products.json file...")

	catalogJSON, err := os.ReadFile("products.json")
	if err != nil {
		log.Warnf("failed to open product catalog json file: %v", err)
		return err
	}

	if err := jsonpb.Unmarshal(bytes.NewReader(catalogJSON), catalog); err != nil {
		log.Warnf("failed to parse the catalog JSON: %v", err)
		return err
	}

	log.Info("successfully parsed product catalog json")
	return nil
}

// mongoProduct represents a product document in MongoDB
type mongoProduct struct {
	ID                   string   `bson:"id"`
	Name                 string   `bson:"name"`
	Description          string   `bson:"description"`
	Picture              string   `bson:"picture"`
	PriceUsdCurrencyCode string   `bson:"price_usd_currency_code"`
	PriceUsdUnits        int64    `bson:"price_usd_units"`
	PriceUsdNanos        int32    `bson:"price_usd_nanos"`
	Categories           []string `bson:"categories"`
}

func loadCatalogFromMongo(catalog *pb.ListProductsResponse) error {
	log.Info("loading catalog from MongoDB...")

	connStr := os.Getenv("MONGO_CONNECTION_STRING")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(connStr))
	if err != nil {
		log.Warnf("failed to connect to MongoDB: %v", err)
		return err
	}
	defer client.Disconnect(ctx)

	// Verify connectivity
	if err := client.Ping(ctx, nil); err != nil {
		log.Warnf("failed to ping MongoDB: %v", err)
		return err
	}

	collection := client.Database("shopdb").Collection("products")
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Warnf("failed to query MongoDB: %v", err)
		return err
	}
	defer cursor.Close(ctx)

	catalog.Products = catalog.Products[:0]
	for cursor.Next(ctx) {
		var mp mongoProduct
		if err := cursor.Decode(&mp); err != nil {
			log.Warnf("failed to decode MongoDB document: %v", err)
			return err
		}

		// Convert categories to lowercase
		for i, cat := range mp.Categories {
			mp.Categories[i] = strings.ToLower(cat)
		}

		product := &pb.Product{
			Id:          mp.ID,
			Name:        mp.Name,
			Description: mp.Description,
			Picture:     mp.Picture,
			PriceUsd: &pb.Money{
				CurrencyCode: mp.PriceUsdCurrencyCode,
				Units:        mp.PriceUsdUnits,
				Nanos:        mp.PriceUsdNanos,
			},
			Categories: mp.Categories,
		}
		catalog.Products = append(catalog.Products, product)
	}

	if err := cursor.Err(); err != nil {
		log.Warnf("cursor error: %v", err)
		return err
	}

	log.Infof("successfully loaded %d products from MongoDB", len(catalog.Products))
	return nil
}
