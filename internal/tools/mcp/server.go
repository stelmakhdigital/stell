package mcp

// MemoryServer is an in-process helper wrapping MemoryTransport.
type MemoryServer struct {
	Transport *MemoryTransport
}

// NewMemoryServer builds a fake MCP server for tests.
func NewMemoryServer(tools []ToolDef, call func(name string, args map[string]any) (string, error)) *MemoryServer {
	return &MemoryServer{Transport: &MemoryTransport{Tools: tools, CallFn: call}}
}
