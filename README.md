# Quake Log Parser API

A robust API for parsing and analyzing Quake 3 Arena game logs, built with Go. This API extracts game statistics, player rankings, and death causes from Quake log files and makes them accessible through a RESTful interface.

## Features

- 📊 Parse Quake 3 Arena log files to extract game statistics
- 🎮 Track player kills, deaths, and other game metrics
- 📈 Generate player rankings across all games
- 🔫 Track kills by weapon/death cause
- 🚀 RESTful API with Swagger documentation
- 💾 MongoDB persistence for game reports
- 🌐 Simple web UI for visualizing data

## Tech Stack

- **Backend**: Go (Gin framework)
- **Database**: MongoDB
- **Documentation**: Swagger/OpenAPI
- **Containerization**: Docker
- **Frontend**: HTML, CSS, JavaScript

## API Endpoints

| Method | Endpoint          | Description                                       |
|--------|-------------------|---------------------------------------------------|
| GET    | /games            | Get all game reports                              |
| GET    | /games/{id}       | Get a single game report by ID                    |
| POST   | /games/upload     | Upload a log file for processing                  |
| DELETE | /games            | Delete all game reports                           |
| DELETE | /games/{id}       | Delete a specific game report                     |
| GET    | /playersranking   | Get aggregated player rankings                    |
| GET    | /swagger/*any     | Swagger UI for API documentation                  |

## Prerequisites

- Go 1.21+ 
- MongoDB 4.4+ 
- Docker 
- Python (simple local web server for the frontend)

## Installation

### Option 1: Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/joazcw/quake_log_parser.git
   cd quake_log_parser
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Make sure MongoDB is running:
   ```bash
   # Default configuration expects MongoDB at mongodb://localhost:27017
   # To start MongoDB with Docker:
   docker run -d -p 27017:27017 --name mongodb mongo:latest
   ```

4. Run the application:
   ```bash
   go run .
   ```

5. Access the API at http://localhost:8080 and the Swagger documentation at http://localhost:8080/swagger/index.html

### Option 2: Docker Deployment

1. Build the Docker image:
   ```bash
   docker build -t quake-log-parser .
   ```

2. Create a Docker network for the app and database:
   ```bash
   docker network create quake-network
   ```

3. Start MongoDB:
   ```bash
   docker run -d --name mongodb --network quake-network mongo:latest
   ```

4. Start the application:
   ```bash
   docker run -d --name quake-api --network quake-network -p 8080:8080 -e MONGO_URI='mongodb://mongodb:27017' quake-log-parser
   ```

5. Access the API at http://localhost:8080 and the Swagger documentation at http://localhost:8080/swagger/index.html

### Web Interface

1. Open a new terminal

2. Enter the frontend folder and run this command 
```bash
# If you have a simple server:
cd frontend
python -m http.server 8000
```
Then open http://localhost:8000 in your browser.



## Project Structure

```
quake_log_parser/
├── data/                # Sample data files
│   └── games.log        # Sample Quake log file
├── database/            # Database interaction layer
│   └── database.go      # MongoDB connection and operations
├── docs/                # Swagger documentation
├── frontend/            # Simple web UI
│   ├── index.html
│   ├── script.js
│   └── style.css
├── parser/              # Log file parsing logic
│   ├── models.go        # Parser data structures
│   └── parser.go        # Log parsing implementation
├── reporter/            # Game report generation
│   ├── models.go        # Report data structures
│   └── reporter.go      # Report formatting
├── main.go              # Application entry point
├── models.go            # API response models
├── routers.go           # API routes and handlers
├── Dockerfile           # Docker configuration
├── go.mod               # Go module definition
└── README.md            # This file
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
