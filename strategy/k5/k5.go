package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/2nth0nyj/go-trade/exchang/ctrader"
)

func main() {
	var waitGroup sync.WaitGroup

	waitGroup.Add(1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		defer waitGroup.Done()

		clientId := "2183_yYYwcVyw40LMplltGPCfbsdBajdICfdrNfSEHSYFkdFbSG4uw1"
		secret := "hHt2FxwSBY06bN0bHbI0ESofrLyyCWLha0wr3Gv8fXOBdwlc3s"
		accessToken := "X1YIVrgGoDmJK5Ky8qufZ-zAMtlw1jHvokJAFHmarYQ"
		var ctid int64 = 37723713
		ctrader := ctrader.NewClient(clientId, secret, accessToken, ctid, false)
		time.Sleep(time.Duration(5 * time.Second))
		b := ctrader.GetBalance()
		fmt.Printf("ctrader balance %v\n", b)
		<-c
	}()

	waitGroup.Wait()
}
