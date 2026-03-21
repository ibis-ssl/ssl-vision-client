package main

import (
	"github.com/RoboCup-SSL/ssl-vision-client/frontend"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/config"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/gc"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/grsim"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/replay"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/tracked"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/vision"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	DetectionProvider func() *vision.SSL_DetectionFrame,
	TrackerProvider func() map[string]*tracked.TrackerWrapperPacket,
	GeometryProvider func() *vision.SSL_GeometryData,
	RefereeProvider func() *gc.Referee,
	cfg *config.Config,
	restarter config.ReceiverRestarter,
	replayService replay.Service,
	grSimSender *grsim.Sender,
) {
	mux.Handle("/", frontend.HandleFrontend())
	mux.Handle("/api/tracker/sources", tracked.HandleTrackerSources(TrackerProvider))
	mux.Handle("/api/tracker", tracked.HandleTracker(TrackerProvider))
	mux.Handle("/api/vision/detection", vision.HandleVisionDetection(DetectionProvider))
	mux.Handle("/api/vision/geometry", vision.HandleVisionGeometry(GeometryProvider))
	mux.Handle("/api/referee", gc.HandleReferee(RefereeProvider))
	mux.Handle("/api/config", config.HandleConfig(cfg, restarter))
	mux.Handle("/api/replay/upload", replay.HandleUpload(replayService))
	mux.Handle("/api/replay/load-path", replay.HandleLoadPath(replayService))
	mux.Handle("/api/replay/state", replay.HandleState(replayService))
	mux.Handle("/api/replay/control", replay.HandleControl(replayService))
	mux.Handle("/api/grsim/ball", grsim.HandleReplaceBall(grSimSender))
	mux.Handle("/api/grsim/robot", grsim.HandleReplaceRobot(grSimSender))
}
