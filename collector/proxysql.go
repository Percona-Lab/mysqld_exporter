// Scrape `SHOW MYSQL STATUS` (on a proxysql admin port).

package collector

import (
	"database/sql"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

const (
	// Subsystem.
	proxysql = "proxysql"
	// Query.
	proxysqlQuery = `SHOW MYSQL STATUS`
)

// ScrapeProxysqlstatus collects from `SHOW MYSQL STATUS`.
func ScrapeProxysqlStatus(db *sql.DB, ch chan<- prometheus.Metric) error {
	var (
		proxysqlRows *sql.Rows
		err          error
	)
	proxysqlRows, err = db.Query(fmt.Sprint(proxysqlQuery))
	if err != nil {
		fmt.Println("Error while getting proxysqlRows: ", err, ", proxysqlQuery was: ", fmt.Sprint(proxysqlQuery))
		return err
	}
	defer proxysqlRows.Close()

	var (
		vname string
		vvalue float64 
	)
	
	for proxysqlRows.Next() {
		if err := proxysqlRows.Scan(&vname, &vvalue); err != nil {
			fmt.Println("Error while running proxysqlRows.Scan: ", err)
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
	return nil
}
