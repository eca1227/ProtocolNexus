package backend

import (
	"bytes"
	"errors"
	"fmt"
	"go.bug.st/serial"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

func FindSerialPort() []string {
	ports, err := serial.GetPortsList()
	if err != nil || len(ports) == 0 {
		fmt.Println("시리얼 포트를 찾는 데 실패했습니다")
		return []string{}
	}

	return ports
}

var (
	managerOnce   sync.Once
	managerSerial *serialManager
)

type serialManager struct {
	connections   map[string]serial.Port
	buffers       map[string]*bytes.Buffer
	disconnecting map[string]bool
	mu            sync.RWMutex
}

func getSerialManager() *serialManager {
	managerOnce.Do(func() {
		managerSerial = &serialManager{
			connections:   make(map[string]serial.Port),
			buffers:       make(map[string]*bytes.Buffer),
			disconnecting: make(map[string]bool),
		}
		fmt.Println("SerialManager가 생성되었습니다.")
	})
	return managerSerial
}

// --- 공개 함수 ---
func SerialConnect(portName string, baudRate int, onDataReceived func(port, dataType, data string)) error {
	return getSerialManager().connect(portName, baudRate, onDataReceived)
}

func SerialDisconnect(portName string) error {
	return getSerialManager().disconnect(portName)
}

func SerialSendData(portName, data string) error {
	return getSerialManager().sendData(portName, data)
}

// --- 비공개 함수 ---
func (m *serialManager) connect(portName string, baudRate int, onDataReceived func(port, dataType, data string)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.connections[portName]; ok {
		return nil
	}

	mode := &serial.Mode{BaudRate: baudRate, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return fmt.Errorf("%s 포트 열기 실패: %v", portName, err)
	}

	m.connections[portName] = port
	m.buffers[portName] = new(bytes.Buffer)
	m.disconnecting[portName] = false
	fmt.Printf("%s 포트에 성공적으로 연결되었습니다.\n", portName)

	go m.startReading(portName, onDataReceived, port)

	return nil
}

func (m *serialManager) startReading(portName string, onDataReceived func(port, dataType, data string), port serial.Port) {
	buff := make([]byte, 512)
	for {
		m.mu.RLock()
		_, isConnected := m.connections[portName]
		m.mu.RUnlock()
		if !isConnected {
			break
		}

		port.SetReadTimeout(1 * time.Second)
		bytesRead, err := port.Read(buff)

		if bytesRead > 0 {
			m.mu.Lock()
			buffer, ok := m.buffers[portName]
			if ok && buffer != nil {
				buffer.Write(buff[:bytesRead])
			}
			m.mu.Unlock()
			// 버퍼 처리 함수 호출
			m.processBuffer(portName, onDataReceived)
		}
		if err != nil {
			if isTimeout(err) {
				continue
			}

			var errno syscall.Errno
			isHandleError := false
			if errors.As(err, &errno) {
				switch runtime.GOOS {
				case "windows":
					if errno == 6 { // ERROR_INVALID_HANDLE
						isHandleError = true
					}
				case "linux", "darwin":
					if errno == syscall.EBADF { // EBADF (코드 9)
						isHandleError = true
					}
				}
			}
			m.mu.Lock()
			isDisconnecting, _ := m.disconnecting[portName]
			m.mu.Unlock()
			if !isDisconnecting && !isHandleError && err != io.EOF {
				onDataReceived(portName, "ERRO", err.Error())
			}
			break
		}
	}

	_ = m.disconnect(portName)
}

func (m *serialManager) disconnect(portName string) error {
	m.mu.Lock()
	m.disconnecting[portName] = true
	port, ok := m.connections[portName]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("%s 포트는 연결되어 있지 않습니다", portName)
	}

	err := port.Close()
	if runtime.GOOS == "windows" {
		time.Sleep(100 * time.Millisecond)
	}

	m.mu.Lock()
	delete(m.buffers, portName)
	delete(m.connections, portName)
	m.mu.Unlock()
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
		return fmt.Errorf("%s 포트 닫기 실패: %v", portName, err)
	}

	fmt.Printf("%s 포트 연결이 해제되었습니다.\n", portName)
	return nil
}

func (m *serialManager) sendData(portName, data string) error {
	m.mu.RLock()
	port, ok := m.connections[portName]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s 포트는 연결되어 있지 않습니다", portName)
	}

	_, err := port.Write([]byte(data + "\r\n"))
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, os.ErrClosed) {
			// 포트가 이미 닫힌 상태
			go m.disconnect(portName)
		}
		return fmt.Errorf("%s 데이터 전송 실패: %v", portName, err)
	}
	fmt.Printf("[%s] 데이터 전송: '%s'\n", portName, data)
	return nil
}

func isTimeout(err error) bool {
	// OS 수준의 타임아웃 오류
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}

	// net.Error 인터페이스를 구현하는 타임아웃 오류
	type timeout interface {
		Timeout() bool
	}
	var e timeout
	if errors.As(err, &e) && e.Timeout() {
		return true
	}
	return false
}

func (m *serialManager) processBuffer(portName string, onDataReceived func(port, dataType, data string)) {
	m.mu.Lock()
	buffer, ok := m.buffers[portName]
	m.mu.Unlock()

	if !ok || buffer == nil {
		return
	}

	delimiter := []byte("\r\n")

	for {
		m.mu.Lock()
		bufferBytes := buffer.Bytes()
		index := bytes.Index(bufferBytes, delimiter)
		if index == -1 {
			m.mu.Unlock()
			break
		}

		messageBytes := make([]byte, index+len(delimiter))
		_, err := buffer.Read(messageBytes) // 버퍼에서 읽기 (읽은 부분은 버퍼에서 제거)
		isDisconnecting, _ := m.disconnecting[portName]
		m.mu.Unlock()

		if isDisconnecting {
			break
		} else {
			if err != nil && err != io.EOF {
				onDataReceived(portName, "ERRO", fmt.Sprintf("버퍼 읽기 오류: %v", err))
				continue
			}
		}

		message := strings.TrimSpace(string(messageBytes[:index]))
		if len(message) > 0 {
			onDataReceived(portName, "RECV", message)
		}
	}
}
