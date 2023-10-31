package xbun

// OrderExpr unifies the construction of ORDER BY expressions.
// If direction is not specified, it defaults to OrderAsc.
func OrderExpr(field string, dir OrderDir) string {
	switch dir {
	case "":
		return field + " " + string(OrderAsc)
	case OrderAsc, OrderDesc:
		return field + " " + string(dir)
	default:
		panic("invalid order direction") // this should never happen
	}
}
