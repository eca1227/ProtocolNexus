package backend

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// --- TCP 매니저 싱글턴 ---
var (
	tcpManagerOnce sync.Once
	managerTCP     *tcpManager
)

// tcpManager 구조체 (private)
type tcpManager struct {
	connections map[string]net.Conn
	mu          sync.Mutex
}

// getTCPManager 함수는 TCP 매니저 인스턴스를 반환합니다.
func getTCPManager() *tcpManager {
	tcpManagerOnce.Do(func() {
		managerTCP = &tcpManager{
			connections: make(map[string]net.Conn),
		}
		fmt.Println("TCPManager가 생성되었습니다.")
	})
	return managerTCP
}

// --- 공개 함수 ---
func TCPConnect(ip string, onDataReceived func(addr, dataType, data string)) error {
	return getTCPManager().connect(ip, onDataReceived)
}

func TCPDisconnect(ip string) error {
	return getTCPManager().disconnect(ip)
}

func TCPSendData(ip string, data string) error {
	return getTCPManager().sendData(ip, data)
}

// --- 비공개 메소드 (실제 로직) ---

func (m *tcpManager) connect(addr string, onDataReceived func(addr, dataType, data string)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.connections[addr]; ok {
		return nil
	}

	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return fmt.Errorf("%s TCP 연결 실패: %v", addr, err)
	}

	m.connections[addr] = conn
	fmt.Printf("%s 에 성공적으로 연결되었습니다.\n", addr)

	// 연결 성공 직후, 실시간 데이터 수신을 위한 고루틴을 시작
	go m.startReading(addr, conn, onDataReceived)

	return nil
}

func (m *tcpManager) disconnect(addr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, ok := m.connections[addr]
	if !ok {
		return nil
	}

	err := conn.Close()
	delete(m.connections, addr)
	if err != nil {
		return fmt.Errorf("%s 연결 해제 실패: %v", addr, err)
	}

	fmt.Printf("%s 연결이 해제되었습니다.\n", addr)
	return nil
}

func (m *tcpManager) sendData(addr string, data string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, ok := m.connections[addr]
	if !ok {
		return fmt.Errorf("%s 는 연결되어 있지 않습니다", addr)
	}

	_, err := conn.Write([]byte(data + "\r\n"))
	if err != nil {
		return fmt.Errorf("%s 데이터 전송 실패: %v", addr, err)
	}
	fmt.Printf("[%s] 데이터 전송: '%s'\n", addr, data)
	return nil
}

// startReading은 내부적으로만 사용
func (m *tcpManager) startReading(addr string, conn net.Conn, onDataReceived func(addr, dataType, data string)) {
	buff := make([]byte, 4096)
	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		bytesRead, err := conn.Read(buff)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // 타임아웃이면 재시도
			}

			m.disconnect(addr)
			// io.EOF는 정상적인 연결 종료
			if !(errors.Is(err, net.ErrClosed) || err == io.EOF) {
				fmt.Printf("[%s] 데이터 읽기 오류: %v\n", addr, err)
				onDataReceived(addr, "ERRO", "연결이 비정상적으로 종료되었습니다: "+err.Error())
			}
			return
		}
		if bytesRead > 0 {
			receivedData := string(buff[:bytesRead])
			// 콜백 함수에 주소(addr)도 함께 전달하여 어느 서버에서 온 데이터인지 구분
			onDataReceived(addr, "RECV", receivedData)
		}
	}
}
