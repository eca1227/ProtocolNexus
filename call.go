package main

import (
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"time"
)

func (a *App) LogPrint(printHandle, dataType, dataText string) {
	runtime.EventsEmit(a.ctx, printHandle, map[string]interface{}{
		"dataType": dataType,
		"dataText": fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05.000"), dataText),
	})
}

func (a *App) DisConn(connLineData ...string) {
	clt := len(connLineData)
	switch clt {
	case 0:
		runtime.EventsEmit(a.ctx, "disconnCommander")
	default:
		return
	}
}
