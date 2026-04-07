package main

import (
	"github.com/RoboCup-SSL/ssl-vision-client/internal/config"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/gc"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/portmonitor"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/replay"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/sslsim"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/tracked"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/vision"
	"net/http"
)

func NewServer(
	DetectionProvider func() *vision.SSL_DetectionFrame,
	TrackerProvider func() map[string]*tracked.TrackerWrapperPacket,
	GeometryProvider func() *vision.SSL_GeometryData,
	RefereeProvider func() *gc.Referee,
	cfg *config.Config,
	restarter config.ReceiverRestarter,
	replayService replay.Service,
	simSender *sslsim.Sender,
	monitor *portmonitor.Monitor,
	reconnector config.SimSenderReconnector,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		DetectionProvider,
		TrackerProvider,
		GeometryProvider,
		RefereeProvider,
		cfg,
		restarter,
		reconnector,
		replayService,
		simSender,
		monitor,
	)
	return mux
}
