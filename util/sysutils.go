/*
 * Author:slive
 * DATE:2020/7/30
 */
package util

import (
	"bufio"
	"hash/crc32"
	"io"
	"os"
	"strings"
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

func LoadProperties(propPath string) map[string]string {
	config := make(map[string]string)
	f, err := os.Open(propPath)
	defer f.Close()
	if err != nil {
		panic(err)
	}

	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		s := strings.TrimSpace(string(b))
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}
		key := strings.TrimSpace(s[:index])
		if len(key) == 0 {
			continue
		}
		value := strings.TrimSpace(s[index+1:])
		if len(value) == 0 {
			continue
		}
		config[key] = value
	}
	return config
}

func CheckFileExist(filename string) bool {
	exist := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
