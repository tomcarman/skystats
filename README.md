<div align="center">
    <img src="docs/logo/logo.jpg" width="500px" align="center" alt="sf metadata linter logo" />
</div>
</br>
<div align="center">
    SkyStats is an application to retrieve, store, and display interesting aircraft ADS-B data received via an SDR.
</div>

## Overview

* Written in Go, using a Postgres database, and a basic html/js website
* ADS-B data is received via [adsb-ultrafeeder](https://github.com/sdr-enthusiasts/docker-adsb-ultrafeeder) / [readsb](https://github.com/wiedehopf/readsb), running on a Raspberry Pi 4 attached to an SDR + aerial ([see it here!](docs/setup/aerial.jpg))
* The application reads raw aircraft data from the readsb [aircraft.json](https://github.com/wiedehopf/readsb-githist/blob/dev/README-json.md) file
* Flight data is stored in a postgres database
* Registration & Routing data is retrieved from the [adsbdb](https://github.com/mrjackwills/adsbdb) API
* "Interesting" aircraft are identified via a local copy of the [plane-alert-db](https://github.com/sdr-enthusiasts/plane-alert-db)
* Various other statistics are (fastest, slowest, highest, lowest) are calculated and stored. Note: This data needs cleansing, as it turns out even the ADS-B world is subject to bad data - with planes reguarly reporting impossible altitudes or speeds.
* A [gin](https://gin-gonic.com/) API surfaces information from the postgres database to the web frontend
* Front end built with vanilla html/js - [Claude Code](https://www.anthropic.com/claude-code) was used liberally, as I am not much of a front end developer!

There are environment variables (`LATITUDE`, `LONGITUDE`, `RADIUS`) that can be used to only process aircraft data that falls within a particular boundary - similar to [planefence](https://github.com/sdr-enthusiasts/docker-planefence). Alternatively, setting the `RADIUS` to something larger than that of your SDR will mean all aircraft data is processed.

## Setup

### Running locally (eg. to develop)
* Clone this repository
* Create the postgres db (e.g. in a Docker container) - `/scripts/schema.sql` can be used to initialise the database
* Rename `.env-example` to `.env` and populate
* Change to the `core` folder e.g. `cd core`
* Compile with `go build -o skystats-daemon`
* Run the app `./skystats-daemon`
* It can be terminated via `kill $(cat skystats/core/skystats.pid)`

### Running in Docker
* Clone this repository
* Populate `docker-compose.yml` with all required values
* Run `docker compose up -d --build`


### Environment Variables

Either set in `.env` for local development, or `docker-compose.yml` when running in docker.

| DB_NAME                   | Name of the postgres database - `skystats_db`                                                                                                             |
|---------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| DB_USER                   | Postgres username - `admin`                                                                                                                               |
| DB_PASSWORD               | Postgres password - `password`                                                                                                                            |
| DB_HOST                   | Postgres host - `192.168.1.xxx`                                                                                                                           |
| DB_PORT                   | Postgres port - `5432`                                                                                                                                    |
| ADSB_HOST                 | IP of the ADSB receiver running readsb - `192.168.1.xxx`                                                                                                  |
| ADSB_PORT                 | Port on ADSB receiver where readsb aircraft.json being served - `8080`                                                                                    |
| LATITUDE                  | Lattitude of your receiver - `xx.xxxxxx`                                                                                                                  |
| LONGITUDE                 | Longitude of your receiver - `yy.yyyyyy`                                                                                                                  |
| RADIUS                    | Distance in km from your receiver that you want to record aircraft. Set to a distance greater than that of your receiver to capture all aircraft - `1000` |
| ADSB_DB_AIRCRAFT_ENDPOINT | ADSB DB aircraft endpoint, used for registration data - `https://api.adsbdb.com/v0/aircraft/`                                                             |
| ADSB_DB_CALLSIGN_ENDPOINT | ADSB DB callsign endpoint, used for route data - `https://api.adsbdb.com/v0/callsign/`                                                                    |

## Screenshots

![General](docs/screenshots/General2.png)
</br>
![MilGov](docs/screenshots/MilGov.png)
</br>
![PolCiv](docs/screenshots/PolCiv.png)
</br>
![Overlay](docs/screenshots/Overlay.png)
</br>
![Stats](docs/screenshots/Stats.png)
