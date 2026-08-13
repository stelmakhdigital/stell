package footer

import "os/exec"

// GitAvailable reports whether git is on PATH.
func GitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}
