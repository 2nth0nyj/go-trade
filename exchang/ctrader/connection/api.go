package connection

import (
	"fmt"
	"strconv"

	openapi "github.com/2nth0nyj/go-trade/exchang/ctrader/connection/proto"
)

func (c *Connection) GetBalance() string {
	req := openapi.ProtoOATraderReq{CtidTraderAccountId: &c.ctid}
	res := c.SendProtoAndWaitResponse(&req)
	if m, ok := res.(*openapi.ProtoOATraderRes); ok {
		trader := m.Trader
		balance := trader.GetBalance()
		moneyDigits := trader.GetMoneyDigits()
		balanceString := strconv.FormatInt(balance, 10)
		if len(balanceString) == 1 {
			balanceString = "0" + balanceString
		}
		m := ""
		for i, d := range balanceString {
			if i == (len(balanceString) - int(moneyDigits)) {
				if i == 0 {
					m += "0"
				}
				m += "."
			}
			m += string(d)
		}
		return m
	} else {
		fmt.Printf("get oa trader error: res: %v\n", res)
	}
	return "0.00"
}
