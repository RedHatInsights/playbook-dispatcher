package protocols

const (
	RunnerDirective    Directive = "rhc-worker-playbook"
	SatelliteDirective Directive = "rhc-cloud-connector-worker"
	LabelRunnerRequest           = "ansible"
	LabelSatRequest              = "satellite"
)

var (
	RunnerProtocol    = &runnerProtocol{}
	SatelliteProtocol = &satelliteProtocol{}
)
