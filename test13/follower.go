package main

import (
  "fmt"
  "net"
  "bufio"
  "os"
  "sync"
  "strings"
)

const DEBUG = true // if true, print debug messages
const DEVMODE = false //false: deploy in kubernetes

var (
  requestsPort string
  values map[string]string
)

func main() {

  var wg sync.WaitGroup
  wg.Add(1) //block the finish of the program while 1 thread alives

  //initialize the values map
  values = make(map[string]string)

  //create a log file
  f, _ := os.Create("/tmp/test13follower-ip"+getMyIP())
  defer f.Close()

  //define which port the FOLLOWER will listen
  requestsPort = "8092"

  //final connection string to leader (ip:port)
  destiny := ""

  if DEVMODE {

    //get server address and port from parameters
    if len(os.Args) < 2 {
      fmt.Printf("Please, specify the LEADER address.\n")
      os.Exit(0)
    }
    if len(os.Args) < 3 {
      fmt.Printf("USAGE: follower accessMode ipAddress [port]\n"+
        "accessMode: direct or port (you only have to specify the port is the 'port' mode)\n")
        os.Exit(0)
    }
    //access with address or with address:port?
    accessMode := os.Args[1]

    if accessMode == "direct" {
      destiny = os.Args[2]
    } else if accessMode == "port" {
      destiny = os.Args[2] + ":" + os.Args[3]
    } else if accessMode[:7] == "hacker=" {   //VERY HACKER SPECIFIC:
      destiny = os.Args[2] + ":" + os.Args[3] //on the same machine,
      requestsPort = accessMode[7:11]        // another port than 8092
                                  // (hacker=8093, for example)
    }
  } else {
    //try to discover the leader service and port from
    //environmental variables
    //TEST13LEADER_SERVICE_HOST=10.247.91.58
    //TEST13LEADER_SERVICE_PORT=8091

    //get the IP of the service
    leaderServiceAddress := os.Getenv("TEST13LEADER_SERVICE_HOST")

    //get the PORT of the service (it should be 8091, but let's ask it anyway)
    leaderServicePort := os.Getenv("TEST13LEADER_SERVICE_PORT")

    destiny = leaderServiceAddress + ":" + leaderServicePort

    //log the retrieved values
    f.WriteString("host IP of leader in environment variables: "+leaderServiceAddress+":"+leaderServicePort)
  }

  if DEBUG {fmt.Println("destiny="+destiny)}

  //connect
  conn, err := net.Dial("tcp", destiny)
  Check(err)

  //get the IP and PORT of this (introspect?)
  myAddress := getMyIP() + ":" + requestsPort

  //try to register with the leader
  cmd := "rgt " + myAddress

  if DEBUG {fmt.Println("trying to register, sending command: ", cmd)}

  //sending the command
  fmt.Fprintf(conn, cmd+"\n")

  //receive the answer
  message, err := bufio.NewReader(conn).ReadString('\n')
  Check(err)

  //get the random ID generated by the leader
  followerID := message[:len(message)-1]

  //print the ID
  fmt.Println("ID provided by the leader: ", followerID)

  //start waiting for messages from leader
  go listenRequests(followerID, &wg)

  wg.Wait()
}

func listenRequests(myID string, localWg *sync.WaitGroup) {

  //create the port listener
  ln, _ := net.Listen("tcp", ":" + requestsPort)

  //create a log file
  //f, _ := os.Create("/tmp/"+serverID)
  //defer f.Close()

  //create the SHARED counter!
  //counter := 0

  fmt.Println("follower waiting for connections")

  //variable that controls the main loop
  goOut := false

  for ; !goOut; {

    //wait for conections
    conn, _ := ln.Accept()

    //read the command
    cmd, _ := bufio.NewReader(conn).ReadString('\n')
    s := string(cmd)

    //debug
    fmt.Println("command received: ", cmd)

    //log the command
    //f.WriteString(cmd)

    //shutdown?
    if s == "sht\n" {
      localWg.Done()
      goOut = true

    } else if s == "get\n" {
      answer := values["counter"]

      if DEBUG {fmt.Println("answering some client, counter="+answer)}

      //send the answer back to the client
      conn.Write([]byte(answer+"\n"))

    } else { //it must be an update command like: key=value

      //prepare the answer
      answer := myID + "=>"

      if DEBUG {fmt.Print("size of commando:",len(s))}
      if DEBUG {fmt.Println(", command: ",s)}

      //parse the command in: key=value
      newValue := strings.Split(s,"=")

      //set the new value
      values[newValue[0]] = newValue[1]

      //complete the answer
      answer += "ok value updated"

      if DEBUG {fmt.Println("answer="+answer)}

      //send the answer back to the client
      conn.Write([]byte(answer+"\n"))
    }
  }
}

func Check(err error) {
  if err!=nil {
    fmt.Println("ERROR: ", err)
    os.Exit(0)
  }
}

func getMyIP() string {
    ifaces, err := net.Interfaces()
    Check(err)

    for _, i := range ifaces {
        addrs, err := i.Addrs()
        if err != nil {
            //log.Print(fmt.Errorf("localAddresses: %v\n", err.Error()))
            continue
        }
        for _, a := range addrs {
            //log.Printf("%v %v\n", i.Name, a)
            var s string
            s = a.String()
            //fmt.Println("a=",s, " part=",s[:6])

            if len(s) >5 && s[:6] == "10.245" {  //10.246 for containers
              fmt.Println("This is an address of container:",s[:10])
              return s[:10]
            }
        }
    }
    return ""
}
