package models

type Scenario struct {
	Name    string  `json:"m_Name"`
	TextMap TextMap `json:"textMap"`
}

type TextMap struct {
	IdToText idToText `json:"idToText"`
}

type idToText struct {
	Values []string `json:"values"`
}
