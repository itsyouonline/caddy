package caddy

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
)

func init() {
	// register CaddyfileInput with gob so it knows into
	// which concrete type to decode an Input interface
	gob.Register(CaddyfileInput{})
}

// signalSuccessToParent tells the parent our status using pipe at index 3.
// If this process is not a restart, this function does nothing.
// Calling this function once this process has successfully initialized
// is vital so that the parent process can unblock and kill itself.
// This function is idempotent; it executes at most once per process.
func signalSuccessToParent() {
	signalParentOnce.Do(func() {
		if IsUpgrade() {
			ppipe := os.NewFile(3, "")               // parent is reading from pipe at index 3
			_, err := ppipe.Write([]byte("success")) // we must send some bytes to the parent
			if err != nil {
				log.Printf("[ERROR] Communicating successful init to parent: %v", err)
			}
			ppipe.Close()
		}
	})
}

// signalParentOnce is used to make sure that the parent is only
// signaled once; doing so more than once breaks whatever socket is
// at fd 4 (TODO: the reason for this is still unclear - to reproduce,
// call Stop() and Start() in succession at least once after a
// restart, then try loading first host of Caddyfile in the browser
// - this was pre-v0.9; this code and godoc is borrowed from the
// implementation then, but I'm not sure if it's been fixed yet, as
// of v0.10.7). Do not use this directly; call signalSuccessToParent
// instead.
var signalParentOnce sync.Once

// transferGob is used if this is a child process as part of
// a graceful upgrade; it is used to map listeners to their
// index in the list of inherited file descriptors. This
// variable is not safe for concurrent access.
var loadedGob transferGob

// Upgrade re-launches the process, preserving the listeners
// for a graceful upgrade. It does NOT load new configuration;
// it only starts the process anew with the current config.
// This makes it possible to perform zero-downtime binary upgrades.
func Upgrade() error {
	log.Println("[INFO] Upgrading")

	// use existing Caddyfile; do not change configuration during upgrade
	currentCaddyfile, _, err := getCurrentCaddyfile()
	if err != nil {
		return err
	}

	if len(os.Args) == 0 { // this should never happen, but...
		os.Args = []string{""}
	}

	// tell the child that it's a restart
	env := os.Environ()
	if !IsUpgrade() {
		env = append(env, "CADDY_UPGRADE=1")
	}

	// prepare our payload to the child process
	cdyfileGob := transferGob{
		ListenerFds: make(map[string]uintptr),
		Caddyfile:   currentCaddyfile,
	}

	// prepare a pipe to the fork's stdin so it can get the Caddyfile
	rpipe, wpipe, err := os.Pipe()
	if err != nil {
		return err
	}

	// prepare a pipe that the child process will use to communicate
	// its success with us by sending > 0 bytes
	sigrpipe, sigwpipe, err := os.Pipe()
	if err != nil {
		return err
	}

	// pass along relevant file descriptors to child process; ordering
	// is very important since we rely on these being in certain positions.
	extraFiles := []*os.File{sigwpipe} // fd 3

	// add file descriptors of all the sockets
	for i, j := 0, 0; ; i++ {
		instancesMu.Lock()
		if i >= len(instances) {
			instancesMu.Unlock()
			break
		}
		inst := instances[i]
		instancesMu.Unlock()

		for _, s := range inst.servers {
			gs, gracefulOk := s.server.(GracefulServer)
			ln, lnOk := s.listener.(Listener)
			pc, pcOk := s.packet.(PacketConn)
			if gracefulOk {
				if lnOk {
					lnFile, _ := ln.File()
					extraFiles = append(extraFiles, lnFile)
					cdyfileGob.ListenerFds["tcp"+gs.Address()] = uintptr(4 + j) // 4 fds come before any of the listeners
					j++
				}
				if pcOk {
					pcFile, _ := pc.File()
					extraFiles = append(extraFiles, pcFile)
					cdyfileGob.ListenerFds["udp"+gs.Address()] = uintptr(4 + j) // 4 fds come before any of the listeners
					j++
				}
			}
		}
	}

	// set up the command
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Stdin = rpipe      // fd 0
	cmd.Stdout = os.Stdout // fd 1
	cmd.Stderr = os.Stderr // fd 2
	cmd.ExtraFiles = extraFiles
	cmd.Env = env

	// spawn the child process
	err = cmd.Start()
	if err != nil {
		return err
	}

	// immediately close our dup'ed fds and the write end of our signal pipe
	for _, f := range extraFiles {
		f.Close()
	}

	// feed Caddyfile to the child
	err = gob.NewEncoder(wpipe).Encode(cdyfileGob)
	if err != nil {
		return err
	}
	wpipe.Close()

	// determine whether child startup succeeded
	answer, readErr := ioutil.ReadAll(sigrpipe)
	if answer == nil || len(answer) == 0 {
		cmdErr := cmd.Wait() // get exit status
		errStr := fmt.Sprintf("child failed to initialize: %v", cmdErr)
		if readErr != nil {
			errStr += fmt.Sprintf(" - additionally, error communicating with child process: %v", readErr)
		}
		return fmt.Errorf(errStr)
	}

	// looks like child is successful; we can exit gracefully.
	log.Println("[INFO] Upgrade finished")
	return Stop()
}

// IsUpgrade returns true if this process is part of an upgrade
// where a parent caddy process spawned this one to upgrade
// the binary.
func IsUpgrade() bool {
	mu.Lock()
	defer mu.Unlock()
	return isUpgrade
}

// transferGob maps bind address to index of the file descriptor
// in the Files array passed to the child process. It also contains
// the Caddyfile contents and any other state needed by the new process.
// Used only during graceful upgrades.
type transferGob struct {
	ListenerFds map[string]uintptr
	Caddyfile   Input
}
