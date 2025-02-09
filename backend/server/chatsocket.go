package server

import (
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"log/slog"
)

/**
HOW TO USE
server: NewServer()
http.Handle("/ws", websocket.Handler(server.handleWS))
*/

type Server struct {
	conns map[*websocket.Conn]string
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]string),
	}
}
func (s *Server) handleWS(ws *websocket.Conn) {
	ctx := ws.Request().Context()
	uid, ok := ctx.Value("userID").(string)
	if !ok {
		slog.Warn(fmt.Sprintf("No userID found in context: %v", ws.RemoteAddr()))
	}
	slog.Info(fmt.Sprintf("new incoming connection from client: %v", ctx.Value("userID")))
	s.conns[ws] = uid
	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			slog.Warn(fmt.Sprintf("read error: %v", err))
			continue
		}
		msg := buf[:n]
		s.broadcast(msg)
	}
}

func (s *Server) broadcast(b []byte) {
	for ws := range s.conns {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				fmt.Println("write error:", err)
			}
		}(ws)
	}
}
