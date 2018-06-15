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
 //"time"
)

const (
 PORT = ":8080"
)

type Page struct {
 ID      int
 Title   string
 RawContent string
 Content template.HTML
 Date    string
}

type PageJSON struct {
 ID int
 Title string
 Content string
 Date string
}

type JSONResponse struct {
 Fields map[string]string
}

var database *sql.DB


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

///////////////////// ServePage /////////////////////

func ServePage(w http.ResponseWriter, r *http.Request){
 vars := mux.Vars(r)
 pageID := vars["id"]
 
 p := Page{}

 row := database.QueryRow("select id, post_title, post_content, post_date from posts2 where id=$1", pageID)

 err := row.Scan(&p.ID, &p.Title, &p.RawContent,&p.Date)
 if err != nil {
  http.Error(w, http.StatusText(404), http.StatusNotFound)
  log.Println("Error: couldn't get data!")
 }else{ 
  p.Content = template.HTML(p.RawContent)
  t, _:= template.ParseFiles("templates/blog.html")
  t.Execute(w, p)
 }
}

///////////////////// RedirIndex /////////////////////

func RedirIndex(w http.ResponseWriter, r *http.Request){
 http.Redirect(w,r,"/home", 301)
}

///////////////////// ServeIndex /////////////////////

func ServeIndex(w http.ResponseWriter, r *http.Request){

 var Pages = [] Page{}
 var pages *sql.Rows

 // Check if we have more than 7 posts
 res := database.QueryRow("SELECT COUNT(*) FROM posts2")
 
 //var count int64
 var count int64
 err2 := res.Scan(&count)
 if err2 != nil {
   fmt.Println("ServeIndex: ERROR scan into count")
 }
 //fmt.Print("ServeIndex: we have a count "); fmt.Println(count)

 if count == 0 {
   //fmt.Println("ServeIndex: we have a count = 0")
   http.ServeFile(w, r, "no_posts.html")
   return
 }

 // Show last 7 posts
 if count > 7 {
   pages_t, err1 := database.Query("SELECT id, post_title,post_content,post_date FROM posts2 limit 7 OFFSET ($1 - 7)", count)
   if err1 != nil {
    fmt.Println("ServeIndex: ERROR in database.Query for count > 7")
    pages_t.Close()
    return
   }
   pages = pages_t
 }else{
   pages_t, err1 := database.Query("SELECT id, post_title,post_content,post_date FROM posts2")
   if err1 != nil {
    fmt.Println("ServeIndex: ERROR in database.Query for count < 7")
    pages_t.Close()
    return
   }
   pages = pages_t
 }
 defer pages.Close()

 for pages.Next() {
  thisPage := Page{}
  pages.Scan(&thisPage.ID, &thisPage.Title, &thisPage.RawContent, &thisPage.Date)
  thisPage.Content = template.HTML(thisPage.RawContent)
  Pages = append(Pages, thisPage) 
 }
 t, err3 := template.ParseFiles("templates/index.html")
 if err3 != nil {
   fmt.Println("ServeIndex: ERROR parsing template")
 }
 t.Execute(w, Pages)  
} 

///////////////////// APIPage /////////////////////

func APIPage(w http.ResponseWriter, r *http.Request){
 vars := mux.Vars(r)
 pageID := vars["id"]
 p := Page{}
 err := database.QueryRow("select post_title, post_content, post_date from posts2 where id = $1", pageID).
   Scan(&p.Title, &p.RawContent, &p.Date)
 if err != nil {
  http.Error(w, http.StatusText(404), http.StatusNotFound)
  log.Println(err)
  return
 }
 
 // Create proper JSON before marshaling
 pID, _ := strconv.Atoi(pageID)
 pj := &PageJSON{ID:pID,Title:p.Title,Content:p.RawContent,Date:p.Date}
 
 APIOutput, err2 := json.Marshal(pj)

 if err2 != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  fmt.Println("Error in APIPAGE - json.Marshal")
  return
 }
 w.Header().Set("Content-Type", "application/json")
 fmt.Fprintln(w, string(APIOutput)) //fmt.Fprintln(w, p)
}

///////////////////// APost /////////////////////

type PostJSON struct {
 Title string
 Content string
 Date string
}

func APost(w http.ResponseWriter, r *http.Request){
 //fmt.Println("BEGIN APost")

if r.URL.Path != "/post" {
   http.Error(w, "404 not found", http.StatusNotFound)
   return
 }
 switch r.Method {
  case "GET":
    //fmt.Println("Get GET branch...")
    http.ServeFile(w, r, "post.html")
  case "POST":
    //fmt.Println("Get POST branch...")
    if err := r.ParseForm(); err != nil {
       fmt.Fprintf(w, "ParseForm() err: %v", err)
    }
    cont := r.FormValue("RawContent")
    title := r.FormValue("Title")
    //fmt.Println("RawContent ", cont)
    //fmt.Println("Title ", title)
    _, err1 := database.Exec("INSERT into posts2 (post_title, post_content) values ($1,$2)",title, cont)
    if err1 != nil {
       fmt.Println("ERROR in APIPOst - database.Exec")
       log.Println(err1.Error)
       // Load err_post.html
       http.ServeFile(w, r, "err_post.html")
       
    }else{
       http.ServeFile(w, r, "post.html")
    }
 /*
     pj := &PostJSON{GUID:guid,Title:title,Content:cont,Date:d}
     APIOutput, err2 := json.Marshal(pj)

     if err2 != nil {
        http.Error(w, err2.Error(), http.StatusInternalServerError)
        fmt.Println("Error in APIPAGE - json.Marshal")
        return
     }
     w.Header().Set("Content-Type", "application/json")
     fmt.Fprintln(w, string(APIOutput)) //fmt.Fprintln(w, p)
*/ 
 default:
    fmt.Fprintf(w, "Only GET and POST are supported")
 }
}

///////////////////// APIPost //////////////////

type shortPost struct {
 Title string
 RawContent string
}

func APIPost(w http.ResponseWriter, r *http.Request){
  //fmt.Println("APIPost begins")
  if err := r.ParseForm(); err != nil {
     fmt.Fprintf(w, http.StatusText(http.StatusInternalServerError))
     //fmt.Println("Error in ParseForm")
     return
  }
  var p shortPost
  for key, _ := range r.Form {

    err := json.Unmarshal([]byte(key), &p); if err != nil {
     //fmt.Println("Error in UNmarshal")
     fmt.Fprintf(w, http.StatusText(http.StatusInternalServerError))
     return
    }
  }

  cont := p.RawContent
  //fmt.Print("Got cont "); fmt.Println(cont)
  title := p.Title
  //fmt.Print("Got title "); fmt.Println(title)

  _, err := database.Exec("INSERT into posts2 (post_title, post_content) values ($1,$2)",title, cont)
  if err != nil {
    fmt.Fprintf(w, http.StatusText(http.StatusInternalServerError))
    //fmt.Println("Error in Exec")
    return
  }
  // Return to sender
  fmt.Fprintf(w, http.StatusText(http.StatusOK))
  return
}

///////////////////// APIPostDelete /////////////////////

func APIPostDelete(w http.ResponseWriter, r *http.Request){
 //fmt.Println("APIPostDelete starts...")
 vars := mux.Vars(r)
 pageID := vars["id"]
 
 _, err := database.Exec("DELETE FROM posts2 WHERE id=$1",pageID)
 if err != nil {
    fmt.Fprintf(w, http.StatusText(http.StatusNotFound))
    fmt.Println("APIPostDelete: Error in Exec - not found")
    return
 }
 // Return to sender
 fmt.Fprintf(w, http.StatusText(http.StatusOK))
 return
} 

///////////////////// HelpFunc /////////////////
func HelpFunc(w http.ResponseWriter, r *http.Request){
  http.ServeFile(w, r, "help.html")
  return
}
///////////////////// MAIN /////////////////////

func main(){

 db , err := sql.Open("sqlite3", "cms.db")
 if err != nil {
   fmt.Println("Cannot connect!")
   log.Println("Couldn't connect!")
   log.Println(err.Error)
 } else {
  fmt.Println("Connected to SQLITE3 DB")
 }

 database = db

 routes := mux.NewRouter()

 routes.HandleFunc("/posts/{id:[0-9a-zA-Z-]+}", ServePage)

 routes.HandleFunc("/home", ServeIndex)

 routes.HandleFunc("/help", HelpFunc)
 
 routes.HandleFunc("/", RedirIndex)

 routes.HandleFunc("/api/posts", APIPage).
  Methods("GET")

 routes.HandleFunc("/api/posts/{id:[0-9a-zA-Z-]+}", APIPage).
  Methods("GET")

 routes.HandleFunc("/post", APost)

 routes.HandleFunc("/api/post", APIPost)

 routes.HandleFunc("/api/post-delete/{id:[0-9a-zA-Z-]+}", APIPostDelete).
  Methods("DELETE")

 http.Handle("/", routes)


 http.ListenAndServe(PORT, nil)
}
