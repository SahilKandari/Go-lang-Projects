package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//Data model

type User struct {
	Id   string
	Name string
}

type Tweet struct {
	Id      string
	UserId  string
	Message string
}

func (t Tweet) IsEmpty() bool {
	return t.Message == ""
}

var users []User
var tweets []Tweet

func main() {
	fmt.Println("Api Tweet")

	//Seed Data
	users = append(users, User{Id: "1", Name: "Sahil"}, User{Id: "2", Name: "Ritul"}, User{Id: "3", Name: "Sam"}, User{Id: "4", Name: "John"}, User{Id: "5", Name: "Doe"})
	tweets = append(tweets, Tweet{Id: "1", UserId: "1", Message: "My first tweet"}, Tweet{Id: "2", UserId: "2", Message: "My second tweet"}, Tweet{Id: "3", UserId: "3", Message: "My third tweet"}, Tweet{Id: "4", UserId: "4", Message: "My fourth tweet"}, Tweet{Id: "5", UserId: "5", Message: "My fifth tweet"})

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/tweets", getAllTweets)
	http.HandleFunc("/tweet", createTweet)
	http.HandleFunc("/tweets/user/{userId}", getUserTweets)

	log.Fatal(http.ListenAndServe(":5000", nil))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Welcome to the Tweet Api")
	w.Write([]byte("Welcome to the Tweet Api"))
}

func getAllTweets(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get all tweets")
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(tweets)
}

func createTweet(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create a tweet")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tweet Tweet
	_ = json.NewDecoder(r.Body).Decode(&tweet)

	if tweet.IsEmpty() {
		json.NewEncoder(w).Encode("Tweet is empty")
		return
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	tweet.Id = strconv.Itoa(rand.Intn(100))

	tweets = append(tweets, tweet)

	json.NewEncoder(w).Encode(tweet)
}

func getUserTweets(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get user tweets")

	params := strings.Split(r.URL.Path, "/")
	userId := params[len(params)-1]

	if userId == "" {
		json.NewEncoder(w).Encode("User Id is required")
		return
	}

	userTweets := []Tweet{}
	for _, tweet := range tweets {
		if tweet.UserId == userId {
			userTweets = append(userTweets, tweet)
		}
	}

	if len(userTweets) > 0 {
		json.NewEncoder(w).Encode(userTweets)
		return
	}
	json.NewEncoder(w).Encode("Unable to find tweets with the given user id")
}
