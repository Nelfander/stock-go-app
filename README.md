# Stock-Go-App: Real-Time Event-Driven Pipeline

A high-performance stock market monitoring system built with **Go** and **Redis Streams**. This project demonstrates a decoupled microservices architecture designed to handle high-throughput financial data.

## 🏗️ Architecture
The system is divided into three main components:
* **Ingester Service**: Simulates a live market feed (Random Walk algorithm) and produces events to a Redis Stream.
* **Analyzer Service**: A stateful consumer that processes price movements in real-time, detecting volatility spikes using a moving-window logic.
* **Shared Library**: A unified data contract (Structs) used by both services to ensure data integrity and prevent "schema drift."



## 🛠️ Technical Highlights
* **Go Workspaces (1.24)**: Managed as a monorepo using `go.work` for seamless local module development.
* **Graceful Shutdown**: Implemented `signal.NotifyContext` to handle `SIGINT/SIGTERM`, ensuring all Redis connections close safely without data loss.
* **Structured Logging**: Utilized the `slog` library for production-ready, searchable logs.
* **Event-Driven Design**: Uses **Redis Streams** for persistent message queuing, allowing services to scale independently.
* **Defensive Programming**: Robust type assertions and JSON unmarshaling to handle malformed data without service panics.

## 🚀 Getting Started

### Prerequisites
* Go 1.24+
* Docker & Docker Compose

### Installation
1. Clone the repo: `git clone https://github.com/nelfander/stock-go-app.git`
2. Start the infrastructure: `docker-compose up -d`
3. Run the Ingester: `cd ingester-service && go run main.go`
4. Run the Analyzer: `cd analyzer-service && go run main.go`