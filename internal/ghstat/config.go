package ghstat

import (
	"bytes"
	"errors"
	"os"

	"github.com/spf13/viper"
)

// Config represents ghstat's configuration format
type Config struct {
	Leads []Lead `yaml:"leads"`
}

// Lead is a Canonical Hiring lead, who has a name and zero or more hiring roles
// that they manage
type Lead struct {
	Name  string  `yaml:"name"`
	Roles []int64 `yaml:"roles"`
}

// ParseConfig locates and parses the ghstat configuration
func ParseConfig(configFile string) (*Config, error) {
	viper.SetConfigType("yaml")

	// If the user specified a path to the config file manually, load that file
	if len(configFile) > 0 {
		b, err := os.ReadFile(configFile)
		if err != nil {
			return nil, errors.New("unable to read specified config file")
		}

		err = viper.ReadConfig(bytes.NewBuffer(b))
		if err != nil {
			return nil, errors.New("error parsing ghstat config file")
		}
	} else {
		// Otherwise check in the default locations
		viper.SetConfigName("ghstat")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/ghstat")

		err := viper.ReadInConfig()
		if err != nil {
			if errors.As(err, &viper.ConfigFileNotFoundError{}) {
				return nil, errors.New("no config file found, see 'ghstat --help' for details")
			}
			return nil, errors.New("error parsing ghstat config file")
		}
	}

	conf := &Config{}
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, errors.New("error parsing ghstat config file")
	}

	return conf, nil
}
