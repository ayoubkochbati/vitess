// Copyright 2014, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mysqlctl

import (
	"os"
	"time"

	log "github.com/golang/glog"
	"github.com/youtube/vitess/go/mysql"
	blproto "github.com/youtube/vitess/go/vt/binlog/proto"
	"github.com/youtube/vitess/go/vt/mysqlctl/proto"
)

/*
This file handles the differences between flavors of mysql.
*/

// MysqlFlavor is the abstract interface for a flavor.
type MysqlFlavor interface {
	// MasterPosition returns the ReplicationPosition of a master.
	MasterPosition(mysqld *Mysqld) (proto.ReplicationPosition, error)

	// SlaveStatus returns the ReplicationStatus of a slave.
	SlaveStatus(mysqld *Mysqld) (*proto.ReplicationStatus, error)

	// PromoteSlaveCommands returns the commands to run to change
	// a slave into a master
	PromoteSlaveCommands() []string

	// StartReplicationCommands returns the commands to start replicating from
	// a given master and position as specified in a ReplicationStatus.
	StartReplicationCommands(params *mysql.ConnectionParams, status *proto.ReplicationStatus) ([]string, error)

	// ParseGTID parses a GTID in the canonical format of this MySQL flavor into
	// a proto.GTID interface value.
	ParseGTID(string) (proto.GTID, error)

	// ParseReplicationPosition parses a replication position in the canonical
	// format of this MySQL flavor into a proto.ReplicationPosition struct.
	ParseReplicationPosition(string) (proto.ReplicationPosition, error)

	// SendBinlogDumpCommand sends the flavor-specific version of the
	// COM_BINLOG_DUMP command to start dumping raw binlog events over a slave
	// connection, starting at a given GTID.
	SendBinlogDumpCommand(mysqld *Mysqld, conn *SlaveConnection, startPos proto.ReplicationPosition) error

	// MakeBinlogEvent takes a raw packet from the MySQL binlog stream connection
	// and returns a BinlogEvent through which the packet can be examined.
	MakeBinlogEvent(buf []byte) blproto.BinlogEvent

	// WaitMasterPos waits until slave replication reaches at least targetPos.
	WaitMasterPos(mysqld *Mysqld, targetPos proto.ReplicationPosition, waitTimeout time.Duration) error
}

var mysqlFlavors map[string]MysqlFlavor = make(map[string]MysqlFlavor)

func mysqlFlavor() MysqlFlavor {
	f := os.Getenv("MYSQL_FLAVOR")
	if f == "" {
		if len(mysqlFlavors) == 1 {
			for k, v := range mysqlFlavors {
				log.Infof("Only one MySQL flavor declared, using %v", k)
				return v
			}
		}
		if v, ok := mysqlFlavors["GoogleMysql"]; ok {
			log.Info("MYSQL_FLAVOR is not set, using GoogleMysql flavor by default")
			return v
		}
		log.Fatal("MYSQL_FLAVOR is not set, and no GoogleMysql flavor registered")
	}
	if v, ok := mysqlFlavors[f]; ok {
		log.Infof("Using MySQL flavor %v", f)
		return v
	}
	log.Fatalf("MYSQL_FLAVOR is set to unknown value %v", f)
	panic("")
}
