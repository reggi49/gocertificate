package domain

// StorageInfo represents local disk and database statistics.
type StorageInfo struct {
	DataDir          string `json:"dataDir"`
	DatabaseSizeByte int64  `json:"databaseSizeByte"`
	ProjectsSizeByte int64  `json:"projectsSizeByte"`
	TotalProjects    int    `json:"totalProjects"`
	ActiveProjects   int    `json:"activeProjects"`
	ArchivedProjects int    `json:"archivedProjects"`
}
