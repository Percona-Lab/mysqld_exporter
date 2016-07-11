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
		return err
	}
	defer proxysqlRows.Close()

	proxysqlCols, err := proxysqlRows.Columns()
	if err != nil {
		return err
	}

	for proxysqlRows.Next() {
		// As the number of columns varies with mysqld versions,
		// and sql.Scan requires []interface{}, we need to create a
		// slice of pointers to the elements of slaveData.
		scanArgs := make([]interface{}, len(proxysqlCols))
		for i := range scanArgs {
			scanArgs[i] = &sql.RawBytes{}
		}
		var varn string
		var varv float64
		if err := proxysqlRows.Scan(&varn, &varv); err != nil {
			return err
		}

		if varn != "Client_Connections_created" &&
			varn != "Client_Connections_connected" &&
			varn != "Server_Connections_created" &&
			varn != "Active_Transactions" {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, proxysql, strings.ToLower(varn)),
				"ProxySQL metric from SHOW MYSQL STATUS.",
				[]string{varn},
				nil,
			),
			prometheus.UntypedValue,
			varv,
			varn,
		)
	}
	return nil
}
