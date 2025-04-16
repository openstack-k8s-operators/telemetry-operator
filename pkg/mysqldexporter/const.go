/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package mysqldexporter provides functionality for managing MySQL exporter components
package mysqldexporter

const (
	// ServiceName -
	ServiceName = "mysqld-exporter"

	// MysqldExporterPort -
	MysqldExporterPort = 9104

	// DatabaseUsernamePrefix -
	DatabaseUsernamePrefix = "mysqld-exporter"

	// configFlag -
	configFlag = "--config.my-cnf"

	// configPath -
	configPath = "/etc/mysqld-exporter/config.cnf"

	// webConfigFlag -
	webConfigFlag = "--web.config.file"

	// webConfigPath -
	webConfigPath = "/etc/mysqld-exporter/web.cnf"
)

// collectorArgs - Arguments to mysqld_exporter for enabling / disabling collectors. In the future enable any additional collectors here.
var collectorArgs = []string{
	"--collect.global_status",
	"--collect.global_variables",
	"--no-collect.auto_increment.columns",
	"--no-collect.binlog_size",
	"--no-collect.engine_innodb_status",
	"--no-collect.engine_tokudb_status",
	"--no-collect.heartbeat",
	"--no-collect.heartbeat.utc",
	"--no-collect.info_schema.clientstats",
	"--no-collect.info_schema.innodb_metrics",
	"--no-collect.info_schema.innodb_tablespaces",
	"--no-collect.info_schema.innodb_cmp",
	"--no-collect.info_schema.innodb_cmpmem",
	"--no-collect.info_schema.processlist",
	"--no-collect.info_schema.query_response_time",
	"--no-collect.info_schema.replica_host",
	"--no-collect.info_schema.tables",
	"--no-collect.info_schema.tablestats",
	"--no-collect.info_schema.schemastats",
	"--no-collect.info_schema.userstats",
	"--no-collect.mysql.user",
	"--no-collect.perf_schema.eventsstatements",
	"--no-collect.perf_schema.eventsstatementssum",
	"--no-collect.perf_schema.eventswaits",
	"--no-collect.perf_schema.file_events",
	"--no-collect.perf_schema.file_instances",
	"--no-collect.perf_schema.indexiowaits",
	"--no-collect.perf_schema.memory_events",
	"--no-collect.perf_schema.tableiowaits",
	"--no-collect.perf_schema.tablelocks",
	"--no-collect.perf_schema.replication_group_members",
	"--no-collect.perf_schema.replication_group_member_stats",
	"--no-collect.perf_schema.replication_applier_status_by_worker",
	"--no-collect.slave_status",
	"--no-collect.slave_hosts",
	"--no-collect.sys.user_summary",
}
