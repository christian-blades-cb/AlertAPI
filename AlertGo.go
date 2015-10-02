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

func main() {
	StartDatabase()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/system", SystemIndex)
	router.HandleFunc("/message", SaveMessage).Methods("POST")
	router.HandleFunc("/alerts/{system}", GetAlerts)

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

func GetAlerts(w http.ResponseWriter, r *http.Request) {
	var messages []Message

	vars := mux.Vars(r)
	system := vars["system"]

	db := StartDatabase()
	defer db.Close()

	rows, err := db.Query("SELECT id, system, type, title, message FROM messages WHERE system='all' OR system=? ORDER BY id", system)
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.Id, &m.System, &m.Type, &m.Title, &m.Message)
		if err != nil {
			panic(err.Error())
		}
		messages = append(messages, m)
	}

	j, err := json.Marshal(MessagesJSON{Messages: messages})
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
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
