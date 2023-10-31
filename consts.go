package xbun

// OrderDir is the direction of sorting.
type OrderDir string

const (
	OrderAsc  OrderDir = "ASC"
	OrderDesc OrderDir = "DESC"
)

// Sub-condition separator constraints that you can use in bun's WhereGroup calls.
const (
	SepAND = " AND "
	SepOR  = " OR "
	SepNOT = " NOT "
)

// RETURNING clause argument constraints.
// Useful with Returning query option.
const (
	RetAll = "*"
)

// QueryFlag is a flag that you can use in QueryOptions.
type QueryFlag uint8

const (
	QueryFlagNone QueryFlag = iota
	QueryFlagWith
	QueryFlagOnly
)
