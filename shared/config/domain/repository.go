package domain

import aiDomain "github.com/jcastilloa/context-distill/shared/ai/domain"

type Repository interface {
	OpenAIProviderConfig() aiDomain.ProviderConfig
	DistillProviderConfig() aiDomain.ProviderConfig
	ServiceConfig() ServiceConfig
}
