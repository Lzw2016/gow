package gowatch

import (
	"fmt"
)

const version = "v0.0.4"

// 打印当前版本
func PrintVersion() {
	fmt.Printf("version: %s", version)
}
