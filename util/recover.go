package util

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

func RecoverFunc() {
	if v := recover(); v != nil {
		fileLine := ""
		fileLineSlice := make([]string, 0)
		for i := 0; i < 10; i++ {
			if _, file, line, ok := runtime.Caller(i); ok && file != "" {
				if binPath, err := os.Getwd(); err == nil {
					file = strings.ReplaceAll(file, binPath, "")
				}
				fileLineSlice = append(fileLineSlice, fmt.Sprintf("%s:%d", file, line))
			}
		}
		if len(fileLineSlice) > 0 {
			fileLine = strings.Join(fileLineSlice, "  |  ")
		}
		fmt.Println(fileLine)
	}
}
