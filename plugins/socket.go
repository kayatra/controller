package plugins

import(
  "github.com/gorilla/websocket"
  "time"
  log "github.com/Sirupsen/logrus"
  "net/http"
  "net"
  "sync/atomic"
  "github.com/home-control/core/transport"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  4096,
    WriteBufferSize: 4096,
}

const pingInterval = time.Second*45
var connectionIdCounter uint64 = 0
var commandIdCounter uint64 = 0

type socketMsg struct{
  Type              string      `json:"message_type"`
  Body              interface{} `json:"body"`
  CommandId         uint64      `json:"command_id"`
  Time              time.Time   `json:"time"`
}

type Connection struct{
  ConnectionId  uint64
  RemoteAddr    net.Addr
  Connection    *websocket.Conn
}

func (c *Connection) NewMessage(mt string, b interface{}) *socketMsg{
  m := socketMsg{}
  m.CommandId = atomic.AddUint64(&commandIdCounter, 1)
  m.Type = mt
  m.Time = time.Now().UTC()
  m.Body = b

  return &m
}

func (c *Connection) SendMessage(mt string, b interface{}){
  msg := c.NewMessage(mt, b)

  log.WithFields(log.Fields{
    "id": c.ConnectionId,
    "type": mt,
    "body": b,
  }).Debug("Sending message")

  err := c.Connection.WriteJSON(msg)
  if err != nil {
    log.WithFields(log.Fields{
      "address": c.RemoteAddr,
      "id": c.ConnectionId,
      "error": err,
      "type": mt,
      "body": b,
    }).Error("Error sending message to client")
  }
}

func ping(c Connection, close chan struct{}) {
  ticker := time.NewTicker(pingInterval)
  defer ticker.Stop()
  for {
    select{
      case <-ticker.C:
        c.SendMessage("ping", "")
      case <-close:
        return
    }
  }
}

func Transport(w http.ResponseWriter, r *http.Request) {
  ws, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.WithFields(log.Fields{
      "err": err,
    }).Error("Failed to set websocket upgrade")
    return
  }

  defer ws.Close()

  c := Connection{}
  c.RemoteAddr = ws.RemoteAddr()
  c.Connection = ws
  c.ConnectionId = atomic.AddUint64(&connectionIdCounter, 1)

  log.WithFields(log.Fields{
    "address": c.RemoteAddr,
    "id": c.ConnectionId,
  }).Debug("New socket connection")

  done := make(chan struct{})

  h := transport.HeloCommand{}
  h.ConnectionId = c.ConnectionId
  h.PingInterval = pingInterval
  c.SendMessage("helo", h)

  go ping(c, done)

  for{
    msgType, msg, err := ws.ReadMessage()
    if err != nil{
      if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
        log.WithFields(log.Fields{
          "id": c.ConnectionId,
        }).Debug("Client disconnect")
      } else {
        log.WithFields(log.Fields{
          "id": c.ConnectionId,
          "err": err,
        }).Error("Error reading message")
      }
      break
    }

    log.WithFields(log.Fields{
      "id": c.ConnectionId,
      "type": msgType,
      "msg": msg,
    }).Debug("Got message from client")
  }

  log.WithFields(log.Fields{
    "id": c.ConnectionId,
  }).Debug("Closing client connection")

  close(done)
}
