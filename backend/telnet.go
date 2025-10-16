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

// TelnetSession은 하나의 Telnet 연결 상태를 저장합니다.
type TelnetSession struct {
	Conn   net.Conn
	Reader *bufio.Reader
}

// telnetManager 구조체 (private)
type telnetManager struct {
	sessions map[string]*TelnetSession
	mu       sync.Mutex
}

// getTelnetManager는 Telnet 매니저 인스턴스를 반환합니다.
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

	fmt.Printf("Telnet 연결 시도 for %s...\n", ip)
	var conn net.Conn
	conn, err = net.DialTimeout("tcp", ip, 2*time.Second)
	if err != nil {
		// Dial 실패 시에는 명확하게 연결 실패 에러만 반환
		return fmt.Errorf("%s Telnet 연결 실패: %v", ip, err)
	}

	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	if err = clearInitialBuffer(conn); err != nil {
		// 각 단계에서 에러 발생 시, 명명된 반환 값 'err'에 할당된 후 defer가 실행됨
		return fmt.Errorf("[%s] 초기 버퍼 클리어 실패: %v", ip, err)
	}

	reader := bufio.NewReader(conn)

	if err = skipTelnetNegotiation(conn, reader, 2*time.Second); err != nil {
		return fmt.Errorf("[%s] 협상 실패: %v", ip, err)
	}

	_, err = readUntil(conn, reader, "Password: ", 1*time.Second)
	if err != nil {
		return fmt.Errorf("[%s] 'Password:' 프롬프트 대기 실패: %v", ip, err)
	}
	conn.Write([]byte("Help\r\n"))

	if _, err = readUntil(conn, reader, "GPL:", 1*time.Second); err != nil {
		// GPL 프롬프트가 없어도 인증은 성공한 것으로 간주하고 경고만 출력할 수 있음
		// fmt.Printf("[%s] 경고: 'GPL:' 프롬프트를 찾지 못했지만 연결은 계속합니다.\n", ip)
		// 또는 기존처럼 에러 처리
		return fmt.Errorf("[%s] 'GPL:' 프롬프트 대기 실패: %v", ip, err)
	}

	m.mu.Lock()
	if _, ok := m.sessions[ip]; ok && m.sessions[ip] == nil {
		session := &TelnetSession{Conn: conn, Reader: reader}
		m.sessions[ip] = session
		fmt.Printf("%s 에 성공적으로 연결 및 인증되었습니다.\n", ip)
	} else {
		err = fmt.Errorf("[%s] 연결 과정 중 외부에서 연결 해제 요청이 있었습니다", ip)
	}
	m.mu.Unlock()

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

// sendData는 명령어를 보내고 응답을 기다려 반환합니다.
func (m *telnetManager) sendData(ip string, command string) (string, error) {
	m.mu.Lock()
	session, ok := m.sessions[ip]
	if !ok {
		m.mu.Unlock()
		return "", fmt.Errorf("%s 는 연결되어 있지 않습니다", ip)
	}
	m.mu.Unlock()

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
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

		b, err := reader.ReadByte()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return "", err
		}

		if b == 0xFF {
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
	return "", fmt.Errorf("readUntil 시간 초과 (%v)", timeout)
}

func skipTelnetNegotiation(conn net.Conn, reader *bufio.Reader, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		b, err := reader.ReadByte()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // 타임아웃 다시 시도
			}
			if err == io.EOF {
				return nil // EOF면 정상 종료
			}
			continue // 그 외 에러는 무시하고 루프 유지
		}
		if b == 0xFF {
			if reader.Buffered() >= 2 {
				reader.Discard(2)
			}
			continue
		}
		return nil
	}
	return fmt.Errorf("skipTelnetNegotiation 시간 초과 (%v)", timeout)
}
