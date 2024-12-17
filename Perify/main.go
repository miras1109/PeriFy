package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Product структура для продукта
type Product struct {
	ID          string `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string `json:"name" bson:"name"`
	Price       int    `json:"price" bson:"price"`
	Description string `json:"description" bson:"description"`
}

var collection *mongo.Collection

// Middleware для CORS
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Подключение к MongoDB
func connectDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("Ошибка подключения к MongoDB:", err)
	}
	collection = client.Database("perify").Collection("products")
	fmt.Println("Подключение к MongoDB успешно")
}

// Создание продукта (POST)
func createProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}
	if product.Name == "" || product.Price <= 0 {
		http.Error(w, "Все поля должны быть заполнены корректно", http.StatusBadRequest)
		return
	}
	collection.InsertOne(context.Background(), product)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Продукт добавлен"})
}

// Получение всех продуктов (GET)
func getProducts(w http.ResponseWriter, r *http.Request) {
	var products []Product
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, "Ошибка получения данных", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var product Product
		cursor.Decode(&product)
		products = append(products, product)
	}
	json.NewEncoder(w).Encode(products)
}

// Обновление продукта (PUT)
func updateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	var updatedProduct Product
	if err := json.NewDecoder(r.Body).Decode(&updatedProduct); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$set": updatedProduct})
	if err != nil {
		http.Error(w, "Ошибка обновления продукта", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Продукт обновлен"})
}

// Удаление продукта (DELETE)
func deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Ошибка удаления продукта", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Продукт удален"})
}

func main() {
	connectDB()

	mux := http.NewServeMux()
	mux.HandleFunc("/products", getProducts)
	mux.HandleFunc("/products/create", createProduct)
	mux.HandleFunc("/products/update", updateProduct)
	mux.HandleFunc("/products/delete", deleteProduct)

	handler := enableCORS(mux)

	fmt.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
