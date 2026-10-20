package repo

import (
	"path/filepath"
	"testing"

	"github.com/algorandfoundation/ARCs/arckit/internal/config"
	"github.com/algorandfoundation/ARCs/arckit/internal/testutil"
)

func writeConfig(t *testing.T, root string, content string) {
	t.Helper()
	testutil.WriteTrimmedFile(t, filepath.Join(root, config.FileName), content)
}

func writeVettedAdopters(t *testing.T, root string, content string) {
	t.Helper()
	testutil.WriteTrimmedFile(t, filepath.Join(root, "adoption", "vetted-adopters.yaml"), content)
}
