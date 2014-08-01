MONTY
=====

Monty is a server health monitoring and alert service written in Golang.

## Installation

The following instructions are incomplete, see Dockerfile.

# DB setup
	su -c "psql" postgres
	CREATE USER monty WITH PASSWORD 'insecure';
	CREATE DATABASE monty OWNER monty;

docker build -t monty .
docker run -e "MONTY_TOKEN=sekrit" -e "MONTY_DBSPEC=postgres://pqgotest:password@localhost/pqgotest" -p 9000:9000 -d monty

## Endpoints

See monty/conf/routes

### Status

Performing a GET to /api/status/:id will allow you to retrieve the most up-to-date Results for all Probes on a Node. The :id key is the 'mid' field specified in the Node during it's POST / creation.

- /api/status/:id

### Nodes

A node offers the ability to map and register a specific service (like hackerlist.net) by a unique string label. 

GET a Node by :id or POST a new Node
- /api/nodes/:id 

GET a list of all Nodes registered within monty
- /api/nodes/

### Probes

A probe is a type of daemon/worker which is configured to test/monitor/report various aspects of a Node's health. A Node may be instrumented with several probes, each which is unique programmed/scripted and responsible for reporting a specific service of a Node. For instance, one probe may be configured to test the uptime of an HTTP server while enough may check that port 22 is open and receiving SSH connections.

GET a Probe by :id or POST a new Probe
- /api/probes/:id

GET a list of all Probes
- /api/probes/

### Scripts

Each Probe has a Script component, containing a string of Lua code, which instructs the Probe how to execute its probing strategy.

GET an existing Script by :id or POST a new Script for a Probe
- /api/scripts/:id

GET a list of all Scripts
- /api/scripts/

### Results

Every time a Probe runs its Script, a Result is generated.

GET a specific Result by :id
- /api/results/:id

GET a list of all Results
- /api/results/

