package run

import (
	"os/exec"
	"strings"
)

// Runner executes external commands. Prod uses Exec; tests use Fake.
// RunInput is like Run but pipes `stdin` to the command's standard input —
// used to hand secrets (PATs) to `git credential approve` without putting
// them on the argv (where they'd leak into process listings / Fake.Calls).
type Runner interface {
	Run(name string, args ...string) (string, error)
	RunInput(stdin string, name string, args ...string) (string, error)
}

type Exec struct{}

func (Exec) Run(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (Exec) RunInput(stdin, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

type Fake struct {
	Outputs map[string]string // keyed by full "name arg arg" command line
	Errs    map[string]error
	Calls   []string
	Inputs  []string // stdin payloads passed to RunInput, in call order
}

func (f *Fake) Run(name string, args ...string) (string, error) {
	key := strings.TrimSpace(name + " " + strings.Join(args, " "))
	f.Calls = append(f.Calls, key)
	if f.Errs != nil {
		if e, ok := f.Errs[key]; ok {
			return "", e
		}
	}
	if f.Outputs != nil {
		if o, ok := f.Outputs[key]; ok {
			return o, nil
		}
	}
	return "", nil
}

func (f *Fake) RunInput(stdin, name string, args ...string) (string, error) {
	f.Inputs = append(f.Inputs, stdin)
	return f.Run(name, args...)
}
