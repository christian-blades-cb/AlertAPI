package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Message struct {
	Id      int    `json:"id"`
	System  string `json:"system"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type MessagesJSON struct {
	Messages []Message `json:"notifications"`
}

var messages []Message

func main() {
	StartDatabase()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/system", SystemIndex)
	router.HandleFunc("/message", SaveMessage).Methods("POST")
	router.HandleFunc("/alerts/{system}", SendMessage)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func SystemIndex(w http.ResponseWriter, r *http.Request) {
	db := StartDatabase()
	results := GetSystems(db)
	results = append(results, "all", "4.0", "4.1")
	b, err := json.Marshal(results)
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Write(b)
	defer db.Close()
}

func SaveMessage(w http.ResponseWriter, r *http.Request) {
	db := StartDatabase()
	//validate that the message and place are there
	message := r.FormValue("message")
	system := r.FormValue("system")
	typearo := r.FormValue("type")
	title := r.FormValue("title")

	fmt.Print(w, message+system+typearo+title)
	//create a new entry into database
	_, err := db.Exec("INSERT INTO messages(title,message,system,type) VALUES(?,?,?,?)", title, message, system, typearo)
	if err != nil {
		panic(err.Error())
	}
	fmt.Fprintln(w, "sent")
	defer db.Close()
}

func SendMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	system := vars["system"]
	db := StartDatabase()
	results := GetMessage(db, system)
	b, err := json.Marshal(results)
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Write(b)
	defer db.Close()

}

func GetMessage(db *sql.DB, system string) [][]string {
	rows, err := db.Query("SELECT id,type,title,message FROM messages WHERE system='cb1-luceo.dev'")
	fmt.Print(system)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	results := GetResultsTwo(rows)
	defer rows.Close()
	return results
}

func GetSystems(db *sql.DB) []string {
	rows, err := db.Query("SELECT strAdresse FROM site WHERE SUBSTRING_INDEX(strVersion, '.', 1) > 3")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	results := GetResults(rows)
	defer rows.Close()
	return results
}

func StartDatabase() *sql.DB {
	db, err := sql.Open("mysql", "luceo:luceo_password@(cb1-luceo.dev:3306)/ate")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	return (db)
}

func GetResults(rows *sql.Rows) []string {

	var (
		results []string
		result  string
	)
	for rows.Next() {
		err := rows.Scan(&result)
		if err != nil {
			fmt.Println(err)
		}
		results = append(results, result)
	}
	return results
}
func GetResultsTwo(rows *sql.Rows) [][]string {
	var (
		results [][]string
		id      string
		types   string
		title   string
		message string
		i       int
	)
	results = make([][]string, 1)
	i = 0
	for rows.Next() {
		err := rows.Scan(&id, &types, &title, &message)
		if err != nil {
			fmt.Println(err)
		}
		messages := []string{id, types, title, message}
		fmt.Println(messages)
		for index, element := range messages {
			results[i][index] = element
		}
		i++
	}
	return results
}
