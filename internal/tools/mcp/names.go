package mcp

import (
	"fmt"
	"strings"
)

// FormatName builds a namespaced MCP tool id: namespace:tool@version.
func FormatName(namespace, tool, version string) string {
	if version == "" {
		version = "latest"
	}
	return namespace + ":" + tool + "@" + version
}

// ParseName splits namespace:tool@version. Tool may contain extra colons.
func ParseName(name string) (namespace, tool, version string, err error) {
	at := strings.LastIndex(name, "@")
	if at < 0 {
		return "", "", "", fmt.Errorf("mcp name %q: missing @version", name)
	}
	version = name[at+1:]
	rest := name[:at]
	colon := strings.Index(rest, ":")
	if colon < 0 {
		return "", "", "", fmt.Errorf("mcp name %q: missing namespace:", name)
	}
	namespace = rest[:colon]
	tool = rest[colon+1:]
	if namespace == "" || tool == "" || version == "" {
		return "", "", "", fmt.Errorf("mcp name %q: empty part", name)
	}
	return namespace, tool, version, nil
}
