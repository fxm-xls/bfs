package response

// csr
type CsrRes struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Attribute struct {
	Priority int    `json:"priority"`
	Status   int    `json:"status"`
	Sample   string `json:"sample"`
	Desc     string `json:"desc"`
}

type AttributeRes struct {
	Priority string `json:"priority"`
	Status   string `json:"status"`
	Sample   string `json:"sample"`
	Desc     string `json:"desc"`
}
