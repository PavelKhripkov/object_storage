package item_data

type PartInfo struct {
	fileServerID string
	path         string
	size         int64
	md5          string
}

type Part interface {
	GetInfo() PartInfo
	GetData() []byte
}

type ItemData struct {
	data []Part
}
