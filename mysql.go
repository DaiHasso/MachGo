package MachGo

import (
	"fmt"

	logging "github.com/daihasso/slogging"

	_ "github.com/go-sql-driver/mysql" // Drivers are not directly utilized.
	"github.com/jmoiron/sqlx"
)

var mysqlAddressTemplate = "%s:%s@tcp(%s:%s)/%s" +
	"?parseTime=true&loc=US%%2FPacific"

func getMysqlDatabase(
	username,
	password,
	serverAddress,
	port,
	dbName string,
) (*sqlx.DB, error) {
	fullAddress := fmt.Sprintf(
		mysqlAddressTemplate,
		username,
		password,
		serverAddress,
		port,
		dbName,
	)

	logging.Debug(fullAddress).Send()

	return sqlx.Connect("mysql", fullAddress)
}
