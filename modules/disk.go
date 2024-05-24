package modules

import (
	"encoding/json"
	"time"

	"golang.org/x/sys/unix"
)

type diskConfig struct {
	Enable   bool
	Interval time.Duration
	Disks    []string
}

func disk(ch chan<- Message, cfg *diskConfig) {
	go loopMessage(ch, "Disk", cfg.Enable, cfg.Interval, func() json.RawMessage {
		type diskStruct struct {
			Free, Total, Used int
			UsedPerc          float64
		}

		var (
			statfs            unix.Statfs_t
			disks             map[string]diskStruct
			v                 string
			free, total, used int
			err               error
		)

		disks = make(map[string]diskStruct)

		for _, v = range cfg.Disks {
			err = unix.Statfs(v, &statfs)
			if err != nil {
				panic(err)
			}

			free = int(statfs.Bfree) * int(statfs.Bsize)
			total = int(statfs.Blocks) * int(statfs.Bsize)
			used = total - free

			disks[v] = diskStruct{
				Free:     free,
				Total:    total,
				Used:     total - free,
				UsedPerc: float64(used) / float64(total) * 100,
			}
		}

		return marshalRawJson(disks)
	})
}