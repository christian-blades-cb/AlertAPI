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
	router.HandleFunc("/alert", PostAlert).Methods("POST")
	router.HandleFunc("/alert/{id}", DeleteAlert).Methods("DELETE")
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

func PostAlert(w http.ResponseWriter, r *http.Request) {
	m := &Message{
		System:  r.FormValue("system"),
		Type:    r.FormValue("type"),
		Title:   r.FormValue("title"),
		Message: r.FormValue("message"),
	}

	db := StartDatabase()
	defer db.Close()

	_, err := db.Exec("INSERT INTO messages (system, type, title, message) VALUES (?, ?, ?, ?)", m.System, m.Type, m.Title, m.Message)
	if err != nil {
		panic(err.Error())
	}
}

func DeleteAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	db := StartDatabase()
	defer db.Close()

	_, err := db.Exec("DELETE FROM messages WHERE id=?", id)
	if err != nil {
		panic(err.Error())
	}
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
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
