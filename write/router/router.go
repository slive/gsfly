/*
 * Author:slive
 * DATE:2020/7/21
 */
package router

import gconn "gsfly/connection"

type Router interface {

	PreSend(packet gconn.Packet) error
}
