package config

import (
	"errors"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Config is the information keeper for generating go structs from type names.
type Config struct {
	// LogLevel sets the logging level
	LogLevel string `yaml:"log_level,omitempty"`
	// Endpoint is the URL for the GraphQL API
	Endpoint string `yaml:"endpoint"`
	// Auth contains details about how to authenticate to the API in the case that it's required.
	Auth AuthConfig `yaml:"auth"`
	// Cache contains information on how and where to store the schema.
	Cache CacheConfig `yaml:"cache"`
	// Packages contain the information on how to break up the schema into code packages.
	Packages []PackageConfig `yaml:"packages,omitempty"`
	// Generators configure the work engine of this project.
	Generators []GeneratorConfig `yaml:"generators,omitempty"`
}

// AuthConfig is the information necessary to authenticate to the NerdGraph API.
type AuthConfig struct {
	// Header is the name of the API request header that is used to authenticate.
	Header string `yaml:"header,omitempty"`
	// EnvVar is the name of the environment variable to attach to the above header.
	EnvVar string `yaml:"api_key_env_var,omitempty"`
}

// CacheConfig is the information necessary to store the NerdGraph schema in JSON.
type CacheConfig struct {
	// Enable or disable the schema caching.
	Enable bool `yaml:",omitempty"`
	// SchemaFile is the location where the schema should be cached.
	SchemaFile string `yaml:"schema_file,omitempty"`
}

// PackageConfig is the information about a single package, which types to include from the schema, and which generators to use for this package.
type PackageConfig struct {
	// Name is the string that is used to refer to the name of the package.
	Name string `yaml:"name,omitempty"`
	// Path is the relative path within the project.
	Path string `yaml:"path,omitempty"`
	// Types is a list of Type configurations to include in the package.
	Types []TypeConfig `yaml:"types,omitempty"`
	// Methods is a list of Method configurations to include in the package.
	Methods []MethodConfig `yaml:"methods,omitempty"`
	Queries []QueryConfig  `yaml:"queries,omitempty"`
	// Generators is a list of names that reference a generator in the Config struct.
	Generators []string `yaml:"generators,omitempty"`
	// Imports is a list of strings to represent what pacakges to import for a given package.
	Imports []string `yaml:"imports,omitempty"`
}

// GeneratorConfig is the information necessary to execute a generator.
type GeneratorConfig struct {
	// Name is the string that is used to reference a generator.
	Name string `yaml:"name,omitempty"`
	// TemplateDir is the path to the directory that contains all of the templates.
	TemplateDir string `yaml:"template_dir,omitempty"`
	// FileName is the target file that is to be generated.
	FileName string `yaml:"fileName,omitempty"`
	// TemplateName is the name of the template within the TemplateDir.
	TemplateName string `yaml:"templateName,omitempty"`
}

// MethodConfig is the information about the GraphQL methods.
type MethodConfig struct {
	// Name is the name of the GraphQL method.
	Name string `yaml:"name"`
}

// TypeConfig is the information about which types to render and any data specific to handling of the type.
type TypeConfig struct {
	Name string `yaml:"name"`
	// FieldTypeOverride is the Golang type to override whatever the default detected type would be for a given field.
	FieldTypeOverride string `yaml:"field_type_override,omitempty"`
	// CreateAs is used when creating a new scalar type to determine which Go type to use.
	CreateAs string `yaml:"create_as,omitempty"`
	// SkipTypeCreate allows the user to skip creating a Scalar type.
	SkipTypeCreate bool `yaml:"skip_type_create,omitempty"`
}

type QueryConfig struct {
	Path string `yaml:"path"`
}

const (
	DefaultCacheEnable     = false
	DefaultCacheSchemaFile = "schema.json"
	DefaultLogLevel        = "info"
	DefaultAuthHeader      = "Api-Key"
	DefaultAuthEnvVar      = "TUTONE_API_KEY"
)

// LoadConfig will load a config file at the specified path or error.
func LoadConfig(file string) (*Config, error) {
	if file == "" {
		return nil, errors.New("config file name required")
	}
	log.WithFields(log.Fields{
		"file": file,
	}).Debug("loading package definition")

	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}
	log.Tracef("definition: %+v", config)

	return &config, nil
}
