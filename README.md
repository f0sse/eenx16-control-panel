# Control Panel

A tiny control panel written in Go, to control the heater in the Björkö
laboratory. The intended feature set was to have controls for the heater's
state, graphs showing the time series data gathered by collection agents from
another repository in this organisation, a way to export said data, and a panel
to easily make updates to the project's control system logic.

Since there was not enough time to develop the full feature set, only the
controls were implemented, and Graphana as a substitute for making our own
graphs.

**Preview**:

![Preview](preview.png?raw=true "Preview")

## MQTT

The Raspberry Pi at the lab runs an MQTT broker, which the Shelly Plug connects
to.

## Shelly Plug

Wireless wall-plug that can be remotely controlled.

## InfluxDB

Time-series database, storing gathered metrics over time.
