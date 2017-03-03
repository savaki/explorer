# explorer

utility to explore a container environments from within the environment

## Overview

```explorer``` is docker container that lets you explore the container environment that it's being 
run in.  It offers a number of ways to do this:

* ```/*``` - behaves a file server with a docroot of /
* ```/_/env``` - returns the environment variables
* ```/_/echo``` - returns the http header and body 

## Installation

```
docker run -d -e HEARTBEAT=true -p 5002:5002 savaki/explorer 
```

## Usage

```
NAME:
   explorer - web server to introspect a running container

USAGE:
   main [global options] command [command options] [arguments...]

VERSION:
   0.3.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value   port number (default: "5002") [$PORT]
   --heartbeat    heartbeat [$HEARTBEAT]
   --delay value  seconds to wait after shutdown completes before exiting (default: 5) [$DELAY]
   --help, -h     show help
   --version, -v  print the version
```
