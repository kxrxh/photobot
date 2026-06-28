package ws

import "sync"

type clientWriter struct {
	conn      wsConn
	send      chan Message
	closeOnce sync.Once
}

func newClientWriter(conn wsConn) *clientWriter {
	cw := &clientWriter{
		conn: conn,
		send: make(chan Message, 32),
	}
	go cw.run()
	return cw
}

func (cw *clientWriter) run() {
	for msg := range cw.send {
		if err := cw.conn.WriteJSON(msg); err != nil {
			log.Error().Err(err).Msg("Error sending message to WebSocket client")
		}
	}
}

func (cw *clientWriter) enqueue(msg Message) {
	select {
	case cw.send <- msg:
	default:
		log.Warn().Msg("WebSocket client send buffer full, dropping message")
	}
}

func (cw *clientWriter) close() {
	cw.closeOnce.Do(func() {
		close(cw.send)
	})
}
