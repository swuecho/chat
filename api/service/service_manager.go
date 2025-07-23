package service

import (
	"github.com/swuecho/chat_backend/repository"
)

type serviceManager struct {
	chat  ChatService
	auth  AuthService
	model ModelService
	file  FileService
}

// NewServiceManager creates a new service manager with all services
func NewServiceManager(repos repository.CoreRepositoryManager) ServiceManager {
	return &serviceManager{
		chat:  NewChatService(repos),
		auth:  NewAuthService(repos),
		model: NewModelService(repos),
		file:  NewFileService(repos),
	}
}

func (sm *serviceManager) Chat() ChatService {
	return sm.chat
}

func (sm *serviceManager) Auth() AuthService {
	return sm.auth
}

func (sm *serviceManager) Model() ModelService {
	return sm.model
}

func (sm *serviceManager) File() FileService {
	return sm.file
}