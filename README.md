# Receipt Processor

The **Receipt Processor** is a web service that processes receipts, calculates points based on specific rules, and stores the data in-memory using **Redis**. The service provides two main API endpoints:

1. **`POST /receipts/process`**: Receives a JSON receipt, processes it, and returns a unique ID.
2. **`GET /receipts/{id}/points`**: Retrieves the points awarded for a given receipt ID.

## Installation

To get started, ensure that **Go** and **Redis** are installed on your machine. The instructions below will help you set up the project.

### 1. Clone the Repository

Clone this repository to your local machine:

```sh
git clone https://github.com/codezilla322/receipt-processor-challenge.git
cd receipt-processor
```

### 2. Install Dependencies

Run the following command to download required Go modules:

```sh
go mod tidy
```

### 3. Start Redis Server

Make sure Redis is running locally on port 6379

```sh
redis-server
```

### 4. Run the Application

After Redis is running, you can start the Go application by running:

```sh
go run main.go
```

This will start the web service on port 8080.
