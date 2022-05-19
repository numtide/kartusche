package wsclient

import "github.com/gorilla/websocket"

type WSClient struct {
	Conn *websocket.Conn
}

func (w *WSClient) ReadJson() (interface{}, error) {
	var v interface{}
	err := w.Conn.ReadJSON(&v)
	if err != nil {
		return nil, err
	}

	return v, nil
}
