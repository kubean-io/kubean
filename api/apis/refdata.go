package apis

type DataRef struct {
	NameSpace string `json:"namespace"`
	Name      string `json:"name"`
}

func (data *DataRef) IsEmpty() bool {
	if data == nil || len(data.Name) == 0 {
		return true
	}
	return false
}

type ConfigMapRef = DataRef

type SecretRef = DataRef

type PodRef = DataRef

type JobRef = DataRef
