// Scrape `SHOW MYSQL STATUS` (on a proxysql admin port).

package collector

import (
	"database/sql"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"strconv"
)

const (
	// Subsystem.
	proxysql = "proxysql"
	// Query.
	proxysqlQuery = `SHOW MYSQL STATUS`
	// query for command counters
	commandCountersQuery = `select Command, Total_Time_us, Total_cnt, cnt_100us, cnt_500us, cnt_1ms, cnt_5ms, cnt_10ms, cnt_50ms, cnt_100ms, cnt_500ms, cnt_1s, cnt_5s, cnt_10s, cnt_INFs from stats.stats_mysql_commands_counters` 
)

// ScrapeProxysqlstatus collects from `SHOW MYSQL STATUS`.
func ScrapeProxysqlStatus(db *sql.DB, ch chan<- prometheus.Metric) error {
	var (
		proxysqlRows *sql.Rows
		commandCountersRows *sql.Rows
		err          error
	)
	
	proxysqlRows, err = db.Query(fmt.Sprint(proxysqlQuery))	
	if err != nil {
		return err
	}
	
	commandCountersRows, err = db.Query(fmt.Sprint(commandCountersQuery))
	if err != nil {
		return err
	}
	
	defer proxysqlRows.Close()
	defer commandCountersRows.Close()

	var (
		vname string
		vvalue float64 
	)
	
	for proxysqlRows.Next() {
		if err := proxysqlRows.Scan(&vname, &vvalue); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, proxysql, vname), "Generic metric from SHOW MYSQL STATUS. ", []string{strings.ToLower(vname)}, nil,
			),			
			prometheus.UntypedValue,
			vvalue,
			vname,
		)
	}
	err = proxysqlRows.Err()
	if err == nil {
		proxysqlRows.Close()
	}

	cols, colerr := commandCountersRows.Columns()
	if colerr != nil {
		return colerr
	}
	counterVals := make([]interface{}, len(cols))
	for i, _ := range cols {
		//counterVals[i] = new(sql.RawBytes)
		counterVals[i] = new(string)
	}
	
	for commandCountersRows.Next() {
		if err := commandCountersRows.Scan(counterVals...); err != nil {
			return err
		}
		var command string
		var metric float64
		for i, _ := range counterVals {
			if i == 0 {
				command = *counterVals[i].(*string)
			} else {
				metric, _  = strconv.ParseFloat(*counterVals[i].(*string),64)
			}
			if i > 0 {
		    ch <- prometheus.MustNewConstMetric(
			    prometheus.NewDesc(
				    prometheus.BuildFQName(namespace, proxysql, fmt.Sprintf("%s_counter",command)), "Generic metric from stats_mysql_commands_counter. ", []string{strings.ToLower(command)}, nil,
			    ),			
			    prometheus.UntypedValue,
			    metric,
			    cols[i],
		    )
			}
		}

	}
	return nil
}
