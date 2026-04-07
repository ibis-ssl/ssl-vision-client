package sslsim

import (
	"fmt"
	"math"
	"net"
	"sync"

	"google.golang.org/protobuf/proto"
)

// Sender SSL Simulation Protocol の UDP 送信を管理
type Sender struct {
	address string
	conn    *net.UDPConn
	mu      sync.Mutex
}

// NewSender 新しい Sender を作成
func NewSender(address string) *Sender {
	return &Sender{
		address: address,
	}
}

// dial は指定アドレスへの UDP 接続を確立し、既存接続を置き換える。
// ネットワーク I/O をミューテックスの外で行い、接続の差し替えのみをロック下で行う。
func (s *Sender) dial(newAddr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", newAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %w", err)
	}

	newConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	s.mu.Lock()
	old := s.conn
	s.conn = newConn
	s.address = newAddr
	s.mu.Unlock()

	if old != nil {
		old.Close()
	}
	return nil
}

// Connect シミュレータへの UDP 接続を確立
func (s *Sender) Connect() error {
	return s.dial(s.address)
}

// Close 接続を閉じる
func (s *Sender) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}

// Reconnect 新しいアドレスで接続を張り直す
func (s *Sender) Reconnect(newAddr string) error {
	return s.dial(newAddr)
}

// sendCommand SimulatorCommand を UDP で送信
func (s *Sender) sendCommand(cmd *SimulatorCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		return fmt.Errorf("not connected")
	}

	data, err := proto.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	_, err = s.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	return nil
}

// ReplaceBall ボールを指定位置に配置（SSL Simulation Protocol）
func (s *Sender) ReplaceBall(x, y float64) error {
	cmd := &SimulatorCommand{
		Control: &SimulatorControl{
			TeleportBall: &TeleportBall{
				X: proto.Float32(float32(x)),
				Y: proto.Float32(float32(y)),
			},
		},
	}
	return s.sendCommand(cmd)
}

// ReplaceRobot ロボットを指定位置に配置（SSL Simulation Protocol）
// dir は度単位（フロントエンドの慣習に合わせ）、内部で SSL Sim Protocol のラジアンに変換する
func (s *Sender) ReplaceRobot(x, y, dir float64, id uint32, yellowTeam bool) error {
	team := SslSimTeam_BLUE
	if yellowTeam {
		team = SslSimTeam_YELLOW
	}

	orientationRad := float32(dir * math.Pi / 180.0)

	cmd := &SimulatorCommand{
		Control: &SimulatorControl{
			TeleportRobot: []*TeleportRobot{
				{
					Id: &SslSimRobotId{
						Id:   proto.Uint32(id),
						Team: team.Enum(),
					},
					X:           proto.Float32(float32(x)),
					Y:           proto.Float32(float32(y)),
					Orientation: proto.Float32(orientationRad),
				},
			},
		},
	}
	return s.sendCommand(cmd)
}
