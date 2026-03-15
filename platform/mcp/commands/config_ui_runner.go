package commands

type ConfigUIRunner interface {
	Run(serviceName string) error
}
