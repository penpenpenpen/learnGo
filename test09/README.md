#test09
## one master and clients

A leader is a POD running under a service leader. The image is **hvescovi/mastersvc**

Some executions examples:

```
./client port 10.245.1.3 30761 add 1
server ID = master, answer = add 1
./client port 10.245.1.3 30761 get
server ID = master, answer = 1
./client port 10.245.1.3 30761 add 5
server ID = master, answer = add 5
./client port 10.245.1.3 30761 get
server ID = master, answer = 6
```

By default, the master runs on port 8091, but accessing kubernetes from outside of the cluster, we want to access through a port of the node.
