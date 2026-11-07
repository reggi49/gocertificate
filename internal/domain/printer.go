package domain

// PrinterInfo represents physical or virtual printer available in the operating system.
type PrinterInfo struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}
