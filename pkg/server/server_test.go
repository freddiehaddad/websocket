package main

import "testing"

func TestNewRouter(t *testing.T) {
	router := newRouter()

	if router.messages == nil {
		t.Errorf("messages channel was not initialized")
	}

	if router.register == nil {
		t.Errorf("register channel was not initialized")
	}

	if router.unregister == nil {
		t.Errorf("unregister channel was not initialized")
	}

	if router.clients == nil {
		t.Errorf("clients map was not initialized")
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name []byte
	}{
		{name: []byte("foo")},
	}

	for index, test := range tests {
		client := newClient(test.name, nil)
		name := string(test.name)
		if client.name != name {
			t.Errorf("test[%d]: client name wrong. expected=%s got=%s", index, name, client.name)
		}
	}
}
