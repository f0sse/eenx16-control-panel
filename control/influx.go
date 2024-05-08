/*
 * influx.go
 * lucas@pamorana.net (2024)
 *
 * Communication with a time series database.
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
|*                               INFLUX METRICS                               *|
\* -------------------------------------------------------------------------- */

package control

import (
	/* standard library */
	"os"
	"fmt"
	"time"
	"context"

	/* remote dependencies */
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

const temperature_query = `from(bucket: "temperature")
|> range(start: -1d)
|> filter(fn: (r) => r["_measurement"] == "temperature")
|> filter(fn: (r) => r["_field"] == "temperature")
|> filter(fn: (r) => r["location"] == "björkö")
|> aggregateWindow(every: 2d, fn: last, createEmpty: false)
|> yield(name: "last")`

var client influxdb2.Client = nil
var wapi   api.WriteAPI     = nil
var qapi   api.QueryAPI     = nil

func b2float (b bool) float64 {
	if (b) {
		return 1.
	}
	return 0.
}

func temperatureHandler() {
	var lastTemp = 0.0
	for {
		time.Sleep(5 * time.Second)
		temp, err := InfluxReadCurrentTemperature()
		if err != nil {
			continue
		}
		if temp != lastTemp {
			lastTemp = temp
			heaterState.Temperature = temp
			PushPlugState(false)
		}
	}
}

func InfluxConnect() {
	client = influxdb2.NewClient(influx_url, os.Getenv("INFLUXDB_TOKEN"))
	wapi = client.WriteAPI("Kandidatarbete", "electricity")
	qapi = client.QueryAPI("Kandidatarbete")
	go temperatureHandler()
}

func InfluxWriteHeaterState(state HeaterStatus) {
	if wapi != nil {
		p := influxdb2.NewPointWithMeasurement("heater").
			AddTag("meter", "shellyplug").
			AddField("on",     b2float(state.State)).
			AddField("power",  state.Power).
			AddField("energy", state.Energy).
			SetTime(state.Time)

		wapi.WritePoint(p)
	}
}

func InfluxReadCurrentTemperature() (float64, error) {
	if qapi != nil {
		var temp = 0.0
		result, err := qapi.Query(context.Background(), temperature_query)
		if err != nil {
			return 0, err
		}
		for result.Next() {
			value, ok := result.Record().Value().(float64)
			if !ok {
				result.Close()
				return 0.0, fmt.Errorf("Could not parse temperature as float")
			}
			temp = value
		}
		if result.Err() != nil {
			return 0.0, result.Err()
		} else {
			result.Close()
		}
		return temp, nil
	}
	return 0.0, fmt.Errorf("Not connected to InfluxDB")
}

func InfluxDisconnect() {
	if client != nil {
		client.Close()
	}
}
