package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

//Data model

type User struct {
	Id   string `json: "id"`
	Name string `json: "name"`
}

type Tweet struct {
	Id       string `json: "id"`
	AuthorId string `json: "authorId"`
	Message  string `json: "message"`
}

func (t Tweet) IsEmpty() bool {
	return t.Message == ""
}

func (t Tweet) IsUserPresent(u []User) bool {
	for _, user := range u {
		if t.AuthorId == user.Id {
			return true
		}
	}
	return false
}

var users []User
var tweets []Tweet

func main() {
	fmt.Println("Api Tweet")

	//Seed Data
	users = append(users, User{Id: "1", Name: "Sahil"}, User{Id: "2", Name: "Ritul"}, User{Id: "3", Name: "Sam"}, User{Id: "4", Name: "John"}, User{Id: "5", Name: "Doe"})
	tweets = append(tweets, Tweet{Id: "1", AuthorId: "1", Message: "My first tweet"}, Tweet{Id: "2", AuthorId: "2", Message: "My second tweet"}, Tweet{Id: "3", AuthorId: "3", Message: "My third tweet"}, Tweet{Id: "4", AuthorId: "4", Message: "My fourth tweet"}, Tweet{Id: "5", AuthorId: "5", Message: "My fifth tweet"})

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/listtweets", listTweets)
	http.HandleFunc("/createtweet", createTweet)
	http.HandleFunc("/getTweets/{id}", getTweets)

	log.Fatal(http.ListenAndServe(":5000", nil))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Welcome to the Tweet Api")
	w.Write([]byte("Welcome to the Tweet Api"))
}

func listTweets(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get all tweets")
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(tweets)
	if err != nil {
		http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)
	}
}

func createTweet(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create a tweet")
	w.Header().Set("Content-Type", "application/json")

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

	if tweet.IsEmpty() {
		http.Error(w, "Tweet message is empty", http.StatusNotFound)
		return
	}

	if tweet.IsUserPresent(users) == false {
		http.Error(w, "User does not exist", http.StatusNotFound)
		return
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	tweet.Id = strconv.Itoa(rand.Intn(100))

	tweets = append(tweets, tweet)

	respErr := json.NewEncoder(w).Encode(tweet)
	if respErr != nil {
		http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)
	}
}

func getTweets(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get user tweets")
	w.Header().Set("Content-Type", "application/json")

	// params := strings.Split(r.URL.Path, "/")
	// userId := params[len(params)-1]

	userId := r.PathValue("id")

	if userId == "" {
		http.Error(w, "Please provide the User Id", http.StatusNotFound)
		return
	}

	userTweets := []Tweet{}
	for _, tweet := range tweets {
		if tweet.AuthorId == userId {
			userTweets = append(userTweets, tweet)
		}
	}

	if len(userTweets) > 0 {
		json.NewEncoder(w).Encode(userTweets)
		return
	}
	err := json.NewEncoder(w).Encode("Unable to find tweets with the given user id")
	if err != nil {
		http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)
	}
}
