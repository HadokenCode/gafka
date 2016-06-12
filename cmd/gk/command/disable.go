package command

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/gocli"
	"github.com/go-ozzo/ozzo-dbx"
)

type Disable struct {
	Ui  cli.Ui
	Cmd string

	zone, cluster, topic, partitions string
	listMode                         bool
}

func (this *Disable) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("disable", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.StringVar(&this.cluster, "c", "", "")
	cmdFlags.StringVar(&this.topic, "t", "", "")
	cmdFlags.StringVar(&this.partitions, "p", "", "")
	cmdFlags.BoolVar(&this.listMode, "l", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	zkzone := zk.NewZkZone(zk.DefaultConfig(this.zone, ctx.ZoneZkAddrs(this.zone)))
	dsn, err := zkzone.KatewayMysqlDsn()
	if err != nil {
		this.Ui.Error(err.Error())
		return 1
	}

	db, err := dbx.Open("mysql", dsn)
	swallow(err)
	defer db.Close()

	if this.listMode {
		this.listDeadPartitions(db)
		return 0
	}

	if validateArgs(this, this.Ui).
		require("-z", "-c", "-t", "-p").
		invalid(args) {
		return 2
	}

	for _, p := range strings.Split(this.partitions, ",") {
		partitionId, err := strconv.Atoi(p)
		swallow(err)

		_, err = db.Insert("dead_partition", dbx.Params{
			"KafkaTopic": this.topic,
			"Partition":  this.partitions,
		}).Execute()
		swallow(err)

		this.Ui.Info(fmt.Sprintf("%s/%d: disabled", this.topic, partitionId))
	}

	return
}

func (this *Disable) listDeadPartitions(db *dbx.DB) {
	type DeadPartition struct {
		KafkaTopic string `db:"KafkaTopic"`
		Partition  int32  `db:"Partition"`
	}

	var deadPartitions []DeadPartition
	q := db.NewQuery("SELECT KafkaTopic, Partition FROM dead_partition")
	swallow(q.All(&deadPartitions))
	for _, dp := range deadPartitions {
		this.Ui.Output(fmt.Sprintf("%s/%d marked dead", dp.KafkaTopic, dp.Partition))
	}
}

func (this *Disable) Synopsis() string {
	return "Disable Pub topic partition"
}

func (this *Disable) Help() string {
	help := fmt.Sprintf(`
Usage: %s disable [options]

    Disable Pub topic partition

Options:

    -l
      List all disabled partitions

    -z zone

    -c cluster

    -t topic

    -p partition ids seperated by comma
      e,g. 0,2

`, this.Cmd)
	return strings.TrimSpace(help)
}