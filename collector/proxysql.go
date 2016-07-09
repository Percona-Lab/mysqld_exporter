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

// func columnIndex(proxysqlCols []string, colName string) int {
// 	for idx := range proxysqlCols {
// 		if proxysqlCols[idx] == colName {
// 			return idx
// 		}
// 	}
// 	return -1
// }
//
// func columnValue(scanArgs []interface{}, proxysqlCols []string, colName string) string {
// 	var columnIndex = columnIndex(proxysqlCols, colName)
// 	if columnIndex == -1 {
// 		return ""
// 	}
// 	return string(*scanArgs[columnIndex].(*sql.RawBytes))
// }

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

	proxysqlCols, err := proxysqlRows.Columns()
	if err != nil {
		fmt.Println("Error while getting proxysqlCols: ", err)
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

		if err := proxysqlRows.Scan(scanArgs...); err != nil {
			fmt.Println("Error while running proxysqlRows.Scan: ", err)
			return err
		}

		clientConnectionsCreated := columnValue(scanArgs, proxysqlCols, "Client_Connections_created")
		clientConnectionsConnected := columnValue(scanArgs, proxysqlCols, "Client_Connections_connected")
		serverConnectionsCreated := columnValue(scanArgs, proxysqlCols, "Server_Connections_created")
		activeTransactions := columnValue(scanArgs, proxysqlCols, "Active_Transactions")

		for i, col := range proxysqlCols {
			fmt.Println("Scraping ", col)
			if value, ok := parseStatus(*scanArgs[i].(*sql.RawBytes)); ok { // Silently skip unparsable values.
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						prometheus.BuildFQName(namespace, proxysql, strings.ToLower(col)),
						"ProxySQL metric from SHOW MASTER STATUS.",
						[]string{"Client_Connections_created", "Client_Connections_connected", "Server_Connections_created", "Active_Transactions"},
						nil,
					),
					prometheus.UntypedValue,
					value,
					clientConnectionsCreated, clientConnectionsConnected, serverConnectionsCreated, activeTransactions,
				)
			}
		}
	}
	return nil
}
