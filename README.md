# Fair-p

Fair-p is an HTTP(S) proxy with throughput guarantees.

## Installation

### Prerequisites
- Ensure Docker is installed on your system. Follow this [guide for Ubuntu](https://docs.docker.com/engine/install/ubuntu/).
- Docker Compose should be included (modern Docker versions have it built-in).

### Setup

**Setup proper linux limits (optional)**

Run [this](https://gist.githubusercontent.com/galqiwi/9b90ac2ea2093c6dbec83dabb7162cce/raw/beab3b6d70acd00a5e7975996d80c9c7e343209e/setup-1m-sockets.sh) script.

**Clone the Repository:**

```git clone https://github.com/galqiwi/fair-p.git```
    
**Configure docker-compose.yml:**
- **Max Throughput:** Update the --max-throughput value in docker-compose.yml to the maximum throughput (in MB/s) you want the proxy to handle.
- **Port:** Optionally, you can change the proxy's port by modifying the --port parameter.

### Build and Run

**Using Docker:**

```docker compose build && docker compose up -d```

**Using Go (without Docker):**

Prerequisites:
- Go 1.22.1 or later

Build and run:
```bash
go build -o fair-p ./cmd/passer
./fair-p --port 8080 --max-throughput 100
```

Your Fair-p proxy should now be up and running!
