/*
 * Author:slive
 * DATE:2020/7/17
 */
package tcp

import (
	"net/http"
)

type HttpConn struct {
	handleHttpFunc HandleHttpFunc
}

func StartHttpConn(handleHttpFunc HandleHttpFunc) *HttpConn {
	ch := &HttpConn{handleHttpFunc: handleHttpFunc}
	return ch
}

func (b *HttpConn) DoHttp(response http.ResponseWriter, request *http.Request) error {
	return b.handleHttpFunc(response, request)
}

type HandleHttpFunc func(response http.ResponseWriter, request *http.Request) error
