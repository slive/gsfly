/*
 * Author:slive
 * DATE:2020/7/17
 */
package handler

import (
	gconn "gsfly/connection"
)

type Handler interface {

	Handle(packet gconn.Packet)

}
