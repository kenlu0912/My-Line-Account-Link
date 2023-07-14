package main

import (
	b64 "encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"database/sql"
	_ "github.com/lib/pq"
)

const (
	host     = "dpg-cinbnoh8g3nafl536o5g-a.oregon-postgres.render.com"
	port     = 5432
	user     = "userlist_user"
	password = "8wLaPwtCbV6h5O4qIEJRGVoI5PgeZgjz"
	dbname   = "user"
)

// CustData : Customers data for provider website.
type CustData struct {
	ID    string
	PW    string
	Name  string
	Age   int
	Desc  string
	Nonce string
}

type UserData struct {
	id       int
	username string
	password string
	name     string
	age      int
	descri   string
	nonce    string
}

var customers []CustData
var users     []UserData

func init() {
	//Init customer data in memory
	customers = append(customers, []CustData{
		CustData{ID: "11", PW: "pw11", Name: "Tom", Age: 43, Desc: "He is from A corp. likes to read comic books."},
		CustData{ID: "22", PW: "pw22", Name: "John", Age: 25, Desc: "He is from B corp. likes to read news paper"},
		CustData{ID: "44", PW: "pw44", Name: "Mary", Age: 13, Desc: "She is a student, like to read science books"},
	}...)

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)
  
	// open database
	db, err := sql.Open("postgres", psqlconn)
	checkErr(err)

	//查詢資料
	rows, err := db.Query("SELECT * FROM test")
	checkErr(err)

	for rows.Next() {
		var SQLid int
		var SQLusername string
		var SQLpassword string
		var SQLname string
		var SQLage int
		var SQLdescri string
		var SQLnonce string
		err = rows.Scan(&SQLid, &SQLusername, &SQLpassword, &SQLname, &SQLage, &SQLdescri, &SQLnonce)
		checkErr(err)
		fmt.Println(SQLid)
		fmt.Println(SQLusername)
		fmt.Println(SQLpassword)
		fmt.Println(SQLname)
		fmt.Println(SQLage)
		fmt.Println(SQLdescri)
		fmt.Println(SQLnonce)

		users = append(users, []UserData{
			UserData{id: SQLid, username: SQLusername, password: SQLpassword, name: SQLname, age: SQLage, descri: SQLdescri, nonce: SQLnonce},
		}...)
	}

	db.Close()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// WEB: List all user in memory
func listCust(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Bookstore customer list as follow:\n")
	// for i, usr := range customers {
	// 	fmt.Fprintf(w, "%d \tID: %s \tName: %s \tPW: %s \tDesc:%s \n", i, usr.ID, usr.Name, usr.PW, usr.Desc)
	// }
	for i, usr := range users {
		fmt.Fprintf(w, "%d \tid: %d \tusername: %s \tpassword: %s \tname:%s \tage:%d \tdescri:%s \tnonce:%s \n", i, usr.id, usr.username, usr.password, usr.name, usr.age, usr.descri, usr.nonce)
	}
}

// WEB: For login (just for demo)
func login(w http.ResponseWriter, r *http.Request) {
	//7. The user enters his/her credentials.
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm() err: %v\n", err)
		return
	}
	name := r.FormValue("user")
	pw := r.FormValue("pass")
	token := r.FormValue("token")
	// for i, usr := range customers {
	// 	if usr.ID == name {
	// 		if pw == usr.PW {
	// 			//8. The web server acquires the user ID from the provider's service and uses that to generate a nonce.
	// 			sNonce := generateNonce(token, name, pw)

	// 			//update nonce to provider DB to store it.
	// 			customers[i].Nonce = sNonce

	// 			//9. The web server redirects the user to the account-linking endpoint.
	// 			//10. The user accesses the account-linking endpoint.
	// 			//Print link to user to click it.
	// 			targetURL := fmt.Sprintf("https://access.line.me/dialog/bot/accountLink?linkToken=%s&nonce=%s", token, sNonce)
	// 			log.Println("generate nonce, targetURL=", targetURL)
	// 			tmpl := template.Must(template.ParseFiles("link.tmpl"))
	// 			if err := tmpl.Execute(w, targetURL); err != nil {
	// 				log.Println("Template err:", err)
	// 			}
	// 			return
	// 		}
	// 	}
	// }
	for i, usr := range users {
		if usr.username == name {
			if pw == usr.password {
				//8. The web server acquires the user ID from the provider's service and uses that to generate a nonce.
				sNonce := generateNonce(token, name, pw)

				//update nonce to provider DB to store it.
				users[i].nonce = sNonce

				psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)
         
        // open database
				db, err := sql.Open("postgres", psqlconn)
				checkErr(err)

				psqlUpdate := fmt.Sprintf("UPDATE test SET nonce = '%s' WHERE username = '%s';", users[i].nonce, users[i].username)
				rows, err := db.Query(psqlUpdate)
				// rows, err := db.Query("UPDATE test SET nonce = '456' WHERE username = 'Joe';")
				_ = rows
   		  checkErr(err)

				db.Close()

				//9. The web server redirects the user to the account-linking endpoint.
				//10. The user accesses the account-linking endpoint.
				//Print link to user to click it.
				targetURL := fmt.Sprintf("https://access.line.me/dialog/bot/accountLink?linkToken=%s&nonce=%s", token, sNonce)
				log.Println("generate nonce, targetURL=", targetURL)
				tmpl := template.Must(template.ParseFiles("link.tmpl"))
				if err := tmpl.Execute(w, targetURL); err != nil {
					log.Println("Template err:", err)
				}
				return
			}
		}
	}
	fmt.Fprintf(w, "您輸入的帳號或密碼錯誤，請再輸入一遍")
}

// WEB: For account link
func link(w http.ResponseWriter, r *http.Request) {
	//5. The user accesses the linking URL.
	TOKEN := r.FormValue("linkToken")
	if TOKEN == "" {
		log.Println("No token.")
		return
	}

	log.Println("token = ", TOKEN)
	tmpl := template.Must(template.ParseFiles("login.tmpl"))
	//6. The web server displays the login screen.
	if err := tmpl.Execute(w, TOKEN); err != nil {
		log.Println("Template err:", err)
	}
}

// generate nonce (currently nonce combine by token + name + pw)
func generateNonce(token, name, pw string) string {
	return b64.StdEncoding.EncodeToString([]byte(token + name + pw))
}
