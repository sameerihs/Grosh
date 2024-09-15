package main

import (
    "log"
    "net/http"
    "fmt"
    "mygoshare/handlers"
    "mygoshare/middleware"
    "mygoshare/database"
    "encoding/json"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/joho/godotenv"
    "os"
)

var s3Client *s3.S3

func init() {
	// this is for loading variables from .env file
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }

    // this fetches aws s3 values from .env 
    accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
    secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
    region := os.Getenv("AWS_REGION")

    sess, err := session.NewSession(&aws.Config{
        Region:      aws.String(region),
        Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
    })
    if err != nil {
        log.Fatal("Failed to create session:", err)
    }

    s3Client = s3.New(sess)
}

func main() {
    database.InitCache()
    database.InitDB()
    http.HandleFunc("/register", handlers.RegisterHandler)
    http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/search/files", middleware.JWTAuthMiddleware(handlers.SearchFiles))

    // these are protect routed with jwt middleware 
    http.HandleFunc("/upload", middleware.JWTAuthMiddleware(uploadHandler))
    http.HandleFunc("/files", middleware.JWTAuthMiddleware(handlers.GetFilesHandler))

    log.Println("Server started at :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    // the request contains user's email, which we use to get user id
    userEmail, ok := r.Context().Value("userEmail").(string)
    if !ok {
        http.Error(w, "User email not found", http.StatusUnauthorized)
        return
    }

    // im making sure that the file is less than 10mb
    err := r.ParseMultipartForm(10 << 20)
    if err != nil {
        http.Error(w, "Unable to parse form", http.StatusBadRequest)
        return
    }

    file, fileHeader, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Unable to get file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // this is for uploading it to S3 bucket
    fileName := fileHeader.Filename
    filePath := fmt.Sprintf("uploads/%s", fileName)
    _, err = s3Client.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(os.Getenv("S3_BUCKET_NAME")),
        Key:    aws.String(filePath),
        Body:   file,
        ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
    })
    if err != nil {
        log.Println(err)
        http.Error(w, "Unable to upload file", http.StatusInternalServerError)
        return
    }

    // Generate public URL
    fileURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", os.Getenv("S3_BUCKET_NAME"), filePath)

    // Fetch user ID from database
    var userID int
    err = database.DB.QueryRow("SELECT id FROM users WHERE email = $1", userEmail).Scan(&userID)
    if err != nil {
        http.Error(w, "User ID not found", http.StatusInternalServerError)
        return
    }

    // Get file size (assuming you have a way to determine this)
    fileSize := fileHeader.Size

    // Insert file metadata into database
    err = database.InsertFileMetadata(userID, fileName, int(fileSize), fileURL)
    if err != nil {
        http.Error(w, "Failed to store file metadata", http.StatusInternalServerError)
        return
    }

    // Return file URL
    response := map[string]string{"fileUrl": fileURL}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
