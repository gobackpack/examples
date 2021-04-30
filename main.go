package main

import (
	"fmt"
	"github.com/gobackpack/examples/ws/websocket"
	"github.com/sirupsen/logrus"
	"sync"
)

func main() {
	//arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	//logrus.Info(arr)
	//
	//logrus.Info(arr[1:5])
	//
	//// copy rest of the array
	//copy(arr[2:], arr[2+1:]) // dst: 3, 4..., start: 4, 5...
	////arr[len(arr)-1] = nil
	//arr = arr[:len(arr)-1]
	//
	//logrus.Info(arr)
	//
	//// replace with last element
	//arr[3] = arr[len(arr)-1]
	////arr[len(arr)-1] = nil
	//arr = arr[:len(arr)-1]
	//
	//logrus.Info(arr)

	hub := websocket.NewHub()

	done := make(chan bool)
	cancelled := hub.ListenConnections(done)

	group1Count := 2000
	group2Count := 100
	wg := sync.WaitGroup{}
	wg.Add(group1Count)

	for i := 0; i < group1Count; i++ {
		go func(i int, wg *sync.WaitGroup) {
			hub.Connect <- websocket.Group{
				GroupId:      "group_1",
				ConnectionId: fmt.Sprintf("client_%d", i),
			}
			wg.Done()
		}(i, &wg)
	}

	wg.Add(group2Count)
	for i := 0; i < group2Count; i++ {
		go func(i int, wg *sync.WaitGroup) {
			hub.Connect <- websocket.Group{
				GroupId:      "group_2",
				ConnectionId: fmt.Sprintf("client_%d", i),
			}
			wg.Done()
		}(i, &wg)
	}

	wg.Wait()

	hub.Disconnect <- websocket.Group{
		GroupId:      "group_2",
		ConnectionId: "client_99",
	}

	hub.Disconnect <- websocket.Group{
		GroupId:      "group_2",
		ConnectionId: "client_98",
	}

	close(done)
	<-cancelled

	logrus.Info("groups: ", len(hub.Groups))
	logrus.Info("group_1: ", len(hub.Groups["group_1"]))
	logrus.Info("group_2: ", len(hub.Groups["group_2"]))

	logrus.Info("server exited")
}
