package db

import (
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

var (
	serviceDB *ORM
)

type ORM struct {
	*xorm.EngineGroup
}

type DBConfig struct {
	DSN          string        // write data source name.
	ReadDSN      []string      // read data source name.
	Active       int           // pool
	Idle         int           // pool
	ShowSql      bool          // pool
	IdleTimeout  time.Duration // connect max life time.
	QueryTimeout time.Duration // query sql timeout
	ExecTimeout  time.Duration // execute sql timeout
	TranTimeout  time.Duration // transaction sql timeout
}

func Client() *ORM {
	return serviceDB
}
func SetDB(c *DBConfig) *ORM {
	dbtype := "mysql"
	master, err := xorm.NewEngine(dbtype, c.DSN)
	if err != nil {
		panic("new orm master engine error " + err.Error())
	}
	var slaves []*xorm.Engine
	for _, v := range c.ReadDSN {
		slave, err := xorm.NewEngine(dbtype, v)
		if err != nil {
			panic("new orm slave engine error : " + v)
		}
		slaves = append(slaves, slave)
	}
	eg, err := xorm.NewEngineGroup(master, slaves)
	eg.DatabaseTZ = time.Local
	eg.SetMaxOpenConns(c.Idle)
	eg.SetMaxIdleConns(c.Active)
	eg.ShowSQL(c.ShowSql)
	if err != nil {
		panic("new orm engine group error : ")
	}
	serviceDB = &ORM{eg}
	return serviceDB
}

func parseDSNType(dsn string) string {
	return strings.Split(dsn, "://")[0]
}
