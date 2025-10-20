package backend

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type AsyncLogger struct {
	logChan   chan string
	wg        sync.WaitGroup
	closeOnce sync.Once
}

var LoggingList = make(map[string]*AsyncLogger)

func logWorker(logFilePath string, logChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "비동기 로거 파일 열기 실패 (%s): %v\n", logFilePath, err)
		return
	}
	defer file.Close()

	logger := log.New(file, "", 0)
	logger.Println("------ 로그 기록 시작 ------")
	for msg := range logChan {
		logger.Println(msg)
	}
	logger.Println("------ 로그 기록 종료 ------")
}

// logFilePath: "C:/logs/comm1.txt" 또는 "./logs/app.log"
func NewLogger(logFilePath string) *AsyncLogger {
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil
	}

	logger := &AsyncLogger{
		logChan: make(chan string, 1000),
	}

	logger.wg.Add(1)
	go logWorker(logFilePath, logger.logChan, &logger.wg)

	fmt.Printf("로그 기능 활성화: %s\n", logFilePath)
	return logger
}

// 특정 로거 인스턴스의 로깅을 종료
func (l *AsyncLogger) Close() {
	// closeOnce.Do 함수가 여러 번 호출돼도 채널 닫기가 한 번만 실행되도록 보장
	l.closeOnce.Do(func() {
		close(l.logChan)
		l.wg.Wait()
		fmt.Println("로그 기능이 비활성화되었습니다.")
	})
}

// Log는 특정 로거 인스턴스에 로그를 전송
func (l *AsyncLogger) Log(wlog string) {
	// Close()가 호출된 후 l.logChan <- wlog가 실행되면 패닉이 발생가능
	// recover()로 패닉을 방지
	defer func() {
		if r := recover(); r != nil {
			// 'send on closed channel' 패닉 무시
		}
	}()

	// 채널 버퍼가 가득 찼으면 여기서 대기
	l.logChan <- wlog
}
