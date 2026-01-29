package main

import (
	"github.com/RoboCup-SSL/ssl-vision-client/internal/config"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/gc"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/grsim"
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
	grSimSender *grsim.Sender,
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
		grSimSender,
	)
	return mux
}
