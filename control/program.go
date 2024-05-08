/*
 * program.go
 * lucas@pamorana.net (2024)
 *
 * Execution of a Python control program.
 *
 * This program is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more
 * details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */

/* -------------------------------------------------------------------------- *\
|*                               CONTROL SYSTEM                               *|
\* -------------------------------------------------------------------------- */

package control

import (
	/* standard library */
	"os"
	"io"
	"fmt"
	"time"
	"sync"
	"bytes"
	"syscall"
	"os/exec"

	/* remote modules */
	"github.com/gorilla/websocket"
)


/* -------------------------------------------------------------------------- *\
|*                               PROGRAM  STATE                               *|
\* -------------------------------------------------------------------------- */

var controlState ControlSystemState = ControlSystemState{
	Time: time.Now(),
	Running: false,
	Output: "",
}

type ControlSystemState struct {
	Mutex       sync.Mutex   `json:"-"`
	Buffer      bytes.Buffer `json:"-"`
	Command     exec.Cmd     `json:"-"`
	Time        time.Time    `json:"time"`    // when it was started
	Running     bool         `json:"running"` // "running" or not
	Output      string       `json:"output"`  // program's output buffer
}

func systemSetProgramState(state bool) {
	controlState.Mutex.Lock()
	if state {
		if !controlState.Running {
			systemStartControlProgram()
		}
	} else {
		if controlState.Running {
			systemStopControlProgram()
		}
	}
	controlState.Mutex.Unlock()
}

func systemToggleProgramState() {
	controlState.Mutex.Lock()
	if controlState.Running {
		systemStopControlProgram()
	} else {
		systemStartControlProgram()
	}
	controlState.Mutex.Unlock()
}

func SystemUpdateProgramState() {
	for {
		time.Sleep(1 * time.Second)
		controlState.Mutex.Lock()
		if controlState.Running {
			controlState.Output = string(controlState.Buffer.Bytes())
			pushSystemState()
		}
		controlState.Mutex.Unlock()
	}
}

// assume that the lock is held
func systemStopControlProgram() {
	err := controlState.Command.Process.Signal(syscall.SIGTERM)
	if err == nil {
		e := make(chan error)
		go func() {
			// kill the program if it hasn't 
			time.Sleep(time.Second * 5)
			e <- controlState.Command.Process.Kill()
		}()
		for {
			select {
				case <-e:
					// assume dead
					return
				default:
					err := controlState.Command.Process.Signal(syscall.Signal(0))
					if err != nil {
						// assume exited
						return
					}
					time.Sleep(time.Millisecond * 10)
			}
		}
	}
	// the goroutine will take care of updating
	// control state once the process exits
	return
}

// assume that the lock is held
func systemStartControlProgram() error {
	// wipe old buffer content
	controlState.Buffer.Reset()

	// the control system has its own virtual environment.
	// "-u" to use unbuffered I/O
	cmd := exec.Command(control_py_venv + "/bin/python3", "-u",
						control_repo_path + "/control.py")

	// open its standard input for writing
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	// for importing of the other components to work properly
	cmd.Dir = control_repo_path

	// the webhook secret is not needed, so remove it from the env.
	cmd.Env = append(os.Environ(), "WEBHOOK_SECRET=")

	cmd.Stdout = &controlState.Buffer
	cmd.Stderr = &controlState.Buffer

	controlState.Time = time.Now()

	if err := cmd.Start(); err != nil {
		controlState.Running = false
		controlState.Output = fmt.Sprintf("%v")
		return err
	} else {
		controlState.Command = *cmd
		controlState.Running = true
	}

	waited := make(chan error)

	// write current temperature to stdin in 5 second intervals
	go func(f io.WriteCloser) {
		for {
			select {
				case <-waited:
					// process stopped
					return
				default:
					_, err := fmt.Fprintf(f, "%.2f\n", heaterState.Temperature)
					if err != nil {
						// assume it's closed
						return
					}
					time.Sleep(time.Second * 1)
			}
		}
	}(stdin)

	// wait for the command to finish in a goroutine
	// concurrently. cannot assume that lock is held
	go func() {
		waited <- cmd.Wait()
		// process has exited
		controlState.Mutex.Lock()
		controlState.Running = false
		controlState.Output = string(controlState.Buffer.Bytes())
		pushSystemState()
		controlState.Mutex.Unlock()
	}()

	pushSystemState()

	return nil
}


/* -------------------------------------------------------------------------- *\
|*                              WEBSOCKET STREAM                              *|
\* -------------------------------------------------------------------------- */

/* websocket connections */
var SystemClients = make(map[*websocket.Conn]bool)

func systemWebSocketConnection(conn *websocket.Conn) {
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			delete(SystemClients, conn)
			break
		}
	}
}

// don't call outside synchronized scope
func pushSystemState() {
	for client := range SystemClients {
		err := client.WriteJSON(controlState)
		if err != nil {
			client.Close()
			delete(SystemClients, client)
		}
	}
}
