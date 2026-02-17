package discordshim

const CoverageTarget = 0.8

type RestEndpoint string

const (
	RestChannelsMessages  RestEndpoint = "rest.channels.messages"
	RestChannelsReactions RestEndpoint = "rest.channels.reactions"
	RestGuildMembers      RestEndpoint = "rest.guilds.members"
	RestGatewayBot        RestEndpoint = "rest.gateway.bot"
)

var RequiredRestEndpoints = []RestEndpoint{
	RestChannelsMessages,
	RestChannelsReactions,
	RestGatewayBot,
}

type GatewayIntent string

const (
	IntentGuildMembers GatewayIntent = "intent.guild.members"
	IntentMessages     GatewayIntent = "intent.messages"
	IntentPresence     GatewayIntent = "intent.presence"
)

var RequiredGatewayIntents = []GatewayIntent{
	IntentMessages,
	IntentGuildMembers,
}

type CoverageReport struct {
	RestCoverage    float64
	GatewayCoverage float64
}

func restScore(covered []RestEndpoint) float64 {
	if len(RequiredRestEndpoints) == 0 {
		return 1
	}
	remaining := map[RestEndpoint]struct{}{}
	for _, candidate := range RequiredRestEndpoints {
		remaining[candidate] = struct{}{}
	}
	for _, c := range covered {
		delete(remaining, c)
	}
	coveredCount := len(RequiredRestEndpoints) - len(remaining)
	return float64(coveredCount) / float64(len(RequiredRestEndpoints))
}

func gatewayScore(covered []GatewayIntent) float64 {
	if len(RequiredGatewayIntents) == 0 {
		return 1
	}
	remaining := map[GatewayIntent]struct{}{}
	for _, candidate := range RequiredGatewayIntents {
		remaining[candidate] = struct{}{}
	}
	for _, c := range covered {
		delete(remaining, c)
	}
	coveredCount := len(RequiredGatewayIntents) - len(remaining)
	return float64(coveredCount) / float64(len(RequiredGatewayIntents))
}

func CoverageFor(coveredRest []RestEndpoint, coveredGateway []GatewayIntent) CoverageReport {
	return CoverageReport{
		RestCoverage:    restScore(coveredRest),
		GatewayCoverage: gatewayScore(coveredGateway),
	}
}

func CombinedScore(report CoverageReport) float64 {
	return (report.RestCoverage + report.GatewayCoverage) / 2
}

func MeetsTarget(score float64) bool {
	return score >= CoverageTarget
}
