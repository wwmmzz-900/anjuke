package main

import (
	"log"
	"sync"
)

var (
	wg    sync.WaitGroup
	mutex sync.Mutex
	flag  int
)

func main() {
	wg.Add(3)
	go LogCat()
	go LogDog()
	go LogFish()
	wg.Wait()
}
func LogCat() {
	defer wg.Done()
	for i := 0; i < 100; i++ {
		log.Println("cat")
	}

}

func LogDog() {
	defer wg.Done()
	for i := 0; i < 100; i++ {
		log.Println("dog")
	}

}

func LogFish() {
	defer wg.Done()
	for i := 0; i < 100; i++ {
		log.Println("fish")
	}

}
