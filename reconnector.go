package reconnector

import "fmt"
import "time"

//import "bufio"

func SleepMillis(m int) {
	time.Sleep(time.Duration(m) * time.Millisecond)
}

type WriteFunc func(func())
type DisconnectFunc func(error)
type KeepReedingFunc func(wf func(func())) error

//type KeepReedingFunc func(df DisconnectFunc, wf WriteFunc) // I don't know why I can't do this

func Do(waitMillis int, connect func() error, keepReeding KeepReedingFunc, close func() error) (wf WriteFunc, cc func(bool)) {
	// todo: force close

	disconnectChan := make(chan error)
	writeChan := make(chan func())
	closeChan := make(chan bool)

	writeFunc := func(theFunc func()) {
		writeChan <- theFunc
	}

	closeCon := func(reconnect bool) {
		closeChan <- reconnect
	}

	go func() {
		goto doTheConnection

	sleepThenReconnect:
		close()
		fmt.Println("Some sort of error, waiting " + fmt.Sprint(waitMillis) + " millis.")
		SleepMillis(waitMillis)

	doTheConnection:
		err := connect() // you could use the channel for the error here too?
		if err != nil {
			fmt.Println(err.Error())
			goto sleepThenReconnect
		}

		// read in another goroutine
		// "Multiple goroutines may invoke methods on a Conn simultaneously."
		go func() {
			for {
				err := keepReeding(writeFunc)
				if err != nil {
					fmt.Println("Yo broke!!!!!!!!!!")
					disconnectChan <- err
					break
				} else {
					fmt.Println("yay I got someting!!! 000000000000")
				}
			}
			fmt.Println("finally got here!!!!!---------------------!(2) ")
		}()

		shouldReconnect := true
	thefor:
		for {
			select {
			case e := <-disconnectChan:
				fmt.Println("disconnecting you!!!!!!!!!!!!!!!!!" + e.Error())
				if shouldReconnect {
					fmt.Println("but reconnecting")
					goto sleepThenReconnect
				} else {
					fmt.Println("but NOT reconnecting")
					break thefor
				}
			case reconnect := <-closeChan:
				shouldReconnect = reconnect
				fmt.Println("<close me>")
				close()
				fmt.Println("</close me>")
			case toWrite := <-writeChan:
				// TODO: handle error here?
				fmt.Println("trying to write...")
				go toWrite() // should I do this in a goroutine?
				// default read bytes?
			}
			fmt.Println("finally got here!!!!!---------------------!(3)")
		}

		fmt.Println("finally got here!!!!!---------------------!(1)")
	}()

	return writeFunc, closeCon
}
