package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

type Receipt struct {
	ID           string  `json:"id,omitempty"`
	Retailer     string  `json:"retailer"`
	PurchaseDate string  `json:"purchaseDate"`
	PurchaseTime string  `json:"purchaseTime"`
	Total        string  `json:"total"`
	Items        []Item  `json:"items"`
	Points       int     `json:"points,omitempty"`
}

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

var ctx = context.Background()
var rdb *redis.Client

// Initialize Redis
func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Update this if Redis is running on a different host
		Password: "",               // Set if Redis has a password
		DB:       0,                // Default DB
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
}

// ProcessReceipt handles receipt processing
func ProcessReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Generate unique ID and calculate points
	receipt.ID = uuid.New().String()
	receipt.Points = calculatePoints(receipt)

	// Convert receipt to JSON and store in Redis
	receiptJSON, _ := json.Marshal(receipt)
	err := rdb.Set(ctx, receipt.ID, receiptJSON, 0).Err()
	if err != nil {
		http.Error(w, "Failed to store receipt", http.StatusInternalServerError)
		return
	}

	// Respond with generated receipt ID
	response := map[string]string{"id": receipt.ID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPoints retrieves the points for a given receipt ID
func GetPoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Retrieve receipt from Redis
	receiptJSON, err := rdb.Get(ctx, id).Result()
	if err == redis.Nil {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to fetch receipt", http.StatusInternalServerError)
		return
	}

	// Deserialize and return points
	var receipt Receipt
	json.Unmarshal([]byte(receiptJSON), &receipt)
	response := map[string]int{"points": receipt.Points}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// calculatePoints applies the rules to determine receipt points
func calculatePoints(receipt Receipt) int {
	points := 0

	// Rule: One point per alphanumeric character in retailer name
	alnumRegex := regexp.MustCompile(`[a-zA-Z0-9]`)
	points += len(alnumRegex.FindAllString(receipt.Retailer, -1))

	// Convert total to float
	total, err := strconv.ParseFloat(receipt.Total, 64)
	if err == nil {
		// Rule: 50 points if total is a round dollar amount
		if total == math.Floor(total) {
			points += 50
		}
		// Rule: 25 points if total is a multiple of 0.25
		if math.Mod(total, 0.25) == 0 {
			points += 25
		}
	}

	// Rule: 5 points for every two items
	points += (len(receipt.Items) / 2) * 5

	// Rule: If item description length is a multiple of 3, add price * 0.2 (rounded up)
	for _, item := range receipt.Items {
		trimmedDesc := strings.TrimSpace(item.ShortDescription)
		if len(trimmedDesc)%3 == 0 {
			price, err := strconv.ParseFloat(item.Price, 64)
			if err == nil {
				points += int(math.Ceil(price * 0.2))
			}
		}
	}

	// Rule: 6 points if purchase day is odd
	dateParts := strings.Split(receipt.PurchaseDate, "-")
	if len(dateParts) == 3 {
		day, err := strconv.Atoi(dateParts[2])
		if err == nil && day%2 == 1 {
			points += 6
		}
	}

	// Rule: 10 points if purchase time is between 2:00 PM and 4:00 PM
	purchaseTime, err := time.Parse("15:04", receipt.PurchaseTime)
	if err == nil {
		if purchaseTime.Hour() >= 14 && purchaseTime.Hour() < 16 {
			points += 10
		}
	}

	return points
}

func main() {
	initRedis()

	router := mux.NewRouter()
	router.HandleFunc("/receipts/process", ProcessReceipt).Methods("POST")
	router.HandleFunc("/receipts/{id}/points", GetPoints).Methods("GET")

	fmt.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}