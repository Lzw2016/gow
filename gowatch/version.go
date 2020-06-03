package gowatch

import (
	"fmt"
)

const version = "v0.0.2"

// 打印当前版本
func PrintVersion() {
	fmt.Printf("version: %s", version)
}
