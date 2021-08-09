package auth

type Metadata struct {
	TenantID string
	Scopes   []string
}

func (md Metadata) Empty() bool {
	return md.TenantID == ""
}

func (md Metadata) ValidateScope(scope string) bool {
	for _, scp := range md.Scopes {
		if scp == scope {
			return true
		}
	}

	return false
}
