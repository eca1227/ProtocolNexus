package main

import (
	"ProtocolNexus/backend"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"time"
)

func (a *App) LogPrint(printHandle, dataType, dataText string) {
	log := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05.000"), dataText)
	runtime.EventsEmit(a.ctx, printHandle, map[string]interface{}{
		"dataType": dataType,
		"dataText": log,
	})
	backend.LoggingList[printHandle].Log(fmt.Sprintf("[%-4s] %s", dataType, log))
}

func (a *App) CommanderDisconn(connLineData ...string) {
	clt := len(connLineData)
	switch clt {
	case 0:
		runtime.EventsEmit(a.ctx, "disconnCommander")
	default:
		return
	}
}
