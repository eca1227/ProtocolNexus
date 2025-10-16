package backend

import (
	"fmt"
	"net"
	"strings"
)

func EthernetConnection(name string) (bool, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return false, fmt.Errorf("네트워크 인터페이스 조회 실패: %w", err)
	}

	for _, iface := range interfaces {
		if iface.Name == name {
			if iface.Flags&net.FlagUp != 0 {
				return true, nil // 인터페이스를 찾았고, 활성화 상태임
			} else {
				return false, nil // 인터페이스를 찾았지만, 비활성화 상태임
			}
		}
	}

	return false, fmt.Errorf("'%s' 이름의 네트워크 인터페이스를 찾을 수 없습니다", name)
}

func FindEthernet() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("네트워크 인터페이스 조회 실패")
		return nil
	}

	fmt.Println("--- 존재하는 모든 이더넷 인터페이스 목록 ---")

	var interfaceList []string
	for _, iface := range interfaces {
		// 루프백 인터페이스 제외
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// MAC 주소(하드웨어 주소) 존재 확인
		if len(iface.HardwareAddr) == 0 {
			continue
		}

		// 이름 필터링
		lowerName := strings.ToLower(iface.Name)
		isEthernetByName := strings.Contains(lowerName, "ethernet") || strings.Contains(lowerName, "eth") || strings.Contains(iface.Name, "이더넷")

		if !isEthernetByName {
			continue
		}

		// 모든 필터링을 통과한 인터페이스 정보 출력
		fmt.Printf("이름: %s\n", iface.Name)
		fmt.Printf("  MAC 주소: %s\n", iface.HardwareAddr)
		fmt.Printf("  플래그: %s\n", iface.Flags) // 랜선이 연결 안됐으면 'up' 플래그가 없습니다.
		fmt.Println("------------------------------------")
		interfaceList = append(interfaceList, iface.Name)
	}
	return interfaceList
}
