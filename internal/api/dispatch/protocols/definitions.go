package protocols

const (
	RunnerDirective    Directive = "playbook"
	SatelliteDirective Directive = "playbook-sat"
	LabelRunnerRequest           = "ansible"
	LabelSatRequest              = "satellite"
)

var (
	RunnerProtocol    = &runnerProtocol{}
	SatelliteProtocol = &satelliteProtocol{}
)
