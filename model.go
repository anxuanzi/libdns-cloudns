package cloudns

type ApiDnsRecord struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Record   string `json:"record"`
	Failover string `json:"failover"`
	Ttl      string `json:"ttl"`
	Status   int    `json:"status"`
}

type ApiResponse struct {
	Status            string `json:"status"`
	StatusDescription string `json:"statusDescription"`
	Data              struct {
		Id int `json:"id"`
	} `json:"data,omitempty"`
}
