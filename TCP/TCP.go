package main

import (
	"fmt"
	"time"
)

type Package struct {
	SYN  bool
	ACK  bool
	Seq  int
	Ack  int
	Data string
}

func main() {
	c2s := make(chan Package)
	s2c := make(chan Package)

	go client(c2s, s2c)
	go server(c2s, s2c)

	time.Sleep(200 * time.Second)

}

func client(c2s chan Package, s2c chan Package) {
	c2s <- Package{SYN: true, Seq: 100}
	fmt.Println("client sending syn 100")
	result := <-s2c
	fmt.Println("client received ", result)
	c2s <- Package{ACK: true, SYN: false, Seq: result.Seq, Ack: result.Ack + 1}

}

func server(c2s chan Package, s2c chan Package) {
	result := <-c2s
	fmt.Println("server received", result)

	s2c <- Package{SYN: true, ACK: true, Seq: result.Seq + 1, Ack: 300}
	fmt.Println("server sent syn-ack")

	result = <-c2s
	fmt.Println("server received ", result)

}
