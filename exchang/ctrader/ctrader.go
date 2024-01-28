package ctrader

import (
	"github.com/2nth0nyj/go-trade/exchang/ctrader/connection"
)

type Client struct {
	conn              *connection.Connection
	accountParameters map[string]interface{}
	balance           *string
}

func NewClient(clientId, clientSecret, accessToken string, ctid int64, live bool) *Client {
	ctraderConnection := connection.NewConnection(clientId, clientSecret, accessToken, ctid, live)
	m := map[string]interface{}{"clientId": clientId, "clientSecret": clientSecret, "accessToken": accessToken, "ctid": ctid, "live": live}

	return &Client{
		conn:              ctraderConnection,
		accountParameters: m,
	}
}

func (c Client) Balance() string {
	if c.balance == nil {
		balance := c.conn.GetBalance()
		c.balance = &balance
		return *c.balance
	}
	return *c.balance
}

func (c Client) Broker() string {
	return "Spotware CTrader"
}
