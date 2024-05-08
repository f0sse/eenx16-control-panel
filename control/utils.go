/*
 * utils.go
 * lucas@pamorana.net (2024)
 *
 * Helper functions.
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

package control

import (
	/* standard library */
	"fmt"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/rand"
)

/* -------------------------------------------------------------------------- *\
|*                             UTILITY FUNCTIONS                              *|
\* -------------------------------------------------------------------------- */

/* generate "n" random bytes */
func RandomBytes(n int) ([]byte, error) {
	 b := make([]byte, n)
	_, err := rand.Read(b)

	// no error only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

/*
 * ValidateGithubRequest:
 * Checks if the webhook payload came from Github.
 */
func ValidateGithubRequest(secret, headerHash string, payload []byte) bool {
	hash := HashRequestBody(secret, payload)
	return hmac.Equal(
		[]byte(hash),
		[]byte(headerHash),
	)
}

/*
 * HashRequestBody
 * Calculate the authentication code using our shared secret and the request
 * body. Only Github knows the secret.
 * https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries
 */
func HashRequestBody(secret string, playloadBody []byte) string {
	hm := hmac.New(sha256.New, []byte(secret))
	hm.Write(playloadBody)
	sum := hm.Sum(nil)
	return fmt.Sprintf("%x", sum)
}

/*
 * Last:
 * Return the last element in a slice.
 */
func Last[E any](s []E) (E, bool) {
	if len(s) == 0 {
		var zero E
		return zero, false
	}
	return s[len(s)-1], true
}
