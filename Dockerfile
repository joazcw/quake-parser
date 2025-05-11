# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
# These are copied separately to leverage Docker cache for dependencies
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
RUN go mod verify

# Copy the source code into the container
# This will copy main.go, routers.go, and the parser/, reporter/, database/ packages
COPY . .

# Build the Go app
# CGO_ENABLED=0 creates a statically-linked binary
# -ldflags "-w -s" strips debug symbols and information, reducing binary size
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/server .

# Stage 2: Create the final, small image
FROM alpine:latest

# The /app directory will be the working directory
WORKDIR /app

# Copy the Pre-built binary file from the builder stage
COPY --from=builder /app/server /app/server

# Copy the data directory which contains games.log
# If you prefer to mount this data as a volume, you can remove this line.
# This is useful if your API's /games/upload endpoint is the primary way to get data.
COPY data ./data

# Expose port 8080 to the outside world (the port the Gin server runs on)
EXPOSE 8080

# Command to run the executable
# Since main.go now only runs in API mode, the "-api-only" flag is no longer strictly necessary here.
# The MongoDB instance is expected to be running and accessible.
# You might need to configure the MongoDB URI in your app (e.g., via env vars) 
# if localhost:27017 (the default in your database.go) is not appropriate for your Docker setup.
CMD ["/app/server"] 