package main

import (
 "database/sql"
 "fmt"
 "github.com/gorilla/mux"
 _ "github.com/mattn/go-sqlite3"
 "log"
 "net/http"
)

const (
 PORT = ":8080"
)

type Page struct {
 ID      int
 GUID    string
 Title   string
 Content string
 Date    string
}

var database *sql.DB

func ServePage(w http.ResponseWriter, r *http.Request){
 vars := mux.Vars(r)
 pageID := vars["id"]
 fmt.Println("Need to get a page for id = " + pageID)
 
 p := Page{}

 row := database.QueryRow("select id, post_guid, post_title, post_content, post_date from posts where id=$1", pageID)

 err := row.Scan(&p.ID, &p.GUID, &p.Title, &p.Content,&p.Date)
 if err != nil {
  log.Println("Error in func querying a database")
  log.Println(err.Error)
 } 

 html := `<html><head><title>`+p.Title+`</title></head><body>`+p.Title+`</h1><div>`+p.Content+`</div></body></html>`
 fmt.Fprintln(w,html)

}

func main(){

 db , err := sql.Open("sqlite3", "cms.db")
 if err != nil {
   fmt.Println("Cannot connect!")
   log.Println("Couldn't connect!")
   log.Println(err.Error)
 } else {
  fmt.Println("SUCCESS! Connected to SQLITE3 DB!")
 }

 database = db

/*
 row := database.QueryRow("select id, post_guid, post_title, post_content, post_date from posts where id=$1", 0)

 pp := Page{}

 err = row.Scan(&pp.ID, &pp.GUID, &pp.Title, &pp.Content, &pp.Date)
 if err != nil {
  log.Println("Error in MAIN scanning a database")
  log.Println(err.Error)
 }else{

  fmt.Println("In main: title="+pp.Title+";content="+pp.Content+";date="+pp.Date)

 } 
*/

 routes := mux.NewRouter()
 routes.HandleFunc("/page/{id:[0-9]+}", ServePage)
 http.Handle("/", routes)
 http.ListenAndServe(PORT, nil)
}

