/*
 * control.go
 * lucas@pamorana.net (2024)
 *
 * Configuration of constants.
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

/* -------------------------------------------------------------------------- *\
|*                             CONSTANTS / CONFIG                             *|
\* -------------------------------------------------------------------------- */

// set to true when deployed behind a TLS proxy
const http_secure = true

// HTTP cookie domain for which it is valid
const cookie_domain = "8e.nu"

// key for cookie
const cookie_name = "token"

// Gin context storage key for the username
const session_user = "session_username"

// Gin context storage key for the session token
const session_token = "session_token"

// the bcrypt hash cost for passwords (range 4-31)
const password_cost = 14

// MQTT broker URL
const mqtt_url = "mqtts://<REDACTED>:8883"

// InfluxDB server URL
const influx_url = "http://127.0.0.1:8086"

// wss or ws depending on TLS proxy in front or not
const websocket_protocol = "wss"

// where in the file system the control system repository exists
const control_repo_path = "<REDACTED>/kontrollsystem"

// where in the filesystem the python virtual environment is
const control_py_venv = "<REDACTED>/pyenv"
