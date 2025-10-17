package backend

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// --- Telnet 매니저 싱글턴 ---

var (
	telnetManagerOnce sync.Once
	managerTelnet     *telnetManager
)

// TelnetSession은 하나의 Telnet 연결 상태를 저장
type TelnetSession struct {
	Conn   net.Conn
	Reader *bufio.Reader
}

// telnetManager 구조체 (private)
type telnetManager struct {
	sessions map[string]*TelnetSession
	mu       sync.Mutex
}

// getTelnetManager는 Telnet 매니저 인스턴스를 반환
func getTelnetManager() *telnetManager {
	telnetManagerOnce.Do(func() {
		managerTelnet = &telnetManager{
			sessions: make(map[string]*TelnetSession),
		}
		fmt.Println("TelnetManager가 생성되었습니다.")
	})
	return managerTelnet
}

// --- 공개 함수 ---
func TelnetConnect(ip string) error {
	return getTelnetManager().connect(ip)
}

func TelnetDisconnect(ip string) error {
	return getTelnetManager().disconnect(ip)
}

func TelnetSendData(ip string, command string) (string, error) {
	return getTelnetManager().sendData(ip, command)
}

// --- 비공개 메소드 ---
func (m *telnetManager) connect(ip string) (err error) {
	m.mu.Lock()
	if _, ok := m.sessions[ip]; ok {
		m.mu.Unlock()
		return nil
	}
	m.sessions[ip] = nil
	m.mu.Unlock()

	defer func() {
		if err != nil {
			m.mu.Lock()
			if session, ok := m.sessions[ip]; ok && session == nil {
				delete(m.sessions, ip)
			}
			m.mu.Unlock()
		}
	}()
	for i := 0; i < 3; i++ {
		fmt.Printf("Telnet 연결 시도 #%d for %s...\n", i+1, ip)
		var conn net.Conn
		conn, err = net.DialTimeout("tcp", ip, 2*time.Second)
		if err != nil {
			time.Sleep(200 * time.Millisecond) // 실패 시 잠시 대기
			continue                           // 다음 시도
		}

		// 연결에 성공하면, 실패 시 conn을 닫도록 defer 설정
		success := false
		defer func() {
			if !success {
				conn.Close()
			}
		}()

		reader := bufio.NewReader(conn)

		if err = skipTelnetNegotiation(conn, reader, 2*time.Second); err != nil {
			fmt.Printf("협상 실패: %v\n", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}

		if _, err = readUntil(conn, reader, "Password: ", 2*time.Second); err != nil {
			fmt.Printf("'Password:' 대기 실패: %v\n", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		conn.Write([]byte("Help\r\n"))

		if _, err = readUntil(conn, reader, "GPL:", 2*time.Second); err != nil {
			fmt.Printf("'GPL:' 대기 실패: %v\n", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}

		// 모든 과정 성공!
		m.mu.Lock()
		if _, ok := m.sessions[ip]; ok && m.sessions[ip] == nil {
			session := &TelnetSession{Conn: conn, Reader: reader}
			m.sessions[ip] = session
			fmt.Printf("%s 에 성공적으로 연결 및 인증되었습니다.\n", ip)
			success = true // 성공했으므로 defer에서 conn을 닫지 않도록 플래그 설정
			m.mu.Unlock()
			return nil // 에러 없이 함수 종료
		} else {
			// 그 사이에 연결 해제 요청이 들어온 경우
			err = fmt.Errorf("[%s] 연결 과정 중 외부에서 연결 해제 요청이 있었습니다", ip)
			m.mu.Unlock()
			return err
		}
	}

	return err
}

func (m *telnetManager) disconnect(ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[ip]
	if !ok {
		return nil // 이미 연결 끊김
	}

	if session.Conn != nil {
		if tcpConn, ok := session.Conn.(*net.TCPConn); ok {
			tcpConn.SetLinger(0)
		}
		session.Conn.Close()
	}

	delete(m.sessions, ip)
	fmt.Printf("%s Telnet 연결이 해제되었습니다.\n", ip)
	return nil
}

// sendData는 명령어를 보내고 응답을 기다려 반환
func (m *telnetManager) sendData(ip string, command string) (string, error) {
	m.mu.Lock()
	session, ok := m.sessions[ip]
	if !ok {
		m.mu.Unlock()
		return "", fmt.Errorf("%s 는 연결되어 있지 않습니다", ip)
	}
	m.mu.Unlock()

	clearInitialBuffer(session.Conn)

	if _, err := session.Conn.Write([]byte(command + "\r\n")); err != nil {
		m.disconnect(ip)
		return "", fmt.Errorf("[%s] 명령어 전송 실패: %v", ip, err)
	}
	fmt.Printf("[%s] Telnet 명령어 전송: '%s'\n", ip, command)

	// 응답 읽기
	response, err := m.readAllResponse(session, 5*time.Second)
	if err != nil {
		m.disconnect(ip)
		return "", fmt.Errorf("[%s] 응답 읽기 실패: %v", ip, err)
	}

	idx := strings.Index(response, "\n")
	if idx != -1 {
		response = response[idx+1:]
	} else {
		response = ""
	}

	fmt.Printf("[%s] Telnet 응답 수신\n", ip)
	return strings.TrimSpace(response), nil
}

func clearInitialBuffer(conn net.Conn) error {
	// 버퍼를 비우는 작업은 최대 0.5초만 시도
	overallDeadline := time.Now().Add(500 * time.Millisecond)
	readBuf := make([]byte, 1024) // 데이터를 읽어 담아둘 임시 버퍼

	for time.Now().Before(overallDeadline) {
		// 1. 매우 짧은 읽기 타임아웃 설정 (50ms)
		err := conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		if err != nil {
			return err
		}

		// 2. 데이터 읽기 시도
		_, err = conn.Read(readBuf)
		if err != nil {
			// 3. 타임아웃 에러가 발생하면, 이는 "더 이상 읽을 데이터가 없음"을 의미하므로 정상 종료
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// 최종적으로 ReadDeadline을 초기화하여 이후 작업에 영향이 없도록 함
				conn.SetReadDeadline(time.Time{})
				return nil
			}
			if err == io.EOF {
				return nil
			}
			// 그 외의 에러는 실제 연결 문제일 수 있으므로 에러 반환
			return err
		}
		// 데이터가 읽혔다면, 버퍼에 더 있을 수 있으므로 루프 계속
	}
	return nil
}

/* 헬퍼 함수들의 전체 구현 내용 */
func (m *telnetManager) readAllResponse(session *TelnetSession, timeout time.Duration) (string, error) {
	var buf [4096]byte // 버퍼 크기를 넉넉하게
	var response strings.Builder
	deadline := time.Now().Add(timeout)
	search := []byte("GPL:")
	var leftover []byte

	for time.Now().Before(deadline) {
		session.Conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, err := session.Reader.Read(buf[:])
		if n > 0 {
			cur := append(leftover, buf[:n]...)
			if idx := bytes.Index(cur, search); idx >= 0 {
				response.Write(cur[:idx])
				break
			}
			response.Write(cur)
			if len(cur) >= 4 {
				leftover = cur[len(cur)-4:]
			} else {
				leftover = cur
			}
		}
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // 타임아웃은 정상적인 대기 상태이므로 계속 진행
			}
			return response.String(), err // 그 외의 에러는 반환
		}
	}
	return response.String(), nil
}

func readUntil(conn net.Conn, reader *bufio.Reader, delim string, timeout time.Duration) (string, error) {
	var buf strings.Builder
	delimLen := len(delim)
	conn.SetReadDeadline(time.Now().Add(timeout))
	defer conn.SetReadDeadline(time.Time{})

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		if b == 0xFF { // IAC
			if reader.Buffered() < 2 {
				continue
			}
			reader.Discard(2)
			continue
		}
		buf.WriteByte(b)

		if buf.Len() >= delimLen && buf.String()[buf.Len()-delimLen:] == delim {
			return buf.String(), nil
		}
	}
}

func skipTelnetNegotiation(conn net.Conn, reader *bufio.Reader, timeout time.Duration) error {
	conn.SetReadDeadline(time.Now().Add(timeout))
	defer conn.SetReadDeadline(time.Time{})

	for {
		b, err := reader.Peek(1)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// 데이터를 하나도 못받고 타임아웃 -> 협상할 내용이 없는 것으로 간주하고 성공 처리
				return nil
			}
			return err
		}

		if b[0] == 0xFF {
			if reader.Buffered() < 3 {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			reader.Discard(3)
		} else {
			return nil // 협상 코드가 아니면 즉시 종료
		}
	}
}
