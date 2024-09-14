package middleware

import (
    "github.com/dgrijalva/jwt-go"
    "net/http"
    "strings"
    "time"
    "context"
	// "mygoshare/database"
)

var jwtKey = []byte("my_secret_key") // Secret key for signing the JWT

type Claims struct {
    Email string `json:"email"`
    jwt.StandardClaims
}

// GenerateJWT creates a new JWT token
func GenerateJWT(email string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)

    claims := &Claims{
        Email: email,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        return "", err
    }
    return tokenString, nil
}

// ValidateJWT validates the given JWT token
func ValidateJWT(tokenString string) (*Claims, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })

    if err != nil || !token.Valid {
        return nil, err
    }

    return claims, nil
}


// JWTAuthMiddleware is a middleware that checks for JWT authentication
func JWTAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
            return
        }

        tokenString := strings.Split(authHeader, "Bearer ")[1]

        claims, err := ValidateJWT(tokenString)
        if err != nil {
            http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
            return
        }

        r = r.WithContext(context.WithValue(r.Context(), "userEmail", claims.Email))
        next.ServeHTTP(w, r)
    })
}
