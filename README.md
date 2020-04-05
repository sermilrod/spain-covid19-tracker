# spain-covid19-tracker
This is the data repository for the 2019 Novel Coronavirus in Spain.

## Data Sources
* [Instituto de Salud Carlos III](https://covid19.isciii.es/)

## Usage

Start the InfluxDB and Grafana by running the docker-compose:
```bash
$ docker-compose up
```

(Only first time) Create the database:
```shell
$ docker exec -ti influxdb bash
root@89af36489b74:/# influx
Connected to http://localhost:8086 version 1.7.10
InfluxDB shell version: 1.7.10
> create database spain_covid19
```

Seed the database running the collector:
```bash
$ go run main.go
```

## Collected metrics

| Metric Name | Description |
|-------------|-------------|
| spain_covid19_cases_total | Accumulated value of all types of cases nation wide |
| spain_covid19_cases_region | Accumulated value of all types of cases in region |
| spain_covid19_cases_per_100000_region | Accumulated value per 100000 inhabitatns of total cases in region |
| spain_covid19_infection_rate_region | Rate of new infections in region |

### Available regions

| Region | DB Normalised Name |
|--------|--------------------|
| Andalucía | andalucia |
| Aragón | aragon |
| Principado de Asturias | asturias |
| Canarias | canarias |
| Cantabria | cantabria |
| Castilla-La Mancha | castilla_la_mancha |
| Castilla y León | castilla_leon |
| Cataluña | catalunya |
| Ciudad Autónoma de Ceuta | ceuta |
| Comunitat Valenciana | comunidad_valenciana |
| Extremadura | extremadura |
| Galicia | galicia |
| Illes Balears | islas_baleares |
| La Rioja | la_rioja |
| Comunidad de Madrid | madrid |
| Ciudad Autónoma de Melilla | melilla |
| Región de Murcia | murcia |
| Comunidad Foral de Navarra | navarra |
| País Vasco | pais_vasco |

### spain_covid19_cases_total fields

| Field | Description |
|-------|-------------|
| active | Number of active cases: Total - (hospital + deaths) |
| cases | Total number of registered cases |
| critical | Number of cases in ICU (UCI)|
| deaths | Number of defunctions  |
| hospitalised | Number of cases in hospital |
| recovered | Number of cases that got over the illness |

### spain_covid19_cases_region fields

| Field | Description |
|-------|-------------|
| active | Number of active cases: Total - (hospital + deaths) in region |
| cases | Total number of registered cases in region |
| critical | Number of cases in ICU (UCI) in region |
| deaths | Number of defunctions in region |
| hospitalised | Number of cases in hospital in region |
| recovered | Number of cases that got over the illness in region |

### spain_covid19_cases_per_100000_region

| Field | Description |
|-------|-------------|
| cases | Total number of registered cases per 100000 inhabitatns in region |

### spain_covid19_infection_rate_region 

| Field | Description |
|-------|-------------|
| rate | Rate of new infections in region |
