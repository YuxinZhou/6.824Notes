package main

import (
	"time"
	"fmt"
)

func main()  {
	ch := make(chan int)
	go func() {
		fmt.Print("111")
		ch <- 1
		fmt.Print("222")
		fmt.Print("333")
	}()
	time.Sleep(time.Second)

}

