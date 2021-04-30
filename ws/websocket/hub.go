package websocket

type Hub struct {
	Connect    chan Group
	Disconnect chan Group
	Groups     map[string]map[string]bool
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
				cancelled <- true
				return
			}
		}
	}(done)

	return cancelled
}