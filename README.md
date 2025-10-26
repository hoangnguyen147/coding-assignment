This is a coding assignment.

The answer for assignment is in the folder 1.1, 1.2, 1.3, 3 specific for the part of each question in assignment.

I have each README for each module. Please check them for more details.

## How to Run This Project Locally

### Prerequisites

- **Go 1.24.3** or later installed on your system
- **curl** or **Postman** (for testing HTTP endpoints)

### Project Structure

This project contains multiple Go modules:

- **1.1**: Payment service with mutex-based concurrency control and idempotency
- **1.2**: Worker pool with ordered output using `sync.Cond`
- **1.3**: Simple HTTP handler demonstrating data race issues and mutex solution
- **3**: Advanced HTTP service with both mutex and channel-based approaches

### Running Individual Modules

#### Module 1.1 - Payment Service

```bash
cd 1.1
go run main.go
```

The service will start on `http://localhost:8080`

**Test the payment endpoint:**

```bash
# Set initial balance for a user
curl -X POST http://localhost:8080/pay \
  -H "Content-Type: application/json" \
  -d '{
    "userID": "user123",
    "amount": 100,
    "transactionID": "txn001"
  }'

# Test idempotency - send the same request again
curl -X POST http://localhost:8080/pay \
  -H "Content-Type: application/json" \
  -d '{
    "userID": "user123",
    "amount": 50,
    "transactionID": "txn002"
  }'
```

**Run tests:**

```bash
cd 1.1
go test -v
```

#### Module 1.2 - Worker Pool with Ordered Output

```bash
cd 1.2
go run main.go
```

This will process 100 numbers through 5 workers and print results in order using `sync.Cond`.

#### Module 1.3 - Simple HTTP Handler

```bash
cd 1.3
go run main.go
```

The service will start on `http://localhost:8080`

**Test the handler:**

```bash
# POST data
curl -X POST http://localhost:8080/ \
  -d "Hello, World!"

# Try with GET (should fail)
curl http://localhost:8080/
```

#### Folder 2.1 and 2.2:

Just simple folder answer for SQL part. I have added the answer in the file answer.sql and folder's README.md.

#### Module 3 - Advanced HTTP Service (Mutex vs Channel)

```bash
cd 3
go run main.go
```

The service will start on `http://localhost:8080`

**Test mutex-based endpoints:**

```bash
# Set data using mutex
curl -X POST http://localhost:8080/mutex/set \
  -d "data with mutex"

# Get data using mutex
curl http://localhost:8080/mutex/get
```

**Test channel-based endpoints:**

```bash
# Set data using channel
curl -X POST http://localhost:8080/channel/set \
  -d "data with channel"

# Get data using channel
curl http://localhost:8080/channel/get
```

**Run tests:**

```bash
cd 3
go test -v
```

### Running All Modules with Go Workspace

This project uses Go workspaces to manage multiple modules. To run tests for all modules:

```bash
# From the project root
go test ./1.1 ./1.2 ./3 -v
```

Or run each module individually as shown above.

Honestly, I generated this README.md file with AI. I think it's a good tool to help us generate documentation. But we should always review and make sure it's correct.
