package auth

type demo struct {
}

func (d *demo) Do(apiKey string) (Metadata, error) {
	return Metadata{
		TenantID: "demo",
		Scopes:   []string{ScopeReadEvents, ScopeWriteEvents},
	}, nil
}

func NewDemoValidator() ApiKeyValidator {
	return &demo{}
}
