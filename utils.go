package main

// Helper functions to create pointers

// int64Ptr creates a pointer to an int64 value
func int64Ptr(i int64) *int64 {
	return &i
}

// boolPtr creates a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}
