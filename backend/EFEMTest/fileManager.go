package EFEMTest

import (
	. "ProtocolNexus/backend"
	"os"
	"path/filepath"
)

var filePath = filepath.Join(ProgramFolderPath, "EFEMTest")

func f() {
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return
	}
}
