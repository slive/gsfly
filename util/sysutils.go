/*
 * Author:slive
 * DATE:2020/7/30
 */
package util

import "os"

// GetPwd 获取当前目录路径
func GetPwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		return "/"
	}
	return pwd
}
