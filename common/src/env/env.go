package env

type Spec struct {
	DataDir string `envconfig:"DATA_DIR" default:"data"`
}
