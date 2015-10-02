package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type Message struct {
	Id      int    `json:"id"`
	System  string `json:"system"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Version string `json:"version"`
	Server  string `json:"server"`
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
	router.HandleFunc("/alert/{id}", PutAlert).Methods("PUT")
	router.HandleFunc("/alerts/{system}", GetAlerts)
	router.HandleFunc("/alerts", DeleteAlerts).Methods("DELETE")

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
		Version: r.FormValue("version"),
		Server:  r.FormValue("server"),
	}

	db := StartDatabase()
	defer db.Close()

	_, err := db.Exec("INSERT INTO messages (system, type, title, message, version, server) VALUES (?, ?, ?, ?, ?, ?)", m.System, m.Type, m.Title, m.Message, m.Version, m.Server)
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

func DeleteAlerts(w http.ResponseWriter, r *http.Request) {
	db := StartDatabase()
	defer db.Close()

	_, err := db.Exec("DELETE FROM messages")
	if err != nil {
		panic(err.Error())
	}

	_, err = db.Exec("ALTER TABLE messages AUTO_INCREMENT=1")
	if err != nil {
		panic(err.Error())
	}
}

func PutAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	m := &Message{
		Id:      id,
		System:  r.FormValue("system"),
		Type:    r.FormValue("type"),
		Title:   r.FormValue("title"),
		Message: r.FormValue("message"),
		Version: r.FormValue("version"),
		Server:  r.FormValue("server"),
	}

	db := StartDatabase()
	defer db.Close()

	_, err := db.Exec("UPDATE messages SET system=?, type=?, title=?, message=?, version=?, server=? WHERE id=?", m.System, m.Type, m.Title, m.Message, m.Version, m.Server, m.Id)
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

	sql := `SELECT m.id, m.system, m.type, m.title, m.message, m.version, m.server
FROM messages AS m
	JOIN (
		SELECT pkSite AS id, strAdresse AS system, strVersion AS version, strNom AS server1, strIP AS server2
		FROM site
			LEFT JOIN r_serveur AS server
				ON fkServeurFichier=pkServeur
        		    OR fkServeurBD=pkServeur
		            OR fkServeurBatch=pkServeur
        		    OR fkServeurMail=pkServeur

		WHERE
		    strAdresse=?
	) AS site
		ON (m.system='all' OR m.system=site.system)
			AND (m.version='' OR m.version=site.version)
			AND (m.server='' OR m.server=site.server1 OR m.server=site.server2)
GROUP BY m.id
ORDER BY m.id`
	rows, err := db.Query(sql, system)
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.Id, &m.System, &m.Type, &m.Title, &m.Message, &m.Version, &m.Server)
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
