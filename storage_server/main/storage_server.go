package main

import (
	"context"
	"di_store/storage_server"
	"di_store/tracing"
	"di_store/util"
	"os"

	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	logReportCaller = kingpin.Flag("log-report-caller", "whether the log will include the calling method").Default("false").Bool()
	logLevel        *string
	hostname        = kingpin.Flag("hostname", "hostname of StorageServer").Default("").String()
	confPath        = kingpin.Arg("conf_path", "configure file path for StorageServer").Required().File()
)

func initLogLevel() {
	levelList := make([]string, len(log.AllLevels))
	for i, level := range log.AllLevels {
		levelName, _ := level.MarshalText()
		levelList[i] = string(levelName)
	}
	logLevel = kingpin.Flag("log-level", "Log level.").Short('l').Default("info").OverrideDefaultFromEnvar("LOG_LEVEL").Enum(levelList...)
}

func main() {
	initLogLevel()
	kingpin.Parse()

	level, _ := log.ParseLevel(*logLevel)
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(level)
	log.SetReportCaller(*logReportCaller)

	log.Debug("SetReportCaller: ", *logReportCaller)
	log.Debug("SetLevel: ", *logLevel)

	if tracing.Enabled {
		log.Info("Tracing enabled.")
		defer tracing.Close()
	}

	conf, err := util.ReadConfig((*confPath).Name())
	if err != nil {
		log.Fatal(err)
	}

	if *hostname == "" {
		h, err := os.Hostname()
		if err != nil {
			log.Fatal(err)
		}
		hostname = &h
	}

	serverInfo, err := conf.StorageServer(*hostname)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("StorageServer config: ", serverInfo)

	if len(conf.NodeTrackers) == 0 {
		log.Fatal("can not find any node tracker from config file")
	}
	trackerInfo := conf.NodeTrackers[0]

	// todo context
	server, err := storage_server.NewStorageServer(
		context.Background(), *hostname, serverInfo.RpcPort,
		trackerInfo.Hostname, trackerInfo.RpcPort, serverInfo.PlasmaSocket,
		serverInfo.PlasmaMemoryByte, 8, 8)
	if err != nil {
		log.Fatalf("failed to create StorageServer: %v", err)
	}
	server.Serve()
}
