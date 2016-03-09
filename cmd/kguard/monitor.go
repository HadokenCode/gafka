package main

import (
	"flag"
	"sync"
	"time"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/go-metrics"
	log "github.com/funkygao/log4go"
)

type Monitor struct {
	zone           string
	influxdbAddr   string
	influxdbDbName string

	stop chan struct{}

	executors []Executor
}

func (this *Monitor) Init() {
	var logFile string
	flag.StringVar(&logFile, "log", "stdout", "log filename")
	flag.StringVar(&this.zone, "z", "", "zone, required")
	flag.StringVar(&this.influxdbAddr, "influxAddr", "", "influxdb addr, required")
	flag.StringVar(&this.influxdbDbName, "db", "", "influxdb db name, required")
	flag.Parse()

	if this.zone == "" || this.influxdbDbName == "" || this.influxdbAddr == "" {
		panic("run help ")
	}

	this.executors = make([]Executor, 0)

	if logFile == "stdout" {
		log.AddFilter("stdout", log.INFO, log.NewConsoleLogWriter())
	} else {
		log.DeleteFilter("stdout")

		filer := log.NewFileLogWriter(logFile, false)
		filer.SetFormat("[%d %T] [%L] (%S) %M")
		log.AddFilter("file", log.INFO, filer)
	}
}

func (this *Monitor) addExecutor(e Executor) {
	this.executors = append(this.executors, e)
}

func (this *Monitor) Stop() {
	close(this.stop)
}

func (this *Monitor) ServeForever() {
	ctx.LoadFromHome()

	go InfluxDB(ctx.Hostname(), metrics.DefaultRegistry, time.Minute,
		this.influxdbAddr, this.influxdbDbName, "", "")

	zkzone := zk.NewZkZone(zk.DefaultConfig(this.zone, ctx.ZoneZkAddrs(this.zone)))
	defer zkzone.Close()

	wg := new(sync.WaitGroup)
	this.addExecutor(&MonitorTopics{zkzone: zkzone, tick: time.Minute, stop: this.stop, wg: wg})
	this.addExecutor(&MonitorBrokers{zkzone: zkzone, tick: time.Minute, stop: this.stop, wg: wg})
	this.addExecutor(&MonitorReplicas{zkzone: zkzone, tick: time.Minute, stop: this.stop, wg: wg})
	this.addExecutor(&MonitorConsumers{zkzone: zkzone, tick: time.Minute, stop: this.stop, wg: wg})
	this.addExecutor(&MonitorClusters{zkzone: zkzone, tick: time.Minute, stop: this.stop, wg: wg})
	this.addExecutor(&MonitorF5{tick: time.Minute, stop: this.stop, wg: wg})

	for _, e := range this.executors {
		wg.Add(1)
		go e.Run()
	}

	log.Info("all executors ready")

	<-this.stop
	wg.Wait()
}
