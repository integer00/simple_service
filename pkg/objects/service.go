package objects

import (
	"github.com/hashicorp/consul/api"
)

type Service struct {
	Name    string
	Address string
	Port    int
	ID      string
}

func (s *Service) CreateAndRegisterConsulService() *api.Client {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic("Failed to create client")
	}

	err = client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		Name:    s.Name,
		Address: s.Address,
		Port:    s.Port,
		ID:      s.Name,
	})
	if err != nil {
		panic("Failed to create client")
	}

	return client
}
