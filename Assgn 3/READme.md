Koe Jia Yee 1003017

# PSET3
Total number of clients in network (excl.CM) = 10 to 20\
This value can be changed in the `main` function, `var numNodes`

Node 999 = CM\
Node 99x = Replicas\
Currently there are 3 Replicas. This value can be changed in the `main` function, `var numNodes`

# Q2
Fault detection is implemented using periodic pinging of the replicas to the Pri CM.\
Election is implemented using Bully.\
Broadcasting of requests to replicas is done every time a new requests is added to the Pri CM's queue.

# Q3
It preserves sequential consistency if the CM successfully broadcasts all requests to all the replicas. 

All the client machines execute their requests according to program order. Requests are served using a FIFO queue given that the go routine, which is a single thread,  is listening to messages sent to the channel. The primary CM orders the read and write requests in an array. It then broadcasts the requests list to all the secondary replicas which maintains that order. Hence, all machines will observe results according to some total order.

However, under certain conditions, sequential consistency is not guaranteed: 

- If the broadcasted requests are not sent successfully to the replicas, not all servers will see the same total order. 
- If the replica fails temporarily. After it revives, the messages it failed to receive are not retrieved anymore, thus violating sequential consistency.  

# Experiment
All timings can be found in `DS PSET3 Timings.pdf`

# Evaluation
1. Fault Tolerant version of the Ivy Protocol takes a longer time to complete requests. 
The implemented fault detection mechanism was that replicas periodically ping the primary CM at certain time intervals. The Central manager has to take additional CPU clock cycles to reply the pings to indicate that it's alive.

The Pri CM also broadcasts the request list to replicas everytime it adds a request to it's request queue. The sending of message also results in some communication overhead for The Central server and result in delay in its servicing of the rest of the requests from other clients

2. Pri CM failure without revival takes a shorter time to complete requests than when the Pri CM revives mid-way

On revival, the revived node calls for another election, which is why it takes longer than in the case without revival. 


# Running Instructions
## Basic Ivy without Replicas
1. ```go run BasicIvy.go [numRequests]``` where `numRequests` = number of clients simultaneously requesting

## Fault Tolerant Ivy 

When all requests complete, press `Enter` to terminate the program & pinging.

### Without Faults
1. `cd Ivy1`
2. ```go run CM1.go Ivy1.go [numRequests]``` where `numRequests` = number of clients simultaneously requesting 

### With Faults - Pri CM dies
1. `cd Ivy2`
2. ```go run CM2.go Ivy2.go 20``` as at least 20 nodes are needed to help demonstrate the full simulation

### With Faults - Pri CM dies and revives
1. Run steps 1 & 2 as above
2. When the print statement

`==================
99X WON ELECTION. START COMPLETING REQUESTS
==================`

appears, press `Enter` to simulate nonde revival.
