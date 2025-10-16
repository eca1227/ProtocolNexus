package backend

import (
	"fmt"
	"go.bug.st/serial"
	"io"
	"sync"
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
	managerOnce   sync.Once      // manager를 단 한 번만 생성하기 위한 장치
	managerSerial *serialManager // 애플리케이션 전역에서 사용될 단일 인스턴스
)

// SerialManager 구조체는 외부로 노출할 필요가 없으므로 소문자로 시작 (private)
type serialManager struct {
	connections map[string]io.ReadWriteCloser
	mu          sync.Mutex
}

// getSerialManager 함수는 managerTelnet 인스턴스를 반환합니다.
// 최초 호출 시에만 인스턴스를 생성합니다.
func getSerialManager() *serialManager {
	managerOnce.Do(func() {
		managerSerial = &serialManager{
			connections: make(map[string]io.ReadWriteCloser),
		}
		fmt.Println("SerialManager가 생성되었습니다.")
	})
	return managerSerial
}

// --- 공개 함수 ---
func SerialConnect(portName string, baudRate int, onDataReceived func(port string, data string)) error {
	return getSerialManager().connect(portName, baudRate, onDataReceived)
}

func SerialDisconnect(portName string) error {
	return getSerialManager().disconnect(portName)
}

func SerialSendData(portName, data string) error {
	return getSerialManager().sendData(portName, data)
}

// --- 비공개 함수 ---
func (m *serialManager) connect(portName string, baudRate int, onDataReceived func(port string, data string)) error {
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
	fmt.Printf("%s 포트에 성공적으로 연결되었습니다.\n", portName)

	go func() {
		// 이제 이 고루틴은 connect 메소드에 의해 시작됩니다.
		buff := make([]byte, 128)
		for {
			bytesRead, err := port.Read(buff)
			if err != nil {
				fmt.Printf("[%s] 데이터 읽기 중단: %v\n", portName, err)
				m.disconnect(portName) // 내부적으로 disconnect 호출
				return
			}
			if bytesRead > 0 {
				receivedData := string(buff[:bytesRead])
				onDataReceived(portName, receivedData)
			}
		}
	}()

	return nil
}

func (m *serialManager) disconnect(portName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	port, ok := m.connections[portName]
	if !ok {
		return fmt.Errorf("%s 포트는 연결되어 있지 않습니다", portName)
	}

	err := port.Close()
	delete(m.connections, portName) // 맵에서 연결 정보 삭제
	if err != nil {
		return fmt.Errorf("%s 포트 닫기 실패: %v", portName, err)
	}

	fmt.Printf("%s 포트 연결이 해제되었습니다.\n", portName)
	return nil
}

func (m *serialManager) sendData(portName, data string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	port, ok := m.connections[portName]
	if !ok {
		return fmt.Errorf("%s 포트는 연결되어 있지 않습니다", portName)
	}

	_, err := port.Write([]byte(data + "\r\n"))
	if err != nil {
		return fmt.Errorf("%s 포트로 데이터 전송 실패: %v", portName, err)
	}
	fmt.Printf("[%s] 데이터 전송: '%s'\n", portName, data)
	return nil
}
