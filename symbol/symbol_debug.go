package symbol

import (
	"fmt"
)

func DebugDumpTable() {
	for id, s := range table {
		fmt.Printf(":: id:%d\n", id)
		fmt.Printf(":: tag:%d\n", s.tag)
		fmt.Printf("[variable] BlockId:%d\n", s.variable.BlockId)
		fmt.Printf("[variable] Name:%s\n", s.variable.Name)
		fmt.Printf("[variable] Offset:%d\n", s.variable.Offset)
	}
}
