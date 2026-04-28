# PetHealth Service 🐾

High-performance metrics collection and processing system built with **Go**. Designed for scalability, data integrity, and observability.

## 🏗 Architecture Highlights

The project demonstrates advanced backend patterns used in distributed systems:

* **Database Sharding:** Manual horizontal sharding for PostgreSQL to handle high-write throughput.
* **Transactional Outbox Pattern:** Ensures 100% data consistency between PostgreSQL and Kafka, preventing data loss during network failures.
* **Worker Pool Pattern:** Concurrent background workers for processing the Outbox table and consuming Kafka events.
* **Event-Driven Design:** Asynchronous processing of health alerts via Apache Kafka.
* **Cloud Native:** Fully containerized and ready for Kubernetes deployment.

## 🛠 Tech Stack

* **Language:** Go (Golang) 1.25+
* **API:** Gin Gonic (REST)
* **Data Serialization:** **Protocol Buffers (Protobuf)** for efficient binary serialization and API contract definition.
* **Databases:** PostgreSQL (Multiple Shards), GORM
* **Messaging:** Apache Kafka (segmentio/kafka-go)
* **Logging:** Uber-Zap (Structured Logging)
* **Infrastructure:** Docker, Docker Compose, Kubernetes (Minikube)
* **Monitoring:** Prometheus (coming soon)

## 🚀 Getting Started

### Prerequisites
* Docker & Docker Compose
* Minikube (for K8s deployment)

### Local Development (Hybrid Mode)
To run the infrastructure (DBs and Kafka) locally while keeping the app flexible:

# Start the infrastructure:
   ```bash
   docker-compose up -d
  ```  
# Run the application:
  ```bash
  go run cmd/server/main.go
 ```

# Kubernetes Deployment
 
# Apply configurations
 ```bash
kubectl apply -f k8s/config.yaml
 ```
# Deploy the application
 ```bash
kubectl apply -f k8s/deployment.yaml
```
# Access the API
 ```bash
kubectl port-forward deployment/pethealth-deployment 8080:8080
```

📊 Sharding Logic
Data is distributed across shards based on the pet_id using a consistent hashing approach:
shard_index = pet_id % total_shards

This ensures that all data for a specific pet always resides on the same shard, simplifying queries and maintaining locality.

📝 Future Roadmap

[ ] Integration with Prometheus & Grafana for shard-level monitoring.

[ ] Implementation of Graceful Shutdown for K8s pods.

[ ] Expansion of Sharding logic to support dynamic re-balancing.