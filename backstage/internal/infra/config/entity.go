package infraconfig

// Config 是应用启动配置的根结构，对应 configs/config.yaml。
type Config struct {
	Port     int            `yaml:"port"`
	Database DatabaseConfig `yaml:"database"`
}

// DatabaseConfig 保存 MySQL 连接参数。
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}
