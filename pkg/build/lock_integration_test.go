//go:build integration

package build

import (
	"context"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestLockEnvironment checks that locking pins every package of the build
// environment to an exact version, without dropping any of the packages the
// configuration asked for.
func TestLockEnvironment(t *testing.T) {
	ctx := context.Background()

	p := filepath.Join("testdata", "build_configs", "sed") + ".yaml"

	b, err := New(
		ctx,
		WithConfig(p),
		WithArch("x86_64"),
		WithConfigFileRepositoryURL("https://github.com/wolfi-dev/os"),
		WithConfigFileRepositoryCommit("c0ffee"),
		WithExtraRepos([]string{"https://packages.wolfi.dev/os"}),
		WithExtraKeys([]string{"https://packages.wolfi.dev/os/wolfi-signing.rsa.pub"}),
	)
	if err != nil {
		t.Fatalf("New(%q) returned error %v, expected a build", p, err)
	}

	if err := b.Compile(ctx); err != nil {
		t.Fatalf("Compile(%q) returned error %v, expected nil", p, err)
	}

	want := slices.Clone(b.Configuration.Environment.Contents.Packages)

	if err := b.LockEnvironment(ctx); err != nil {
		t.Fatalf("LockEnvironment(%q) returned error %v, expected nil", p, err)
	}

	got := b.Configuration.Environment.Contents.Packages

	locked := make(map[string]string, len(got))
	for _, pkg := range got {
		name, version, ok := strings.Cut(pkg, "=")
		if !ok {
			t.Errorf("LockEnvironment(%q) left package %q unpinned, expected a name=version constraint", p, pkg)
			continue
		}
		locked[name] = version
	}

	for _, pkg := range want {
		if _, ok := locked[pkg]; !ok {
			t.Errorf("LockEnvironment(%q) dropped package %q, expected it in the locked list %v", p, pkg, got)
		}
	}
}
