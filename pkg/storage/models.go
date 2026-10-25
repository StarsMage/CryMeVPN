package storage

import "slices"

type ServerData struct {
	ID           string
	Name         string
	ServerIp     string
	Port         uint16
	ClientConfig string
	HTTPPort     uint16
}

type AppStorage struct {
	Servers []ServerData
}

// Method ServerData
func NewServer(id, name, ip, client, prkey string, port uint16) *ServerData {
	return &ServerData{
		ID:         id,
		Name:       name,
		ServerIp:   ip,
		Port:       port,
		ClientConfig: client,
	}
}

// Method AppStorage
func (a *AppStorage) AddServer(server ServerData) {
	a.Servers = append(a.Servers, server)
}

func (a *AppStorage) DeleteServer(id string) {
	a.Servers = slices.DeleteFunc(a.Servers, func(s ServerData) bool {
		return s.ID == id 
	})
}