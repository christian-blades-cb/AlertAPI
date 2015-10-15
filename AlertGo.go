package main // import "github.com/greygore/AlertAPI"

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

type Alert struct {
	Id      int    `json:"id"`
	System  string `json:"system"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Version string `json:"version"`
	Server  string `json:"server"`
}

type AlertsJSON struct {
	Alerts []Alert `json:"notifications"`
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
	a := &Alert{
		System:  r.FormValue("system"),
		Type:    r.FormValue("type"),
		Title:   r.FormValue("title"),
		Message: r.FormValue("message"),
		Version: r.FormValue("version"),
		Server:  r.FormValue("server"),
	}

	db := StartDatabase()
	defer db.Close()

	_, err := db.Exec("INSERT INTO alerts (system, type, title, message, version, server) VALUES (?, ?, ?, ?, ?, ?)", a.System, a.Type, a.Title, a.Message, a.Version, a.Server)
	if err != nil {
		panic(err.Error())
	}
}

func DeleteAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	db := StartDatabase()
	defer db.Close()

	_, err := db.Exec("DELETE FROM alerts WHERE id=?", id)
	if err != nil {
		panic(err.Error())
	}
}

func DeleteAlerts(w http.ResponseWriter, r *http.Request) {
	db := StartDatabase()
	defer db.Close()

	_, err := db.Exec("DELETE FROM alerts")
	if err != nil {
		panic(err.Error())
	}

	_, err = db.Exec("ALTER TABLE alerts AUTO_INCREMENT=1")
	if err != nil {
		panic(err.Error())
	}
}

func PutAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	a := &Alert{
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

	_, err := db.Exec("UPDATE alerts SET system=?, type=?, title=?, message=?, version=?, server=? WHERE id=?", a.System, a.Type, a.Title, a.Message, a.Version, a.Server, a.Id)
	if err != nil {
		panic(err.Error())
	}
}

func GetAlerts(w http.ResponseWriter, r *http.Request) {
	var alerts []Alert

	vars := mux.Vars(r)
	system := vars["system"]

	db := StartDatabase()
	defer db.Close()

	sql := `SELECT a.id, a.system, a.type, a.title, a.message, a.version, a.server
FROM alerts AS a
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
		UNION
		SELECT 0 AS id, 'cb1-luceo.dev' AS system, '4.0' AS version, 'cb1-luceo.dev' AS server1, 'cb1-luceo.dev' AS server2
		UNION
		SELECT -1 AS id, 'demo-us-ocb1.cb1-luceo.dev' AS system, '4.0' AS version, 'cb1-luceo.dev' AS server1, 'demo-us-ocb1.cb1-luceo.dev' AS server2
	) AS site
		ON (a.system='all' OR a.system=site.system)
			AND (a.version='' OR a.version=site.version)
			AND (a.server='' OR a.server=site.server1 OR a.server=site.server2)
GROUP BY a.id
ORDER BY a.id`
	rows, err := db.Query(sql, system)
	for rows.Next() {
		var a Alert
		err := rows.Scan(&a.Id, &a.System, &a.Type, &a.Title, &a.Message, &a.Version, &a.Server)
		if err != nil {
			panic(err.Error())
		}
		alerts = append(alerts, a)
	}

	j, err := json.Marshal(AlertsJSON{Alerts: alerts})
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
