# Blockchian-IPFS-Datasharing

### Overview of system
##### The project code is divided into two main parts: application and blockchain.
The blockchain folder contains scripts for starting and shutting down the blockchain network and specific blockchain network configurations and test configurations. 
The application folder contains the client and system backend. Specifically, the client folder mainly contains the front-end of the web page, and the server-side folder mainly contains the system routing and function implementation under the Gin framework.

### Steps to run
1. Environment
ubuntu 20.04, docker, docker-compose, go1.19
2. Run the blockchain network start script
```bash
cd blockchain
./start.sh
```
3. Start the server
```bash
cd application/server
go run main.go
```

3. Stop the blockchain network
```bash
cd blockchain
./stop.sh
```

### Tape testing
Under the /blockchain/tape, running the config.yaml with setting number of clients/nodes to specific number
```bash
./tape --config=config.yaml --number=n
```