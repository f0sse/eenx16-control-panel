/*
 * main.go
 * lucas@pamorana.net (2024)
 *
 * Main, Control Panel.
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

package main

import (
	/* standard library */
	"html/template"

	/* remote modules */
	"github.com/gin-gonic/gin"

	/* local modules */
	"control"
)

func activePageNavigation(item, page string) string {
	if (item == page) {
		return "active"
	} else {
		return ""
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.SetTrustedProxies([]string{"127.0.0.0/8", "::1/128"})

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Static("/css",    "./static/css")
	r.Static("/img",    "./static/img")
	r.Static("/js",     "./static/js")
	r.Static("/vendor", "./static/vendor")

	//r.StaticFile("/favicon.ico", "./img/favicon.ico")

	r.SetFuncMap(template.FuncMap{
		"activePageNavigation": activePageNavigation,
	})

	r.LoadHTMLGlob("./templates/**/*")

	/* routes */
	r.GET ("/",              control.PageControlGet)
	r.GET ("/control",       control.PageControlGet)
	r.POST("/",              control.PageControlPost)
	r.POST("/control",       control.PageControlPost)
	r.GET ("/graphs",        control.PageGraphsGet)
	r.GET ("/program",       control.PageProgramGet)
	r.POST("/program",       control.PageProgramPost)
	r.GET ("/export",        control.PageExportGet)
	r.GET ("/login",         control.PageLoginGet)
	r.POST("/login",         control.PageLoginPost)
	r.GET ("/register",      control.PageRegisterGet)
	r.POST("/register",      control.PageRegisterPost)
	r.POST("/commit",        control.PageCommitPost)
	r.GET ("/logout",        control.PageLogoutGet)

	r.GET ("/ds/heater.js",  control.ScriptHeaterJsGet)
	r.GET ("/ds/program.js", control.ScriptProgramJsGet)

	/* websocket */
	r.GET("/socket", control.PageHeaterDataSocket)
	r.GET("/system", control.PageControlSystemDataSocket)

	err := control.MQTTConnect()
	if (err != nil) {
		panic("Could not connect to MQTT broker")
	}
	defer control.MQTTDisconnect(250)

	control.InfluxConnect()
	defer control.InfluxDisconnect()

	go control.SystemUpdateProgramState()

	r.Run("127.0.0.1:8123")
}
