package websocket

import (
	"sync"

	"github.com/eolinker/eosc/eocontext"
	websocket_context "github.com/eolinker/eosc/eocontext/websocket-context"
	"github.com/eolinker/eosc/log"
)

type Finisher struct {
}

func (f *Finisher) Finish(org eocontext.EoContext) error {
	ctx, err := websocket_context.Assert(org)
	if err != nil {
		return err
	}

	clientConn := ctx.ClientConn()
	upstreamConn := ctx.UpstreamConn()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msgType, msg, err := clientConn.ReadMessage()
			if err != nil {
				log.Error("read:", err)
				break
			}
			err = upstreamConn.WriteMessage(msgType, msg)
			if err != nil {
				log.Error("write message error: ", err)
				break
			}
		}

	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msgType, msg, err := upstreamConn.ReadMessage()
			if err != nil {
				log.Error("read upstream message err: ", err)
				return
			}
			err = clientConn.WriteMessage(msgType, msg)
			if err != nil {
				log.Error("write client message err: ", err)
				return
			}
		}
	}()
	wg.Wait()

	return nil
}
