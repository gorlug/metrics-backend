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
				Log: map[string]any{
					"__REALTIME_TIMESTAMP": "1723552255278199",
					"SYSLOG_TIMESTAMP":     "2024-08-12T23:13:00.437669742Z",
					"_HOSTNAME":            "couchdb-1",
				},
			},
			{
				Id:   0,
				Time: getTime(1723552255349199, location),
				Hash: "de774812e635dd91ea6e8d1434d1cf68e5bc655d32214088054b878556c6cd3e",
				Log: map[string]any{
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

	t.Run("fix parsing failure", func(t *testing.T) {
		logs := `{"_MACHINE_ID":"c6f55861b8c4481ab08d2e40135beed7","_CMDLINE":"/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock","SYSLOG_TIMESTAMP":"2024-08-16T08:21:49.016715493Z","__REALTIME_TIMESTAMP":"1723796509016948","_HOSTNAME":"couchdb-1","__MONOTONIC_TIMESTAMP":"3185836231135","_PID":"2368","CONTAINER_NAME":"gitlab-runner","_SELINUX_CONTEXT":"unconfined\n","_SYSTEMD_SLICE":"system.slice","CONTAINER_ID":"e58d961e4c34","_UID":"0","_EXE":"/usr/bin/dockerd","CONTAINER_LOG_ORDINAL":"1281","_SYSTEMD_CGROUP":"/system.slice/docker.service","_TRANSPORT":"journal","CONTAINER_TAG":"e58d961e4c34","CONTAINER_LOG_EPOCH":"04819d45be105e7e3c2782adc1a827d22e9b3016ce07313aef964640e95a5fb9","_CAP_EFFECTIVE":"1ffffffffff","__CURSOR":"s=0aa8c6cb82724d9183495653ce21f45b;i=100bbd98;b=32ece28cb42a40c3a0b41d2e83d95054;m=2e5c2a309df;t=61fc8a695c374;x=ba676ad6755499af","CONTAINER_ID_FULL":"e58d961e4c34bc1fe9143da258bf078dfbf30c4d1a79cb4a38b098f5df52f41a","_BOOT_ID":"32ece28cb42a40c3a0b41d2e83d95054","_SYSTEMD_UNIT":"docker.service","_SYSTEMD_INVOCATION_ID":"cdd697cf4ca04683a4e4c3db574f9f2c","IMAGE_NAME":"gitlab/gitlab-runner:v17.1.0","_COMM":"dockerd","_GID":"0","MESSAGE":[67,104,101,99,107,105,110,103,32,102,111,114,32,106,111,98,115,46,46,46,32,114,101,99,101,105,118,101,100,32,32,32,32,32,32,32,32,32,32,32,32,32,32,32,32,32,32,32,32,32,27,91,48,59,109,32,32,106,111,98,27,91,48,59,109,61,53,53,52,50,32,114,101,112,111,95,117,114,108,27,91,48,59,109,61,104,116,116,112,115,58,47,47,103,105,116,46,103,111,114,108,117,103,46,100,101,47,97,99,104,105,109,47,109,101,116,114,105,99,115,45,98,97,99,107,101,110,100,46,103,105,116,32,114,117,110,110,101,114,27,91,48,59,109,61,57,50,49,48,56,57,97,101],"PRIORITY":"3","SYSLOG_IDENTIFIER":"e58d961e4c34","_SOURCE_REALTIME_TIMESTAMP":"1723796509016783"}`

		entries := ParseJournalLogs(logs)
		PrintLogsEntries(entries)
		assert.Equal(t, 1, len(entries))
		assert.Equal(t, entries[0].Log["CONTAINER_NAME"], "gitlab-runner")
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
