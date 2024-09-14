// In handlers/file.go (create this file if it doesn't exist)

package handlers

import (
    "encoding/json"
    "net/http"
    "mygoshare/database"
	"log"
	"time"
	
)

// File represents a file record in the database
type File struct {
    ID          int       `json:"id"`
    UserID      int       `json:"user_id"`
    FileName    string    `json:"file_name"`
    FileURL     string    `json:"file_url"`
    FileSize    int64     `json:"file_size"`
    UploadDate  time.Time `json:"upload_date"`
}


// GetFilesHandler retrieves files for the user identified by email
func GetFilesHandler(w http.ResponseWriter, r *http.Request) {
    // Extract user email from context
    userEmail, ok := r.Context().Value("userEmail").(string)
    if !ok {
        http.Error(w, "User email not found", http.StatusUnauthorized)
        return
    }

    // Fetch user ID from database
    var userID int
    err := database.DB.QueryRow("SELECT id FROM users WHERE email = $1", userEmail).Scan(&userID)
    if err != nil {
        http.Error(w, "User ID not found", http.StatusInternalServerError)
        return
    }

    // Retrieve files for the user
    files, err := database.GetFilesByUser(userID)
    if err != nil {
        http.Error(w, "Failed to retrieve files", http.StatusInternalServerError)
        return
    }

    // Cache file metadata
    for _, file := range files {
        err := database.CacheFileMetadata(file.ID, file)
        if err != nil {
            log.Printf("Failed to cache file metadata: %v", err)
        }
    }

    // Return file list as JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(files)
}


// SearchFiles handles search requests
func SearchFiles(w http.ResponseWriter, r *http.Request) {
    // Extract user email from context (you may need to adjust how you get the user email)
    userEmail, ok := r.Context().Value("userEmail").(string)
    if !ok {
        http.Error(w, "User email not found", http.StatusUnauthorized)
        return
    }

    // Fetch user ID from database
    var userID int
    err := database.DB.QueryRow("SELECT id FROM users WHERE email = $1", userEmail).Scan(&userID)
    if err != nil {
        http.Error(w, "User ID not found", http.StatusInternalServerError)
        return
    }

    // Extract query parameters
    name := r.URL.Query().Get("name")
    date := r.URL.Query().Get("date")
    fileType := r.URL.Query().Get("file_type")

    // Build query
    query := "SELECT id, user_id, file_name, file_url, file_size, upload_date FROM files WHERE user_id = $1"
    var args []interface{}
    args = append(args, userID) // Add userID to the query

    if name != "" {
        query += " AND file_name ILIKE $2"
        args = append(args, "%"+name+"%")
    }
    if date != "" {
        query += " AND upload_date::date = $3"
        args = append(args, date)
    }
    if fileType != "" {
        query += " AND file_url LIKE $4"
        args = append(args, "%."+fileType)
    }

    rows, err := database.DB.Query(query, args...)
    if err != nil {
        http.Error(w, "Error querying database", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var files []File
    for rows.Next() {
        var file File
        if err := rows.Scan(&file.ID, &file.UserID, &file.FileName, &file.FileURL, &file.FileSize, &file.UploadDate); err != nil {
            log.Println(err)
            http.Error(w, "Error scanning row", http.StatusInternalServerError)
            return
        }
        files = append(files, file)
    }

    // Return file list as JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(files)
}
