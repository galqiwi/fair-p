# Fair-p

Fair-p is an HTTP(S) proxy with throughput guarantees.

## Installation

### Prerequisites
- Ensure Docker is installed on your system. Follow this [guide for Ubuntu](https://docs.docker.com/engine/install/ubuntu/).
- Docker Compose should be included (modern Docker versions have it built-in).

### Setup

**Clone the Repository:**

```git clone https://github.com/galqiwi/fair-p.git```
    
**Configure docker-compose.yml:**
- **Max Throughput:** Update the --max-throughput value in docker-compose.yml to the maximum throughput (in MB/s) you want the proxy to handle.
- **Port:** Optionally, you can change the proxy's port by modifying the --port parameter.

### Build and Run

```docker compose build && docker compose up -d```

Your Fair-p proxy should now be up and running!
