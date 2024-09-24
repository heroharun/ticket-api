package ticket

//{
//"name": "example",
//"desc": "sample description",
//"allocation": 100
//}

type TicketOption struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`

	Allocation int `json:"allocation"`
}

// {
// "quantity": 2,
// "user_id": "406c1d05-bbb2-4e94-b183-7d208c2692e1"
// }

type Purchase struct {
	Quantity       int    `json:"quantity"`
	UserID         string `json:"user_id"`
	TicketOptionID string
}
