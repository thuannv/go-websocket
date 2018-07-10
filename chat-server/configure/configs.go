package configure

import (
	"fmt"

	"github.com/spf13/viper"
)

// AppConfigs is the struct containing all of application configs
type AppConfigs struct {
	// WebSocket server host
	WebSocketHost string
	// WebSocket listen port
	WebSocketPort int32
	// WebSocket max allow connections
	WebSocketConns int32
	//GRPC host
	GrpcHost string
	//GRPC port
	GrpcPort int32
}

// Load configs from file
func LoadConfigs(name, path string) (*AppConfigs, bool) {
	c := NewConfigs()
	err := c.Load(name, path)
	if err != nil {
		fmt.Println(err)
		return nil, false
	}
	return c, true
}

// NewConfigs is to create new empty configs object
func NewConfigs() *AppConfigs {
	return &AppConfigs{}
}

// Load configs from file specified by name which is located in path
func (c *AppConfigs) Load(name string, path string) error {
	viper.SetConfigName(name)
	viper.AddConfigPath(path)

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Cannot read configs for name=%s, path=%s", name, path)
	}

	c.WebSocketHost = viper.GetString("ws.host")
	c.WebSocketPort = viper.GetInt32("ws.port")
	c.WebSocketConns = viper.GetInt32("ws.conn")
	c.GrpcHost = viper.GetString("grpc.host")
	c.GrpcPort = viper.GetInt32("grpc.port")
	return nil
}

func (c *AppConfigs) String() string {
	return fmt.Sprintf(
		"ws.host=%s, ws.port=%d, ws.conn=%d, grpc.host=%s, grpc.port=%d",
		c.WebSocketHost,
		c.WebSocketPort,
		c.WebSocketConns,
		c.GrpcHost,
		c.GrpcPort)
}
