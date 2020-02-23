package git

import (
	"testing"
	"time"
)



var reflogRepo *Repository


func setupTestReflog(t *testing.T) (fixture *Reflog) {
	reflogRepo = createTestRepo(t)

	_, _ = seedTestRepo(t, reflogRepo)

	return fixture
}

func cleanupTestReflog(t *testing.T) {
	cleanupTestRepo(t, reflogRepo)
}

func TestReflog_Append(t *testing.T) {
	var (
		reflog *Reflog
		sig *Signature
		ref *Reference
		loc *time.Location
		err error
	)

	setupTestReflog(t)
	defer cleanupTestReflog(t)

	loc, err = time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)

	sig = &Signature{
		Name:  "C O Mitter",
		Email: "committer@example.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	err = reflogRepo.References.EnsureLog("refs/heads/master")
	checkFatal(t, err)

	ref, err = reflogRepo.References.Lookup("refs/heads/master")
	checkFatal(t, err)
	defer ref.Free()

	reflog, err = ref.Log()
	checkFatal(t, err)
	defer reflog.Free()

	count := reflog.EntryCount()
	if count == 0 {
		t.Fatal("expected EntryCount to be > 0")
	}

	commitOid, _ := updateReadme(t, reflogRepo, "content is king\n")
	err = reflog.Append(commitOid, sig, "this is a reflog message\n")
	checkFatal(t, err)

	newCount := reflog.EntryCount()
	if newCount <= count {
		t.Fatalf("expected number of reflog entries to increase, was %d now %d", count, newCount)
	}
}
