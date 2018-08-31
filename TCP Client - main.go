package main

import (
	"fmt"
	"net"
)

func main() {
	//Connect TCP
	conn, err := net.Dial("tcp", "127.0.0.1:1111")
	if err != nil {
		fmt.Print(err)
	}
	defer conn.Close()

	conn.Write([]byte("S"))

	for {
		//simple Read
		buffer := make([]byte, 1024)

		n, _ := conn.Read(buffer)

		fmt.Printf("%s\n", buffer[:n])
	}

}
