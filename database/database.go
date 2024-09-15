package database

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/go-redis/redis/v8"
    _ "github.com/lib/pq" // PostgreSQL driver
)

var (
    DB          *sql.DB
    redisClient *redis.Client
    ctx         = context.Background()
)

func InitDB() {
    var err error

    connStr := "host=localhost port=5432 user=postgres password=password dbname=MyGoShare sslmode=disable"
    DB, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Failed to open database:", err)
    }

    err = DB.Ping()
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    fmt.Println("Connected to PostgreSQL successfully")
}

func InitCache() {
    redisClient = redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    _, err := redisClient.Ping(ctx).Result()
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }
    fmt.Println("Connected to Redis successfully")
}

func InsertFileMetadata(userID int, fileName string, fileSize int, fileURL string) error {
    query := `INSERT INTO files (file_name, file_size, file_url, upload_date, user_id) VALUES ($1, $2, $3, NOW(), $4)`
    _, err := DB.Exec(query, fileName, fileSize, fileURL, userID)
    if err != nil {
        return fmt.Errorf("error inserting file metadata: %w", err)
    }

    // Cache the file metadata
    fileID := getLastInsertID()
    metadata := FileMetadata{
        FileName:   fileName,
        FileSize:   fileSize,
        FileURL:    fileURL,
        UploadDate: time.Now(),
    }
    err = CacheFileMetadata(fileID, metadata)
    if err != nil {
        log.Printf("Failed to cache file metadata: %v", err)
    }

    return nil
}

func getLastInsertID() int {
    var lastID int
    err := DB.QueryRow("SELECT LASTVAL()").Scan(&lastID)
    if err != nil {
        log.Fatalf("Failed to get last insert ID: %v", err)
    }
    return lastID
}

func CacheFileMetadata(fileID int, metadata FileMetadata) error {
    key := fmt.Sprintf("file_metadata:%d", fileID)
    data, err := json.Marshal(metadata)
    if err != nil {
        return fmt.Errorf("failed to marshal metadata: %w", err)
    }

    err = redisClient.Set(ctx, key, data, time.Minute*10).Err()
    if err != nil {
        return fmt.Errorf("failed to cache metadata: %w", err)
    }
    return nil
}

func GetCachedFileMetadata(fileID int) (*FileMetadata, error) {
    key := fmt.Sprintf("file_metadata:%d", fileID)
    data, err := redisClient.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, nil // Cache miss
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get cache: %w", err)
    }

    var metadata FileMetadata
    err = json.Unmarshal([]byte(data), &metadata)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
    }
    return &metadata, nil
}

func GetFileMetadata(fileID int) (string, int, string, error) {
    // Try to get metadata from cache
    metadata, err := GetCachedFileMetadata(fileID)
    if err != nil {
        return "", 0, "", err
    }
    if metadata != nil {
        return metadata.FileName, metadata.FileSize, metadata.FileURL, nil
    }

    // Fallback to database
    if DB == nil {
        return "", 0, "", fmt.Errorf("database not initialized")
    }

    var fileName string
    var fileSize int
    var fileURL string

    query := `SELECT file_name, file_size, file_url FROM files WHERE id = $1`
    row := DB.QueryRow(query, fileID)
    err = row.Scan(&fileName, &fileSize, &fileURL)
    if err != nil {
        if err == sql.ErrNoRows {
            return "", 0, "", fmt.Errorf("file not found")
        }
        return "", 0, "", fmt.Errorf("failed to retrieve file metadata: %v", err)
    }

    // Cache file metadata
    metadata = &FileMetadata{
        FileName: fileName,
        FileSize: fileSize,
        FileURL:  fileURL,
    }
    err = CacheFileMetadata(fileID, *metadata)
    if err != nil {
        log.Printf("Failed to cache file metadata: %v", err)
    }

    return fileName, fileSize, fileURL, nil
}

func GetFilesByUser(userID int) ([]FileMetadata, error) {
    if DB == nil {
        return nil, fmt.Errorf("database not initialized")
    }

    query := `SELECT id, file_name, file_size, file_url, upload_date FROM files WHERE user_id = $1`
    rows, err := DB.Query(query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve files: %w", err)
    }
    defer rows.Close()

    var files []FileMetadata
    for rows.Next() {
        var file FileMetadata
        if err := rows.Scan(&file.ID, &file.FileName, &file.FileSize, &file.FileURL, &file.UploadDate); err != nil {
            return nil, fmt.Errorf("failed to scan file row: %w", err)
        }
        files = append(files, file)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating over rows: %w", err)
    }

    return files, nil
}

// Define the FileMetadata struct
type FileMetadata struct {
    ID         int
    FileName   string
    FileSize   int
    FileURL    string
    UploadDate time.Time
}
