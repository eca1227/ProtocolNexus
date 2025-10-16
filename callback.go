package main

import (
	"ProtocolNexus/backend"
	"fmt"
	"net"
	"strconv"
	"sync"
)

var connTypeList = make(map[string]int)
var connTypeMu sync.Mutex

// connType {1 == Serial, 2 == TCP, 3 == Telnet}
func (a *App) SendData(address, address2, data string) {
	if net.ParseIP(address) != nil {
		address = fmt.Sprintf("%s:%s", address, address2)
	}

	connTypeMu.Lock()
	connType, ok := connTypeList[address]
	connTypeMu.Unlock()

	if !ok {
		fmt.Println("Connection type not found for:", address)
		return
	}

	var err error
	switch connType {
	case 1:
		err = backend.SerialSendData(address, data)
	case 2:
		err = backend.TCPSendData(address, data)
	case 3:
		telData, err := backend.TelnetSendData(address, data)
		if err != nil {
			a.LogPrint("CommanderLog", "error", err.Error())
			a.DisConn()
			return
		}

		a.LogPrint("CommanderLog", "sent", data)
		a.LogPrint("CommanderLog", "received", telData)
		return
	default:
		fmt.Println("Unsupported connection type:", connType)
		return
	}

	if err != nil {
		a.LogPrint("CommanderLog", "error", err.Error())
	} else {
		a.LogPrint("CommanderLog", "sent", data)
	}
}

// connType = {-1 == Serial Disconnect, -2 == TCP Disconnect, 1 == Serial, 2 == TCP}
func (a *App) CommanderConn(connType int, address, address2 string) bool {
	ct := connType
	if ct < 0 {
		ct = -ct
	}
	if ct > 1 {
		address = fmt.Sprintf("%s:%s", address, address2)
	}
	var err error
	switch connType {
	case -1:
		err = backend.SerialDisconnect(address)
	case -2:
		err = backend.TCPDisconnect(address)
	case -3:
		err = backend.TelnetDisconnect(address)
	case 1:
		aNum, ok := strconv.Atoi(address2)
		if ok != nil {
			return false
		}
		dataHandler := func(port string, data string) {
			a.LogPrint("CommanderLog", "received", data)
		}
		err = backend.SerialConnect(address, aNum, dataHandler)
		if err != nil {
			return false
		}
	case 2:
		dataHandler := func(addr, dataType, data string) {
			a.LogPrint("CommanderLog", dataType, data)
			if dataType != "received" {
				a.DisConn()
			}
		}
		err = backend.TCPConnect(address, dataHandler)
		if err != nil {
			return false
		}
	case 3:
		err = backend.TelnetConnect(address)
		if err != nil {
			return false
		}
	default:
		return false
	}
	if err != nil {
		a.LogPrint("CommanderLog", "error", err.Error())
	}
	connTypeMu.Lock()
	defer connTypeMu.Unlock()
	if connType > 0 {
		a.LogPrint("CommanderLog", "info", fmt.Sprintf("%s Connected", address))
		connTypeList[address] = connType
		return true
	} else {
		a.LogPrint("CommanderLog", "info", fmt.Sprintf("%s Disconnected", address))
		delete(connTypeList, address)
		return false
	}
}

func (a *App) SerialList() []string {
	return backend.FindSerialPort()
}

func (a *App) SetPage(page string) {
	connTypeMu.Lock()
	defer connTypeMu.Unlock()
	var err error
	for k, v := range connTypeList {
		fmt.Println(k, v)
		switch v {
		case 1:
			err = backend.SerialDisconnect(k)
		case 2:
			err = backend.TCPDisconnect(k)
		case 3:
			err = backend.TelnetDisconnect(k)
		}
		delete(connTypeList, k)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	switch page {
	case "Commander":
		fmt.Println("C")
	case "EFEM Test":
		fmt.Println("E")
	case "LP Maint":
		fmt.Println("L")
	}
}
