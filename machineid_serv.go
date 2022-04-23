//go:build freebsd || netbsd || openbsd || dragonfly || solaris || darwin || linux || windows

package teomon

import "github.com/denisbrodbeck/machineid"

func getMachineID() (id string, err error) {
	return machineid.ID()
}
