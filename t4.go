package main

import (
 "database/sql"
 "fmt"
 "github.com/gorilla/mux"
 _ "github.com/mattn/go-sqlite3"
 "log"
 "net/http"
 "html/template"
 //"crypto/tls"
 "strconv"
 "encoding/json"
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

type JSONResponse struct {
 Fields map[string]string
}

var database *sql.DB

func APICommentPost(w http.ResponseWriter, r *http.Request){
 var commentAdded bool
 err := r.ParseForm()
 if err != nil {
  log.Println(err.Error)
 }
 name  := r.FormValue("name")
 email := r.FormValue("email")
 comments := r.FormValue("comments")

 res, err1 := database.Exec("INSERT INTO comments SET comment_name = $1, comment_email = $2, comemnt_text = $3", 
   name, email, comments)
 if err1 != nil {
  fmt.Println("Error adding info to comments table") 
  log.Println(err.Error)
 }

 id, err3 := res.LastInsertId()
 if err3 != nil {
  commentAdded = false
 }else{
  commentAdded = true
 }
 
 commentAddedBool := strconv.FormatBool(commentAdded)
 var resp JSONResponse
 resp.Fields["id"] = string(id)
 resp.Fields["added"] = commentAddedBool
 jsonResp, _ := json.Marshal(resp)
 w.Header().Set("Content-Type", "application/json")
 fmt.Fprintln(w, jsonResp) 
}

func (p Page) TruncateText() template.HTML {
 chars := 0
 for i, _ := range p.RawContent {
  chars++
  if chars > 50 {
    p.Content = template.HTML(p.RawContent[:i] + "...")
    return p.Content
  }
 }
 return p.Content
}

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

func APIPage(w http.ResponseWriter, r *http.Request){
 vars := mux.Vars(r)
 pageGUID := vars["post_guid"]
 thisPage := Page{}
 fmt.Println(pageGUID)
 err := database.QueryRow("select post_title, post_content, post_date from posts where post_guid = $1", pageGUID).
   Scan(&thisPage.Title, &thisPage.RawContent, &thisPage.Date)
 if err != nil {
  http.Error(w, http.StatusText(404), http.StatusNotFound)
  log.Println(err)
  return
 }
 APIOutput, err2 := json.Marshal(thisPage)
 fmt.Println(APIOutput) 
 if err2 != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  fmt.Println("Error in APIPAGE - json.Marshal")
  return
 }
 w.Header().Set("Content-Type", "application/json")
 fmt.Fprintln(w, thisPage)
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

 routes.HandleFunc("/api/pages", APIPage).
  Methods("GET")

 routes.HandleFunc("/api/pages/{post_guid:[0-9a-zA-Z-]+}", APIPage).
  Methods("GET")

 http.Handle("/", routes)

 /*
               	certificates, err1 := tls.LoadX509KeyPair("certificate.pem","key.pem")
 		if err1 != nil {
  			fmt.Println("ERROR loading x509 key pair")
 		}
 		tlsConf := tls.Config{Certificates:[]tls.Certificate{certificates}}
 		_, err2 := tls.Listen("tcp", ":443", &tlsConf)
 		if err2 != nil{
  			fmt.Println("Error listening!")
  			fmt.Println(err2.Error())
 		}
 		//defer ln.Close()
 */
 http.ListenAndServe(PORT, nil)
}

