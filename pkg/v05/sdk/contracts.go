package sdk

type ExpectationLevel string

const (
	ExpectationRequired ExpectationLevel = "required"
	ExpectationOptional ExpectationLevel = "optional"
)

type GoSDKExpectation struct {
	Name       string
	Level      ExpectationLevel
	MinVersion string
}

type Profile string

const (
	ProfileCore      Profile = "go-sdk-core"
	ProfileCommunity Profile = "go-sdk-community"
)

type ProfileContract struct {
	ProfileName  Profile
	Expectations []GoSDKExpectation
}

func CommunityProfileContract() ProfileContract {
	return ProfileContract{
		ProfileName: ProfileCommunity,
		Expectations: []GoSDKExpectation{
			{Name: "context-aware clients", Level: ExpectationRequired, MinVersion: "1.0"},
			{Name: "deterministic serialization", Level: ExpectationRequired, MinVersion: "1.0"},
			{Name: "traceable event hooks", Level: ExpectationOptional, MinVersion: "1.3"},
		},
	}
}

func (p ProfileContract) HasExpectation(name string) bool {
	for _, expect := range p.Expectations {
		if expect.Name == name {
			return true
		}
	}
	return false
}

func (p ProfileContract) RequiredNames() []string {
	var out []string
	for _, expect := range p.Expectations {
		if expect.Level == ExpectationRequired {
			out = append(out, expect.Name)
		}
	}
	return out
}

func (p ProfileContract) Meets(expectations []GoSDKExpectation) bool {
	needed := map[string]GoSDKExpectation{}
	for _, expect := range p.Expectations {
		if expect.Level == ExpectationRequired {
			needed[expect.Name] = expect
		}
	}
	for _, candidate := range expectations {
		delete(needed, candidate.Name)
	}
	return len(needed) == 0
}
