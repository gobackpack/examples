package websocket

import (
	websocketLib "github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocketLib.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Hub struct {
	Connect    chan Group
	Disconnect chan Group
	Groups     map[string]map[string]bool
	Connection *websocketLib.Conn
}

type Group struct {
	GroupId      string
	ConnectionId string
}

func NewHub() *Hub {
	return &Hub{
		Connect:    make(chan Group),
		Disconnect: make(chan Group),
		Groups:     make(map[string]map[string]bool, 0),
	}
}

func (hub *Hub) ListenConnections(done chan bool) chan bool {
	cancelled := make(chan bool)

	go func(done chan bool) {
		defer func() {
			cancelled <- true
		}()

		for {
			select {
			case gr := <-hub.Connect:
				if hub.Groups[gr.GroupId] == nil {
					hub.Groups[gr.GroupId] = make(map[string]bool, 0)
				}
				hub.Groups[gr.GroupId][gr.ConnectionId] = true
				break
			case gr := <-hub.Disconnect:
				if hub.Groups[gr.GroupId] != nil {
					delete(hub.Groups[gr.GroupId], gr.ConnectionId)
				}
				break
			case <-done:
				return
			}
		}
	}(done)

	return cancelled
}

func (hub *Hub) ConnectToGroup(w http.ResponseWriter, r *http.Request, groupId string) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	hub.Connection = conn

	return nil
}

func (hub *Hub) SendToGroup(groupId string) {

}

func (hub *Hub) SendToConnectionId(groupId, connectionId string) {

}

func (hub *Hub) SendToAllGroups() {

}
