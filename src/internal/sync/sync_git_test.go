package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCloneRepo(t *testing.T) {
	t.Parallel()

	originPath := createBareRepo(t)
	seedOriginWithCommit(t, originPath, "first")

	localCloneDir := t.TempDir()
	repo := RepoSync{Name: "origin", URL: originPath}
	options := SyncRepoOptions{LocalCloneDir: localCloneDir}

	if err := cloneRepo(repo, options); err != nil {
		t.Fatalf("cloneRepo returned error: %v", err)
	}

	clonedPath := filepath.Join(localCloneDir, "origin.git")
	if _, err := os.Stat(clonedPath); err != nil {
		t.Fatalf("cloned repository not found at %s: %v", clonedPath, err)
	}

	if output := runGit(t, clonedPath, "rev-parse", "--is-bare-repository"); strings.TrimSpace(output) != "true" {
		t.Fatalf("expected cloned repository to be bare, got %q", strings.TrimSpace(output))
	}
}

func TestValidateLocalRepo(t *testing.T) {
	t.Parallel()

	originPath := createBareRepo(t)
	seedOriginWithCommit(t, originPath, "first")

	localCloneDir := t.TempDir()
	repo := RepoSync{Name: "origin", URL: originPath}
	options := SyncRepoOptions{LocalCloneDir: localCloneDir}

	if err := cloneRepo(repo, options); err != nil {
		t.Fatalf("cloneRepo returned error: %v", err)
	}

	if err := validateLocalRepo(repo, options); err != nil {
		t.Fatalf("validateLocalRepo returned error: %v", err)
	}
}

func TestFetchRepo(t *testing.T) {
	t.Parallel()

	originPath := createBareRepo(t)
	seedOriginWithCommit(t, originPath, "first")

	localCloneDir := t.TempDir()
	repo := RepoSync{Name: "origin", URL: originPath}
	options := SyncRepoOptions{LocalCloneDir: localCloneDir}

	if err := cloneRepo(repo, options); err != nil {
		t.Fatalf("cloneRepo returned error: %v", err)
	}

	localMirrorPath := filepath.Join(localCloneDir, "origin.git")

	before := strings.TrimSpace(runGit(t, localMirrorPath, "rev-list", "--all", "--count"))

	addCommitToRepo(t, originPath, "second")

	if err := fetchRepo(repo, options); err != nil {
		t.Fatalf("fetchRepo returned error: %v", err)
	}

	after := strings.TrimSpace(runGit(t, localMirrorPath, "rev-list", "--all", "--count"))
	if before == after {
		t.Fatalf("expected fetch to update refs, commit count unchanged: before=%s after=%s", before, after)
	}
}

func TestPushToMirror(t *testing.T) {
	t.Parallel()

	originPath := createBareRepo(t)
	seedOriginWithCommit(t, originPath, "first")

	localCloneDir := t.TempDir()
	repo := RepoSync{Name: "origin", URL: originPath}
	options := SyncRepoOptions{LocalCloneDir: localCloneDir}

	if err := cloneRepo(repo, options); err != nil {
		t.Fatalf("cloneRepo returned error: %v", err)
	}

	if err := fetchRepo(repo, options); err != nil {
		t.Fatalf("fetchRepo returned error: %v", err)
	}

	mirrorPath := createBareRepo(t)
	mirror := RepoSync{Name: "mirror", URL: mirrorPath}

	if err := pushToMirror(repo, mirror, options); err != nil {
		t.Fatalf("pushToMirror returned error: %v", err)
	}

	count := strings.TrimSpace(runGit(t, mirrorPath, "rev-list", "--all", "--count"))
	if count == "0" {
		t.Fatalf("expected mirrored repository to have commits")
	}
}

func createBareRepo(t *testing.T) string {
	t.Helper()

	repoPath := filepath.Join(t.TempDir(), "repo.git")
	runGit(t, "", "init", "--bare", repoPath)
	return repoPath
}

func seedOriginWithCommit(t *testing.T, originPath string, marker string) {
	t.Helper()

	workRoot := t.TempDir()
	workDir := filepath.Join(workRoot, "work")
	runGit(t, "", "clone", originPath, workDir)
	runGit(t, workDir, "config", "user.email", "test@example.com")
	runGit(t, workDir, "config", "user.name", "Test User")
	checkoutMainBranch(t, workDir)

	filePath := filepath.Join(workDir, "README.md")
	content := []byte(fmt.Sprintf("commit-%s\n", marker))
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}

	runGit(t, workDir, "add", ".")
	runGit(t, workDir, "commit", "-m", "seed "+marker)
	runGit(t, workDir, "push", "origin", "HEAD:main")
}

func addCommitToRepo(t *testing.T, path string, marker string) {
	t.Helper()

	workRoot := t.TempDir()
	workDir := filepath.Join(workRoot, "work")
	runGit(t, "", "clone", path, workDir)
	runGit(t, workDir, "config", "user.email", "test@example.com")
	runGit(t, workDir, "config", "user.name", "Test User")
	checkoutMainBranch(t, workDir)

	filePath := filepath.Join(workDir, "README.md")
	content := []byte(fmt.Sprintf("commit-%s\n", marker))
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}

	runGit(t, workDir, "add", ".")
	runGit(t, workDir, "commit", "-m", "seed "+marker)
	runGit(t, workDir, "push", "origin", "HEAD:main")
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
	return string(output)
}

func checkoutMainBranch(t *testing.T, workDir string) {
	t.Helper()

	mainRef := strings.TrimSpace(runGit(t, workDir, "ls-remote", "--heads", "origin", "main"))
	if mainRef == "" {
		runGit(t, workDir, "checkout", "--orphan", "main")
		return
	}

	runGit(t, workDir, "fetch", "origin", "main")
	runGit(t, workDir, "checkout", "-B", "main", "origin/main")
}