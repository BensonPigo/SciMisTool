// routing_key.go
package mq

// RoutingKey 是我們自訂的字串型別
type RoutingKey string

// 一組可用的 RoutingKey 列舉
const (
	RoutingKeyDDL RoutingKey = "bi_ddl.key"
	RoutingKeyDML RoutingKey = "bi_dml.key"
	// 若日後擴充，只要在這裡新增一行
)

// AllRoutingKeys 幫助批次綁定時迭代
var AllRoutingKeys = []RoutingKey{
	RoutingKeyDDL,
	RoutingKeyDML,
}
