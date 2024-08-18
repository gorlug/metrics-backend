# Metrics Backend

## Motivation

Having alerts when you're running your own infrastructure is important. Otherwise, how can you know if it is not running anymore.

In the past I used to run an Elasticsearch cluster on my own. The Servers would send their metrics to it and via Kibana I set up alerts for certain metrics like "container isn't running", "disk is full". I used Elasticsearch also as a small JSON database. But maintaining and updating an Elasticsearch cluster is quite time-consuming. 

Next I used AWS Cloudwatch metrics for this purpose. I built a small Cli in Dart myself that

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
