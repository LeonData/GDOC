package main

import (
 "database/sql"
 "fmt"
 "github.com/gorilla/mux"
 _ "github.com/mattn/go-sqlite3"
 "log"
 "net/http"
 "html/template"
)

const (
 PORT = ":8080"
)

type Page struct {
 ID      int
 GUID    string
 Title   string
 RawContent string
 Content template.HTML
 Date    string
}

var database *sql.DB

func ServePage(w http.ResponseWriter, r *http.Request){
 vars := mux.Vars(r)
 pageID := vars["post_guid"]
 fmt.Println("Need to get a page for GUID = " + pageID)
 
 p := Page{}

 row := database.QueryRow("select id, post_guid, post_title, post_content, post_date from posts where post_guid=$1", pageID)

 err := row.Scan(&p.ID, &p.GUID, &p.Title, &p.RawContent,&p.Date)
 if err != nil {
  http.Error(w, http.StatusText(404), http.StatusNotFound)
  log.Println("Error: couldn't get data!")
 }else{ 
  p.Content = template.HTML(p.RawContent)
  t, _:= template.ParseFiles("templates/blog.html")
  t.Execute(w, p)
 }
}


func RedirIndex(w http.ResponseWriter, r *http.Request){
 http.Redirect(w,r,"/home", 301)
}

func ServeIndex(w http.ResponseWriter, r *http.Request){
 var Pages = [] Page{}
 pages, err := database.Query("select post_title,post_content,post_date from posts order by $1 desc", "post_date")
 if err != nil {
   fmt.Fprintln(w, err.Error)
 }
 defer pages.Close()
 for pages.Next() {
  thisPage := Page{}
  pages.Scan(&thisPage.Title, &thisPage.RawContent, &thisPage.Date)
  thisPage.Content = template.HTML(thisPage.RawContent)
  Pages = append(Pages, thisPage)
 }
 t, _ := template.ParseFiles("templates/index.html")
 t.Execute(w, Pages)  
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

 routes := mux.NewRouter()
 routes.HandleFunc("/page/{post_guid:[0-9a-zA-Z-]+}", ServePage)
 routes.HandleFunc("/home", ServeIndex)
 routes.HandleFunc("/", RedirIndex)
 http.Handle("/", routes)
 http.ListenAndServe(PORT, nil)
}

