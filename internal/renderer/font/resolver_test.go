package font_test

import (
	"path/filepath"
	"testing"

	"certificate-generator/internal/renderer/font"
)

func TestFontResolver(t *testing.T) {
	fontDir := filepath.Join("..", "..", "..", "assets", "fonts")
	resolver := font.NewResolver(fontDir)

	fonts := resolver.AvailableFonts()
	if len(fonts) == 0 {
		t.Fatalf("Expected registered fonts in assets/fonts, got 0")
	}

	// 1. Direct family match
	name, path, ok := resolver.Resolve("Ubuntu")
	if !ok || name != "Ubuntu" || path == "" {
		t.Errorf("Failed to resolve Ubuntu: name=%s, ok=%v", name, ok)
	}

	// 2. Fallback sans-serif -> Ubuntu
	name, _, ok = resolver.Resolve("sans-serif")
	if !ok || name != "Ubuntu" {
		t.Errorf("Expected sans-serif to resolve to Ubuntu, got %s", name)
	}

	// 3. Fallback serif -> LiberationSerif
	name, _, ok = resolver.Resolve("serif")
	if !ok || name != "LiberationSerif" {
		t.Errorf("Expected serif to resolve to LiberationSerif, got %s", name)
	}

	// 4. Case insensitivity
	name, _, ok = resolver.Resolve("liberationserif")
	if !ok || name != "LiberationSerif" {
		t.Errorf("Expected liberationserif to resolve to LiberationSerif, got %s", name)
	}

	// 5. Sloop Scriptone
	name, _, ok = resolver.Resolve("Sloop Scriptone")
	if !ok || name != "SloopScriptOne" {
		t.Errorf("Expected 'Sloop Scriptone' to resolve to SloopScriptOne, got %s (ok=%v)", name, ok)
	}

	// 6. Trajan Pro Bold
	name, _, ok = resolver.Resolve("Trajan Pro Bold")
	if !ok || name != "TrajanPro-Bold" {
		t.Errorf("Expected 'Trajan Pro Bold' to resolve to TrajanPro-Bold, got %s (ok=%v)", name, ok)
	}

	// 7. Embedded font resolution without disk directory (Standalone binary simulation)
	emptyResolver := font.NewResolver("")
	eName, eBytes, eOk := emptyResolver.ResolveBytes("LiberationSerif")
	if !eOk || eName != "LiberationSerif" || len(eBytes) == 0 {
		t.Errorf("Expected embedded LiberationSerif to resolve from empty resolver, got name=%s, ok=%v, bytesLen=%d", eName, eOk, len(eBytes))
	}
	sName, sBytes, sOk := emptyResolver.ResolveBytes("Sloop Scriptone")
	if !sOk || len(sBytes) == 0 {
		t.Errorf("Expected embedded Sloop Scriptone to resolve, got name=%s, ok=%v", sName, sOk)
	}
}
