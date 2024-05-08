/*
 * pages.go
 * lucas@pamorana.net (2024)
 *
 * Web page handlers.
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
|*                               PAGE  HANDLERS                               *|
\* -------------------------------------------------------------------------- */

package control

import (
	/* standard library */
	"os"
	"fmt"
	"slices"
	"strings"
	"os/exec"
	"net/http"
	"encoding/json"

	/* remote modules */
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)


/* -------------------------------------------------------------------------- *\
|*                              DYNAMIC  SCRIPTS                              *|
\* -------------------------------------------------------------------------- */

func ScriptHeaterJsGet(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"scripts/heater.js",
		gin.H{
			"ws_proto": websocket_protocol,
		},
	)
}

func ScriptProgramJsGet(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"scripts/program.js",
		gin.H{
			"ws_proto": websocket_protocol,
		},
	)
}

/* -------------------------------------------------------------------------- *\
|*                               ROUTE HANDLERS                               *|
\* -------------------------------------------------------------------------- */

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// allow all connections
		return true
	},
}

func PageControlGet(c *gin.Context) {
	if (!SessionIsLoggedIn(c)) {
		c.Redirect(http.StatusFound, "/login")
	} else {
		username := c.GetString(session_user)
		c.HTML(
			http.StatusOK,
			"views/control.html",
			gin.H{
				"pagename": "control",
				"username": username,
			},
		)
	}
}

func PageGraphsGet(c *gin.Context) {
	if (!SessionIsLoggedIn(c)) {
		c.Redirect(http.StatusFound, "/login")
	} else {
		username := c.GetString(session_user)
		c.HTML(
			http.StatusOK,
			"views/graphs.html",
			gin.H{
				"pagename": "graphs",
				"username": username,
			},
		)
	}
}

func PageProgramGet(c *gin.Context) {
	if (!SessionIsLoggedIn(c)) {
		c.Redirect(http.StatusFound, "/login")
	} else {
		username := c.GetString(session_user)
		c.HTML(
			http.StatusOK,
			"views/program.html",
			gin.H{
				"pagename": "program",
				"username": username,
			},
		)
	}
}

func PageExportGet(c *gin.Context) {
	if (!SessionIsLoggedIn(c)) {
		c.Redirect(http.StatusFound, "/login")
	} else {
		username := c.GetString(session_user)
		c.HTML(
			http.StatusOK,
			"views/export.html",
			gin.H{
				"pagename": "export",
				"username": username,
			},
		)
	}
}

func PageLoginGet(c *gin.Context) {
	if (SessionIsLoggedIn(c)) {
		c.Redirect(http.StatusFound, "/")
	} else {
		c.HTML(
			http.StatusOK,
			"views/login.html",
			gin.H{
				"pagename":     "login",
				"info_animate": "hidden",  /* hidden, popup */
				"info_type":    "success", /* error, success */
				"info_message": "success",
			},
		)
	}
}

func PageLoginPost(c *gin.Context) {
	if (SessionIsLoggedIn(c)) {
		c.Redirect(http.StatusFound, "/")
		return
	}

	username := c.PostForm("username")
	password := c.PostForm("password")
	remember := c.PostForm("remember")

	success := SessionPerformLogin(c, username, password, remember != "")

	if (success) {
		c.Redirect(http.StatusFound, "/")
	} else {
		c.HTML(
			http.StatusUnauthorized,
			"views/login.html",
			gin.H{
				"pagename":     "login",
				"info_animate": "popup", /* hidden, popup */
				"info_type":    "error", /* error, success */
				"info_message": "Login failure",
			},
		)
	}
}

func PageLogoutGet(c *gin.Context) {
	if (SessionIsLoggedIn(c)) {
		token := c.GetString(session_token)
		delete(sessions, token)
		SessionDeleteCookie(c)
	}
	c.Redirect(http.StatusFound, "/")
}

func PageRegisterGet(c *gin.Context) {
	if (SessionIsLoggedIn(c)) {
		c.Redirect(http.StatusFound, "/")
	} else {
		c.HTML(
			http.StatusOK,
			"views/register.html",
			gin.H{
				"pagename":     "register",
				"info_animate": "hidden",  /* hidden, popup */
				"info_type":    "success", /* error, success */
				"info_message": "success",
			},
		)
	}
}

func PageRegisterPost(c *gin.Context) {
	if (SessionIsLoggedIn(c)) {
		c.Redirect(http.StatusFound, "/")
		return
	}

	username := c.PostForm("username")
	password := c.PostForm("password")

	if hash, err := HashPassword(password); err == nil {
		c.HTML(
			http.StatusOK,
			"views/register.html",
			gin.H{
				"pagename":     "register",
				"info_animate": "static", /* hidden, popup */
				"info_type":    "success", /* error, success */
				"hash_label":   username + ":",
				"hash_value":   hash,
				"hash_focus":   "autofocus",
				"hash_class":   "",
			},
		)
	} else {
		c.HTML(
			http.StatusInternalServerError,
			"views/register.html",
			gin.H{
				"pagename":     "register",
				"info_animate": "popup", /* hidden, popup */
				"info_type":    "error", /* error, success */
				"hash_label":   "Internal server error (500)",
				"hash_value":   "",
				"hash_focus":   "",
				"hash_class":   "hidden",
			},
		)
	}
}

func PageControlPost(c *gin.Context) {
	if (!SessionIsLoggedIn(c)) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	action := c.PostForm("action")

	switch action {
	case "toggle":
		MQTTHeaterPlugCommand("toggle")
		break
	default:
		break
	}

	c.Status(http.StatusNoContent)
}

func PageHeaterDataSocket(c *gin.Context) {
	if s, _, found := SessionGet(c); !found || s.isExpired() {
		c.Status(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	HeaterClients[conn] = true

	if err = conn.WriteJSON(heaterState); err != nil {
		conn.Close()
		delete(HeaterClients, conn)
	} else {
		go heaterWebSocketConnection(conn)
	}
}

func PageProgramPost(c *gin.Context) {
	if (!SessionIsLoggedIn(c)) {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	action := c.PostForm("action")

	switch action {
	case "toggle":
		systemToggleProgramState()
		break
	case "on":
		systemSetProgramState(true)
		break
	case "off":
		systemSetProgramState(false)
		break
	default:
		break
	}

	c.Status(http.StatusNoContent)
}

func PageControlSystemDataSocket(c *gin.Context) {
	if s, _, found := SessionGet(c); !found || s.isExpired() {
		c.Status(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	SystemClients[conn] = true

	controlState.Mutex.Lock()
	err = conn.WriteJSON(controlState)
	controlState.Mutex.Unlock()

	if err != nil {
		conn.Close()
		delete(SystemClients, conn)
	} else {
		go systemWebSocketConnection(conn)
	}
}


/* -------------------------------------------------------------------------- *\
|*                               GITHUB WEBHOOK                               *|
\* -------------------------------------------------------------------------- */

type Repository struct {
	Archived       bool   `json:"archived"`
	Disabled       bool   `json:"disabled"`
	DefaultBranch  string `json:"default_branch"`
	Name           string `json:"name"`
	Organization   string `json:"organization"`
}
type PushEvent struct {
	After      string     `json:"after"`
	Before     string     `json:"before"`
	Created    bool       `json:"created"`
	Deleted    bool       `json:"deleted"`
	Private    bool       `json:"private"`
	Forced     bool       `json:"forced"`
	Ref        string     `json:"ref"`
	Repository Repository `json:"repository"`
}

func PageCommitPost(c *gin.Context) {
	// Validate that the request comes from Github
	secret := os.Getenv("WEBHOOK_SECRET")

	body, err := c.GetRawData()
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	signature, found := c.Request.Header["X-Hub-Signature-256"]
	if !found || len(signature) != 1 {
		c.Status(http.StatusBadRequest)
		return
	}

	parts := strings.SplitN(signature[0], "=", 2)
	if len(parts) != 2 {
		c.Status(http.StatusBadRequest)
		return
	}

	sig_type := parts[0]
	sig_hash := parts[1]

	if sig_type != "sha256" || !ValidateGithubRequest(secret, sig_hash, body) {
		fmt.Printf("%s\n", "could not validate request")
		c.Status(http.StatusUnauthorized)
		return
	}

	if events, found := c.Request.Header["X-GitHub-Event"]; found {
		if !slices.Contains(events, "push") {
			fmt.Printf("%s\n", "not a push event")
			c.Status(http.StatusOK)
			return
		}
	}

	var event PushEvent

	// using unmarshal because instead of c.ShouldBindJSON,
	// since the request body has already been read.
	err = json.Unmarshal(body, &event)
	if err != nil {
		fmt.Printf("%s\n", "could not bind")
		c.Status(http.StatusBadRequest)
		return
	}

	branch, found := Last(strings.Split(event.Ref, "/"))

	if !found || branch != event.Repository.DefaultBranch {
		// another branch
		fmt.Printf("different branch")
		c.Status(http.StatusOK)
		return
	}

	if !event.Repository.Archived &&
	   !event.Repository.Disabled &&
	    event.Repository.Organization == "KandidatarbeteElkraft" &&
	    event.Repository.Name         == "KontrollSystem" {
		// event.forced if force-pushed
		go pullNewContent()
	}
	c.Status(http.StatusOK)
}

func pullNewContent() {
	cmd := exec.Command("git", "reset", "--hard")
	cmd.Dir = control_repo_path
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "KontrollSystem: Could not reset: %v\n", err)
		return
	}

	cmd = exec.Command("git", "pull", "-p", "--no-tags")
	cmd.Dir = control_repo_path
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "KontrollSystem: Could not pull: %v\n", err)
	}
}
