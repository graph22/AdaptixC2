package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Adaptix-Framework/axc2"
)

func (m *ModuleExtender) HandlerListenerValid(data string) error {
	var conf DNSConfig
	if err := json.Unmarshal([]byte(data), &conf); err != nil {
		return err
	}
	if conf.PortBind < 1 || conf.PortBind > 65535 {
		return errors.New("Port must be in the range 1-65535")
	}
	if conf.Domain == "" {
		return errors.New("domain is required")
	}
	return nil
}

func (m *ModuleExtender) HandlerCreateListenerDataAndStart(name string, configData string, listenerCustomData []byte) (adaptix.ListenerData, []byte, any, error) {
	var listenerData adaptix.ListenerData
	var customData []byte
	var listener *DNS
	var conf DNSConfig
	var err error

	if listenerCustomData == nil {
		err = json.Unmarshal([]byte(configData), &conf)
		if err != nil {
			return listenerData, customData, nil, err
		}
		randSlice := make([]byte, 16)
		_, _ = rand.Read(randSlice)
		conf.EncryptKey = randSlice[:16]
		conf.Protocol = "dns"
	} else {
		err = json.Unmarshal(listenerCustomData, &conf)
		if err != nil {
			return listenerData, customData, nil, err
		}
	}

	listener = &DNS{
		Name:   name,
		Config: conf,
	}

	err = listener.Start(m.ts)
	if err != nil {
		return listenerData, customData, nil, err
	}

	listenerData = adaptix.ListenerData{
		BindHost:  listener.Config.HostBind,
		BindPort:  fmt.Sprintf("%d", listener.Config.PortBind),
		AgentAddr: fmt.Sprintf("%s:%d", listener.Config.HostBind, listener.Config.PortBind),
		Status:    "Listen",
	}

	if !listener.Active {
		listenerData.Status = "Closed"
	}

	var buffer bytes.Buffer
	_ = json.NewEncoder(&buffer).Encode(listener.Config)
	customData = buffer.Bytes()

	return listenerData, customData, listener, nil
}

func (m *ModuleExtender) HandlerEditListenerData(name string, listenerObject any, configData string) (adaptix.ListenerData, []byte, bool) {
	var listenerData adaptix.ListenerData
	var customData []byte
	var ok bool

	listener := listenerObject.(*DNS)
	if listener.Name != name {
		return listenerData, customData, false
	}

	var conf DNSConfig
	if err := json.Unmarshal([]byte(configData), &conf); err != nil {
		return listenerData, customData, false
	}

	listener.Config.Domain = conf.Domain

	listenerData = adaptix.ListenerData{
		BindHost:  listener.Config.HostBind,
		BindPort:  fmt.Sprintf("%d", listener.Config.PortBind),
		AgentAddr: fmt.Sprintf("%s:%d", listener.Config.HostBind, listener.Config.PortBind),
		Status:    "Listen",
	}
	if !listener.Active {
		listenerData.Status = "Closed"
	}

	var buffer bytes.Buffer
	_ = json.NewEncoder(&buffer).Encode(listener.Config)
	customData = buffer.Bytes()

	ok = true
	return listenerData, customData, ok
}

func (m *ModuleExtender) HandlerListenerStop(name string, listenerObject any) (bool, error) {
	listener := listenerObject.(*DNS)
	if listener.Name != name {
		return false, nil
	}
	err := listener.Stop()
	return true, err
}

func (m *ModuleExtender) HandlerListenerGetProfile(name string, listenerObject any) ([]byte, bool) {
	listener := listenerObject.(*DNS)
	if listener.Name != name {
		return nil, false
	}
	var object bytes.Buffer
	_ = json.NewEncoder(&object).Encode(listener.Config)
	return object.Bytes(), true
}

func (m *ModuleExtender) HandlerListenerInteralHandler(name string, data []byte, listenerObject any) (string, error, bool) {
	return "", nil, false
}
