package version

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/projectdiscovery/pdtm/pkg/types"
)

// writeFakeTool drops an executable script at basePath/name that prints script
// to stdout and exits with code. Returns the base dir.
func writeFakeTool(t *testing.T, name, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake-tool shell scripts are not executable on windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatalf("write fake tool: %v", err)
	}
	return dir
}

func TestExtractInstalledVersion(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		want    string
		wantErr error
	}{
		{
			name:   "supports --version",
			script: `[ "$1" = "--version" ] && echo "mytool v1.2.3" || exit 1`,
			want:   "1.2.3",
		},
		{
			name:   "only supports version subcommand",
			script: `[ "$1" = "version" ] && echo "Current version 2.0.1" || exit 2`,
			want:   "2.0.1",
		},
		{
			name:    "installed but no version in output",
			script:  `echo "usage: mytool [flags]"`,
			wantErr: ErrVersionUnknown,
		},
		{
			name:    "installed but empty output and non-zero exit",
			script:  `exit 3`,
			wantErr: ErrVersionUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeFakeTool(t, "mytool", tt.script)
			got, err := ExtractInstalledVersion(types.Tool{Name: "mytool"}, dir)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want wrap of %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tt.want {
				t.Fatalf("version = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractInstalledVersionNotInstalled(t *testing.T) {
	dir := t.TempDir() // empty: binary does not exist
	_, err := ExtractInstalledVersion(types.Tool{Name: "missing-tool"}, dir)
	if !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("err = %v, want wrap of ErrNotInstalled", err)
	}
}
