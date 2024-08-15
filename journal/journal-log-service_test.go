package journal

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestJournalLogService(t *testing.T) {
	location := GetLocation()
	t.Run("Should properly parse journal log string", func(t *testing.T) {
		logs := `{"__REALTIME_TIMESTAMP":"1723552255278199","SYSLOG_TIMESTAMP":"2024-08-12T23:13:00.437669742Z","_HOSTNAME":"couchdb-1"}
{"_HOSTNAME":"couchdb-1","_SOURCE_REALTIME_TIMESTAMP":"1723504381256471","__REALTIME_TIMESTAMP":"1723552255349199"}`

		expectedEntries := []*LogsEntry{
			{
				Id:   0,
				Time: getTime(1723552255278199, location),
				Hash: "be44de7412847a49f94abb202015b94ee690320e3c0f4a5d320933364fc9c5f9",
				Log: map[string]string{
					"__REALTIME_TIMESTAMP": "1723552255278199",
					"SYSLOG_TIMESTAMP":     "2024-08-12T23:13:00.437669742Z",
					"_HOSTNAME":            "couchdb-1",
				},
			},
			{
				Id:   0,
				Time: getTime(1723552255349199, location),
				Hash: "de774812e635dd91ea6e8d1434d1cf68e5bc655d32214088054b878556c6cd3e",
				Log: map[string]string{
					"_HOSTNAME":                  "couchdb-1",
					"_SOURCE_REALTIME_TIMESTAMP": "1723504381256471",
					"__REALTIME_TIMESTAMP":       "1723552255349199",
				},
			},
		}

		entries := ParseJournalLogs(logs)
		PrintLogsEntries(entries)

		assert.EqualValues(t, expectedEntries, entries)
	})
}

func getTime(value int64, location *time.Location) time.Time {
	return time.UnixMicro(value).In(location)
}

func checkError(err error) {
	if err != nil {
		log.Println(err)
		panic(err)
	}
}
