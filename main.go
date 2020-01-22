package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kataras/go-sessions"
	// "os"
)

var db *sql.DB
var err error

type user struct {
	ID       int
	Username string
	Password string
}

type article struct {
	ID        int
	Title     string
	Post      string
	IsPublish string
}

type message struct {
	ID   int
	Post string
}

func connect_db() {
	db, err = sql.Open("mysql", "root:delete21@tcp(127.0.0.1)/GolangProject")

	if err != nil {
		log.Fatalln(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
}

func routes() {
	http.HandleFunc("/", Home)
	http.HandleFunc("/addArticle", Add)
	http.HandleFunc("/allArticle", ShowArticle)
	http.HandleFunc("/about", About)
	http.HandleFunc("/contactUs", ContactUser)
	http.HandleFunc("/contactUsAdmin", ContactAdmin)
	http.HandleFunc("/edit", Edit)
	http.HandleFunc("/delete", Delete)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/logout", Logout)
}

func main() {
	connect_db()
	routes()

	defer db.Close()

	fmt.Println("Server running on port :8000")
	http.ListenAndServe(":8000", nil)
}

func checkErr(w http.ResponseWriter, r *http.Request, err error) bool {
	if err != nil {

		fmt.Println(r.Host + r.URL.Path)

		http.Redirect(w, r, r.Host+r.URL.Path, 301)
		return false
	}

	return true
}

func QueryUser(username string) user {
	var users = user{}
	err = db.QueryRow(`
		SELECT id, username, password 
		FROM Users 
		WHERE username=?
		`, username).
		Scan(
			&users.ID,
			&users.Username,
			&users.Password,
		)
	return users
}

func QueryGetArticleAdmin() []article {
	var oneArticle = article{}
	var allArticle = []article{}
	query, err := db.Query(`
	SELECT * FROM Articles
	ORDER BY id DESC`)
	if err != nil {
		fmt.Println(err.Error())
	}
	for query.Next() {
		var id int
		var title string
		var post string
		var isPublish string
		err = query.Scan(&id, &title, &post, &isPublish)
		if err != nil {
			fmt.Println(err.Error())
		}
		oneArticle.ID = id
		oneArticle.Title = title
		oneArticle.Post = post
		oneArticle.IsPublish = isPublish
		allArticle = append(allArticle, oneArticle)
	}
	return allArticle
}

func QueryGetArticleHome() []article {
	var oneArticle = article{}
	var allArticle = []article{}
	query, err := db.Query(`
	SELECT * FROM Articles
	WHERE isPublish="Yes"
	ORDER BY id DESC`)
	if err != nil {
		fmt.Println(err.Error())
	}
	for query.Next() {
		var id int
		var title string
		var post string
		var isPublish string
		err = query.Scan(&id, &title, &post, &isPublish)
		if err != nil {
			fmt.Println(err.Error())
		}
		oneArticle.ID = id
		oneArticle.Title = title
		oneArticle.Post = post
		oneArticle.IsPublish = isPublish
		allArticle = append(allArticle, oneArticle)
	}
	return allArticle
}

func QueryGetOneArticle(nid string) article {
	var oneArticle = article{}
	query, err := db.Query(`
	SELECT * FROM Articles
	WHERE id = ?`, nid)
	if err != nil {
		fmt.Println(err.Error())
	}
	for query.Next() {
		var id int
		var title string
		var post string
		var isPublish string
		err = query.Scan(&id, &title, &post, &isPublish)
		if err != nil {
			fmt.Println(err.Error())
		}
		oneArticle.ID = id
		oneArticle.Title = title
		oneArticle.Post = post
		oneArticle.IsPublish = isPublish
	}
	return oneArticle
}

func QueryGetMessage() []message {
	var oneMessage = message{}
	var allMessage = []message{}
	query, err := db.Query(`
	SELECT * FROM Messages`)
	if err != nil {
		fmt.Println(err.Error())
	}
	for query.Next() {
		var id int
		var post string
		err = query.Scan(&id, &post)
		if err != nil {
			fmt.Println(err.Error())
		}
		oneMessage.ID = id
		oneMessage.Post = post
		allMessage = append(allMessage, oneMessage)
	}
	return allMessage
}

func Home(w http.ResponseWriter, r *http.Request) {
	allArticles := QueryGetArticleHome()
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		var t, err = template.ParseFiles("views/homeBeforeLogin.html")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		t.Execute(w, allArticles)
		return
	} else {
		var t, err = template.ParseFiles("views/home.html")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		t.Execute(w, allArticles)
		return
	}
}

func ShowArticle(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/login", 301)
	}
	allArticles := QueryGetArticleAdmin()
	var t, err = template.ParseFiles("views/article.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	t.Execute(w, allArticles)
	return
}

func About(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		var t, err = template.ParseFiles("views/aboutBeforeLogin.html")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		t.Execute(w, nil)
		return
	} else {
		if len(session.GetString("username")) == 0 {
			http.Redirect(w, r, "/login", 301)
		}
		var t, err = template.ParseFiles("views/about.html")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		t.Execute(w, nil)
		return
	}
}

func Add(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/login", 301)
	}
	if r.Method != "POST" {
		http.ServeFile(w, r, "views/add.html")
	}
	if r.Method == "POST" {
		title := r.FormValue("title")
		post := r.FormValue("post")
		addForm, err := db.Prepare(`
	INSERT INTO Articles (title, post) VALUES (?,?)`)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		addForm.Exec(title, post)
		allArticles := QueryGetArticleAdmin()
		session := sessions.Start(w, r)
		if len(session.GetString("username")) == 0 {
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			http.ServeFile(w, r, "views/login.html")
			return
		} else {
			var t, err = template.ParseFiles("views/article.html")
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			t.Execute(w, allArticles)
			return
		}
	}
}

func Contact(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if r.Method != "POST" && len(session.GetString("username")) == 0 {
		http.ServeFile(w, r, "views/contactBeforeLogin.html")
	} else if r.Method != "POST" && len(session.GetString("username")) == 1 {
		http.ServeFile(w, r, "views/contact.html")
	} else if r.Method == "POST" && len(session.GetString("username")) == 0 {
		post := r.FormValue("post")
		addForm, err := db.Prepare(`
	INSERT INTO Messages (post) VALUES (?)`)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		addForm.Exec(post)
		http.Redirect(w, r, "/", 302)
	} else if r.Method == "POST" && len(session.GetString("username")) == 1 {
		post := r.FormValue("post")
		addForm, err := db.Prepare(`
	INSERT INTO Messages (post) VALUES (?)`)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		addForm.Exec(post)
		http.Redirect(w, r, "/", 302)
	}
}

func ContactAdmin(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/login", 301)
	}
	if r.Method != "POST" {
		http.ServeFile(w, r, "views/contact.html")
	}
	if r.Method == "POST" {
		post := r.FormValue("post")
		addForm, err := db.Prepare(`
	INSERT INTO Messages (post) VALUES (?)`)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		addForm.Exec(post)
		http.Redirect(w, r, "/", 302)
	}
}

func ContactUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "views/contactBeforeLogin.html")
	}
	if r.Method == "POST" {
		post := r.FormValue("post")
		addForm, err := db.Prepare(`
	INSERT INTO Messages (post) VALUES (?)`)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		addForm.Exec(post)
		http.Redirect(w, r, "/", 302)
	}
}

func Edit(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/login", 301)
	}
	if r.Method != "POST" {
		http.ServeFile(w, r, "views/edit.html")
		return
	}
	if r.Method == "POST" {
		nid := r.FormValue("id")
		var oneArticle = QueryGetOneArticle(nid)
		var title string
		var post string
		var isPublish string
		if r.FormValue("title") == "" {
			title = oneArticle.Title
		} else {
			title = r.FormValue("title")
		}
		if r.FormValue("title") == "" {
			post = oneArticle.Post
		} else {
			post = r.FormValue("post")
		}
		if r.FormValue("isPublish") == "" {
			isPublish = oneArticle.IsPublish
		} else {
			isPublish = r.FormValue("isPublish")
		}
		id := r.FormValue("id")
		editForm, err := db.Prepare(`
	UPDATE Articles SET title=?, post=?, isPublish=?
	WHERE id=?`)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		editForm.Exec(title, post, isPublish, id)
		allArticles := QueryGetArticleAdmin()
		session := sessions.Start(w, r)
		if len(session.GetString("username")) == 0 {
			var t, err = template.ParseFiles("views/login.html")
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			t.Execute(w, allArticles)
			return
		} else {
			var t, err = template.ParseFiles("views/article.html")
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			t.Execute(w, allArticles)
			return
		}
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) != 0 && checkErr(w, r, err) {
		http.Redirect(w, r, "/", 302)
	}
	if r.Method != "POST" {
		http.ServeFile(w, r, "views/login.html")
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	users := QueryUser(username)
	var password_tes = bcrypt.CompareHashAndPassword([]byte(users.Password), []byte(password))
	if password_tes == nil {
		session := sessions.Start(w, r)
		session.Set("username", users.Username)
		http.Redirect(w, r, "/", 302)
	} else {
		http.Redirect(w, r, "/login", 302)
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/login", 301)
	}
	oneArticle := r.URL.Query().Get("id")
	deleteForm, err := db.Prepare(`
	DELETE FROM Articles WHERE id = ?`)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	deleteForm.Exec(oneArticle)
	allArticles := QueryGetArticleAdmin()
	if len(session.GetString("username")) == 0 {
		var t, err = template.ParseFiles("views/login.html")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		t.Execute(w, allArticles)
		return
	} else {
		var t, err = template.ParseFiles("views/article.html")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		t.Execute(w, allArticles)
		return
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	session.Clear()
	sessions.Destroy(w, r)
	http.Redirect(w, r, "/", 302)
}
