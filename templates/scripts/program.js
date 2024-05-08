/*
 * program.js
 * lucas@pamorana.net (2024)
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
{{ define "scripts/program.js" }}
(function ($) {
	'use strict'

	var socket = null;
	var connected = false;

	function reconnect() {
		if (!connected) {
			socket = new WebSocket("{{.ws_proto}}://" + location.host + "/system");
			socket.onopen = function(e) {
				connected = true;
			};
			socket.onclose = function (e) {
				connected = false;
				socket = null;
			};
			socket.onerror = function (e) {
				socket.close();
			};
			socket.onmessage = function (e) {
				let data = JSON.parse(e.data);
				/*
				 * data = {
				 *   time:   ...,
				 *   state:  ...,
				 *   power:  ...,
				 *   energy: ...,
				 * }
				 */
				$('textarea#program-output').val(data.output)
				if (data.running) {
					$('button#program-toggle')
						.addClass('btn-success')
						.removeClass('btn-info')
						.removeClass('btn-danger');
					$('button#program-toggle').attr('value', 'off');
					$('span#program-running').html("running");
				} else {
					$('button#program-toggle')
						.addClass('btn-danger')
						.removeClass('btn-info')
						.removeClass('btn-success');
					$('button#program-toggle').attr('value', 'on');
					$('span#program-running').html("not running");
				}
			};
			$(window).on('beforeunload', function() {
				socket.onclose = function() {};
				socket.close();
			});
		}
	}

	reconnect()
	setInterval(reconnect, 5000);

	$("form#toggleButton").on("submit", function(e) {
		var val = $('button#program-toggle').attr('value');
		if (!val)
			val = "toggle";
		e.preventDefault();
		$.ajax({
			type: "POST",
			url: "/program",
			data: "action=" + val,
		});
	});
})(jQuery)
{{end}}

