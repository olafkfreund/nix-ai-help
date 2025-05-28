package ai

type AIProvider interface {
	Query(prompt string) (string, error)
}

type ProviderFactory struct {
	providers map[string]AIProvider
}

func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: make(map[string]AIProvider),
	}
}

func (f *ProviderFactory) RegisterProvider(name string, provider AIProvider) {
	f.providers[name] = provider
}

func (f *ProviderFactory) GetProvider(name string) (AIProvider, bool) {
	provider, exists := f.providers[name]
	return provider, exists
}
