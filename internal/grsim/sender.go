package grsim

import (
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/protobuf/proto"
)

// Sender grSimへのUDP送信を管理
type Sender struct {
	address string
	conn    *net.UDPConn
	mu      sync.Mutex
}

// NewSender 新しいSenderを作成
func NewSender(address string) *Sender {
	return &Sender{
		address: address,
	}
}

// Connect grSimへのUDP接続を確立
func (s *Sender) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		s.conn.Close()
	}

	addr, err := net.ResolveUDPAddr("udp", s.address)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	s.conn = conn
	return nil
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

// SendReplacement Replacementパケットを送信
func (s *Sender) SendReplacement(replacement *GrSim_Replacement) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		return fmt.Errorf("not connected")
	}

	// grSimはcommandsフィールドも必要とする
	// 空のcommandsでも必須フィールドを設定する必要がある
	packet := &GrSim_Packet{
		Commands: &GrSim_Commands{
			Timestamp:    proto.Float64(0.0),
			Isteamyellow: proto.Bool(false),
		},
		Replacement: replacement,
	}

	data, err := proto.Marshal(packet)
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %w", err)
	}

	_, err = s.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send packet: %w", err)
	}

	return nil
}

// ReplaceBall ボールを指定位置に配置
func (s *Sender) ReplaceBall(x, y float64) error {
	replacement := &GrSim_Replacement{
		Ball: &GrSim_BallReplacement{
			X:  proto.Float64(x),
			Y:  proto.Float64(y),
			Vx: proto.Float64(0),
			Vy: proto.Float64(0),
		},
	}
	return s.SendReplacement(replacement)
}

// ReplaceRobot ロボットを指定位置に配置
func (s *Sender) ReplaceRobot(x, y, dir float64, id uint32, yellowTeam bool) error {
	replacement := &GrSim_Replacement{
		Robots: []*GrSim_RobotReplacement{
			{
				X:          proto.Float64(x),
				Y:          proto.Float64(y),
				Dir:        proto.Float64(dir),
				Id:         proto.Uint32(id),
				Yellowteam: proto.Bool(yellowTeam),
				Turnon:     proto.Bool(true),
			},
		},
	}
	return s.SendReplacement(replacement)
}
