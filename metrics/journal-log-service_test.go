package metrics

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestJournalLogService(t *testing.T) {
	t.Run("Should properly parse journal log string", func(t *testing.T) {
		logs := `{"_SOURCE_REALTIME_TIMESTAMP":"1723504380437713","SYSLOG_TIMESTAMP":"2024-08-12T23:13:00.437669742Z","_HOSTNAME":"couchdb-1"}
{"_HOSTNAME":"couchdb-1","_SOURCE_REALTIME_TIMESTAMP":"1723504381256471"}`

		expectedEntries := []*LogsEntry{
			{
				Id:   0,
				Time: getTime(1723504380437713),
				Hash: "1269ae2e091f514181b4c5d1445bb31bd6bab06949dd1fd8541d5462f4daa55a",
				Log: map[string]string{
					"_SOURCE_REALTIME_TIMESTAMP": "1723504380437713",
					"SYSLOG_TIMESTAMP":           "2024-08-12T23:13:00.437669742Z",
					"_HOSTNAME":                  "couchdb-1",
				},
			},
			{
				Id:   0,
				Time: getTime(1723504381256471),
				Hash: "0889d7854c79a7ea0e34e5502a1f077431a83372f32619a2e1ed00e535d27ab9",
				Log: map[string]string{
					"_HOSTNAME":                  "couchdb-1",
					"_SOURCE_REALTIME_TIMESTAMP": "1723504381256471",
				},
			},
		}

		entries := ParseJournalLogs(logs)
		PrintLogsEntries(entries)

		assert.EqualValues(t, expectedEntries, entries)
	})
}

func getTime(value int64) time.Time {
	return time.UnixMicro(value)
}

func checkError(err error) {
	if err != nil {
		log.Println(err)
		panic(err)
	}
}
