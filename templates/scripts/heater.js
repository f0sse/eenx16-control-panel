/*
 * heater.js
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
{{ define "scripts/heater.js" }}
(function ($) {
	'use strict'

	var socket = null;
	var connected = false;

	function reconnect() {
		if (!connected) {
			socket = new WebSocket("{{.ws_proto}}://" + location.host + "/socket");
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
				$('span#heater-power').html(data.power.toFixed(2));
				$('span#heater-energy').html(data.energy.toFixed(2));
				$('span#heater-temperature').html(data.temperature.toFixed(2));

				if (data.state) {
					$('button#heater-toggle')
						.addClass('btn-success')
						.removeClass('btn-info')
						.removeClass('btn-danger');
					$('span#heater-state').html("on");
				} else {
					$('button#heater-toggle')
						.addClass('btn-danger')
						.removeClass('btn-info')
						.removeClass('btn-success');
					$('span#heater-state').html("off");
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
		e.preventDefault();
		$.ajax({
			type: "POST",
			url: "/control",
			data: "action=toggle",
		});
	});
})(jQuery)
{{end}}

