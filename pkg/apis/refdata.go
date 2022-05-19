package apis

type ConfigMapRef struct {
	NameSpace string `json:"nameSpace"`
	Name      string `json:"name"`
}

type SecretRef struct {
	NameSpace string `json:"nameSpace"`
	Name      string `json:"name"`
}

type PodRef struct {
	NameSpace string `json:"nameSpace"`
	Name      string `json:"name"`
}
