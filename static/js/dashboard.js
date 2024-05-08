/*
 * dashboard.js
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

(function ($) {
	'use strict'

	// Graphs
	var ctx = document.getElementById('myChart')

	// eslint-disable-next-line no-unused-vars
	var myChart = new Chart(ctx, {
		type: 'line',
		data: {
			labels: [
				'Sunday',
				'Monday',
				'Tuesday',
				'Wednesday',
				'Thursday',
				'Friday',
				'Saturday'
			],
			datasets: [{
				data: [
					15339,
					21345,
					18483,
					24003,
					23489,
					24092,
					12034
				],
				lineTension: 0,
				backgroundColor: 'transparent',
				borderColor: '#007bff',
				borderWidth: 4,
				pointBackgroundColor: '#007bff'
			}]
		},
		options: {
			scales: {
				yAxes: [{
					ticks: {
						beginAtZero: false
					}
				}]
			},
			legend: {
				display: false
			}
		}
	})
})(jQuery)
