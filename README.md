# PetHealth Service 🐾

A high-performance metrics collection and processing system built with **Go**. The architecture focuses on write-throughput, data integrity, and observability.

## 🏗 Architecture Decisions

The project implements several distributed systems patterns:

*   **Transactional Outbox:** Ensures atomicity between state persistence in PostgreSQL and event publishing to Kafka.
*   **Database Sharding:** Manual horizontal sharding of PostgreSQL based on `pet_id` (consistent hashing) to distribute write load.
*   **Worker Pool:** Concurrent task processing with a bounded goroutine pool to manage resource consumption and handle graceful shutdown.
*   **Graceful Shutdown:** Proper handling of `SIGTERM`/`SIGINT` signals to ensure active workers complete tasks without data loss.
*   **Event-Driven Design:** Asynchronous component interaction via Kafka for service decoupling.
* **Cloud Native:** Fully containerized and ready for Kubernetes deployment.

## 🛠 Tech Stack

*   **Runtime:** Go 1.25+
*   **API:** Gin Gonic (REST)
*   **Serialization:** **Protocol Buffers (Protobuf)** for efficient binary serialization
*   **Storage:** PostgreSQL (Sharded), GORM
*   **Messaging:** Apache Kafka (segmentio/kafka-go)
*   **Observability:** Prometheus (metrics), Uber-Zap (logging)
*   **Infrastructure:** Kubernetes, Docker Compose

## 🚀 Operations & Deployment

### Local Infrastructure
* Minikube (for K8s deployment)
```bash
minikube start
```
* Docker Compose
```bash
# Start infrastructure (DB shards, Kafka, kafdrop)
docker-compose up -d
```

## Run the application locally
```bash
go run cmd/server/main.go
```
## Kubernetes Management

### 1. Apply configurations and secrets
kubectl apply -f k8s/config.yaml

### 2. Deploy the application
kubectl apply -f k8s/deployment.yaml

### 3. Resource inspection
kubectl get all
kubectl describe pod <pod_name>
kubectl get configmaps

### 4. Debugging and logs
kubectl logs -f deployment/pethealth-deployment
kubectl exec -it <pod_name> -- /bin/sh

### 5. Port-forwarding for API and Monitoring
kubectl port-forward deployment/pethealth-deployment 8080:8080
kubectl port-forward svc/prometheus-service 9090:9090

### 6. Cleanup
kubectl delete -f k8s/deployment.yaml
kubectl delete -f k8s/config.yaml


## 📊 Sharding Logic

Data distribution across shards is calculated using a consistent hashing approach:

$$shard\_index = pet\_id \pmod{total\_shards}$$

*   **Data Locality:** Ensures all data for a specific pet resides on the same shard.
*   **Scalability:** Simplifies horizontal scaling of the database layer.

## 📝 Roadmap

*   [ ] **Reliability Layer:** Implementation of **Retry Policy** and **Dead Letter Queues (DLQ)** for Kafka consumers.
*   [ ] **Distributed Transactions:** **Saga Pattern** for the "Pet Registration & Alert Generation" flow.
*   [ ] **Observability:** Grafana dashboards for metrics visualization (Kafka lag, Worker Pool state).
*   [ ] **Tracing:** **Distributed Tracing** integration (OpenTelemetry/Jaeger).