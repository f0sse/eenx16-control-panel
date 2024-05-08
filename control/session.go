/*
 * session.go
 * lucas@pamorana.net (2024)
 *
 * User session handling.
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
|*                             SESSION  HANDLING                              *|
\* -------------------------------------------------------------------------- */

package control

import (
	/* standard library */
	"time"
	"encoding/base64"

	/* remote modules */
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/language"
	"golang.org/x/text/cases"
	"github.com/gin-gonic/gin"
)


/* -------------------------------------------------------------------------- *\
|*                               SESSION TYPES                                *|
\* -------------------------------------------------------------------------- */

type user struct {
	username string `json:"username"`
	hash     string `json:"hash"`
}

var users = []user{
	{username: "example",  hash: "$2a$14$9v1ewG7AAtyymLwf3v0MBOhtrCjWETSL7HzdlVN6jbdYmtUKWRxOi" },
}

type session struct {
	username string
	expiry   time.Time
}


/* -------------------------------------------------------------------------- *\
|*                          SESSION  IMPLEMENTATION                           *|
\* -------------------------------------------------------------------------- */

/*
 * simple memory storage for active user sessions.
 * this will not persist a device/program reboot/restart.
 *
 * for a more complex program, use
 */
var sessions = map[string] session{}


/* -------------------------------------------------------------------------- *\
|*                            PASSWORD  FUNCTIONS                             *|
\* -------------------------------------------------------------------------- */

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), password_cost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

/* -------------------------------------------------------------------------- *\
|*                             SESSION FUNCTIONS                              *|
\* -------------------------------------------------------------------------- */

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

func SessionDeleteCookie(c *gin.Context) {
	c.SetCookie(cookie_name, "", -1, "/", cookie_domain, http_secure, true)
}

func SessionCreateToken() string {
	bytes, _ := RandomBytes(24)
	return base64.StdEncoding.EncodeToString(bytes)
}

func SessionGet(c *gin.Context) (*session, string, bool) {
	if token, err := c.Cookie(cookie_name); err == nil {
		s, found := sessions[token]
		return &s, token, found
	}
	return nil, "", false
}

func SessionIsLoggedIn(c *gin.Context) bool {
	c.Set(session_user,  nil)
	c.Set(session_token, nil)
	s, token, found := SessionGet(c)
	if (found) {
		if s.isExpired() {
			delete(sessions, token)
			SessionDeleteCookie(c)
			return false
		} else {
			c.Set(session_user, s.username)
			c.Set(session_token, token)
			return true
		}
	} else {
		SessionDeleteCookie(c)
	}
	return false
}

func SessionPerformLogin(c *gin.Context, usr string, pass string, remember bool) (bool) {
	found := false

	for _, u := range users {
		found = CheckPasswordHash(pass, u.hash)

		if (found) {
			break
		}
	}

	if (found) {
		multiplier := time.Second * 5

		if (remember) {
			multiplier = 65536 * time.Second
		}

		timeout := multiplier * time.Minute
		expiry := time.Now().Add(timeout)
		token := ""

		for {
			token = SessionCreateToken()

			if _, exists := sessions[token]; exists {
				continue
			}

			break
		}

		sessions[token] = session{
			username: cases.Title(language.English, cases.Compact).String(usr),
			expiry:   expiry,
		}

		c.SetCookie(cookie_name, token, int(timeout), "/", cookie_domain, http_secure, true)

		return true
	}

	SessionDeleteCookie(c)

	return false
}
