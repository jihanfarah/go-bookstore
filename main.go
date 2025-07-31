package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "github.com/gorilla/mux"
    "database/sql"
    "log"
    "strconv"

    _ "github.com/go-sql-driver/mysql"
)

type Book struct {
    ID int `json:"id"`
    Title string `json:"title"`
    Author string  `json:"author"`
    Price float64 `json:"price"`
}


var db *sql.DB

func initDB() {
    var err error
    dsn := "root:123456789@tcp(127.0.0.1:3306)/bookstore"
    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal("Failed to open DB:", err)
    }
    if err = db.Ping(); err != nil {
        log.Fatal("Failed to connect to open DB:", err)
    }

    if err = db.Ping(); err != nil {
        log.Fatal("Failed to connect to DB:", err)
    }

fmt.Println("Connected to MySQL successfully!")
}

func main() {
    http.HandleFunc("/greet", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "📚 Welcome to the Go Bookstore!")
    })

    initDB()
    router:= mux.NewRouter()

    router.HandleFunc("/ping", pingHandler).Methods("GET")
    router.HandleFunc("/books", getAllBooks).Methods("GET")
    router.HandleFunc("/books", createBook).Methods("POST")
    router.HandleFunc("/books/{id}", getBookByID).Methods("GET")
    router.HandleFunc("/books/{id}", updateBook).Methods("PUT")
    router.HandleFunc("/books/{id}", deleteBook).Methods("DELETE")


    fmt.Println("Server is running on port 8080...")
    http.ListenAndServe(":8080", enableCorsMiddleware(router))
}

func pingHandler(w http.ResponseWriter, r *http.Request){
        fmt.Fprintf(w, "👋 Hello from Go!")
}

func enableCorsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}


func getAllBooks(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, title, author, price FROM books")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var books []Book
    for rows.Next() {
        var b Book
        if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Price); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        books = append(books, b)
    }

    json.NewEncoder(w).Encode(books)
}


func getBookByID(w http.ResponseWriter, r *http.Request){
     w.Header().Set("Content-Type", "application/json")
     params := mux.Vars(r)

     var book Book
     err := db.QueryRow("SELECT id, title, author, price FROM books WHERE id = ?", params["id"]).Scan(&book.ID, &book.Title, &book.Author, &book.Price)
     if err != nil {
        http.Error(w, "Book not found", http.StatusNotFound)
        return
     }

    json.NewEncoder(w).Encode(book)
}

func createBook(w http.ResponseWriter, r *http.Request){
    w.Header().Set("Content-Type", "application/json")

    var newBook Book

    err := json.NewDecoder(r.Body).Decode(&newBook)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    result, err := db.Exec("INSERT INTO books (title, author, price) VALUES (?, ?, ?)", newBook.Title, newBook.Author, newBook.Price)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    lastID, err := result.LastInsertId()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    newBook.ID = int(lastID)

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(newBook)
}

func updateBook(w http.ResponseWriter, r *http.Request){
    w.Header().Set("Content-Type", "application/json")

    params := mux.Vars(r)

    var updateBook Book
    err := json.NewDecoder(r.Body).Decode(&updateBook)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    result, err := db.Exec("UPDATE books SET title=?, author=?, price=? WHERE id=?", updateBook.Title, updateBook.Author, updateBook.Price, params["id"])
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected, err := result.RowsAffected()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        if rowsAffected == 0 {
            http.Error(w, "Book not found", http.StatusNotFound)
            return
        }

    updateBook.ID, _ = strconv.Atoi(params["id"])
    json.NewEncoder(w).Encode(updateBook)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    params := mux.Vars(r)

    _, err := db.Exec("DELETE FROM books WHERE id = ?", params["id"])
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{"message": "Book deleted"})
}
