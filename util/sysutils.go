/*
 * Author:slive
 * DATE:2020/7/30
 */
package util

import (
	"hash/crc32"
	"os"
)

// GetPwd 获取当前目录路径
func GetPwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		return "/"
	}
	return pwd
}

func Hashcode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}
