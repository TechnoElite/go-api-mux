package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/lib/pq"

	"github.com/subosito/gotenv"

	"github.com/gorilla/mux"
)

/*
post iss
type User struct {
	FullName string `json:"fullname"`
	UserName string `json:"username"`
	Email    string `json:"email"`
}
*/

// Post asd
type Post struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Author string `json:"author"`
}

var posts []Post = []Post{}
var db *sql.DB

func init() {
	gotenv.Load()
}
func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	router := mux.NewRouter()
	pgURL, err := pq.ParseURL(os.Getenv("ELEPHANTSQL_URL"))
	logFatal(err)
	db, err = sql.Open("postgres", pgURL)
	logFatal(err)

	err = db.Ping()
	logFatal(err)
	log.Println(pgURL)
	router.HandleFunc("/posts", additem).Methods("POST")
	router.HandleFunc("/posts", getAllPost).Methods("GET")
	router.HandleFunc("/posts/{id}", getPost).Methods("GET")
	router.HandleFunc("/posts/{id}", updatePost).Methods("PUT")
	router.HandleFunc("/posts/{id}", patchPost).Methods("PATCH")
	router.HandleFunc("/posts/{id}", deletePost).Methods("DELETE")
	http.ListenAndServe(":5000", router)
}
func additem(w http.ResponseWriter, r *http.Request) {
	var newpost Post
	var postid int
	json.NewDecoder(r.Body).Decode(&newpost)

	err := db.QueryRow("insert into book (title , body, author) values ($1,$2,$3) RETURNING id;",
		newpost.Title, newpost.Body, newpost.Author).Scan(&postid)
	logFatal(err)
	posts = append(posts, newpost)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(postid)

}
func getAllPost(w http.ResponseWriter, r *http.Request) {
	var post Post
	row, err := db.Query("select * from book")
	logFatal(err)
	defer row.Close()

	for row.Next() {
		err := row.Scan(&post.ID, &post.Title, &post.Body, &post.Author)
		logFatal(err)

		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
func getPost(w http.ResponseWriter, r *http.Request) {
	var temp string = mux.Vars(r)["id"]
	id, err := strconv.Atoi(temp)

	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("ID could not be converted to integer"))
	}

	if id > len(posts) {
		w.WriteHeader(404)
		w.Write([]byte("No post found with specified ID"))
		return
	}

	post := posts[id]
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)

}
func updatePost(w http.ResponseWriter, r *http.Request) {
	var temp string = mux.Vars(r)["id"]
	id, err := strconv.Atoi(temp)

	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("ID could not be converted to integer"))
		return
	}

	if id > len(posts) {
		w.WriteHeader(400)
		w.Write([]byte("No post found with specified I"))
		return
	}

	var updatePost Post
	json.NewDecoder(r.Body).Decode(&updatePost)
	result, err := db.Exec("update book set title=$1, body=$2, author=$3 where id=$4 RETURNING id",
		&updatePost.Title, &updatePost.Body, &updatePost.Author, &updatePost.ID)

	rowsupdated, err := result.RowsAffected()
	log.Fatal(err)
	posts[id] = updatePost
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rowsupdated)
}
func patchPost(w http.ResponseWriter, r *http.Request) {

	var idParam string = mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("ID could not be converted to integer"))
		return
	}

	if id >= len(posts) {
		w.WriteHeader(404)
		w.Write([]byte("No post found with specified ID"))
		return
	}

	post := &posts[id]
	json.NewDecoder(r.Body).Decode(post)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func deletePost(w http.ResponseWriter, r *http.Request) {

	var idParam string = mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("ID could not be converted to integer"))
		return
	}

	if id >= len(posts) {
		w.WriteHeader(404)
		w.Write([]byte("No post found with specified ID"))
		return
	}

	posts = append(posts[:id], posts[id+1:]...)

	w.WriteHeader(200)
}
