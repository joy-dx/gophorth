package netdto

type NetClientType string

type NetClient struct {
	Name        string        `json:"name" yaml:"name"`
	Ref         string        `json:"ref" yaml:"ref"`
	ClientType  NetClientType `json:"client_type" yaml:"client_type"`
	Description string        `json:"description" yaml:"description"`
}
