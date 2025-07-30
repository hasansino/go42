package grpc

// PermissionRegistry hold list of permissions for each grpc method.
// It is responsibility of the adapter to register permissions.
type PermissionRegistry struct {
	storage map[string][]string
}

func NewPermissionRegistry() *PermissionRegistry {
	return &PermissionRegistry{
		storage: make(map[string][]string),
	}
}

func (pr *PermissionRegistry) Register(method string, permissions ...string) {
	if _, exists := pr.storage[method]; !exists {
		pr.storage[method] = make([]string, 0)
	}
	pr.storage[method] = append(pr.storage[method], permissions...)
}

func (pr *PermissionRegistry) PermissionsForMethod(method string) []string {
	if permissions, exists := pr.storage[method]; exists {
		return permissions
	}
	return nil
}
