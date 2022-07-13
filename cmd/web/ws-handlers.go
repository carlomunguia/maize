package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	*websocket.Conn
}

type WsPayload struct {
	Action      string              `json:"action"`
	Message     string              `json:"message"`
	MessageType string              `json:"messageType"`
	UserName    string              `json:"username"`
	Conn        WebSocketConnection `json:"-"`
}

type WsJsonResponse struct {
	Action  string `json:"action"`
	Message string `json:"message"`
	UserID  int    `json:"userID"`
}

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[WebSocketConnection]string)

var wsChan = make(chan WsPayload)

func (app *application) WsEndPoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Printf("New connection from %s", r.RemoteAddr)
	var response WsJsonResponse
	response.Message = "Connected to the server"

	err = ws.WriteJSON(response)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	go app.ListenForWS(&conn)

}

func (app *application) ListenForWS(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			app.errorLog.Println("ListenForWS recovered from panic:", r)
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			app.errorLog.Println(err)
			break
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

func (app *application) ListenToWSChannel() {
	var response WsJsonResponse
	for {
		e := <-wsChan
		switch e.Action {
		case "deleteUser":
			response.Action = "logout"
			response.Message = "Your account has been deleted"
			app.broadcastToAll(response)
		default:
		}
	}
}

func (app *application) broadcastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			app.errorLog.Printf("Websocket error: %s on %s", response.Action, err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}
