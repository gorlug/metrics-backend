package metrics

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type MetricsService interface {
	SaveMetric(metric MetricValues) error
	SaveState(metric MetricValues, state MetricState) error
	GetAllMetrics() ([]Metric, error)
}

type DbMetricsService struct {
	connPool *pgxpool.Pool
}

func NewDBMetricsService(dbUrl string, alerter Alerter) (*DbMetricsService, error) {
	connPool, err := pgxpool.NewWithConfig(context.Background(), Config(dbUrl))
	if err != nil {
		log.Fatal("Error while creating connection to the database!!")
	}

	connection, err := connPool.Acquire(context.Background())
	if err != nil {
		err := alerter.NewAlert(GetFailedToGetMetricsMetric())
		if err != nil {
			log.Println("Failed to send alert", err)
		}
		log.Println("wat")
		log.Fatalf("Error while acquiring connection from the database pool!! %v", err)
	}
	defer connection.Release()

	err = connection.Ping(context.Background())
	if err != nil {
		log.Fatal("Could not ping database")
	}

	fmt.Println("Connected to the database!!")

	return &DbMetricsService{connPool: connPool}, nil
}

func (s *DbMetricsService) SaveMetric(metric MetricValues) error {
	insertDynStmt := `
insert into "metric" ("host", "name", "timestamp", "type", "value", "state")
values ($1, $2, $3, $4, $5, $6)
on conflict ("host", "name") do update
    set timestamp = $3,
        type      = $4,
        value     = $5
`
	_, e := s.connPool.Exec(context.Background(), insertDynStmt, metric.Host, metric.Name, metric.Timestamp, metric.Type, metric.Value, OK)
	return e
}

func (s *DbMetricsService) SaveState(metric MetricValues, state MetricState) error {
	insertDynStmt := `
update "metric" set state = $1 where host = $2 and name = $3;
`
	_, e := s.connPool.Exec(context.Background(), insertDynStmt, state, metric.Host, metric.Name)
	return e
}

func (s *DbMetricsService) GetAllMetrics() ([]Metric, error) {
	rows, err := s.connPool.Query(context.Background(), `select host, name, timestamp, type, value, state, id from metric
order by state desc, host, name 
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var metricValues MetricValues
		err := rows.Scan(&metricValues.Host, &metricValues.Name, &metricValues.Timestamp, &metricValues.Type, &metricValues.Value, &metricValues.State, &metricValues.Id)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, NewMetricBuilder().WithMetricValues(metricValues).Build())
	}

	return metrics, nil
}

func (s *DbMetricsService) Close() {
	s.connPool.Close()
}

func (s *DbMetricsService) DeleteMetric(id int) error {
	deleteDynStmt := `delete from "metric" where id = $1`
	_, e := s.connPool.Exec(context.Background(), deleteDynStmt, id)
	return e
}

func Config(dbUrl string) *pgxpool.Config {
	log.Printf("database url length: %v", len(dbUrl))
	const defaultMaxConns = int32(4)
	const defaultMinConns = int32(0)
	const defaultMaxConnLifetime = time.Hour
	const defaultMaxConnIdleTime = time.Minute * 30
	const defaultHealthCheckPeriod = time.Minute
	const defaultConnectTimeout = time.Second * 5

	dbConfig, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		log.Fatal("Failed to create a config, error: ", err)
	}

	dbConfig.MaxConns = defaultMaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifetime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	dbConfig.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
		log.Println("Before acquiring the connection pool to the database!!")
		return true
	}

	dbConfig.AfterRelease = func(c *pgx.Conn) bool {
		log.Println("After releasing the connection pool to the database!!")
		return true
	}

	dbConfig.BeforeClose = func(c *pgx.Conn) {
		log.Println("Closed the connection pool to the database!!")
	}

	return dbConfig
}
