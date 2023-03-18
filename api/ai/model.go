package ai

import (
	"encoding/json"
	"fmt"
)

type Role int

const (
	System Role = iota
	User
	Assistant
)

func (r Role) String() string {
	switch r {
	case System:
		return "system"
	case User:
		return "user"
	case Assistant:
		return "assistant"
	default:
		return ""
	}
}

func StringToRole(s string) (Role, error) {
	switch s {
	case "system":
		return System, nil
	case "user":
		return User, nil
	case "assistant":
		return Assistant, nil
	default:
		return 0, fmt.Errorf("invalid role string: %s", s)
	}
}

func (r *Role) UnmarshalJSON(data []byte) error {
	var roleStr string
	err := json.Unmarshal(data, &roleStr)
	if err != nil {
		return err
	}
	switch roleStr {
	case "system":
		*r = System
	case "user":
		*r = User
	case "assistant":
		*r = Assistant
	default:
		return fmt.Errorf("invalid role string: %s", roleStr)
	}
	return nil
}

func (r Role) MarshalJSON() ([]byte, error) {
	switch r {
	case System:
		return json.Marshal("system")
	case User:
		return json.Marshal("user")
	case Assistant:
		return json.Marshal("assistant")
	default:
		return nil, fmt.Errorf("invalid role value: %d", r)
	}
}
