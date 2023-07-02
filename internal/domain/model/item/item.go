package item

import "github.com/PavelKhripkov/object_storage/internal/domain/model/item_data"

type Item struct {
	id       string
	name     string
	path     string
	metadata map[string]string
	bucketId string
	data     item_data.ItemData
}
