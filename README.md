# Metrics Backend

## Features

### Metrics

Clients send Metrics to this Backend like a server/Docker container still being up ping or disk usage. If a server or container is down, an alert is sent via Telegram bot. If the server/container is up again, another message is sent, that it is ok again. It works the same for disk usage alerts.

There is a dashboard UI to view the current state of the metrics:

<img src="doc/dashboard.png" width="500">

### Journal logs

The backend also receives journal logs from the clients. These are stored in a timescale Postgres database. A simple UI is provided to view these logs:

<img src="doc/journal.png" width="500">

## Requirements

* Go >= 1.22
* Database 
  * A Postgres database
  * A Postgres database with [timescale](https://www.timescale.com/) extension.
  * It can be a single Postgres instance if it has at least timescale installed

## Setup

### Database

Under `docker/metrics-backend/docker-compose.yml` is an example database setup with one normal Postgres and one timescale Postgres database. You can use this as a starting point.

To start them go to `docker/metrics-backend` and run:

```bash
docker compose up -d
```

### Docker image creation

The easiest way to get started is by using Docker.

To do that first generate the binary:

```bash
go build -o docker/production/metrics-backend
```

Copy the views to the Dockerfile folder:

```bash
cp -r views docker/production/
````

Go to `docker/production` 

Add an `.env` file with the following example content:

```bash
DATABASE_URL="postgres://postgres:local@127.0.0.1:5432/metric"
TIMESCALE_DATABASE_URL="postgres://postgres:local@127.0.0.1:5433/journal"
TELEGRAM_TOKEN="token"
TELEGRAM_CHAT_ID="12345"
GOOGLE_CLIENT_ID="client-id"
GOOGLE_CLIENT_SECRET="client-secret"
GOOGLE_CALLBACK_URL="http://localhost:8080/auth/google/callback"
SESSION_SECRET="some_super_duper_secret"
````

Replace the values with your own.

### Docker compose setup

Finally run:

```bash
SERVER_IP=127.0.0.1
docker compose up -d --build
````

Replace `SERVER_IP` with whatever your actual IP is.

If you get

```
Error response from daemon: failed to create task for container: failed to initialize logging driver: journald is not enabled on this host
```

Remove

```
logging:
  driver: journald
```

from the `docker-compose.yml` file.

## Usage

There isn't much to do. The [metrics-sender](https://github.com/gorlug/metrics-sender) Cli is the main way to configure what is sent to the server.

If a metric turns into an alert, a message is sent to the Telegram bot. If the alert is resolved, another message is sent.

### Dashboard

To view the current state of the metrics go to the dashboard at [http://localhost:8080/dashboard](http://localhost:8080/dashboard). Replace localhost with your host if needed.

<img src="doc/dashboard.png" width="500">

Clicking delete the button "Delete" will remove the metric from the database.

### Journal logs

To view the journal logs go to [http://localhost:8080/journal](http://localhost:8080/journal). Replace localhost with your host if needed.

<img src="doc/journal.png" width="500">

The inputs:

* Start: Start date time of the logs, inclusive
* End: End date time of the logs, inclusive
* Page size: Number of logs per page
* Container: Filter by Docker container name
* Host: Filter by host name
* Container message filter: Apply a substring search to the Docker container message
* Apply range: Applies the above filters and shows the logs
* Previous: Shows the previous page of logs
* Next: Shows the next page of logs

## Motivation

Having alerts when you're running your own infrastructure is important. Otherwise, how can you know if it is not running anymore.

In the past I used to run an Elasticsearch cluster on my own. The Servers would send their metrics to it and via Kibana I set up alerts for certain metrics like "container isn't running", "disk is full". I used Elasticsearch also as a small JSON database. But maintaining and updating an Elasticsearch cluster is quite time-consuming. 

Next I used AWS Cloudwatch metrics for this purpose. I built a small Cli in Dart that sent the metrics I was interested in to Cloudwatch. Cloudwatch was that configured to alert me if certain thresholds were reached. Now I didn't have to maintain the Elasticsearch cluster.

I realized that the metrics and alerts are not as cheap as I thought. Still cheap, but I was like "I could save that money by building something simple myself". Of course there is always the pitfall that it takes way too much time.

Finally, I started watching more of ThePrimeagen and he was hyping Go and HTMX so much that I wanted to try it out for myself. So I thought this metrics idea would be a first good project to start with. It is not too simple, because it involves a Database, a Cron Job and a Webserver. But also not overly complicated, the actual business logic is quite simple.

## Metrics

A Webserver runs that takes metrics POST requests from clients. These use the [metrics-sender](https://github.com/gorlug/metrics-sender) Cli tool that I also wrote to send the metrics.

The metrics look like this:

```go
type MetricValues struct {
Host      string     `json:"host"`
Name      string     `json:"name"`
Type      MetricType `json:"type"`
Timestamp time.Time  `json:"timestamp"`
// optional
Value string      `json:"value,omitempty"`
State MetricState `json:"state,omitempty"`
Id    int
}


type MetricType string

const (
Disk MetricType = "disk"
Ping MetricType = "ping"
)


type MetricState string

const (
OK    MetricState = "ok"
Alert MetricState = "alert"
)
```

### Disk metric

There are two types of metrics. The simplest is `disk` which just sends the disk usage in percent. Here an alert is sent if the disk usage is above 90%.

### Ping metric

The second type is `ping`. Here the idea is that the client constantly sends this ping metric every minute. If the backend stops receiving a ping metric then after 5 minutes an alert is sent.

Optionally with the ping metric a value can be sent. The value is the number of minutes until the alert should be triggered. This is useful for backups that run only once a day. Here the alert would always be triggered because it does not happen mor often than every 24 hours. So if you specify 60 minutes * 24h = 1440 minutes, the alert is only triggered after these 24h.

### State

The state just specifies if the metric is still ok or in an alert state. This way alerts are not sent again if it is still in the state "alert".

## Storing the metrics

The different metrics are saved in a Postgres database. I already had one set up so it was an easy option for me to use.

## Sending alerts

I use Telegram a lot and have used their simple bots API in the past. This made it easy to set up a telegram bot to which alerts are sent.

## Dashboard

I built a simple Dashboard that shows all the metrics in an HTML table. This was to try out Golang HTML templates.

Then I realized I have a bunch of testing metrics that were left over from development. This is what I used as a pretext to try out HTMX, By adding a delete button to each metric.

## Journal logs

### Choosing a database

I had been previously using ElasticSearch also for viewing logs from systemd journal. I lost that when I switched to Cloudwatch Metrics and didn't set that up to forward it to Cloudwatch as well.

On my journey developing the metrics, HTMX seemed so easy I felt like maybe I can implement something simple myself.

The first hurdle was where to store it. My first thought was using Postgres as well, but that would be a lot of data. I thought about using some other database, but I didn't want to have to set up something new.

Luckily I discovered the [timescale](https://www.timescale.com/) extension for Postgres which is optimized for time series data and has automatic deletion of time series data. Now that sounds exactly like what I've been looking for. Because it is Postgres it is super easy to set up. I went with using a separate database because I didn't want to clutter my actual production database with a bunch of logs. If I lose the timescale db it isn't a big deal.

### Prisma

This is my first real Golang project. Before that I used TypeScript. So when thinking about how to set up the Postgres database my first go to in the TypeScript world is [Prisma](https://prisma.io). It works standalone to set up the database so that is what I used. It has some really nice db migration functionality that makes changing the database schema really easy, which is nice for development.

### Timescale DB setup

But Prisma lacked the functionality to set up the timescale extension. For that I used my own SQL script.

### Implementation

On the backend I added another schema for the time series log data. A separate HTML table view lets you browse all the logs using pagination. You can filter by start and end time, Docker container name and host name. 

A POST "/journal" route was added to retrieve journal logs sent by the `metrics-sender` client. The `metrics-sender` was also extended to send the journal logs.

## Authentication

I didn't want too much hassle adding authentication such as adding password hashing. [goth](https://github.com/markbates/goth) was chosen as a simple authentication provider and I simply just added Google authentication, because that is what I use the most. But it could easily be extended to other providers. 

One users table needed to be added to handled authorization. But this table simply contains the email address of the user. If it is present, the user is authorized to access the data.

## Sources

* [Interfaces](https://www.digitalocean.com/community/tutorials/how-to-use-interfaces-in-go)
* [JSON Serialization](https://emretanriverdi.medium.com/json-serialization-in-go-a27aeeb968de)
* [Postgres Connection Pool](https://medium.com/@neelkanthsingh.jr/understanding-database-connection-pools-and-the-pgx-library-in-go-3087f3c5a0c)
* [Cron](https://betterstack.com/community/questions/how-to-run-cron-jobs-in-go/)
* [Testing](https://blog.jetbrains.com/go/2022/11/22/comprehensive-guide-to-testing-in-go/)
* [Options vs Builder](https://blog.matthiasbruns.com/golang-options-vs-builder-pattern)
* [Http Request](https://www.digitalocean.com/community/tutorials/how-to-make-http-requests-in-go)
* [Environment Variables](https://towardsdatascience.com/use-environment-variable-in-your-next-golang-project-39e17c3aaa66)
* https://github.com/sigrdrifa/go-htmx-websockets-example/tree/main
* [Toasts with htmx](https://themurph.hashnode.dev/go-beyond-the-basics-mastering-toast-notifications-with-go-and-htmx)
* [templates cheatsheet](https://docs.gofiber.io/template/next/html/TEMPLATES_CHEATSHEET)
* https://flowbite.com/
* [template layouts/inheritance](https://stackoverflow.com/a/69244593)
* [Postgres Timescale DB](https://saiparvathaneni.medium.com/a-complete-guide-for-postgres-timescale-db-ae75a4d45b8d)
  * [With Prisma](https://medium.com/geekculture/set-up-a-timescaledb-hypertable-with-prisma-9550652cfe97), [or this](https://gist.github.com/janpio/2a425f22673f2de54469772f16af8118)
* https://www.haatos.com/articles/google-authentication-with-goth-in-golang
* [Google OAuth API keys](https://console.cloud.google.com/apis/credentials)
* [timescale postgres extension](https://www.timescale.com/)
