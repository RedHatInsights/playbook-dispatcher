package protocols

const (
	RunnerDirective    Directive = "rhc-worker-playbook"
	SatelliteDirective Directive = "foreman_rh_cloud"
	LabelRunnerRequest           = "ansible"
	LabelSatRequest              = "satellite"
)

var (
	RunnerProtocol    = &runnerProtocol{}
	SatelliteProtocol = &satelliteProtocol{}
)
