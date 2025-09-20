package symbol

import (
	"fmt"
)

func DebugDumpTable() {
	for i := 0; i < len(table); i++ {
		fmt.Printf("--> id:%d\n", table[i].id)
		fmt.Printf("[variable] blockId:%d\n", table[i].variable.blockId)
		fmt.Printf("[variable] name:%s\n", table[i].variable.name)
		fmt.Printf("[variable] offset:%d\n", table[i].variable.offset)
	}
}
