package metrics

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"strconv"
	"strings"
	"time"
)

type JournalLogService struct {
	connPool *pgxpool.Pool
}

func NewJournalLogService(dbUrl string) (*JournalLogService, error) {
	println("dbUrl", dbUrl)
	connPool, err := pgxpool.NewWithConfig(context.Background(), Config(dbUrl))
	if err != nil {
		log.Fatal("Error while creating connection to the database!!", err)
		return nil, err
	}
	return &JournalLogService{connPool: connPool}, nil
}

type LogsEntry struct {
	Id   int
	Time time.Time
	Hash string
	Log  map[string]string
}

func (e *LogsEntry) String() string {
	return fmt.Sprintf("Id: %d, Time: %v, Hash: %s, Log: %v", e.Id, e.Time, e.Hash, e.Log)
}

func PrintLogsEntries(entries []*LogsEntry) {
	for _, entry := range entries {
		fmt.Println(entry.String())
	}
}

type LogsEntryCopyFrom struct {
	entries []*LogsEntry
}

func (c *LogsEntryCopyFrom) Next() bool {
	return len(c.entries) > 0
}

func (c *LogsEntryCopyFrom) Values() ([]any, error) {
	entry := c.entries[0]
	c.entries = c.entries[1:]
	return []any{entry.Time, entry.Hash, entry.Log}, nil
}

func (c *LogsEntryCopyFrom) Err() error {
	return nil
}

func (s *JournalLogService) SaveJournalLogs(logs string) error {
	logEntries := ParseJournalLogs(logs)
	existingHashes, err := s.queryLogsWithTheSameHash(logEntries)
	log.Printf("existing hashes: %v", existingHashes)
	if err != nil {
		return err
	}
	filteredLogEntries := filterOutLogEntriesThatHaveTheSameHash(logEntries, existingHashes)
	if len(filteredLogEntries) == 0 {
		log.Println("No new logs to save")
		return nil
	}

	logEntriesCopyFrom := &LogsEntryCopyFrom{entries: filteredLogEntries}

	copyCount, err := s.connPool.CopyFrom(
		context.Background(),
		pgx.Identifier{"logs"},
		[]string{"time", "hash", "log"},
		logEntriesCopyFrom,
	)
	log.Printf("copied %d logs", copyCount)
	return err
}

func (s *JournalLogService) queryLogsWithTheSameHash(logEntries []*LogsEntry) ([]string, error) {
	hashes := make([]string, len(logEntries))
	for _, entry := range logEntries {
		hashes = append(hashes, entry.Hash)
	}
	query := fmt.Sprintf("select hash from logs where hash in ('%s')", strings.Join(hashes, "','"))
	rows, err := s.connPool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	logs := make([]string, 0)
	for rows.Next() {
		var hashValue string
		err := rows.Scan(&hashValue)
		if err != nil {
			return nil, err
		}
		logs = append(logs, hashValue)
	}
	return logs, nil
}

func filterOutLogEntriesThatHaveTheSameHash(logEntries []*LogsEntry, hashes []string) []*LogsEntry {
	logsMap := map[string]bool{}
	for _, hash := range hashes {
		logsMap[hash] = true
	}
	filteredLogEntries := make([]*LogsEntry, 0)
	for _, entry := range logEntries {
		if _, ok := logsMap[entry.Hash]; !ok {
			filteredLogEntries = append(filteredLogEntries, entry)
		}
	}
	return filteredLogEntries
}

func ParseJournalLogs(logs string) []*LogsEntry {
	splitLogs := strings.Split(logs, "\n")
	logsEntries := make([]*LogsEntry, 0)
	for _, logLine := range splitLogs {
		if logLine == "" {
			continue
		}
		logMap := map[string]string{}
		err := json.Unmarshal([]byte(logLine), &logMap)
		if err != nil {
			log.Printf("Error while unmarshalling log: %v", err)
			continue
		}
		timestampString := logMap["_SOURCE_REALTIME_TIMESTAMP"]
		timestampInt, err := strconv.ParseInt(timestampString, 10, 64)
		if err != nil {
			log.Printf("Error while converting timestamp to int: %v", err)
			continue
		}
		timestamp := time.UnixMicro(timestampInt)
		logsEntries = append(logsEntries, &LogsEntry{
			Time: timestamp,
			Log:  logMap,
			Hash: createHash(logLine),
		})
	}
	return logsEntries
}

func createHash(input string) string {
	h := sha256.New()
	h.Write([]byte(input))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (j *JournalLogService) Close() {
	j.connPool.Close()
}
