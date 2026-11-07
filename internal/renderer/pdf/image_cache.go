package pdf

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"sync"

	"certificate-generator/assets"

	"github.com/signintech/gopdf"
)

// cachedImage holds immutable image bytes plus a precomputed stable identifier.
// Sharing these across renders avoids re-reading the (multi-MB, 300 DPI) background
// from disk / embed.FS and re-hashing it for every single certificate in a batch,
// which previously generated ~6 MB of garbage + an MD5 pass per certificate.
type cachedImage struct {
	id   string
	data []byte
}

// sharedImageHolder implements gopdf.ImageHolder over shared, read-only bytes
// without copying them or recomputing a hash.
type sharedImageHolder struct {
	id string
	*bytes.Reader
}

func (h *sharedImageHolder) ID() string { return h.id }

func (c *cachedImage) holder() gopdf.ImageHolder {
	return &sharedImageHolder{id: c.id, Reader: bytes.NewReader(c.data)}
}

const maxImageCacheEntries = 16

var (
	imageCacheMu sync.Mutex
	imageCache   = make(map[string]*cachedImage)
)

func storeCachedImage(key string, data []byte) *cachedImage {
	sum := sha1.Sum(data)
	img := &cachedImage{id: fmt.Sprintf("%x", sum), data: data}

	imageCacheMu.Lock()
	defer imageCacheMu.Unlock()
	if existing, ok := imageCache[key]; ok {
		return existing
	}
	// Bounded cache: a simple reset is enough for a desktop app with a handful of templates.
	if len(imageCache) >= maxImageCacheEntries {
		imageCache = make(map[string]*cachedImage)
	}
	imageCache[key] = img
	return img
}

func lookupCachedImage(key string) (*cachedImage, bool) {
	imageCacheMu.Lock()
	defer imageCacheMu.Unlock()
	img, ok := imageCache[key]
	return img, ok
}

// loadDiskImage returns cached bytes for a file on disk. The cache key includes
// size and modification time so edits to a template asset are picked up immediately.
func loadDiskImage(path string) (*cachedImage, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory", path)
	}
	key := fmt.Sprintf("disk:%s:%d:%d", path, info.Size(), info.ModTime().UnixNano())
	if img, ok := lookupCachedImage(key); ok {
		return img, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return storeCachedImage(key, data), nil
}

// loadEmbeddedBackground returns cached bytes for an embedded background asset.
func loadEmbeddedBackground(name string) (*cachedImage, error) {
	key := "embed:" + name
	if img, ok := lookupCachedImage(key); ok {
		return img, nil
	}
	data, err := assets.ReadBackground(name)
	if err != nil {
		return nil, err
	}
	return storeCachedImage(key, data), nil
}
