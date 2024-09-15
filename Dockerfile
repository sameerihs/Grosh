# Step 1: Use a Go version compatible with your project
FROM golang:1.23

# Step 2: Set the working directory
WORKDIR /app

# Step 3: Copy the Go module files and install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Step 4: Copy the source code to the container
COPY . .

# Step 5: Build the Go app
RUN go build -o mygoshare .

# Step 6: Expose the port that the app runs on
EXPOSE 8080

# Step 7: Run the Go app
CMD ["./mygoshare"]
