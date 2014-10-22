package reconnector_test

import "fmt"
import "reconnector"
import "testing"
import "net"
import "bufio"
//import "errors"


func startDummyServer(t *testing.T, waitChan chan bool, messageChan chan string) {
  ln, err := net.Listen("tcp", ":8094")
  if err != nil {
    t.Error("Errror creating dummy server\n")
  } 
  waitChan <- true

  for {
    conn, err := ln.Accept()
    if err != nil {
      continue
    }
    go func() {
      for {
        r := bufio.NewReader(conn) 
        str, err := r.ReadString('\n')
        if err != nil {
          fmt.Println("error reading")
          conn.Close()
          break;
        } else {
          conn.Write([]byte("this be server " + str)) // including the delimiter
					messageChan <- str
        }
      } 
    }()
  }
}




func TestDo(t *testing.T) {
  fmt.Println("tested!") 

  waitChan := make(chan bool)
  messageChan := make(chan string)
  gotItChan := make(chan string)
  go startDummyServer(t, waitChan, messageChan)
  <- waitChan
  var c net.Conn
  var e error
  var r *bufio.Reader

  writeFunc, close := reconnector.Do(1000, func() error {
    fmt.Println("connecting")
    c, e = net.Dial("tcp", "localhost:8094")
    if e != nil {
      t.Error("error connecting to server!")
    }
    r = bufio.NewReader(c) 
    return e
  }, func(writeFunc func(func())) error {
    //reply := make([]byte, 1024)
    //_, err := conn.Read(reply)
    str, err := r.ReadString('\n') 
    if err != nil {
      //disconnectFunc(errors.New("error inital connnection"))
      return err
    }
    fmt.Println("Server said: " + str)
		gotItChan <- str
    return nil
  }, func() error {
    return c.Close()
  }) 
  reconnector.SleepMillis(500)
	close(true)
	reconnector.SleepMillis(1100)
  writeFunc(func() {
    // try to rewrite if disconnected?
    c.Write([]byte("Hello world\n"))
  })

	str := <- messageChan

	if str != "Hello world\n" {
		t.Error("it was supposed to get hello world")	
	} else {
		fmt.Println("what the yay!")	
	}
 	got := <- gotItChan 
	if got != "this be server Hello world\n" {
		t.Error("it was supposed to get  a message from the server")	
	} else {
		fmt.Println("what the yay!")	
	}
	//c.Close()
	//c.Write([]byte("foozy\n"))
	reconnector.SleepMillis(2100)
	close(false)
	
	reconnector.SleepMillis(2100)
	fmt.Println("done")

}
