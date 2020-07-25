/*
 * Author:slive
 * DATE:2020/7/24
 */
package logger

import "testing"

func TestDebug(t *testing.T) {
	Debug("测试debug...")
}

func TestInfo(t *testing.T) {
	Debug("测试info...")
}

func TestWarn(t *testing.T) {
	Debug("测试warn...")
}

func TestError(t *testing.T) {
	Debug("测试error...")
}
