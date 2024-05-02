package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type User struct {
	Id   string `json: "id"`
	Name string `json: "name"`
}

type Tweet struct {
	Id       string `json: "id"`
	AuthorId string `json: "authorId"`
	Message  string `json: "message"`
}

func (t Tweet) isEmpty() bool {
	return t.Message == ""
}

func main() {
	fmt.Println("Api Tweet")

	db := connectToDb()

	defer db.Close()

	createTweetTable(db)

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/registerUser", registerUser(db))
	http.HandleFunc("/listTweets", listTweets(db))
	http.HandleFunc("/createTweet", createTweet(db))
	http.HandleFunc("/getTweets/{id}", getTweets(db))

	log.Fatal(http.ListenAndServe(":5000", nil))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Welcome to the Tweet Api")
	w.Write([]byte("Welcome to the Tweet Api"))
}

func listTweets(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Get all tweets")

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		query := `SELECT id, message, author_id FROM tweet`
		rows, err := db.Query(query)

		if err != nil {
			log.Fatal("Unable to fetch tweets from database: ", err)
		}
		var id int
		var message string
		var authorId int

		var allTweets []Tweet

		for rows.Next() {
			errSc := rows.Scan(&id, &message, &authorId)

			if errSc != nil {
				log.Fatal("Unable to scan the data from tweet table rows: ", errSc)
			}

			allTweets = append(allTweets, Tweet{Id: strconv.Itoa(id), AuthorId: strconv.Itoa(authorId), Message: message})
		}

		errEn := json.NewEncoder(w).Encode(allTweets)
		if errEn != nil {
			http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
	}
}

func createTweet(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Create a tweet")

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var tweet Tweet
		err := json.NewDecoder(r.Body).Decode(&tweet)

		if err != nil {
			http.Error(w, "Internal Server Error while decoding the JSON", http.StatusInternalServerError)
			return
		}

		if tweet.isEmpty() {
			http.Error(w, "Tweet message is empty", http.StatusBadRequest)
			return
		}

		query := `INSERT INTO tweet (message, author_id) VALUES ($1, $2) RETURNING id`
		var id int
		errQ := db.QueryRow(query, tweet.Message, tweet.AuthorId).Scan(&id)

		if errQ != nil {
			log.Fatal("Unable to insert tweet into database: ", errQ)
		}

		tweet.Id = strconv.Itoa(id)
		respErr := json.NewEncoder(w).Encode(tweet)

		if respErr != nil {
			http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
	}
}

func getTweets(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Get user tweets")

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userId := r.PathValue("id")

		if userId == "" {
			http.Error(w, "Please provide the User Id", http.StatusBadRequest)
			return
		}

		query := `SELECT id, message, author_id FROM tweet WHERE author_id = $1`
		rows, err := db.Query(query, userId)

		if err != nil {
			log.Fatal("Unable to fetch user tweet from database: ", err)
		}

		userTweets := []Tweet{}
		for rows.Next() {
			var tweet Tweet
			err := rows.Scan(&tweet.Id, &tweet.Message, &tweet.AuthorId)

			if err != nil {
				log.Fatal("Unable to scan the tweet data: ", err)
			}

			userTweets = append(userTweets, tweet)
		}

		if len(userTweets) > 0 {
			err := json.NewEncoder(w).Encode(userTweets)

			if err != nil {
				log.Fatal("Unable to encode the user tweets: ", err)
			}

			w.Header().Set("Content-Type", "application/json")
			return
		}

		http.Error(w, "Unable to find tweets with the given user id", http.StatusInternalServerError)
	}
}

func connectToDb() *sql.DB {
	conStr := "postgres://postgres:sahil123@localhost:5433/tweet?sslmode=disable"

	db, err := sql.Open("postgres", conStr)

	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}

func registerUser(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Registering User")

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var user User
		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil {
			log.Fatal("Error while decoding the user: ", err)
		}

		if user.Name == "" {
			http.Error(w, "User Name not found", http.StatusBadRequest)
			return
		}

		query := `INSERT INTO users (name) VALUES ($1) RETURNING id`
		var id int
		errQ := db.QueryRow(query, user.Name).Scan(&id)

		if errQ != nil {
			log.Fatal("Error while inserting the user into database: ", errQ)
		}

		user.Id = strconv.Itoa(id)
		errEn := json.NewEncoder(w).Encode(user)

		if errEn != nil {
			log.Fatal("Error while encoding the user: ", err)
		}

		w.Header().Set("Content-Type", "application/json")
	}
}

func createTweetTable(db *sql.DB) {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			created timestamp DEFAULT NOW()
		);
			
		CREATE TABLE IF NOT EXISTS tweet (
			id SERIAL PRIMARY KEY,
			message VARCHAR(200) NOT NULL,
			author_id INT,
			created timestamp DEFAULT NOW(),
			CONSTRAINT fk_users FOREIGN KEY(author_id) REFERENCES users(id)
		);`

	_, err := db.Exec(query)
	fmt.Println(err)
	if err != nil {
		log.Fatal(err)
	}
}
