package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
	"log"
    "golang.org/x/crypto/bcrypt"
    "mygoshare/database"
    "mygoshare/middleware" // Ensure this is imported correctly
)

type User struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    if database.DB == nil {
        http.Error(w, "Database not initialized", http.StatusInternalServerError)
        return
    }

    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Error hashing password", http.StatusInternalServerError)
        return
    }

    query := `INSERT INTO users (email, password) VALUES ($1, $2)`
    _, err = database.DB.Exec(query, user.Email, string(hashedPassword))
    if err != nil {
		log.Println(err)
        http.Error(w, "Error creating user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("User registered successfully"))
}



func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Retrieve user from the database
    var hashedPassword string
    query := `SELECT id, password FROM users WHERE email=$1`
    row := database.DB.QueryRow(query, user.Email)

    var userId int
    err = row.Scan(&userId, &hashedPassword)
    if err == sql.ErrNoRows || err != nil {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    // Compare the hashed password
    err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
    if err != nil {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    // Generate JWT token
    token, err := middleware.GenerateJWT(user.Email) // Use the function from middleware package
    if err != nil {
        http.Error(w, "Error generating token", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(token))
}
