package chunk

type CreateChunkDTO struct {
	ItemID       string
	Position     uint8
	FileServerID string
	FilePath     string
	Size         int64
}
