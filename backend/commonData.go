package backend

import (
	"os"
	"path/filepath"
)

var ProgramFolderPath string

func DataSetProcess() {
	// 현재 실행 파일의 전체 경로
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	// 디렉토리 경로만 추출
	ProgramFolderPath = filepath.Join(filepath.Dir(exePath), "ProtocolNexus")

}
