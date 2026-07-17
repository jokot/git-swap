package run

import "testing"

func TestFakeRunnerRecordsAndReturns(t *testing.T) {
	f := &Fake{Outputs: map[string]string{"git config --global user.email": "a@b.com"}}
	out, err := f.Run("git", "config", "--global", "user.email")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out != "a@b.com" {
		t.Fatalf("got %q want a@b.com", out)
	}
	if len(f.Calls) != 1 || f.Calls[0] != "git config --global user.email" {
		t.Fatalf("calls not recorded: %v", f.Calls)
	}
}
