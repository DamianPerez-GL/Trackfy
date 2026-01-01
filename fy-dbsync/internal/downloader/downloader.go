package downloader

import "context"

// Downloader interfaz para los descargadores de bases de datos
type Downloader interface {
	Name() string
	Download(ctx context.Context) error
	GetStats() map[string]interface{}
}
